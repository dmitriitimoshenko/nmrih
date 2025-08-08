package loggraphhandler

import (
	"context"
	"time"

	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/dto"
)

type redisCache interface {
	Get(ctx context.Context, key string) (string, bool, error)
	Set(ctx context.Context, key, value string, ttlOverride *time.Duration) error
}

type csvRepository interface {
	GetAllCSVData() ([]byte, error)
}

type csvParser interface {
	Parse(data []byte) ([]*dto.LogData, error)
}

type graphService interface {
	TopTimeSpent(logs []*dto.LogData) dto.TopTimeSpentList
	TopCountries(logs []*dto.LogData) dto.TopCountriesPercentageList
	PlayersInfo() (*dto.PlayersInfo, error)
	OnlineStatistics(logs []*dto.LogData) dto.OnlineStatistics
}
