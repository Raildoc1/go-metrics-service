package config

import (
	"flag"
	"fmt"
	common "go-metrics-service/cmd/common/config"
	"go-metrics-service/internal/server"
	"go-metrics-service/internal/server/data/backupmemstorage"
	"go-metrics-service/internal/server/database"
	"os"
	"strconv"
	"time"
)

const (
	fileStoragePathFlag    = "f"
	fileStoragePathEnv     = "FILE_STORAGE_PATH"
	storeIntervalFlag      = "i"
	storeIntervalEnv       = "STORE_INTERVAL"
	restoreFlag            = "r"
	restoreEnv             = "RESTORE"
	dbConnectionStringFlag = "d"
	dbConnectionStringEnv  = "DATABASE_DSN"
)

const (
	defaultFileStoragePath       = "./localstorage/data.gz"
	defaultServerShutdownTimeout = 5
	defaultStoreInterval         = 300
	defaultRestore               = true
)

type Config struct {
	Database         database.Config
	BackupMemStorage backupmemstorage.Config
	Server           server.Config
	Production       bool
}

func Load() (Config, error) {
	serverAddress := flag.String(
		common.ServerAddressFlag,
		common.DefaultServerAddress,
		"Server address host:port",
	)

	fileStoragePath := flag.String(
		fileStoragePathFlag,
		defaultFileStoragePath,
		"File path",
	)

	storeInterval := flag.Int(
		storeIntervalFlag,
		defaultStoreInterval,
		"Store interval in seconds",
	)

	needRestore := flag.Bool(
		restoreFlag,
		defaultRestore,
		"Restore true/false",
	)

	dbConnectionString := flag.String(
		dbConnectionStringFlag,
		"",
		"Database connection string",
	)

	flag.Parse()

	if valStr, ok := os.LookupEnv(common.ServerAddressEnv); ok {
		*serverAddress = valStr
	}

	if valStr, ok := os.LookupEnv(fileStoragePathEnv); ok {
		*fileStoragePath = valStr
	}

	if valStr, ok := os.LookupEnv(storeIntervalEnv); ok {
		val, err := strconv.Atoi(valStr)
		if err != nil {
			return Config{}, fmt.Errorf("failed to parse '%s' env: %w", storeIntervalEnv, err)
		}
		*storeInterval = val
	}

	if valStr, ok := os.LookupEnv(restoreEnv); ok {
		val, err := strconv.ParseBool(valStr)
		if err != nil {
			return Config{}, fmt.Errorf("failed to parse '%s' env: %w", restoreEnv, err)
		}
		*needRestore = val
	}

	if valStr, ok := os.LookupEnv(dbConnectionStringEnv); ok {
		*dbConnectionString = valStr
	}

	return Config{
		Database: database.Config{
			ConnectionString: *dbConnectionString,
		},
		BackupMemStorage: backupmemstorage.Config{
			Backup: backupmemstorage.BackupConfig{
				FilePath:      *fileStoragePath,
				StoreInterval: time.Duration(*storeInterval) * time.Second,
			},
			NeedRestore: *needRestore,
		},
		Server: server.Config{
			ServerAddress:   *serverAddress,
			ShutdownTimeout: defaultServerShutdownTimeout * time.Second,
		},
	}, nil
}
