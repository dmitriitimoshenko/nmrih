package logrepository

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type Service struct {
	config config
}

func NewService(config config) *Service {
	return &Service{config: config}
}

func (s *Service) GetLogs() (map[string][]byte, error) {
	pattern := filepath.Join(s.config.LogDirectory, s.config.LogFilesPattern)
	files, err := filepath.Glob(pattern)
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
