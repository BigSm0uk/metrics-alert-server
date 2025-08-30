package domain

import "errors"

// Базовые ошибки домена
var (
	ErrMetricNotFound     = errors.New("metric not found")
	ErrInvalidMetricType  = errors.New("invalid metric type")
	ErrInvalidMetricValue = errors.New("invalid metric value")
	ErrMissingMetricValue = errors.New("missing value")
	ErrMetricExists       = errors.New("metric already exists")
)

// Ошибки репозитория
var (
	ErrRepositoryConnection = errors.New("repository connection error")
	ErrRepositoryTimeout    = errors.New("repository timeout")
	ErrRepositoryNotFound   = errors.New("repository not found")
)

// Ошибки сервиса
var (
	ErrServiceUnavailable = errors.New("service unavailable")
	ErrInternalError      = errors.New("internal server error")
)
