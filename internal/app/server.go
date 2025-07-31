package app

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/router"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/zl"
	"github.com/bigsm0uk/metrics-alert-server/internal/config"
	"github.com/bigsm0uk/metrics-alert-server/internal/handler"
)

type Server struct {
	cfg *config.ServerConfig
	h   *handler.MetricHandler
}

func NewServer(cfg *config.ServerConfig, h *handler.MetricHandler) *Server {
	return &Server{cfg: cfg, h: h}
}

func (a *Server) Run() error {
	r := router.NewRouter(a.h)

	zl.Log.Info("starting server", zap.String("Addr", a.cfg.Addr))
	if err := http.ListenAndServe(a.cfg.Addr, r); err != nil {
		zl.Log.Fatal("failed to start server", zap.Error(err))
	}

	return nil
}
