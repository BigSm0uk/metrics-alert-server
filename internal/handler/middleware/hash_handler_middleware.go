package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/zl"
	"github.com/bigsm0uk/metrics-alert-server/pkg/util/hasher"
	"go.uber.org/zap"
)

func HashHandlerMiddleware(next http.Handler, key string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if key == "" {
			next.ServeHTTP(w, r)
			return
		}
		hash := r.Header.Get("HashSHA256")
		if hash == "" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			zl.Log.Error("failed to read body", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Восстанавливаем тело для следующего обработчика
		r.Body = io.NopCloser(bytes.NewReader(body))

		if !hasher.VerifyHash(string(body), key, hash) {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}
