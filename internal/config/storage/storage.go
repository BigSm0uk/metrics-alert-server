package storage

type StorageConfig struct {
	Type string `yaml:"type" required:"true"`
	//другие поля для других провайдеров в будущем
}
