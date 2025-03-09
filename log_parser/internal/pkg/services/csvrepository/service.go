package csvrepository

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const CSVStorageDirectory = "../data"

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Save(csvBytes []byte, requestTimeStamp time.Time) error {
	// Ensure the directory exists
	if err := os.MkdirAll(CSVStorageDirectory, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Define the file name
	fileName := fmt.Sprintf("logs_%s.csv", requestTimeStamp.Format("2006-01-02_15:04:05"))

	// Define the file path
	filePath := filepath.Join(CSVStorageDirectory, fileName)

	// Write the CSV data to the file
	if err := os.WriteFile(filePath, csvBytes, 0644); err != nil {
		return fmt.Errorf("failed to write CSV file: %w", err)
	}

	return nil
}

func (s *Service) GetLastSavedDate() (*time.Time, error) {
	files, err := os.ReadDir(CSVStorageDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}
	if len(files) == 0 {
		return nil, nil
	}

	var lastTime time.Time
	for _, file := range files {
		name := file.Name()
		// logs_2006-01-02_15:04:05.csv
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
