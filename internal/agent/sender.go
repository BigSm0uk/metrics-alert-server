package agent

import (
	"fmt"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/zl"
	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
	"github.com/bigsm0uk/metrics-alert-server/pkg/util"
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

func (s *MetricsSender) SendMetricsV2(metrics []domain.Metrics) error {

	if len(metrics) == 0 {
		zl.Log.Debug("no metrics to send, skipping")
		return nil
	}

	compressedData, err := util.CompressJSON(metrics)
	if err != nil {
		zl.Log.Error("failed to compress metrics", zap.Error(err))
		return err
	}

	url := fmt.Sprintf("%s/updates", s.serverURL)
	resp, err := s.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Accept-Encoding", "gzip").
		SetBody(compressedData).
		Post(url)

	if err != nil {
		zl.Log.Error("failed to send metrics batch",
			zap.Int("metrics_count", len(metrics)),
			zap.Error(err))
		return err
	}

	zl.Log.Debug("metrics batch sent",
		zap.Int("metrics_count", len(metrics)),
		zap.Int("status", resp.StatusCode()),
		zap.Int("compressed_size", len(compressedData)))

	return nil
}

// SendMetricV2 отправляет одну метрику со сжатием
func (s *MetricsSender) SendMetricV2(metric domain.Metrics) error {
	compressedData, err := util.CompressJSON(metric)
	if err != nil {
		zl.Log.Error("failed to compress metric",
			zap.String("metric", metric.ID),
			zap.Error(err))
		return err
	}

	url := fmt.Sprintf("%s/update", s.serverURL)
	resp, err := s.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Accept-Encoding", "gzip").
		SetBody(compressedData).
		Post(url)

	if err != nil {
		zl.Log.Error("failed to send metric",
			zap.String("metric", metric.ID),
			zap.Error(err))
		return err
	}

	zl.Log.Debug("metric sent",
		zap.String("metric", metric.ID),
		zap.Int("status", resp.StatusCode()),
		zap.Int("compressed_size", len(compressedData)))

	return nil
}
