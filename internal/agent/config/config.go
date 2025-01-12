package config

import "time"

type Config struct {
	ServerAddress   string
	RetryAttempts   int
	PollingInterval time.Duration
	SendingInterval time.Duration
}
