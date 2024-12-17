package config

import (
	"flag"
	"fmt"
	common "go-metrics-service/cmd/common/config"
	"go-metrics-service/internal/server"
	"os"
	"strconv"
	"time"
)

const (
	FileStoragePathFlag = "f"
	FileStoragePathEnv  = "FILE_STORAGE_PATH"
	StoreIntervalFlag   = "i"
	StoreIntervalEnv    = "STORE_INTERVAL"
	RestoreFlag         = "r"
	RestoreEnv          = "RESTORE"
)

const (
	DefaultFileStoragePath       = "./data.gz"
	DefaultServerShutdownTimeout = 5
	DefaultStoreInterval         = 300
	DefaultRestore               = true
)

type Config struct {
	Server     server.Config
	Production bool
}

func Load() (Config, error) {
	serverAddress := flag.String(
		common.ServerAddressFlag,
		common.DefaultServerAddress,
		"Server address host:port",
	)

	fileStoragePath := flag.String(
		FileStoragePathFlag,
		DefaultFileStoragePath,
		"File path",
	)

	storeInterval := flag.Int(
		StoreIntervalFlag,
		DefaultStoreInterval,
		"Store interval in seconds",
	)

	needRestore := flag.Bool(
		RestoreFlag,
		DefaultRestore,
		"Restore true/false",
	)

	flag.Parse()

	if valStr, ok := os.LookupEnv(common.ServerAddressEnv); ok {
		*serverAddress = valStr
	}

	if valStr, ok := os.LookupEnv(FileStoragePathEnv); ok {
		*fileStoragePath = valStr
	}

	if valStr, ok := os.LookupEnv(StoreIntervalEnv); ok {
		val, err := strconv.Atoi(valStr)
		if err != nil {
			return Config{}, fmt.Errorf("failed to parse '%s' env: %w", StoreIntervalEnv, err)
		}
		*storeInterval = val
	}

	if valStr, ok := os.LookupEnv(RestoreEnv); ok {
		val, err := strconv.ParseBool(valStr)
		if err != nil {
			return Config{}, fmt.Errorf("failed to parse '%s' env: %w", RestoreEnv, err)
		}
		*needRestore = val
	}

	return Config{
		Server: server.Config{
			ServerAddress:   *serverAddress,
			NeedRestore:     *needRestore,
			FilePath:        *fileStoragePath,
			StoreInterval:   time.Duration(*storeInterval) * time.Second,
			ShutdownTimeout: DefaultServerShutdownTimeout * time.Second,
		},
	}, nil
}
