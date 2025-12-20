package app

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	pb "github.com/bigsm0uk/metrics-alert-server/api/proto"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/config"
	grpcserver "github.com/bigsm0uk/metrics-alert-server/internal/app/grpc"
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
	svc    *service.MetricService
	logger *zap.Logger
}

func NewServer(cfg *config.ServerConfig, h *handler.MetricHandler, ms interfaces.MetricsStore, as *service.AuditService, svc *service.MetricService, logger *zap.Logger) *Server {
	return &Server{cfg: cfg, h: h, ms: ms, as: as, svc: svc, logger: logger}
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

	r := router.NewRouter(a.h, a.cfg.Key, a.logger, privateKey, a.cfg.TrustedSubnet)

	srv := &http.Server{
		Addr:    a.cfg.Addr,
		Handler: r,
	}

	var grpcSrv *grpc.Server
	if a.cfg.GRPCAddr != "" {
		opts := []grpc.ServerOption{}
		if a.cfg.TrustedSubnet != "" {
			opts = append(opts, grpc.UnaryInterceptor(grpcserver.SubnetCheckInterceptor(a.cfg.TrustedSubnet)))
		}
		grpcSrv = grpc.NewServer(opts...)
		pb.RegisterMetricsServer(grpcSrv, grpcserver.NewMetricsServer(a.svc))

		go func() {
			lis, err := net.Listen("tcp", a.cfg.GRPCAddr)
			if err != nil {
				a.logger.Fatal("failed to listen gRPC", zap.Error(err))
			}
			a.logger.Info("starting gRPC server", zap.String("Addr", a.cfg.GRPCAddr))
			if err := grpcSrv.Serve(lis); err != nil {
				a.logger.Fatal("failed to serve gRPC", zap.Error(err))
			}
		}()
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

	if grpcSrv != nil {
		grpcSrv.GracefulStop()
		a.logger.Info("gRPC server stopped")
	}

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
