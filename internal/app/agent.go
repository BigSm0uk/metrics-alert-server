package app

import (
	"time"

	"go.uber.org/zap"

	"github.com/bigsm0uk/metrics-alert-server/internal/agent"
	"github.com/bigsm0uk/metrics-alert-server/internal/config"
)

type Agent struct {
	Cfg       *config.AgentConfig
	Collector *agent.MetricsCollector
	Sender    *agent.MetricsSender
	Logger    *zap.Logger
}

func NewAgent(logger *zap.Logger, cfg *config.AgentConfig) *Agent {
	return &Agent{Logger: logger, Cfg: cfg, Collector: agent.NewMetricsCollector(), Sender: agent.NewMetricsSender(cfg.Server, logger)}
}

func (a *Agent) Run() error {
	a.Logger.Info("starting agent", zap.String("Addr", a.Cfg.Addr))

	pollTicker := time.NewTicker(time.Duration(a.Cfg.PollInterval) * time.Second)
	reportTicker := time.NewTicker(time.Duration(a.Cfg.ReportInterval) * time.Second)

	defer pollTicker.Stop()
	defer reportTicker.Stop()

	for {
		select {
		case <-pollTicker.C:
			a.Collector.CollectRuntimeMetrics()
			a.Logger.Debug("metrics collected")

		case <-reportTicker.C:
			metrics := a.Collector.GetMetrics()
			if err := a.Sender.SendMetrics(metrics); err != nil {
				a.Logger.Error("failed to send metrics", zap.Error(err))
			} else {
				a.Logger.Info("metrics sent", zap.Int("count", len(metrics)))
			}
		}
	}
}
