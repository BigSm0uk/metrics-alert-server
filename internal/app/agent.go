package app

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"go.uber.org/zap"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/agent"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/config"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/semaphore"
)

type Agent struct {
	Cfg       *config.AgentConfig
	Collector *agent.MetricsCollector
	Sender    agent.IMetricsSender
	Sem       *semaphore.Semaphore
	logger    *zap.Logger
}

func NewAgent(cfg *config.AgentConfig, logger *zap.Logger) (*Agent, error) {
	var sender agent.IMetricsSender

	if cfg.GRPCAddr != "" {
		grpcSender, err := agent.NewGRPCMetricsSender(cfg.GRPCAddr, logger)
		if err != nil {
			return nil, err
		}
		sender = grpcSender
	} else {
		httpSender, err := agent.NewMetricsSender(cfg.Addr, logger, cfg.CryptoKey)
		if err != nil {
			return nil, err
		}
		sender = httpSender
	}

	return &Agent{
		Cfg:       cfg,
		Collector: agent.NewMetricsCollector(logger),
		Sender:    sender,
		Sem:       semaphore.NewSemaphore(int(cfg.RateLimit)),
		logger:    logger,
	}, nil
}

func (a *Agent) Run() error {
	wg := sync.WaitGroup{}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	go a.Collector.RunProcess(ctx, &wg, a.Cfg.PollInterval)
	go a.Sender.RunProcess(ctx, &wg, a.Cfg.ReportInterval, a.Collector, a.Sem, a.Cfg.Key)

	<-ctx.Done()
	wg.Wait()

	a.logger.Info("shutting down agent ...")

	if err := a.Sender.Close(); err != nil {
		a.logger.Error("failed to close sender", zap.Error(err))
	}

	return nil
}
