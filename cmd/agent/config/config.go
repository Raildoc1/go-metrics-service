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

type Config struct {
	Agent agent.Config
}

func Load() (Config, error) {
	serverAddress := &common.ServerAddress{
		Host: "localhost",
		Port: 8080,
	}

	flag.Var(serverAddress, commonConfig.ServerAddressFlag, "Server address host:port")
	sendingFreqSeconds := flag.Int(sendingFreqSecondsFlag, 10, "Metrics sending frequency in seconds")
	pollingFreqSeconds := flag.Int(pollingFreqSecondsFlag, 2, "Metrics polling frequency in seconds")

	flag.Parse()

	if *sendingFreqSeconds <= 0 {
		return Config{}, errors.New("sending frequency must be greater than zero")
	}

	if *pollingFreqSeconds <= 0 {
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
		*sendingFreqSeconds = val
	}

	if valStr, ok := os.LookupEnv(pollingFreqSecondsEnv); ok {
		val, err := strconv.Atoi(valStr)
		if err != nil {
			return Config{}, err
		}
		*pollingFreqSeconds = val
	}

	if *sendingFreqSeconds <= 0 {
		return Config{}, errors.New("sending frequency must be greater than zero")
	}

	if *pollingFreqSeconds <= 0 {
		return Config{}, errors.New("polling frequency must be greater than zero")
	}

	pollingFreq := time.Duration(*pollingFreqSeconds) * time.Second
	sendingFreq := time.Duration(*sendingFreqSeconds) * time.Second

	return Config{
		agent.Config{
			ServerAddress: serverAddress.String(),
			SendingFreq:   sendingFreq,
			PollingFreq:   pollingFreq,
		},
	}, nil
}
