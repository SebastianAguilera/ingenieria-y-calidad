package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// BDD-06: el Sprint Backlog (US-13) reutiliza la representación de historia y
// por lo tanto expone horas_estimadas sin cambios adicionales.
func TestSprintBacklogHandlerListarIncluyeHorasEstimadas(t *testing.T) {
	gin.SetMode(gin.TestMode)

	casos := []struct {
		nombre   string
		horas    *float64
		esperado string
	}{
		{"con horas estimadas", floatPtr(8), `"horas_estimadas":8`},
		{"sin horas estimadas emite null", nil, `"horas_estimadas":null`},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			detalle := sprintBacklogDetalleDePrueba()
			detalle.Items[0].Historia.HorasEstimadas = c.horas

			servicio := &mockServicioSprintBacklog{}
			servicio.On("ListarSprintBacklog", mock.Anything, uint(1), uint(10)).Return(detalle, nil)
			handler := NuevoSprintBacklogHandler(servicio)
			r := gin.New()
			r.GET("/api/proyectos/:id/sprints/:sprintId/historias", handler.Listar)

			req := httptest.NewRequest(http.MethodGet, "/api/proyectos/1/sprints/10/historias", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			require.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), c.esperado)

			var resp respuestaSprintBacklog
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			require.Len(t, resp.Items, 1)
			assert.Equal(t, c.horas, resp.Items[0].Historia.HorasEstimadas)
			servicio.AssertExpectations(t)
		})
	}
}
