package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

func nuevoRouterDeCriterios(t *testing.T) (*gin.Engine, *mockServicioCriterios) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	servicio := nuevoMockServicioCriterios(t)
	router := gin.New()
	RegistrarRutasAcceptanceCriteria(router, servicio)
	return router, servicio
}

func decodificarCriterio(t *testing.T, w *httptest.ResponseRecorder) respuestaAcceptanceCriterion {
	t.Helper()
	var respuesta respuestaAcceptanceCriterion
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &respuesta), "cuerpo: %s", w.Body.String())
	return respuesta
}

func TestCrearCriterioHandler(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre        string
		cuerpo        string
		preparar      func(m *mockServicioCriterios)
		estadoHTTP    int
		errorEsperado string
		campoEsperado string
	}{
		{
			nombre: "alta valida",
			cuerpo: `{"descripcion":"El usuario ve el saldo"}`,
			preparar: func(m *mockServicioCriterios) {
				m.On("Crear", mock.Anything, uint(1), uint(2), domain.NuevoAcceptanceCriterion{Descripcion: "El usuario ve el saldo"}).
					Return(&domain.AcceptanceCriterion{ID: 5, HistoriaID: 2, Descripcion: "El usuario ve el saldo", Position: 1}, nil)
			},
			estadoHTTP: http.StatusCreated,
		},
		{
			nombre: "descripcion vacia",
			cuerpo: `{"descripcion":"   "}`,
			preparar: func(m *mockServicioCriterios) {
				m.On("Crear", mock.Anything, uint(1), uint(2), mock.Anything).Return(nil, domain.ErrDescripcionCriterioObligatoria)
			},
			estadoHTTP:    http.StatusUnprocessableEntity,
			errorEsperado: codigoValidacion,
			campoEsperado: "descripcion",
		},
		{
			nombre: "historia inexistente",
			cuerpo: `{"descripcion":"Algo"}`,
			preparar: func(m *mockServicioCriterios) {
				m.On("Crear", mock.Anything, uint(1), uint(2), mock.Anything).Return(nil, domain.ErrHistoriaNoEncontrada)
			},
			estadoHTTP:    http.StatusNotFound,
			errorEsperado: codigoHistoriaNoEncontrada,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			router, servicio := nuevoRouterDeCriterios(t)
			caso.preparar(servicio)

			w := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodPost, "/api/proyectos/1/historias/2/criterios", bytes.NewReader([]byte(caso.cuerpo)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, caso.estadoHTTP, w.Code, "cuerpo: %s", w.Body.String())
			if caso.errorEsperado != "" {
				respuesta := decodificarError(t, w)
				assert.Equal(t, caso.errorEsperado, respuesta.Error)
				assert.Equal(t, caso.campoEsperado, respuesta.Campo)
			}
		})
	}
}

func TestListarCriteriosHandler(t *testing.T) {
	t.Parallel()
	router, servicio := nuevoRouterDeCriterios(t)
	servicio.On("Listar", mock.Anything, uint(1), uint(2)).Return([]domain.AcceptanceCriterion{
		{ID: 5, HistoriaID: 2, Descripcion: "Primero", Position: 1},
		{ID: 6, HistoriaID: 2, Descripcion: "Segundo", Position: 2, IsMet: true},
	}, nil)

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/api/proyectos/1/historias/2/criterios", nil)
	require.NoError(t, err)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var respuesta respuestaListadoAcceptanceCriteria
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &respuesta))
	assert.Equal(t, 2, respuesta.Total)
	assert.Len(t, respuesta.Criterios, 2)
	assert.True(t, respuesta.Criterios[1].IsMet)
}

func TestActualizarCriterioHandler(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre     string
		cuerpo     string
		preparar   func(m *mockServicioCriterios)
		estadoHTTP int
	}{
		{
			nombre: "marca cumplido",
			cuerpo: `{"is_met":true}`,
			preparar: func(m *mockServicioCriterios) {
				cumplido := true
				m.On("Actualizar", mock.Anything, uint(1), uint(2), uint(3), domain.ActualizacionAcceptanceCriterion{IsMet: &cumplido}).
					Return(&domain.AcceptanceCriterion{ID: 3, HistoriaID: 2, Descripcion: "Algo", Position: 1, IsMet: true}, nil)
			},
			estadoHTTP: http.StatusOK,
		},
		{
			nombre: "actualizacion vacia",
			cuerpo: `{}`,
			preparar: func(m *mockServicioCriterios) {
				m.On("Actualizar", mock.Anything, uint(1), uint(2), uint(3), mock.Anything).Return(nil, domain.ErrActualizacionCriterioVacia)
			},
			estadoHTTP: http.StatusUnprocessableEntity,
		},
		{
			nombre: "criterio inexistente",
			cuerpo: `{"is_met":true}`,
			preparar: func(m *mockServicioCriterios) {
				m.On("Actualizar", mock.Anything, uint(1), uint(2), uint(3), mock.Anything).Return(nil, domain.ErrCriterioNoEncontrado)
			},
			estadoHTTP: http.StatusNotFound,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			router, servicio := nuevoRouterDeCriterios(t)
			caso.preparar(servicio)

			w := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodPatch, "/api/proyectos/1/historias/2/criterios/3", bytes.NewReader([]byte(caso.cuerpo)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, caso.estadoHTTP, w.Code, "cuerpo: %s", w.Body.String())
			if caso.estadoHTTP == http.StatusOK {
				assert.True(t, decodificarCriterio(t, w).IsMet)
			}
		})
	}
}

func TestEliminarCriterioHandler(t *testing.T) {
	t.Parallel()
	router, servicio := nuevoRouterDeCriterios(t)
	servicio.On("Eliminar", mock.Anything, uint(1), uint(2), uint(3)).Return(nil)

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodDelete, "/api/proyectos/1/historias/2/criterios/3", nil)
	require.NoError(t, err)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}
