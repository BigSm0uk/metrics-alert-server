package config

import (
	"flag"
	"log"

	"github.com/ilyakaznacheev/cleanenv"
	"go.uber.org/zap"
)

type Config struct {
	Env    string     `yaml:"env" required:"true"`
	Logger zap.Config `yaml:"logger" required:"true"`
}

func MustLoadConfig() *Config {
	cfg := &Config{}
	path := flag.String("config", "", "path to config file")
	flag.Parse()

	if *path == "" {
		log.Fatal("config file is required")
	}

	if err := cleanenv.ReadConfig(*path, cfg); err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	return cfg
}
