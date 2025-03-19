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
	minLogLineLength        = 27
	dateTimeCSVLayout       = "2006-01-02 15:04:05"
	dateTimeLogLayout       = "01/02/2006 - 15:04:05"
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

	log.Printf("[LogParseService] Parsing logs from %s\n", dateFromPtr.Format(dateTimeCSVLayout))

	logs, err := s.logRepository.GetLogs()
	if err != nil {
		return fmt.Errorf("failed to get logs: %w", err)
	}

	log.Printf("[LogParseService] Found %d logs\n", len(logs))

	mappedLogs, err := s.mapLogs(logs, *dateFromPtr)
	if err != nil {
		return fmt.Errorf("failed to structurize the logs: %w", err)
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
		log.Printf("[LogParseService] Last log time is %s\n", lastLogTime.Format(dateTimeCSVLayout))
	}

	if lastLogTime == nil {
		lastLogTime = &requestTimeStamp
	}

	log.Printf("[LogParseService] Last log time is ANYWAY %s\n", lastLogTime.Format(dateTimeCSVLayout))

	if err := s.csvRepository.Save(csvBytes, *lastLogTime); err != nil {
		return fmt.Errorf("failed to save mapped logs as CSV: %w", err)
	}

	log.Print("[LogParseService] Saved CSV\n")

	return nil
}

func (s *Service) mapLogs(logs map[string][]byte, dateFrom time.Time) ([]dto.LogData, error) {
	var (
		wg          sync.WaitGroup
		logDataChan = make(chan dto.LogData, maxConcurrentGoroutines)
		errChan     = make(chan error, maxConcurrentGoroutines)
	)

	for fileName, content := range logs {
		wg.Add(1)
		go func(fileName string, content []byte, dateFrom time.Time) {
			defer wg.Done()
			s.processLogFile(fileName, content, dateFrom, logDataChan, errChan)
		}(fileName, content, dateFrom)
	}

	go func() {
		wg.Wait()
		close(logDataChan)
		close(errChan)
	}()

	var logEntries []dto.LogData
	for entry := range logDataChan {
		logEntries = append(logEntries, entry)
	}
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return logEntries, nil
}

func (s *Service) processLogFile(
	fileName string,
	content []byte,
	dateFrom time.Time,
	logDataChan chan<- dto.LogData,
	errChan chan<- error,
) {
	if len(content) == 0 {
		return
	}
	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		entry, err := s.parseLogLine(fileName, line, dateFrom)
		if err != nil {
			errChan <- fmt.Errorf("file %s: %w", fileName, err)
			continue
		}
		if entry != nil {
			logDataChan <- *entry
		}
	}
	if err := scanner.Err(); err != nil {
		errChan <- fmt.Errorf("error scanning file %s: %w", fileName, err)
	}
}

func (s *Service) parseLogLine(fileName, line string, dateFrom time.Time) (*dto.LogData, error) {
	var action enums.Action
	switch {
	case strings.Contains(line, enums.Actions.Entered().String()):
		action = enums.Actions.Entered()
	case strings.Contains(line, enums.Actions.Disconnected().String()):
		action = enums.Actions.Disconnected()
	case strings.Contains(line, enums.Actions.Connected().String()):
		action = enums.Actions.Connected()
	case strings.Contains(line, enums.Actions.CommittedSuicide().String()):
		action = enums.Actions.CommittedSuicide()
	default:
		return nil, nil
	}

	if len(line) < minLogLineLength {
		return nil, fmt.Errorf("line too short to parse required fields: %q", line)
	}

	// example: "01/02/2006 - 15:04:05"
	timeStampStr := line[2:23]
	parsedTime, err := time.Parse(dateTimeLogLayout, timeStampStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp %q: %w", timeStampStr, err)
	}
	if !parsedTime.After(dateFrom) {
		return nil, nil
	}

	idx := strings.Index(line, "<")
	if idx == -1 || idx < 26 {
		return nil, fmt.Errorf("failed to find nickname in line: %q", line)
	}
	nickName := strings.TrimSpace(line[26:idx])

	entry := dto.LogData{
		Action:    action,
		TimeStamp: parsedTime,
		NickName:  nickName,
	}

	if action == enums.Actions.Connected() {
		ipMatches := tools.IPRegex.FindAllString(line, -1)
		if len(ipMatches) == 0 {
			return nil, fmt.Errorf("no IP address found in line: %q", line)
		}
		if len(ipMatches) > 1 {
			log.Printf("[WARN] file %s: multiple IP addresses found in line, using the last one", fileName)
		}
		ip := ipMatches[len(ipMatches)-1]
		entry.IPAddress = ip
		ipInfo, err := s.ipAPIClient.GetCountryByIP(ip)
		if err != nil {
			return nil, fmt.Errorf("failed to get country for IP %s: %w", ip, err)
		}
		entry.Country = ipInfo.Country
	}
	return &entry, nil
}
