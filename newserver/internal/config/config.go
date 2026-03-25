package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	GrpcPort     int    `json:"grpc_port"`
	GetEndpoint  string `json:"getEndpoint"`
	PostEndpoint string `json:"postEndpoint"`
}

var Cfg *Config

func Load() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(homeDir, ".scurrier", "config", "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	Cfg = &Config{}
	return json.Unmarshal(data, Cfg)
}
