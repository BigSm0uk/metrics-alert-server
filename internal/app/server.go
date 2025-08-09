package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/router"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/zl"
	"github.com/bigsm0uk/metrics-alert-server/internal/config"
	"github.com/bigsm0uk/metrics-alert-server/internal/handler"
	"github.com/bigsm0uk/metrics-alert-server/internal/server"
)

type Server struct {
	cfg *config.ServerConfig
	h   *handler.MetricHandler
	Ms  *server.MetricStore
}

func NewServer(cfg *config.ServerConfig, h *handler.MetricHandler, ms *server.MetricStore) *Server {
	return &Server{cfg: cfg, h: h, Ms: ms}
}

func (a *Server) Run() error {
	r := router.NewRouter(a.h)

	srv := &http.Server{
		Addr:    a.cfg.Addr,
		Handler: r,
	}

	go func() {
		zl.Log.Info("starting server", zap.String("Addr", a.cfg.Addr))

		a.Ms.StartProcess()
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zl.Log.Fatal("failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	zl.Log.Info("shutting down server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := a.Ms.Close(); err != nil {
		zl.Log.Error("failed to close metric store", zap.Error(err))
		return err
	}

	if err := srv.Shutdown(ctx); err != nil {
		zl.Log.Error("server forced to shutdown", zap.Error(err))
		return err
	}

	zl.Log.Info("server exiting")
	return nil
}
