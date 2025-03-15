package ipapiclient

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/dto"
)

type IPAPIClient struct{}

func NewIPAPIClient() *IPAPIClient {
	return &IPAPIClient{}
}

func (c *IPAPIClient) GetCountryByIP(ip string) (*dto.IPInfo, error) {
	resp, err := http.Get("http://ipinfo.io/" + ip)
	if err != nil {
		return nil, fmt.Errorf("failed to get IP info: %w", err)
	}
	defer resp.Body.Close()

	var info *dto.IPInfo
	err = json.NewDecoder(resp.Body).Decode(&info)
	return info, err
}
