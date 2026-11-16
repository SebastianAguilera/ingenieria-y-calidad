package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

type mockServicioPDF struct {
	mock.Mock
}

func (m *mockServicioPDF) GenerarPDFProyecto(ctx context.Context, proyectoID uint) ([]byte, string, error) {
	args := m.Called(ctx, proyectoID)
	return args.Get(0).([]byte), args.String(1), args.Error(2)
}

func (m *mockServicioPDF) GenerarPDFSprint(ctx context.Context, proyectoID, sprintID uint) ([]byte, string, error) {
	args := m.Called(ctx, proyectoID, sprintID)
	return args.Get(0).([]byte), args.String(1), args.Error(2)
}

func nuevoRouterDePDF(t *testing.T) (*gin.Engine, *mockServicioPDF) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	servicio := new(mockServicioPDF)
	router := gin.New()
	RegistrarRutasPDF(router, servicio)
	return router, servicio
}

var pdfValido = []byte("%PDF-1.4 fake pdf content for testing")

func TestPDFHandlerProyecto(t *testing.T) {
	t.Parallel()

	t.Run("E-01 descarga PDF de proyecto exitosamente", func(t *testing.T) {
		t.Parallel()
		router, servicio := nuevoRouterDePDF(t)

		servicio.On("GenerarPDFProyecto", mock.Anything, uint(1)).
			Return(pdfValido, "reporte-proyecto-alpha-proyecto-2026-11-15.pdf", nil).Once()

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/proyectos/1/reporte.pdf", nil)
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
		disposition := w.Header().Get("Content-Disposition")
		assert.Contains(t, disposition, "attachment")
		assert.Contains(t, disposition, ".pdf")
		assert.True(t, strings.HasPrefix(w.Body.String(), "%PDF"))
		servicio.AssertExpectations(t)
	})

	t.Run("E-04 proyecto sin historias genera PDF sin error", func(t *testing.T) {
		t.Parallel()
		router, servicio := nuevoRouterDePDF(t)

		servicio.On("GenerarPDFProyecto", mock.Anything, uint(2)).
			Return(pdfValido, "reporte-proyecto-vacio-proyecto-2026-11-15.pdf", nil).Once()

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/proyectos/2/reporte.pdf", nil)
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
		servicio.AssertExpectations(t)
	})

	t.Run("proyecto no encontrado devuelve 404", func(t *testing.T) {
		t.Parallel()
		router, servicio := nuevoRouterDePDF(t)

		servicio.On("GenerarPDFProyecto", mock.Anything, uint(999)).
			Return([]byte{}, "", domain.ErrProyectoNoEncontrado).Once()

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/proyectos/999/reporte.pdf", nil)
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "PROYECTO_NO_ENCONTRADO")
		servicio.AssertExpectations(t)
	})

	t.Run("id invalido devuelve 400", func(t *testing.T) {
		t.Parallel()
		router, servicio := nuevoRouterDePDF(t)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/proyectos/abc/reporte.pdf", nil)
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		servicio.AssertExpectations(t)
	})
}

func TestPDFHandlerSprint(t *testing.T) {
	t.Parallel()

	t.Run("E-02 descarga PDF de sprint exitosamente", func(t *testing.T) {
		t.Parallel()
		router, servicio := nuevoRouterDePDF(t)

		servicio.On("GenerarPDFSprint", mock.Anything, uint(1), uint(10)).
			Return(pdfValido, "reporte-proyecto-alpha-sprint-10-2026-11-15.pdf", nil).Once()

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/proyectos/1/sprints/10/reporte.pdf", nil)
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
		disposition := w.Header().Get("Content-Disposition")
		assert.Contains(t, disposition, "attachment")
		assert.Contains(t, disposition, ".pdf")
		servicio.AssertExpectations(t)
	})

	t.Run("E-06 sprint no encontrado devuelve 404 SPRINT_NO_ENCONTRADA", func(t *testing.T) {
		t.Parallel()
		router, servicio := nuevoRouterDePDF(t)

		servicio.On("GenerarPDFSprint", mock.Anything, uint(1), uint(999)).
			Return([]byte{}, "", domain.ErrSprintNoEncontrado).Once()

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/proyectos/1/sprints/999/reporte.pdf", nil)
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "SPRINT_NO_ENCONTRADA")
		servicio.AssertExpectations(t)
	})
}
