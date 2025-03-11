package loggraphhandler

import "github.com/dmitriitimoshenko/nmrih/log_parser/internal/pkg/dto"

type CSVRepository interface {
	GetAllCSVData() ([]byte, error)
}

type CSVParser interface {
	Parse(data []byte) ([]*dto.LogData, error)
}
