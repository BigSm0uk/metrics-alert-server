package app

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/router"
	"github.com/bigsm0uk/metrics-alert-server/internal/config"
	"github.com/bigsm0uk/metrics-alert-server/internal/handler"
)

type App struct {
	cfg    *config.Config
	h      *handler.MetricHandler
	logger *zap.Logger
}

func NewApp(cfg *config.Config, h *handler.MetricHandler, logger *zap.Logger) *App {
	return &App{cfg: cfg, h: h, logger: logger}
}

func (a *App) Run() error {

	r := router.NewRouter(a.h, a.logger)

	a.logger.Info("starting server", zap.String("Addr", a.cfg.Addr))
	if err := http.ListenAndServe(a.cfg.Addr, r); err != nil {
		a.logger.Fatal("failed to start server", zap.Error(err))
	}

	return nil
}
