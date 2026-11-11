package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_Environments(t *testing.T) {
	t.Run("Development environment defaults", func(t *testing.T) {
		os.Setenv("APP_ENV", "development")
		os.Setenv("DB_NAME_DEV", "metrics_db")

		cfg, err := Load()
		assert.NoError(t, err)
		assert.Equal(t, "development", cfg.AppEnv)
		assert.Equal(t, "metrics_db", cfg.DBName)
	})

	t.Run("Development fallback DB name", func(t *testing.T) {
		t.Setenv("APP_ENV", "development")
		t.Setenv("DB_NAME_DEV", "")

		cfg, err := Load()
		assert.NoError(t, err)
		assert.Equal(t, "metrics_db", cfg.DBName)
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

// TestCargaJWT cubre E-19 (SDD secciones 2 y 3, CL-07): sin JWT_SECRET la
// aplicacion no arranca, JWT_EXPIRY tiene default 24h y un valor con formato
// invalido tambien impide el arranque. Los casos usan t.Setenv para aislar el
// entorno, por lo que no corren en paralelo.
func TestCargaJWT(t *testing.T) {
	t.Run("E-19 JWT_SECRET ausente impide arrancar", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "")

		cfg, err := Load()

		require.Error(t, err, "sin JWT_SECRET la aplicacion no debe iniciar")
		assert.Nil(t, cfg)
	})

	t.Run("E-19 JWT_EXPIRY con formato invalido impide arrancar", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "secreto-de-prueba")
		t.Setenv("JWT_EXPIRY", "veinticuatro horas")

		cfg, err := Load()

		require.Error(t, err, "JWT_EXPIRY debe admitir solo formato time.ParseDuration")
		assert.Nil(t, cfg)
	})

	t.Run("JWT_EXPIRY sin configurar usa 24h", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "secreto-de-prueba")
		t.Setenv("JWT_EXPIRY", "")

		cfg, err := Load()

		require.NoError(t, err)
		assert.Equal(t, "secreto-de-prueba", cfg.JWTSecret)
		assert.Equal(t, 24*time.Hour, cfg.JWTExpiry)
	})

	t.Run("JWT_EXPIRY configurado se parsea", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "secreto-de-prueba")
		t.Setenv("JWT_EXPIRY", "2h")

		cfg, err := Load()

		require.NoError(t, err)
		assert.Equal(t, 2*time.Hour, cfg.JWTExpiry)
	})
}
