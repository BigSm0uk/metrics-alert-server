package main

import (
	"net/http"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/db"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/di"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/router"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/zl"
	"github.com/bigsm0uk/metrics-alert-server/internal/config"
	"github.com/bigsm0uk/metrics-alert-server/internal/repository"
	"github.com/bigsm0uk/metrics-alert-server/internal/service"
)

func main() {
	// 1. Load config
	cfg := config.MustLoadConfig()
	// 2. Initialize logger
	logger := zl.InitLoggerMust(cfg.Logger)
	defer logger.Sync()
	logger.Info("logger initialized")
	// 3. Initialize database
	storage := db.NewMemStorage()
	repository := repository.NewMemRepository(storage)
	service := service.NewMetricService(repository, logger)
	container := di.NewContainer(logger, service)
	// 4. DI
	r := router.NewRouter(container)
	// 5. Initialize server
	http.ListenAndServe(":8080", r)
	// 6. Shutdown server
	// 7. Shutdown database
	// 8. Shutdown logger
}
