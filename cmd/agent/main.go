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
	err = app.RunV2()
	if err != nil {
		panic(err)
	}
}

func InitializeApp() (*app.Agent, error) {
	cfg, err := config.LoadAgentConfig()
	if err != nil {
		return nil, err
	}
	zl.InitLogger(cfg.Env)
	defer zl.Log.Sync()

	return app.NewAgent(cfg), nil
}
