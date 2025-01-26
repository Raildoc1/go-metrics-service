package config

import (
	"errors"
	"flag"
	"fmt"
	commonConfig "go-metrics-service/cmd/common/config"
	agent "go-metrics-service/internal/agent/config"
	"os"
	"strconv"
	"time"
)

const (
	sendingIntervalSecondsFlag = "r"
	sendingIntervalSecondsEnv  = "REPORT_INTERVAL"
	pollingIntervalSecondsFlag = "p"
	pollingIntervalSecondsEnv  = "POLL_INTERVAL"
	rateLimitFlag              = "l"
	rateLimitEnv               = "RATE_LIMIT"
)

const (
	defaultSendingIntervalSeconds = 10
	defaultPollingIntervalSeconds = 2
	defaultRateLimit              = 2
)

var defaultRetryAttempts = []time.Duration{time.Second, 3 * time.Second, 5 * time.Second}

type Config struct {
	Agent      agent.Config
	Production bool
}

func Load() (Config, error) {
	serverAddress := flag.String(
		commonConfig.ServerAddressFlag,
		commonConfig.DefaultServerAddress,
		"Server address host:port",
	)

	sendingIntervalSeconds := flag.Int(
		sendingIntervalSecondsFlag,
		defaultSendingIntervalSeconds,
		"Metrics sending frequency in seconds",
	)

	pollingIntervalSeconds := flag.Int(
		pollingIntervalSecondsFlag,
		defaultPollingIntervalSeconds,
		"Metrics polling frequency in seconds",
	)

	rateLimit := flag.Int(
		rateLimitFlag,
		defaultRateLimit,
		"Outgoing requests rate limit",
	)

	sha256Key := flag.String(
		commonConfig.SHA256KeyFlag,
		"",
		"SHA256 key",
	)

	flag.Parse()

	if *sendingIntervalSeconds <= 0 {
		return Config{}, errors.New("sending frequency must be greater than zero")
	}

	if *pollingIntervalSeconds <= 0 {
		return Config{}, errors.New("polling frequency must be greater than zero")
	}

	if valStr, ok := os.LookupEnv(commonConfig.ServerAddressEnv); ok {
		*serverAddress = valStr
	}

	if valStr, ok := os.LookupEnv(sendingIntervalSecondsEnv); ok {
		val, err := strconv.Atoi(valStr)
		if err != nil {
			return Config{}, fmt.Errorf("%w: '%s' env variable parsing failed", err, sendingIntervalSecondsEnv)
		}
		*sendingIntervalSeconds = val
	}

	if valStr, ok := os.LookupEnv(pollingIntervalSecondsEnv); ok {
		val, err := strconv.Atoi(valStr)
		if err != nil {
			return Config{}, fmt.Errorf("%w: '%s' env variable parsing failed", err, pollingIntervalSecondsEnv)
		}
		*pollingIntervalSeconds = val
	}

	if valStr, ok := os.LookupEnv(rateLimitEnv); ok {
		val, err := strconv.Atoi(valStr)
		if err != nil {
			return Config{}, fmt.Errorf("%w: '%s' env variable parsing failed", err, rateLimitEnv)
		}
		*rateLimit = val
	}

	if valStr, ok := os.LookupEnv(commonConfig.SHA256KeyEnv); ok {
		*sha256Key = valStr
	}

	if *sendingIntervalSeconds <= 0 {
		return Config{}, errors.New("sending frequency must be greater than zero")
	}

	if *pollingIntervalSeconds <= 0 {
		return Config{}, errors.New("polling frequency must be greater than zero")
	}

	pollingFreq := time.Duration(*pollingIntervalSeconds) * time.Second
	sendingFreq := time.Duration(*sendingIntervalSeconds) * time.Second

	return Config{
		Agent: agent.Config{
			ServerAddress:   *serverAddress,
			SHA256Key:       *sha256Key,
			SendingInterval: sendingFreq,
			PollingInterval: pollingFreq,
			RetryAttempts:   defaultRetryAttempts,
			RateLimit:       *rateLimit,
		},
		Production: false,
	}, nil
}
