package dto

type TopCountry struct {
	Country          string `json:"country"`
	ConnectionsCount int    `json:"connections_count"`
}

type TopCountriesList []TopCountry

type TopCountriesPercentage struct {
	Country    string  `json:"country"`
	Percentage float64 `json:"percentage"`
}

type TopCountriesPercentageList []TopCountriesPercentage
