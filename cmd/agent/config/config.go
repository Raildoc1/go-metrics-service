package config

import (
	"errors"
	"flag"
	"go-metrics-service/cmd/common"
	commonConfig "go-metrics-service/cmd/common/config"
	agent "go-metrics-service/internal/agent/config"
	"os"
	"strconv"
	"time"
)

const (
	sendingFreqSecondsFlag = "r"
	sendingFreqSecondsEnv  = "REPORT_INTERVAL"
	pollingFreqSecondsFlag = "p"
	pollingFreqSecondsEnv  = "POLL_INTERVAL"
)

const (
	defaultSendingIntervalSeconds = 10
	defaultPollingIntervalSeconds = 2
)

type Config struct {
	Agent agent.Config
}

func Load() (Config, error) {
	serverAddress := &common.ServerAddress{
		Host: commonConfig.DefaultServerHost,
		Port: commonConfig.DefaultServerPort,
	}

	flag.Var(serverAddress, commonConfig.ServerAddressFlag, "Server address host:port")
	
	sendingIntervalSeconds := flag.Int(
		sendingFreqSecondsFlag,
		defaultSendingIntervalSeconds,
		"Metrics sending frequency in seconds",
	)

	pollingIntervalSeconds := flag.Int(
		pollingFreqSecondsFlag,
		defaultPollingIntervalSeconds,
		"Metrics polling frequency in seconds",
	)

	flag.Parse()

	if *sendingIntervalSeconds <= 0 {
		return Config{}, errors.New("sending frequency must be greater than zero")
	}

	if *pollingIntervalSeconds <= 0 {
		return Config{}, errors.New("polling frequency must be greater than zero")
	}

	if valStr, ok := os.LookupEnv(commonConfig.ServerAddressEnv); ok {
		err := serverAddress.Set(valStr)
		if err != nil {
			return Config{}, err
		}
	}

	if valStr, ok := os.LookupEnv(sendingFreqSecondsEnv); ok {
		val, err := strconv.Atoi(valStr)
		if err != nil {
			return Config{}, err
		}
		*sendingIntervalSeconds = val
	}

	if valStr, ok := os.LookupEnv(pollingFreqSecondsEnv); ok {
		val, err := strconv.Atoi(valStr)
		if err != nil {
			return Config{}, err
		}
		*pollingIntervalSeconds = val
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
		agent.Config{
			ServerAddress: serverAddress.String(),
			SendingFreq:   sendingFreq,
			PollingFreq:   pollingFreq,
		},
	}, nil
}
