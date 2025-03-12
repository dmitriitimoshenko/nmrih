package logparser

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/dmitriitimoshenko/nmrih/log_parser/internal/pkg/dto"
	"github.com/dmitriitimoshenko/nmrih/log_parser/internal/pkg/enums"
	"github.com/dmitriitimoshenko/nmrih/log_parser/internal/tools"
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

	logs, err := s.logRepository.GetLogs()
	if err != nil {
		return fmt.Errorf("failed to get logs: %w", err)
	}

	mappedLogs, err := s.mapLogs(logs, *dateFromPtr)
	if err != nil {
		return fmt.Errorf("failed to structurize the logs: %w", err)
	}
	if len(mappedLogs) == 0 {
		return nil
	}

	csvBytes, lastLogTime, err := s.csvGenerator.Generate(mappedLogs)
	if err != nil {
		return fmt.Errorf("failed to generate CSV: %w", err)
	}
	if lastLogTime == nil {
		lastLogTime = &requestTimeStamp
	}

	if err := s.csvRepository.Save(csvBytes, *lastLogTime); err != nil {
		return fmt.Errorf("failed to save mapped logs as CSV: %w", err)
	}

	return nil
}

func (s *Service) mapLogs(logs map[string][]byte, dateFrom time.Time) ([]dto.LogData, error) {
	var (
		logData []dto.LogData
		i       int
	)

	for fileName, page := range logs {
		linesCount := s.countLines(page)
		if linesCount == 0 {
			continue
		}

		i = 0

		scanner := bufio.NewScanner(bytes.NewReader(page))
		for scanner.Scan() {
			i++
			line := scanner.Text()

			logDataEntry := dto.LogData{}
			if strings.Contains(line, enums.Actions.Entered().String()) {
				logDataEntry.Action = enums.Actions.Entered()
			} else if strings.Contains(line, enums.Actions.Disconnected().String()) {
				logDataEntry.Action = enums.Actions.Disconnected()
			} else if strings.Contains(line, enums.Actions.Connected().String()) {
				logDataEntry.Action = enums.Actions.Connected()
			} else if strings.Contains(line, enums.Actions.CommittedSuicideAction().String()) {
				logDataEntry.Action = enums.Actions.CommittedSuicideAction()
			} else {
				continue
			}

			if linesCount > i {
				timeStampStr := line[2:23]

				parsedTime, err := time.Parse("01/02/2006 - 15:04:05", timeStampStr)
				if err != nil {
					return nil, fmt.Errorf("failed to parse timeStamp from extracted log: %w", err)
				}
				if parsedTime.Before(dateFrom) {
					continue
				}

				logDataEntry.TimeStamp = parsedTime
				logDataEntry.NickName = line[26:strings.Index(line, "<")]

				if logDataEntry.Action == enums.Actions.Connected() {
					ipMatches := tools.IPRegex.FindAllString(line, -1)
					if len(ipMatches) > 1 {
						log.Println("[WARN] Found more than one IP address in the line [", i, "] of the file [", fileName, "]")
					}
					for _, ip := range ipMatches {
						logDataEntry.IPAddress = ip
						ipInfo, err := s.ipAPIClient.GetCountryByIP(ip)
						if err != nil {
							return nil, fmt.Errorf("failed to get country by IP [%s]: %w", ip, err)
						}
						logDataEntry.Country = ipInfo.Country
					}
				}
			}

			logData = append(logData, logDataEntry)
		}

		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("error reading log extracted from file \"%s\": %w", fileName, err)
		}
	}

	return logData, nil
}

func (s *Service) countLines(data []byte) int {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	lineCount := 0
	for scanner.Scan() {
		lineCount++
	}
	return lineCount
}
