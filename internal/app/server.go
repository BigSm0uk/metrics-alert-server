package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/config"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/router"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/zl"
	"github.com/bigsm0uk/metrics-alert-server/internal/domain/interfaces"
	"github.com/bigsm0uk/metrics-alert-server/internal/handler"
	"github.com/bigsm0uk/metrics-alert-server/internal/service"
)

type Server struct {
	cfg *config.ServerConfig
	h   *handler.MetricHandler
	Ms  interfaces.MetricsStore
	as  *service.AuditService
}

func NewServer(cfg *config.ServerConfig, h *handler.MetricHandler, ms interfaces.MetricsStore, as *service.AuditService) *Server {
	return &Server{cfg: cfg, h: h, Ms: ms, as: as}
}

func (a *Server) Run() error {
	r := router.NewRouter(a.h, a.cfg.Key)

	srv := &http.Server{
		Addr:    a.cfg.Addr,
		Handler: r,
	}

	go func() {
		zl.Log.Info("starting server", zap.String("Addr", "http://"+a.cfg.Addr))

		ctx := context.Background()
		a.Ms.StartProcess(ctx)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			zl.Log.Fatal("failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	zl.Log.Info("shutting down server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := a.Ms.Close(ctx); err != nil {
		zl.Log.Error("failed to close metric store", zap.Error(err))
		return err
	}
	if err := a.h.Close(); err != nil {
		zl.Log.Error("failed to close metric handler", zap.Error(err))
		return err
	}

	if err := srv.Shutdown(ctx); err != nil {
		zl.Log.Error("server forced to shutdown", zap.Error(err))
		return err
	}

	zl.Log.Info("server exiting")
	return nil
}
