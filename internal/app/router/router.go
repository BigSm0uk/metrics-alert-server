package router

import (
	"net/http"
	"time"

	"github.com/bigsm0uk/metrics-alert-server/internal/handler"
	lm "github.com/bigsm0uk/metrics-alert-server/internal/handler/middleware"
	oapiMetric "github.com/bigsm0uk/metrics-alert-server/pkg/openapi/metric"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(h *handler.MetricHandler, key string) *chi.Mux {
	r := chi.NewRouter()

	// Глобальные middleware
	r.Use(middleware.Recoverer)
	r.Use(middleware.GetHead)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(middleware.CleanPath)
	r.Use(middleware.AllowContentType("application/json", "text/xml"))
	r.Use(middleware.Timeout(time.Second * 60))
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(lm.LoggerMiddleware)
	r.Use(lm.GzipDecompressMiddleware)
	r.Use(lm.GzipCompressMiddleware)

	// Создаем wrapper для handler с hash middleware для определенных маршрутов
	wrapper := &MetricHandlerWrapper{
		handler: h,
		key:     key,
	}

	// Монтируем OpenAPI сгенерированный роутер
	oapiMetric.HandlerFromMux(wrapper, r)

	return r
}

// MetricHandlerWrapper оборачивает MetricHandler и добавляет hash middleware для нужных методов
type MetricHandlerWrapper struct {
	handler *handler.MetricHandler
	key     string
}

func (w *MetricHandlerWrapper) GetAllMetrics(rw http.ResponseWriter, r *http.Request) {
	w.handler.GetAllMetrics(rw, r)
}

func (w *MetricHandlerWrapper) GetDocs(rw http.ResponseWriter, r *http.Request) {
	w.handler.GetDocs(rw, r)
}

func (w *MetricHandlerWrapper) HealthCheck(rw http.ResponseWriter, r *http.Request) {
	w.handler.HealthCheck(rw, r)
}

func (w *MetricHandlerWrapper) GetOpenAPI(rw http.ResponseWriter, r *http.Request) {
	w.handler.GetOpenAPI(rw, r)
}

func (w *MetricHandlerWrapper) Ping(rw http.ResponseWriter, r *http.Request) {
	w.handler.Ping(rw, r)
}

func (w *MetricHandlerWrapper) UpdateOrCreateMetricByBody(rw http.ResponseWriter, r *http.Request) {
	lm.HashHandlerMiddleware(http.HandlerFunc(w.handler.UpdateOrCreateMetricByBody), w.key).ServeHTTP(rw, r)
}

func (w *MetricHandlerWrapper) UpdateOrCreateMetricByParam(rw http.ResponseWriter, r *http.Request, mType oapiMetric.UpdateOrCreateMetricByParamParamsType, id oapiMetric.ID, value oapiMetric.Value) {
	lm.HashHandlerMiddleware(
		http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			w.handler.UpdateOrCreateMetricByParam(writer, req, mType, id, value)
		}),
		w.key,
	).ServeHTTP(rw, r)
}

func (w *MetricHandlerWrapper) UpdateOrCreateMetricsBatch(rw http.ResponseWriter, r *http.Request) {
	lm.HashHandlerMiddleware(http.HandlerFunc(w.handler.UpdateOrCreateMetricsBatch), w.key).ServeHTTP(rw, r)
}

func (w *MetricHandlerWrapper) GetValueByBody(rw http.ResponseWriter, r *http.Request) {
	w.handler.GetValueByBody(rw, r)
}

func (w *MetricHandlerWrapper) GetValueByParam(rw http.ResponseWriter, r *http.Request, mType oapiMetric.GetValueByParamParamsType, id oapiMetric.ID) {
	w.handler.GetValueByParam(rw, r, mType, id)
}
