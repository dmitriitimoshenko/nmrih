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
		`L 03/15/2025 - 16:05:12: "BigZeeb<69><[U:1:xxxxxxxxxx]><>" connected, address "123.190.1.1:27005"`,
	))
}

func TestDateTimeRegex(t *testing.T) {
	assert.False(t, tools.DateTimeRegex.MatchString("01/01/2025 - 00:00:00"))

	assert.True(t, tools.DateTimeRegex.MatchString(
		`L 03/23/2025 - 08:05:10: "XXXXX<101><[U:1:xxxxxxxxxx]><>" committed suicide with "world"`,
	))
	assert.False(t, tools.DateTimeRegex.MatchString(
		`L 03/23/2X25 - 08:05:10: "XXXXX<101><[U:1:xxxxxxxxxx]><>" committed suicide with "world"`,
	))
	assert.False(t, tools.DateTimeRegex.MatchString(
		`L 03/23/2025 08:05:10: "XXXXX<101><[U:1:xxxxxxxxxx]><>" committed suicide with "world"`,
	))
}
