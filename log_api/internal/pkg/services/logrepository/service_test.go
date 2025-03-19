package logrepository_test

import (
	"os"
	"testing"

	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/services/logrepository"
	"github.com/dmitriitimoshenko/nmrih/log_api/internal/tools/testhelper"
	"github.com/stretchr/testify/assert"
)

func TestService_GetLogs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		assert func(t *testing.T, logs map[string][]byte, err error)
	}{
		{
			name: "success: got a few log entries",
			assert: func(t *testing.T, logs map[string][]byte, err error) {
				assert.NoError(t, err)
				assert.NotEmpty(t, logs)
				assert.Len(t, logs, 1)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			th := testhelper.NewTestHelper(t)
			th.UseTestEnv()

			cfg := logrepository.NewConfig(os.Getenv("LOGS_STORAGE_DIRECTORY"), os.Getenv("LOGS_FILE_PATTERN"))
			service := logrepository.NewService(*cfg)
			logs, err := service.GetLogs()
			test.assert(t, logs, err)
		})
	}
}
