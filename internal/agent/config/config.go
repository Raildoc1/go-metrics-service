package config

import "time"

type Config struct {
	ServerAddress   string
	SHA256Key       string
	RetryAttempts   []time.Duration
	RateLimit       int
	PollingInterval time.Duration
	SendingInterval time.Duration
}
