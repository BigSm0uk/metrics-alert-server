package domain

import "errors"

// Базовые ошибки домена
var (
	ErrMetricNotFound     = errors.New("metric not found")
	ErrInvalidMetricType  = errors.New("invalid metric type")
	ErrInvalidMetricValue = errors.New("invalid metric value")
	ErrMissingMetricValue = errors.New("missing value")
)
