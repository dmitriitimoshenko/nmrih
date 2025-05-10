package csvrepository

type config struct {
	CsvStorageDirectory string
}

func NewConfig(csvStorageDirectory string) *config {
	return &config{
		CsvStorageDirectory: csvStorageDirectory,
	}
}
