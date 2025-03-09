package logparser

import (
	"github.com/dmitriitimoshenko/nmrih/log_parser/internal/pkg/models"
	"github.com/dmitriitimoshenko/nmrih/log_parser/internal/pkg/services/logrepository"
)

type LogRepository interface {
	GetLogs(fiter *logrepository.Filter) (map[string][]byte, error)
}

type CSVGenerator interface {
	Generate(logData []models.LogData) ([]byte, error)
}

type CSVRepository interface {
	Save(data []byte) error
}
