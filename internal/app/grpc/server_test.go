package grpc

import (
	"context"
	"testing"

	pb "github.com/bigsm0uk/metrics-alert-server/api/proto"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/storage"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/zl"
	"github.com/bigsm0uk/metrics-alert-server/internal/repository/mem"
	"github.com/bigsm0uk/metrics-alert-server/internal/service"
)

func TestMetricsServer_UpdateMetrics(t *testing.T) {
	logger := zl.NewLogger("test")
	ms := storage.NewMemStorage()
	repo := mem.NewMemRepository(ms)
	svc := service.NewService(repo, nil, logger)

	server := NewMetricsServer(svc)

	tests := []struct {
		name    string
		req     *pb.UpdateMetricsRequest
		wantErr bool
	}{
		{
			name: "valid gauge metric",
			req: &pb.UpdateMetricsRequest{
				Metrics: []*pb.Metric{
					{
						Id:    "test_gauge",
						Type:  pb.Metric_GAUGE,
						Value: 123.45,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid counter metric",
			req: &pb.UpdateMetricsRequest{
				Metrics: []*pb.Metric{
					{
						Id:    "test_counter",
						Type:  pb.Metric_COUNTER,
						Delta: 100,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple metrics",
			req: &pb.UpdateMetricsRequest{
				Metrics: []*pb.Metric{
					{
						Id:    "gauge1",
						Type:  pb.Metric_GAUGE,
						Value: 1.0,
					},
					{
						Id:    "counter1",
						Type:  pb.Metric_COUNTER,
						Delta: 1,
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "empty metrics list",
			req:     &pb.UpdateMetricsRequest{Metrics: []*pb.Metric{}},
			wantErr: true,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, err := server.UpdateMetrics(ctx, tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
