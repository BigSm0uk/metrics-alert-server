package server

import (
	"encoding/json"
	"io"
	"os"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/zl"
	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
	"go.uber.org/zap"
)

type storeWriter struct {
	encoder *json.Encoder
	file    *os.File
}

func newStoreWriter(fileName string) (*storeWriter, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		zl.Log.Error("failed to open file", zap.Error(err))
		return nil, err
	}
	return &storeWriter{encoder: json.NewEncoder(file), file: file}, nil
}
func (w *storeWriter) Truncate() error {
	return w.file.Truncate(0)
}

func (w *storeWriter) WriteMetric(metric domain.Metrics) error {
	return w.encoder.Encode(metric)
}

func (w *storeWriter) Close() error {
	return w.file.Close()
}

type storeReader struct {
	decoder *json.Decoder
	file    *os.File
}

func newStoreReader(fileName string) (*storeReader, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		zl.Log.Error("failed to open file", zap.Error(err))
		return nil, err
	}
	return &storeReader{decoder: json.NewDecoder(file), file: file}, nil
}

func (r *storeReader) ReadAll() ([]*domain.Metrics, error) {
	var arr []*domain.Metrics
	for {
		var m domain.Metrics
		err := r.decoder.Decode(&m)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		arr = append(arr, &m)
	}
	return arr, nil
}
func (r *storeReader) Close() error {
	return r.file.Close()
}
