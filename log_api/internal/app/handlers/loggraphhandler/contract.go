package loggraphhandler

import (
	"context"
	"time"

	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/dto"
)

type redisCache interface {
	GetWithTimeout(ctx context.Context, key string, cacheTimeout time.Duration) (*string, error)
	SetWithTimeout(ctx context.Context, key, value string, ttlOverride *time.Duration, cacheTimeout time.Duration) error
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
