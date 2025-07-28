package handler

import (
	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
	"github.com/bigsm0uk/metrics-alert-server/internal/service"
)

type UpdateMetricDTO struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

func (m *UpdateMetricDTO) Validate() error {
	if m.ID == "" {
		return service.ErrMetricNotFound
	}
	if m.Type != domain.Counter && m.Type != domain.Gauge {
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
