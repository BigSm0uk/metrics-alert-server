package handler

import (
	"bytes"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/cache"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/zl"
	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
	oapiMetric "github.com/bigsm0uk/metrics-alert-server/pkg/openapi/metric"
)

// Ping проверяет подключение с БД
func (h *MetricHandler) Ping(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := h.service.Ping(ctx)
	if err != nil {
		zl.Log.Error("database connection failed", zap.Error(err))
		handleInternal(w)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

// HealthCheck проверяет жив ли сервис
func (h *MetricHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// GetDocs возвращает openapi документацию по сервису в формате html
func (h *MetricHandler) GetDocs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeFile(w, r, "api/docs/redoc.html")
}

// GetOpenAPI возвращает openapi документацию по сервису в формате yaml
func (h *MetricHandler) GetOpenAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
	http.ServeFile(w, r, "api/metric/openapi.yaml")
}

// GetAllMetrics отдает html с табличным представлением всех метрик
func (h *MetricHandler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	value, found := h.cache.Get("all_metrics")
	if found {
		w.Write(value.([]byte))
		return
	}

	m, err := h.service.GetAllMetrics(ctx)
	if err != nil {
		zl.Log.Error("failed to get all metrics", zap.Error(err))
		handleInternal(w)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	var buf bytes.Buffer
	if err := h.tmpl.Execute(&buf, m); err != nil {
		zl.Log.Error("failed to execute template", zap.Error(err))
		handleInternal(w)
		return
	}

	h.cache.Set("all_metrics", buf.Bytes(), cache.DefaultExpiration)
	w.Write(buf.Bytes())
}

// GetValueByParam возвращает значение метрики по ее типу и id
func (h *MetricHandler) GetValueByParam(w http.ResponseWriter, r *http.Request, mType oapiMetric.GetValueByParamParamsType, id oapiMetric.ID) {
	ctx := r.Context()

	dto := &GetMetricDTO{
		ID:   id,
		Type: string(mType),
	}
	if err := dto.Validate(); err != nil {
		if err == domain.ErrMetricNotFound {
			handleNotFound(w, err.Error())
			return
		} else {
			handleBadRequest(w, err.Error())
			return
		}
	}
	m, err := h.service.GetMetric(ctx, dto.ID, dto.Type)
	if err != nil {
		handleNotFound(w, err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	var value string
	if m.MType == domain.Gauge && m.Value != nil {
		value = fmt.Sprintf("%g", *m.Value)
	} else if m.MType == domain.Counter && m.Delta != nil {
		value = fmt.Sprintf("%d", *m.Delta)
	}
	w.Write([]byte(value))
}
