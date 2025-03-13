package config

type A2SClientConfig struct {
	Host string
	Port int
}

func NewA2SClientConfig(host string, port int) *A2SClientConfig {
	return &A2SClientConfig{
		Host: host,
		Port: port,
	}
}
