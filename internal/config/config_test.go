package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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
