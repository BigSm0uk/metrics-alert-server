package handler

import (
	"html/template"
	"net/http"

	"github.com/goccy/go-json"
	"go.uber.org/zap"

	"github.com/bigsm0uk/metrics-alert-server/api/templates"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/zl"
	"github.com/bigsm0uk/metrics-alert-server/internal/service"
	oapiMetric "github.com/bigsm0uk/metrics-alert-server/pkg/openapi/metric"
	"github.com/bigsm0uk/metrics-alert-server/pkg/util/hasher"
)

// MetricHandler обслуживает HTTP-запросы практического трека метрик.
// Содержит ссылки на сервис метрик, шаблоны для HTML-рендеринга
// и ключ для расчета хеша ответа.
type MetricHandler struct {
	oapiMetric.Unimplemented
	service *service.MetricService
	tmpl    *template.Template
	key     string
	as      *service.AuditService
}

// NewMetricHandler конструирует экземпляр обработчика метрик.
// templatePath — путь к HTML-шаблону; при ошибке используется встроенный дефолтный шаблон.
// key — секрет для заголовка HashSHA256.
// as — сервис аудита.
func NewMetricHandler(service *service.MetricService, templatePath, key string, as *service.AuditService) *MetricHandler {
	tmpl := initializeTemplate(templatePath)

	return &MetricHandler{
		service: service,
		tmpl:    tmpl,
		key:     key,
		as:      as,
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

// Close корректно закрывает зависимости обработчика (репозиторий и др.).
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

func jsonWithHashValueHandler(w http.ResponseWriter, data any, key string) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		zl.Log.Error("failed to marshal response data", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	withHasherValueHandler(w, jsonData, key)

	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(jsonData); err != nil {
		zl.Log.Error("failed to write response", zap.Error(err))
	}
}

func withHasherValueHandler(w http.ResponseWriter, jsonData []byte, key string) {
	if key != "" && len(jsonData) > 0 {
		hash := hasher.Hash(string(jsonData), key)
		w.Header().Set("HashSHA256", hash)
	}
}
