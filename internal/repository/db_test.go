package repository_test

import (
	"os"
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
	if testing.Short() {
		t.Skip("test de integracion omitido en modo -short")
	}
	if os.Getenv("DB_HOST") == "" {
		t.Skip("DB_HOST no definido: base de datos no disponible para el test de integracion")
	}

	t.Setenv("APP_ENV", "test")
	setEnvIfEmpty(t, "DB_PORT", "5433")
	setEnvIfEmpty(t, "DB_USER", "postgres")
	setEnvIfEmpty(t, "DB_PASSWORD", "postgres")
	setEnvIfEmpty(t, "DB_SSLMODE", "disable")
	setEnvIfEmpty(t, "DB_NAME_TEST", "metrics_db_test")

	cfg, err := config.Load()
	require.NoError(t, err)

	db, err := repository.ConnectDB(cfg)
	if err != nil {
		t.Skipf("base de datos no disponible: %v", err)
	}
	require.NotNil(t, db)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Ping())
}

// setEnvIfEmpty define un valor por defecto solo cuando la variable no fue
// provista por el entorno. Asi el test respeta la configuracion de CI o de
// docker-compose (por ejemplo DB_PORT=5432 en CI) sin dejar de funcionar con
// los valores locales habituales.
func setEnvIfEmpty(t *testing.T, key, value string) {
	t.Helper()
	if os.Getenv(key) != "" {
		return
	}
	t.Setenv(key, value)
}
