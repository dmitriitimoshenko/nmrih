package servercheck_test

import (
	"testing"

	"github.com/dmitriitimoshenko/nmrih/log_api/internal/app/servercheck"
	"github.com/stretchr/testify/assert"
)

func TestPickServer(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		servers []servercheck.SteamServer
		port    int
		assert  func(t *testing.T, server *servercheck.SteamServer, err error)
	}{
		{
			name: "success: the entry on the requested port is picked",
			servers: []servercheck.SteamServer{
				{Addr: "1.2.3.4:27016", GamePort: 27016, AppID: 224260, GameDir: "nmrih"},
				{Addr: "1.2.3.4:27015", GamePort: 27015, AppID: 224260, GameDir: "nmrih", Region: 3},
			},
			port: 27015,
			assert: func(t *testing.T, server *servercheck.SteamServer, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, server)
				assert.Equal(t, "1.2.3.4:27015", server.Addr)
				assert.Equal(t, 3, server.Region)
			},
		},
		{
			name:    "failed: steam lists nothing at this address",
			servers: nil,
			port:    27015,
			assert: func(t *testing.T, server *servercheck.SteamServer, err error) {
				assert.Error(t, err)
				assert.Nil(t, server)
			},
		},
		{
			name: "failed: something is listed, but not on this port",
			servers: []servercheck.SteamServer{
				{Addr: "1.2.3.4:27016", GamePort: 27016},
			},
			port: 27015,
			assert: func(t *testing.T, server *servercheck.SteamServer, err error) {
				assert.Error(t, err)
				assert.Nil(t, server)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			server, err := servercheck.PickServer(test.servers, test.port)
			test.assert(t, server, err)
		})
	}
}

func TestReport_Verdict(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		report   servercheck.Report
		expected servercheck.Verdict
	}{
		{
			name:     "unreachable: no A2S answer",
			report:   servercheck.Report{},
			expected: servercheck.VerdictUnreachable,
		},
		{
			name:     "not listed: answers, but steam does not know it",
			report:   servercheck.Report{Info: &servercheck.Info{Name: "Krich Server"}},
			expected: servercheck.VerdictNotListed,
		},
		{
			name: "ok: answers and is listed",
			report: servercheck.Report{
				Info:  &servercheck.Info{Name: "Krich Server"},
				Steam: &servercheck.SteamServer{GamePort: 27015},
			},
			expected: servercheck.VerdictOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.expected, test.report.Verdict())
		})
	}
}

func TestRegionName(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "Europe", servercheck.RegionName(3))
	assert.Equal(t, "World", servercheck.RegionName(255))
	assert.Equal(t, "unknown", servercheck.RegionName(42))
}
