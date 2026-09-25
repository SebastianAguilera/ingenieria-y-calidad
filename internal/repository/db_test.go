package repository_test

import (
	"testing"

	"ingenieria-y-calidad/internal/config"
	"ingenieria-y-calidad/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectDB_Error(t *testing.T) {
	cfg := &config.Config{
		AppEnv:     "test",
		DBHost:     "invalid_host_xyz",
		DBPort:     "5432",
		DBUser:     "postgres",
		DBPassword: "wrongpassword",
		DBName:     "nonexistentdb",
		DBSSLMode:  "disable",
	}

	db, err := repository.ConnectDB(cfg)
	if err != nil {
		assert.Error(t, err)
		assert.Nil(t, db)
	} else {
		assert.NotNil(t, db)
	}
}

func TestConnectDB_Success(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_USER", "postgres")
	t.Setenv("DB_PASSWORD", "postgres")
	t.Setenv("DB_SSLMODE", "disable")
	t.Setenv("DB_NAME_TEST", "metrics_db_test")

	cfg, err := config.Load()
	require.NoError(t, err)

	db, err := repository.ConnectDB(cfg)
	require.NoError(t, err)
	require.NotNil(t, db)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Ping())
}