package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNombreBaseDeDatos(t *testing.T) {
	casos := []struct {
		nombre   string
		entorno  string
		vars     map[string]string
		esperado string
	}{
		{
			nombre:   "la variable especifica del entorno tiene prioridad sobre DB_NAME",
			entorno:  "development",
			vars:     map[string]string{"DB_NAME_DEV": "metrics_db_dev", "DB_NAME": "otra_db"},
			esperado: "metrics_db_dev",
		},
		{
			nombre:   "DB_NAME cubre development cuando no hay DB_NAME_DEV (caso del pipeline de CI)",
			entorno:  "development",
			vars:     map[string]string{"DB_NAME_DEV": "", "DB_NAME": "metrics_db"},
			esperado: "metrics_db",
		},
		{
			nombre:   "DB_NAME cubre test cuando no hay DB_NAME_TEST",
			entorno:  "test",
			vars:     map[string]string{"DB_NAME_TEST": "", "DB_NAME": "metrics_db"},
			esperado: "metrics_db",
		},
		{
			nombre:   "DB_NAME_PROD tiene prioridad sobre DB_NAME en produccion",
			entorno:  "production",
			vars:     map[string]string{"DB_NAME_PROD": "metrics_db_prod", "DB_NAME": "metrics_db"},
			esperado: "metrics_db_prod",
		},
		{
			nombre:   "sin variables development usa metrics_db_dev",
			entorno:  "development",
			vars:     map[string]string{"DB_NAME_DEV": "", "DB_NAME": ""},
			esperado: "metrics_db_dev",
		},
		{
			nombre:   "sin variables test usa metrics_db_test",
			entorno:  "test",
			vars:     map[string]string{"DB_NAME_TEST": "", "DB_NAME": ""},
			esperado: "metrics_db_test",
		},
		{
			nombre:   "sin variables production usa metrics_db",
			entorno:  "production",
			vars:     map[string]string{"DB_NAME_PROD": "", "DB_NAME": ""},
			esperado: "metrics_db",
		},
		{
			nombre:   "un entorno desconocido cae en los defaults de development",
			entorno:  "staging",
			vars:     map[string]string{"DB_NAME_DEV": "", "DB_NAME": ""},
			esperado: "metrics_db_dev",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			for variable, valor := range caso.vars {
				t.Setenv(variable, valor)
			}
			assert.Equal(t, caso.esperado, nombreBaseDeDatos(caso.entorno))
		})
	}
}

func TestCargarBaseDeDatos(t *testing.T) {
	t.Run("usa los defaults de conexion cuando no hay variables", func(t *testing.T) {
		for _, variable := range []string{
			"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_SSLMODE",
			"DB_NAME", "DB_NAME_DEV",
		} {
			t.Setenv(variable, "")
		}

		bd := cargarBaseDeDatos("development")

		assert.Equal(t, "localhost", bd.host)
		assert.Equal(t, "5432", bd.port)
		assert.Equal(t, "postgres", bd.user)
		assert.Empty(t, bd.password)
		assert.Equal(t, "disable", bd.sslmode)
		assert.Equal(t, "metrics_db_dev", bd.name)
	})

	t.Run("lee las variables definidas", func(t *testing.T) {
		t.Setenv("DB_HOST", "postgresql")
		t.Setenv("DB_PORT", "5433")
		t.Setenv("DB_USER", "postgres")
		t.Setenv("DB_PASSWORD", "secreto")
		t.Setenv("DB_SSLMODE", "require")
		t.Setenv("DB_NAME_DEV", "")
		t.Setenv("DB_NAME", "metrics_db")

		bd := cargarBaseDeDatos("development")

		assert.Equal(t, "postgresql", bd.host)
		assert.Equal(t, "5433", bd.port)
		assert.Equal(t, "postgres", bd.user)
		assert.Equal(t, "secreto", bd.password)
		assert.Equal(t, "require", bd.sslmode)
		assert.Equal(t, "metrics_db", bd.name)
	})
}
