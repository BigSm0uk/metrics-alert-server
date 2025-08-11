package app

import (
	"io"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/zl"
	"github.com/bigsm0uk/metrics-alert-server/internal/config"
	"github.com/bigsm0uk/metrics-alert-server/internal/handler"
	"github.com/bigsm0uk/metrics-alert-server/internal/interfaces"
	"github.com/bigsm0uk/metrics-alert-server/internal/repository"
	"github.com/bigsm0uk/metrics-alert-server/internal/server/store"
	"github.com/bigsm0uk/metrics-alert-server/internal/service"
)

// Container представляет DI контейнер для управления зависимостями
type Container struct {
	config     *config.ServerConfig
	repository interfaces.MetricsRepository
	store      interfaces.MetricsStore
	service    *service.MetricService
	handler    *handler.MetricHandler
	server     *Server
}

// NewContainer создает новый DI контейнер
func NewContainer() *Container {
	return &Container{}
}

// LoadConfig загружает конфигурацию
func (c *Container) LoadConfig() *Container {
	cfg, err := config.LoadServerConfig()
	if err != nil {
		panic(err)
	}
	c.config = cfg
	return c
}

// InitLogger инициализирует логгер
func (c *Container) InitLogger() *Container {
	zl.InitLogger(c.config.Env)
	return c
}

// InitRepository инициализирует репозиторий
func (c *Container) InitRepository() *Container {
	repo, err := repository.InitRepository(c.config)
	if err != nil {
		panic(err)
	}
	c.repository = repo
	return c
}

// InitStore инициализирует хранилище
func (c *Container) InitStore() *Container {
	st, err := store.InitStore(c.repository, &c.config.Store)
	if err != nil {
		panic(err)
	}
	c.store = st
	return c
}

// InitService инициализирует сервис
func (c *Container) InitService() *Container {
	c.service = service.NewService(c.repository, c.store)
	return c
}

// InitHandler инициализирует обработчик
func (c *Container) InitHandler() *Container {
	c.handler = handler.NewMetricHandler(c.service, c.config.TemplatePath)
	return c
}

// RestoreData восстанавливает данные из хранилища
func (c *Container) RestoreData() *Container {
	if c.config.Store.Restore {
		if err := c.store.Restore(); err != nil && err != io.EOF {
			panic(err)
		}
	}
	return c
}

// Build создает готовое приложение
func (c *Container) Build() (*Server, error) {
	c.server = NewServer(c.config, c.handler, c.store)
	return c.server, nil
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
