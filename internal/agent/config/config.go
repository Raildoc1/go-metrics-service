// Package config determines flags, envs, constants and config structs
package config

import (
	"go-metrics-service/internal/agent/sender/driver"
	"time"
)

type Config struct {
	ServerAddress   string
	SHA256Key       string
	RetryAttempts   []time.Duration
	RateLimit       int
	PollingInterval time.Duration
	SendingInterval time.Duration
	RSAPublicKeyPem []byte
	GRPC            *driver.GRPCConfig
}
