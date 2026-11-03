package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

var errInfraestructura = errors.New("fallo de infraestructura")

func TestUsuarioServiceErroresDeInfraestructura(t *testing.T) {
	t.Parallel()

	t.Run("Crear propaga el error del repositorio", func(t *testing.T) {
		t.Parallel()
		usuarios := nuevoMockUsuarioRepository(t)
		servicio := NewUsuarioService(usuarios)
		usuarios.On("Crear", mock.Anything, mock.Anything).Return(errInfraestructura)

		usuario, err := servicio.Crear(context.Background(), domain.UsuarioInput{
			Nombre: "Ana",
			Email:  "ana@example.com",
		})

		require.ErrorIs(t, err, errInfraestructura)
		assert.Nil(t, usuario)
	})

	t.Run("Listar propaga el error del repositorio", func(t *testing.T) {
		t.Parallel()
		usuarios := nuevoMockUsuarioRepository(t)
		servicio := NewUsuarioService(usuarios)
		usuarios.On("Listar", mock.Anything).Return(nil, errInfraestructura)

		lista, err := servicio.Listar(context.Background())

		require.ErrorIs(t, err, errInfraestructura)
		assert.Nil(t, lista)
	})
}

func TestValidarUsuarioLimites(t *testing.T) {
	t.Parallel()

	usuarios := nuevoMockUsuarioRepository(t)
	servicio := NewUsuarioService(usuarios)

	_, err := servicio.Crear(context.Background(), domain.UsuarioInput{
		Nombre: "   ",
		Email:  "ana@example.com",
	})
	assert.ErrorIs(t, err, domain.ErrUsuarioNombreVacio)

	_, err = servicio.Crear(context.Background(), domain.UsuarioInput{
		Nombre: strings.Repeat("a", 101),
		Email:  "ana@example.com",
	})
	assert.ErrorIs(t, err, domain.ErrUsuarioNombreLargo)

	usuarios.AssertNotCalled(t, "Crear", mock.Anything, mock.Anything)
}

func TestListarHistoriasPropagaErrorDelRepositorio(t *testing.T) {
	t.Parallel()

	servicio, proyectos, historias := nuevoServicioHistorias(t)
	proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1}, nil)
	historias.On("ListarPorProyecto", mock.Anything, uint(1)).Return(nil, errInfraestructura)

	lista, err := servicio.Listar(context.Background(), 1)

	require.ErrorIs(t, err, errInfraestructura)
	assert.Nil(t, lista)
}

func TestListarSprintsPropagaErrorYNormalizaNil(t *testing.T) {
	t.Parallel()

	t.Run("error del repositorio", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, sprints := nuevoServicioSprints(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1}, nil)
		sprints.On("ListarPorProyecto", mock.Anything, uint(1)).Return(nil, errInfraestructura)

		lista, err := servicio.Listar(context.Background(), 1)

		require.ErrorIs(t, err, errInfraestructura)
		assert.Nil(t, lista)
	})

	t.Run("listado nil se normaliza a slice vacio", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, sprints := nuevoServicioSprints(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1}, nil)
		sprints.On("ListarPorProyecto", mock.Anything, uint(1)).Return(nil, nil)

		lista, err := servicio.Listar(context.Background(), 1)

		require.NoError(t, err)
		assert.NotNil(t, lista)
		assert.Empty(t, lista)
	})
}

func TestUsuarioServiceListarNormalizaNil(t *testing.T) {
	t.Parallel()

	usuarios := nuevoMockUsuarioRepository(t)
	servicio := NewUsuarioService(usuarios)
	usuarios.On("Listar", mock.Anything).Return(nil, nil)

	lista, err := servicio.Listar(context.Background())

	require.NoError(t, err)
	assert.NotNil(t, lista)
	assert.Empty(t, lista)
}

func TestCrearHistoriaPropagaErrorDeSiguienteNumero(t *testing.T) {
	t.Parallel()

	servicio, proyectos, historias := nuevoServicioHistorias(t)
	proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil)
	historias.On("SiguienteNumero", mock.Anything, uint(1)).Return(0, errInfraestructura)

	historia, err := servicio.Crear(context.Background(), 1, domain.NuevaHistoria{
		Titulo:      "Historia",
		Descripcion: "Descripcion",
		Prioridad:   domain.PrioridadMedia,
	})

	require.ErrorIs(t, err, errInfraestructura)
	assert.Nil(t, historia)
}

func TestCambiarEstadoSprintErroresDeInfraestructura(t *testing.T) {
	t.Parallel()

	t.Run("error al verificar el sprint activo", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, sprints := nuevoServicioSprints(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1}, nil)
		sprints.On("ObtenerPorID", mock.Anything, uint(1), uint(10)).
			Return(sprintDePrueba(10, "Sprint 1", domain.EstadoSprintPlanificado), nil)
		sprints.On("ObtenerActivo", mock.Anything, uint(1)).Return(nil, errInfraestructura)

		sprint, err := servicio.CambiarEstado(context.Background(), 1, 10, domain.EstadoSprintActivo)

		require.ErrorIs(t, err, errInfraestructura)
		assert.Nil(t, sprint)
	})

	t.Run("error al actualizar el estado", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, sprints := nuevoServicioSprints(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1}, nil)
		sprints.On("ObtenerPorID", mock.Anything, uint(1), uint(10)).
			Return(sprintDePrueba(10, "Sprint 1", domain.EstadoSprintPlanificado), nil)
		sprints.On("ObtenerActivo", mock.Anything, uint(1)).Return(nil, nil)
		sprints.On("ActualizarEstado", mock.Anything, mock.Anything).Return(errInfraestructura)

		sprint, err := servicio.CambiarEstado(context.Background(), 1, 10, domain.EstadoSprintActivo)

		require.ErrorIs(t, err, errInfraestructura)
		assert.Nil(t, sprint)
	})
}

func TestProyectoServiceErroresDeInfraestructura(t *testing.T) {
	t.Parallel()

	t.Run("Actualizar propaga el error del repositorio", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, _ := nuevoServicio(t)
		proyecto := proyectoDePrueba(1, "Proyecto", fechaInicio, &fechaFin, domain.EstadoActivo)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyecto, nil)
		proyectos.On("Actualizar", mock.Anything, proyecto).Return(errInfraestructura)

		actualizado, err := servicio.Actualizar(context.Background(), 1, domain.ActualizacionProyecto{
			Nombre:      "Nuevo nombre",
			FechaInicio: fechaInicio,
		})

		require.ErrorIs(t, err, errInfraestructura)
		assert.Nil(t, actualizado)
	})

	t.Run("CambiarEstado propaga el error del repositorio", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, _ := nuevoServicio(t)
		proyecto := proyectoDePrueba(1, "Proyecto", fechaInicio, &fechaFin, domain.EstadoActivo)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyecto, nil)
		proyectos.On("Actualizar", mock.Anything, proyecto).Return(errInfraestructura)

		actualizado, err := servicio.CambiarEstado(context.Background(), 1, domain.EstadoCerrado)

		require.ErrorIs(t, err, errInfraestructura)
		assert.Nil(t, actualizado)
	})

	t.Run("Eliminar propaga el error del repositorio", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, _ := nuevoServicio(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).
			Return(proyectoDePrueba(1, "Proyecto", fechaInicio, &fechaFin, domain.EstadoActivo), nil)
		proyectos.On("TieneHistorial", mock.Anything, uint(1)).Return(false, nil)
		proyectos.On("Eliminar", mock.Anything, uint(1)).Return(errInfraestructura)

		err := servicio.Eliminar(context.Background(), 1)

		require.ErrorIs(t, err, errInfraestructura)
	})
}

func TestAsignarASprintErroresDeInfraestructura(t *testing.T) {
	t.Parallel()

	t.Run("proyecto inexistente", func(t *testing.T) {
		t.Parallel()
		servicio, p, _, _, _ := nuevoServicioSprintBacklog(t)
		p.On("ObtenerPorID", mock.Anything, uint(1)).Return(nil, errInfraestructura)

		err := servicio.AsignarASprint(context.Background(), 1, 10, 100)

		require.ErrorIs(t, err, errInfraestructura)
	})

	t.Run("error al obtener el sprint", func(t *testing.T) {
		t.Parallel()
		servicio, p, sp, _, _ := nuevoServicioSprintBacklog(t)
		p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1}, nil)
		sp.On("ObtenerPorID", mock.Anything, uint(1), uint(10)).Return(nil, errInfraestructura)

		err := servicio.AsignarASprint(context.Background(), 1, 10, 100)

		require.ErrorIs(t, err, errInfraestructura)
	})

	t.Run("error al obtener la historia", func(t *testing.T) {
		t.Parallel()
		servicio, p, sp, h, _ := nuevoServicioSprintBacklog(t)
		p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1}, nil)
		sp.On("ObtenerPorID", mock.Anything, uint(1), uint(10)).Return(sprintActivoDePrueba(10, 1), nil)
		h.On("ObtenerPorID", mock.Anything, uint(1), uint(100)).Return(nil, errInfraestructura)

		err := servicio.AsignarASprint(context.Background(), 1, 10, 100)

		require.ErrorIs(t, err, errInfraestructura)
	})

	t.Run("error al verificar la asignacion", func(t *testing.T) {
		t.Parallel()
		servicio, p, sp, h, sb := nuevoServicioSprintBacklog(t)
		p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1}, nil)
		sp.On("ObtenerPorID", mock.Anything, uint(1), uint(10)).Return(sprintActivoDePrueba(10, 1), nil)
		h.On("ObtenerPorID", mock.Anything, uint(1), uint(100)).Return(historiaBacklogDePrueba(100, 1), nil)
		sb.On("ObtenerPorHistoria", mock.Anything, uint(100)).Return(nil, errInfraestructura)

		err := servicio.AsignarASprint(context.Background(), 1, 10, 100)

		require.ErrorIs(t, err, errInfraestructura)
	})

	t.Run("error al calcular el orden", func(t *testing.T) {
		t.Parallel()
		servicio, p, sp, h, sb := nuevoServicioSprintBacklog(t)
		p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1}, nil)
		sp.On("ObtenerPorID", mock.Anything, uint(1), uint(10)).Return(sprintActivoDePrueba(10, 1), nil)
		h.On("ObtenerPorID", mock.Anything, uint(1), uint(100)).Return(historiaBacklogDePrueba(100, 1), nil)
		sb.On("ObtenerPorHistoria", mock.Anything, uint(100)).Return(nil, nil)
		sb.On("ObtenerMaxOrden", mock.Anything, uint(1), uint(10)).Return(0, errInfraestructura)

		err := servicio.AsignarASprint(context.Background(), 1, 10, 100)

		require.ErrorIs(t, err, errInfraestructura)
	})

	t.Run("error al persistir la asignacion", func(t *testing.T) {
		t.Parallel()
		servicio, p, sp, h, sb := nuevoServicioSprintBacklog(t)
		p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1}, nil)
		sp.On("ObtenerPorID", mock.Anything, uint(1), uint(10)).Return(sprintActivoDePrueba(10, 1), nil)
		h.On("ObtenerPorID", mock.Anything, uint(1), uint(100)).Return(historiaBacklogDePrueba(100, 1), nil)
		sb.On("ObtenerPorHistoria", mock.Anything, uint(100)).Return(nil, nil)
		sb.On("ObtenerMaxOrden", mock.Anything, uint(1), uint(10)).Return(0, nil)
		sb.On("Asignar", mock.Anything, mock.AnythingOfType("*domain.SprintBacklogItem")).Return(errInfraestructura)

		err := servicio.AsignarASprint(context.Background(), 1, 10, 100)

		require.ErrorIs(t, err, errInfraestructura)
	})
}

func TestDesasignarDeSprintErroresYValidaciones(t *testing.T) {
	t.Parallel()

	t.Run("proyecto inexistente", func(t *testing.T) {
		t.Parallel()
		servicio, p, _, _, _ := nuevoServicioSprintBacklog(t)
		p.On("ObtenerPorID", mock.Anything, uint(1)).Return(nil, errInfraestructura)

		err := servicio.DesasignarDeSprint(context.Background(), 1, 10, 100)

		require.ErrorIs(t, err, errInfraestructura)
	})

	t.Run("error al obtener el sprint", func(t *testing.T) {
		t.Parallel()
		servicio, p, sp, _, _ := nuevoServicioSprintBacklog(t)
		p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1}, nil)
		sp.On("ObtenerPorID", mock.Anything, uint(1), uint(10)).Return(nil, errInfraestructura)

		err := servicio.DesasignarDeSprint(context.Background(), 1, 10, 100)

		require.ErrorIs(t, err, errInfraestructura)
	})

	t.Run("error al verificar la asignacion", func(t *testing.T) {
		t.Parallel()
		servicio, p, sp, _, sb := nuevoServicioSprintBacklog(t)
		p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1}, nil)
		sp.On("ObtenerPorID", mock.Anything, uint(1), uint(10)).Return(sprintActivoDePrueba(10, 1), nil)
		sb.On("ObtenerPorHistoria", mock.Anything, uint(100)).Return(nil, errInfraestructura)

		err := servicio.DesasignarDeSprint(context.Background(), 1, 10, 100)

		require.ErrorIs(t, err, errInfraestructura)
	})

	t.Run("historia sin asignacion", func(t *testing.T) {
		t.Parallel()
		servicio, p, sp, _, sb := nuevoServicioSprintBacklog(t)
		p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1}, nil)
		sp.On("ObtenerPorID", mock.Anything, uint(1), uint(10)).Return(sprintActivoDePrueba(10, 1), nil)
		sb.On("ObtenerPorHistoria", mock.Anything, uint(100)).Return(nil, nil)

		err := servicio.DesasignarDeSprint(context.Background(), 1, 10, 100)

		require.ErrorIs(t, err, domain.ErrHistoriaNoPerteneceASprint)
	})

	t.Run("historia asignada a otro sprint", func(t *testing.T) {
		t.Parallel()
		servicio, p, sp, _, sb := nuevoServicioSprintBacklog(t)
		p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1}, nil)
		sp.On("ObtenerPorID", mock.Anything, uint(1), uint(10)).Return(sprintActivoDePrueba(10, 1), nil)
		sb.On("ObtenerPorHistoria", mock.Anything, uint(100)).
			Return(&domain.SprintBacklogItem{ID: 1, ProyectoID: 1, SprintID: 20, HistoriaID: 100}, nil)

		err := servicio.DesasignarDeSprint(context.Background(), 1, 10, 100)

		require.ErrorIs(t, err, domain.ErrHistoriaNoPerteneceASprint)
	})

	t.Run("error al persistir la desasignacion", func(t *testing.T) {
		t.Parallel()
		servicio, p, sp, _, sb := nuevoServicioSprintBacklog(t)
		p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1}, nil)
		sp.On("ObtenerPorID", mock.Anything, uint(1), uint(10)).Return(sprintActivoDePrueba(10, 1), nil)
		sb.On("ObtenerPorHistoria", mock.Anything, uint(100)).
			Return(&domain.SprintBacklogItem{ID: 1, ProyectoID: 1, SprintID: 10, HistoriaID: 100}, nil)
		sb.On("Desasignar", mock.Anything, uint(1), uint(10), uint(100)).Return(errInfraestructura)

		err := servicio.DesasignarDeSprint(context.Background(), 1, 10, 100)

		require.ErrorIs(t, err, errInfraestructura)
	})
}
