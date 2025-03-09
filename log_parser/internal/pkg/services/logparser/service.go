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
	"github.com/dmitriitimoshenko/nmrih/log_parser/internal/pkg/services/logrepository"
	"github.com/dmitriitimoshenko/nmrih/log_parser/internal/tools"
)

const hoursInMonth = 24 * 30

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

func (s *Service) Parse() error {
	dateFrom := time.Now().Add(time.Hour * hoursInMonth)
	logs, err := s.logRepository.GetLogs(&logrepository.Filter{
		DateFrom: &dateFrom,
		DateTo:   nil,
	})
	if err != nil {
		return fmt.Errorf("failed to get logs: %w", err)
	}

	mappedLogs, err := s.mapLogs(logs)
	if err != nil {
		return fmt.Errorf("failed to structurize the logs: %w", err)
	}

	csvBytes, err := s.csvGenerator.Generate(mappedLogs)
	if err != nil {
		return fmt.Errorf("failed to generate CSV: %w", err)
	}

	if err := s.csvRepository.Save(csvBytes); err != nil {
		return fmt.Errorf("failed to save mapped logs as CSV: %w", err)
	}

	return nil
}

func (s *Service) mapLogs(logs map[string][]byte) ([]dto.LogData, error) {
	var (
		logData []dto.LogData
		i       int
	)

	for fileName, page := range logs {
		linesCount := s.countLines(page)
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
