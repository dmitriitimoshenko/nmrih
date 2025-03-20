package csvparser

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/dto"
	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/enums"
)

const minValuesCount = 4

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Parse(data []byte) ([]*dto.LogData, error) {
	reader := csv.NewReader(bytes.NewReader(data))

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read csv: %w", err)
	}

	if len(records) < 1 {
		return nil, errors.New("csv does not contain header")
	}

	var results []*dto.LogData
	for i, record := range records[1:] {
		if len(record) < minValuesCount {
			log.Println("record is too short, skipping: line No.", i+1)
			continue
		}

		ts, err := time.Parse("2006-01-02 15:04:05", record[0])
		if err != nil {
			return nil, fmt.Errorf("failed to parse timestamp in record %d: %w", i+1, err)
		}

		action := enums.Action(record[2])
		if !action.IsValid() {
			log.Printf("invalid action [%s], skipping: line No.%d\n", record[2], i+1)
			continue
		}

		logDataEntry := &dto.LogData{
			TimeStamp: ts,
			NickName:  record[1],
			Action:    action,
			IPAddress: record[3],
			Country:   record[4],
		}

		results = append(results, logDataEntry)
	}

	return results, nil
}
