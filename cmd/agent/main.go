package main

import (
	"github.com/bigsm0uk/metrics-alert-server/internal/app"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/zl"
	"github.com/bigsm0uk/metrics-alert-server/internal/config"
)

func main() {
	app, err := InitializeApp()
	if err != nil {
		panic(err)
	}
	err = app.Run()
	if err != nil {
		panic(err)
	}
}

func InitializeApp() (*app.Agent, error) {
	cfg, err := config.LoadAgentConfig()
	if err != nil {
		return nil, err
	}
	logger := zl.InitDefaultLogger()
	defer logger.Sync()

	return app.NewAgent(logger, cfg), nil
}
