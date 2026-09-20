package graph_test

import (
	"testing"
	"time"

	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/dto"
	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/enums"
	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/services/graph"
	"github.com/stretchr/testify/assert"
)

func connectedLog(nickName, country string) *dto.LogData {
	return &dto.LogData{
		TimeStamp: time.Date(2025, time.March, 2, 12, 0, 0, 0, time.UTC),
		NickName:  nickName,
		Action:    enums.Actions.Connected(),
		IPAddress: "123.234.123.234",
		Country:   country,
	}
}

func TestService_TopCountries(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		logs   []*dto.LogData
		assert func(t *testing.T, actual dto.TopCountriesPercentageList)
	}{
		{
			name: "success: percentages are shared between the countries",
			logs: []*dto.LogData{
				connectedLog("first", "UA"),
				connectedLog("second", "UA"),
				connectedLog("third", "UA"),
				connectedLog("fourth", "DE"),
				connectedLog("fifth", "DE"),
				{
					TimeStamp: time.Date(2025, time.March, 2, 13, 0, 0, 0, time.UTC),
					NickName:  "first",
					Action:    enums.Actions.Disconnected(),
				},
			},
			assert: func(t *testing.T, actual dto.TopCountriesPercentageList) {
				assert.Len(t, actual, 3) // UA, DE and "Other"
				assert.Equal(t, "UA", actual[0].Country)
				assert.InDelta(t, 60.0, actual[0].Percentage, 0.001)
				assert.Equal(t, "DE", actual[1].Country)
				assert.InDelta(t, 40.0, actual[1].Percentage, 0.001)
				assert.Equal(t, "Other", actual[2].Country)
				assert.InDelta(t, 0.0, actual[2].Percentage, 0.001)
			},
		},
		{
			name: "success: connections without a country are counted once as Unknown",
			logs: []*dto.LogData{
				connectedLog("first", ""),
				connectedLog("second", ""),
				connectedLog("third", "UA"),
			},
			assert: func(t *testing.T, actual dto.TopCountriesPercentageList) {
				var total float64
				for _, country := range actual {
					assert.NotEmpty(t, country.Country)
					total += country.Percentage
				}
				assert.InDelta(t, 100.0, total, 0.001)

				assert.Equal(t, "Unknown", actual[0].Country)
				assert.InDelta(t, 200.0/3, actual[0].Percentage, 0.001)
			},
		},
		{
			name: "success: no connections at all",
			logs: nil,
			assert: func(t *testing.T, actual dto.TopCountriesPercentageList) {
				assert.Empty(t, actual)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := graph.NewService(nil)
			test.assert(t, service.TopCountries(test.logs))
		})
	}
}
