package config

import (
	"flag"
	"go-metrics-service/cmd/common"
	commonConfig "go-metrics-service/cmd/common/config"
	"os"
)

type Config struct {
	ServerAddress string
}

func Load() (Config, error) {
	serverAddress := &common.ServerAddress{
		Host: "localhost",
		Port: 8080,
	}

	flag.Var(serverAddress, commonConfig.ServerAddressFlag, "Server address host:port")

	flag.Parse()

	if valStr, ok := os.LookupEnv(commonConfig.ServerAddressEnv); ok {
		err := serverAddress.Set(valStr)
		if err != nil {
			return Config{}, err
		}
	}

	return Config{
		ServerAddress: serverAddress.String(),
	}, nil
}
