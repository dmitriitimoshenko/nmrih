package csvparser_test

import (
	"testing"
	"time"

	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/dto"
	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/enums"
	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/services/csvparser"
	"github.com/stretchr/testify/assert"
)

func TestService_Parse(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		data   []byte
		assert func(t *testing.T, logData []*dto.LogData, err error)
	}{
		{
			name: "success: csv parsed",
			data: []byte{
				0x54, 0x69, 0x6d, 0x65, 0x53, 0x74, 0x61, 0x6d, 0x70, 0x2c, 0x4e, 0x69, 0x63, 0x6b, 0x4e, 0x61,
				0x6d, 0x65, 0x2c, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2c, 0x49, 0x50, 0x41, 0x64, 0x64, 0x72,
				0x65, 0x73, 0x73, 0x2c, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x79, 0xa, 0x32, 0x30, 0x30, 0x30,
				0x2d, 0x31, 0x32, 0x2d, 0x33, 0x31, 0x20, 0x30, 0x30, 0x3a, 0x30, 0x30, 0x3a, 0x30, 0x30, 0x2c,
				0x74, 0x65, 0x73, 0x74, 0x2c, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x65, 0x64, 0x2c, 0x31,
				0x32, 0x33, 0x2e, 0x32, 0x33, 0x34, 0x2e, 0x31, 0x32, 0x33, 0x2e, 0x32, 0x33, 0x34, 0x2c, 0x52,
				0x55, 0xa, 0x32, 0x30, 0x30, 0x31, 0x2d, 0x31, 0x32, 0x2d, 0x33, 0x31, 0x20, 0x30, 0x31, 0x3a,
				0x30, 0x30, 0x3a, 0x30, 0x30, 0x2c, 0x74, 0x65, 0x73, 0x74, 0x2c, 0x64, 0x69, 0x73, 0x63, 0x6f,
				0x6e, 0x6e, 0x65, 0x63, 0x74, 0x65, 0x64, 0x2c, 0x2c, 0xa,
			},
			assert: func(t *testing.T, logData []*dto.LogData, err error) {
				assert.NoError(t, err)
				expectedLogs := []*dto.LogData{
					{
						TimeStamp: time.Date(2000, 12, 31, 0, 0, 0, 0, time.UTC),
						NickName:  "test",
						Action:    enums.Actions.Connected(),
						IPAddress: "123.234.123.234",
						Country:   "RU",
					},
					{
						TimeStamp: time.Date(2001, 12, 31, 1, 0, 0, 0, time.UTC),
						NickName:  "test",
						Action:    enums.Actions.Disconnected(),
					},
				}
				assert.Len(t, logData, len(expectedLogs))
				for i, log := range expectedLogs {
					assert.Equal(t, log.TimeStamp, logData[i].TimeStamp)
					assert.Equal(t, log.NickName, logData[i].NickName)
					assert.Equal(t, log.Action, logData[i].Action)
					assert.Equal(t, log.IPAddress, logData[i].IPAddress)
					assert.Equal(t, log.Country, logData[i].Country)
				}
			},
		},
		{
			name: "success: no records in csv",
			data: []byte{
				0x01, // this is not a valid csv
			},
			assert: func(t *testing.T, logData []*dto.LogData, err error) {
				assert.NoError(t, err)
				assert.Nil(t, logData)
			},
		},
		{
			name: "failed: invalid data",
			data: []byte(nil),
			assert: func(t *testing.T, logData []*dto.LogData, err error) {
				assert.Error(t, err)
				assert.Nil(t, logData)
			},
		},
		{
			name: "success: one line in csv invalid - skipping",
			data: []byte{
				0x54,
				0x69,
				0x6d,
				0x65,
				0x53,
				0x74,
				0x61,
				0x6d,
				0x70,
				0x2c,
				0x4e,
				0x69,
				0x63,
				0x6b,
				0x4e,
				0x61,
				0x6d,
				0x65,
				0x2c,
				0x41,
				0x63,
				0x74,
				0x69,
				0x6f,
				0x6e,
				0x2c,
				0x49,
				0x50,
				0x41,
				0x64,
				0x64,
				0x72,
				0x65,
				0x73,
				0x73,
				0x2c,
				0x43,
				0x6f,
				0x75,
				0x6e,
				0x74,
				0x72,
				0x79,
				0xa,
				0x32,
				0x30,
				0x30,
				0x30,
				0x2d,
				0x31,
				0x32,
				0x2d,
				0x33,
				0x31,
				0x20,
				0x30,
				0x30,
				0x3a,
				0x30,
				0x30,
				0x3a,
				0x30,
				0x30,
				0x2c,
				0x74,
				0x65,
				0x73,
				0x74,
				0x2c,
				0x63,
				0x6f,
				0x6e,
				0x6e,
				0x65,
				0x63,
				0x74,
				0x65,
				0x64,
				0x2c,
				0x31,
				0x32,
				0x33,
				0x2e,
				0x32,
				0x33,
				0x34,
				0x2e,
				0x31,
				0x32,
				0x33,
				0x2e,
				0x32,
				0x33,
				0x34,
				0x2c,
				0x52,
				0x55,
				0xa,
				0x32,
				0x30,
				0x30,
				0x31,
				0x2d,
				0x31,
				0x32,
				0x2d,
				0x33,
				0x31,
				0x20,
				0x30,
				0x31,
				0x3a,
				0x30,
				0x30,
				0x3a,
				0x30,
				0x30,
				0x2c,
				0x74,
				0x65,
				0x73,
				0x74,
				0x2c,
				0x31,
				0x32,
				0x33,
				0x2c,
				0x2c,
				0xa,
			},
			assert: func(t *testing.T, logData []*dto.LogData, err error) {
				assert.NoError(t, err)
				expectedLogs := []*dto.LogData{
					{
						TimeStamp: time.Date(2000, 12, 31, 0, 0, 0, 0, time.UTC),
						NickName:  "test",
						Action:    enums.Actions.Connected(),
						IPAddress: "123.234.123.234",
						Country:   "RU",
					},
				}
				assert.Len(t, logData, len(expectedLogs))
				for i, log := range expectedLogs {
					assert.Equal(t, log.TimeStamp, logData[i].TimeStamp)
					assert.Equal(t, log.NickName, logData[i].NickName)
					assert.Equal(t, log.Action, logData[i].Action)
					assert.Equal(t, log.IPAddress, logData[i].IPAddress)
					assert.Equal(t, log.Country, logData[i].Country)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := csvparser.NewService()
			logData, err := service.Parse(test.data)
			test.assert(t, logData, err)
		})
	}
}
