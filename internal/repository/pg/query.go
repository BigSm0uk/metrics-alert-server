package pg

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/cenkalti/backoff/v4"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	pgerrors "github.com/bigsm0uk/metrics-alert-server/internal/app/storage/pgerror"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/zl"
	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
)

func (r *PostgresRepository) Metric(ctx context.Context, id, metricType string) (*domain.Metrics, error) {
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

func (r *PostgresRepository) MetricList(ctx context.Context) ([]domain.Metrics, error) {
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

	var count int
	err = r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM metrics").Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("scan count: %w", err)
	}

	var metrics []domain.Metrics
	metrics = make([]domain.Metrics, 0, count)
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

func (r *PostgresRepository) MetricListByType(ctx context.Context, metricType string) ([]domain.Metrics, error) {
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
