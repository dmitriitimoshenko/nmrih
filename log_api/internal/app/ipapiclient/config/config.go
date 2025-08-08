package config

type IPAPIClientConfig struct {
	IPInfoAPIToken string
}

func NewIPAPIClientConfig(IPInfoAPIToken string) *IPAPIClientConfig {
	return &IPAPIClientConfig{
		IPInfoAPIToken: IPInfoAPIToken,
	}
}
