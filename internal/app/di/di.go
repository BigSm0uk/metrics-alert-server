package di

import (
	"github.com/bigsm0uk/metrics-alert-server/internal/service"
	"go.uber.org/zap"
)

type Container struct {
	Logger  *zap.Logger
	Service *service.MetricService
}

func NewContainer(logger *zap.Logger, service *service.MetricService) *Container {
	return &Container{Logger: logger, Service: service}
}
