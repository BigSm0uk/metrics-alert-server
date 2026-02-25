package config

import (
	"os"
	"testing"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerConfigJSONParsing(t *testing.T) {
	// Создаем временный JSON файл
	jsonContent := `{
		"address": "localhost:9090",
		"store": {
			"restore": true,
			"store_interval": "5s",
			"store_file": "/tmp/test-metrics.db"
		},
		"storage": {
			"database_dsn": "postgres://user:pass@localhost/testdb"
		}
	}`

	tmpFile, err := os.CreateTemp("", "server-config-*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(jsonContent)
	require.NoError(t, err)
	tmpFile.Close()

	// Тестируем прямое чтение JSON файла
	cfg := InitDefaultConfig()
	err = cleanenv.ReadConfig(tmpFile.Name(), cfg)
	require.NoError(t, err)

	assert.Equal(t, "localhost:9090", cfg.Addr)
	assert.Equal(t, true, cfg.Store.Restore)
	assert.Equal(t, "5s", cfg.Store.StoreInterval)
	assert.Equal(t, "/tmp/test-metrics.db", cfg.Store.FileStoragePath)
	assert.Equal(t, "postgres://user:pass@localhost/testdb", cfg.Storage.ConnectionString)

}

func TestAgentConfigJSONParsing(t *testing.T) {
	// Создаем временный JSON файл
	jsonContent := `{
		"address": "localhost:9090",
		"report_interval": 5,
		"poll_interval": 3
	}`

	tmpFile, err := os.CreateTemp("", "agent-config-*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(jsonContent)
	require.NoError(t, err)
	tmpFile.Close()

	// Тестируем прямое чтение JSON файла
	cfg := &AgentConfig{
		Env:            EnvDevelopment,
		Addr:           "localhost:8080",
		ReportInterval: 10,
		PollInterval:   2,
		RateLimit:      1,
		Key:            "1234567890",
	}

	err = cleanenv.ReadConfig(tmpFile.Name(), cfg)
	require.NoError(t, err)

	assert.Equal(t, "localhost:9090", cfg.Addr)
	assert.Equal(t, uint(5), cfg.ReportInterval)
	assert.Equal(t, uint(3), cfg.PollInterval)

}
