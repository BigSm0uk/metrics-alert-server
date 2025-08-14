package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/bigsm0uk/metrics-alert-server/internal/config/storage"
	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
	"github.com/bigsm0uk/metrics-alert-server/internal/interfaces"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

var _ interfaces.MetricsRepository = (*PostgresRepository)(nil)

func NewPostgresRepository(ctx context.Context, cfg *storage.StorageConfig) (*PostgresRepository, error) {
	config, err := pgxpool.ParseConfig(cfg.ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	config.MaxConns = cfg.MaxConns
	config.MinConns = cfg.MinConns
	config.MaxConnLifetime = time.Duration(cfg.MaxConnLifetime) * time.Second
	config.MaxConnIdleTime = time.Duration(cfg.MaxConnIdleTime) * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	return &PostgresRepository{pool: pool}, nil
}

// Save сохраняет метрику в базу данных (UPSERT операция)
func (r *PostgresRepository) Save(ctx context.Context, metric *domain.Metrics) error {

	return fmt.Errorf("not implemented")
}

// Get получает метрику по ID и типу
func (r *PostgresRepository) Get(ctx context.Context, id, metricType string) (*domain.Metrics, error) {

	return nil, fmt.Errorf("not implemented")
}

// GetAll получает все метрики
func (r *PostgresRepository) GetAll(ctx context.Context) ([]domain.Metrics, error) {
	return nil, fmt.Errorf("not implemented")
}

// Delete удаляет метрику по ID
func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	return fmt.Errorf("not implemented")
}

// SaveBatch сохраняет множество метрик за одну транзакцию
func (r *PostgresRepository) SaveBatch(ctx context.Context, metrics []domain.Metrics) error {
	return fmt.Errorf("not implemented")
}

// GetByType получает все метрики определенного типа
func (r *PostgresRepository) GetByType(ctx context.Context, metricType string) ([]domain.Metrics, error) {
	return nil, fmt.Errorf("not implemented")
}

// Ping проверяет соединение с базой данных
func (r *PostgresRepository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

// Close закрывает пул соединений
func (r *PostgresRepository) Close() error {
	r.pool.Close()
	return nil
}
