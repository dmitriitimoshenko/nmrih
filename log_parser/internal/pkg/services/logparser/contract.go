package logparser

import (
	"github.com/dmitriitimoshenko/nmrih/log_parser/internal/pkg/dto"
	"github.com/dmitriitimoshenko/nmrih/log_parser/internal/pkg/services/logrepository"
)

type LogRepository interface {
	GetLogs(fiter *logrepository.Filter) (map[string][]byte, error)
}

type CSVGenerator interface {
	Generate(logData []dto.LogData) ([]byte, error)
}

type CSVRepository interface {
	Save(data []byte) error
}

type IPAPIClient interface {
	GetCountryByIP(ip string) (*dto.IPInfo, error)
}
