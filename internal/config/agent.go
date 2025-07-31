package config

import (
	"flag"

	"github.com/ilyakaznacheev/cleanenv"
)

type AgentConfig struct {
	Env            string `env:"ENV"`
	Addr           string `env:"AGENT_ADDRESS"`
	ReportInterval uint   `env:"REPORT_INTERVAL"`
	PollInterval   uint   `env:"POLL_INTERVAL"`
	Server         string `env:"ADDRESS"`
}

func LoadAgentConfig() (*AgentConfig, error) {
	cfg := &AgentConfig{}
	flag.StringVar(&cfg.Env, "e", EnvDevelopment, "environment")
	flag.StringVar(&cfg.Addr, "a", ":8090", "agent address")
	flag.UintVar(&cfg.ReportInterval, "r", 10, "report interval")
	flag.UintVar(&cfg.PollInterval, "p", 2, "poll interval")
	flag.StringVar(&cfg.Server, "s", "http://localhost:8080", "server address")
	flag.Parse()

	err := cleanenv.ReadEnv(cfg)

	return cfg, err
}
