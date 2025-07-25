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
	dto := &MetricDTO{
		ID:    chi.URLParam(r, "id"),
		Type:  chi.URLParam(r, "type"),
		Value: chi.URLParam(r, "value"),
	}
	if err := dto.Validate(); err != nil {
		if err == service.ErrMetricNotFound {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(err.Error()))

			return
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
	}

	err := h.service.UpdateMetric(dto.ID, dto.Type, dto.Value)

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
