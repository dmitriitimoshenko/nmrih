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

func (s *Service) Save(csvBytes []byte) error {
	// Ensure the directory exists
	if err := os.MkdirAll(CSVStorageDirectory, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Define the file name
	fileName := fmt.Sprintf("logs_%s.csv", time.Now().Format("20060102_150405"))

	// Define the file path
	filePath := filepath.Join(CSVStorageDirectory, fileName)

	// Write the CSV data to the file
	if err := os.WriteFile(filePath, csvBytes, 0644); err != nil {
		return fmt.Errorf("failed to write CSV file: %w", err)
	}

	return nil
}
