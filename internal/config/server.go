package config

import (
	"flag"
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"

	S "github.com/bigsm0uk/metrics-alert-server/internal/config/storage"
)

const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
)

type ServerConfig struct {
	Env          string          `yaml:"env" env-default:"development"`
	Addr         string          `yaml:"port" env-default:":8080"`
	Storage      S.StorageConfig `yaml:"storage" required:"true"`
	TemplatePath string          `yaml:"template_path" env-default:"api/templates/metrics.html"`
}

func LoadServerConfig() (*ServerConfig, error) {
	cfg := &ServerConfig{}
	path := flag.String("config", "config/config.dev.yaml", "path to config file")
	flag.Parse()

	if err := cleanenv.ReadConfig(*path, cfg); err != nil {
		fmt.Printf("failed to read config: %v. In development mode. Using default config...", err)
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
		TemplatePath: "api/templates/metrics.html",
		Env:          EnvDevelopment,
	}
}
