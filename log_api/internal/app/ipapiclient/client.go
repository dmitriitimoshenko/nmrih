package ipapiclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/dto"
)

type getCountriesByIPsPayload struct {
	Query    string `json:"query"`
	Fields   string `json:"fields"`
	Language string `json:"lang"`
}

type getCountriesByIPsResponse struct {
	Query   string `json:"query"`
	Country string `json:"country"`
}

type IPAPIClient struct{}

func NewIPAPIClient() *IPAPIClient {
	return &IPAPIClient{}
}

func (c *IPAPIClient) GetCountriesByIPs(ips []string) (dto.IPInfo, error) {
	fmt.Println("IPs: ", ips)

	var payload []getCountriesByIPsPayload
	for _, ip := range ips {
		payload = append(payload, getCountriesByIPsPayload{
			Query:    ip,
			Fields:   "country,query",
			Language: "en",
		})
	}

	encodedPyaload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshall data [%+v]: %w", payload, err)
	}

	fmt.Println(string(encodedPyaload))

	reader := bytes.NewReader(encodedPyaload)
	request, err := http.NewRequest(http.MethodPost, "http://ip-api.com/batch", reader)
	if err != nil {
		return nil, fmt.Errorf("failed to compose request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "Batch-Example/1.0")

	client := http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status returned is not 200 but %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var ipCountryList []getCountriesByIPsResponse
	if err := json.Unmarshal(body, &ipCountryList); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	ipInfo := make(dto.IPInfo, len(ipCountryList))
	for _, ipCountry := range ipCountryList {
		ipInfo[ipCountry.Query] = ipCountry.Country
	}

	return ipInfo, nil
}
