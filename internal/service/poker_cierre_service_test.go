package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

// Tests de US-17: acuerdo, cierre y cancelación de Planning Poker. Los
// escenarios E-XX refieren a
// docs/features/US-17-acuerdo-cierre-cancelacion-planning-poker.feature.

const (
	sesionCierreID = uint(9)
	rondaCierreID  = uint(14)
)

var inicioProyectoCierre = time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC)

func sesionCierreEn(estado domain.EstadoSesionPoker) *domain.SesionPoker {
	return &domain.SesionPoker{
		ID: sesionCierreID, ProyectoID: proyectoPokerID, HistoriaID: historiaPokerID, CreadoPor: anaPokerID,
		CreadorNombre: "Ana", RondaActual: 1, Estado: estado,
	}
}

func historiaConTresPuntos() *domain.Historia {
	h := historiaDePoker(domain.EstadoHistoriaBacklog)
	tres := 3.0
	h.StoryPoints = &tres
	return h
}

// prepararCierre deja el proyecto existente y la primera lectura de la sesión.
func prepararCierre(t *testing.T, estadoInicial domain.EstadoSesionPoker) *dependenciasPoker {
	t.Helper()
	d := nuevasDependenciasPoker(t)
	d.proyectos.On("ObtenerPorID", mock.Anything, proyectoPokerID).
		Return(proyectoDePrueba(proyectoPokerID, "Metrics", inicioProyectoCierre, nil, domain.EstadoActivo), nil)
	d.poker.On("ObtenerPorID", mock.Anything, proyectoPokerID, sesionCierreID).
		Return(sesionCierreEn(estadoInicial), nil).Once()
	return d
}

// esperarDetalleFinal configura la relectura de la sesión tras el commit y las
// consultas que arman el detalle.
func esperarDetalleFinal(d *dependenciasPoker, final *domain.SesionPoker, rondaRevelada bool) {
	ronda := &domain.RondaPoker{ID: rondaCierreID, SesionID: sesionCierreID, NumeroRonda: 1}
	if rondaRevelada {
		revelada := time.Date(2026, 11, 14, 18, 0, 0, 0, time.UTC)
		ronda.ReveladaEn = &revelada
	}
	d.poker.On("ObtenerPorID", mock.Anything, proyectoPokerID, sesionCierreID).Return(final, nil).Once()
	d.poker.On("ObtenerRonda", mock.Anything, sesionCierreID, 1).Return(ronda, nil)
	d.poker.On("ContarMiembros", mock.Anything, proyectoPokerID).Return(3, nil)
	d.poker.On("ContarVotosPorRonda", mock.Anything, rondaCierreID).Return(2, nil)
	d.poker.On("ObtenerVotoUsuario", mock.Anything, rondaCierreID, anaPokerID).Return(nil, nil)
	d.poker.On("ListarVotosPorRonda", mock.Anything, rondaCierreID).Return([]domain.VotoPoker{
		{UsuarioID: anaPokerID, Nombre: "Ana", Puntos: 3},
		{UsuarioID: benPokerID, Nombre: "Ben", Puntos: 5},
	}, nil).Maybe()
	d.poker.On("ListarRondasPorSesion", mock.Anything, sesionCierreID).Return([]domain.RondaPoker{*ronda}, nil)
}

func sesionCerradaCon(puntos float64) *domain.SesionPoker {
	s := sesionCierreEn(domain.EstadoSesionCerrada)
	cerrada := time.Date(2026, 11, 14, 18, 30, 0, 0, time.UTC)
	s.PuntosAcordados = &puntos
	s.CerradoEn = &cerrada
	return s
}

func TestAcordarSesionPoker(t *testing.T) {
	ctx := context.Background()

	acuerdoExitoso := func(t *testing.T, puntos float64) (*dependenciasPoker, *domain.DetalleSesionPoker) {
		t.Helper()
		d := prepararCierre(t, domain.EstadoSesionRevelada)
		tx := d.uow.repos
		tx.poker.On("CerrarSesion", mock.Anything, sesionCierreID, puntos).
			Return(&domain.RondaPoker{ID: rondaCierreID, SesionID: sesionCierreID, NumeroRonda: 1}, nil)
		tx.historias.On("ObtenerPorID", mock.Anything, proyectoPokerID, historiaPokerID).
			Return(historiaConTresPuntos(), nil)
		tx.historias.On("Actualizar", mock.Anything, mock.MatchedBy(func(h *domain.Historia) bool {
			return h.ID == historiaPokerID && h.StoryPoints != nil && *h.StoryPoints == puntos
		})).Return(nil)
		tx.estimaciones.On("Crear", mock.Anything, mock.MatchedBy(func(e *domain.StoryEstimate) bool {
			return e.HistoriaID == historiaPokerID &&
				e.SesionID != nil && *e.SesionID == sesionCierreID &&
				e.RondaID != nil && *e.RondaID == rondaCierreID &&
				e.EstimadoPor != nil && *e.EstimadoPor == anaPokerID &&
				e.Puntos == puntos &&
				e.Fuente == domain.FuenteEstimacionPlanningPoker
		})).Return(nil)

		final := sesionCerradaCon(puntos)
		esperarDetalleFinal(d, final, true)
		sesionID := sesionCierreID
		d.estimaciones.On("ListarPorSesiones", mock.Anything, []uint{sesionCierreID}).Return([]domain.StoryEstimate{
			{ID: 11, HistoriaID: historiaPokerID, SesionID: &sesionID, Puntos: puntos, Fuente: domain.FuenteEstimacionPlanningPoker},
		}, nil)

		detalle, err := d.servicio.Acordar(ctx, proyectoPokerID, sesionCierreID, anaPokerID, puntos)
		require.NoError(t, err)
		return d, detalle
	}

	t.Run("E-01 acordar cierra la sesión, estima la historia y crea el story_estimate en una transacción", func(t *testing.T) {
		d, detalle := acuerdoExitoso(t, 5)

		assert.Equal(t, 1, d.uow.ejecuciones, "CA-6 todo ocurre en una única transacción")
		assert.True(t, d.uow.confirmada)
		assert.Equal(t, domain.EstadoSesionCerrada, detalle.Sesion.Estado)
		require.NotNil(t, detalle.Sesion.PuntosAcordados)
		assert.Equal(t, 5.0, *detalle.Sesion.PuntosAcordados)
		require.NotNil(t, detalle.Sesion.Estimacion, "RB-13 el detalle incluye el story_estimate")
		assert.Equal(t, uint(11), detalle.Sesion.Estimacion.ID)
		require.Len(t, detalle.Votos, 2, "en Cerrada siguen visibles los votos revelados")
		d.uow.repos.historias.AssertExpectations(t)
		d.uow.repos.estimaciones.AssertExpectations(t)
		d.historias.AssertNotCalled(t, "Actualizar", mock.Anything, mock.Anything)
		d.estimaciones.AssertNotCalled(t, "Crear", mock.Anything, mock.Anything)
	})

	t.Run("E-02 acepta 0 y 0.5", func(t *testing.T) {
		for _, puntos := range []float64{0, 0.5} {
			t.Run(fmt.Sprintf("%v", puntos), func(t *testing.T) {
				_, detalle := acuerdoExitoso(t, puntos)
				assert.Equal(t, puntos, *detalle.Sesion.PuntosAcordados)
			})
		}
	})

	t.Run("E-06 puntos fuera de la escala de US-05 no abren la transacción", func(t *testing.T) {
		for _, puntos := range []float64{4, -1, 34, 1.5} {
			t.Run(fmt.Sprintf("%v", puntos), func(t *testing.T) {
				d := prepararCierre(t, domain.EstadoSesionRevelada)

				_, err := d.servicio.Acordar(ctx, proyectoPokerID, sesionCierreID, anaPokerID, puntos)

				assert.ErrorIs(t, err, domain.ErrPuntosPokerInvalidos)
				assert.Zero(t, d.uow.ejecuciones, "CA-4 no se toca la historia")
			})
		}
	})

	t.Run("E-04 sesión en Votando devuelve SESION_POKER_NO_REVELADA sin escribir", func(t *testing.T) {
		d := prepararCierre(t, domain.EstadoSesionVotando)
		d.uow.repos.poker.On("CerrarSesion", mock.Anything, sesionCierreID, 5.0).
			Return(nil, domain.ErrSesionPokerNoRevelada)

		_, err := d.servicio.Acordar(ctx, proyectoPokerID, sesionCierreID, anaPokerID, 5)

		assert.ErrorIs(t, err, domain.ErrSesionPokerNoRevelada)
		assert.True(t, d.uow.revertida)
		d.uow.repos.historias.AssertNotCalled(t, "Actualizar", mock.Anything, mock.Anything)
		d.uow.repos.estimaciones.AssertNotCalled(t, "Crear", mock.Anything, mock.Anything)
	})

	t.Run("E-05 sesión Cerrada o Cancelada devuelve SESION_POKER_CERRADA sin escribir", func(t *testing.T) {
		for _, estado := range []domain.EstadoSesionPoker{domain.EstadoSesionCerrada, domain.EstadoSesionCancelada} {
			t.Run(string(estado), func(t *testing.T) {
				d := prepararCierre(t, estado)
				d.uow.repos.poker.On("CerrarSesion", mock.Anything, sesionCierreID, 5.0).
					Return(nil, domain.ErrSesionPokerCerrada)

				_, err := d.servicio.Acordar(ctx, proyectoPokerID, sesionCierreID, anaPokerID, 5)

				assert.ErrorIs(t, err, domain.ErrSesionPokerCerrada)
				d.uow.repos.historias.AssertNotCalled(t, "Actualizar", mock.Anything, mock.Anything)
				d.uow.repos.estimaciones.AssertNotCalled(t, "Crear", mock.Anything, mock.Anything)
			})
		}
	})

	t.Run("E-08 una falla al crear el story_estimate revierte la transacción", func(t *testing.T) {
		d := prepararCierre(t, domain.EstadoSesionRevelada)
		tx := d.uow.repos
		falla := errors.New("insert story_estimates falló")
		tx.poker.On("CerrarSesion", mock.Anything, sesionCierreID, 5.0).
			Return(&domain.RondaPoker{ID: rondaCierreID, SesionID: sesionCierreID, NumeroRonda: 1}, nil)
		tx.historias.On("ObtenerPorID", mock.Anything, proyectoPokerID, historiaPokerID).Return(historiaConTresPuntos(), nil)
		tx.historias.On("Actualizar", mock.Anything, mock.Anything).Return(nil)
		tx.estimaciones.On("Crear", mock.Anything, mock.Anything).Return(falla)

		detalle, err := d.servicio.Acordar(ctx, proyectoPokerID, sesionCierreID, anaPokerID, 5)

		assert.Nil(t, detalle)
		assert.ErrorIs(t, err, falla)
		assert.True(t, d.uow.revertida, "CA-5 la falla provoca rollback total")
		assert.False(t, d.uow.confirmada)
	})

	t.Run("E-09 una falla al actualizar la historia revierte sin crear story_estimate", func(t *testing.T) {
		d := prepararCierre(t, domain.EstadoSesionRevelada)
		tx := d.uow.repos
		falla := errors.New("update historias falló")
		tx.poker.On("CerrarSesion", mock.Anything, sesionCierreID, 5.0).
			Return(&domain.RondaPoker{ID: rondaCierreID, SesionID: sesionCierreID, NumeroRonda: 1}, nil)
		tx.historias.On("ObtenerPorID", mock.Anything, proyectoPokerID, historiaPokerID).Return(historiaConTresPuntos(), nil)
		tx.historias.On("Actualizar", mock.Anything, mock.Anything).Return(falla)

		_, err := d.servicio.Acordar(ctx, proyectoPokerID, sesionCierreID, anaPokerID, 5)

		assert.ErrorIs(t, err, falla)
		assert.True(t, d.uow.revertida)
		tx.estimaciones.AssertNotCalled(t, "Crear", mock.Anything, mock.Anything)
	})

	t.Run("E-18 sesión inexistente o de otro proyecto", func(t *testing.T) {
		d := nuevasDependenciasPoker(t)
		d.proyectos.On("ObtenerPorID", mock.Anything, proyectoPokerID).
			Return(proyectoDePrueba(proyectoPokerID, "Metrics", inicioProyectoCierre, nil, domain.EstadoActivo), nil)
		d.poker.On("ObtenerPorID", mock.Anything, proyectoPokerID, sesionCierreID).
			Return(nil, domain.ErrSesionPokerNoEncontrada)

		_, err := d.servicio.Acordar(ctx, proyectoPokerID, sesionCierreID, anaPokerID, 4)

		assert.ErrorIs(t, err, domain.ErrSesionPokerNoEncontrada, "RB-16 el 404 precede al 422")
		assert.Zero(t, d.uow.ejecuciones)
	})

	t.Run("proyecto inexistente", func(t *testing.T) {
		d := nuevasDependenciasPoker(t)
		d.proyectos.On("ObtenerPorID", mock.Anything, proyectoPokerID).Return(nil, domain.ErrProyectoNoEncontrado)

		_, err := d.servicio.Acordar(ctx, proyectoPokerID, sesionCierreID, anaPokerID, 5)

		assert.ErrorIs(t, err, domain.ErrProyectoNoEncontrado)
		assert.Zero(t, d.uow.ejecuciones)
	})
}

func TestCancelarSesionPoker(t *testing.T) {
	ctx := context.Background()

	t.Run("E-11 cancela una sesión activa sin tocar la historia", func(t *testing.T) {
		for _, estado := range []domain.EstadoSesionPoker{domain.EstadoSesionVotando, domain.EstadoSesionRevelada} {
			t.Run(string(estado), func(t *testing.T) {
				d := prepararCierre(t, estado)
				d.poker.On("CancelarSesion", mock.Anything, sesionCierreID).Return(nil)
				final := sesionCierreEn(domain.EstadoSesionCancelada)
				cancelada := time.Date(2026, 11, 14, 19, 0, 0, 0, time.UTC)
				final.CanceladoEn = &cancelada
				esperarDetalleFinal(d, final, estado == domain.EstadoSesionRevelada)

				detalle, err := d.servicio.Cancelar(ctx, proyectoPokerID, sesionCierreID, anaPokerID)

				require.NoError(t, err)
				assert.Equal(t, domain.EstadoSesionCancelada, detalle.Sesion.Estado)
				assert.NotNil(t, detalle.Sesion.CanceladoEn)
				assert.Nil(t, detalle.Sesion.PuntosAcordados)
				assert.Nil(t, detalle.Sesion.Estimacion, "RB-09 la cancelación no crea story_estimate")
				assert.Zero(t, d.uow.ejecuciones)
				d.historias.AssertNotCalled(t, "Actualizar", mock.Anything, mock.Anything)
				d.estimaciones.AssertNotCalled(t, "Crear", mock.Anything, mock.Anything)
			})
		}
	})

	t.Run("cancelar en Votando no expone los votos ocultos de la ronda", func(t *testing.T) {
		d := prepararCierre(t, domain.EstadoSesionVotando)
		d.poker.On("CancelarSesion", mock.Anything, sesionCierreID).Return(nil)
		esperarDetalleFinal(d, sesionCierreEn(domain.EstadoSesionCancelada), false)

		detalle, err := d.servicio.Cancelar(ctx, proyectoPokerID, sesionCierreID, anaPokerID)

		require.NoError(t, err)
		assert.Empty(t, detalle.Votos)
		assert.Nil(t, detalle.Dispersion)
		d.poker.AssertNotCalled(t, "ListarVotosPorRonda", mock.Anything, mock.Anything)
	})

	t.Run("E-14 cancelar en Revelada conserva los votos revelados", func(t *testing.T) {
		d := prepararCierre(t, domain.EstadoSesionRevelada)
		d.poker.On("CancelarSesion", mock.Anything, sesionCierreID).Return(nil)
		esperarDetalleFinal(d, sesionCierreEn(domain.EstadoSesionCancelada), true)

		detalle, err := d.servicio.Cancelar(ctx, proyectoPokerID, sesionCierreID, anaPokerID)

		require.NoError(t, err)
		assert.Len(t, detalle.Votos, 2)
		require.Len(t, detalle.HistorialRondas, 1)
		assert.Len(t, detalle.HistorialRondas[0].Votos, 2)
	})

	t.Run("E-12 sesión Cerrada o Cancelada devuelve SESION_POKER_CERRADA", func(t *testing.T) {
		for _, estado := range []domain.EstadoSesionPoker{domain.EstadoSesionCerrada, domain.EstadoSesionCancelada} {
			t.Run(string(estado), func(t *testing.T) {
				d := prepararCierre(t, estado)
				d.poker.On("CancelarSesion", mock.Anything, sesionCierreID).Return(domain.ErrSesionPokerCerrada)

				_, err := d.servicio.Cancelar(ctx, proyectoPokerID, sesionCierreID, anaPokerID)

				assert.ErrorIs(t, err, domain.ErrSesionPokerCerrada)
			})
		}
	})

	t.Run("E-18 sesión inexistente", func(t *testing.T) {
		d := nuevasDependenciasPoker(t)
		d.proyectos.On("ObtenerPorID", mock.Anything, proyectoPokerID).
			Return(proyectoDePrueba(proyectoPokerID, "Metrics", inicioProyectoCierre, nil, domain.EstadoActivo), nil)
		d.poker.On("ObtenerPorID", mock.Anything, proyectoPokerID, sesionCierreID).
			Return(nil, domain.ErrSesionPokerNoEncontrada)

		_, err := d.servicio.Cancelar(ctx, proyectoPokerID, sesionCierreID, anaPokerID)

		assert.ErrorIs(t, err, domain.ErrSesionPokerNoEncontrada)
		d.poker.AssertNotCalled(t, "CancelarSesion", mock.Anything, mock.Anything)
	})

	t.Run("proyecto inexistente", func(t *testing.T) {
		d := nuevasDependenciasPoker(t)
		d.proyectos.On("ObtenerPorID", mock.Anything, proyectoPokerID).Return(nil, domain.ErrProyectoNoEncontrado)

		_, err := d.servicio.Cancelar(ctx, proyectoPokerID, sesionCierreID, anaPokerID)

		assert.ErrorIs(t, err, domain.ErrProyectoNoEncontrado)
	})
}

func TestDetalleSesionCerradaIncluyeStoryEstimate(t *testing.T) {
	d := prepararCierre(t, domain.EstadoSesionCerrada)
	// prepararCierre ya registró la única lectura que hace Detalle; se
	// reemplaza por la sesión cerrada con puntos.
	d.poker.ExpectedCalls = nil
	final := sesionCerradaCon(8)
	esperarDetalleFinal(d, final, true)
	sesionID := sesionCierreID
	d.estimaciones.On("ListarPorSesiones", mock.Anything, []uint{sesionCierreID}).Return([]domain.StoryEstimate{
		{ID: 12, HistoriaID: historiaPokerID, SesionID: &sesionID, Puntos: 8, Fuente: domain.FuenteEstimacionPlanningPoker},
	}, nil)

	detalle, err := d.servicio.Detalle(context.Background(), proyectoPokerID, sesionCierreID, anaPokerID)

	require.NoError(t, err)
	require.NotNil(t, detalle.Sesion.Estimacion)
	assert.Equal(t, 8.0, detalle.Sesion.Estimacion.Puntos)
}
