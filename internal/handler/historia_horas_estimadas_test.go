package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

func TestHorasEstimadasHandler(t *testing.T) {
	t.Parallel()

	t.Run("BDD-01 registrar horas estimadas en el alta", func(t *testing.T) {
		t.Parallel()

		router, servicio := nuevoRouterDeHistorias(t)
		historia := historiaDeEjemplo()
		historia.HorasEstimadas = floatPtr(5)
		servicio.On("Crear", mock.Anything, uint(1), domain.NuevaHistoria{
			Titulo:         "Historia estimada",
			HorasEstimadas: floatPtr(5),
		}).Return(historia, nil)

		w := ejecutar(t, router, http.MethodPost, "/api/proyectos/1/historias", `{"titulo":"Historia estimada","horas_estimadas":5}`)

		require.Equal(t, http.StatusCreated, w.Code)
		respuesta := decodificarHistoria(t, w)
		require.NotNil(t, respuesta.HorasEstimadas)
		assert.Equal(t, 5.0, *respuesta.HorasEstimadas)
		servicio.AssertExpectations(t)
	})

	t.Run("BDD-02 el alta sin horas estimadas responde null", func(t *testing.T) {
		t.Parallel()

		router, servicio := nuevoRouterDeHistorias(t)
		servicio.On("Crear", mock.Anything, uint(1), domain.NuevaHistoria{Titulo: "Sin horas"}).
			Return(historiaDeEjemplo(), nil)

		w := ejecutar(t, router, http.MethodPost, "/api/proyectos/1/historias", `{"titulo":"Sin horas"}`)

		require.Equal(t, http.StatusCreated, w.Code)
		assert.Nil(t, decodificarHistoria(t, w).HorasEstimadas)
		assert.Contains(t, w.Body.String(), `"horas_estimadas":null`, "el campo se emite siempre, con null")
		servicio.AssertExpectations(t)
	})

	t.Run("BDD-03 el alta con horas estimadas null equivale a no informado", func(t *testing.T) {
		t.Parallel()

		router, servicio := nuevoRouterDeHistorias(t)
		servicio.On("Crear", mock.Anything, uint(1), domain.NuevaHistoria{Titulo: "Con null"}).
			Return(historiaDeEjemplo(), nil)

		w := ejecutar(t, router, http.MethodPost, "/api/proyectos/1/historias", `{"titulo":"Con null","horas_estimadas":null}`)

		require.Equal(t, http.StatusCreated, w.Code)
		assert.Nil(t, decodificarHistoria(t, w).HorasEstimadas)
		servicio.AssertExpectations(t)
	})

	t.Run("BDD-04 modificar horas estimadas", func(t *testing.T) {
		t.Parallel()

		router, servicio := nuevoRouterDeHistorias(t)
		historia := historiaDeEjemplo()
		historia.HorasEstimadas = floatPtr(8)
		servicio.On("Actualizar", mock.Anything, uint(1), uint(5), domain.ActualizacionHistoria{
			HorasEstimadas: floatPtr(8),
		}).Return(historia, nil)

		w := ejecutar(t, router, http.MethodPatch, "/api/proyectos/1/historias/5", `{"horas_estimadas":8}`)

		require.Equal(t, http.StatusOK, w.Code)
		respuesta := decodificarHistoria(t, w)
		require.NotNil(t, respuesta.HorasEstimadas)
		assert.Equal(t, 8.0, *respuesta.HorasEstimadas)
		servicio.AssertExpectations(t)
	})

	t.Run("BDD-05 las horas estimadas aparecen en el listado", func(t *testing.T) {
		t.Parallel()

		router, servicio := nuevoRouterDeHistorias(t)
		historia := historiaDeEjemplo()
		historia.HorasEstimadas = floatPtr(8)
		servicio.On("Listar", mock.Anything, uint(1)).Return([]domain.Historia{*historia}, nil)

		w := ejecutar(t, router, http.MethodGet, "/api/proyectos/1/historias", nil)

		require.Equal(t, http.StatusOK, w.Code)
		var respuesta struct {
			Total     int                 `json:"total"`
			Historias []respuestaHistoria `json:"historias"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &respuesta))
		require.Len(t, respuesta.Historias, 1)
		require.NotNil(t, respuesta.Historias[0].HorasEstimadas)
		assert.Equal(t, 8.0, *respuesta.Historias[0].HorasEstimadas)
		servicio.AssertExpectations(t)
	})

	t.Run("CL-04 null en la modificacion equivale a no informado", func(t *testing.T) {
		t.Parallel()

		router, servicio := nuevoRouterDeHistorias(t)
		servicio.On("Actualizar", mock.Anything, uint(1), uint(5), domain.ActualizacionHistoria{}).
			Return(nil, domain.ErrActualizacionHistoriaVacia)

		w := ejecutar(t, router, http.MethodPatch, "/api/proyectos/1/historias/5", `{"horas_estimadas":null}`)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		assert.Empty(t, decodificarError(t, w).Campo)
		servicio.AssertExpectations(t)
	})

	valoresInvalidos := []struct {
		nombre    string
		valor     float64
		escenario string
	}{
		{"cero", 0, "BDD-09"},
		{"negativo", -2, "BDD-10"},
	}
	for _, c := range valoresInvalidos {
		t.Run("alta con horas estimadas "+c.nombre+" es rechazada", func(t *testing.T) {
			t.Parallel()

			router, servicio := nuevoRouterDeHistorias(t)
			servicio.On("Crear", mock.Anything, uint(1), domain.NuevaHistoria{
				Titulo:         "Historia invalida",
				HorasEstimadas: floatPtr(c.valor),
			}).Return(nil, domain.ErrHorasEstimadasInvalidas)

			cuerpo := fmt.Sprintf(`{"titulo":"Historia invalida","horas_estimadas":%v}`, c.valor)
			w := ejecutar(t, router, http.MethodPost, "/api/proyectos/1/historias", cuerpo)

			assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "escenario BDD %s", c.escenario)
			respuesta := decodificarError(t, w)
			assert.Equal(t, codigoValidacion, respuesta.Error)
			assert.Equal(t, "horas_estimadas", respuesta.Campo)
			servicio.AssertExpectations(t)
		})
		t.Run("modificacion con horas estimadas "+c.nombre+" es rechazada", func(t *testing.T) {
			t.Parallel()

			router, servicio := nuevoRouterDeHistorias(t)
			servicio.On("Actualizar", mock.Anything, uint(1), uint(5), domain.ActualizacionHistoria{
				HorasEstimadas: floatPtr(c.valor),
			}).Return(nil, domain.ErrHorasEstimadasInvalidas)

			cuerpo := fmt.Sprintf(`{"horas_estimadas":%v}`, c.valor)
			w := ejecutar(t, router, http.MethodPatch, "/api/proyectos/1/historias/5", cuerpo)

			assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "escenario BDD %s", c.escenario)
			respuesta := decodificarError(t, w)
			assert.Equal(t, codigoValidacion, respuesta.Error)
			assert.Equal(t, "horas_estimadas", respuesta.Campo)
			servicio.AssertExpectations(t)
		})
	}

	valoresNoNumericos := []struct {
		nombre string
		valor  string
	}{
		{"texto", `"cinco"`},
		{"booleano", `true`},
		{"objeto", `{"a":1}`},
		{"arreglo", `[5]`},
	}
	for _, c := range valoresNoNumericos {
		t.Run("BDD-11 no numerico en alta: "+c.nombre, func(t *testing.T) {
			t.Parallel()

			router, _ := nuevoRouterDeHistorias(t)

			w := ejecutar(t, router, http.MethodPost, "/api/proyectos/1/historias", `{"titulo":"x","horas_estimadas":`+c.valor+`}`)

			assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
			respuesta := decodificarError(t, w)
			assert.Equal(t, codigoValidacion, respuesta.Error)
			assert.Equal(t, "horas_estimadas", respuesta.Campo)
		})
		t.Run("BDD-11 no numerico en modificacion: "+c.nombre, func(t *testing.T) {
			t.Parallel()

			router, _ := nuevoRouterDeHistorias(t)

			w := ejecutar(t, router, http.MethodPatch, "/api/proyectos/1/historias/5", `{"horas_estimadas":`+c.valor+`}`)

			assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
			respuesta := decodificarError(t, w)
			assert.Equal(t, codigoValidacion, respuesta.Error)
			assert.Equal(t, "horas_estimadas", respuesta.Campo)
		})
	}
}
