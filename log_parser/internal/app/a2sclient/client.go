package a2sclient

import (
	"fmt"

	"github.com/dmitriitimoshenko/nmrih/log_parser/internal/app/a2sclient/config"
	"github.com/rumblefrog/go-a2s"
)

func NewA2SClient(config *config.A2SClientConfig) (*a2s.Client, error) {
	serverAddress := config.Host + ":" + fmt.Sprint(config.Port)

	client, err := a2s.NewClient(serverAddress)
	if err != nil {
		return nil, err
	}

	return client, nil
}
