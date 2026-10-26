package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

func nuevoServicioSprintBacklog(t *testing.T) (*SprintBacklogService, *MockProyectoRepository, *MockSprintRepository, *MockHistoriaRepository, *MockSprintBacklogRepository) {
	t.Helper()
	proyectos := nuevoMockProyectoRepository(t)
	sprints := nuevoMockSprintRepository(t)
	historias := nuevoMockHistoriaRepository(t)
	sprintBacklog := nuevoMockSprintBacklogRepository(t)
	return NewSprintBacklogService(proyectos, sprints, historias, sprintBacklog), proyectos, sprints, historias, sprintBacklog
}

func historiaBacklogDePrueba(id uint, proyectoID uint) *domain.Historia {
	sp := 3
	return &domain.Historia{
		ID:          id,
		ProyectoID:  proyectoID,
		Numero:      1,
		Codigo:      domain.FormatearCodigoHistoria(1),
		Titulo:      "Historia de prueba",
		Descripcion: "Descripcion",
		Prioridad:   domain.PrioridadMedia,
		Estado:      domain.EstadoHistoriaBacklog,
		StoryPoints: &sp,
	}
}

func historiaEnProgresoDePrueba(id uint, proyectoID uint) *domain.Historia {
	sp := 5
	return &domain.Historia{
		ID:          id,
		ProyectoID:  proyectoID,
		Numero:      2,
		Codigo:      domain.FormatearCodigoHistoria(2),
		Titulo:      "Historia en progreso",
		Descripcion: "Descripcion",
		Prioridad:   domain.PrioridadAlta,
		Estado:      domain.EstadoHistoriaEnProgreso,
		StoryPoints: &sp,
	}
}

func sprintActivoDePrueba(id uint, proyectoID uint) *domain.Sprint {
	return &domain.Sprint{
		ID:            id,
		ProyectoID:    proyectoID,
		Identificador: "Sprint 1",
		SprintGoal:    "Goal",
		Estado:        domain.EstadoSprintActivo,
		FechaInicio:   time.Now().Add(-1 * time.Hour),
		FechaFin:      time.Now().Add(24 * time.Hour),
	}
}

func sprintPlanificadoDePrueba(id uint, proyectoID uint) *domain.Sprint {
	s := sprintActivoDePrueba(id, proyectoID)
	s.Estado = domain.EstadoSprintPlanificado
	return s
}

func sprintCerradoDePrueba(id uint, proyectoID uint) *domain.Sprint {
	s := sprintActivoDePrueba(id, proyectoID)
	s.Estado = domain.EstadoSprintCerrado
	return s
}

func TestAsignarASprint(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre      string
		proyectoID  uint
		sprintID    uint
		historiaID  uint
		preparar    func(p *MockProyectoRepository, sp *MockSprintRepository, h *MockHistoriaRepository, sb *MockSprintBacklogRepository)
		errEsperado error
		verificar   func(t *testing.T, sb *MockSprintBacklogRepository)
		escenario   string
	}{
		{
			nombre:     "BDD-01 asignar historia pendiente a sprint activo",
			proyectoID: 1,
			sprintID:   10,
			historiaID: 100,
			preparar: func(p *MockProyectoRepository, sp *MockSprintRepository, h *MockHistoriaRepository, sb *MockSprintBacklogRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil)
				sp.On("ObtenerPorID", mock.Anything, uint(1), uint(10)).Return(sprintActivoDePrueba(10, 1), nil)
				h.On("ObtenerPorID", mock.Anything, uint(1), uint(100)).Return(historiaBacklogDePrueba(100, 1), nil)
				sb.On("ObtenerPorHistoria", mock.Anything, uint(100)).Return(nil, nil)
				sb.On("ObtenerMaxOrden", mock.Anything, uint(1), uint(10)).Return(0, nil)
				sb.On("Asignar", mock.Anything, mock.AnythingOfType("*domain.SprintBacklogItem")).Return(nil)
			},
			verificar: func(t *testing.T, sb *MockSprintBacklogRepository) {
				sb.AssertCalled(t, "Asignar", mock.Anything, mock.AnythingOfType("*domain.SprintBacklogItem"))
			},
			escenario: "BDD-01",
		},
		{
			nombre:     "BDD-02 primer elemento del sprint backlog",
			proyectoID: 1,
			sprintID:   10,
			historiaID: 100,
			preparar: func(p *MockProyectoRepository, sp *MockSprintRepository, h *MockHistoriaRepository, sb *MockSprintBacklogRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil)
				sp.On("ObtenerPorID", mock.Anything, uint(1), uint(10)).Return(sprintActivoDePrueba(10, 1), nil)
				h.On("ObtenerPorID", mock.Anything, uint(1), uint(100)).Return(historiaBacklogDePrueba(100, 1), nil)
				sb.On("ObtenerPorHistoria", mock.Anything, uint(100)).Return(nil, nil)
				sb.On("ObtenerMaxOrden", mock.Anything, uint(1), uint(10)).Return(0, nil)
				sb.On("Asignar", mock.Anything, mock.MatchedBy(func(item *domain.SprintBacklogItem) bool {
					return item.Orden == 1
				})).Return(nil)
			},
			verificar: func(t *testing.T, sb *MockSprintBacklogRepository) {
				sb.AssertCalled(t, "Asignar", mock.Anything, mock.MatchedBy(func(item *domain.SprintBacklogItem) bool {
					return item.Orden == 1
				}))
			},
			escenario: "BDD-02",
		},
		{
			nombre:     "BDD-03 historia ya asignada a otro sprint",
			proyectoID: 1,
			sprintID:   10,
			historiaID: 100,
			preparar: func(p *MockProyectoRepository, sp *MockSprintRepository, h *MockHistoriaRepository, sb *MockSprintBacklogRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil)
				sp.On("ObtenerPorID", mock.Anything, uint(1), uint(10)).Return(sprintActivoDePrueba(10, 1), nil)
				h.On("ObtenerPorID", mock.Anything, uint(1), uint(100)).Return(historiaBacklogDePrueba(100, 1), nil)
				sb.On("ObtenerPorHistoria", mock.Anything, uint(100)).Return(&domain.SprintBacklogItem{ID: 1, SprintID: 5, HistoriaID: 100}, nil)
			},
			errEsperado: domain.ErrHistoriaYaAsignada,
			escenario:   "BDD-03",
		},
		{
			nombre:     "BDD-04 historia no pendiente",
			proyectoID: 1,
			sprintID:   10,
			historiaID: 100,
			preparar: func(p *MockProyectoRepository, sp *MockSprintRepository, h *MockHistoriaRepository, sb *MockSprintBacklogRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil)
				sp.On("ObtenerPorID", mock.Anything, uint(1), uint(10)).Return(sprintActivoDePrueba(10, 1), nil)
				h.On("ObtenerPorID", mock.Anything, uint(1), uint(100)).Return(historiaEnProgresoDePrueba(100, 1), nil)
			},
			errEsperado: domain.ErrHistoriaNoPendiente,
			escenario:   "BDD-04",
		},
		{
			nombre:     "BDD-05 sprint no activo",
			proyectoID: 1,
			sprintID:   10,
			historiaID: 100,
			preparar: func(p *MockProyectoRepository, sp *MockSprintRepository, h *MockHistoriaRepository, sb *MockSprintBacklogRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil)
				sp.On("ObtenerPorID", mock.Anything, uint(1), uint(10)).Return(sprintPlanificadoDePrueba(10, 1), nil)
			},
			errEsperado: domain.ErrSprintNoActivo,
			escenario:   "BDD-05",
		},
		{
			nombre:     "BDD-05 sprint cerrado",
			proyectoID: 1,
			sprintID:   10,
			historiaID: 100,
			preparar: func(p *MockProyectoRepository, sp *MockSprintRepository, h *MockHistoriaRepository, sb *MockSprintBacklogRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil)
				sp.On("ObtenerPorID", mock.Anything, uint(1), uint(10)).Return(sprintCerradoDePrueba(10, 1), nil)
			},
			errEsperado: domain.ErrSprintNoActivo,
			escenario:   "BDD-05",
		},
		{
			nombre:     "BDD-09 fuera de ventana temporal - antes de inicio",
			proyectoID: 1,
			sprintID:   10,
			historiaID: 100,
			preparar: func(p *MockProyectoRepository, sp *MockSprintRepository, h *MockHistoriaRepository, sb *MockSprintBacklogRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil)
				s := sprintActivoDePrueba(10, 1)
				s.FechaInicio = time.Now().Add(24 * time.Hour)
				s.FechaFin = time.Now().Add(48 * time.Hour)
				sp.On("ObtenerPorID", mock.Anything, uint(1), uint(10)).Return(s, nil)
				h.On("ObtenerPorID", mock.Anything, uint(1), uint(100)).Return(historiaBacklogDePrueba(100, 1), nil)
				sb.On("ObtenerPorHistoria", mock.Anything, uint(100)).Return(nil, nil)
			},
			errEsperado: domain.ErrFueraDeVentanaTemporal,
			escenario:   "BDD-09",
		},
		{
			nombre:     "BDD-09 fuera de ventana temporal - despues de fin",
			proyectoID: 1,
			sprintID:   10,
			historiaID: 100,
			preparar: func(p *MockProyectoRepository, sp *MockSprintRepository, h *MockHistoriaRepository, sb *MockSprintBacklogRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil)
				s := sprintActivoDePrueba(10, 1)
				s.FechaInicio = time.Now().Add(-48 * time.Hour)
				s.FechaFin = time.Now().Add(-24 * time.Hour)
				sp.On("ObtenerPorID", mock.Anything, uint(1), uint(10)).Return(s, nil)
				h.On("ObtenerPorID", mock.Anything, uint(1), uint(100)).Return(historiaBacklogDePrueba(100, 1), nil)
				sb.On("ObtenerPorHistoria", mock.Anything, uint(100)).Return(nil, nil)
			},
			errEsperado: domain.ErrFueraDeVentanaTemporal,
			escenario:   "BDD-09",
		},
		{
			nombre:     "BDD-08 registra fecha de asignacion",
			proyectoID: 1,
			sprintID:   10,
			historiaID: 100,
			preparar: func(p *MockProyectoRepository, sp *MockSprintRepository, h *MockHistoriaRepository, sb *MockSprintBacklogRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil)
				s := sprintActivoDePrueba(10, 1)
				s.FechaInicio = time.Now().Add(-1 * time.Hour)
				s.FechaFin = time.Now().Add(24 * time.Hour)
				sp.On("ObtenerPorID", mock.Anything, uint(1), uint(10)).Return(s, nil)
				h.On("ObtenerPorID", mock.Anything, uint(1), uint(100)).Return(historiaBacklogDePrueba(100, 1), nil)
				sb.On("ObtenerPorHistoria", mock.Anything, uint(100)).Return(nil, nil)
				sb.On("ObtenerMaxOrden", mock.Anything, uint(1), uint(10)).Return(2, nil)
				sb.On("Asignar", mock.Anything, mock.MatchedBy(func(item *domain.SprintBacklogItem) bool {
					return !item.FechaAsignacion.IsZero() && item.Orden == 3
				})).Return(nil)
			},
			verificar: func(t *testing.T, sb *MockSprintBacklogRepository) {
				sb.AssertCalled(t, "Asignar", mock.Anything, mock.MatchedBy(func(item *domain.SprintBacklogItem) bool {
					return !item.FechaAsignacion.IsZero()
				}))
			},
			escenario: "BDD-08",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			servicio, p, sp, h, sb := nuevoServicioSprintBacklog(t)
			if c.preparar != nil {
				c.preparar(p, sp, h, sb)
			}

			err := servicio.AsignarASprint(context.Background(), c.proyectoID, c.sprintID, c.historiaID)

			if c.errEsperado != nil {
				require.Error(t, err, "escenario BDD %s", c.escenario)
				assert.ErrorIs(t, err, c.errEsperado, "escenario BDD %s", c.escenario)
			} else {
				require.NoError(t, err, "escenario BDD %s", c.escenario)
				if c.verificar != nil {
					c.verificar(t, sb)
				}
			}
			p.AssertExpectations(t)
			sp.AssertExpectations(t)
			h.AssertExpectations(t)
			sb.AssertExpectations(t)
		})
	}
}

func TestDesasignarDeSprint(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre      string
		proyectoID  uint
		sprintID    uint
		historiaID  uint
		preparar    func(p *MockProyectoRepository, sp *MockSprintRepository, sb *MockSprintBacklogRepository)
		errEsperado error
		escenario   string
	}{
		{
			nombre:     "BDD-06 desasignar vuelve a product backlog",
			proyectoID: 1,
			sprintID:   10,
			historiaID: 100,
			preparar: func(p *MockProyectoRepository, sp *MockSprintRepository, sb *MockSprintBacklogRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil)
				sp.On("ObtenerPorID", mock.Anything, uint(1), uint(10)).Return(sprintActivoDePrueba(10, 1), nil)
				sb.On("ObtenerPorHistoria", mock.Anything, uint(100)).Return(&domain.SprintBacklogItem{ID: 1, ProyectoID: 1, SprintID: 10, HistoriaID: 100}, nil)
				sb.On("Desasignar", mock.Anything, uint(1), uint(10), uint(100)).Return(nil)
			},
			escenario: "BDD-06",
		},
		{
			nombre:     "BDD-07 sprint cerrado no permite desasignar",
			proyectoID: 1,
			sprintID:   10,
			historiaID: 100,
			preparar: func(p *MockProyectoRepository, sp *MockSprintRepository, sb *MockSprintBacklogRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil)
				sp.On("ObtenerPorID", mock.Anything, uint(1), uint(10)).Return(sprintCerradoDePrueba(10, 1), nil)
			},
			errEsperado: domain.ErrSprintCerrado,
			escenario:   "BDD-07",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			servicio, p, sp, _, sb := nuevoServicioSprintBacklog(t)
			if c.preparar != nil {
				c.preparar(p, sp, sb)
			}

			err := servicio.DesasignarDeSprint(context.Background(), c.proyectoID, c.sprintID, c.historiaID)

			if c.errEsperado != nil {
				require.Error(t, err, "escenario BDD %s", c.escenario)
				assert.ErrorIs(t, err, c.errEsperado, "escenario BDD %s", c.escenario)
			} else {
				require.NoError(t, err, "escenario BDD %s", c.escenario)
			}
			p.AssertExpectations(t)
			sp.AssertExpectations(t)
			sb.AssertExpectations(t)
		})
	}
}
