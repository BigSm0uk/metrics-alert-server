package di

import (
	"github.com/bigsm0uk/metrics-alert-server/internal/handler"
	"github.com/bigsm0uk/metrics-alert-server/internal/service"
	"go.uber.org/zap"
)

type Container struct {
	Logger  *zap.Logger
	Service *service.MetricService
	Handler *handler.Handler
}

func NewContainer(logger *zap.Logger, service *service.MetricService, handler *handler.Handler) *Container {
	return &Container{Logger: logger, Service: service, Handler: handler}
}
