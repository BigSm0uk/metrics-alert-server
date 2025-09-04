package config

import (
	"flag"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
)

type AgentConfig struct {
	Env            string `env:"ENV"`
	Addr           string `env:"ADDRESS"`
	ReportInterval uint   `env:"REPORT_INTERVAL"`
	PollInterval   uint   `env:"POLL_INTERVAL"`
	Key            string `env:"KEY"`
}

func LoadAgentConfig() (*AgentConfig, error) {
	cfg := &AgentConfig{}
	flag.StringVar(&cfg.Env, "e", EnvDevelopment, "environment")
	flag.StringVar(&cfg.Addr, "a", "localhost:8080", "http server address")
	flag.UintVar(&cfg.ReportInterval, "r", 10, "report interval")
	flag.UintVar(&cfg.PollInterval, "p", 2, "poll interval")
	flag.StringVar(&cfg.Key, "k", "", "key")
	flag.Parse()

	err := cleanenv.ReadEnv(cfg)

	if !isValidURL(cfg.Addr) {
		cfg.Addr = "http://" + cfg.Addr
	}

	return cfg, err
}
func isValidURL(addr string) bool {
	return strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://")
}
