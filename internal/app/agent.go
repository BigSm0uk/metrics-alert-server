package app

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/agent"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/config"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/semaphore"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/zl"
)

type Agent struct {
	Cfg       *config.AgentConfig
	Collector *agent.MetricsCollector
	Sender    *agent.MetricsSender
	Sem       *semaphore.Semaphore
}

func NewAgent(cfg *config.AgentConfig) *Agent {
	return &Agent{Cfg: cfg, Collector: agent.NewMetricsCollector(), Sender: agent.NewMetricsSender(cfg.Addr), Sem: semaphore.NewSemaphore(int(cfg.RateLimit))}
}

func (a *Agent) Run() error {

	wg := sync.WaitGroup{}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go a.Collector.RunProcess(ctx, &wg, a.Cfg.PollInterval)
	go a.Sender.RunProcess(ctx, &wg, a.Cfg.ReportInterval, a.Collector, a.Sem, a.Cfg.Key)

	<-ctx.Done()
	wg.Wait()

	zl.Log.Info("shutting down agent ...")

	return nil
}
