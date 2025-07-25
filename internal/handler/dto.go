package handler

import (
	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
	"github.com/bigsm0uk/metrics-alert-server/internal/service"
)

type MetricDTO struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

func (m *MetricDTO) Validate() error {
	if m.ID == "" {
		return service.ErrMetricNotFound
	}
	if m.Type != domain.Counter && m.Type != domain.Gauge {
		return service.ErrInvalidMetricType
	}
	return nil
}
