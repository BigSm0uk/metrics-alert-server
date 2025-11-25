package main

import (
	"log"

	"github.com/bigsm0uk/metrics-alert-server/internal/app"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/config"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/zl"
)

func main() {
	app, err := InitializeApp()
	if err != nil {
		log.Fatal(err)
	}
	err = app.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func InitializeApp() (*app.Agent, error) {
	cfg, err := config.LoadAgentConfig()
	if err != nil {
		return nil, err
	}
	logger := zl.NewLogger(cfg.Env)
	defer logger.Sync()

	return app.NewAgent(cfg, logger), nil
}
