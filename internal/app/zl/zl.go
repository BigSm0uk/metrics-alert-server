package zl

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/config"
)

// NewLogger создает новый логгер в зависимости от окружения.
// Доступные окружения: production, local, development (по умолчанию).
func NewLogger(env string) *zap.Logger {
	switch env {
	case config.EnvProduction:
		return productionLogger()
	case config.EnvLocal:
		return localLogger()
	default:
		return developmentLogger()
	}
}

func developmentLogger() *zap.Logger {
	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.DebugLevel),
		Development: true,
		Encoding:    "console",
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:     "message",
			LevelKey:       "level",
			TimeKey:        "time",
			NameKey:        "logger",
			StacktraceKey:  "stacktrace",
			EncodeLevel:    zapcore.CapitalColorLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeName:     zapcore.FullNameEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		InitialFields:    map[string]any{},
	}

	logger, _ := cfg.Build(zap.AddStacktrace(zapcore.DPanicLevel))
	return logger
}

func localLogger() *zap.Logger {
	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.DebugLevel),
		Development: true,
		Encoding:    "console",
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:     "message",
			LevelKey:       "level",
			TimeKey:        "time",
			NameKey:        "logger",
			StacktraceKey:  "stacktrace",
			EncodeLevel:    zapcore.CapitalColorLevelEncoder,
			EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeName:     zapcore.FullNameEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		InitialFields:    map[string]any{},
	}

	logger, _ := cfg.Build(zap.AddStacktrace(zapcore.DPanicLevel))
	return logger
}

func productionLogger() *zap.Logger {
	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.InfoLevel),
		Development: false,
		Encoding:    "json",
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:     "message",
			LevelKey:       "level",
			TimeKey:        "time",
			NameKey:        "logger",
			StacktraceKey:  "stacktrace",
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeName:     zapcore.FullNameEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		InitialFields:    map[string]any{},
	}

	logger, _ := cfg.Build(zap.AddStacktrace(zapcore.DPanicLevel))
	return logger
}
