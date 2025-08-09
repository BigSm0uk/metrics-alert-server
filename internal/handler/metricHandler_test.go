package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bigsm0uk/metrics-alert-server/internal/config"
	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
	"github.com/bigsm0uk/metrics-alert-server/internal/repository"
	"github.com/bigsm0uk/metrics-alert-server/internal/server"
	"github.com/bigsm0uk/metrics-alert-server/internal/service"
)

func setupTestServer(t *testing.T) (*httptest.Server, *resty.Client) {
	cfg := config.InitDefaultConfig()

	r, err := repository.InitRepository(cfg)
	require.NoError(t, err)

	ms, err := server.NewMetricStore(r, &cfg.Store)
	require.NoError(t, err)

	svc := service.NewService(r, ms)
	h := NewMetricHandler(svc, cfg.TemplatePath)

	router := chi.NewRouter()
	router.Route("/update", func(r chi.Router) {
		r.Post("/", h.UpdateMetricByBody)
		r.Post("/{type}/{id}/{value}", h.UpdateMetricByParam)
	})
	router.Route("/value", func(r chi.Router) {
		r.Post("/", h.EnrichMetric)
		r.Get("/{type}/{id}", h.GetMetric)
	})

	server := httptest.NewServer(router)
	client := resty.New().SetBaseURL(server.URL)

	return server, client
}

func TestMetricHandler_UpdateMetricByParam(t *testing.T) {
	server, client := setupTestServer(t)
	defer server.Close()

	tests := []struct {
		name       string
		metricType string
		id         string
		value      string
		wantStatus int
	}{
		{
			name:       "valid gauge",
			metricType: domain.Gauge,
			id:         "test_gauge",
			value:      "123.45",
			wantStatus: http.StatusOK,
		},
		{
			name:       "valid counter",
			metricType: domain.Counter,
			id:         "test_counter",
			value:      "100",
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid type",
			metricType: "invalid",
			id:         "test",
			value:      "123",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid gauge value",
			metricType: domain.Gauge,
			id:         "test_gauge",
			value:      "invalid",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.R().
				Post("/update/" + tt.metricType + "/" + tt.id + "/" + tt.value)

			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode())
		})
	}
}

func TestMetricHandler_UpdateMetricByBody(t *testing.T) {
	server, client := setupTestServer(t)
	defer server.Close()

	tests := []struct {
		name       string
		body       interface{}
		wantStatus int
	}{
		{
			name: "valid gauge",
			body: map[string]interface{}{
				"id":    "test_gauge",
				"type":  domain.Gauge,
				"value": 123.45,
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "valid counter",
			body: map[string]interface{}{
				"id":    "test_counter",
				"type":  domain.Counter,
				"delta": int64(100),
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid json",
			body:       "invalid json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing value",
			body: map[string]interface{}{
				"id":   "test",
				"type": domain.Gauge,
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.R().
				SetHeader("Content-Type", "application/json").
				SetBody(tt.body).
				Post("/update")

			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode())
		})
	}
}

func TestMetricHandler_GetMetric(t *testing.T) {
	server, client := setupTestServer(t)
	defer server.Close()

	// Setup test data
	_, err := client.R().Post("/update/gauge/test_gauge/123.45")
	require.NoError(t, err)

	_, err = client.R().Post("/update/counter/test_counter/100")
	require.NoError(t, err)

	tests := []struct {
		name       string
		metricType string
		id         string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "get gauge",
			metricType: domain.Gauge,
			id:         "test_gauge",
			wantStatus: http.StatusOK,
			wantBody:   "123.45",
		},
		{
			name:       "get counter",
			metricType: domain.Counter,
			id:         "test_counter",
			wantStatus: http.StatusOK,
			wantBody:   "100",
		},
		{
			name:       "not found",
			metricType: domain.Gauge,
			id:         "not_found",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.R().
				Get("/value/" + tt.metricType + "/" + tt.id)

			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode())

			if tt.wantStatus == http.StatusOK {
				assert.Equal(t, tt.wantBody, string(resp.Body()))
			}
		})
	}
}

func TestMetricHandler_GetEnrichMetric(t *testing.T) {
	server, client := setupTestServer(t)
	defer server.Close()

	// Setup test data
	_, err := client.R().Post("/update/gauge/test_gauge/123.45")
	require.NoError(t, err)

	tests := []struct {
		name       string
		body       interface{}
		wantStatus int
	}{
		{
			name: "get existing gauge",
			body: map[string]interface{}{
				"id":   "test_gauge",
				"type": domain.Gauge,
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "get non-existing metric",
			body: map[string]interface{}{
				"id":   "not_found",
				"type": domain.Gauge,
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.R().
				SetHeader("Content-Type", "application/json").
				SetBody(tt.body).
				Post("/value")

			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode())

			if tt.wantStatus == http.StatusOK {
				assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
			}
		})
	}
}
