package domain

type AuditMessage struct {
	TS      int64    `json:"ts"`         // Время события
	Metrics []string `json:"metrics"`    // Наименования полученный метрик
	IPAddr  string   `json:"ip_address"` // IP адрес входящего запроса
}
type AuditMessages []AuditMessage
