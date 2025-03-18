package logrepository

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type Service struct {
	config Config
}

func NewService(config Config) *Service {
	return &Service{config: config}
}

func (s *Service) GetLogs() (map[string][]byte, error) {
	files, err := filepath.Glob(s.config.LogFilesPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search for log files: %w", err)
	}

	if files == nil {
		return nil, nil
	}

	// map [ file name ] -> content
	logs := make(map[string][]byte)

	log.Println(files)

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("reading logs error: %s: %w", file, err)
		}
		logs[file] = data
	}

	return logs, nil
}
