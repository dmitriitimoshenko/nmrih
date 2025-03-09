package logrepository

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	logsDirectory      = "../logs/"
	logFileNamePattern = "l*.log"
)

type Service struct{}

type Filter struct {
	DateFrom *time.Time
	DateTo   *time.Time
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) GetLogs(fiter *Filter) (map[string][]byte, error) {
	pattern := filepath.Join(logsDirectory, logFileNamePattern)
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
