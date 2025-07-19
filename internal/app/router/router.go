package router

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/di"
	"github.com/bigsm0uk/metrics-alert-server/internal/handler"
	lm "github.com/bigsm0uk/metrics-alert-server/internal/handler/middleware"

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

	MapRoutes(r, container.Handler)

	return r
}

func MapRoutes(r *chi.Mux, handler *handler.Handler) {
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	r.Post("/update/{type}/{id}/{value}", handler.UpdateMetrics)
	r.Get("/", handler.GetAllMetrics)
}
