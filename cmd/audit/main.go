package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/goccy/go-json"
	"go.uber.org/zap"

	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
)

const (
	addr = ":8081"
)

// Отдельное приложение для тестирования аудита
func main() {
	// создать роутер на указанном адресе
	r := chi.NewRouter()
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}
	// создать логгер
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()
	// создать обработчик post запроса на /audit
	r.Post("/audit", func(w http.ResponseWriter, r *http.Request) {
		var auditMessage domain.AuditMessage
		if err := json.NewDecoder(r.Body).Decode(&auditMessage); err != nil {
			logger.Error("failed to unmarshal audit message", zap.Error(err))
			http.Error(w, "failed to unmarshal audit message", http.StatusInternalServerError)
			return
		}
		logger.Info("audit request", zap.Any("auditMessage", auditMessage))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("audit message received"))
	})
	// запустить сервер
	logger.Info("starting server", zap.String("addr", addr))
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("failed to start server", zap.Error(err))
		}
	}()
	// завершить работу при получении сигнала SIGINT или SIGTERM
	quit := make(chan os.Signal, 2)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("failed to shutdown server", zap.Error(err))
	}
	logger.Info("server exiting")
	os.Exit(0)
}
