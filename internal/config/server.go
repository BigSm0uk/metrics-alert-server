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
	path := flag.String("config", "config/config.dev.yaml", "path to config file")
	flag.Parse()

	if err := cleanenv.ReadConfig(*path, cfg); err != nil {
		fmt.Printf("failed to read config: %v", err)
		cfg = InitDefaultConfig()
	}

	return cfg, nil
}
func InitDefaultConfig() *ServerConfig {
	return &ServerConfig{
		Addr: ":8080",
		Storage: S.StorageConfig{
			Type: "mem",
		},
		Logger: zap.NewDevelopmentConfig(),
	}
}
