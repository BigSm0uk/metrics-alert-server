package main

import (
	"fmt"
	"os"

	"github.com/bigsm0uk/metrics-alert-server/internal/app"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/zl"
	"github.com/bigsm0uk/metrics-alert-server/internal/config"
	"github.com/bigsm0uk/metrics-alert-server/internal/handler"
	"github.com/bigsm0uk/metrics-alert-server/internal/repository"
	"github.com/bigsm0uk/metrics-alert-server/internal/service"
)

func main() {
	fmt.Printf("Server starting with args: %v\n", os.Args)
	fmt.Printf("Working directory: %s\n", func() string {
		wd, _ := os.Getwd()
		return wd
	}())

	app, err := InitializeApp()
	if err != nil {
		panic(err)
	}
	err = app.Run()
	if err != nil {
		panic(err)
	}
}

func InitializeApp() (*app.Server, error) {
	cfg, err := config.LoadServerConfig()
	if err != nil {
		return nil, err
	}
	logger := zl.InitLogger(cfg.Env)
	defer logger.Sync()

	r, err := repository.InitRepository(cfg)
	if err != nil {
		return nil, err
	}
	service := service.NewService(r, logger)

	handler := handler.NewMetricHandler(service, cfg.TemplatePath)
	return app.NewServer(cfg, handler, logger), nil
}
