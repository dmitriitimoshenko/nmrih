package tools_test

import (
	"testing"

	"github.com/dmitriitimoshenko/nmrih/log_api/internal/tools"
	"github.com/stretchr/testify/assert"
)

func TestIPRegex(t *testing.T) {
	assert.Equal(t, true, tools.IPRegex.MatchString("100.100.100.100"))
	assert.Equal(t, true, tools.IPRegex.MatchString("10.10.10.10"))
	assert.Equal(t, true, tools.IPRegex.MatchString("1.1.1.1"))
	assert.Equal(t, true, tools.IPRegex.MatchString("0.0.0.0"))
	assert.Equal(t, true, tools.IPRegex.MatchString("255.255.255.255"))

	assert.Equal(t, false, tools.IPRegex.MatchString("256.255.255.255"))
	// assert.Equal(t, false, tools.IPRegex.MatchString("-1.255.255.255"))
	assert.Equal(t, false, tools.IPRegex.MatchString("255.255.255"))
	assert.Equal(t, false, tools.IPRegex.MatchString(".255.255.255"))
	assert.Equal(t, false, tools.IPRegex.MatchString("255.255.255."))
	assert.Equal(t, false, tools.IPRegex.MatchString("255.255.255.INVALID"))

	assert.True(t, tools.IPRegex.MatchString(
		`L 03/15/2025 - 16:05:12: "BigZeeb<69><[U:1:1419671927]><>" connected, address "173.172.252.206:27005"`,
	))
}

func TestDateTimeRegex(t *testing.T) {
	assert.True(t, tools.DateTimeRegex.MatchString("01/01/2025 - 00:00:00"))
	assert.True(t, tools.DateTimeRegex.MatchString("31/12/2099 - 23:59:59"))
	assert.False(t, tools.DateTimeRegex.MatchString("1/01/2025 - 00:00:00"))
	assert.False(t, tools.DateTimeRegex.MatchString("01/01/2025 - 0:00:00"))
	assert.False(t, tools.DateTimeRegex.MatchString("a1/01/2025 - 00:00:00"))
	assert.False(t, tools.DateTimeRegex.MatchString("01/01/2025  00:00:00"))
}
