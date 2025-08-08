package config

import "time"

type RedisConfig struct {
	Addr       string
	Password   string
	DB         int
	DefaultTTL time.Duration
}

func NewRedisConfig(addr, password string, db int, defaultTTL time.Duration) *RedisConfig {
	return &RedisConfig{
		Addr:       addr,
		Password:   password,
		DB:         db,
		DefaultTTL: defaultTTL,
	}
}
