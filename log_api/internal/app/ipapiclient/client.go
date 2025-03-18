package ipapiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/dto"
)

type IPAPIClient struct{}

func NewIPAPIClient() *IPAPIClient {
	return &IPAPIClient{}
}

func (c *IPAPIClient) GetCountryByIP(ip string) (*dto.IPInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
	var info *dto.IPInfo
	err = json.NewDecoder(resp.Body).Decode(&info)
	return info, err
}
