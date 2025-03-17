package csvgenerator

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"sort"
	"time"

	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/dto"
)

type CSVGenerator struct{}

func NewCSVGenerator() *CSVGenerator {
	return &CSVGenerator{}
}

func (c *CSVGenerator) Generate(logData []dto.LogData) ([]byte, *time.Time, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	header := []string{"TimeStamp", "NickName", "Action", "IPAddress", "Country"}
	if err := writer.Write(header); err != nil {
		return nil, nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	sort.Slice(logData, func(i, j int) bool {
		return logData[i].TimeStamp.Before(logData[j].TimeStamp)
	})

	for _, data := range logData {
		row := []string{
			data.TimeStamp.Format("2006-01-02 15:04:05"),
			data.NickName,
			data.Action.String(),
			data.IPAddress,
			data.Country,
		}
		if err := writer.Write(row); err != nil {
			return nil, nil, fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, nil, fmt.Errorf("failed to flush CSV writer: %w", err)
	}

	return buf.Bytes(), &logData[len(logData)-1].TimeStamp, nil
}
