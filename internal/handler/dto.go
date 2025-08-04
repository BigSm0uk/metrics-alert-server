package handler

import (
	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
	"github.com/bigsm0uk/metrics-alert-server/internal/service"
)

type ParamMetric struct {
	ID    string `json:"id"`
	MType string `json:"type"`
	Value string `json:"value"`
}

func (m *ParamMetric) Validate() error {
	if m.ID == "" {
		return service.ErrMetricNotFound
	}
	if m.MType != domain.Counter && m.MType != domain.Gauge {
		return service.ErrInvalidMetricType
	}
	return nil
}

type GetMetricDTO struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

func (gm *GetMetricDTO) Validate() error {
	if gm.ID == "" {
		return service.ErrMetricNotFound
	}
	if gm.Type != domain.Counter && gm.Type != domain.Gauge {
		return service.ErrInvalidMetricType
	}
	return nil
}

type BodyMetric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (m *BodyMetric) Validate() error {
	if m.ID == "" {
		return service.ErrMetricNotFound
	}
	if m.MType != domain.Counter && m.MType != domain.Gauge {
		return service.ErrInvalidMetricType
	}
	return nil
}
