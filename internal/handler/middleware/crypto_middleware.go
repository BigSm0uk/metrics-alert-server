package middleware

import (
	"bytes"
	"crypto/rsa"
	"io"
	"net/http"

	"github.com/bigsm0uk/metrics-alert-server/pkg/util/crypto"
)

const maxEncryptedBodySize = 10 << 20 // 10 мегабайт (10 * 2^20 = 10_485_760 байт)

func WithDecryption(privateKey *rsa.PrivateKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if privateKey == nil {
				next.ServeHTTP(w, r)
				return
			}

			if r.Header.Get("Content-Encryption") != "rsa" {
				next.ServeHTTP(w, r)
				return
			}

			r.Body = http.MaxBytesReader(w, r.Body, maxEncryptedBodySize)
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				return
			}
			r.Body.Close()

			// Расшифровываем данные
			decryptedData, err := crypto.Decrypt(body, privateKey)
			if err != nil {
				http.Error(w, "Failed to decrypt request body", http.StatusBadRequest)
				return
			}

			// Восстанавливаем тело запроса для последующих обработчиков
			r.Body = io.NopCloser(bytes.NewBuffer(decryptedData))

			next.ServeHTTP(w, r)
		})
	}
}
