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

	MapRoutes(r, h)

	return r
}

func MapRoutes(r *chi.Mux, h *handler.MetricHandler) {
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	r.Post("/update/{type}/{id}/{value}", h.UpdateMetrics)
	r.Get("/", h.GetAllMetrics)
	r.Get("/value/{type}/{id}", h.GetMetric)
}
