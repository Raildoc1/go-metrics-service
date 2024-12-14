package config

import (
	"flag"
	commonConfig "go-metrics-service/cmd/common/config"
	"os"
)

type Config struct {
	ServerAddress string
	Production    bool
}

func Load() (Config, error) {
	serverAddress := flag.String(
		commonConfig.ServerAddressFlag,
		commonConfig.DefaultServerAddress,
		"Server address host:port",
	)

	flag.Parse()

	if valStr, ok := os.LookupEnv(commonConfig.ServerAddressEnv); ok {
		*serverAddress = valStr
	}

	return Config{
		ServerAddress: *serverAddress,
	}, nil
}
