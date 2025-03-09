package csvgenerator

import (
	"bytes"
	"encoding/csv"
	"fmt"

	"github.com/dmitriitimoshenko/nmrih/log_parser/internal/pkg/models"
)

type CSVGenerator struct{}

func NewCSVGenerator() *CSVGenerator {
	return &CSVGenerator{}
}

func (c *CSVGenerator) Generate(logData []models.LogData) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write CSV header
	header := []string{"TimeStamp", "NickName", "Action", "IPAddress"}
	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write CSV rows
	for _, data := range logData {
		row := []string{
			data.TimeStamp.Format("2006-01-02 15:04:05"),
			data.NickName,
			data.Action.String(),
			data.IPAddress,
		}
		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	// Flush the writer
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("failed to flush CSV writer: %w", err)
	}

	return buf.Bytes(), nil
}
