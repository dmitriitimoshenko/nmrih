package tools_test

import (
	"testing"
	"time"

	"github.com/dmitriitimoshenko/nmrih/log_api/internal/tools"
	"github.com/stretchr/testify/assert"
)

func TestGetCETLocation(t *testing.T) {
	loc := tools.GetCETLocation()
	assert.NotNil(t, loc, "Location should not be nil")

	name := loc.String()
	assert.True(t, name == "CET" || name == "UTC", "Location name should be CET or UTC, got: %s", name)

	now := time.Now().In(loc)
	zone, _ := now.Zone()
	assert.NotEmpty(t, zone, "Zone name should not be empty")
}
