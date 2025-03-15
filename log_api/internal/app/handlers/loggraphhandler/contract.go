package loggraphhandler

import (
	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/dto"
)

type CSVRepository interface {
	GetAllCSVData() ([]byte, error)
}

type CSVParser interface {
	Parse(data []byte) ([]*dto.LogData, error)
}

type GraphService interface {
	TopTimeSpent(logs []*dto.LogData) dto.TopTimeSpentList
	TopCountries(logs []*dto.LogData) dto.TopCountriesPercentageList
	PlayersInfo() (*dto.PlayersInfo, error)
	OnlineStatistics(logs []*dto.LogData) dto.OnlineStatistics
}
