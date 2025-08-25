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
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

const maxRetries = 3
const retryDelay = time.Second

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

	repo := &PostgresRepository{pool: pool}

	repo.MustBootstrap(ctx)

	return repo, nil
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
);

CREATE INDEX IF NOT EXISTS idx_metrics_type ON metrics(type);
CREATE INDEX IF NOT EXISTS idx_metrics_updated_at ON metrics(updated_at);

ALTER TABLE metrics ADD CONSTRAINT IF NOT EXISTS chk_counter_has_delta 
    CHECK ((type = 'counter' AND delta IS NOT NULL) OR type != 'counter');

ALTER TABLE metrics ADD CONSTRAINT IF NOT EXISTS chk_gauge_has_value 
    CHECK ((type = 'gauge' AND value IS NOT NULL) OR type != 'gauge');`

	pgErrClassifier := pgerrors.NewPostgresErrorClassifier()
	for i := range maxRetries {
		_, err := r.pool.Exec(ctx, sql)
		if err == nil {
			return
		}
		if pgErrClassifier.Classify(err) == pgerrors.NonRetriable {
			zl.Log.Error("failed to bootstrap database", zap.Error(err))
		}
		time.Sleep(retryDelay * time.Duration(i+1))
	}
	zl.Log.Error("failed to bootstrap database", zap.String("dsn", r.pool.Config().ConnString()))
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
	pgErrClassifier := pgerrors.NewPostgresErrorClassifier()
	for i := range maxRetries {
		_, err = r.pool.Exec(ctx, sqlQuery, args...)
		if err == nil {
			return nil
		}
		if pgErrClassifier.Classify(err) == pgerrors.NonRetriable {
			return fmt.Errorf("exec query: %w", err)
		}
		time.Sleep(retryDelay * time.Duration(i+1))
	}
	return fmt.Errorf("exec query: %w", err)
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
	pgErrClassifier := pgerrors.NewPostgresErrorClassifier()
	for i := range maxRetries {
		err := r.pool.QueryRow(ctx, sqlQuery, args...).Scan(&gotID, &gotType, &value, &delta, &hash)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrMetricNotFound
		}
		if err != nil {
			if pgErrClassifier.Classify(err) == pgerrors.NonRetriable {
				return nil, fmt.Errorf("query row: %w", err)
			}
			time.Sleep(retryDelay * time.Duration(i+1))
			continue
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

	pgErrClassifier := pgerrors.NewPostgresErrorClassifier()
	var rows pgx.Rows
	for i := range maxRetries {
		r, err := r.pool.Query(ctx, sqlQuery, args...)
		if err != nil {
			if pgErrClassifier.Classify(err) == pgerrors.NonRetriable {
				return nil, fmt.Errorf("exec query: %w", err)
			}
			time.Sleep(retryDelay * time.Duration(i+1))
			continue
		}
		rows = r
		break
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
	pgErrClassifier := pgerrors.NewPostgresErrorClassifier()
	for i := range maxRetries {
		_, err = r.pool.Exec(ctx, sqlQuery, args...)
		if err == nil {
			return nil
		}
		if pgErrClassifier.Classify(err) == pgerrors.NonRetriable {
			return fmt.Errorf("exec query: %w", err)
		}
		time.Sleep(retryDelay * time.Duration(i+1))
	}
	return fmt.Errorf("exec query: %w", err)
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

	pgErrClassifier := pgerrors.NewPostgresErrorClassifier()
	var rows pgx.Rows
	for i := range maxRetries {
		r, err := r.pool.Query(ctx, sqlQuery, args...)
		if err != nil {
			if pgErrClassifier.Classify(err) == pgerrors.NonRetriable {
				return nil, fmt.Errorf("exec query: %w", err)
			}
			time.Sleep(retryDelay * time.Duration(i+1))
			continue
		}
		rows = r
		break
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
	pgErrClassifier := pgerrors.NewPostgresErrorClassifier()
	var err error
	for i := range maxRetries {
		err = r.pool.Ping(ctx)
		if err == nil {
			return nil
		}
		if pgErrClassifier.Classify(err) == pgerrors.NonRetriable {
			return fmt.Errorf("ping attempt: %d: %w", i+1, err)
		}
		time.Sleep(retryDelay * time.Duration(i+1))
	}
	return fmt.Errorf("ping attempt: %d: %w", maxRetries+1, err)
}

func (r *PostgresRepository) Close() error {
	r.pool.Close()
	return nil
}
