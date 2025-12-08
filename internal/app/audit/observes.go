package audit

import (
	"go.uber.org/zap"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/config/audit"
	"github.com/bigsm0uk/metrics-alert-server/internal/domain/interfaces"
)

// createAuditObservers создает observers на основе конфигурации.
// Возвращает список observers, которые должны быть зарегистрированы.
func CreateAuditObservers(cfg *audit.AuditConfig, log *zap.Logger) []interfaces.AuditObserver {
	var observers []interfaces.AuditObserver

	if cfg.AuditFile != "" && cfg.AuditURL != "" {
		observers = append(observers, NewFileObserver(cfg.AuditFile, log))
		observers = append(observers, NewURLObserver(cfg.AuditURL, log))
	}

	return observers
}
