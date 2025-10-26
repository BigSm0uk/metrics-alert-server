package pg

import (
	"context"
	"fmt"
	"time"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/config/storage"
	pgerrors "github.com/bigsm0uk/metrics-alert-server/internal/app/storage/pgerror"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/zl"
	"github.com/bigsm0uk/metrics-alert-server/internal/domain/interfaces"

	"github.com/cenkalti/backoff/v4"
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
func (r *PostgresRepository) Bootstrap(ctx context.Context) error {
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
		return err
	}
	return nil
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
