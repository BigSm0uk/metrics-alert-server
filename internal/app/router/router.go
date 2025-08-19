package router

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/bigsm0uk/metrics-alert-server/internal/handler"
	lm "github.com/bigsm0uk/metrics-alert-server/internal/handler/middleware"
)

func NewRouter(h *handler.MetricHandler) *chi.Mux {
	r := chi.NewRouter()
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

	MapRoutes(r, h)

	return r
}

func MapRoutes(r *chi.Mux, h *handler.MetricHandler) {
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	r.Route("/update", func(r chi.Router) {
		r.Post("/", h.UpdateMetricByBody)
		r.Post("/{type}/{id}/{value}", h.UpdateMetricByParam)
	})
	r.Route("/updates", func(r chi.Router) {
		r.Post("/", h.UpdateMetricsBatch)
	})
	r.Get("/", h.GetAllMetrics)
	r.Route("/value", func(r chi.Router) {
		r.Post("/", h.EnrichMetric)
		r.Get("/{type}/{id}", h.GetMetric)
	})
	r.Get("/ping", h.Ping)

	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		http.ServeFile(w, r, "docs/redoc.html")
	})
	r.Get("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/yaml")
		http.ServeFile(w, r, "api/openapi.yaml")
	})
}
