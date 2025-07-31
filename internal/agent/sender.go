package agent

import (
	"fmt"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/zl"
	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
)

type MetricsSender struct {
	client    *resty.Client
	serverURL string
}

func NewMetricsSender(serverURL string) *MetricsSender {
	return &MetricsSender{
		client:    resty.New(),
		serverURL: serverURL,
	}
}

func (s *MetricsSender) SendMetrics(metrics []domain.Metrics) error {
	for _, metric := range metrics {
		var value string
		if metric.MType == domain.Gauge && metric.Value != nil {
			value = fmt.Sprintf("%g", *metric.Value)
		} else if metric.MType == domain.Counter && metric.Delta != nil {
			value = fmt.Sprintf("%d", *metric.Delta)
		}

		url := fmt.Sprintf("%s/update/%s/%s/%s", s.serverURL, metric.MType, metric.ID, value)

		resp, err := s.client.R().SetHeader("Content-Type", "text/plain").Post(url)
		if err != nil {
			zl.Log.Error("failed to send metric",
				zap.String("metric", metric.ID),
				zap.Error(err))
			return err
		}

		zl.Log.Debug("metric sent",
			zap.String("metric", metric.ID),
			zap.Int("status", resp.StatusCode()))
	}

	return nil
}
