package csvgenerator_test

import (
	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/dto"
	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/enums"
	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/services/csvgenerator"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCSVGenerator_Generate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		logData []dto.LogData
		assert  func(t *testing.T, b []byte, timeStamp *time.Time, err error)
	}{
		{
			name: "success: csv generated",
			logData: []dto.LogData{
				{
					TimeStamp: time.Date(2001, 12, 31, 1, 0, 0, 0, time.UTC),
					NickName:  "test",
					Action:    enums.Actions.Disconnected(),
				},
				{
					TimeStamp: time.Date(2000, 12, 31, 0, 0, 0, 0, time.UTC),
					NickName:  "test",
					Action:    enums.Actions.Connected(),
					IPAddress: "123.234.123.234",
					Country:   "RU",
				},
			},
			assert: func(t *testing.T, b []byte, timeStamp *time.Time, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, timeStamp)
				assert.Equal(t, time.Date(2001, 12, 31, 1, 0, 0, 0, time.UTC), *timeStamp)
				assert.Equal(
					t,
					[]byte{
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
					b,
				)
			},
		},
		{
			name:    "success: empty input (nil)",
			logData: nil,
			assert: func(t *testing.T, b []byte, timeStamp *time.Time, err error) {
				assert.NoError(t, err)
				assert.Nil(t, timeStamp)
				assert.Equal(t, []byte(nil), b)
			},
		},
		{
			name:    "success: empty input (empty)",
			logData: []dto.LogData{},
			assert: func(t *testing.T, b []byte, timeStamp *time.Time, err error) {
				assert.NoError(t, err)
				assert.Nil(t, timeStamp)
				assert.Equal(t, []byte(nil), b)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			c := csvgenerator.NewCSVGenerator()
			b, timeStamp, err := c.Generate(test.logData)
			test.assert(t, b, timeStamp, err)
		})
	}
}
