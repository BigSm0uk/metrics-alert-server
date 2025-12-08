package router

import (
	"crypto/rsa"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"

	"github.com/bigsm0uk/metrics-alert-server/internal/handler"
	lm "github.com/bigsm0uk/metrics-alert-server/internal/handler/middleware"
	oapiMetric "github.com/bigsm0uk/metrics-alert-server/pkg/openapi/metric"
)

// NewRouter создает и настраивает HTTP-роутер chi с middleware и маршрутами OpenAPI.
// key используется для валидации/добавления хеша ответа.
// logger используется для логирования в middleware.
// privateKey используется для расшифровки зашифрованных запросов.
func NewRouter(h *handler.MetricHandler, key string, logger *zap.Logger, privateKey *rsa.PrivateKey) *chi.Mux {
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
	r.Use(lm.LoggerMiddleware(logger))
	r.Use(lm.GzipDecompressMiddleware)
	r.Use(lm.GzipCompressMiddleware)
	r.Use(lm.WithDecryption(privateKey, logger))
	r.Use(lm.WithHashValidation(key, logger))

	// Монтируем OpenAPI сгенерированный роутер
	oapiMetric.HandlerFromMux(h, r)

	return r
}
