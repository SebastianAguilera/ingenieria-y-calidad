package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValorPriorizacionValido(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		valor     float64
		esperado  bool
		escenario string
	}{
		{"uno", 1, true, "E-05"},
		{"dos", 2, true, "E-05"},
		{"tres", 3, true, "E-05"},
		{"cinco", 5, true, "E-05"},
		{"ocho", 8, true, "E-05"},
		{"trece", 13, true, "E-05"},
		{"veintiuno", 21, true, "E-05"},
		{"cero fuera de escala", 0, false, "E-05"},
		{"medio punto fuera de escala", 0.5, false, "E-05"},
		{"cuatro fuera de escala", 4, false, "E-05"},
		{"un decimal fuera de escala", 1.5, false, "E-05"},
		{"veintidos fuera de escala", 22, false, "E-05"},
		{"negativo", -1, false, "E-05"},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, c.esperado, ValorPriorizacionValido(c.valor), "escenario BDD %s", c.escenario)
		})
	}
}

func TestCalcularROI(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre       string
		valorNegocio *float64
		storyPoints  *float64
		esperado     *float64
		escenario    string
	}{
		{"ocho sobre cinco", punteroFloat64(8), punteroFloat64(5), punteroFloat64(1.6), "E-01"},
		{"tres sobre tres", punteroFloat64(3), punteroFloat64(3), punteroFloat64(1), "CA-01"},
		{"sin story points", punteroFloat64(8), nil, nil, "E-03"},
		{"story points cero", punteroFloat64(8), punteroFloat64(0), nil, "E-03"},
		{"sin valor de negocio", nil, punteroFloat64(5), nil, "E-04"},
		{"medio punto", punteroFloat64(3), punteroFloat64(0.5), punteroFloat64(6), "CL-06"},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			assertFloatPtr(t, c.esperado, CalcularROI(c.valorNegocio, c.storyPoints), "escenario BDD %s", c.escenario)
		})
	}
}

func TestCalcularCostoDemora(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre           string
		valorNegocio     *float64
		criticidadTiempo *float64
		reduccionRiesgo  *float64
		esperado         float64
		escenario        string
	}{
		{"todos cargados", punteroFloat64(3), punteroFloat64(1), punteroFloat64(1), 5, "E-02"},
		{"solo valor de negocio", punteroFloat64(8), nil, nil, 8, "CL-04"},
		{"solo criticidad y riesgo", nil, punteroFloat64(3), punteroFloat64(2), 5, "E-04"},
		{"criticidad vacia", punteroFloat64(3), nil, punteroFloat64(1), 4, "CL-04"},
		{"riesgo vacio", punteroFloat64(3), punteroFloat64(1), nil, 4, "CL-04"},
		{"todos vacios", nil, nil, nil, 0, "CL-05"},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			obtenido := CalcularCostoDemora(c.valorNegocio, c.criticidadTiempo, c.reduccionRiesgo)
			assert.InDelta(t, c.esperado, obtenido, 1e-9, "escenario BDD %s", c.escenario)
		})
	}
}

func TestCalcularWSJF(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre      string
		costoDemora float64
		storyPoints *float64
		esperado    *float64
		escenario   string
	}{
		{"cinco sobre tres", 5, punteroFloat64(3), punteroFloat64(5.0 / 3.0), "E-02"},
		{"trece sobre cinco", 13, punteroFloat64(5), punteroFloat64(2.6), "E-01"},
		{"sin story points", 5, nil, nil, "E-03"},
		{"story points cero", 5, punteroFloat64(0), nil, "E-03"},
		{"costo cero", 0, punteroFloat64(5), punteroFloat64(0), "CL-05"},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			assertFloatPtr(t, c.esperado, CalcularWSJF(c.costoDemora, c.storyPoints), "escenario BDD %s", c.escenario)
		})
	}
}

func TestPriorizarBacklog(t *testing.T) {
	t.Parallel()

	historias := []Historia{
		{ID: 1, Estado: EstadoHistoriaBacklog, StoryPoints: punteroFloat64(5), ValorNegocio: punteroFloat64(8), CriticidadTiempo: punteroFloat64(3), ReduccionRiesgo: punteroFloat64(2)},
		{ID: 2, Estado: EstadoHistoriaBacklog, StoryPoints: punteroFloat64(5), ValorNegocio: punteroFloat64(3), CriticidadTiempo: punteroFloat64(1), ReduccionRiesgo: punteroFloat64(1)},
		{ID: 3, Estado: EstadoHistoriaBacklog, StoryPoints: punteroFloat64(5), ValorNegocio: punteroFloat64(1), CriticidadTiempo: punteroFloat64(1), ReduccionRiesgo: punteroFloat64(1)},
		{ID: 4, Estado: EstadoHistoriaBacklog, StoryPoints: punteroFloat64(5), ValorNegocio: punteroFloat64(1), CriticidadTiempo: punteroFloat64(1), ReduccionRiesgo: punteroFloat64(1)},
		{ID: 5, Estado: EstadoHistoriaEnProgreso, ValorNegocio: punteroFloat64(21)},
		{ID: 6, Estado: EstadoHistoriaHecho, StoryPoints: punteroFloat64(5), ValorNegocio: punteroFloat64(21)},
		{ID: 7, Estado: EstadoHistoriaBacklog, StoryPoints: punteroFloat64(0), ValorNegocio: punteroFloat64(21)},
	}

	priorizadas := PriorizarBacklog(historias)

	require.Len(t, priorizadas, 6, "las historias Hecho no se priorizan")
	ids := make([]uint, 0, len(priorizadas))
	for _, p := range priorizadas {
		ids = append(ids, p.Historia.ID)
	}
	assert.Equal(t, []uint{1, 2, 3, 4, 5, 7}, ids,
		"orden WSJF descendente, nulos al final y desempate por ID")

	assertFloatPtr(t, punteroFloat64(1.6), priorizadas[0].ROI, "ROI de la historia 1")
	assertFloatPtr(t, punteroFloat64(2.6), priorizadas[0].WSJF, "WSJF de la historia 1")
	assertFloatPtr(t, punteroFloat64(0.6), priorizadas[1].WSJF, "WSJF de la historia 2")
	assertFloatPtr(t, punteroFloat64(0.6), priorizadas[2].WSJF, "WSJF de la historia 3")
	assert.Nil(t, priorizadas[4].ROI, "la historia sin story points no tiene ROI")
	assert.Nil(t, priorizadas[4].WSJF, "la historia sin story points no tiene WSJF")
	assert.Nil(t, priorizadas[5].ROI, "story_points cero no tiene ROI")
	assert.Nil(t, priorizadas[5].WSJF, "story_points cero no tiene WSJF")

	vacio := PriorizarBacklog(nil)
	assert.NotNil(t, vacio, "una entrada vacia devuelve slice vacio, no nil")
	assert.Empty(t, vacio)
}
