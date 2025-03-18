package csvrepository

type Config struct {
	CsvStorageDirectory string
}

func NewConfig(csvStorageDirectory string) *Config {
	return &Config{
		CsvStorageDirectory: csvStorageDirectory,
	}
}
