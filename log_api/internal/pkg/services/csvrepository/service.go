package csvrepository

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	fileNameDateLayout   = "2006-01-02_15:04:05"
	fileNamePrefixLength = len("logs_")
	fileNameDateLength   = len(fileNameDateLayout)
)

type Service struct {
	config config
}

func NewService(config config) *Service {
	return &Service{config: config}
}

func (s *Service) GetLastSavedDate() (*time.Time, error) {
	files, err := os.ReadDir(s.config.CsvStorageDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var (
		lastTime time.Time
		found    bool
	)
	for _, file := range files {
		// example: logs_2006-01-02_15:04:05.csv
		name := file.Name()
		if file.IsDir() ||
			!strings.HasSuffix(name, ".csv") ||
			len(name) < fileNamePrefixLength+fileNameDateLength {
			continue
		}
		dateString := name[fileNamePrefixLength : fileNamePrefixLength+fileNameDateLength]
		parsedTime, err := time.Parse(fileNameDateLayout, dateString)
		if err != nil {
			return nil, fmt.Errorf("failed to parse time: %w", err)
		}
		if !found || parsedTime.After(lastTime) {
			lastTime = parsedTime
			found = true
		}
	}

	if !found {
		return nil, nil
	}
	return &lastTime, nil
}

func (s *Service) Save(csvBytes []byte, requestTimeStamp time.Time) error {
	if err := os.MkdirAll(s.config.CsvStorageDirectory, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	fileName := fmt.Sprintf("logs_%s.csv", requestTimeStamp.Format(fileNameDateLayout))
	filePath := filepath.Join(s.config.CsvStorageDirectory, fileName)

	if err := os.WriteFile(filePath, csvBytes, 0o600); err != nil {
		return fmt.Errorf("failed to write CSV file: %w", err)
	}

	return nil
}

func (s *Service) GetAllCSVData() ([]byte, error) {
	files, err := os.ReadDir(s.config.CsvStorageDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var combined bytes.Buffer
	firstFile := true

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".csv") {
			filePath := filepath.Join(s.config.CsvStorageDirectory, file.Name())
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
