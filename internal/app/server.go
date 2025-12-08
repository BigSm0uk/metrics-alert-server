package app

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/config"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/router"
	"github.com/bigsm0uk/metrics-alert-server/internal/domain/interfaces"
	"github.com/bigsm0uk/metrics-alert-server/internal/handler"
	"github.com/bigsm0uk/metrics-alert-server/internal/service"
	"github.com/bigsm0uk/metrics-alert-server/pkg/util/crypto"
)

type Server struct {
	cfg    *config.ServerConfig
	h      *handler.MetricHandler
	ms     interfaces.MetricsStore
	as     *service.AuditService
	logger *zap.Logger
}

func NewServer(cfg *config.ServerConfig, h *handler.MetricHandler, ms interfaces.MetricsStore, as *service.AuditService, logger *zap.Logger) *Server {
	return &Server{cfg: cfg, h: h, ms: ms, as: as, logger: logger}
}

func (a *Server) Run() error {
	var privateKey *rsa.PrivateKey
	var err error
	if a.cfg.CryptoKey != "" {
		privateKey, err = crypto.LoadPrivateKey(a.cfg.CryptoKey)
		if err != nil {
			return fmt.Errorf("failed to load private key: %w", err)
		}
		a.logger.Info("private key loaded for decryption", zap.String("path", a.cfg.CryptoKey))
	}

	r := router.NewRouter(a.h, a.cfg.Key, a.logger, privateKey)

	srv := &http.Server{
		Addr:    a.cfg.Addr,
		Handler: r,
	}
	go func() {
		a.logger.Info("pprof server listening on :6060")
		a.logger.Info("error starting pprof server", zap.Error(http.ListenAndServe("localhost:6060", nil)))
	}()
	go func() {
		a.logger.Info("starting server", zap.String("Addr", "http://"+a.cfg.Addr))

		ctx := context.Background()
		a.ms.StartProcess(ctx)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Fatal("failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit

	a.logger.Info("shutting down server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := a.ms.Close(ctx); err != nil {
		a.logger.Error("failed to close metric store", zap.Error(err))
		return err
	}
	if err := a.h.Close(); err != nil {
		a.logger.Error("failed to close metric handler", zap.Error(err))
		return err
	}

	if err := srv.Shutdown(ctx); err != nil {
		a.logger.Error("server forced to shutdown", zap.Error(err))
		return err
	}

	a.logger.Info("server exiting")
	return nil
}
