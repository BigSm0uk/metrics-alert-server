package middleware

import (
	"bytes"
	"crypto/rsa"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/bigsm0uk/metrics-alert-server/pkg/util/crypto"
)

func WithDecryption(privateKey *rsa.PrivateKey, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Если приватный ключ не задан, пропускаем расшифровку
		if privateKey == nil {
			next.ServeHTTP(w, r)
			return
		}

		// Проверяем заголовок Content-Encryption
		if r.Header.Get("Content-Encryption") != "rsa" {
			next.ServeHTTP(w, r)
			return
		}

		// Читаем тело запроса
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error("Failed to read request body for decryption", zap.Error(err))
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		r.Body.Close()

		// Расшифровываем данные
		decryptedData, err := crypto.Decrypt(body, privateKey)
		if err != nil {
			logger.Error("Failed to decrypt request body", zap.Error(err))
			http.Error(w, "Failed to decrypt request body", http.StatusBadRequest)
			return
		}

		// Восстанавливаем тело запроса для последующих обработчиков
		r.Body = io.NopCloser(bytes.NewBuffer(decryptedData))

		next.ServeHTTP(w, r)
		})
	}
}

