package handler

import (
	"strconv"

	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
)

// validateMetricType проверяет валидность типа метрики
func validateMetricType(id, mtype string) error {
	if id == "" {
		return domain.ErrMetricNotFound
	}
	if mtype != domain.Counter && mtype != domain.Gauge {
		return domain.ErrInvalidMetricType
	}
	return nil
}

type ParamMetric struct {
	ID    string `json:"id"`
	MType string `json:"type"`
	Value string `json:"value"`
}

func (m *ParamMetric) Validate() (*domain.Metrics, error) {
	if err := validateMetricType(m.ID, m.MType); err != nil {
		return nil, err
	}

	metric := &domain.Metrics{ID: m.ID, MType: m.MType}

	switch m.MType {
	case domain.Counter:
		if m.Value == "" {
			return nil, domain.ErrInvalidMetricValue
		}
		delta, err := strconv.ParseInt(m.Value, 10, 64)
		if err != nil {
			return nil, domain.ErrInvalidMetricValue
		}
		metric.Delta = &delta
	case domain.Gauge:
		value, err := strconv.ParseFloat(m.Value, 64)
		if err != nil {
			return nil, domain.ErrInvalidMetricValue
		}
		metric.Value = &value
	default:
		return nil, domain.ErrInvalidMetricType
	}
	return metric, nil
}

type GetMetricDTO struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

func (gm *GetMetricDTO) Validate() error {
	return validateMetricType(gm.ID, gm.Type)
}

type BodyMetric struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func (m *BodyMetric) Validate() (*domain.Metrics, error) {
	if err := validateMetricType(m.ID, m.MType); err != nil {
		return nil, err
	}

	// Проверяем, что для каждого типа метрики установлено правильное поле
	switch m.MType {
	case domain.Counter:
		if m.Delta == nil {
			return nil, domain.ErrMissingMetricValue
		}
	case domain.Gauge:
		if m.Value == nil {
			return nil, domain.ErrMissingMetricValue
		}
	}

	return &domain.Metrics{
		ID:    m.ID,
		MType: m.MType,
		Value: m.Value,
		Delta: m.Delta,
	}, nil
}
