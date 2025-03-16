package logparser

import (
	"time"

	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/dto"
)

type LogRepository interface {
	GetLogs() (map[string][]byte, error)
}

type CSVGenerator interface {
	Generate(logData []*dto.LogData) ([]byte, *time.Time, error)
}

type CSVRepository interface {
	Save(data []byte, requestTimeStamp time.Time) error
	GetLastSavedDate() (*time.Time, error)
}

type IPAPIClient interface {
	GetCountriesByIPs(ips []string) (dto.IPInfo, error)
}
