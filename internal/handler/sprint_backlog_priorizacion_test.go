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

func TestSprintBacklogHandlerListarIncluyeValoresDeNegocio(t *testing.T) {
	gin.SetMode(gin.TestMode)

	casos := []struct {
		nombre           string
		valorNegocio     *float64
		criticidadTiempo *float64
		reduccionRiesgo  *float64
	}{
		{"con los tres valores", floatPtr(8), floatPtr(3), floatPtr(2)},
		{"sin valores emite null", nil, nil, nil},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			detalle := sprintBacklogDetalleDePrueba()
			detalle.Items[0].Historia.ValorNegocio = c.valorNegocio
			detalle.Items[0].Historia.CriticidadTiempo = c.criticidadTiempo
			detalle.Items[0].Historia.ReduccionRiesgo = c.reduccionRiesgo

			servicio := &mockServicioSprintBacklog{}
			servicio.On("ListarSprintBacklog", mock.Anything, uint(1), uint(10)).Return(detalle, nil)
			handler := NuevoSprintBacklogHandler(servicio)
			r := gin.New()
			r.GET("/api/proyectos/:id/sprints/:sprintId/historias", handler.Listar)

			req := httptest.NewRequest(http.MethodGet, "/api/proyectos/1/sprints/10/historias", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			require.Equal(t, http.StatusOK, w.Code)
			var resp respuestaSprintBacklog
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			require.Len(t, resp.Items, 1)
			assert.Equal(t, c.valorNegocio, resp.Items[0].Historia.ValorNegocio)
			assert.Equal(t, c.criticidadTiempo, resp.Items[0].Historia.CriticidadTiempo)
			assert.Equal(t, c.reduccionRiesgo, resp.Items[0].Historia.ReduccionRiesgo)
			servicio.AssertExpectations(t)
		})
	}
}
