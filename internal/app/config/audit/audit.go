package audit

type AuditConfig struct {
	AuditURL  string `yaml:"audit_url" env:"AUDIT_URL"`
	AuditFile string `yaml:"audit_file" env:"AUDIT_FILE"`
}

func (ac *AuditConfig) IsEnabled() bool {
	return ac.AuditURL != "" && ac.AuditFile != ""
}
