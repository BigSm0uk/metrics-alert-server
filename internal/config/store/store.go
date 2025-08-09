package store

type StoreConfig struct {
	StoreInterval   string `env:"STORE_INTERVAL" env-default:"300s"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" env-default:"store.json"`
	Restore         bool   `env:"RESTORE" env-default:"true"`
}
