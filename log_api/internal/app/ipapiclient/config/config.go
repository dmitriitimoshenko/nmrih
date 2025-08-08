package config

type IPAPIClientConfig struct {
	IpInfoAPIToken string
}

func NewIPAPIClientConfig(ipInfoAPIToken string) *IPAPIClientConfig {
	return &IPAPIClientConfig{
		IpInfoAPIToken: ipInfoAPIToken,
	}
}
