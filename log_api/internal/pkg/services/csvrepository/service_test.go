package csvrepository_test

import (
	"os"
	"testing"
	"time"

	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/services/csvrepository"
	"github.com/dmitriitimoshenko/nmrih/log_api/internal/tools/testhelper"
	"github.com/stretchr/testify/assert"
)

func TestService_GetLastSavedDate(t *testing.T) {
	defaultDateTime := time.Date(2012, 12, 23, 12, 54, 45, 0, time.UTC)

	t.Parallel()
	tests := []struct {
		name   string
		assert func(t *testing.T, actualDate *time.Time, err error)
	}{
		{
			name: "success: return last saved date",
			assert: func(t *testing.T, actualDateTime *time.Time, err error) {
				assert.Nil(t, err)
				assert.NotEmpty(t, actualDateTime)
				assert.Equal(t, defaultDateTime, *actualDateTime)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			th := testhelper.NewTestHelper(t)
			th.UseTestEnv()

			cfg := csvrepository.NewConfig(os.Getenv("CSV_STORAGE_DIRECTORY"))
			service := csvrepository.NewService(*cfg)
			actualDateTime, err := service.GetLastSavedDate()
			test.assert(t, actualDateTime, err)
		})
	}
}
