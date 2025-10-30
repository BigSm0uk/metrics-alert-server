package domain

type AuditMessage struct {
	Ts      int64    `json:"ts"`         // Время события
	Metrics []string `json:"metrics"`    // Наименования полученный метрик
	IpAddr  string   `json:"ip_address"` // IP адрес входящего запроса
}
type AuditMessages []AuditMessage
