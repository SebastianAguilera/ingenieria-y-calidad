package handler

import (
	"bytes"
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

type mockServicioSprintBacklog struct {
	mock.Mock
}

func (m *mockServicioSprintBacklog) AsignarASprint(ctx context.Context, proyectoID, sprintID, historiaID uint) error {
	return m.Called(ctx, proyectoID, sprintID, historiaID).Error(0)
}

func (m *mockServicioSprintBacklog) DesasignarDeSprint(ctx context.Context, proyectoID, sprintID, historiaID uint) error {
	return m.Called(ctx, proyectoID, sprintID, historiaID).Error(0)
}

func (m *mockServicioSprintBacklog) ListarSprintBacklog(ctx context.Context, proyectoID, sprintID uint) (*domain.SprintBacklogDetalle, error) {
	ret := m.Called(ctx, proyectoID, sprintID)
	var detalle *domain.SprintBacklogDetalle
	if ret.Get(0) != nil {
		detalle = ret.Get(0).(*domain.SprintBacklogDetalle)
	}
	return detalle, ret.Error(1)
}

func TestSprintBacklogHandlerAsignar(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("asigna historia correctamente", func(t *testing.T) {
		servicio := &mockServicioSprintBacklog{}
		servicio.On("AsignarASprint", mock.Anything, uint(1), uint(10), uint(100)).Return(nil)
		handler := NuevoSprintBacklogHandler(servicio)
		r := gin.New()
		r.POST("/api/proyectos/:id/sprints/:sprintId/historias", handler.Asignar)

		body := bytes.NewBufferString(`{"historia_id":100}`)
		req := httptest.NewRequest(http.MethodPost, "/api/proyectos/1/sprints/10/historias", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusNoContent, w.Code)
		servicio.AssertExpectations(t)
	})

	t.Run("historia ya asignada", func(t *testing.T) {
		servicio := &mockServicioSprintBacklog{}
		servicio.On("AsignarASprint", mock.Anything, uint(1), uint(10), uint(100)).Return(domain.ErrHistoriaYaAsignada)
		handler := NuevoSprintBacklogHandler(servicio)
		r := gin.New()
		r.POST("/api/proyectos/:id/sprints/:sprintId/historias", handler.Asignar)

		body := bytes.NewBufferString(`{"historia_id":100}`)
		req := httptest.NewRequest(http.MethodPost, "/api/proyectos/1/sprints/10/historias", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusConflict, w.Code)
		var resp respuestaError
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "HISTORIA_YA_ASIGNADA", resp.Error)
		servicio.AssertExpectations(t)
	})

	t.Run("sprint no activo", func(t *testing.T) {
		servicio := &mockServicioSprintBacklog{}
		servicio.On("AsignarASprint", mock.Anything, uint(1), uint(10), uint(100)).Return(domain.ErrSprintNoActivo)
		handler := NuevoSprintBacklogHandler(servicio)
		r := gin.New()
		r.POST("/api/proyectos/:id/sprints/:sprintId/historias", handler.Asignar)

		body := bytes.NewBufferString(`{"historia_id":100}`)
		req := httptest.NewRequest(http.MethodPost, "/api/proyectos/1/sprints/10/historias", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		var resp respuestaError
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "SPRINT_NO_ACTIVO", resp.Error)
		servicio.AssertExpectations(t)
	})

	t.Run("historia no pendiente", func(t *testing.T) {
		servicio := &mockServicioSprintBacklog{}
		servicio.On("AsignarASprint", mock.Anything, uint(1), uint(10), uint(100)).Return(domain.ErrHistoriaNoPendiente)
		handler := NuevoSprintBacklogHandler(servicio)
		r := gin.New()
		r.POST("/api/proyectos/:id/sprints/:sprintId/historias", handler.Asignar)

		body := bytes.NewBufferString(`{"historia_id":100}`)
		req := httptest.NewRequest(http.MethodPost, "/api/proyectos/1/sprints/10/historias", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		var resp respuestaError
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "HISTORIA_NO_PENDIENTE", resp.Error)
		servicio.AssertExpectations(t)
	})

	t.Run("fuera de ventana temporal", func(t *testing.T) {
		servicio := &mockServicioSprintBacklog{}
		servicio.On("AsignarASprint", mock.Anything, uint(1), uint(10), uint(100)).Return(domain.ErrFueraDeVentanaTemporal)
		handler := NuevoSprintBacklogHandler(servicio)
		r := gin.New()
		r.POST("/api/proyectos/:id/sprints/:sprintId/historias", handler.Asignar)

		body := bytes.NewBufferString(`{"historia_id":100}`)
		req := httptest.NewRequest(http.MethodPost, "/api/proyectos/1/sprints/10/historias", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		var resp respuestaError
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "FUERA_DE_VENTANA_TEMPORAL", resp.Error)
		servicio.AssertExpectations(t)
	})
}

func TestSprintBacklogHandlerDesasignar(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("desasigna historia correctamente", func(t *testing.T) {
		servicio := &mockServicioSprintBacklog{}
		servicio.On("DesasignarDeSprint", mock.Anything, uint(1), uint(10), uint(100)).Return(nil)
		handler := NuevoSprintBacklogHandler(servicio)
		r := gin.New()
		r.DELETE("/api/proyectos/:id/sprints/:sprintId/historias/:historiaId", handler.Desasignar)

		req := httptest.NewRequest(http.MethodDelete, "/api/proyectos/1/sprints/10/historias/100", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusNoContent, w.Code)
		servicio.AssertExpectations(t)
	})

	t.Run("sprint cerrado no permite desasignar", func(t *testing.T) {
		servicio := &mockServicioSprintBacklog{}
		servicio.On("DesasignarDeSprint", mock.Anything, uint(1), uint(10), uint(100)).Return(domain.ErrSprintCerrado)
		handler := NuevoSprintBacklogHandler(servicio)
		r := gin.New()
		r.DELETE("/api/proyectos/:id/sprints/:sprintId/historias/:historiaId", handler.Desasignar)

		req := httptest.NewRequest(http.MethodDelete, "/api/proyectos/1/sprints/10/historias/100", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusConflict, w.Code)
		var resp respuestaError
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "SPRINT_CERRADO", resp.Error)
		servicio.AssertExpectations(t)
	})

	t.Run("historia no pertenece al sprint", func(t *testing.T) {
		servicio := &mockServicioSprintBacklog{}
		servicio.On("DesasignarDeSprint", mock.Anything, uint(1), uint(10), uint(100)).Return(domain.ErrHistoriaNoPerteneceASprint)
		handler := NuevoSprintBacklogHandler(servicio)
		r := gin.New()
		r.DELETE("/api/proyectos/:id/sprints/:sprintId/historias/:historiaId", handler.Desasignar)

		req := httptest.NewRequest(http.MethodDelete, "/api/proyectos/1/sprints/10/historias/100", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusConflict, w.Code)
		var resp respuestaError
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "HISTORIA_NO_PERTENECE_A_SPRINT", resp.Error)
		servicio.AssertExpectations(t)
	})
}

func sprintBacklogDetalleDePrueba() *domain.SprintBacklogDetalle {
	inicio := time.Date(2026, time.October, 6, 0, 0, 0, 0, time.UTC)
	fin := time.Date(2026, time.October, 20, 0, 0, 0, 0, time.UTC)
	asignada := time.Date(2026, time.October, 6, 10, 0, 0, 0, time.UTC)
	puntos := 3.0
	return &domain.SprintBacklogDetalle{
		Sprint: &domain.Sprint{
			ID:            1,
			ProyectoID:    1,
			Identificador: "Sprint 1",
			SprintGoal:    "Entregar el MVP",
			Estado:        domain.EstadoSprintActivo,
			FechaInicio:   inicio,
			FechaFin:      fin,
			CreadoEn:      asignada,
			ActualizadoEn: asignada,
		},
		Items: []domain.SprintBacklogItemDetalle{
			{
				Item: domain.SprintBacklogItem{
					ID:              10,
					ProyectoID:      1,
					SprintID:        1,
					HistoriaID:      100,
					Orden:           1,
					FechaAsignacion: asignada,
				},
				Historia: domain.Historia{
					ID:            100,
					ProyectoID:    1,
					Numero:        1,
					Codigo:        "US-01",
					Titulo:        "Crear proyecto",
					Descripcion:   "Como usuario...",
					Prioridad:     domain.PrioridadAlta,
					Estado:        domain.EstadoHistoriaEnProgreso,
					StoryPoints:   &puntos,
					CreadoEn:      asignada,
					ActualizadoEn: asignada,
				},
			},
		},
	}
}

func TestSprintBacklogHandlerListar(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("E-01 responde 200 con sprint, total e items", func(t *testing.T) {
		servicio := &mockServicioSprintBacklog{}
		servicio.On("ListarSprintBacklog", mock.Anything, uint(1), uint(10)).
			Return(sprintBacklogDetalleDePrueba(), nil)
		handler := NuevoSprintBacklogHandler(servicio)
		r := gin.New()
		r.GET("/api/proyectos/:id/sprints/:sprintId/historias", handler.Listar)

		req := httptest.NewRequest(http.MethodGet, "/api/proyectos/1/sprints/10/historias", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var resp respuestaSprintBacklog
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, uint(1), resp.Sprint.ID)
		assert.Equal(t, "Sprint 1", resp.Sprint.Identificador)
		assert.Equal(t, "Activo", string(resp.Sprint.Estado))
		assert.Equal(t, 1, resp.Total)
		require.Len(t, resp.Items, 1)
		assert.Equal(t, uint(10), resp.Items[0].ID)
		assert.Equal(t, 1, resp.Items[0].Orden)
		assert.Equal(t, "2026-10-06T10:00:00Z", resp.Items[0].FechaAsignacion)
		assert.Equal(t, uint(100), resp.Items[0].Historia.ID)
		assert.Equal(t, "US-01", resp.Items[0].Historia.Codigo)
		assert.Equal(t, "En Progreso", string(resp.Items[0].Historia.Estado))
		servicio.AssertExpectations(t)
	})

	t.Run("E-02 sprint sin historias emite items vacio, no null", func(t *testing.T) {
		detalle := sprintBacklogDetalleDePrueba()
		detalle.Items = []domain.SprintBacklogItemDetalle{}
		servicio := &mockServicioSprintBacklog{}
		servicio.On("ListarSprintBacklog", mock.Anything, uint(1), uint(10)).Return(detalle, nil)
		handler := NuevoSprintBacklogHandler(servicio)
		r := gin.New()
		r.GET("/api/proyectos/:id/sprints/:sprintId/historias", handler.Listar)

		req := httptest.NewRequest(http.MethodGet, "/api/proyectos/1/sprints/10/historias", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var cuerpo map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &cuerpo))
		assert.JSONEq(t, `[]`, string(cuerpo["items"]), "escenario BDD E-02")
		assert.JSONEq(t, `0`, string(cuerpo["total"]))
		servicio.AssertExpectations(t)
	})

	t.Run("E-06 sprint inexistente responde 404", func(t *testing.T) {
		servicio := &mockServicioSprintBacklog{}
		servicio.On("ListarSprintBacklog", mock.Anything, uint(1), uint(99)).
			Return(nil, domain.ErrSprintNoEncontrado)
		handler := NuevoSprintBacklogHandler(servicio)
		r := gin.New()
		r.GET("/api/proyectos/:id/sprints/:sprintId/historias", handler.Listar)

		req := httptest.NewRequest(http.MethodGet, "/api/proyectos/1/sprints/99/historias", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusNotFound, w.Code)
		var resp respuestaError
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "SPRINT_NO_ENCONTRADA", resp.Error)
		servicio.AssertExpectations(t)
	})

	t.Run("E-08 proyecto inexistente responde 404", func(t *testing.T) {
		servicio := &mockServicioSprintBacklog{}
		servicio.On("ListarSprintBacklog", mock.Anything, uint(99), uint(10)).
			Return(nil, domain.ErrProyectoNoEncontrado)
		handler := NuevoSprintBacklogHandler(servicio)
		r := gin.New()
		r.GET("/api/proyectos/:id/sprints/:sprintId/historias", handler.Listar)

		req := httptest.NewRequest(http.MethodGet, "/api/proyectos/99/sprints/10/historias", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusNotFound, w.Code)
		var resp respuestaError
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "PROYECTO_NO_ENCONTRADO", resp.Error)
		servicio.AssertExpectations(t)
	})

	t.Run("E-10 id de proyecto invalido responde 400", func(t *testing.T) {
		servicio := &mockServicioSprintBacklog{}
		handler := NuevoSprintBacklogHandler(servicio)
		r := gin.New()
		r.GET("/api/proyectos/:id/sprints/:sprintId/historias", handler.Listar)

		req := httptest.NewRequest(http.MethodGet, "/api/proyectos/abc/sprints/10/historias", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		var resp respuestaError
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "PARAMETRO_INVALIDO", resp.Error)
		servicio.AssertNotCalled(t, "ListarSprintBacklog", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("E-11 id de sprint invalido responde 400", func(t *testing.T) {
		servicio := &mockServicioSprintBacklog{}
		handler := NuevoSprintBacklogHandler(servicio)
		r := gin.New()
		r.GET("/api/proyectos/:id/sprints/:sprintId/historias", handler.Listar)

		req := httptest.NewRequest(http.MethodGet, "/api/proyectos/1/sprints/abc/historias", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		var resp respuestaError
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "PARAMETRO_INVALIDO", resp.Error)
		servicio.AssertNotCalled(t, "ListarSprintBacklog", mock.Anything, mock.Anything, mock.Anything)
	})
}
