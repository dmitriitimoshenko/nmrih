// Command servercheck reports whether the game server answers queries from the
// outside and whether Steam lists it in the global server browser.
//
//	go run ./cmd/servercheck
//	go run ./cmd/servercheck -addr=example.com -port=27015 -json
//
// It exits with 0 when the server is reachable and listed, 1 when it does not
// answer at all, and 2 when it answers but Steam does not know about it, so it
// can be dropped into a cron job or a monitoring check as is.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dmitriitimoshenko/nmrih/log_api/internal/app/servercheck"
)

const (
	defaultHost    = "rulat-bot.duckdns.org"
	duckDNSSuffix  = ".duckdns.org"
	overallTimeout = 30 * time.Second
)

func main() {
	// Everything happens in run() so that the deferred calls are not skipped by
	// the os.Exit below.
	os.Exit(run())
}

func run() int {
	host := flag.String("addr", publicHost(), "server address")
	port := flag.Int("port", envIntOr("NMRIH_PORT", envIntOr("SERVER_PORT", servercheck.DefaultPort)),
		"server query port")
	timeout := flag.Duration("timeout", servercheck.DefaultTimeout, "timeout of a single query")
	asJSON := flag.Bool("json", false, "print the report as JSON")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), overallTimeout)
	defer cancel()

	report := servercheck.Check(ctx, *host, *port, *timeout)

	if *asJSON {
		if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
			fmt.Fprintln(os.Stderr, "failed to encode the report:", err)
			return int(servercheck.VerdictUnreachable)
		}
	} else {
		printReport(report)
	}

	return int(report.Verdict())
}

func printReport(report *servercheck.Report) {
	fmt.Printf("%s (%s) at %s\n\n", report.Address, report.IP, report.QueriedAt.Format(time.RFC3339))

	printA2S(report)
	fmt.Println()
	printSteam(report)
	fmt.Println()
	printVerdict(report)
}

func printA2S(report *servercheck.Report) {
	fmt.Println("A2S query from the public internet")
	if report.Info == nil {
		fmt.Println("  no answer:", report.InfoError)
		fmt.Println("  the query port is not reachable, or the server is down")
		return
	}

	info := report.Info
	fmt.Printf("  name      : %s\n", info.Name)
	fmt.Printf("  map       : %s\n", info.Map)
	fmt.Printf("  game      : %s (%s, appid %d)\n", info.Game, info.Folder, info.AppID)
	fmt.Printf("  players   : %d/%d (%d bots)\n", info.Players, info.MaxPlayers, info.Bots)
	fmt.Printf("  password  : %s\n", yesNo(info.PasswordSet))
	fmt.Printf("  VAC       : %s\n", yesNo(info.VAC))

	if info.Keywords != "" {
		fmt.Printf("  tags      : %s\n", info.Keywords)
	}
	if len(info.PlayerNames) > 0 {
		fmt.Printf("  online    : %s\n", strings.Join(info.PlayerNames, ", "))
	}
	if info.ReportedID == 0 && info.AppID != 0 {
		fmt.Println("  note      : the info packet reports appid 0, the real one was read from the 64-bit GameID")
	}
}

func printSteam(report *servercheck.Report) {
	fmt.Println("Steam global server list")
	if report.Steam == nil {
		fmt.Println("  not listed:", report.SteamError)
		fmt.Println("  check sv_lan 0, sv_region, and whether the server registers with Steam on startup")
		return
	}

	steam := report.Steam
	fmt.Printf("  listed as : %s\n", steam.Addr)
	fmt.Printf("  game      : appid %d (%s)\n", steam.AppID, steam.GameDir)
	fmt.Printf("  region    : %d (%s)\n", steam.Region, servercheck.RegionName(steam.Region))
	fmt.Printf("  sv_lan    : %s\n", yesNo(steam.LAN))
	fmt.Printf("  VAC       : %s\n", yesNo(steam.Secure))
}

func printVerdict(report *servercheck.Report) {
	switch report.Verdict() {
	case servercheck.VerdictOK:
		fmt.Println("Verdict: reachable and listed globally.")
		if report.Info != nil && report.Info.Players == 0 {
			fmt.Println("  It is empty, and both server browsers push empty servers to the bottom")
			fmt.Println("  or hide them outright. Search it by name, or connect by address.")
		}
	case servercheck.VerdictUnreachable:
		fmt.Println("Verdict: the server does not answer queries from the outside.")
	case servercheck.VerdictNotListed:
		fmt.Println("Verdict: the server is up but missing from the global list.")
	default:
		fmt.Println("Verdict: unknown.")
	}
}

func yesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

// publicHost picks the address to check from the outside. SERVER_ADDR is
// deliberately not used: since the game server moved into this compose stack it
// holds a container name, and checking that would prove nothing about whether
// the server is reachable from the internet.
func publicHost() string {
	if value := os.Getenv("SERVER_PUBLIC_ADDR"); value != "" {
		return value
	}
	if subdomain := os.Getenv("DUCKDNS_SUBDOMAIN"); subdomain != "" {
		return subdomain + duckDNSSuffix
	}
	return defaultHost
}

func envIntOr(name string, fallback int) int {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}

	parsed := 0
	if _, err := fmt.Sscanf(value, "%d", &parsed); err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
