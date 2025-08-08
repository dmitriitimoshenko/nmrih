package config

type IPAPIClientConfig struct {
	IPInfoAPIToken string
}

func NewIPAPIClientConfig(iPInfoAPIToken string) *IPAPIClientConfig {
	return &IPAPIClientConfig{
		IPInfoAPIToken: iPInfoAPIToken,
	}
}
