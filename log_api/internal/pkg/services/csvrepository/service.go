package csvrepository

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const csvStorageDirectory = "/data"

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Save(csvBytes []byte, requestTimeStamp time.Time) error {
	if err := os.MkdirAll(csvStorageDirectory, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	fileName := fmt.Sprintf("logs_%s.csv", requestTimeStamp.Format("2006-01-02_15:04:05"))
	filePath := filepath.Join(csvStorageDirectory, fileName)

	if err := os.WriteFile(filePath, csvBytes, 0o600); err != nil {
		return fmt.Errorf("failed to write CSV file: %w", err)
	}

	return nil
}

func (s *Service) GetLastSavedDate() (*time.Time, error) {
	files, err := os.ReadDir(csvStorageDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}
	if len(files) == 0 {
		return nil, nil
	}

	var lastTime time.Time
	for _, file := range files {
		name := file.Name()
		// example: logs_2006-01-02_15:04:05.csv
		dateString := name[5:24]
		parsedTime, err := time.Parse("2006-01-02_15:04:05", dateString)
		if err != nil {
			return nil, fmt.Errorf("failed to parse time: %w", err)
		}
		if parsedTime.After(lastTime) {
			lastTime = parsedTime
		}
	}

	return &lastTime, nil
}

func (s *Service) GetAllCSVData() ([]byte, error) {
	files, err := os.ReadDir(csvStorageDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var combined bytes.Buffer
	firstFile := true

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".csv") {
			filePath := filepath.Join(csvStorageDirectory, file.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
			}

			if !firstFile {
				if pos := bytes.IndexByte(content, '\n'); pos != -1 && pos+1 < len(content) {
					content = content[pos+1:]
				}
			}
			combined.Write(content)
			combined.WriteString("\n")
			firstFile = false
		}
	}

	return combined.Bytes(), nil
}
