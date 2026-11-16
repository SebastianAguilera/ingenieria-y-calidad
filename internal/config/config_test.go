package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_Environments(t *testing.T) {
	// Load exige JWT_SECRET (US-10, E-19); los demas valores siguen probando
	// el manejo de entornos.
	t.Setenv("JWT_SECRET", "secreto-de-prueba")
	t.Setenv("JWT_EXPIRY", "")

	t.Run("Development environment defaults", func(t *testing.T) {
		t.Setenv("APP_ENV", "development")
		t.Setenv("DB_NAME_DEV", "metrics_db")

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
		t.Setenv("APP_ENV", "test")
		t.Setenv("DB_NAME_TEST", "metrics_db_test")

		cfg, err := Load()
		assert.NoError(t, err)
		assert.Equal(t, "test", cfg.AppEnv)
		assert.Equal(t, "metrics_db_test", cfg.DBName)
	})

	t.Run("Production environment defaults", func(t *testing.T) {
		t.Setenv("APP_ENV", "production")
		t.Setenv("DB_NAME_PROD", "metrics_db_prod")
		t.Setenv("CORS_ALLOWED_ORIGINS", "https://ingenieria-y-calidad.vercel.app")

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
	t.Setenv("APP_ENV", "test")

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

// TestCargaCORS cubre la allowlist de origenes: en produccion es obligatoria
// (fail-closed) y solo admite origenes exactos; en desarrollo/test puede
// omitirse porque CORS permite cualquier origen.
func TestCargaCORS(t *testing.T) {
	t.Run("produccion sin CORS_ALLOWED_ORIGINS impide arrancar", func(t *testing.T) {
		t.Setenv("APP_ENV", "production")
		t.Setenv("JWT_SECRET", "secreto-de-prueba")
		t.Setenv("CORS_ALLOWED_ORIGINS", "")

		cfg, err := Load()

		require.Error(t, err, "sin CORS_ALLOWED_ORIGINS la API no debe iniciar en produccion")
		assert.Nil(t, cfg)
	})

	t.Run("produccion con origen exacto lo expone normalizado", func(t *testing.T) {
		t.Setenv("APP_ENV", "production")
		t.Setenv("JWT_SECRET", "secreto-de-prueba")
		t.Setenv("CORS_ALLOWED_ORIGINS", " https://ingenieria-y-calidad.vercel.app/ ")

		cfg, err := Load()

		require.NoError(t, err)
		assert.Equal(t, []string{"https://ingenieria-y-calidad.vercel.app"}, cfg.CORSAllowedOrigins)
	})

	t.Run("produccion con varios origenes separados por coma", func(t *testing.T) {
		t.Setenv("APP_ENV", "production")
		t.Setenv("JWT_SECRET", "secreto-de-prueba")
		t.Setenv("CORS_ALLOWED_ORIGINS", "https://a.example.com, https://b.example.com")

		cfg, err := Load()

		require.NoError(t, err)
		assert.Equal(t, []string{"https://a.example.com", "https://b.example.com"}, cfg.CORSAllowedOrigins)
	})

	t.Run("produccion rechaza comodin", func(t *testing.T) {
		t.Setenv("APP_ENV", "production")
		t.Setenv("JWT_SECRET", "secreto-de-prueba")
		t.Setenv("CORS_ALLOWED_ORIGINS", "*")

		cfg, err := Load()

		require.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("produccion rechaza origen sin esquema", func(t *testing.T) {
		t.Setenv("APP_ENV", "production")
		t.Setenv("JWT_SECRET", "secreto-de-prueba")
		t.Setenv("CORS_ALLOWED_ORIGINS", "ingenieria-y-calidad.vercel.app")

		cfg, err := Load()

		require.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("produccion rechaza origen con path", func(t *testing.T) {
		t.Setenv("APP_ENV", "production")
		t.Setenv("JWT_SECRET", "secreto-de-prueba")
		t.Setenv("CORS_ALLOWED_ORIGINS", "https://ingenieria-y-calidad.vercel.app/api")

		cfg, err := Load()

		require.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("development sin CORS_ALLOWED_ORIGINS sigue arrancando", func(t *testing.T) {
		t.Setenv("APP_ENV", "development")
		t.Setenv("JWT_SECRET", "secreto-de-prueba")
		t.Setenv("CORS_ALLOWED_ORIGINS", "")

		cfg, err := Load()

		require.NoError(t, err)
		assert.Empty(t, cfg.CORSAllowedOrigins)
	})
}

// TestCargaRateLimit cubre los limites de autenticacion: 20 solicitudes por
// minuto por defecto y validacion de valores configurados.
func TestCargaRateLimit(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("JWT_SECRET", "secreto-de-prueba")

	t.Run("valores por defecto", func(t *testing.T) {
		t.Setenv("AUTH_RATE_LIMIT_MAX", "")
		t.Setenv("AUTH_RATE_LIMIT_WINDOW", "")

		cfg, err := Load()

		require.NoError(t, err)
		assert.Equal(t, 20, cfg.AuthRateLimitMax)
		assert.Equal(t, time.Minute, cfg.AuthRateLimitWindow)
	})

	t.Run("valores configurados", func(t *testing.T) {
		t.Setenv("AUTH_RATE_LIMIT_MAX", "7")
		t.Setenv("AUTH_RATE_LIMIT_WINDOW", "30s")

		cfg, err := Load()

		require.NoError(t, err)
		assert.Equal(t, 7, cfg.AuthRateLimitMax)
		assert.Equal(t, 30*time.Second, cfg.AuthRateLimitWindow)
	})

	t.Run("maximo invalido impide arrancar", func(t *testing.T) {
		t.Setenv("AUTH_RATE_LIMIT_MAX", "0")
		t.Setenv("AUTH_RATE_LIMIT_WINDOW", "1m")

		cfg, err := Load()

		require.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("ventana invalida impide arrancar", func(t *testing.T) {
		t.Setenv("AUTH_RATE_LIMIT_MAX", "20")
		t.Setenv("AUTH_RATE_LIMIT_WINDOW", "una hora")

		cfg, err := Load()

		require.Error(t, err)
		assert.Nil(t, cfg)
	})
}
