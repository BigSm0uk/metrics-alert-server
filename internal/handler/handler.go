package handler

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/goccy/go-json"
	"go.uber.org/zap"

	"github.com/bigsm0uk/metrics-alert-server/api/templates"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/zl"
	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
	"github.com/bigsm0uk/metrics-alert-server/internal/service"
	oapiMetric "github.com/bigsm0uk/metrics-alert-server/pkg/openapi/metric"
)

type MetricHandler struct {
	oapiMetric.Unimplemented
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

	err = h.service.UpdateMetric(ctx, m)
	if err != nil {
		handleBadRequest(w, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Metric updated"))
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

	err = h.service.UpdateMetric(ctx, m)
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedMetric)
}

// GetAllMetrics отдает html с табличным представлением всех метрик
func (h *MetricHandler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	m, err := h.service.GetAllMetrics(ctx)
	if err != nil {
		zl.Log.Error("failed to get all metrics", zap.Error(err))
		handleInternal(w)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if err := h.tmpl.Execute(w, m); err != nil {
		zl.Log.Error("failed to execute template", zap.Error(err))
		handleInternal(w)
		return
	}
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(m)
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

		err = h.service.UpdateMetric(ctx, m)
		if err != nil {
			handleBadRequest(w, err.Error())
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Successfully updated %d metrics", len(metrics))
}

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

func (h *MetricHandler) Close() error {
	return h.service.Close()
}

func handleInternal(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(oapiMetric.InternalServerError{
		Code:    http.StatusInternalServerError,
		Message: http.StatusText(http.StatusInternalServerError),
	})
}

func handleBadRequest(w http.ResponseWriter, errText string) {
	w.Header().Set("Content-Type", "application/json")

	message := http.StatusText(http.StatusBadRequest)
	if errText != "" {
		message = errText
	}

	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(oapiMetric.BadRequestError{
		Code:    http.StatusBadRequest,
		Message: message,
	})
}

func handleNotFound(w http.ResponseWriter, errText string) {
	w.Header().Set("Content-Type", "application/json")

	message := http.StatusText(http.StatusBadRequest)
	if errText != "" {
		message = errText
	}
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(oapiMetric.BadRequestError{
		Code:    http.StatusNotFound,
		Message: message,
	})
}
