package handler

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/goccy/go-json"

	"github.com/bigsm0uk/metrics-alert-server/api/templates"
	"github.com/bigsm0uk/metrics-alert-server/internal/service"
)

type MetricHandler struct {
	service *service.MetricService
	tmpl    *template.Template
}

func NewMetricHandler(service *service.MetricService, templatePath string) *MetricHandler {
	tmpl := initializeTemplate(templatePath)

	return &MetricHandler{
		service: service,
		tmpl:    tmpl,
	}
}

func initializeTemplate(path string) *template.Template {
	funcMap := template.FuncMap{
		"derefFloat": func(f *float64) float64 {
			if f == nil {
				return 0
			}
			return *f
		},
		"derefInt": func(i *int64) int64 {
			if i == nil {
				return 0
			}
			return *i
		},
	}

	// Пытаемся загрузить из файла
	tmpl, err := template.New("metrics.html").Funcs(funcMap).ParseFiles(path)
	if err != nil {
		// Если ошибка, используем дефолтный шаблон
		tmpl = template.Must(
			template.New("metrics.html").Funcs(funcMap).Parse(templates.DefaultMetricsHTML),
		)
	}

	return tmpl
}

func (h *MetricHandler) UpdateMetricByParam(w http.ResponseWriter, r *http.Request) {
	dto := &ParamMetric{
		ID:    chi.URLParam(r, "id"),
		MType: chi.URLParam(r, "type"),
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

	err := h.service.UpdateMetric(dto.ID, dto.MType, dto.Value)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Metric updated"))
}
func (h *MetricHandler) UpdateMetricByBody(w http.ResponseWriter, r *http.Request) {
	var dto BodyMetric

	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid JSON"))
		return
	}

	if err := dto.Validate(); err != nil {
		if err == service.ErrMetricNotFound {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(err.Error()))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	var value string
	if dto.MType == "gauge" && dto.Value != nil {
		value = fmt.Sprintf("%g", *dto.Value)
	} else if dto.MType == "counter" && dto.Delta != nil {
		value = fmt.Sprintf("%d", *dto.Delta)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing value"))
		return
	}

	err := h.service.UpdateMetric(dto.ID, dto.MType, value)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
func (h *MetricHandler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	m, err := h.service.GetAllMetrics()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if err := h.tmpl.Execute(w, m); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
func (h *MetricHandler) GetMetric(w http.ResponseWriter, r *http.Request) {
	dto := &GetMetricDTO{
		ID:   chi.URLParam(r, "id"),
		Type: chi.URLParam(r, "type"),
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
	m, err := h.service.GetMetric(dto.ID, dto.Type)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	var value string
	if m.MType == "gauge" && m.Value != nil {
		value = fmt.Sprintf("%g", *m.Value)
	} else if m.MType == "counter" && m.Delta != nil {
		value = fmt.Sprintf("%d", *m.Delta)
	}
	w.Write([]byte(value))
}
func (h *MetricHandler) GetEnrichMetric(w http.ResponseWriter, r *http.Request) {
	var dto BodyMetric

	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid JSON"))
		return
	}

	if err := dto.Validate(); err != nil {
		if err == service.ErrMetricNotFound {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(err.Error()))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	m, err := h.service.GetEnrichMetric(dto.ID, dto.MType)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(m)
}
