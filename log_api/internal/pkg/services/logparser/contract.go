package logparser

import (
	"time"

	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/dto"
)

type logRepository interface {
	GetLogs() (map[string][]byte, error)
}

type csvGenerator interface {
	Generate(logData []dto.LogData) ([]byte, *time.Time, error)
}

type csvRepository interface {
	Save(data []byte, requestTimeStamp time.Time) error
	GetLastSavedDate() (*time.Time, error)
}

type ipAPIClient interface {
	GetCountryByIP(ip string) (*dto.IPInfo, error)
}
