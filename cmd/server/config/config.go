// Package config determines flags, envs, constants and config structs
package config

import (
	"errors"
	"flag"
	"fmt"
	common "go-metrics-service/cmd/common/config"
	"go-metrics-service/cmd/common/config/flagtypes"
	"go-metrics-service/internal/server"
	"go-metrics-service/internal/server/data/storages/backupmemstorage"
	"go-metrics-service/internal/server/database"
	"os"
	"strconv"
	"time"
)

const (
	fileStoragePathFlag    = "f"
	fileStoragePathEnv     = "FILE_STORAGE_PATH"
	fileStoragePathJSON    = "store_file"
	storeIntervalFlag      = "i"
	storeIntervalEnv       = "STORE_INTERVAL"
	storeIntervalJSON      = "store_interval"
	needRestoreFlag        = "r"
	needRestoreEnv         = "RESTORE"
	needRestoreJSON        = "restore"
	dbConnectionStringFlag = "d"
	dbConnectionStringEnv  = "DATABASE_DSN"
	dbConnectionStringJSON = "database_dsn"
	rsaPrivateKeyFileFlag  = "crypto-key"
	rsaPrivateKeyFileEnv   = "CRYPTO_KEY"
	rsaPrivateKeyFileJSON  = "crypto_key"
)

const (
	defaultFileStoragePath       = "./localstorage/data.gz"
	defaultServerShutdownTimeout = 5 * time.Second
	defaultAppShutdownTimeout    = 10 * time.Second
	defaultStoreInterval         = 300 * time.Second
	defaultNeedRestore           = true
	defaultDBConnectionString    = ""
	defaultSHA256Key             = ""
	defaultRSAPrivateKeyFilePath = ""
)

var defaultRetryAttempts = []time.Duration{time.Second, 3 * time.Second, 5 * time.Second}

type Config struct {
	SHA256Key        string
	Database         database.Config
	BackupMemStorage backupmemstorage.Config
	Server           server.Config
	ShutdownTimeout  time.Duration
	Production       bool
	RSAPrivateKeyPem string
}

func Load() (Config, error) {

	serverAddress := common.DefaultServerAddress
	fileStoragePath := defaultFileStoragePath
	storeInterval := defaultStoreInterval
	needRestore := defaultNeedRestore
	dbConnectionString := defaultDBConnectionString
	sha256Key := defaultSHA256Key
	rsaPrivateKeyFilePath := defaultRSAPrivateKeyFilePath

	// Flags Definition.

	configFlagVal := flagtypes.NewString()
	flag.Var(configFlagVal, common.ConfigFlag, "JSON config path")

	serverAddressFlagVal := flagtypes.NewString()
	flag.Var(serverAddressFlagVal, common.ServerAddressFlag, "Server address host:port")

	fileStoragePathFlagVal := flagtypes.NewString()
	flag.Var(fileStoragePathFlagVal, fileStoragePathFlag, "File storage path")

	storeIntervalFlagVal := flagtypes.NewInt()
	flag.Var(storeIntervalFlagVal, storeIntervalFlag, "Store interval in seconds")

	needRestoreFlagVal := flagtypes.NewBool()
	flag.Var(needRestoreFlagVal, needRestoreFlag, "Need restore true/false")

	dbConnectionStringFlagVal := flagtypes.NewString()
	flag.Var(dbConnectionStringFlagVal, dbConnectionStringFlag, "Database connection string")

	sha256KeyFlagVal := flagtypes.NewString()
	flag.Var(sha256KeyFlagVal, common.SHA256KeyFlag, "SHA256 key")

	rsaPrivateKeyFilePathFlagVal := flagtypes.NewString()
	flag.Var(rsaPrivateKeyFilePathFlagVal, rsaPrivateKeyFileFlag, "RSA private key file path")

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
		if val, ok := rawJSON[needRestoreJSON]; ok {
			needRestore = val.(bool)
		}
		if val, ok := rawJSON[storeIntervalJSON]; ok {
			storeInterval, err = time.ParseDuration(val.(string))
			if err != nil {
				return Config{}, fmt.Errorf("invalid value for store interval: %w", err)
			}
		}
		if val, ok := rawJSON[fileStoragePathJSON]; ok {
			fileStoragePath = val.(string)
		}
		if val, ok := rawJSON[dbConnectionStringJSON]; ok {
			dbConnectionString = val.(string)
		}
		if val, ok := rawJSON[rsaPrivateKeyFileJSON]; ok {
			rsaPrivateKeyFilePath = val.(string)
		}
	}

	// Flags Parse.

	if val, ok := serverAddressFlagVal.Value(); ok {
		serverAddress = val
	}

	if val, ok := fileStoragePathFlagVal.Value(); ok {
		fileStoragePath = val
	}

	if val, ok := storeIntervalFlagVal.Value(); ok {
		storeInterval = time.Duration(val) * time.Second
	}

	if val, ok := needRestoreFlagVal.Value(); ok {
		needRestore = val
	}

	if val, ok := dbConnectionStringFlagVal.Value(); ok {
		dbConnectionString = val
	}

	if val, ok := sha256KeyFlagVal.Value(); ok {
		sha256Key = val
	}

	if val, ok := rsaPrivateKeyFilePathFlagVal.Value(); ok {
		rsaPrivateKeyFilePath = val
	}

	// Environment Variables.

	if valStr, ok := os.LookupEnv(common.ServerAddressEnv); ok {
		serverAddress = valStr
	}

	if valStr, ok := os.LookupEnv(fileStoragePathEnv); ok {
		fileStoragePath = valStr
	}

	if valStr, ok := os.LookupEnv(storeIntervalEnv); ok {
		val, err := strconv.Atoi(valStr)
		if err != nil {
			return Config{}, fmt.Errorf("%w: '%s' env variable parsing failed", err, storeIntervalEnv)
		}
		storeInterval = time.Duration(val) * time.Second
	}

	if valStr, ok := os.LookupEnv(needRestoreEnv); ok {
		val, err := strconv.ParseBool(valStr)
		if err != nil {
			return Config{}, fmt.Errorf("%w: '%s' env variable parsing failed", err, needRestoreEnv)
		}
		needRestore = val
	}

	if valStr, ok := os.LookupEnv(dbConnectionStringEnv); ok {
		dbConnectionString = valStr
	}

	if valStr, ok := os.LookupEnv(common.SHA256KeyEnv); ok {
		sha256Key = valStr
	}

	if valStr, ok := os.LookupEnv(rsaPrivateKeyFileEnv); ok {
		rsaPrivateKeyFilePath = valStr
	}

	// Validation.

	if storeInterval < time.Duration(0) {
		return Config{}, errors.New("store internal must be greater than zero")
	}

	// RSA pem file reading.

	var rsaPrivateKeyPem []byte = nil

	if rsaPrivateKeyFilePath != "" {
		prv, err := os.ReadFile(rsaPrivateKeyFilePath)
		if err != nil {
			return Config{}, fmt.Errorf("failed to read file '%s': %w", rsaPrivateKeyFilePath, err)
		}
		rsaPrivateKeyPem = prv
	}

	return Config{
		Database: database.Config{
			ConnectionString: dbConnectionString,
			RetryAttempts:    defaultRetryAttempts,
		},
		BackupMemStorage: backupmemstorage.Config{
			Backup: backupmemstorage.BackupConfig{
				FilePath:      fileStoragePath,
				StoreInterval: storeInterval,
			},
			NeedRestore: needRestore,
		},
		Server: server.Config{
			ServerAddress:   serverAddress,
			ShutdownTimeout: defaultServerShutdownTimeout,
		},
		SHA256Key:        sha256Key,
		ShutdownTimeout:  defaultAppShutdownTimeout,
		RSAPrivateKeyPem: string(rsaPrivateKeyPem),
	}, nil
}
