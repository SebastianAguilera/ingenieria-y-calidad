package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

type mockServicioReporte struct {
	mock.Mock
}

func (m *mockServicioReporte) ObtenerReporteProyecto(ctx context.Context, proyectoID uint) (domain.ReporteProyecto, error) {
	args := m.Called(ctx, proyectoID)
	return args.Get(0).(domain.ReporteProyecto), args.Error(1)
}

func (m *mockServicioReporte) ObtenerReporteSprint(ctx context.Context, proyectoID, sprintID uint) (domain.ReporteSprint, error) {
	args := m.Called(ctx, proyectoID, sprintID)
	return args.Get(0).(domain.ReporteSprint), args.Error(1)
}

func nuevoRouterDeReporte(t *testing.T) (*gin.Engine, *mockServicioReporte) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	servicio := new(mockServicioReporte)
	router := gin.New()
	RegistrarRutasReporte(router, servicio)
	return router, servicio
}

func TestReporteHandlerProyecto(t *testing.T) {
	t.Parallel()

	t.Run("E-01 proyecto con historias y defectos", func(t *testing.T) {
		t.Parallel()
		router, servicio := nuevoRouterDeReporte(t)

		fecha := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
		sp := 5.0
		vel := 5.0
		servicio.On("ObtenerReporteProyecto", mock.Anything, uint(1)).Return(domain.ReporteProyecto{
			Base:        "estado_actual",
			ProyectoID:  1,
			Nombre:      "Proyecto Alpha",
			Descripcion: "Desc",
			FechaInicio: fecha,
			Estado:      domain.EstadoActivo,
			Historias: []domain.Historia{
				{
					ID:          10,
					Codigo:      "US-01",
					Titulo:      "Login",
					Estado:      domain.EstadoHistoriaHecho,
					StoryPoints: &sp,
				},
			},
			Metricas: domain.MetricasProyecto{
				Velocidad: &vel,
			},
			Defectos: []domain.Defecto{
				{
					ID:                100,
					Descripcion:       "Bug login",
					Severidad:         domain.SeveridadCritical,
					Estado:            domain.EstadoDefectoAbierto,
					SprintDeteccionID: 2,
				},
			},
		}, nil).Once()

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/proyectos/1/reporte", nil)
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var resp respuestaReporteProyecto
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "estado_actual", resp.Base)
		assert.Equal(t, "Proyecto Alpha", resp.Nombre)
		require.Len(t, resp.Historias, 1)
		assert.Nil(t, resp.Historias[0].HorasEstimadas)
		require.Len(t, resp.Defectos, 1)
		assert.Equal(t, "DEF-100", resp.Defectos[0].Codigo)
		assert.NotNil(t, resp.Metricas.Velocidad)
		assert.Equal(t, 5.0, *resp.Metricas.Velocidad)
		servicio.AssertExpectations(t)
	})

	t.Run("E-03 proyecto sin historias devuelve listas vacias no null", func(t *testing.T) {
		t.Parallel()
		router, servicio := nuevoRouterDeReporte(t)

		fecha := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
		servicio.On("ObtenerReporteProyecto", mock.Anything, uint(1)).Return(domain.ReporteProyecto{
			Base:        "estado_actual",
			ProyectoID:  1,
			Nombre:      "Proyecto Vacio",
			FechaInicio: fecha,
			Estado:      domain.EstadoActivo,
			Historias:   []domain.Historia{},
			Defectos:    []domain.Defecto{},
		}, nil).Once()

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/proyectos/1/reporte", nil)
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"historias":[]`)
		assert.Contains(t, w.Body.String(), `"defectos":[]`)
		servicio.AssertExpectations(t)
	})
}

func TestReporteHandlerSprint(t *testing.T) {
	t.Parallel()

	t.Run("E-02 sprint con historias y defectos", func(t *testing.T) {
		t.Parallel()
		router, servicio := nuevoRouterDeReporte(t)

		fecha := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
		servicio.On("ObtenerReporteSprint", mock.Anything, uint(1), uint(10)).Return(domain.ReporteSprint{
			Base:          "estado_actual",
			ProyectoID:    1,
			SprintID:      10,
			Identificador: "Sprint 10",
			SprintGoal:    "Meta",
			FechaInicio:   fecha,
			FechaFin:      fecha.AddDate(0, 0, 14),
			Estado:        domain.EstadoSprintActivo,
			Historias:     []domain.Historia{},
			Defectos:      []domain.Defecto{},
		}, nil).Once()

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/proyectos/1/sprints/10/reporte", nil)
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var resp respuestaReporteSprint
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "estado_actual", resp.Base)
		assert.Equal(t, uint(10), resp.SprintID)
		assert.Equal(t, "Sprint 10", resp.Identificador)
		assert.Empty(t, resp.Historias)
		assert.Empty(t, resp.Defectos)
		servicio.AssertExpectations(t)
	})

	t.Run("E-06 sprint no encontrado devuelve 404 SPRINT_NO_ENCONTRADA", func(t *testing.T) {
		t.Parallel()
		router, servicio := nuevoRouterDeReporte(t)

		servicio.On("ObtenerReporteSprint", mock.Anything, uint(1), uint(999)).Return(domain.ReporteSprint{}, domain.ErrSprintNoEncontrado).Once()

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/proyectos/1/sprints/999/reporte", nil)
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "SPRINT_NO_ENCONTRADA")
		servicio.AssertExpectations(t)
	})
}
