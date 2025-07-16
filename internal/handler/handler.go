package handler

import (
	"net/http"

	"github.com/bigsm0uk/metrics-alert-server/internal/repository"
)

type Handler struct {
	repository *repository.MetricsRepository
}

func NewHandler(repository *repository.MetricsRepository) *Handler {
	return &Handler{repository: repository}
}

func (h *Handler) HandleMetrics(w http.ResponseWriter, r *http.Request) {

}
