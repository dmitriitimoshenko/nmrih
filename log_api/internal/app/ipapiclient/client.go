package ipapiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/dto"
)

const timeoutSeconds = 3

type IPAPIClient struct{}

func NewIPAPIClient() *IPAPIClient {
	return &IPAPIClient{}
}

func (c *IPAPIClient) GetCountryByIP(ip string) (*dto.IPInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutSeconds*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://ipinfo.io/"+ip, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get IP info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		log.Printf(
			"[IPClient][GetCountryByIP] Failed to get country for IP [%s] with response code [%s]\n",
			ip, resp.Status,
		)
	}

	var info *dto.IPInfo
	if err = json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode IP info: %w", err)
	}
	return info, nil
}
