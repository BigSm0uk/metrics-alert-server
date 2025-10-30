package interfaces

import "github.com/bigsm0uk/metrics-alert-server/internal/domain"

// AuditObserver - реагирует на события аудита
type AuditObserver interface {
	GetID() string
	Notify(domain.AuditMessage)
}

// AuditSubject - управляет наблюдателями
type AuditSubject interface {
	Attach(AuditObserver)
	Detach(string)
	NotifyAll(domain.AuditMessage)
}
