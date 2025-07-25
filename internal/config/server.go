package config

import (
	"flag"
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
	"go.uber.org/zap"

	S "github.com/bigsm0uk/metrics-alert-server/internal/config/storage"
)

type ServerConfig struct {
	Addr    string          `yaml:"port" env-default:":8080"`
	Storage S.StorageConfig `yaml:"storage" required:"true"`
	Logger  zap.Config      `yaml:"logger" required:"true"`
}

func LoadServerConfig() (*ServerConfig, error) {
	cfg := &ServerConfig{}
	path := flag.String("config", "./config/config.dev.yaml", "path to config file")
	flag.Parse()

	if *path == "" {
		return nil, fmt.Errorf("config file is required")
	}

	if err := cleanenv.ReadConfig(*path, cfg); err != nil {
		return nil, fmt.Errorf("failed to read config: %v", err)
	}

	return cfg, nil
}
