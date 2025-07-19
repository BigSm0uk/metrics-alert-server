package handler

import (
	"encoding/json"
	"net/http"

	"github.com/bigsm0uk/metrics-alert-server/internal/service"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *service.MetricService
}

func NewHandler(service *service.MetricService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) UpdateMetrics(w http.ResponseWriter, r *http.Request) {
	t := chi.URLParam(r, "type")
	id := chi.URLParam(r, "id")
	value := chi.URLParam(r, "value")

	err := h.service.UpdateMetric(t, id, value)

	switch err {
	case service.ErrInvalidMetricType:
		{
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(err.Error()))
			return
		}
	case service.ErrInvalidMetricValue:
		{
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
	}
	w.Write([]byte("Metric updated"))
}
func (h *Handler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
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
