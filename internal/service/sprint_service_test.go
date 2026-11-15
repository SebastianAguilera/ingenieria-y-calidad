package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

func nuevoServicioSprints(t *testing.T) (*SprintService, *MockProyectoRepository, *MockSprintRepository) {
	t.Helper()
	proyectos := nuevoMockProyectoRepository(t)
	sprints := nuevoMockSprintRepository(t)
	return NewSprintService(proyectos, sprints), proyectos, sprints
}

func fechaSprint(mes time.Month, dia int) time.Time {
	return time.Date(2026, mes, dia, 0, 0, 0, 0, time.UTC)
}

func altoSprintValido() domain.NuevoSprint {
	return domain.NuevoSprint{
		Identificador: "Sprint 1",
		SprintGoal:    "Entregar el MVP del motor de métricas",
		FechaInicio:   fechaSprint(time.October, 6),
		FechaFin:      fechaSprint(time.October, 20),
	}
}

func prepararAltaValida(p *MockProyectoRepository, s *MockSprintRepository) {
	p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil)
	s.On("HaySolapamiento", mock.Anything, uint(1), mock.Anything, mock.Anything).Return(false, nil)
	s.On("Crear", mock.Anything, mock.Anything).Return(nil)
}

func TestCrearSprint(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre      string
		nuevo       domain.NuevoSprint
		preparar    func(p *MockProyectoRepository, s *MockSprintRepository)
		errEsperado error
		verificar   func(t *testing.T, sprint *domain.Sprint)
		escenario   string
	}{
		{
			nombre: "E-01 alta con todos los campos",
			nuevo:  altoSprintValido(),
			preparar: func(p *MockProyectoRepository, s *MockSprintRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil)
				s.On("HaySolapamiento", mock.Anything, uint(1), mock.Anything, mock.Anything).Return(false, nil)
				s.On("Crear", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
					args.Get(1).(*domain.Sprint).ID = 1
				})
			},
			verificar: func(t *testing.T, sprint *domain.Sprint) {
				require.NotNil(t, sprint)
				assert.Equal(t, uint(1), sprint.ID)
				assert.Equal(t, uint(1), sprint.ProyectoID)
				assert.Equal(t, "Sprint 1", sprint.Identificador)
				assert.Equal(t, "Entregar el MVP del motor de métricas", sprint.SprintGoal)
				assert.Equal(t, domain.EstadoSprintPlanificado, sprint.Estado, "escenario BDD E-01")
				assert.Equal(t, fechaSprint(time.October, 6), sprint.FechaInicio)
				assert.Equal(t, fechaSprint(time.October, 20), sprint.FechaFin)
			},
			escenario: "E-01",
		},
		{
			nombre: "E-02 identificador y goal con espacios se normalizan",
			nuevo: domain.NuevoSprint{
				Identificador: "   Sprint 1   ",
				SprintGoal:    "  Entregar el MVP  ",
				FechaInicio:   fechaSprint(time.October, 6),
				FechaFin:      fechaSprint(time.October, 20),
			},
			preparar: func(p *MockProyectoRepository, s *MockSprintRepository) {
				prepararAltaValida(p, s)
			},
			verificar: func(t *testing.T, sprint *domain.Sprint) {
				assert.Equal(t, "Sprint 1", sprint.Identificador)
				assert.Equal(t, "Entregar el MVP", sprint.SprintGoal)
			},
			escenario: "E-02",
		},
		{
			nombre:      "E-03 identificador vacio",
			nuevo:       domain.NuevoSprint{Identificador: "", SprintGoal: "Goal", FechaInicio: fechaSprint(time.October, 6), FechaFin: fechaSprint(time.October, 20)},
			errEsperado: domain.ErrIdentificadorSprintObligatorio,
			escenario:   "E-03",
		},
		{
			nombre:      "E-03 identificador solo espacios",
			nuevo:       domain.NuevoSprint{Identificador: "    ", SprintGoal: "Goal", FechaInicio: fechaSprint(time.October, 6), FechaFin: fechaSprint(time.October, 20)},
			errEsperado: domain.ErrIdentificadorSprintObligatorio,
			escenario:   "E-03",
		},
		{
			nombre:      "E-04 identificador solo tabuladores",
			nuevo:       domain.NuevoSprint{Identificador: "\t\t", SprintGoal: "Goal", FechaInicio: fechaSprint(time.October, 6), FechaFin: fechaSprint(time.October, 20)},
			errEsperado: domain.ErrIdentificadorSprintObligatorio,
			escenario:   "E-04",
		},
		{
			nombre:      "E-05 sprint goal vacio",
			nuevo:       domain.NuevoSprint{Identificador: "Sprint 1", SprintGoal: "", FechaInicio: fechaSprint(time.October, 6), FechaFin: fechaSprint(time.October, 20)},
			errEsperado: domain.ErrSprintGoalObligatorio,
			escenario:   "E-05",
		},
		{
			nombre:      "E-05 sprint goal solo espacios",
			nuevo:       domain.NuevoSprint{Identificador: "Sprint 1", SprintGoal: "   ", FechaInicio: fechaSprint(time.October, 6), FechaFin: fechaSprint(time.October, 20)},
			errEsperado: domain.ErrSprintGoalObligatorio,
			escenario:   "E-05",
		},
		{
			nombre: "E-06 identificador de exactamente 50 caracteres",
			nuevo: domain.NuevoSprint{
				Identificador: strings.Repeat("a", 50),
				SprintGoal:    "Goal",
				FechaInicio:   fechaSprint(time.October, 6),
				FechaFin:      fechaSprint(time.October, 20),
			},
			preparar: func(p *MockProyectoRepository, s *MockSprintRepository) {
				prepararAltaValida(p, s)
			},
			verificar: func(t *testing.T, sprint *domain.Sprint) {
				assert.Len(t, []rune(sprint.Identificador), 50)
			},
			escenario: "E-06",
		},
		{
			nombre: "E-07 identificador de 51 caracteres",
			nuevo: domain.NuevoSprint{
				Identificador: strings.Repeat("a", 51),
				SprintGoal:    "Goal",
				FechaInicio:   fechaSprint(time.October, 6),
				FechaFin:      fechaSprint(time.October, 20),
			},
			errEsperado: domain.ErrIdentificadorSprintDemasiadoLargo,
			escenario:   "E-07",
		},
		{
			nombre: "E-08 sprint goal de exactamente 300 caracteres",
			nuevo: domain.NuevoSprint{
				Identificador: "Sprint 1",
				SprintGoal:    strings.Repeat("g", 300),
				FechaInicio:   fechaSprint(time.October, 6),
				FechaFin:      fechaSprint(time.October, 20),
			},
			preparar: func(p *MockProyectoRepository, s *MockSprintRepository) {
				prepararAltaValida(p, s)
			},
			verificar: func(t *testing.T, sprint *domain.Sprint) {
				assert.Len(t, []rune(sprint.SprintGoal), 300)
			},
			escenario: "E-08",
		},
		{
			nombre: "E-09 sprint goal de 301 caracteres",
			nuevo: domain.NuevoSprint{
				Identificador: "Sprint 1",
				SprintGoal:    strings.Repeat("g", 301),
				FechaInicio:   fechaSprint(time.October, 6),
				FechaFin:      fechaSprint(time.October, 20),
			},
			errEsperado: domain.ErrSprintGoalDemasiadoLargo,
			escenario:   "E-09",
		},
		{
			nombre: "E-10 fecha de fin anterior a la de inicio",
			nuevo: domain.NuevoSprint{
				Identificador: "Sprint 1",
				SprintGoal:    "Goal",
				FechaInicio:   fechaSprint(time.October, 6),
				FechaFin:      fechaSprint(time.October, 5),
			},
			errEsperado: domain.ErrSprintFechaFinAnteriorAlInicio,
			escenario:   "E-10",
		},
		{
			nombre: "E-11 sprint de un solo dia",
			nuevo: domain.NuevoSprint{
				Identificador: "Sprint 1",
				SprintGoal:    "Daily de ajuste",
				FechaInicio:   fechaSprint(time.October, 6),
				FechaFin:      fechaSprint(time.October, 6),
			},
			preparar: func(p *MockProyectoRepository, s *MockSprintRepository) {
				prepararAltaValida(p, s)
			},
			verificar: func(t *testing.T, sprint *domain.Sprint) {
				assert.Equal(t, sprint.FechaInicio, sprint.FechaFin)
			},
			escenario: "E-11",
		},
		{
			nombre: "E-12 identificador duplicado en el proyecto",
			nuevo:  altoSprintValido(),
			preparar: func(p *MockProyectoRepository, s *MockSprintRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1}, nil)
				s.On("HaySolapamiento", mock.Anything, uint(1), mock.Anything, mock.Anything).Return(false, nil)
				s.On("Crear", mock.Anything, mock.Anything).Return(domain.ErrIdentificadorSprintDuplicado)
			},
			errEsperado: domain.ErrIdentificadorSprintDuplicado,
			escenario:   "E-12",
		},
		{
			nombre: "E-14 solapamiento parcial de fechas",
			nuevo: domain.NuevoSprint{
				Identificador: "Sprint 2",
				SprintGoal:    "Continuar el MVP",
				FechaInicio:   fechaSprint(time.October, 19),
				FechaFin:      fechaSprint(time.November, 3),
			},
			preparar: func(p *MockProyectoRepository, s *MockSprintRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1}, nil)
				s.On("HaySolapamiento", mock.Anything, uint(1), mock.Anything, mock.Anything).Return(true, nil)
			},
			errEsperado: domain.ErrSprintFechasSolapadas,
			escenario:   "E-14",
		},
		{
			nombre: "E-16 sprints contiguos no se solapan",
			nuevo: domain.NuevoSprint{
				Identificador: "Sprint 2",
				SprintGoal:    "Continuar el MVP",
				FechaInicio:   fechaSprint(time.October, 20),
				FechaFin:      fechaSprint(time.November, 3),
			},
			preparar: func(p *MockProyectoRepository, s *MockSprintRepository) {
				prepararAltaValida(p, s)
			},
			verificar: func(t *testing.T, sprint *domain.Sprint) {
				assert.Equal(t, fechaSprint(time.October, 20), sprint.FechaInicio)
			},
			escenario: "E-16",
		},
		{
			nombre: "E-17 proyecto inexistente",
			nuevo:  altoSprintValido(),
			preparar: func(p *MockProyectoRepository, s *MockSprintRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(nil, domain.ErrProyectoNoEncontrado)
			},
			errEsperado: domain.ErrProyectoNoEncontrado,
			escenario:   "E-17",
		},
		{
			nombre: "E-18 fecha de inicio ausente",
			nuevo: domain.NuevoSprint{
				Identificador: "Sprint 1",
				SprintGoal:    "Goal",
				FechaInicio:   time.Time{},
				FechaFin:      fechaSprint(time.October, 20),
			},
			errEsperado: domain.ErrSprintFechaInicioObligatoria,
			escenario:   "E-18",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			servicio, proyectos, sprints := nuevoServicioSprints(t)
			if c.preparar != nil {
				c.preparar(proyectos, sprints)
			}

			sprint, err := servicio.Crear(context.Background(), uint(1), c.nuevo)

			if c.errEsperado != nil {
				require.Error(t, err, "escenario BDD %s", c.escenario)
				assert.ErrorIs(t, err, c.errEsperado, "escenario BDD %s", c.escenario)
				assert.Nil(t, sprint)
			} else {
				require.NoError(t, err, "escenario BDD %s", c.escenario)
				require.NotNil(t, sprint)
				if c.verificar != nil {
					c.verificar(t, sprint)
				}
			}
			proyectos.AssertExpectations(t)
			sprints.AssertExpectations(t)
		})
	}
}

func TestCrearSprintPropagaErroresDeInfraestructura(t *testing.T) {
	t.Parallel()

	t.Run("falla la verificacion de solapamiento", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, sprints := nuevoServicioSprints(t)
		fallo := errors.New("sin conexion")
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1}, nil)
		sprints.On("HaySolapamiento", mock.Anything, uint(1), mock.Anything, mock.Anything).Return(false, fallo)

		_, err := servicio.Crear(context.Background(), 1, altoSprintValido())

		require.ErrorIs(t, err, fallo)
	})

	t.Run("falla el alta en el repositorio", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, sprints := nuevoServicioSprints(t)
		fallo := errors.New("sin conexion")
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1}, nil)
		sprints.On("HaySolapamiento", mock.Anything, uint(1), mock.Anything, mock.Anything).Return(false, nil)
		sprints.On("Crear", mock.Anything, mock.Anything).Return(fallo)

		_, err := servicio.Crear(context.Background(), 1, altoSprintValido())

		require.ErrorIs(t, err, fallo)
	})
}

func TestListarSprints(t *testing.T) {
	t.Parallel()

	t.Run("E-25 listado ordenado por fecha de inicio", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, sprints := nuevoServicioSprints(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1}, nil)
		sprints.On("ListarPorProyecto", mock.Anything, uint(1)).Return([]domain.Sprint{
			*sprintDePrueba(1, "Sprint 1", domain.EstadoSprintPlanificado),
			*sprintDePrueba(2, "Sprint 2", domain.EstadoSprintPlanificado),
		}, nil)

		lista, err := servicio.Listar(context.Background(), 1)

		require.NoError(t, err)
		require.Len(t, lista, 2)
		assert.Equal(t, "Sprint 1", lista[0].Identificador)
		assert.Equal(t, "Sprint 2", lista[1].Identificador)
	})

	t.Run("E-26 listado vacio no es nil", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, sprints := nuevoServicioSprints(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1}, nil)
		sprints.On("ListarPorProyecto", mock.Anything, uint(1)).Return([]domain.Sprint{}, nil)

		lista, err := servicio.Listar(context.Background(), 1)

		require.NoError(t, err)
		assert.NotNil(t, lista)
		assert.Empty(t, lista)
	})

	t.Run("E-27 proyecto inexistente", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, _ := nuevoServicioSprints(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(99)).Return(nil, domain.ErrProyectoNoEncontrado)

		lista, err := servicio.Listar(context.Background(), 99)

		require.ErrorIs(t, err, domain.ErrProyectoNoEncontrado)
		assert.Nil(t, lista)
	})
}

func TestCambiarEstadoSprint(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre      string
		proyectoID  uint
		sprintID    uint
		destino     domain.EstadoSprint
		preparar    func(p *MockProyectoRepository, s *MockSprintRepository)
		errEsperado error
		verificar   func(t *testing.T, sprint *domain.Sprint)
		escenario   string
	}{
		{
			nombre:     "E-29 activar un sprint Planificado",
			proyectoID: 1,
			sprintID:   1,
			destino:    domain.EstadoSprintActivo,
			preparar: func(p *MockProyectoRepository, s *MockSprintRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1}, nil)
				s.On("ObtenerPorID", mock.Anything, uint(1), uint(1)).
					Return(sprintDePrueba(1, "Sprint 1", domain.EstadoSprintPlanificado), nil)
				s.On("ObtenerActivo", mock.Anything, uint(1)).Return(nil, nil)
				s.On("ActualizarEstado", mock.Anything, mock.Anything).Return(nil)
			},
			verificar: func(t *testing.T, sprint *domain.Sprint) {
				assert.Equal(t, domain.EstadoSprintActivo, sprint.Estado)
			},
			escenario: "E-29",
		},
		{
			nombre:     "E-30 cerrar un sprint Activo",
			proyectoID: 1,
			sprintID:   1,
			destino:    domain.EstadoSprintCerrado,
			preparar: func(p *MockProyectoRepository, s *MockSprintRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1}, nil)
				s.On("ObtenerPorID", mock.Anything, uint(1), uint(1)).
					Return(sprintDePrueba(1, "Sprint 1", domain.EstadoSprintActivo), nil)
				s.On("ActualizarEstado", mock.Anything, mock.Anything).Return(nil)
			},
			verificar: func(t *testing.T, sprint *domain.Sprint) {
				assert.Equal(t, domain.EstadoSprintCerrado, sprint.Estado)
			},
			escenario: "E-30",
		},
		{
			nombre:     "E-31 Activo a Planificado es invalido",
			proyectoID: 1,
			sprintID:   1,
			destino:    domain.EstadoSprintPlanificado,
			preparar: func(p *MockProyectoRepository, s *MockSprintRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1}, nil)
				s.On("ObtenerPorID", mock.Anything, uint(1), uint(1)).
					Return(sprintDePrueba(1, "Sprint 1", domain.EstadoSprintActivo), nil)
			},
			errEsperado: domain.ErrTransicionSprintInvalida,
			escenario:   "E-31",
		},
		{
			nombre:     "E-31 Planificado a Cerrado es invalido",
			proyectoID: 1,
			sprintID:   1,
			destino:    domain.EstadoSprintCerrado,
			preparar: func(p *MockProyectoRepository, s *MockSprintRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1}, nil)
				s.On("ObtenerPorID", mock.Anything, uint(1), uint(1)).
					Return(sprintDePrueba(1, "Sprint 1", domain.EstadoSprintPlanificado), nil)
			},
			errEsperado: domain.ErrTransicionSprintInvalida,
			escenario:   "E-31",
		},
		{
			nombre:     "E-31 Cerrado a Activo es invalido",
			proyectoID: 1,
			sprintID:   1,
			destino:    domain.EstadoSprintActivo,
			preparar: func(p *MockProyectoRepository, s *MockSprintRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1}, nil)
				s.On("ObtenerPorID", mock.Anything, uint(1), uint(1)).
					Return(sprintDePrueba(1, "Sprint 1", domain.EstadoSprintCerrado), nil)
			},
			errEsperado: domain.ErrTransicionSprintInvalida,
			escenario:   "E-31",
		},
		{
			nombre:     "E-31 Activo a Activo es invalido",
			proyectoID: 1,
			sprintID:   1,
			destino:    domain.EstadoSprintActivo,
			preparar: func(p *MockProyectoRepository, s *MockSprintRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1}, nil)
				s.On("ObtenerPorID", mock.Anything, uint(1), uint(1)).
					Return(sprintDePrueba(1, "Sprint 1", domain.EstadoSprintActivo), nil)
			},
			errEsperado: domain.ErrTransicionSprintInvalida,
			escenario:   "E-31",
		},
		{
			nombre:      "E-32 literal de estado fuera del conjunto",
			proyectoID:  1,
			sprintID:    1,
			destino:     domain.EstadoSprint("Terminado"),
			errEsperado: domain.ErrEstadoSprintInvalido,
			escenario:   "E-32",
		},
		{
			nombre:     "E-34 sprint inexistente",
			proyectoID: 1,
			sprintID:   5,
			destino:    domain.EstadoSprintActivo,
			preparar: func(p *MockProyectoRepository, s *MockSprintRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1}, nil)
				s.On("ObtenerPorID", mock.Anything, uint(1), uint(5)).Return(nil, domain.ErrSprintNoEncontrado)
			},
			errEsperado: domain.ErrSprintNoEncontrado,
			escenario:   "E-34",
		},
		{
			nombre:     "E-36 proyecto inexistente",
			proyectoID: 99,
			sprintID:   1,
			destino:    domain.EstadoSprintActivo,
			preparar: func(p *MockProyectoRepository, s *MockSprintRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(99)).Return(nil, domain.ErrProyectoNoEncontrado)
			},
			errEsperado: domain.ErrProyectoNoEncontrado,
			escenario:   "E-36",
		},
		{
			nombre:     "E-38 ya existe otro sprint activo",
			proyectoID: 1,
			sprintID:   2,
			destino:    domain.EstadoSprintActivo,
			preparar: func(p *MockProyectoRepository, s *MockSprintRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1}, nil)
				s.On("ObtenerPorID", mock.Anything, uint(1), uint(2)).
					Return(sprintDePrueba(2, "Sprint 2", domain.EstadoSprintPlanificado), nil)
				s.On("ObtenerActivo", mock.Anything, uint(1)).
					Return(sprintDePrueba(1, "Sprint 1", domain.EstadoSprintActivo), nil)
			},
			errEsperado: domain.ErrSprintActivoExistente,
			escenario:   "E-38",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			servicio, proyectos, sprints := nuevoServicioSprints(t)
			if c.preparar != nil {
				c.preparar(proyectos, sprints)
			}

			sprint, err := servicio.CambiarEstado(context.Background(), c.proyectoID, c.sprintID, c.destino)

			if c.errEsperado != nil {
				require.Error(t, err, "escenario BDD %s", c.escenario)
				assert.ErrorIs(t, err, c.errEsperado, "escenario BDD %s", c.escenario)
				assert.Nil(t, sprint)
			} else {
				require.NoError(t, err, "escenario BDD %s", c.escenario)
				if c.verificar != nil {
					c.verificar(t, sprint)
				}
			}
			proyectos.AssertExpectations(t)
			sprints.AssertExpectations(t)
		})
	}
}

func TestE39CerrarElActivoYActivarElSiguiente(t *testing.T) {
	t.Parallel()

	servicio, proyectos, sprints := nuevoServicioSprints(t)
	proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1}, nil)
	sprints.On("ObtenerPorID", mock.Anything, uint(1), uint(1)).
		Return(sprintDePrueba(1, "Sprint 1", domain.EstadoSprintActivo), nil)
	sprints.On("ActualizarEstado", mock.Anything, mock.Anything).Return(nil)
	sprints.On("ObtenerPorID", mock.Anything, uint(1), uint(2)).
		Return(sprintDePrueba(2, "Sprint 2", domain.EstadoSprintPlanificado), nil)
	sprints.On("ObtenerActivo", mock.Anything, uint(1)).Return(nil, nil)

	cerrado, err := servicio.CambiarEstado(context.Background(), 1, 1, domain.EstadoSprintCerrado)
	require.NoError(t, err, "escenario BDD E-39")
	assert.Equal(t, domain.EstadoSprintCerrado, cerrado.Estado)

	activo, err := servicio.CambiarEstado(context.Background(), 1, 2, domain.EstadoSprintActivo)
	require.NoError(t, err, "escenario BDD E-39")
	assert.Equal(t, domain.EstadoSprintActivo, activo.Estado)
}
