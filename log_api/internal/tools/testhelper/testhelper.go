package testhelper

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

type TestHelper struct {
	t *testing.T
}

func NewTestHelper(t *testing.T) *TestHelper {
	return &TestHelper{
		t: t,
	}
}

func (th *TestHelper) UseTestEnv() {
	homeDir := "/home/runner/work/nmrih"
	_, err := os.ReadDir(homeDir)
	if err != nil {
		homeDir, err = os.UserHomeDir()
		assert.NoError(th.t, err)
	}
	envPath := filepath.Join(homeDir, "nmrih", "log_api", ".env.test")
	assert.NoError(th.t, godotenv.Load(envPath))
}
