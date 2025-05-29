package logrepository

type config struct {
	LogDirectory    string
	LogFilesPattern string
}

//nolint:revive // no sense in export here
func NewConfig(
	logDirectory string,
	logFilesPattern string,
) *config {
	return &config{
		LogDirectory:    logDirectory,
		LogFilesPattern: logFilesPattern,
	}
}
