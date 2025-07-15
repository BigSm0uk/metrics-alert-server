package main

import (
	"net/http"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/router"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/zl"
	"github.com/bigsm0uk/metrics-alert-server/internal/config"
)

func main() {
	// 1. Load config
	cfg := config.MustLoadConfig()
	// 2. Initialize logger
	logger := zl.InitLoggerMust(cfg.Logger)
	defer logger.Sync()
	logger.Info("logger initialized")
	// 3. Initialize database
	// 4. DI
	r := router.NewRouter(logger)
	http.ListenAndServe(":8080", r)
	// 5. Initialize server
	// 6. Start server
	// 7. Shutdown server
	// 8. Shutdown database
	// 9. Shutdown logger
}
