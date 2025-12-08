package storage

type StorageConfig struct {
	ConnectionString string `yaml:"connection_string" json:"database_dsn" env:"DATABASE_DSN"`
	MaxConns         int32  `yaml:"max_conns" json:"max_conns" env-default:"25"`
	MinConns         int32  `yaml:"min_conns" json:"min_conns" env-default:"5"`
	MaxConnLifetime  int    `yaml:"max_conn_lifetime" json:"max_conn_lifetime" env-default:"3600"`
	MaxConnIdleTime  int    `yaml:"max_conn_idle_time" json:"max_conn_idle_time" env-default:"1800"`
}
