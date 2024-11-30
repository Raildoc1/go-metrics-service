package config

import "time"

type Config struct {
	ServerAddress string
	PollingFreq   time.Duration
	SendingFreq   time.Duration
}
