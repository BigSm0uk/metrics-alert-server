package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/bigsm0uk/metrics-alert-server/pkg/util/hasher"
)

func HashHandlerMiddleware(next http.Handler, key string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if key == "" {
			next.ServeHTTP(w, r)
			return
		}
		hash := r.Header.Get("HashSHA256")
		if hash == "" {
			http.Error(w, "HashSHA256 header is required", http.StatusBadRequest)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusInternalServerError)
			return
		}

		// Восстанавливаем тело для следующего обработчика
		r.Body = io.NopCloser(bytes.NewReader(body))

		if !hasher.VerifyHash(string(body), key, hash) {
			http.Error(w, "Invalid hash", http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}
