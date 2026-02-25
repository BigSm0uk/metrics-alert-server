package store

type StoreConfig struct {
	StoreInterval   string `yaml:"store_interval" json:"store_interval" env:"STORE_INTERVAL" env-default:"300"`
	FileStoragePath string `yaml:"store_file" json:"store_file" env:"FILE_STORAGE_PATH" env-default:"store.json"`
	Restore         bool   `yaml:"restore" json:"restore" env:"RESTORE" env-default:"true"`
	SFormat         string `yaml:"format" json:"format" env:"FORMAT" env-default:"json"`
	UseStore        bool   `yaml:"-" json:"-"`
}
