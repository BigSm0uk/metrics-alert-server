package handler

import (
	"fmt"
	"net/http"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/zl"
	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
	oapiMetric "github.com/bigsm0uk/metrics-alert-server/pkg/openapi/metric"
	"github.com/goccy/go-json"
	"go.uber.org/zap"
)

// UpdateOrCreateMetricByParam обновляет или создает метрику по query параметрам
func (h *MetricHandler) UpdateOrCreateMetricByParam(w http.ResponseWriter, r *http.Request, mType oapiMetric.UpdateOrCreateMetricByParamParamsType, id oapiMetric.ID, value oapiMetric.Value) {
	ctx := r.Context()

	dto := &ParamMetric{
		ID:    id,
		MType: string(mType),
		Value: value,
	}

	m, err := dto.Validate()
	if err != nil {
		if err == domain.ErrMetricNotFound {
			handleNotFound(w, err.Error())
			return
		} else {
			handleBadRequest(w, err.Error())
			return
		}
	}

	err = h.service.SaveOrUpdateMetric(ctx, m)
	if err != nil {
		handleBadRequest(w, err.Error())
		return
	}
	jsonWithHashValueHandler(w, m, h.key)
}

// UpdateOrCreateMetricByBody обновляет или создает метрику по body запроса
func (h *MetricHandler) UpdateOrCreateMetricByBody(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var dto BodyMetric

	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		handleBadRequest(w, err.Error())
		return
	}

	m, err := dto.Validate()
	if err != nil {
		if err == domain.ErrMetricNotFound {
			handleNotFound(w, err.Error())
			return
		}
		handleBadRequest(w, err.Error())
		return
	}

	err = h.service.SaveOrUpdateMetric(ctx, m)
	if err != nil {
		handleBadRequest(w, err.Error())
		return
	}

	updatedMetric, err := h.service.GetEnrichMetric(ctx, m.ID, m.MType)
	if err != nil {
		zl.Log.Error("failed to get enriched metric", zap.Error(err))
		handleInternal(w)
		return
	}

	jsonWithHashValueHandler(w, updatedMetric, h.key)
}

// UpdateOrCreateMetricsBatch Обновляет/сохраняет метрики batch запросов
func (h *MetricHandler) UpdateOrCreateMetricsBatch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var metrics []BodyMetric

	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		handleBadRequest(w, err.Error())
		return
	}

	for _, metric := range metrics {
		m, err := metric.Validate()
		if err != nil {
			if err == domain.ErrMissingMetricValue {
				handleBadRequest(w, fmt.Sprintf("missing value for metric %s", metric.ID))
			} else {
				handleBadRequest(w, fmt.Sprintf("invalid metric %s: %s", metric.ID, err.Error()))
			}
			return
		}

		err = h.service.SaveOrUpdateMetric(ctx, m)
		if err != nil {
			handleBadRequest(w, err.Error())
			return
		}
	}

	jsonWithHashValueHandler(w, metrics, h.key)
}

// GetValueByBody возвращает метрику по ее типу и id из body запроса
func (h *MetricHandler) GetValueByBody(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var dto GetMetricDTO

	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		handleBadRequest(w, err.Error())
		return
	}

	if err := dto.Validate(); err != nil {
		handleBadRequest(w, err.Error())
		return
	}

	m, err := h.service.GetEnrichMetric(ctx, dto.ID, dto.Type)
	if err != nil {
		handleNotFound(w, err.Error())
		return
	}
	jsonWithHashValueHandler(w, m, h.key)
}
