package handler

import (
	"bytes"
	"context"
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

type mockServicioSprintBacklog struct {
	mock.Mock
}

func (m *mockServicioSprintBacklog) AsignarASprint(ctx context.Context, proyectoID, sprintID, historiaID uint) error {
	return m.Called(ctx, proyectoID, sprintID, historiaID).Error(0)
}

func (m *mockServicioSprintBacklog) DesasignarDeSprint(ctx context.Context, proyectoID, sprintID, historiaID uint) error {
	return m.Called(ctx, proyectoID, sprintID, historiaID).Error(0)
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
