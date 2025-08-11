package jsonStore

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/zl"
	"github.com/bigsm0uk/metrics-alert-server/internal/config/store"
	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
	"github.com/bigsm0uk/metrics-alert-server/internal/interfaces"

	"go.uber.org/zap"
)

type Store struct {
	r   interfaces.MetricsRepository
	cfg *store.StoreConfig

	storeInterval time.Duration
	syncMode      bool // true если интервал = 0
	ticker        *time.Ticker
	stopChan      chan struct{}
}

var _ interfaces.MetricsStore = (*Store)(nil)

func NewStore(r interfaces.MetricsRepository, cfg *store.StoreConfig) (*Store, error) {
	interval, err := time.ParseDuration(cfg.StoreInterval + "s")
	if err != nil {
		return nil, err
	}

	syncMode := interval == 0

	ms := &Store{
		r:             r,
		cfg:           cfg,
		storeInterval: interval,
		syncMode:      syncMode,
		ticker:        time.NewTicker(interval),
		stopChan:      make(chan struct{}),
	}

	return ms, nil
}
func (s *Store) StartProcess() {
	if s.syncMode {
		return
	}
	s.startPeriodicSave()
}
func (s *Store) IsSyncMode() bool {
	return s.syncMode
}
func (s *Store) startPeriodicSave() {
	s.ticker = time.NewTicker(s.storeInterval)
	go func() {
		for {
			select {
			case <-s.ticker.C:
				if err := s.SaveAllMetrics(); err != nil {
					zl.Log.Error("failed to save metrics during periodic save", zap.Error(err))
				}
			case <-s.stopChan:
				return
			}
		}
	}()
}
func (s *Store) WriteMetric(metric domain.Metrics) error {
	file, err := os.OpenFile(s.cfg.FileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		zl.Log.Error("failed to open file for write", zap.Error(err), zap.String("filePath", s.cfg.FileStoragePath))
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	encoder := json.NewEncoder(writer)
	if err := encoder.Encode(metric); err != nil {
		zl.Log.Error("failed to encode metric", zap.Error(err))
		return err
	}

	return nil
}
func (s *Store) SaveAllMetrics() error {
	metrics, err := s.r.GetAll()
	if err != nil {
		return err
	}

	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)

	for _, metric := range metrics {
		if err := encoder.Encode(metric); err != nil {
			zl.Log.Error("failed to encode metric to buffer", zap.Error(err))
			return err
		}
	}

	if err := os.WriteFile(s.cfg.FileStoragePath, buffer.Bytes(), 0644); err != nil {
		zl.Log.Error("failed to write buffer to file", zap.Error(err), zap.String("filePath", s.cfg.FileStoragePath))
		return err
	}

	zl.Log.Info("successfully saved all metrics", zap.Int("count", len(metrics)))
	return nil
}

func (s *Store) Restore() error {
	file, err := os.Open(s.cfg.FileStoragePath)
	if err != nil {
		if os.IsNotExist(err) {
			zl.Log.Info("storage file does not exist, starting with empty store")
			return nil
		}
		zl.Log.Error("failed to open file for read", zap.Error(err), zap.String("filePath", s.cfg.FileStoragePath))
		return err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	decoder := json.NewDecoder(reader)

	count := 0
	for {
		var metric domain.Metrics
		err := decoder.Decode(&metric)
		if err == io.EOF {
			break
		}
		if err != nil {
			zl.Log.Error("failed to decode metric", zap.Error(err))
			return err
		}

		if err := s.r.Save(&metric); err != nil {
			zl.Log.Error("failed to save metric during restore", zap.Error(err))
			return err
		}
		count++
	}

	zl.Log.Info("successfully restored metrics", zap.Int("count", count))
	return nil
}
func (s *Store) Close() error {
	if s.ticker != nil {
		s.ticker.Stop()
		close(s.stopChan)
	}

	if err := s.SaveAllMetrics(); err != nil {
		zl.Log.Error("failed to save metrics during close", zap.Error(err))
		return err
	}

	zl.Log.Info("store closed successfully")
	return nil
}
