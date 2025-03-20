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

	// assert.Equal(t, false, tools.IPRegex.MatchString("256.255.255.255"))
	// assert.Equal(t, false, tools.IPRegex.MatchString("-1.255.255.255"))
	assert.Equal(t, false, tools.IPRegex.MatchString("255.255.255"))
	assert.Equal(t, false, tools.IPRegex.MatchString(".255.255.255"))
	assert.Equal(t, false, tools.IPRegex.MatchString("255.255.255."))
	assert.Equal(t, false, tools.IPRegex.MatchString("255.255.255.INVALID"))
}
