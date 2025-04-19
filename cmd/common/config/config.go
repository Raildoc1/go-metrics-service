// Package config determines flags, envs, constants and config structs
package config

import (
	"encoding/json"
	"fmt"
	"os"
)

const (
	ServerAddressFlag = "a"
	ServerAddressEnv  = "ADDRESS"
	ServerAddressJSON = "address"
	SHA256KeyFlag     = "k"
	SHA256KeyEnv      = "KEY"
	ConfigFlag        = "c"
	ConfigEnv         = "CONFIG"
)

const (
	DefaultServerAddress = "localhost:8080"
)

func GetRawJSON(path string) (map[string]any, error) {
	jsonCfgBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read '%s' config file: %w", path, err)
	}
	rawJson := make(map[string]any)
	err = json.Unmarshal(jsonCfgBytes, &rawJson)
	if err != nil {
		return nil, fmt.Errorf("failed to parse '%s' config file: %w", path, err)
	}
	return rawJson, nil
}
