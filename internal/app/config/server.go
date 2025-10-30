package config

import (
	"flag"
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/config/audit"
	S "github.com/bigsm0uk/metrics-alert-server/internal/app/config/storage"
	Store "github.com/bigsm0uk/metrics-alert-server/internal/app/config/store"
)

const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
	EnvLocal       = "local"

	StorageTypeMem      = "mem"
	StorageTypePostgres = "postgres"
)

type ServerConfig struct {
	Env          string            `yaml:"env"  env-default:"development"`
	Storage      S.StorageConfig   `yaml:"storage" required:"true"`
	TemplatePath string            `yaml:"template_path" env-default:"api/templates/metrics.html"`
	Addr         string            `env:"ADDRESS"`
	Store        Store.StoreConfig `required:"true"`
	Key          string            `env:"KEY"`
	Audit        audit.AuditConfig `yaml:"audit"`
}

func LoadServerConfig() (*ServerConfig, error) {
	cfg := InitDefaultConfig()
	path := flag.String("config", "config/config.dev.yaml", "path to config file")

	var (
		flagAddr      = flag.String("a", "", "server address")
		flagFile      = flag.String("f", "", "path to store file")
		flagRestore   = flag.Bool("r", true, "restore store from file")
		flagInterval  = flag.String("i", "", "store interval")
		flagDB        = flag.String("d", "", "database connection string")
		flagKey       = flag.String("k", "", "key")
		flagAuditURL  = flag.String("audit-url", "", "audit URL")
		flagAuditFile = flag.String("audit-file", "", "audit file")
	)

	flag.Parse()

	// Пытаемся прочитать YAML файл (он перезапишет дефолты + применит env переменные)
	if err := cleanenv.ReadConfig(*path, cfg); err != nil {
		fmt.Printf("failed to read config: %v. In development mode. Using default config...", err)
		// Если файла нет, применяем env переменные напрямую
		_ = cleanenv.ReadEnv(cfg)

	}
	// Применяем флаги командной строки ТОЛЬКО если они были явно указаны
	// ENV переменные имеют приоритет над флагами для KEY!
	if *flagAddr != "" {
		cfg.Addr = *flagAddr
	}
	if *flagFile != "" {
		cfg.Store.FileStoragePath = *flagFile
	}
	// flagRestore всегда имеет значение (bool), проверяем через flag.Visit
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "r" {
			cfg.Store.Restore = *flagRestore
		}
	})
	if *flagInterval != "" {
		cfg.Store.StoreInterval = *flagInterval
	}

	if *flagDB != "" {
		cfg.Storage.ConnectionString = *flagDB
	}

	if cfg.Key == "" && *flagKey != "" {
		cfg.Key = *flagKey
	}
	cfg.Store.UseStore = cfg.isActiveStore()

	if *flagAuditURL != "" {
		cfg.Audit.AuditURL = *flagAuditURL
	}
	if *flagAuditFile != "" {
		cfg.Audit.AuditFile = *flagAuditFile
	}

	return cfg, nil
}

func (s *ServerConfig) isActiveStore() bool {
	return !s.IsPgStoreStorage() && s.Store.FileStoragePath != ""
}

func (s *ServerConfig) IsPgStoreStorage() bool {
	return s.Storage.ConnectionString != ""
}

func InitDefaultConfig() *ServerConfig {
	return &ServerConfig{
		Addr: "localhost:8080",
		Storage: S.StorageConfig{
			ConnectionString: "",
		},
		TemplatePath: "api/templates/metrics.html",
		Env:          EnvDevelopment,
		Store: Store.StoreConfig{
			UseStore:      true,
			StoreInterval: "300",
			SFormat:       "json",
		},
	}
}
