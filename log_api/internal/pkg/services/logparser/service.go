package logparser

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/dto"
	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/enums"
	"github.com/dmitriitimoshenko/nmrih/log_api/internal/tools"
)

const (
	maxConcurrentGoroutines = 100
	loggingTimeFormat       = "2006-01-02 15:04:05"
)

type Service struct {
	logRepository LogRepository
	csvGenerator  CSVGenerator
	csvRepository CSVRepository
	ipAPIClient   IPAPIClient
}

func NewService(
	logRepository LogRepository,
	csvGenerator CSVGenerator,
	csvRepository CSVRepository,
	ipAPIClient IPAPIClient,
) *Service {
	return &Service{
		logRepository: logRepository,
		csvGenerator:  csvGenerator,
		csvRepository: csvRepository,
		ipAPIClient:   ipAPIClient,
	}
}

func (s *Service) Parse(requestTimeStamp time.Time) error {
	dateFromPtr, err := s.csvRepository.GetLastSavedDate()
	if err != nil {
		return fmt.Errorf("failed to get last saved date: %w", err)
	}
	if dateFromPtr == nil {
		dateFromPtr = tools.ToPtr(time.Date(2025, time.March, 1, 0, 0, 0, 0, time.Local))
	}

	log.Printf("[LogParseService] Parsing logs from %s\n", dateFromPtr.Format(loggingTimeFormat))

	logs, err := s.logRepository.GetLogs()
	if err != nil {
		return fmt.Errorf("failed to get logs: %w", err)
	}

	log.Printf("[LogParseService] Found %d logs\n", len(logs))

	mappedLogs, err := s.mapLogs(logs, *dateFromPtr)
	if err != nil {
		err = fmt.Errorf("failed to structurize the logs: %w", err)
		log.Println(err)
		return err
	}

	log.Printf("[LogParseService] Mapped %d logs\n", len(mappedLogs))

	if len(mappedLogs) == 0 {
		return nil
	}

	log.Print("[LogParseService] Mapped logs are going to get into CSV\n")

	csvBytes, lastLogTime, err := s.csvGenerator.Generate(mappedLogs)
	if err != nil {
		return fmt.Errorf("failed to generate CSV: %w", err)
	}

	log.Printf(
		"[LogParseService] Generated CSV with data (not null: %v)\n",
		csvBytes != nil,
	)
	if lastLogTime == nil {
		log.Print("[LogParseService] Last log time is nil\n")
	} else {
		log.Printf("[LogParseService] Last log time is %s\n", lastLogTime.Format(loggingTimeFormat))
	}

	if lastLogTime == nil || lastLogTime.IsZero() {
		lastLogTime = &requestTimeStamp
	}

	log.Printf("[LogParseService] Last log time is ANYWAY %s\n", lastLogTime.Format(loggingTimeFormat))

	if err := s.csvRepository.Save(csvBytes, *lastLogTime); err != nil {
		return fmt.Errorf("failed to save mapped logs as CSV: %w", err)
	}

	log.Print("[LogParseService] Saved CSV\n")

	return nil
}

func (s *Service) mapLogs(logs map[string][]byte, dateFrom time.Time) ([]dto.LogData, error) {
	var (
		logData []dto.LogData
		wg      sync.WaitGroup
	)

	errChan := make(chan error, maxConcurrentGoroutines)
	logDataChan := make(chan dto.LogData, maxConcurrentGoroutines)

	for fileName, page := range logs {
		wg.Add(1)
		go func(fileName string, page []byte, dateFrom time.Time, errChan chan error, logDataChan chan dto.LogData) {
			defer wg.Done()

			linesCount := s.countLines(page)
			if linesCount == 0 {
				return
			}

			i := 0
			scanner := bufio.NewScanner(bytes.NewReader(page))
			for scanner.Scan() {
				line := scanner.Text()
				i++
				if linesCount <= i {
					break
				}
				s.processLine(fileName, line, dateFrom, logDataChan, errChan)
			}

			if err := scanner.Err(); err != nil {
				errChan <- fmt.Errorf("error reading log extracted from file \"%s\": %w", fileName, err)
			}
		}(fileName, page, dateFrom, errChan, logDataChan)
	}

	go func() {
		wg.Wait()
		close(logDataChan)
		close(errChan)
	}()

	var errs []error
	for logDataChan != nil || errChan != nil {
		select {
		case data, opened := <-logDataChan:
			if !opened {
				logDataChan = nil
				continue
			}
			logData = append(logData, data)
		case err, opened := <-errChan:
			if !opened {
				errChan = nil
				continue
			}
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return logData, nil
}

func (s *Service) processLine(
	fileName, line string,
	dateFrom time.Time,
	logDataChan chan dto.LogData,
	errChan chan error,
) {
	if line == "" {
		return
	}

	logDataEntry := dto.LogData{}

	switch {
	case strings.Contains(line, enums.Actions.Disconnected().String()):
		logDataEntry.Action = enums.Actions.Disconnected()
	case strings.Contains(line, enums.Actions.Connected().String()):
		logDataEntry.Action = enums.Actions.Connected()
	case strings.Contains(line, enums.Actions.Entered().String()):
		logDataEntry.Action = enums.Actions.Entered()
	case strings.Contains(line, enums.Actions.CommittedSuicide().String()):
		logDataEntry.Action = enums.Actions.CommittedSuicide()
	default:
		return
	}
	s.addNickAndTimeStamp(line, &logDataEntry, dateFrom, errChan)
	if logDataEntry.Action == enums.Actions.Connected() {
		s.addCountryIfIPAvailable(fileName, line, &logDataEntry, errChan)
	}

	if err := logDataEntry.Validate(); err != nil {
		errChan <- fmt.Errorf("failed to validate log data entry on line [%s]: %w", line, err)
		return
	}
	logDataChan <- logDataEntry
}

func (s *Service) countLines(data []byte) int {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	lineCount := 0
	for scanner.Scan() {
		lineCount++
	}
	return lineCount
}

func (s *Service) addNickAndTimeStamp(
	line string,
	logDataEntry *dto.LogData,
	dateFrom time.Time,
	errChan chan error,
) {
	timeStampStr := line[2:23]
	parsedTime, err := time.Parse("01/02/2006 - 15:04:05", timeStampStr)
	if err != nil {
		errChan <- fmt.Errorf("failed to parse timeStamp from extracted log: %w", err)
		return
	}
	if !parsedTime.After(dateFrom) {
		return
	}

	logDataEntry.TimeStamp = parsedTime
	logDataEntry.NickName = line[26:strings.Index(line, "<")]
}

func (s *Service) addCountryIfIPAvailable(
	fileName, line string,
	logDataEntry *dto.LogData,
	errChan chan error,
) {
	ipMatches := tools.IPRegex.FindAllString(line, -1)
	if len(ipMatches) > 1 {
		log.Println(
			"[WARN] Found more than one IP address in file [",
			fileName,
			"]",
		)
	}
	if len(ipMatches) == 0 {
		log.Println(
			"[WARN] Found no IP address in file [",
			fileName,
			"]",
		)
	} else {
		ip := ipMatches[len(ipMatches)-1]
		logDataEntry.IPAddress = ip
		ipInfo, err := s.ipAPIClient.GetCountryByIP(ip)
		if err != nil {
			errChan <- fmt.Errorf("failed to get country by IP [%s]: %w", ip, err)
			return
		}
		logDataEntry.Country = ipInfo.Country
	}
}
