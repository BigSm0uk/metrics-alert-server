package config

import (
	"flag"
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"

	S "github.com/bigsm0uk/metrics-alert-server/internal/config/storage"
	Store "github.com/bigsm0uk/metrics-alert-server/internal/config/store"
)

const (
	EnvDevelopment = "development"
	EnvProduction  = "production"

	StorageTypeMem      = "mem"
	StorageTypePostgres = "postgres"
)

type ServerConfig struct {
	Env          string            `yaml:"env"  env-default:"development"`
	Storage      S.StorageConfig   `yaml:"storage" required:"true"`
	TemplatePath string            `yaml:"template_path" env-default:"api/templates/metrics.html"`
	Addr         string            `env:"ADDRESS"`
	Store        Store.StoreConfig `required:"true"`
}

func LoadServerConfig() (*ServerConfig, error) {
	cfg := &ServerConfig{}
	path := flag.String("config", "config/config.dev.yaml", "path to config file")
	flag.StringVar(&cfg.Addr, "a", "localhost:8080", "server address")

	flag.StringVar(&cfg.Store.FileStoragePath, "f", "store.json", "path to store file")
	flag.BoolVar(&cfg.Store.Restore, "r", true, "restore store from file")
	flag.StringVar(&cfg.Store.StoreInterval, "i", "300", "store interval")

	flag.StringVar(&cfg.Storage.ConnectionString, "d", "", "database connection string")

	flag.Parse()

	if err := cleanenv.ReadConfig(*path, cfg); err != nil {
		fmt.Printf("failed to read config: %v. In development mode. Using default config...", err)
		cfg = InitDefaultConfig()
	}

	cfg.Store.UseStore = cfg.isActiveStore()

	return cfg, nil
}
func (s *ServerConfig) isActiveStore() bool {
	return s.Storage.Type != StorageTypePostgres && s.Store.FileStoragePath != ""

}
func InitDefaultConfig() *ServerConfig {
	return &ServerConfig{
		Addr: "localhost:8080",
		Storage: S.StorageConfig{
			Type: "mem",
		},
		TemplatePath: "api/templates/metrics.html",
		Env:          EnvDevelopment,
		Store: Store.StoreConfig{
			UseStore: false,
		},
	}
}
