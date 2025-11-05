package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/goccy/go-json"
	"go.uber.org/zap"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/zl"
	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
	oapiMetric "github.com/bigsm0uk/metrics-alert-server/pkg/openapi/metric"
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
	h.notifyAudit(r.RemoteAddr, m)
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
	h.notifyAudit(r.RemoteAddr, updatedMetric)
}

// UpdateOrCreateMetricsBatch Обновляет/сохраняет метрики batch запросов
func (h *MetricHandler) UpdateOrCreateMetricsBatch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var bodyMetrics []BodyMetric

	if err := json.NewDecoder(r.Body).Decode(&bodyMetrics); err != nil {
		handleBadRequest(w, err.Error())
		return
	}

	metrics := make([]*domain.Metrics, len(bodyMetrics))

	for i, bodyMetric := range bodyMetrics {
		m, err := bodyMetric.Validate()
		if err != nil {
			if err == domain.ErrMissingMetricValue {
				handleBadRequest(w, fmt.Sprintf("missing value for metric %s", bodyMetric.ID))
			} else {
				handleBadRequest(w, fmt.Sprintf("invalid metric %s: %s", bodyMetric.ID, err.Error()))
			}
			return
		}

		metrics[i] = m
	}
	err := h.service.SaveOrUpdateMetricsBatch(ctx, metrics)
	if err != nil {
		handleBadRequest(w, err.Error())
		return
	}
	jsonWithHashValueHandler(w, metrics, h.key)
	h.notifyAudit(r.RemoteAddr, metrics...)
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

func (h *MetricHandler) notifyAudit(ip string, metrics ...*domain.Metrics) {
	auditMessage := domain.AuditMessage{
		TS:      time.Now().Unix(),
		Metrics: make([]string, len(metrics)),
		IPAddr:  ip,
	}
	for i, metric := range metrics {
		auditMessage.Metrics[i] = metric.ID
	}
	h.as.NotifyAll(auditMessage)
}
