package app

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/bigsm0uk/metrics-alert-server/internal/agent"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/semaphore"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/zl"
	"github.com/bigsm0uk/metrics-alert-server/internal/config"
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
	zl.Log.Info("starting agent, to send metrics to", zap.String("Addr", a.Cfg.Addr))

	pollTicker := time.NewTicker(time.Duration(a.Cfg.PollInterval) * time.Second)
	reportTicker := time.NewTicker(time.Duration(a.Cfg.ReportInterval) * time.Second)

	defer pollTicker.Stop()
	defer reportTicker.Stop()

	for {
		select {
		case <-pollTicker.C:
			a.Collector.CollectRuntimeMetrics()
			zl.Log.Debug("metrics collected")

		case <-reportTicker.C:
			metrics := a.Collector.GetMetrics()
			if err := a.Sender.SendMetricsV2(metrics, a.Cfg.Key); err != nil {
				zl.Log.Error("failed to send metrics", zap.Error(err))
			} else {
				zl.Log.Info("metrics sent", zap.Int("count", len(metrics)))
			}
		}
	}
}
func (a *Agent) RunV2() error {

	go a.Collector.RunProcess(a.Cfg.PollInterval)
	go a.Sender.RunProcess(a.Cfg.ReportInterval, a.Collector, a.Sem, a.Cfg.Key)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	zl.Log.Info("shutting down agent ...")

	return nil
}
