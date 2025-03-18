package tools_test

import (
	"github.com/dmitriitimoshenko/nmrih/log_api/internal/tools"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToPtr(t *testing.T) {
	assert.Equal(t, *tools.ToPtr(1), 1)
	assert.Equal(t, *tools.ToPtr("qwe"), "qwe")
	assert.Equal(t, *tools.ToPtr(10.2), 10.2)
	assert.Equal(t, *tools.ToPtr(struct{ a int }{a: 1}), struct{ a int }{a: 1})
}
