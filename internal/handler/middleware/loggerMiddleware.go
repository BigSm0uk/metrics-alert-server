package middleware

import (
	"net/http"

	"go.uber.org/zap"
)

func LoggerMiddleware(next http.Handler, logger *zap.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("request", zap.String("method", r.Method), zap.String("path", r.URL.Path), zap.String("remote_addr", r.RemoteAddr))
		next.ServeHTTP(w, r)
	})
}
