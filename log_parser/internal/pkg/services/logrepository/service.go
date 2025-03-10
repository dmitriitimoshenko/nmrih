package logrepository

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	logsStorageDirectory = "/logs/"
	logFileNamePattern   = "l*.log"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) GetLogs() (map[string][]byte, error) {
	pattern := filepath.Join(logsStorageDirectory, logFileNamePattern)
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search for log files: %w", err)
	}

	if files == nil {
		return nil, nil
	}

	// map [ file name ] -> content
	logs := make(map[string][]byte)

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("reading logs error: %s: %w", file, err)
		}
		logs[file] = data
	}

	return logs, nil
}
