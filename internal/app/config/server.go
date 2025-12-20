package config

import (
	"flag"
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/config/audit"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/config/cache"
	S "github.com/bigsm0uk/metrics-alert-server/internal/app/config/storage"
	Store "github.com/bigsm0uk/metrics-alert-server/internal/app/config/store"
)

const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
	EnvLocal       = "local"
)

type ServerConfig struct {
	Env           string            `yaml:"env" json:"env" env-default:"development"`
	Storage       S.StorageConfig   `yaml:"storage" json:"storage" required:"true"`
	TemplatePath  string            `yaml:"template_path" json:"template_path" env-default:"api/templates/metrics.html"`
	Addr          string            `yaml:"address" json:"address" env:"ADDRESS"`
	GRPCAddr      string            `yaml:"grpc_address" json:"grpc_address" env:"GRPC_ADDRESS"`
	Store         Store.StoreConfig `yaml:"store" json:"store" required:"true"`
	Key           string            `yaml:"key" json:"key" env:"KEY"`
	CryptoKey     string            `yaml:"crypto_key" json:"crypto_key" env:"CRYPTO_KEY"`
	TrustedSubnet string            `yaml:"trusted_subnet" json:"trusted_subnet" env:"TRUSTED_SUBNET"`
	Audit         audit.AuditConfig `yaml:"audit" json:"audit"`
	Cache         cache.CacheConfig `yaml:"cache" json:"cache"`
	ConfigFile    string            `env:"CONFIG"`
}

func LoadServerConfig() (*ServerConfig, error) {
	cfg := InitDefaultConfig()

	var (
		flagConfig        = flag.String("c", "", "path to config file")
		flagConfigLong    = flag.String("config", "", "path to config file")
		flagAddr          = flag.String("a", "", "server address")
		flagGRPCAddr      = flag.String("g", "", "gRPC server address")
		flagFile          = flag.String("f", "", "path to store file")
		flagRestore       = flag.Bool("r", true, "restore store from file")
		flagInterval      = flag.String("i", "", "store interval")
		flagDB            = flag.String("d", "", "database connection string")
		flagKey           = flag.String("k", "", "key")
		flagCryptoKey     = flag.String("crypto-key", "", "path to private key file for decryption")
		flagTrustedSubnet = flag.String("t", "", "trusted subnet (CIDR)")
		flagAuditURL      = flag.String("audit-url", "", "audit URL")
		flagAuditFile     = flag.String("audit-file", "", "audit file")
	)

	flag.Parse()

	// Сначала применяем переменные окружения
	_ = cleanenv.ReadEnv(cfg)

	// Определяем путь к конфигурационному файлу
	configPath := ""
	if *flagConfig != "" {
		configPath = *flagConfig
	} else if *flagConfigLong != "" {
		configPath = *flagConfigLong
	} else if cfg.ConfigFile != "" {
		configPath = cfg.ConfigFile
	}

	// Если указан конфигурационный файл, читаем его
	if configPath != "" {
		if err := cleanenv.ReadConfig(configPath, cfg); err != nil {
			fmt.Printf("failed to read config file %s: %v. Using environment variables and defaults...\n", configPath, err)
		}
	}

	// Флаги имеют наивысший приоритет
	if *flagAddr != "" {
		cfg.Addr = *flagAddr
	}
	if *flagGRPCAddr != "" {
		cfg.GRPCAddr = *flagGRPCAddr
	}
	if *flagFile != "" {
		cfg.Store.FileStoragePath = *flagFile
	}
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

	if *flagKey != "" {
		cfg.Key = *flagKey
	}
	if *flagCryptoKey != "" {
		cfg.CryptoKey = *flagCryptoKey
	}
	if *flagTrustedSubnet != "" {
		cfg.TrustedSubnet = *flagTrustedSubnet
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
		Addr:     "localhost:8080",
		GRPCAddr: "localhost:9090",
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
