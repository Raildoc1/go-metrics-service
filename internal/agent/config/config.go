package config

import "time"

type Config struct {
	ServerAddress   string
	PollingInterval time.Duration
	SendingInterval time.Duration
}
