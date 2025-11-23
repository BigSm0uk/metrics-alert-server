package service

import (
	"strings"

	"go.uber.org/zap"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/config/audit"
	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
	"github.com/bigsm0uk/metrics-alert-server/internal/domain/interfaces"
)

type AuditService struct {
	logger    *zap.Logger
	observers map[string]interfaces.AuditObserver
	cfg       *audit.AuditConfig
}

var _ interfaces.AuditSubject = &AuditService{}

func NewAuditService(cfg *audit.AuditConfig, log *zap.Logger) *AuditService {
	logger := log.Named("audit-service")
	return &AuditService{cfg: cfg, observers: map[string]interfaces.AuditObserver{}, logger: logger}
}

func (s *AuditService) NotifyAll(message domain.AuditMessage) {
	if !s.cfg.IsEnabled() {
		s.logger.Warn("audit service is not enabled, skipping notification")
		return
	}
	s.logger.Debug("notifying all observers",
		zap.Int("observers_count", len(s.observers)),
		zap.String("metrics", strings.Join(message.Metrics, ",")),
	)
	for _, o := range s.observers {
		s.logger.Debug("notifying observer", zap.String("observer_id", o.GetID()))
		o.Notify(message)
	}
}

func (s *AuditService) Attach(observers ...interfaces.AuditObserver) {
	for _, o := range observers {
		s.observers[o.GetID()] = o
		s.logger.Debug("attaching observer", zap.String("observer_id", o.GetID()))
	}
}

func (s *AuditService) Detach(ID string) {
	delete(s.observers, ID)
	s.logger.Debug("detaching observer", zap.String("observer_id", ID))
}
