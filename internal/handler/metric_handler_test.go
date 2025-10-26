package handler

// import (
// 	"context"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"
// 	"time"

// 	"github.com/go-chi/chi/v5"
// 	"github.com/go-resty/resty/v2"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"

// 	"github.com/bigsm0uk/metrics-alert-server/internal/app/config"
// 	"github.com/bigsm0uk/metrics-alert-server/internal/app/server/store"
// 	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
// 	"github.com/bigsm0uk/metrics-alert-server/internal/repository"

// 	"github.com/bigsm0uk/metrics-alert-server/internal/service"
// )

// func setupTestServer(t *testing.T) (*httptest.Server, *resty.Client) {
// 	cfg := config.InitDefaultConfig()
// 	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
// 	defer cancel()

// 	r, err := repository.InitRepository(ctx, cfg)
// 	require.NoError(t, err)

// 	ms, err := store.NewJSONStore(r, &cfg.Store)
// 	require.NoError(t, err)

// 	svc := service.NewService(r, ms)
// 	h := NewMetricHandler(svc, cfg.TemplatePath)

// 	router := chi.NewRouter()
// 	router.Route("/update", func(r chi.Router) {
// 		r.Post("/", h.UpdateMetricByBody)
// 		r.Post("/{type}/{id}/{value}", h.UpdateMetricByParam)
// 	})
// 	router.Route("/updates", func(r chi.Router) {
// 		r.Post("/", h.UpdateMetricsBatch)
// 	})
// 	router.Route("/value", func(r chi.Router) {
// 		r.Post("/", h.EnrichMetric)
// 		r.Get("/{type}/{id}", h.GetMetric)
// 	})

// 	server := httptest.NewServer(router)
// 	client := resty.New().SetBaseURL(server.URL)

// 	return server, client
// }

// func TestMetricHandler_UpdateMetricByParam(t *testing.T) {
// 	server, client := setupTestServer(t)
// 	defer server.Close()

// 	tests := []struct {
// 		name       string
// 		metricType string
// 		id         string
// 		value      string
// 		wantStatus int
// 	}{
// 		{
// 			name:       "valid gauge",
// 			metricType: domain.Gauge,
// 			id:         "test_gauge",
// 			value:      "123.45",
// 			wantStatus: http.StatusOK,
// 		},
// 		{
// 			name:       "valid counter",
// 			metricType: domain.Counter,
// 			id:         "test_counter",
// 			value:      "100",
// 			wantStatus: http.StatusOK,
// 		},
// 		{
// 			name:       "invalid type",
// 			metricType: "invalid",
// 			id:         "test",
// 			value:      "123",
// 			wantStatus: http.StatusBadRequest,
// 		},
// 		{
// 			name:       "invalid gauge value",
// 			metricType: domain.Gauge,
// 			id:         "test_gauge",
// 			value:      "invalid",
// 			wantStatus: http.StatusBadRequest,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			resp, err := client.R().
// 				Post("/update/" + tt.metricType + "/" + tt.id + "/" + tt.value)

// 			require.NoError(t, err)
// 			assert.Equal(t, tt.wantStatus, resp.StatusCode())
// 		})
// 	}
// }

// func TestMetricHandler_UpdateMetricByBody(t *testing.T) {
// 	server, client := setupTestServer(t)
// 	defer server.Close()

// 	tests := []struct {
// 		name       string
// 		body       interface{}
// 		wantStatus int
// 	}{
// 		{
// 			name: "valid gauge",
// 			body: map[string]interface{}{
// 				"id":    "test_gauge",
// 				"type":  domain.Gauge,
// 				"value": 123.45,
// 			},
// 			wantStatus: http.StatusOK,
// 		},
// 		{
// 			name: "valid counter",
// 			body: map[string]interface{}{
// 				"id":    "test_counter",
// 				"type":  domain.Counter,
// 				"delta": int64(100),
// 			},
// 			wantStatus: http.StatusOK,
// 		},
// 		{
// 			name:       "invalid json",
// 			body:       "invalid json",
// 			wantStatus: http.StatusBadRequest,
// 		},
// 		{
// 			name: "missing value",
// 			body: map[string]interface{}{
// 				"id":   "test",
// 				"type": domain.Gauge,
// 			},
// 			wantStatus: http.StatusBadRequest,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			resp, err := client.R().
// 				SetHeader("Content-Type", "application/json").
// 				SetBody(tt.body).
// 				Post("/update")

// 			require.NoError(t, err)
// 			assert.Equal(t, tt.wantStatus, resp.StatusCode())
// 		})
// 	}
// }

// func TestMetricHandler_GetMetric(t *testing.T) {
// 	server, client := setupTestServer(t)
// 	defer server.Close()

// 	// Setup test data
// 	_, err := client.R().Post("/update/gauge/test_gauge/123.45")
// 	require.NoError(t, err)

// 	_, err = client.R().Post("/update/counter/test_counter/100")
// 	require.NoError(t, err)

// 	tests := []struct {
// 		name       string
// 		metricType string
// 		id         string
// 		wantStatus int
// 		wantBody   string
// 	}{
// 		{
// 			name:       "get gauge",
// 			metricType: domain.Gauge,
// 			id:         "test_gauge",
// 			wantStatus: http.StatusOK,
// 			wantBody:   "123.45",
// 		},
// 		{
// 			name:       "get counter",
// 			metricType: domain.Counter,
// 			id:         "test_counter",
// 			wantStatus: http.StatusOK,
// 			wantBody:   "100",
// 		},
// 		{
// 			name:       "not found",
// 			metricType: domain.Gauge,
// 			id:         "not_found",
// 			wantStatus: http.StatusNotFound,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			resp, err := client.R().
// 				Get("/value/" + tt.metricType + "/" + tt.id)

// 			require.NoError(t, err)
// 			assert.Equal(t, tt.wantStatus, resp.StatusCode())

// 			if tt.wantStatus == http.StatusOK {
// 				assert.Equal(t, tt.wantBody, string(resp.Body()))
// 			}
// 		})
// 	}
// }

// func TestMetricHandler_GetEnrichMetric(t *testing.T) {
// 	server, client := setupTestServer(t)
// 	defer server.Close()

// 	// Setup test data
// 	_, err := client.R().Post("/update/gauge/test_gauge/123.45")
// 	require.NoError(t, err)

// 	tests := []struct {
// 		name       string
// 		body       interface{}
// 		wantStatus int
// 	}{
// 		{
// 			name: "get existing gauge",
// 			body: map[string]interface{}{
// 				"id":   "test_gauge",
// 				"type": domain.Gauge,
// 			},
// 			wantStatus: http.StatusOK,
// 		},
// 		{
// 			name: "get non-existing metric",
// 			body: map[string]interface{}{
// 				"id":   "not_found",
// 				"type": domain.Gauge,
// 			},
// 			wantStatus: http.StatusNotFound,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			resp, err := client.R().
// 				SetHeader("Content-Type", "application/json").
// 				SetBody(tt.body).
// 				Post("/value")

// 			require.NoError(t, err)
// 			assert.Equal(t, tt.wantStatus, resp.StatusCode())

// 			if tt.wantStatus == http.StatusOK {
// 				assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
// 			}
// 		})
// 	}
// }

// func TestMetricHandler_UpdateMetricsBatch(t *testing.T) {
// 	server, client := setupTestServer(t)
// 	defer server.Close()

// 	tests := []struct {
// 		name       string
// 		body       interface{}
// 		wantStatus int
// 		wantBody   string
// 	}{
// 		{
// 			name: "valid batch with mixed metrics",
// 			body: []map[string]interface{}{
// 				{
// 					"id":    "gauge1",
// 					"type":  domain.Gauge,
// 					"value": 123.45,
// 				},
// 				{
// 					"id":    "counter1",
// 					"type":  domain.Counter,
// 					"delta": int64(100),
// 				},
// 				{
// 					"id":    "gauge2",
// 					"type":  domain.Gauge,
// 					"value": 678.90,
// 				},
// 			},
// 			wantStatus: http.StatusOK,
// 			wantBody:   "Successfully updated 3 metrics",
// 		},
// 		{
// 			name:       "empty batch",
// 			body:       []map[string]interface{}{},
// 			wantStatus: http.StatusOK,
// 			wantBody:   "Successfully updated 0 metrics",
// 		},
// 		{
// 			name: "single metric in batch",
// 			body: []map[string]interface{}{
// 				{
// 					"id":    "single_gauge",
// 					"type":  domain.Gauge,
// 					"value": 42.0,
// 				},
// 			},
// 			wantStatus: http.StatusOK,
// 			wantBody:   "Successfully updated 1 metrics",
// 		},
// 		{
// 			name:       "invalid json",
// 			body:       "invalid json string",
// 			wantStatus: http.StatusBadRequest,
// 			wantBody:   "invalid JSON",
// 		},
// 		{
// 			name: "batch with invalid metric type",
// 			body: []map[string]interface{}{
// 				{
// 					"id":    "valid_gauge",
// 					"type":  domain.Gauge,
// 					"value": 123.45,
// 				},
// 				{
// 					"id":    "invalid_metric",
// 					"type":  "invalid_type",
// 					"value": 456.78,
// 				},
// 			},
// 			wantStatus: http.StatusBadRequest,
// 			wantBody:   "invalid metric invalid_metric: invalid metric type",
// 		},
// 		{
// 			name: "batch with missing value for gauge",
// 			body: []map[string]interface{}{
// 				{
// 					"id":   "missing_value",
// 					"type": domain.Gauge,
// 					// missing value field
// 				},
// 			},
// 			wantStatus: http.StatusBadRequest,
// 			wantBody:   "missing value for metric missing_value",
// 		},
// 		{
// 			name: "batch with missing delta for counter",
// 			body: []map[string]interface{}{
// 				{
// 					"id":   "missing_delta",
// 					"type": domain.Counter,
// 					// missing delta field
// 				},
// 			},
// 			wantStatus: http.StatusBadRequest,
// 			wantBody:   "missing value for metric missing_delta",
// 		},
// 		{
// 			name: "batch with empty metric id",
// 			body: []map[string]interface{}{
// 				{
// 					"id":    "",
// 					"type":  domain.Gauge,
// 					"value": 123.45,
// 				},
// 			},
// 			wantStatus: http.StatusBadRequest,
// 			wantBody:   "invalid metric : metric not found",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			resp, err := client.R().
// 				SetHeader("Content-Type", "application/json").
// 				SetBody(tt.body).
// 				Post("/updates")

// 			require.NoError(t, err)
// 			assert.Equal(t, tt.wantStatus, resp.StatusCode())

// 			if tt.wantBody != "" {
// 				assert.Contains(t, string(resp.Body()), tt.wantBody)
// 			}
// 		})
// 	}
// }

// func TestMetricHandler_UpdateMetricsBatch_Integration(t *testing.T) {
// 	server, client := setupTestServer(t)
// 	defer server.Close()

// 	// Send batch update
// 	batchPayload := []map[string]any{
// 		{
// 			"id":    "cpu_usage",
// 			"type":  domain.Gauge,
// 			"value": 85.5,
// 		},
// 		{
// 			"id":    "requests_total",
// 			"type":  domain.Counter,
// 			"delta": int64(150),
// 		},
// 	}

// 	resp, err := client.R().
// 		SetHeader("Content-Type", "application/json").
// 		SetBody(batchPayload).
// 		Post("/updates")

// 	require.NoError(t, err)
// 	assert.Equal(t, http.StatusOK, resp.StatusCode())

// 	// Verify the metrics were actually saved by retrieving them
// 	gaugeResp, err := client.R().Get("/value/gauge/cpu_usage")
// 	require.NoError(t, err)
// 	assert.Equal(t, http.StatusOK, gaugeResp.StatusCode())
// 	assert.Equal(t, "85.5", string(gaugeResp.Body()))

// 	counterResp, err := client.R().Get("/value/counter/requests_total")
// 	require.NoError(t, err)
// 	assert.Equal(t, http.StatusOK, counterResp.StatusCode())
// 	assert.Equal(t, "150", string(counterResp.Body()))

// 	// Send another batch to test counter accumulation
// 	secondBatch := []map[string]any{
// 		{
// 			"id":    "requests_total",
// 			"type":  domain.Counter,
// 			"delta": int64(50),
// 		},
// 	}

// 	resp2, err := client.R().
// 		SetHeader("Content-Type", "application/json").
// 		SetBody(secondBatch).
// 		Post("/updates")

// 	require.NoError(t, err)
// 	assert.Equal(t, http.StatusOK, resp2.StatusCode())

// 	// Verify counter was incremented
// 	counterResp2, err := client.R().Get("/value/counter/requests_total")
// 	require.NoError(t, err)
// 	assert.Equal(t, http.StatusOK, counterResp2.StatusCode())
// 	assert.Equal(t, "200", string(counterResp2.Body())) // 150 + 50
// }
