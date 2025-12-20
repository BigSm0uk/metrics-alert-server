package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "github.com/bigsm0uk/metrics-alert-server/api/proto"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/semaphore"
	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
)

type GRPCMetricsSender struct {
	client  pb.MetricsClient
	conn    *grpc.ClientConn
	logger  *zap.Logger
	localIP string
}

func NewGRPCMetricsSender(serverAddr string, logger *zap.Logger) (*GRPCMetricsSender, error) {
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	client := pb.NewMetricsClient(conn)
	localIP := getLocalIP()
	if localIP != "" {
		logger.Info("local IP detected for gRPC", zap.String("ip", localIP))
	}

	return &GRPCMetricsSender{
		client:  client,
		conn:    conn,
		logger:  logger,
		localIP: localIP,
	}, nil
}

func (s *GRPCMetricsSender) Close() error {
	return s.conn.Close()
}

func (s *GRPCMetricsSender) SendMetrics(ctx context.Context, metrics []domain.Metrics) error {
	if len(metrics) == 0 {
		s.logger.Debug("no metrics to send, skipping")
		return nil
	}

	pbMetrics := make([]*pb.Metric, 0, len(metrics))
	for _, m := range metrics {
		pbMetric := &pb.Metric{
			Id: m.ID,
		}

		switch m.MType {
		case domain.Gauge:
			pbMetric.Type = pb.Metric_GAUGE
			if m.Value != nil {
				pbMetric.Value = *m.Value
			}
		case domain.Counter:
			pbMetric.Type = pb.Metric_COUNTER
			if m.Delta != nil {
				pbMetric.Delta = *m.Delta
			}
		}

		pbMetrics = append(pbMetrics, pbMetric)
	}

	req := &pb.UpdateMetricsRequest{
		Metrics: pbMetrics,
	}

	if s.localIP != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-real-ip", s.localIP)
	}

	_, err := s.client.UpdateMetrics(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to send metrics via gRPC: %w", err)
	}

	s.logger.Debug("metrics sent via gRPC", zap.Int("count", len(metrics)))
	return nil
}

func (s *GRPCMetricsSender) RunProcess(ctx context.Context, wg *sync.WaitGroup, reportInterval uint, collector Collector, sem *semaphore.Semaphore) {
	ticker := time.NewTicker(time.Duration(reportInterval) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			metrics := collector.GetMetrics()
			wg.Add(1)
			go func() {
				defer wg.Done()
				sem.Acquire()
				defer sem.Release()

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				if err := s.SendMetrics(ctx, metrics); err != nil {
					s.logger.Error("failed to send metrics via gRPC", zap.Error(err))
				}
			}()
		}
	}
}
