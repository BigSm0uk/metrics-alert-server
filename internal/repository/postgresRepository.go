package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/bigsm0uk/metrics-alert-server/internal/config/storage"
	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
	"github.com/bigsm0uk/metrics-alert-server/internal/interfaces"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

func (r *PostgresRepository) Save(ctx context.Context, metric *domain.Metrics) error {
	b := sq.
		Insert("metrics").
		Columns("id", "type", "value", "delta", "hash").
		Values(metric.ID, metric.MType, metric.Value, metric.Delta, metric.Hash).
		Suffix(`
			ON CONFLICT (id, type)
			DO UPDATE SET
				delta = CASE
					WHEN metrics.type = 'counter' THEN COALESCE(metrics.delta, 0) + COALESCE(EXCLUDED.delta, 0)
					ELSE EXCLUDED.delta
				END,
				value = CASE
					WHEN metrics.type = 'gauge' THEN EXCLUDED.value
					ELSE metrics.value
				END,
				hash = EXCLUDED.hash,
				updated_at = NOW()
		`).
		PlaceholderFormat(sq.Dollar)

	sqlQuery, args, err := b.ToSql()
	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}
	if _, err := r.pool.Exec(ctx, sqlQuery, args...); err != nil {
		return fmt.Errorf("exec query: %w", err)
	}
	return nil
}

func (r *PostgresRepository) Get(ctx context.Context, id, metricType string) (*domain.Metrics, error) {
	sqlQuery, args, err := sq.
		Select("id", "type", "value", "delta", "hash").
		From("metrics").
		Where(sq.Eq{"id": id, "type": metricType}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	var (
		gotID   string
		gotType string
		value   *float64
		delta   *int64
		hash    *string
	)

	err = r.pool.QueryRow(ctx, sqlQuery, args...).Scan(&gotID, &gotType, &value, &delta, &hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrMetricNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("exec query: %w", err)
	}

	m := &domain.Metrics{
		ID:    gotID,
		MType: gotType,
		Value: value,
		Delta: delta,
	}
	if hash != nil {
		m.Hash = *hash
	}
	return m, nil
}

func (r *PostgresRepository) GetAll(ctx context.Context) ([]domain.Metrics, error) {
	sqlQuery, args, err := sq.
		Select("id", "type", "value", "delta", "hash").
		From("metrics").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	rows, err := r.pool.Query(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("exec query: %w", err)
	}
	defer rows.Close()

	var metrics []domain.Metrics
	for rows.Next() {
		var m domain.Metrics
		err = rows.Scan(&m.ID, &m.MType, &m.Value, &m.Delta, &m.Hash)
		if err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		metrics = append(metrics, m)
	}
	return metrics, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	sqlQuery, args, err := sq.
		Delete("metrics").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	_, err = r.pool.Exec(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("exec query: %w", err)
	}
	return nil
}

func (r *PostgresRepository) SaveBatch(ctx context.Context, metrics []domain.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	b := sq.
		Insert("metrics").
		Columns("id", "type", "value", "delta", "hash").
		PlaceholderFormat(sq.Dollar)

	for _, m := range metrics {
		b = b.Values(m.ID, m.MType, m.Value, m.Delta, m.Hash)
	}

	b = b.Suffix(`
		ON CONFLICT (id, type)
		DO UPDATE SET
			delta = CASE
				WHEN metrics.type = 'counter' THEN COALESCE(metrics.delta, 0) + COALESCE(EXCLUDED.delta, 0)
				ELSE EXCLUDED.delta
			END,
			value = CASE
				WHEN metrics.type = 'gauge' THEN EXCLUDED.value
				ELSE metrics.value
			END,
			hash = EXCLUDED.hash,
			updated_at = NOW()
	`)

	sqlQuery, args, err := b.ToSql()
	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}
	if _, err := r.pool.Exec(ctx, sqlQuery, args...); err != nil {
		return fmt.Errorf("exec query: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetByType(ctx context.Context, metricType string) ([]domain.Metrics, error) {
	sqlQuery, args, err := sq.
		Select("id", "type", "value", "delta", "hash").
		From("metrics").
		Where(sq.Eq{"type": metricType}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	rows, err := r.pool.Query(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("exec query: %w", err)
	}
	defer rows.Close()

	var metrics []domain.Metrics
	for rows.Next() {
		var m domain.Metrics
		err = rows.Scan(&m.ID, &m.MType, &m.Value, &m.Delta, &m.Hash)
		if err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		metrics = append(metrics, m)
	}
	return metrics, nil
}

func (r *PostgresRepository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

func (r *PostgresRepository) Close() error {
	r.pool.Close()
	return nil
}
