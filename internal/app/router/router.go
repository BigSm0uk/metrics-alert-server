package router

import (
	"net/http"
	"time"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/di"
	lm "github.com/bigsm0uk/metrics-alert-server/internal/handler/middleware"
	"github.com/bigsm0uk/metrics-alert-server/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(container *di.Container) *chi.Mux {
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
	r.Use(func(next http.Handler) http.Handler {
		return lm.LoggerMiddleware(next, container.Logger)
	})

	MapRoutes(r, container.Service)

	return r
}

func MapRoutes(r *chi.Mux, service *service.MetricService) {
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	r.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Metrics"))
	})

	r.Post("/update/{type}/{id}/{value}", func(w http.ResponseWriter, r *http.Request) {
		t := chi.URLParam(r, "type")
		id := chi.URLParam(r, "id")
		value := chi.URLParam(r, "value")

		service.UpdateMetric(t, id, value)

		w.Write([]byte("Metric updated"))
	})
}
