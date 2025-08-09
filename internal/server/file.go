package server

import (
	"time"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/zl"
	"github.com/bigsm0uk/metrics-alert-server/internal/config/store"
	"github.com/bigsm0uk/metrics-alert-server/internal/interfaces"
	"go.uber.org/zap"
)

type MetricStore struct {
	r   interfaces.MetricsRepository
	cfg *store.StoreConfig
	Sw  *storeWriter
	Sr  *storeReader

	storeInterval time.Duration
	syncMode      bool // true если интервал = 0
	ticker        *time.Ticker
	stopChan      chan struct{}
}

func NewMetricStore(r interfaces.MetricsRepository, cfg *store.StoreConfig) (*MetricStore, error) {

	interval, err := time.ParseDuration(cfg.StoreInterval + "s")
	if err != nil {
		return nil, err
	}

	syncMode := interval == 0

	sw, err := newStoreWriter(cfg.FileStoragePath)
	if err != nil {
		zl.Log.Error("failed to create store writer", zap.Error(err))
		return nil, err
	}
	sr, err := newStoreReader(cfg.FileStoragePath)
	if err != nil {
		zl.Log.Error("failed to create store reader", zap.Error(err))
		return nil, err
	}

	ms := &MetricStore{r: r, cfg: cfg, Sw: sw, Sr: sr, storeInterval: interval, syncMode: syncMode, ticker: time.NewTicker(interval), stopChan: make(chan struct{})}

	return ms, nil
}
func (s *MetricStore) StartProcess() {
	if s.syncMode {
		return
	}
	s.startPeriodicSave()
}
func (s *MetricStore) IsSyncMode() bool {
	return s.syncMode
}
func (s *MetricStore) startPeriodicSave() {
	s.ticker = time.NewTicker(s.storeInterval)
	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.Sw.Truncate()
				s.saveAllMetrics()
			case <-s.stopChan:
				return
			}
		}
	}()
}

func (s *MetricStore) saveAllMetrics() error {
	metrics, err := s.r.GetAll()
	if err != nil {
		return err
	}

	for _, m := range metrics {
		if err := s.Sw.WriteMetric(m); err != nil {
			return err
		}
	}
	return nil
}

func (s *MetricStore) Restore() error {
	metrics, err := s.Sr.ReadAll()
	if err != nil {
		zl.Log.Warn("failed to read metrics", zap.Error(err))
		return err
	}
	for _, m := range metrics {
		if err := s.r.Save(m); err != nil {
			zl.Log.Error("failed to save metric", zap.Error(err))
			return err
		}
	}
	return nil
}
func (s *MetricStore) Close() error {
	if s.ticker != nil {
		s.ticker.Stop()
		close(s.stopChan)
	}
	if err := s.Sw.Truncate(); err != nil {
		zl.Log.Error("failed to truncate store writer", zap.Error(err))
		return err
	}
	if err := s.saveAllMetrics(); err != nil {
		zl.Log.Error("failed to save metrics", zap.Error(err))
		return err
	}
	if err := s.Sw.Close(); err != nil {
		zl.Log.Error("failed to close store writer", zap.Error(err))
		return err
	}
	if err := s.Sr.Close(); err != nil {
		zl.Log.Error("failed to close store reader", zap.Error(err))
		return err
	}
	return nil
}
