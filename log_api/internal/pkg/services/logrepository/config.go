package logrepository

type Config struct {
	LogDirectory    string
	LogFilesPattern string
}

func NewConfig(
	logDirectory string,
	logFilesPattern string,
) *Config {
	return &Config{
		LogDirectory:    logDirectory,
		LogFilesPattern: logFilesPattern,
	}
}
