package logrepository

type config struct {
	LogDirectory    string
	LogFilesPattern string
}

func NewConfig(
	logDirectory string,
	logFilesPattern string,
) *config {
	return &config{
		LogDirectory:    logDirectory,
		LogFilesPattern: logFilesPattern,
	}
}
