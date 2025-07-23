package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/bigsm0uk/metrics-alert-server/internal/service"
)

type MetricHandler struct {
	service *service.MetricService
}

func NewMetricHandler(service *service.MetricService) *MetricHandler {
	return &MetricHandler{service: service}
}

func (h *MetricHandler) UpdateMetrics(w http.ResponseWriter, r *http.Request) {
	t := chi.URLParam(r, "type")
	id := chi.URLParam(r, "id")
	value := chi.URLParam(r, "value")

	if id == "" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("metric name is required"))
		return
	}

	err := h.service.UpdateMetric(t, id, value)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Metric updated"))
}
func (h *MetricHandler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	m, err := h.service.GetAllMetrics()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(m); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
