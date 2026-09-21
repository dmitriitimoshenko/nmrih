// Package servercheck answers one question: is the game server reachable from
// the outside, and does Steam actually list it in the global server browser.
package servercheck

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/rumblefrog/go-a2s"
)

const (
	// DefaultPort is the port a Source dedicated server answers A2S queries on.
	DefaultPort = 27015
	// DefaultTimeout is the deadline for a single query.
	DefaultTimeout = 5 * time.Second

	steamServersAtAddressURL = "https://api.steampowered.com/ISteamApps/GetServersAtAddress/v0001/"

	// The A2S info packet stores the AppID in 16 bits, so anything above 65535
	// comes back truncated or zeroed. The 64-bit GameID keeps it intact.
	appIDMask = 0xFFFFFF

	regionUSEast       = 0
	regionUSWest       = 1
	regionSouthAmerica = 2
	regionEurope       = 3
	regionAsia         = 4
	regionAustralia    = 5
	regionMiddleEast   = 6
	regionAfrica       = 7
	regionWorld        = 255
)

// Verdict is the outcome of a check, also used as the process exit code.
type Verdict int

const (
	// VerdictOK means the server answers queries and Steam lists it.
	VerdictOK Verdict = 0
	// VerdictUnreachable means the server did not answer an A2S query.
	VerdictUnreachable Verdict = 1
	// VerdictNotListed means the server is up but Steam does not know it.
	VerdictNotListed Verdict = 2
)

// Report is everything a single check found out.
type Report struct {
	Address    string       `json:"address"`
	IP         string       `json:"ip,omitempty"`
	QueriedAt  time.Time    `json:"queried_at"`
	Info       *Info        `json:"a2s,omitempty"`
	InfoError  string       `json:"a2s_error,omitempty"`
	Steam      *SteamServer `json:"steam_registration,omitempty"`
	SteamError string       `json:"steam_error,omitempty"`
}

// Info is what the server itself reports over A2S.
type Info struct {
	Name        string   `json:"name"`
	Map         string   `json:"map"`
	Game        string   `json:"game"`
	Folder      string   `json:"folder"`
	AppID       uint64   `json:"app_id"`
	ReportedID  uint16   `json:"reported_app_id"`
	Players     uint8    `json:"players"`
	MaxPlayers  uint8    `json:"max_players"`
	Bots        uint8    `json:"bots"`
	PasswordSet bool     `json:"password_protected"`
	VAC         bool     `json:"vac_secured"`
	Keywords    string   `json:"keywords,omitempty"`
	PlayerNames []string `json:"player_names,omitempty"`
}

// SteamServer is one entry of what Steam knows about an address.
type SteamServer struct {
	Addr     string `json:"addr"`
	GamePort int    `json:"gameport"`
	AppID    int    `json:"appid"`
	GameDir  string `json:"gamedir"`
	Region   int    `json:"region"`
	Secure   bool   `json:"secure"`
	LAN      bool   `json:"lan"`
}

type steamPayload struct {
	Response struct {
		Success bool          `json:"success"`
		Servers []SteamServer `json:"servers"`
	} `json:"response"`
}

// Verdict reports whether the server is reachable and globally listed.
func (r *Report) Verdict() Verdict {
	switch {
	case r.Info == nil:
		return VerdictUnreachable
	case r.Steam == nil:
		return VerdictNotListed
	default:
		return VerdictOK
	}
}

// RegionName turns the region code Steam stores into something readable.
func RegionName(region int) string {
	switch region {
	case regionUSEast:
		return "US East"
	case regionUSWest:
		return "US West"
	case regionSouthAmerica:
		return "South America"
	case regionEurope:
		return "Europe"
	case regionAsia:
		return "Asia"
	case regionAustralia:
		return "Australia"
	case regionMiddleEast:
		return "Middle East"
	case regionAfrica:
		return "Africa"
	case regionWorld:
		return "World"
	default:
		return "unknown"
	}
}

// Check queries the server and Steam. Failures are recorded in the report
// instead of being returned, so a half-answered check still tells you something.
func Check(ctx context.Context, host string, port int, timeout time.Duration) *Report {
	address := net.JoinHostPort(host, strconv.Itoa(port))

	report := &Report{
		Address:   address,
		QueriedAt: time.Now().UTC(),
	}

	ip, err := resolve(ctx, host)
	if err != nil {
		report.InfoError = err.Error()
		report.SteamError = err.Error()
		return report
	}
	report.IP = ip

	info, err := queryA2S(address, timeout)
	if err != nil {
		report.InfoError = err.Error()
	} else {
		report.Info = info
	}

	steam, err := querySteam(ctx, ip, port, timeout)
	if err != nil {
		report.SteamError = err.Error()
	} else {
		report.Steam = steam
	}

	return report
}

func resolve(ctx context.Context, host string) (string, error) {
	if ip := net.ParseIP(host); ip != nil {
		return host, nil
	}

	addresses, err := net.DefaultResolver.LookupHost(ctx, host)
	if err != nil {
		return "", fmt.Errorf("failed to resolve %s: %w", host, err)
	}
	if len(addresses) == 0 {
		return "", fmt.Errorf("no addresses for %s", host)
	}

	return addresses[0], nil
}

func queryA2S(address string, timeout time.Duration) (*Info, error) {
	client, err := a2s.NewClient(address, a2s.TimeoutOption(timeout))
	if err != nil {
		return nil, fmt.Errorf("failed to create A2S client: %w", err)
	}
	defer func() { _ = client.Close() }()

	serverInfo, err := client.QueryInfo()
	if err != nil {
		return nil, fmt.Errorf("A2S_INFO query failed: %w", err)
	}

	info := &Info{
		Name:        serverInfo.Name,
		Map:         serverInfo.Map,
		Game:        serverInfo.Game,
		Folder:      serverInfo.Folder,
		AppID:       uint64(serverInfo.ID),
		ReportedID:  serverInfo.ID,
		Players:     serverInfo.Players,
		MaxPlayers:  serverInfo.MaxPlayers,
		Bots:        serverInfo.Bots,
		PasswordSet: serverInfo.Visibility,
		VAC:         serverInfo.VAC,
	}

	if extended := serverInfo.ExtendedServerInfo; extended != nil {
		info.Keywords = extended.Keywords
		if appID := extended.GameID & appIDMask; appID != 0 {
			info.AppID = appID
		}
	}

	// Player names are a bonus: a server may refuse this query while still
	// answering the info one, and that is not worth failing the check over.
	if players, playerErr := client.QueryPlayer(); playerErr == nil && players != nil {
		for _, player := range players.Players {
			if player != nil && player.Name != "" {
				info.PlayerNames = append(info.PlayerNames, player.Name)
			}
		}
	}

	return info, nil
}

func querySteam(ctx context.Context, ip string, port int, timeout time.Duration) (*SteamServer, error) {
	requestURL := steamServersAtAddressURL + "?addr=" + url.QueryEscape(ip)

	requestCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	request, err := http.NewRequestWithContext(requestCtx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build Steam request: %w", err)
	}

	client := &http.Client{Timeout: timeout}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to reach the Steam API: %w", err)
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("steam API answered with %s", response.Status)
	}

	var payload steamPayload
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode the Steam response: %w", err)
	}

	return PickServer(payload.Response.Servers, port)
}

// PickServer returns the entry for the given game port, or the only one there
// is. An address can host several servers, so the port is what disambiguates.
func PickServer(servers []SteamServer, port int) (*SteamServer, error) {
	if len(servers) == 0 {
		return nil, errors.New("steam does not list any server at this address")
	}

	for i := range servers {
		if servers[i].GamePort == port {
			return &servers[i], nil
		}
	}

	return nil, fmt.Errorf("steam lists %d server(s) at this address, none on port %d", len(servers), port)
}
