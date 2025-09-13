package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	pgerrors "github.com/bigsm0uk/metrics-alert-server/internal/app/storage/pgerror"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/zl"
	"github.com/bigsm0uk/metrics-alert-server/internal/config/storage"
	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
	"github.com/bigsm0uk/metrics-alert-server/internal/interfaces"
	"github.com/cenkalti/backoff/v4"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

var _ interfaces.MetricsRepository = (*PostgresRepository)(nil)

// newBackoff создает конфигурацию backoff для retry операций
func newBackoff() *backoff.ExponentialBackOff {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 1 * time.Second
	b.Multiplier = 2.0
	b.MaxInterval = 5 * time.Second
	b.MaxElapsedTime = 10 * time.Second
	return b
}

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
func (r *PostgresRepository) MustBootstrap(ctx context.Context) {
	sql := `CREATE TABLE IF NOT EXISTS metrics (
    id VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('counter', 'gauge')),
    delta BIGINT,
    value DOUBLE PRECISION,
    hash VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    PRIMARY KEY (id, type)
);`

	operation := func() error {
		_, err := r.pool.Exec(ctx, sql)
		if err != nil {
			pgErrClassifier := pgerrors.NewPostgresErrorClassifier()
			if pgErrClassifier.Classify(err) == pgerrors.NonRetriable {
				zl.Log.Error("failed to bootstrap database", zap.Error(err))
				return backoff.Permanent(err)
			}
			return err
		}
		return nil
	}

	if err := backoff.Retry(operation, newBackoff()); err != nil {
		zl.Log.Error("failed to bootstrap database after retries",
			zap.Error(err),
			zap.String("dsn", r.pool.Config().ConnString()))
	}
}

func (r *PostgresRepository) SaveOrUpdate(ctx context.Context, metric *domain.Metrics) error {
	b := sq.
		Insert("metrics").
		Columns("id", "type", "value", "delta", "hash").
		Values(metric.ID, metric.MType, metric.Value, metric.Delta, metric.Hash).
		Suffix(`
			ON CONFLICT (id, type)
			DO UPDATE SET
				delta = CASE
					WHEN metrics.type = 'counter' THEN COALESCE(EXCLUDED.delta, 0)
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

	operation := func() error {
		_, err := r.pool.Exec(ctx, sqlQuery, args...)
		if err != nil {
			pgErrClassifier := pgerrors.NewPostgresErrorClassifier()
			if pgErrClassifier.Classify(err) == pgerrors.NonRetriable {
				return backoff.Permanent(err)
			}
			return err
		}
		return nil
	}

	return backoff.Retry(operation, newBackoff())
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

	operation := func() error {
		err := r.pool.QueryRow(ctx, sqlQuery, args...).Scan(&gotID, &gotType, &value, &delta, &hash)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				zl.Log.Debug("metric not found", zap.String("id", id), zap.String("type", metricType))
				return backoff.Permanent(domain.ErrMetricNotFound)
			}
			pgErrClassifier := pgerrors.NewPostgresErrorClassifier()
			if pgErrClassifier.Classify(err) == pgerrors.NonRetriable {
				return backoff.Permanent(err)
			}
			return err
		}
		return nil
	}

	if err := backoff.Retry(operation, newBackoff()); err != nil {
		if errors.Is(err, domain.ErrMetricNotFound) {
			return nil, domain.ErrMetricNotFound
		}
		return nil, fmt.Errorf("query row: %w", err)
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

	var rows pgx.Rows

	operation := func() error {
		r, err := r.pool.Query(ctx, sqlQuery, args...)
		if err != nil {
			pgErrClassifier := pgerrors.NewPostgresErrorClassifier()
			if pgErrClassifier.Classify(err) == pgerrors.NonRetriable {
				return backoff.Permanent(err)
			}
			return err
		}
		rows = r
		return nil
	}

	if err := backoff.Retry(operation, newBackoff()); err != nil {
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

func (r *PostgresRepository) SaveOrUpdateBatch(ctx context.Context, metrics []domain.Metrics) error {
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

	operation := func() error {
		_, err := r.pool.Exec(ctx, sqlQuery, args...)
		if err != nil {
			pgErrClassifier := pgerrors.NewPostgresErrorClassifier()
			if pgErrClassifier.Classify(err) == pgerrors.NonRetriable {
				return backoff.Permanent(err)
			}
			return err
		}
		return nil
	}

	return backoff.Retry(operation, newBackoff())
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

	var rows pgx.Rows

	operation := func() error {
		r, err := r.pool.Query(ctx, sqlQuery, args...)
		if err != nil {
			pgErrClassifier := pgerrors.NewPostgresErrorClassifier()
			if pgErrClassifier.Classify(err) == pgerrors.NonRetriable {
				return backoff.Permanent(err)
			}
			return err
		}
		rows = r
		return nil
	}

	if err := backoff.Retry(operation, newBackoff()); err != nil {
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
	operation := func() error {
		err := r.pool.Ping(ctx)
		if err != nil {
			pgErrClassifier := pgerrors.NewPostgresErrorClassifier()
			if pgErrClassifier.Classify(err) == pgerrors.NonRetriable {
				return backoff.Permanent(err)
			}
			return err
		}
		return nil
	}

	return backoff.Retry(operation, newBackoff())
}

func (r *PostgresRepository) Close() error {
	r.pool.Close()
	return nil
}
