package audit

import (
	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
	"github.com/bigsm0uk/metrics-alert-server/internal/domain/interfaces"
)

type AuditService struct {
	observers map[string]interfaces.AuditObserver
}

var _ interfaces.AuditSubject = &AuditService{}

func (s *AuditService) NotifyAll(message domain.AuditMessage) {
	for _, o := range s.observers {
		o.Notify(message)
	}
}

func (s *AuditService) Attach(o interfaces.AuditObserver) {
	if s.observers == nil {
		s.observers = make(map[string]interfaces.AuditObserver)
	}
	s.observers[o.GetID()] = o
}

func (s *AuditService) Detach(ID string) {
	delete(s.observers, ID)
}
