package config

import (
	"flag"
)

type AgentConfig struct {
	Env            string
	Addr           string
	ReportInterval uint
	PollInterval   uint
	Server         string
}

func LoadAgentConfig() (*AgentConfig, error) {
	cfg := &AgentConfig{}
	flag.StringVar(&cfg.Env, "e", EnvDevelopment, "environment")
	flag.StringVar(&cfg.Addr, "a", ":8090", "agent address")
	flag.UintVar(&cfg.ReportInterval, "r", 10, "report interval")
	flag.UintVar(&cfg.PollInterval, "p", 2, "poll interval")
	flag.StringVar(&cfg.Server, "s", "http://localhost:8080", "server address")
	flag.Parse()

	return cfg, nil
}
