package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/bigsm0uk/metrics-alert-server/api/proto"
	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
	"github.com/bigsm0uk/metrics-alert-server/internal/service"
)

type MetricsServer struct {
	pb.UnimplementedMetricsServer
	svc *service.MetricService
}

func NewMetricsServer(svc *service.MetricService) *MetricsServer {
	return &MetricsServer{svc: svc}
}

func (s *MetricsServer) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	if req == nil || len(req.Metrics) == 0 {
		return nil, status.Error(codes.InvalidArgument, "metrics list is empty")
	}

	metrics := make([]*domain.Metrics, 0, len(req.Metrics))
	for _, m := range req.Metrics {
		metric := &domain.Metrics{
			ID: m.Id,
		}

		switch m.Type {
		case pb.Metric_GAUGE:
			metric.MType = domain.Gauge
			metric.Value = &m.Value
		case pb.Metric_COUNTER:
			metric.MType = domain.Counter
			metric.Delta = &m.Delta
		default:
			return nil, status.Errorf(codes.InvalidArgument, "unknown metric type: %v", m.Type)
		}

		metrics = append(metrics, metric)
	}

	if err := s.svc.SaveOrUpdateMetricsBatch(ctx, metrics); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update metrics: %v", err)
	}

	return &pb.UpdateMetricsResponse{}, nil
}
