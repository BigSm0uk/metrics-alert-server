package app

import (
	"context"
	"io"
	"time"

	"go.uber.org/zap"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/audit"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/cache"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/config"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/server/store"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/zl"
	"github.com/bigsm0uk/metrics-alert-server/internal/domain/interfaces"
	"github.com/bigsm0uk/metrics-alert-server/internal/handler"
	"github.com/bigsm0uk/metrics-alert-server/internal/repository"
	"github.com/bigsm0uk/metrics-alert-server/internal/service"
)

// Container представляет DI контейнер для управления зависимостями
type Container struct {
	config       *config.ServerConfig
	logger       *zap.Logger
	repository   interfaces.MetricsRepository
	store        interfaces.MetricsStore
	service      *service.MetricService
	handler      *handler.MetricHandler
	auditService *service.AuditService
	cache        interfaces.MetricsCache
}

// GetRepository возвращает репозиторий (для тестирования)
func (c *Container) GetRepository() interfaces.MetricsRepository {
	return c.repository
}

// GetStore возвращает хранилище (для тестирования)
func (c *Container) GetStore() interfaces.MetricsStore {
	return c.store
}

// GetService возвращает сервис (для тестирования)
func (c *Container) GetService() *service.MetricService {
	return c.service
}

// ContainerOptions представляет функцию для инициализации контейнера
type ContainerOptions func(*Container) error

// NewContainerWithOptions создает новый контейнер с заданными опциями
func NewContainerWithOptions(opts ...ContainerOptions) (*Container, error) {
	c := &Container{}
	for _, opt := range opts {
		err := opt(c)
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

// WithConfig инициализирует конфигурацию
func WithConfig() ContainerOptions {
	return func(c *Container) error {
		cfg, err := config.LoadServerConfig()
		if err != nil {
			return err
		}
		c.config = cfg
		return nil
	}
}

// WithLogger инициализирует логгер
func WithLogger() ContainerOptions {
	return func(c *Container) error {
		c.logger = zl.NewLogger(c.config.Env)
		return nil
	}
}

// WithRepository инициализирует репозиторий
func WithRepository() ContainerOptions {
	return func(c *Container) error {
		repo, err := repository.InitRepository(context.Background(), c.config, c.logger)
		if err != nil {
			return err
		}
		c.repository = repo
		return nil
	}
}

// WithStore инициализирует хранилище
func WithStore() ContainerOptions {
	return func(c *Container) error {
		st, err := store.InitStore(c.repository, &c.config.Store, c.logger)
		if err != nil {
			return err
		}
		c.store = st
		return nil
	}
}

// WithService инициализирует сервис
func WithService() ContainerOptions {
	return func(c *Container) error {
		c.service = service.NewService(c.repository, c.store, c.logger)
		return nil
	}
}

// WithCache инициализирует кеш
func WithCache() ContainerOptions {
	return func(c *Container) error {
		c.cache = cache.NewCache(c.config.Cache.CleanupInterval)
		return nil
	}
}

// WithHandler инициализирует обработчик
func WithHandler() ContainerOptions {
	return func(c *Container) error {
		c.handler = handler.NewMetricHandler(c.service, c.config.TemplatePath, c.config.Key, c.auditService, c.cache, c.logger)
		return nil
	}
}

// WithRestoreData инициализирует восстановление данных
func WithRestoreData() ContainerOptions {
	return func(c *Container) error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if c.config.Store.Restore && c.store.IsActive() {
			if err := c.store.Restore(ctx); err != nil && err != io.EOF {
				return err
			}
		}
		return nil
	}
}

// WithBootstrap инициализирует миграции в базу данных
func WithBootstrap() ContainerOptions {
	return func(c *Container) error {
		if err := c.repository.Bootstrap(context.Background()); err != nil {
			return err
		}
		return nil
	}
}

// WithAuditService инициализирует сервис аудита
func WithAuditService() ContainerOptions {
	return func(c *Container) error {
		c.auditService = service.NewAuditService(&c.config.Audit, c.logger)

		observers := audit.CreateAuditObservers(&c.config.Audit, c.logger)
		c.auditService.Attach(observers...)

		return nil
	}
}

// Build создает новый сервер
func Build(c *Container) *Server {
	return NewServer(c.config, c.handler, c.store, c.auditService, c.logger)
}
