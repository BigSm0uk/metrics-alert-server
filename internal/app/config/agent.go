package config

import (
	"flag"
	"fmt"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
)

type AgentConfig struct {
	Env            string `json:"env" env:"ENV"`
	Addr           string `json:"address" env:"ADDRESS"`
	GRPCAddr       string `json:"grpc_address" env:"GRPC_ADDRESS"`
	ReportInterval uint   `json:"report_interval" env:"REPORT_INTERVAL"`
	PollInterval   uint   `json:"poll_interval" env:"POLL_INTERVAL"`
	RateLimit      uint   `json:"rate_limit" env:"RATE_LIMIT"`
	Key            string `json:"key" env:"KEY"`
	CryptoKey      string `json:"crypto_key" env:"CRYPTO_KEY"`
	ConfigFile     string `json:"-" env:"CONFIG"`
}

func LoadAgentConfig() (*AgentConfig, error) {
	cfg := &AgentConfig{
		Env:            EnvDevelopment,
		Addr:           "localhost:8080",
		ReportInterval: 10,
		PollInterval:   2,
		RateLimit:      1,
		Key:            "1234567890",
	}

	var (
		flagConfig     = flag.String("c", "", "path to config file")
		flagConfigLong = flag.String("config", "", "path to config file")
	)

	flag.StringVar(&cfg.Env, "e", cfg.Env, "environment")
	flag.StringVar(&cfg.Addr, "a", cfg.Addr, "http server address")
	flag.StringVar(&cfg.GRPCAddr, "g", "", "grpc server address")
	flag.UintVar(&cfg.ReportInterval, "r", cfg.ReportInterval, "report interval")
	flag.UintVar(&cfg.PollInterval, "p", cfg.PollInterval, "poll interval")
	flag.UintVar(&cfg.RateLimit, "l", cfg.RateLimit, "rate limit")
	flag.StringVar(&cfg.Key, "k", cfg.Key, "key")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", "", "path to public key file for encryption")
	flag.Parse()

	// Сначала применяем переменные окружения
	err := cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}

	// Определяем путь к конфигурационному файлу
	configPath := ""
	if *flagConfig != "" {
		configPath = *flagConfig
	} else if *flagConfigLong != "" {
		configPath = *flagConfigLong
	} else if cfg.ConfigFile != "" {
		configPath = cfg.ConfigFile
	}

	// Если указан конфигурационный файл, читаем его
	if configPath != "" {
		if err := cleanenv.ReadConfig(configPath, cfg); err != nil {
			return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
		}
	}

	if !isValidURL(cfg.Addr) {
		cfg.Addr = "http://" + cfg.Addr
	}

	return cfg, nil
}

func isValidURL(addr string) bool {
	return strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://")
}
