package pg

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	pgerrors "github.com/bigsm0uk/metrics-alert-server/internal/app/storage/pgerror"
	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
	"github.com/cenkalti/backoff/v4"
)

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
