package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

const minCompressibleSize = 1400

// gzipResponseWriter оборачивает http.ResponseWriter для сжатия ответов
type gzipResponseWriter struct {
	http.ResponseWriter
	writer      *gzip.Writer
	wroteHeader bool
	size        int
}

func (g *gzipResponseWriter) WriteHeader(code int) {
	if g.wroteHeader {
		return
	}
	g.wroteHeader = true

	contentType := g.Header().Get("Content-Type")
	if !shouldCompress(contentType) {
		g.ResponseWriter.WriteHeader(code)
		return
	}

	g.Header().Set("Content-Encoding", "gzip")
	g.Header().Add("Vary", "Accept-Encoding")
	g.ResponseWriter.WriteHeader(code)
}

func (g *gzipResponseWriter) Write(data []byte) (int, error) {
	if !g.wroteHeader {
		g.WriteHeader(200)
	}

	g.size += len(data)

	if g.Header().Get("Content-Encoding") == "gzip" {
		return g.writer.Write(data)
	}

	return g.ResponseWriter.Write(data)
}

func (g *gzipResponseWriter) Close() error {
	return g.writer.Close()
}

// shouldCompress проверяет, нужно ли сжимать данный Content-Type
func shouldCompress(contentType string) bool {
	compressibleTypes := []string{
		"text/html",
		"text/plain",
		"text/css",
		"text/xml",
		"application/json",
		"application/javascript",
		"application/xml",
	}

	for _, cType := range compressibleTypes {
		if strings.Contains(contentType, cType) {
			return true
		}
	}
	return false
}

// Пул для переиспользования gzip writers
var gzipWriterPool = sync.Pool{
	New: func() any {
		return gzip.NewWriter(io.Discard)
	},
}

// GzipDecompressMiddleware middleware для автоматической декомпрессии gzip запросов
func GzipDecompressMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gzipReader, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Failed to decompress gzip content", http.StatusBadRequest)
				return
			}
			defer gzipReader.Close()
			defer r.Body.Close()

			r.Body = io.NopCloser(gzipReader)

			r.Header.Del("Content-Encoding")
		}

		next.ServeHTTP(w, r)
	})
}

// GzipCompressMiddleware middleware для автоматического сжатия ответов
func GzipCompressMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gzipWriter := gzipWriterPool.Get().(*gzip.Writer)
		defer gzipWriterPool.Put(gzipWriter)

		// Сбрасываем и настраиваем writer
		gzipWriter.Reset(w)
		defer gzipWriter.Close()

		gzipRW := &gzipResponseWriter{
			ResponseWriter: w,
			writer:         gzipWriter,
		}

		next.ServeHTTP(gzipRW, r)
	})
}
