package csvrepository

type config struct {
	CsvStorageDirectory string
}

//nolint:revive // no sense in export here
func NewConfig(csvStorageDirectory string) *config {
	return &config{
		CsvStorageDirectory: csvStorageDirectory,
	}
}
