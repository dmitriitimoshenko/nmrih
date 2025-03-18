package logrepository

type Config struct {
	LogFilesPattern string
}

func NewConfig(logFilesPattern string) *Config {
	return &Config{
		LogFilesPattern: logFilesPattern,
	}
}
