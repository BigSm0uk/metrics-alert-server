package agent

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/goccy/go-json"
	"go.uber.org/zap"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/semaphore"
	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
	"github.com/bigsm0uk/metrics-alert-server/pkg/util"
	"github.com/bigsm0uk/metrics-alert-server/pkg/util/crypto"
	"github.com/bigsm0uk/metrics-alert-server/pkg/util/hasher"
)

type IMetricsSender interface {
	RunProcess(ctx context.Context, wg *sync.WaitGroup, reportInterval uint, collector Collector, sem *semaphore.Semaphore, key string)
	Close() error
}

type MetricsSender struct {
	client    *resty.Client
	serverURL string
	logger    *zap.Logger
	publicKey *rsa.PublicKey
	localIP   string
}

const (
	maxRetries = 3
	retryDelay = time.Second
)

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}
	return ""
}

func NewMetricsSender(serverURL string, logger *zap.Logger, cryptoKeyPath string) (*MetricsSender, error) {
	c := resty.New()
	c.SetRetryCount(maxRetries)
	c.SetRetryWaitTime(retryDelay)

	var publicKey *rsa.PublicKey
	var err error
	if cryptoKeyPath != "" {
		publicKey, err = crypto.LoadPublicKey(cryptoKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load public key: %w", err)
		}
		logger.Info("public key loaded for encryption", zap.String("path", cryptoKeyPath))
	}

	localIP := getLocalIP()
	if localIP != "" {
		logger.Info("local IP detected", zap.String("ip", localIP))
	}

	return &MetricsSender{
		client:    c,
		serverURL: serverURL,
		logger:    logger,
		publicKey: publicKey,
		localIP:   localIP,
	}, nil
}

// prepareRequest подготавливает HTTP запрос с шифрованием и заголовками
func (s *MetricsSender) prepareRequest(jsonData []byte, key string) (*resty.Request, []byte, error) {
	compressedData, err := util.CompressJSON(jsonData)
	if err != nil {
		return nil, nil, err
	}

	// Шифруем данные, если публичный ключ доступен
	if s.publicKey != nil {
		compressedData, err = crypto.Encrypt(compressedData, s.publicKey)
		if err != nil {
			return nil, nil, err
		}
	}

	req := s.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip")

	if s.localIP != "" {
		req.SetHeader("X-Real-IP", s.localIP)
	}
	if s.publicKey != nil {
		req.SetHeader("Content-Encryption", "rsa")
	}
	if key != "" {
		req.SetHeader("HashSHA256", hasher.Hash(string(jsonData), key))
	}

	req.SetBody(compressedData)
	return req, compressedData, nil
}

func (s *MetricsSender) SendMetrics(metrics []domain.Metrics, key string) error {
	if len(metrics) == 0 {
		s.logger.Debug("no metrics to send, skipping")
		return nil
	}

	jsonMetrics, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/updates", s.serverURL)
	req, compressedData, err := s.prepareRequest(jsonMetrics, key)
	if err != nil {
		return err
	}

	resp, err := req.Post(url)
	if err != nil {
		return err
	}

	s.logger.Debug("metrics batch sent",
		zap.Int("metrics_count", len(metrics)),
		zap.Int("status", resp.StatusCode()),
		zap.Int("compressed_size", len(compressedData)))
	return nil
}

// SendMetric отправляет одну метрику со сжатием
func (s *MetricsSender) SendMetric(metric domain.Metrics, key string) error {
	jsonMetric, err := json.Marshal(metric)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/update", s.serverURL)
	req, compressedData, err := s.prepareRequest(jsonMetric, key)
	if err != nil {
		return err
	}

	resp, err := req.Post(url)
	if err != nil {
		return err
	}

	s.logger.Debug("metric sent",
		zap.String("metric", metric.ID),
		zap.Int("status", resp.StatusCode()),
		zap.Int("compressed_size", len(compressedData)))
	return nil
}

type Collector interface {
	GetMetrics() []domain.Metrics
}

func (s *MetricsSender) Close() error {
	return nil
}

func (s *MetricsSender) RunProcess(ctx context.Context, wg *sync.WaitGroup, reportInterval uint, collector Collector, sem *semaphore.Semaphore, key string) {
	ticker := time.NewTicker(time.Duration(reportInterval) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			metrics := collector.GetMetrics()
			wg.Add(1)
			go func() {
				defer wg.Done()
				sem.Acquire()
				defer sem.Release()
				if err := s.SendMetrics(metrics, key); err != nil {
					s.logger.Error("failed to send metrics", zap.Error(err))
				}
			}()
		}
	}
}
