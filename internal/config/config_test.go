package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_Environments(t *testing.T) {
	t.Run("Development environment defaults", func(t *testing.T) {
		os.Setenv("APP_ENV", "development")
		os.Setenv("DB_NAME_DEV", "metrics_db_dev")

		cfg, err := Load()
		assert.NoError(t, err)
		assert.Equal(t, "development", cfg.AppEnv)
		assert.Equal(t, "metrics_db_dev", cfg.DBName)
	})

	t.Run("Test environment defaults", func(t *testing.T) {
		os.Setenv("APP_ENV", "test")
		os.Setenv("DB_NAME_TEST", "metrics_db_test")

		cfg, err := Load()
		assert.NoError(t, err)
		assert.Equal(t, "test", cfg.AppEnv)
		assert.Equal(t, "metrics_db_test", cfg.DBName)
	})

	t.Run("Production environment defaults", func(t *testing.T) {
		os.Setenv("APP_ENV", "production")
		os.Setenv("DB_NAME_PROD", "metrics_db_prod")

		cfg, err := Load()
		assert.NoError(t, err)
		assert.Equal(t, "production", cfg.AppEnv)
		assert.Equal(t, "metrics_db_prod", cfg.DBName)
	})
}

func TestLoadConfig_Defaults(t *testing.T) {
	for _, clave := range []string{
		"APP_ENV", "DB_NAME_DEV", "DB_NAME_TEST", "DB_NAME_PROD",
		"PORT", "DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_SSLMODE",
	} {
		t.Setenv(clave, "")
	}

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "development", cfg.AppEnv)
	assert.Equal(t, "8081", cfg.Port)
	assert.Equal(t, "localhost", cfg.DBHost)
	assert.Equal(t, "5432", cfg.DBPort)
	assert.Equal(t, "postgres", cfg.DBUser)
	assert.Empty(t, cfg.DBPassword)
	assert.Equal(t, "metrics_db_dev", cfg.DBName)
	assert.Equal(t, "disable", cfg.DBSSLMode)
}

func TestLoadConfig_NombresDeBasePorDefecto(t *testing.T) {
	t.Run("test sin DB_NAME_TEST", func(t *testing.T) {
		t.Setenv("APP_ENV", "test")
		t.Setenv("DB_NAME_TEST", "")

		cfg, err := Load()
		require.NoError(t, err)
		assert.Equal(t, "metrics_db_test", cfg.DBName)
	})

	t.Run("production sin DB_NAME_PROD", func(t *testing.T) {
		t.Setenv("APP_ENV", "production")
		t.Setenv("DB_NAME_PROD", "")

		cfg, err := Load()
		require.NoError(t, err)
		assert.Equal(t, "metrics_db", cfg.DBName)
	})

	t.Run("development sin DB_NAME_DEV", func(t *testing.T) {
		t.Setenv("APP_ENV", "development")
		t.Setenv("DB_NAME_DEV", "")

		cfg, err := Load()
		require.NoError(t, err)
		assert.Equal(t, "metrics_db_dev", cfg.DBName)
	})
}

func TestGetEnvOrDefault(t *testing.T) {
	t.Run("devuelve el valor definido", func(t *testing.T) {
		t.Setenv("CLAVE_DE_PRUEBA", "valor")

		assert.Equal(t, "valor", getEnvOrDefault("CLAVE_DE_PRUEBA", "por-defecto"))
	})

	t.Run("devuelve el valor por defecto si esta vacio", func(t *testing.T) {
		t.Setenv("CLAVE_DE_PRUEBA", "")

		assert.Equal(t, "por-defecto", getEnvOrDefault("CLAVE_DE_PRUEBA", "por-defecto"))
	})
}
