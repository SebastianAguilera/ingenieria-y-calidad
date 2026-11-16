package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

func nuevoRouterDeMetricas(t *testing.T) (*gin.Engine, *mockServicioMetricas) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	servicio := nuevoMockServicioMetricas(t)
	router := gin.New()
	RegistrarRutasMetricas(router, servicio)
	return router, servicio
}

func ejecutarMetricas(router *gin.Engine, metodo, ruta, cuerpo string) *httptest.ResponseRecorder {
	var request *http.Request
	if cuerpo == "" {
		request = httptest.NewRequest(metodo, ruta, nil)
	} else {
		request = httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func clavesJSON(t *testing.T, cuerpo []byte) []string {
	t.Helper()
	var objeto map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(cuerpo, &objeto), "cuerpo: %s", string(cuerpo))
	claves := make([]string, 0, len(objeto))
	for clave := range objeto {
		claves = append(claves, clave)
	}
	sort.Strings(claves)
	return claves
}

func TestMetricasHandlerSprint(t *testing.T) {
	t.Parallel()

	porcentaje := 0.6153846153846154
	desviacion := 0.2

	casos := []struct {
		nombre   string
		preparar func(m *mockServicioMetricas)
		estado   int
		cuerpo   string
	}{
		{
			nombre: "E-01 sprint con historias mixtas",
			preparar: func(m *mockServicioMetricas) {
				m.On("CalcularPorSprint", mock.Anything, uint(1), uint(10)).Return(domain.MetricasSprint{
					StoryPointsPlanificados: 13,
					StoryPointsCompletados:  8,
					PorcentajeCompletado:    &porcentaje,
					HistoriasCompletadas:    2,
					HorasEstimadas:          20,
					HorasReales:             24,
					DesviacionEsfuerzo:      &desviacion,
					DefectosDetectados:      2,
					DefectosResueltos:       1,
				}, nil)
			},
			estado: http.StatusOK,
		},
		{
			nombre: "E-02 sprint sin historias devuelve nulos",
			preparar: func(m *mockServicioMetricas) {
				m.On("CalcularPorSprint", mock.Anything, uint(1), uint(10)).Return(domain.MetricasSprint{}, nil)
			},
			estado: http.StatusOK,
		},
	}

	for _, caso := range casos {
		caso := caso
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()
			router, servicio := nuevoRouterDeMetricas(t)
			caso.preparar(servicio)

			w := ejecutarMetricas(router, http.MethodGet, "/api/proyectos/1/sprints/10/metricas", "")

			assert.Equal(t, caso.estado, w.Code)
			var respuesta respuestaMetricasSprint
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &respuesta), "cuerpo: %s", w.Body.String())
			assert.Equal(t, uint(10), respuesta.SprintID)
			assert.Equal(t, "estado_actual", respuesta.Base)
			assert.Equal(t, []string{
				"base",
				"defectos_detectados",
				"defectos_resueltos",
				"desviacion_esfuerzo",
				"historias_completadas",
				"horas_estimadas",
				"horas_reales",
				"porcentaje_completado",
				"sprint_id",
				"story_points_completados",
				"story_points_planificados",
			}, clavesJSON(t, w.Body.Bytes()))
			servicio.AssertExpectations(t)
		})
	}
}

func TestMetricasHandlerSprintValoresYNumerosNulos(t *testing.T) {
	t.Parallel()
	router, servicio := nuevoRouterDeMetricas(t)
	porcentaje := 0.625
	servicio.On("CalcularPorSprint", mock.Anything, uint(1), uint(10)).Return(domain.MetricasSprint{
		StoryPointsPlanificados: 8,
		StoryPointsCompletados:  5,
		PorcentajeCompletado:    &porcentaje,
		HistoriasCompletadas:    1,
		HorasEstimadas:          20,
		HorasReales:             24,
		DesviacionEsfuerzo:      nil,
		DefectosDetectados:      2,
		DefectosResueltos:       1,
	}, nil)

	w := ejecutarMetricas(router, http.MethodGet, "/api/proyectos/1/sprints/10/metricas", "")
	require.Equal(t, http.StatusOK, w.Code)

	var cuerpo map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &cuerpo))
	assert.Equal(t, 8.0, cuerpo["story_points_planificados"])
	assert.Equal(t, 5.0, cuerpo["story_points_completados"])
	assert.Equal(t, 0.625, cuerpo["porcentaje_completado"])
	assert.Equal(t, 1.0, cuerpo["historias_completadas"])
	assert.Nil(t, cuerpo["desviacion_esfuerzo"])

	servicio.AssertExpectations(t)
}

func TestMetricasHandlerSprintSprintVacioDevuelveCerosYNull(t *testing.T) {
	t.Parallel()
	router, servicio := nuevoRouterDeMetricas(t)
	servicio.On("CalcularPorSprint", mock.Anything, uint(1), uint(10)).Return(domain.MetricasSprint{}, nil)

	w := ejecutarMetricas(router, http.MethodGet, "/api/proyectos/1/sprints/10/metricas", "")
	require.Equal(t, http.StatusOK, w.Code)

	var cuerpo map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &cuerpo))
	assert.Equal(t, 0.0, cuerpo["story_points_planificados"])
	assert.Equal(t, 0.0, cuerpo["story_points_completados"])
	assert.Nil(t, cuerpo["porcentaje_completado"])
	assert.Equal(t, 0.0, cuerpo["historias_completadas"])
	assert.Equal(t, 0.0, cuerpo["horas_estimadas"])
	assert.Equal(t, 0.0, cuerpo["horas_reales"])
	assert.Nil(t, cuerpo["desviacion_esfuerzo"])
	servicio.AssertExpectations(t)
}

func TestMetricasHandlerProyecto(t *testing.T) {
	t.Parallel()

	velocidad := 6.0
	porcentaje := 0.5
	desviacion := 0.2
	ratio := 1.0 / 3.0

	router, servicio := nuevoRouterDeMetricas(t)
	servicio.On("CalcularPorProyecto", mock.Anything, uint(1)).Return(domain.MetricasProyecto{
		Velocidad:                      &velocidad,
		PorcentajeHistoriasCompletadas: &porcentaje,
		HorasEstimadas:                 40,
		HorasReales:                    48,
		DesviacionEsfuerzo:             &desviacion,
		DefectosDetectados:             3,
		DefectosResueltos:              2,
		RatioDefectos:                  &ratio,
	}, nil)

	w := ejecutarMetricas(router, http.MethodGet, "/api/proyectos/1/metricas", "")

	require.Equal(t, http.StatusOK, w.Code)
	var respuesta respuestaMetricasProyecto
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &respuesta), "cuerpo: %s", w.Body.String())
	assert.Equal(t, uint(1), respuesta.ProyectoID)
	assert.Equal(t, "estado_actual", respuesta.Base)
	assert.Equal(t, 6.0, *respuesta.Velocidad)
	assert.Equal(t, 0.5, *respuesta.PorcentajeHistoriasCompletadas)
	assert.InDelta(t, 1.0/3.0, *respuesta.RatioDefectos, 1e-9)

	assert.Equal(t, []string{
		"base",
		"defectos_abiertos",
		"defectos_detectados",
		"defectos_por_severidad",
		"defectos_resueltos",
		"desviacion_esfuerzo",
		"horas_estimadas",
		"horas_reales",
		"lead_time_aproximado",
		"lead_time_estimado",
		"porcentaje_historias_completadas",
		"proyecto_id",
		"ratio_defectos",
		"sprints_restantes",
		"tasa_entrega",
		"velocidad",
		"wip",
	}, clavesJSON(t, w.Body.Bytes()))
	assert.NotContains(t, clavesJSON(t, w.Body.Bytes()), "historias_completadas")
	assert.NotContains(t, clavesJSON(t, w.Body.Bytes()), "historias_totales")

	servicio.AssertExpectations(t)
}

func TestMetricasHandlerProyectoExtensiones(t *testing.T) {
	t.Parallel()

	sprintsRestantes := 7.0
	tasaEntrega := 2.0
	leadTime := 3.0

	router, servicio := nuevoRouterDeMetricas(t)
	servicio.On("CalcularPorProyecto", mock.Anything, uint(1)).Return(domain.MetricasProyecto{
		SprintsRestantes:   &sprintsRestantes,
		WIP:                6,
		TasaEntrega:        &tasaEntrega,
		LeadTimeEstimado:   &leadTime,
		LeadTimeAproximado: true,
		DefectosAbiertos:   2,
		DefectosPorSeveridad: domain.DefectosPorSeveridad{
			Minor:    1,
			Major:    1,
			Critical: 2,
			Blocker:  0,
		},
	}, nil)

	w := ejecutarMetricas(router, http.MethodGet, "/api/proyectos/1/metricas", "")
	require.Equal(t, http.StatusOK, w.Code)

	var cuerpo map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &cuerpo), "cuerpo: %s", w.Body.String())
	assert.Equal(t, 7.0, cuerpo["sprints_restantes"])
	assert.Equal(t, 6.0, cuerpo["wip"])
	assert.Equal(t, 2.0, cuerpo["tasa_entrega"])
	assert.Equal(t, 3.0, cuerpo["lead_time_estimado"])
	assert.Equal(t, true, cuerpo["lead_time_aproximado"])
	assert.Equal(t, 2.0, cuerpo["defectos_abiertos"])

	severidad, ok := cuerpo["defectos_por_severidad"].(map[string]any)
	require.True(t, ok, "defectos_por_severidad debe ser un objeto")
	assert.Equal(t, map[string]any{
		"minor":    1.0,
		"major":    1.0,
		"critical": 2.0,
		"blocker":  0.0,
	}, severidad)

	servicio.AssertExpectations(t)
}

func TestMetricasHandlerProyectoExtensionesSinDatos(t *testing.T) {
	t.Parallel()

	router, servicio := nuevoRouterDeMetricas(t)
	servicio.On("CalcularPorProyecto", mock.Anything, uint(1)).Return(domain.MetricasProyecto{}, nil)

	w := ejecutarMetricas(router, http.MethodGet, "/api/proyectos/1/metricas", "")
	require.Equal(t, http.StatusOK, w.Code)

	var cuerpo map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &cuerpo), "cuerpo: %s", w.Body.String())
	assert.Nil(t, cuerpo["sprints_restantes"])
	assert.Equal(t, 0.0, cuerpo["wip"])
	assert.Nil(t, cuerpo["tasa_entrega"])
	assert.Nil(t, cuerpo["lead_time_estimado"])
	assert.Equal(t, false, cuerpo["lead_time_aproximado"])
	assert.Equal(t, 0.0, cuerpo["defectos_abiertos"])

	severidad, ok := cuerpo["defectos_por_severidad"].(map[string]any)
	require.True(t, ok, "defectos_por_severidad debe ser un objeto incluso sin datos")
	assert.Equal(t, map[string]any{
		"minor":    0.0,
		"major":    0.0,
		"critical": 0.0,
		"blocker":  0.0,
	}, severidad)

	servicio.AssertExpectations(t)
}

func TestMetricasHandlerProyectoVacioDevuelveNull(t *testing.T) {
	t.Parallel()
	router, servicio := nuevoRouterDeMetricas(t)
	servicio.On("CalcularPorProyecto", mock.Anything, uint(1)).Return(domain.MetricasProyecto{}, nil)

	w := ejecutarMetricas(router, http.MethodGet, "/api/proyectos/1/metricas", "")
	require.Equal(t, http.StatusOK, w.Code)

	var cuerpo map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &cuerpo))
	assert.Nil(t, cuerpo["velocidad"])
	assert.Nil(t, cuerpo["porcentaje_historias_completadas"])
	assert.Nil(t, cuerpo["desviacion_esfuerzo"])
	assert.Nil(t, cuerpo["ratio_defectos"])
	assert.Equal(t, 0.0, cuerpo["horas_estimadas"])
	assert.Equal(t, 0.0, cuerpo["horas_reales"])
	servicio.AssertExpectations(t)
}

func TestMetricasHandlerParametrosInvalidos(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string
		ruta   string
	}{
		{nombre: "id no numerico", ruta: "/api/proyectos/abc/metricas"},
		{nombre: "id cero", ruta: "/api/proyectos/0/metricas"},
		{nombre: "sprintId no numerico", ruta: "/api/proyectos/1/sprints/abc/metricas"},
		{nombre: "sprintId cero", ruta: "/api/proyectos/1/sprints/0/metricas"},
	}

	for _, caso := range casos {
		caso := caso
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()
			router, servicio := nuevoRouterDeMetricas(t)

			w := ejecutarMetricas(router, http.MethodGet, caso.ruta, "")

			assert.Equal(t, http.StatusBadRequest, w.Code)
			var respuesta respuestaError
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &respuesta))
			assert.Equal(t, "PARAMETRO_INVALIDO", respuesta.Error)
			assert.Empty(t, clavesQueIncluyanBase(t, w.Body.Bytes()))
			servicio.AssertNotCalled(t, "CalcularPorSprint", mock.Anything, mock.Anything, mock.Anything)
			servicio.AssertNotCalled(t, "CalcularPorProyecto", mock.Anything, mock.Anything)
		})
	}
}

func clavesQueIncluyanBase(t *testing.T, cuerpo []byte) []string {
	t.Helper()
	var objeto map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(cuerpo, &objeto))
	claves := make([]string, 0)
	for clave := range objeto {
		if clave == "base" {
			claves = append(claves, clave)
		}
	}
	return claves
}

func TestMetricasHandlerErroresDeDominio(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre   string
		ruta     string
		preparar func(m *mockServicioMetricas)
		estado   int
		codigo   string
	}{
		{
			nombre: "E-12 proyecto no encontrado en sprint",
			ruta:   "/api/proyectos/1/sprints/10/metricas",
			preparar: func(m *mockServicioMetricas) {
				m.On("CalcularPorSprint", mock.Anything, uint(1), uint(10)).Return(domain.MetricasSprint{}, domain.ErrProyectoNoEncontrado)
			},
			estado: http.StatusNotFound,
			codigo: "PROYECTO_NO_ENCONTRADO",
		},
		{
			nombre: "E-12 sprint no encontrado",
			ruta:   "/api/proyectos/1/sprints/99/metricas",
			preparar: func(m *mockServicioMetricas) {
				m.On("CalcularPorSprint", mock.Anything, uint(1), uint(99)).Return(domain.MetricasSprint{}, domain.ErrSprintNoEncontrado)
			},
			estado: http.StatusNotFound,
			codigo: "SPRINT_NO_ENCONTRADA",
		},
		{
			nombre: "E-12 proyecto no encontrado en proyecto",
			ruta:   "/api/proyectos/1/metricas",
			preparar: func(m *mockServicioMetricas) {
				m.On("CalcularPorProyecto", mock.Anything, uint(1)).Return(domain.MetricasProyecto{}, domain.ErrProyectoNoEncontrado)
			},
			estado: http.StatusNotFound,
			codigo: "PROYECTO_NO_ENCONTRADO",
		},
		{
			nombre: "error de persistencia sprint",
			ruta:   "/api/proyectos/1/sprints/10/metricas",
			preparar: func(m *mockServicioMetricas) {
				m.On("CalcularPorSprint", mock.Anything, uint(1), uint(10)).Return(domain.MetricasSprint{}, errors.New("db caida"))
			},
			estado: http.StatusInternalServerError,
			codigo: "ERROR_INTERNO",
		},
		{
			nombre: "error de persistencia proyecto",
			ruta:   "/api/proyectos/1/metricas",
			preparar: func(m *mockServicioMetricas) {
				m.On("CalcularPorProyecto", mock.Anything, uint(1)).Return(domain.MetricasProyecto{}, errors.New("db caida"))
			},
			estado: http.StatusInternalServerError,
			codigo: "ERROR_INTERNO",
		},
	}

	for _, caso := range casos {
		caso := caso
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()
			router, servicio := nuevoRouterDeMetricas(t)
			caso.preparar(servicio)

			w := ejecutarMetricas(router, http.MethodGet, caso.ruta, "")

			assert.Equal(t, caso.estado, w.Code)
			var respuesta respuestaError
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &respuesta))
			assert.Equal(t, caso.codigo, respuesta.Error)
			assert.Empty(t, clavesQueIncluyanBase(t, w.Body.Bytes()))
			servicio.AssertExpectations(t)
		})
	}
}

func TestMetricasHandlerIgnoraCuerpoYQuery(t *testing.T) {
	t.Parallel()
	router, servicio := nuevoRouterDeMetricas(t)
	servicio.On("CalcularPorProyecto", mock.Anything, uint(1)).Return(domain.MetricasProyecto{}, nil)

	w := ejecutarMetricas(router, http.MethodGet, "/api/proyectos/1/metricas?ignorado=1", `{"cualquier":"cosa"}`)

	assert.Equal(t, http.StatusOK, w.Code)
	servicio.AssertExpectations(t)
}
