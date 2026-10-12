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

func nuevoServicioHistorias(t *testing.T) (*HistoriaService, *MockProyectoRepository, *MockHistoriaRepository) {
	t.Helper()
	proyectos := nuevoMockProyectoRepository(t)
	historias := nuevoMockHistoriaRepository(t)
	return NewHistoriaService(proyectos, historias), proyectos, historias
}

func strPtr(v string) *string { return &v }

func prioridadPtr(v domain.PrioridadHistoria) *domain.PrioridadHistoria { return &v }

func estadoPtr(v domain.EstadoHistoria) *domain.EstadoHistoria { return &v }

func TestCrearHistoria(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre      string
		nuevo       domain.NuevaHistoria
		preparar    func(p *MockProyectoRepository, h *MockHistoriaRepository)
		errEsperado error
		verificar   func(t *testing.T, historia *domain.Historia)
		escenario   string
	}{
		{
			nombre: "E-01 alta con todos los campos",
			nuevo: domain.NuevaHistoria{
				Titulo:      "Como Product Builder quiero agregar historias",
				Descripcion: "Permitir administrar el alcance del proyecto",
				Prioridad:   domain.PrioridadAlta,
				Estado:      domain.EstadoHistoriaBacklog,
			},
			preparar: func(p *MockProyectoRepository, h *MockHistoriaRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil)
				h.On("SiguienteNumero", mock.Anything, uint(1)).Return(1, nil)
				h.On("Crear", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
					args.Get(1).(*domain.Historia).ID = 10
				})
			},
			verificar: func(t *testing.T, historia *domain.Historia) {
				require.NotNil(t, historia)
				assert.Equal(t, "US-01", historia.Codigo)
				assert.Equal(t, 1, historia.Numero)
				assert.Equal(t, domain.PrioridadAlta, historia.Prioridad)
				assert.Equal(t, domain.EstadoHistoriaBacklog, historia.Estado)
				assert.Equal(t, uint(1), historia.ProyectoID)
				assert.Equal(t, uint(10), historia.ID)
			},
			escenario: "E-01",
		},
		{
			nombre: "E-02 sin prioridad ni estado toma los valores por defecto",
			nuevo: domain.NuevaHistoria{
				Titulo: "Historia con defaults",
			},
			preparar: func(p *MockProyectoRepository, h *MockHistoriaRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil)
				h.On("SiguienteNumero", mock.Anything, uint(1)).Return(1, nil)
				h.On("Crear", mock.Anything, mock.Anything).Return(nil)
			},
			verificar: func(t *testing.T, historia *domain.Historia) {
				assert.Equal(t, domain.PrioridadMedia, historia.Prioridad)
				assert.Equal(t, domain.EstadoHistoriaBacklog, historia.Estado)
			},
			escenario: "E-02",
		},
		{
			nombre: "E-03 descripcion opcional vacia se acepta",
			nuevo: domain.NuevaHistoria{
				Titulo: "Sin descripcion",
			},
			preparar: func(p *MockProyectoRepository, h *MockHistoriaRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil)
				h.On("SiguienteNumero", mock.Anything, uint(1)).Return(1, nil)
				h.On("Crear", mock.Anything, mock.Anything).Return(nil)
			},
			verificar: func(t *testing.T, historia *domain.Historia) {
				assert.Empty(t, historia.Descripcion)
			},
			escenario: "E-03",
		},
		{
			nombre: "E-06 titulo vacio",
			nuevo: domain.NuevaHistoria{
				Titulo: "",
			},
			errEsperado: domain.ErrTituloHistoriaObligatorio,
			escenario:   "E-06",
		},
		{
			nombre: "E-07 titulo compuesto solo por espacios",
			nuevo: domain.NuevaHistoria{
				Titulo: "   \t  ",
			},
			errEsperado: domain.ErrTituloHistoriaObligatorio,
			escenario:   "E-07",
		},
		{
			nombre: "E-08 titulo de 200 caracteres es valido",
			nuevo: domain.NuevaHistoria{
				Titulo: strings.Repeat("a", 200),
			},
			preparar: func(p *MockProyectoRepository, h *MockHistoriaRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil)
				h.On("SiguienteNumero", mock.Anything, uint(1)).Return(1, nil)
				h.On("Crear", mock.Anything, mock.Anything).Return(nil)
			},
			verificar: func(t *testing.T, historia *domain.Historia) {
				assert.Len(t, []rune(historia.Titulo), 200)
			},
			escenario: "E-08",
		},
		{
			nombre: "E-09 titulo de 201 caracteres es rechazado",
			nuevo: domain.NuevaHistoria{
				Titulo: strings.Repeat("a", 201),
			},
			errEsperado: domain.ErrTituloHistoriaDemasiadoLargo,
			escenario:   "E-09",
		},
		{
			nombre: "E-10 prioridad invalida",
			nuevo: domain.NuevaHistoria{
				Titulo:    "Historia valida",
				Prioridad: domain.PrioridadHistoria("Urgente"),
			},
			errEsperado: domain.ErrPrioridadHistoriaInvalida,
			escenario:   "E-10",
		},
		{
			nombre: "E-11 estado invalido",
			nuevo: domain.NuevaHistoria{
				Titulo: "Historia valida",
				Estado: domain.EstadoHistoria("En revision"),
			},
			errEsperado: domain.ErrEstadoHistoriaInvalido,
			escenario:   "E-11",
		},
		{
			nombre: "E-12 segunda historia usa el consecutivo US-02",
			nuevo: domain.NuevaHistoria{
				Titulo: "Segunda historia",
			},
			preparar: func(p *MockProyectoRepository, h *MockHistoriaRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil)
				h.On("SiguienteNumero", mock.Anything, uint(1)).Return(2, nil)
				h.On("Crear", mock.Anything, mock.Anything).Return(nil)
			},
			verificar: func(t *testing.T, historia *domain.Historia) {
				assert.Equal(t, "US-02", historia.Codigo)
			},
			escenario: "E-12",
		},
		{
			nombre: "E-14 colision de codigo concurrente",
			nuevo: domain.NuevaHistoria{
				Titulo: "Historia duplicada",
			},
			preparar: func(p *MockProyectoRepository, h *MockHistoriaRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil)
				h.On("SiguienteNumero", mock.Anything, uint(1)).Return(1, nil)
				h.On("Crear", mock.Anything, mock.Anything).Return(domain.ErrCodigoHistoriaDuplicado)
			},
			errEsperado: domain.ErrCodigoHistoriaDuplicado,
			escenario:   "E-14",
		},
		{
			nombre: "E-15 proyecto inexistente",
			nuevo: domain.NuevaHistoria{
				Titulo: "Historia en proyecto inexistente",
			},
			preparar: func(p *MockProyectoRepository, h *MockHistoriaRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(99)).Return(nil, domain.ErrProyectoNoEncontrado)
			},
			errEsperado: domain.ErrProyectoNoEncontrado,
			escenario:   "E-15",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()

			servicio, proyectos, historias := nuevoServicioHistorias(t)
			if c.preparar != nil {
				c.preparar(proyectos, historias)
			}

			idProyecto := uint(1)
			if c.escenario == "E-15" {
				idProyecto = 99
			}

			historia, err := servicio.Crear(context.Background(), idProyecto, c.nuevo)

			if c.errEsperado != nil {
				require.Error(t, err, "escenario BDD %s", c.escenario)
				assert.ErrorIs(t, err, c.errEsperado, "escenario BDD %s", c.escenario)
				assert.Nil(t, historia)
				return
			}

			require.NoError(t, err, "escenario BDD %s", c.escenario)
			if c.verificar != nil {
				c.verificar(t, historia)
			}
			proyectos.AssertExpectations(t)
			historias.AssertExpectations(t)
		})
	}
}

func TestCrearHistoriaNoConsultaRepositorioSiLaValidacionFalla(t *testing.T) {
	t.Parallel()

	servicio, proyectos, historias := nuevoServicioHistorias(t)

	_, err := servicio.Crear(context.Background(), 1, domain.NuevaHistoria{Titulo: "  "})
	require.ErrorIs(t, err, domain.ErrTituloHistoriaObligatorio)

	proyectos.AssertNotCalled(t, "ObtenerPorID", mock.Anything, mock.Anything)
	historias.AssertNotCalled(t, "Crear", mock.Anything, mock.Anything)
}

func TestListarHistorias(t *testing.T) {
	t.Parallel()

	t.Run("E-18 listado con historias", func(t *testing.T) {
		t.Parallel()

		servicio, proyectos, historias := nuevoServicioHistorias(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil)
		historias.On("ListarPorProyecto", mock.Anything, uint(1)).Return([]domain.Historia{
			{ID: 1, ProyectoID: 1, Numero: 1, Codigo: "US-01", Titulo: "Primera"},
			{ID: 2, ProyectoID: 1, Numero: 2, Codigo: "US-02", Titulo: "Segunda"},
		}, nil)

		lista, err := servicio.Listar(context.Background(), 1)

		require.NoError(t, err)
		require.Len(t, lista, 2)
		assert.Equal(t, "US-01", lista[0].Codigo)
		proyectos.AssertExpectations(t)
		historias.AssertExpectations(t)
	})

	t.Run("E-19 listado vacio devuelve slice no nulo", func(t *testing.T) {
		t.Parallel()

		servicio, proyectos, historias := nuevoServicioHistorias(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil)
		historias.On("ListarPorProyecto", mock.Anything, uint(1)).Return(nil, nil)

		lista, err := servicio.Listar(context.Background(), 1)

		require.NoError(t, err)
		assert.NotNil(t, lista)
		assert.Empty(t, lista)
	})

	t.Run("E-20 listado de proyecto inexistente", func(t *testing.T) {
		t.Parallel()

		servicio, proyectos, _ := nuevoServicioHistorias(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(99)).Return(nil, domain.ErrProyectoNoEncontrado)

		_, err := servicio.Listar(context.Background(), 99)

		assert.ErrorIs(t, err, domain.ErrProyectoNoEncontrado)
	})
}

func TestActualizarHistoria(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre        string
		actualizacion domain.ActualizacionHistoria
		preparar      func(p *MockProyectoRepository, h *MockHistoriaRepository)
		errEsperado   error
		verificar     func(t *testing.T, historia *domain.Historia)
		escenario     string
	}{
		{
			nombre:        "E-23 modificar solo la prioridad",
			actualizacion: domain.ActualizacionHistoria{Prioridad: prioridadPtr(domain.PrioridadAlta)},
			preparar: func(p *MockProyectoRepository, h *MockHistoriaRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil)
				h.On("ObtenerPorID", mock.Anything, uint(1), uint(5)).Return(&domain.Historia{
					ID: 5, ProyectoID: 1, Codigo: "US-01", Titulo: "Original",
					Prioridad: domain.PrioridadBaja, Estado: domain.EstadoHistoriaBacklog,
				}, nil)
				h.On("Actualizar", mock.Anything, mock.Anything).Return(nil)
			},
			verificar: func(t *testing.T, historia *domain.Historia) {
				assert.Equal(t, domain.PrioridadAlta, historia.Prioridad)
				assert.Equal(t, "Original", historia.Titulo, "los campos no informados no cambian")
			},
			escenario: "E-23",
		},
		{
			nombre: "E-24 modificar titulo, descripcion y estado",
			actualizacion: domain.ActualizacionHistoria{
				Titulo:      strPtr("Titulo nuevo"),
				Descripcion: strPtr("Descripcion nueva"),
				Estado:      estadoPtr(domain.EstadoHistoriaHecho),
			},
			preparar: func(p *MockProyectoRepository, h *MockHistoriaRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil)
				h.On("ObtenerPorID", mock.Anything, uint(1), uint(5)).Return(&domain.Historia{
					ID: 5, ProyectoID: 1, Codigo: "US-01", Titulo: "Original", Prioridad: domain.PrioridadMedia,
				}, nil)
				h.On("Actualizar", mock.Anything, mock.Anything).Return(nil)
			},
			verificar: func(t *testing.T, historia *domain.Historia) {
				assert.Equal(t, "Titulo nuevo", historia.Titulo)
				assert.Equal(t, "Descripcion nueva", historia.Descripcion)
				assert.Equal(t, domain.EstadoHistoriaHecho, historia.Estado)
				assert.Equal(t, domain.PrioridadMedia, historia.Prioridad)
			},
			escenario: "E-24",
		},
		{
			nombre:        "E-26 sin ningun campo informado",
			actualizacion: domain.ActualizacionHistoria{},
			errEsperado:   domain.ErrActualizacionHistoriaVacia,
			escenario:     "E-26",
		},
		{
			nombre:        "E-27 titulo vacio en la modificacion",
			actualizacion: domain.ActualizacionHistoria{Titulo: strPtr("   ")},
			errEsperado:   domain.ErrTituloHistoriaObligatorio,
			escenario:     "E-27",
		},
		{
			nombre:        "E-27 prioridad invalida en la modificacion",
			actualizacion: domain.ActualizacionHistoria{Prioridad: prioridadPtr(domain.PrioridadHistoria("Urgente"))},
			errEsperado:   domain.ErrPrioridadHistoriaInvalida,
			escenario:     "E-27",
		},
		{
			nombre:        "E-27 estado invalido en la modificacion",
			actualizacion: domain.ActualizacionHistoria{Estado: estadoPtr(domain.EstadoHistoria("Cerrada"))},
			errEsperado:   domain.ErrEstadoHistoriaInvalido,
			escenario:     "E-27",
		},
		{
			nombre:        "E-30 historia inexistente",
			actualizacion: domain.ActualizacionHistoria{Prioridad: prioridadPtr(domain.PrioridadAlta)},
			preparar: func(p *MockProyectoRepository, h *MockHistoriaRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil)
				h.On("ObtenerPorID", mock.Anything, uint(1), uint(5)).Return(nil, domain.ErrHistoriaNoEncontrada)
			},
			errEsperado: domain.ErrHistoriaNoEncontrada,
			escenario:   "E-30",
		},
		{
			nombre:        "E-31 proyecto inexistente",
			actualizacion: domain.ActualizacionHistoria{Prioridad: prioridadPtr(domain.PrioridadAlta)},
			preparar: func(p *MockProyectoRepository, h *MockHistoriaRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(99)).Return(nil, domain.ErrProyectoNoEncontrado)
			},
			errEsperado: domain.ErrProyectoNoEncontrado,
			escenario:   "E-31",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()

			servicio, proyectos, historias := nuevoServicioHistorias(t)
			if c.preparar != nil {
				c.preparar(proyectos, historias)
			}

			idProyecto := uint(1)
			if c.escenario == "E-31" {
				idProyecto = 99
			}

			historia, err := servicio.Actualizar(context.Background(), idProyecto, 5, c.actualizacion)

			if c.errEsperado != nil {
				require.Error(t, err, "escenario BDD %s", c.escenario)
				assert.ErrorIs(t, err, c.errEsperado, "escenario BDD %s", c.escenario)
				assert.Nil(t, historia)
				return
			}

			require.NoError(t, err, "escenario BDD %s", c.escenario)
			if c.verificar != nil {
				c.verificar(t, historia)
			}
			proyectos.AssertExpectations(t)
			historias.AssertExpectations(t)
		})
	}
}

func TestActualizarHistoriaPropagaErrorDePersistencia(t *testing.T) {
	t.Parallel()

	servicio, proyectos, historias := nuevoServicioHistorias(t)
	errorPersistencia := errors.New("conexion caida")
	proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Proyecto{ID: 1, Estado: domain.EstadoActivo}, nil)
	historias.On("ObtenerPorID", mock.Anything, uint(1), uint(5)).Return(&domain.Historia{ID: 5, ProyectoID: 1}, nil)
	historias.On("Actualizar", mock.Anything, mock.Anything).Return(errorPersistencia)

	_, err := servicio.Actualizar(context.Background(), 1, 5, domain.ActualizacionHistoria{Estado: estadoPtr(domain.EstadoHistoriaHecho)})

	require.Error(t, err)
	assert.ErrorIs(t, err, errorPersistencia)
}
