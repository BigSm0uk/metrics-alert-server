package main

import (
	"io"

	"github.com/bigsm0uk/metrics-alert-server/internal/app"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/zl"
	"github.com/bigsm0uk/metrics-alert-server/internal/config"
	"github.com/bigsm0uk/metrics-alert-server/internal/handler"
	"github.com/bigsm0uk/metrics-alert-server/internal/repository"
	"github.com/bigsm0uk/metrics-alert-server/internal/server"
	"github.com/bigsm0uk/metrics-alert-server/internal/service"
)

func initializeApp() (*app.Server, error) {
	cfg, err := config.LoadServerConfig()
	if err != nil {
		return nil, err
	}
	zl.InitLogger(cfg.Env)
	defer zl.Log.Sync()

	r, err := repository.InitRepository(cfg)
	if err != nil {
		return nil, err
	}
	ms, err := server.NewMetricStore(r, &cfg.Store)
	if err != nil {
		return nil, err
	}
	if cfg.Store.Restore {
		if err := ms.Restore(); err != nil && err != io.EOF {
			return nil, err
		}
	}
	service := service.NewService(r, ms)

	handler := handler.NewMetricHandler(service, cfg.TemplatePath)

	return app.NewServer(cfg, handler, ms), nil
}
