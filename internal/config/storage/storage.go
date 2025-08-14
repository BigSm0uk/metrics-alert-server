package storage

type StorageConfig struct {
	Type             string `yaml:"type" required:"true"`
	ConnectionString string `yaml:"connection_string" env:"DATABASE_DSN"`
	MaxConns         int32  `yaml:"max_conns" env-default:"25"`
	MinConns         int32  `yaml:"min_conns" env-default:"5"`
	MaxConnLifetime  int    `yaml:"max_conn_lifetime" env-default:"3600"`
	MaxConnIdleTime  int    `yaml:"max_conn_idle_time" env-default:"1800"`
}
