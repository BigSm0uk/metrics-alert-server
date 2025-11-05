package handler

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/config/audit"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/storage"
	"github.com/bigsm0uk/metrics-alert-server/internal/repository/mem"
	"github.com/bigsm0uk/metrics-alert-server/internal/service"
	"go.uber.org/zap"
)

// ExampleMetricHandler_UpdateOrCreateMetricByBody демонстрирует обновление одной метрики через body.
func ExampleMetricHandler_UpdateOrCreateMetricByBody() {
	// Подготовка зависимостей: in-memory storage и репозиторий
	memStore := storage.NewMemStorage()
	repo := mem.NewMemRepository(memStore)
	svc := service.NewService(repo, nil)

	// Аудит (выключен, чтобы не мешал примеру)
	as := service.NewAuditService(&audit.AuditConfig{AuditURL: "", AuditFile: ""}, zap.NewNop())

	h := NewMetricHandler(svc, "api/templates/metrics.html", "", as)

	// Тело запроса: counter метрика
	body := `{"id":"requests","type":"counter","delta":5}`
	req := httptest.NewRequest(http.MethodPost, "/update/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.UpdateOrCreateMetricByBody(rr, req)

	// Выведем код ответа и часть тела
	out := rr.Body.String()
	if len(out) > 60 {
		out = out[:60]
	}
	fmt.Println(rr.Code, out)
	// Output:
	// 200 {"id":"requests","type":"counter","delta":5}
}

// ExampleMetricHandler_UpdateOrCreateMetricsBatch демонстрирует batch обновление метрик.
func ExampleMetricHandler_UpdateOrCreateMetricsBatch() {
	memStore := storage.NewMemStorage()
	repo := mem.NewMemRepository(memStore)
	svc := service.NewService(repo, nil)
	as := service.NewAuditService(&audit.AuditConfig{AuditURL: "", AuditFile: ""}, zap.NewNop())

	h := NewMetricHandler(svc, "api/templates/metrics.html", "", as)

	body := `[
      {"id":"requests","type":"counter","delta":2},
      {"id":"cpu","type":"gauge","value":0.9}
    ]`
	req := httptest.NewRequest(http.MethodPost, "/updates/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.UpdateOrCreateMetricsBatch(rr, req)

	fmt.Println(rr.Code)
	// Output:
	// 200
}

// ExampleMetricHandler_GetAllMetrics демонстрирует получение HTML страницы со списком метрик.
func ExampleMetricHandler_GetAllMetrics() {
	memStore := storage.NewMemStorage()
	repo := mem.NewMemRepository(memStore)
	svc := service.NewService(repo, nil)
	as := service.NewAuditService(&audit.AuditConfig{AuditURL: "", AuditFile: ""}, zap.NewNop())

	h := NewMetricHandler(svc, "api/templates/metrics.html", "", as)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	h.GetAllMetrics(rr, req)

	// Проверим, что вернулся HTML
	fmt.Println(rr.Code, hasPrefix(rr.Body.Bytes(), []byte("<!DOCTYPE html>")))
	// Output:
	// 200 true
}

// Вспомогательные функции для примера
func hasPrefix(b, pref []byte) bool {
	return bytes.HasPrefix(b, pref)
}
