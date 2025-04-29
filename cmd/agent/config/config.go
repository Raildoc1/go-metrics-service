// Package config determines flags, envs, constants and config structs
package config

import (
	"errors"
	"flag"
	"fmt"
	common "go-metrics-service/cmd/common/config"
	"go-metrics-service/cmd/common/config/flagtypes"
	agent "go-metrics-service/internal/agent/config"
	"os"
	"strconv"
	"time"
)

const (
	sendingIntervalSecondsFlag = "r"
	sendingIntervalSecondsEnv  = "REPORT_INTERVAL"
	sendingIntervalSecondsJSON = "report_interval"
	pollingIntervalSecondsFlag = "p"
	pollingIntervalSecondsEnv  = "POLL_INTERVAL"
	pollingIntervalSecondsJSON = "poll_interval"
	rateLimitFlag              = "l"
	rateLimitEnv               = "RATE_LIMIT"
	rsaPublicKeyFileFlag       = "crypto-key"
	rsaPublicKeyFileEnv        = "CRYPTO_KEY"
	rsaPublicKeyFileJSON       = "crypto_key"
)

const (
	defaultSendingInterval = time.Second * 10
	defaultPollingInterval = time.Second * 2
	defaultRateLimit       = 2
	defaultRSAPublicKey    = ""
	defaultSHA256Key       = ""
)

var defaultRetryAttempts = []time.Duration{time.Second, 3 * time.Second, 5 * time.Second}

type Config struct {
	Agent      agent.Config
	Production bool
}

func Load() (Config, error) {

	serverAddress := common.DefaultServerAddress
	sendingInterval := defaultSendingInterval
	pollingInterval := defaultPollingInterval
	rsaPublicKeyFilePath := defaultRSAPublicKey
	sha256Key := defaultSHA256Key
	rateLimit := defaultRateLimit

	// Flags Definition.

	configFlagVal := flagtypes.NewString()
	flag.Var(configFlagVal, common.ConfigFlag, "JSON config path")

	serverAddressFlagVal := flagtypes.NewString()
	flag.Var(serverAddressFlagVal, common.ServerAddressFlag, "Server address host:port")

	sendingIntervalSecondsFlagVal := flagtypes.NewInt()
	flag.Var(sendingIntervalSecondsFlagVal, sendingIntervalSecondsFlag, "Metrics sending frequency in seconds")

	pollingIntervalSecondsFlagVal := flagtypes.NewInt()
	flag.Var(pollingIntervalSecondsFlagVal, pollingIntervalSecondsFlag, "Metrics polling frequency in seconds")

	rateLimitFlagVal := flagtypes.NewInt()
	flag.Var(rateLimitFlagVal, rateLimitFlag, "Outgoing requests rate limit")

	sha256KeyFlagVal := flagtypes.NewString()
	flag.Var(sha256KeyFlagVal, common.SHA256KeyFlag, "SHA256 key")

	rsaPublicKeyFilePathFlagVal := flagtypes.NewString()
	flag.Var(rsaPublicKeyFilePathFlagVal, rsaPublicKeyFileFlag, "RSA public key file path")

	flag.Parse()

	// Config JSON.

	var cfgPath *string = nil

	if val, ok := configFlagVal.Value(); ok {
		cfgPath = &val
	}

	if valStr, ok := os.LookupEnv(common.ConfigEnv); ok {
		cfgPath = &valStr
	}

	if cfgPath != nil {
		rawJSON, err := common.GetRawJSON(*cfgPath)
		if err != nil {
			return Config{}, err
		}
		if val, ok := rawJSON[common.ServerAddressJSON]; ok {
			serverAddress = val.(string)
		}
		if val, ok := rawJSON[sendingIntervalSecondsJSON]; ok {
			sendingInterval, err = time.ParseDuration(val.(string))
			if err != nil {
				return Config{}, fmt.Errorf("invalid value for sending interval: %w", err)
			}
		}
		if val, ok := rawJSON[pollingIntervalSecondsJSON]; ok {
			pollingInterval, err = time.ParseDuration(val.(string))
			if err != nil {
				return Config{}, fmt.Errorf("invalid value for polling interval: %w", err)
			}
		}
		if val, ok := rawJSON[rsaPublicKeyFileJSON]; ok {
			rsaPublicKeyFilePath = val.(string)
		}
	}

	// Flags Parse.

	if val, ok := serverAddressFlagVal.Value(); ok {
		serverAddress = val
	}

	if val, ok := sendingIntervalSecondsFlagVal.Value(); ok {
		sendingInterval = time.Duration(val) * time.Second
	}

	if val, ok := pollingIntervalSecondsFlagVal.Value(); ok {
		pollingInterval = time.Duration(val) * time.Second
	}

	if val, ok := rateLimitFlagVal.Value(); ok {
		rateLimit = val
	}

	if val, ok := sha256KeyFlagVal.Value(); ok {
		sha256Key = val
	}

	if val, ok := rsaPublicKeyFilePathFlagVal.Value(); ok {
		rsaPublicKeyFilePath = val
	}

	// Environment Variables.

	if valStr, ok := os.LookupEnv(common.ServerAddressEnv); ok {
		serverAddress = valStr
	}

	if valStr, ok := os.LookupEnv(sendingIntervalSecondsEnv); ok {
		val, err := strconv.Atoi(valStr)
		if err != nil {
			return Config{}, fmt.Errorf("%w: '%s' env variable parsing failed", err, sendingIntervalSecondsEnv)
		}
		sendingInterval = time.Duration(val) * time.Second
	}

	if valStr, ok := os.LookupEnv(pollingIntervalSecondsEnv); ok {
		val, err := strconv.Atoi(valStr)
		if err != nil {
			return Config{}, fmt.Errorf("%w: '%s' env variable parsing failed", err, pollingIntervalSecondsEnv)
		}
		pollingInterval = time.Duration(val) * time.Second
	}

	if valStr, ok := os.LookupEnv(rateLimitEnv); ok {
		val, err := strconv.Atoi(valStr)
		if err != nil {
			return Config{}, fmt.Errorf("%w: '%s' env variable parsing failed", err, rateLimitEnv)
		}
		rateLimit = val
	}

	if valStr, ok := os.LookupEnv(common.SHA256KeyEnv); ok {
		sha256Key = valStr
	}

	if valStr, ok := os.LookupEnv(rsaPublicKeyFileEnv); ok {
		rsaPublicKeyFilePath = valStr
	}

	// Validation.

	if sendingInterval < time.Duration(0) {
		return Config{}, errors.New("sending frequency must be greater than zero")
	}

	if pollingInterval < time.Duration(0) {
		return Config{}, errors.New("polling frequency must be greater than zero")
	}

	// RSA pem file reading.

	var rsaPublicKeyPem []byte = nil

	if rsaPublicKeyFilePath != "" {
		pub, err := os.ReadFile(rsaPublicKeyFilePath)
		if err != nil {
			return Config{}, fmt.Errorf("failed to read file '%s': %w", rsaPublicKeyFilePath, err)
		}
		rsaPublicKeyPem = pub
	}

	return Config{
		Agent: agent.Config{
			ServerAddress:   serverAddress,
			SHA256Key:       sha256Key,
			SendingInterval: sendingInterval,
			PollingInterval: pollingInterval,
			RetryAttempts:   defaultRetryAttempts,
			RateLimit:       rateLimit,
			RSAPublicKeyPem: rsaPublicKeyPem,
		},
		Production: false,
	}, nil
}
