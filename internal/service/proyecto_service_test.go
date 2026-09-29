package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

var (
	fechaInicio = time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC)
	fechaFin    = time.Date(2026, 12, 18, 0, 0, 0, 0, time.UTC)
	fechaAntes  = time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	fechaMisma  = time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC)
	fechaHoy    = time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
)

func nuevoServicio(t *testing.T) (*ProyectoService, *MockProyectoRepository, *MockIntegranteRepository) {
	t.Helper()
	proyectos := nuevoMockProyectoRepository(t)
	integrantes := nuevoMockIntegranteRepository(t)
	return NewProyectoService(proyectos, integrantes), proyectos, integrantes
}

func TestCrearProyecto(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre          string
		nuevo           domain.NuevoProyecto
		preparar        func(p *MockProyectoRepository, i *MockIntegranteRepository)
		errEsperado     error
		proyectoID      uint
		integrantes     []domain.Integrante
		escenario       string
		consultaLlamada bool
	}{
		{
			nombre: "E-01 alta con nombre, fechas e integrantes",
			nuevo: domain.NuevoProyecto{
				Nombre:      "Software Metrics & Estimation",
				FechaInicio: fechaInicio,
				FechaFin:    &fechaFin,
				Integrantes: []domain.IntegranteInput{
					{Nombre: "Aguilera Sebastián", Email: "sebas.aguilera@utn.edu.ar"},
					{Nombre: "Choquevillca Celeste", Email: "celeste.choque@utn.edu.ar"},
				},
			},
			preparar: func(p *MockProyectoRepository, i *MockIntegranteRepository) {
				i.On("ObtenerPorEmail", mock.Anything, "sebas.aguilera@utn.edu.ar").Return(nil, domain.ErrIntegranteNoEncontrado)
				i.On("ObtenerPorEmail", mock.Anything, "celeste.choque@utn.edu.ar").Return(nil, domain.ErrIntegranteNoEncontrado)
				p.On("CrearConIntegrantes", mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(asignarIDProyecto(1))
			},
			proyectoID: 1,
			integrantes: []domain.Integrante{
				{Nombre: "Aguilera Sebastián", Email: "sebas.aguilera@utn.edu.ar"},
				{Nombre: "Choquevillca Celeste", Email: "celeste.choque@utn.edu.ar"},
			},
			escenario: "E-01",
		},
		{
			nombre: "E-04 fecha de fin igual a la de inicio es valida",
			nuevo: domain.NuevoProyecto{
				Nombre:      "Proyecto de un dia",
				FechaInicio: fechaInicio,
				FechaFin:    &fechaMisma,
			},
			preparar: func(p *MockProyectoRepository, i *MockIntegranteRepository) {
				p.On("CrearConIntegrantes", mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(asignarIDProyecto(1))
			},
			proyectoID: 1,
			escenario:  "E-04",
		},
		{
			nombre: "E-02 fecha de fin ausente se persiste como nula",
			nuevo: domain.NuevoProyecto{
				Nombre:      "Proyecto sin fecha de fin",
				FechaInicio: fechaInicio,
			},
			preparar: func(p *MockProyectoRepository, i *MockIntegranteRepository) {
				p.On("CrearConIntegrantes", mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(asignarIDProyecto(1))
			},
			proyectoID: 1,
			escenario:  "E-02",
		},
		{
			nombre: "E-03 alta sin integrantes",
			nuevo: domain.NuevoProyecto{
				Nombre:      "Proyecto sin equipo",
				FechaInicio: fechaInicio,
				Integrantes: []domain.IntegranteInput{},
			},
			preparar: func(p *MockProyectoRepository, i *MockIntegranteRepository) {
				p.On("CrearConIntegrantes", mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(asignarIDProyecto(1))
			},
			proyectoID: 1,
			escenario:  "E-03",
		},
		{
			nombre: "RB-01 nombre vacio",
			nuevo: domain.NuevoProyecto{
				Nombre:      "",
				FechaInicio: fechaInicio,
			},
			errEsperado: domain.ErrNombreObligatorio,
			escenario:   "E-08",
		},
		{
			nombre: "E-09 nombre compuesto solo por espacios",
			nuevo: domain.NuevoProyecto{
				Nombre:      "   \t  ",
				FechaInicio: fechaInicio,
			},
			errEsperado: domain.ErrNombreObligatorio,
			escenario:   "E-09",
		},
		{
			nombre: "E-06 nombre de 150 caracteres se acepta",
			nuevo: domain.NuevoProyecto{
				Nombre:      strings.Repeat("A", 150),
				FechaInicio: fechaInicio,
			},
			preparar: func(p *MockProyectoRepository, i *MockIntegranteRepository) {
				p.On("CrearConIntegrantes", mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(asignarIDProyecto(1))
			},
			proyectoID: 1,
			escenario:  "E-06",
		},
		{
			nombre: "E-07 nombre de 151 caracteres se rechaza",
			nuevo: domain.NuevoProyecto{
				Nombre:      strings.Repeat("A", 151),
				FechaInicio: fechaInicio,
			},
			errEsperado: domain.ErrNombreDemasiadoLargo,
			escenario:   "E-07",
		},
		{
			nombre: "RB-02 fecha de inicio ausente",
			nuevo: domain.NuevoProyecto{
				Nombre: "Proyecto sin inicio",
			},
			errEsperado: domain.ErrFechaInicioObligatoria,
			escenario:   "E-10",
		},
		{
			nombre: "E-11 fecha de fin anterior a la de inicio",
			nuevo: domain.NuevoProyecto{
				Nombre:      "Proyecto invertido",
				FechaInicio: fechaInicio,
				FechaFin:    &fechaAntes,
			},
			errEsperado: domain.ErrFechaFinAnteriorAlInicio,
			escenario:   "E-11",
		},
		{
			nombre: "RB-07 integrante con nombre vacio",
			nuevo: domain.NuevoProyecto{
				Nombre:      "Proyecto con integrante invalido",
				FechaInicio: fechaInicio,
				Integrantes: []domain.IntegranteInput{
					{Nombre: "   ", Email: "ada@utn.edu.ar"},
				},
			},
			errEsperado: domain.ErrIntegranteNombreVacio,
			escenario:   "E-15",
		},
		{
			nombre: "RB-07 integrante con email invalido",
			nuevo: domain.NuevoProyecto{
				Nombre:      "Proyecto con email invalido",
				FechaInicio: fechaInicio,
				Integrantes: []domain.IntegranteInput{
					{Nombre: "Ada Lovelace", Email: "ada.utn.edu.ar"},
				},
			},
			errEsperado: domain.ErrIntegranteEmailInvalido,
			escenario:   "E-15",
		},
		{
			nombre: "E-17 email duplicado con distinta capitalizacion en el mismo request",
			nuevo: domain.NuevoProyecto{
				Nombre:      "Proyecto con duplicado",
				FechaInicio: fechaInicio,
				Integrantes: []domain.IntegranteInput{
					{Nombre: "Ada Lovelace", Email: "Ada@utn.edu.ar"},
					{Nombre: "ada lovelace", Email: "ada@utn.edu.ar"},
				},
			},
			errEsperado: domain.ErrIntegranteYaAsociado,
			escenario:   "E-17",
		},
		{
			nombre: "E-19 el nombre se valida antes que la fecha de inicio",
			nuevo: domain.NuevoProyecto{
				Nombre:      "",
				FechaInicio: time.Time{},
			},
			errEsperado: domain.ErrNombreObligatorio,
			escenario:   "E-19",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			servicio, proyectos, integrantes := nuevoServicio(t)
			if c.preparar != nil {
				c.preparar(proyectos, integrantes)
			}

			proyecto, err := servicio.Crear(context.Background(), c.nuevo)

			if c.errEsperado != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, c.errEsperado, "escenario BDD %s", c.escenario)
				assert.Nil(t, proyecto)
				proyectos.AssertNotCalled(t, "CrearConIntegrantes", mock.Anything, mock.Anything, mock.Anything)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, proyecto)
			assert.Equal(t, c.proyectoID, proyecto.ID, "escenario BDD %s", c.escenario)
			assert.Equal(t, domain.EstadoActivo, proyecto.Estado, "RB-04: todo proyecto nace Activo")
			assert.Equal(t, strings.TrimSpace(c.nuevo.Nombre), proyecto.Nombre)
			assert.False(t, c.nuevo.FechaInicio.IsZero() && proyecto.FechaInicio.IsZero())
			if c.nuevo.FechaFin == nil {
				assert.Nil(t, proyecto.FechaFin)
			}
			if c.integrantes != nil {
				require.Len(t, proyecto.Integrantes, len(c.integrantes))
				for idx, esperado := range c.integrantes {
					assert.Equal(t, esperado.Email, proyecto.Integrantes[idx].Email)
					assert.Equal(t, esperado.Nombre, proyecto.Integrantes[idx].Nombre)
				}
			}
			proyectos.AssertExpectations(t)
			integrantes.AssertExpectations(t)
		})
	}
}

func TestCrearProyectoReutilizaIntegranteExistente(t *testing.T) {
	t.Parallel()

	servicio, proyectos, integrantes := nuevoServicio(t)
	existente := &domain.Integrante{ID: 7, Nombre: "ada", Email: "ada@utn.edu.ar"}

	integrantes.On("ObtenerPorEmail", mock.Anything, "ada@utn.edu.ar").Return(existente, nil)
	integrantes.On("ActualizarNombre", mock.Anything, uint(7), "Ada Lovelace").Return(nil)
	proyectos.On("CrearConIntegrantes", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	proyecto, err := servicio.Crear(context.Background(), domain.NuevoProyecto{
		Nombre:      "Proyecto reutilizado",
		FechaInicio: fechaInicio,
		Integrantes: []domain.IntegranteInput{{Nombre: "Ada Lovelace", Email: "Ada@UTN.edu.ar "}},
	})

	require.NoError(t, err, "escenario BDD E-18 y E-41")
	require.Len(t, proyecto.Integrantes, 1)
	assert.Equal(t, uint(7), proyecto.Integrantes[0].ID, "RB-09: se reutiliza el registro existente")
	proyectos.AssertExpectations(t)
	integrantes.AssertExpectations(t)
}

func TestCrearProyectoPropagaErrorDePersistencia(t *testing.T) {
	t.Parallel()

	servicio, proyectos, _ := nuevoServicio(t)
	proyectos.On("CrearConIntegrantes", mock.Anything, mock.Anything, mock.Anything).
		Return(fmt.Errorf("violacion de integridad: %w", domain.ErrErrorDePersistencia))

	proyecto, err := servicio.Crear(context.Background(), domain.NuevoProyecto{
		Nombre:      "Proyecto que falla",
		FechaInicio: fechaInicio,
	})

	require.Error(t, err, "escenario BDD E-49")
	assert.ErrorIs(t, err, domain.ErrErrorDePersistencia)
	assert.Nil(t, proyecto)
}

// TestCrearProyectoRechazaEmailsDuplicadosEnElMismoAlta cubre E-14 y E-43: si el
// mismo email aparece dos veces en el payload, el alta se rechaza antes de
// tocar la base. Sin esta guarda el servicio reutilizaba el primer integrante
// para ambas apariciones y el proyecto se creaba con menos gente de la que se
// pidio, en silencio.
func TestCrearProyectoRechazaEmailsDuplicadosEnElMismoAlta(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string
		emails []string
	}{
		{"identico repetido", []string{"dup@utn.edu.ar", "dup@utn.edu.ar"}},
		{"variacion de mayusculas", []string{"dup@utn.edu.ar", "DUP@utn.edu.ar"}},
		{"variacion de espacios", []string{"dup@utn.edu.ar", " dup@utn.edu.ar "}},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			servicio, proyectos, _ := nuevoServicio(t)
			equipo := make([]domain.IntegranteInput, 0, len(caso.emails))
			for _, email := range caso.emails {
				equipo = append(equipo, domain.IntegranteInput{Nombre: "Ada", Email: email})
			}

			proyecto, err := servicio.Crear(context.Background(), domain.NuevoProyecto{
				Nombre:      "Proyecto con duplicado",
				FechaInicio: fechaInicio,
				Integrantes: equipo,
			})

			require.Error(t, err, "el alta debe rechazarse: el email esta repetido")
			assert.ErrorIs(t, err, domain.ErrIntegranteYaAsociado)
			assert.Nil(t, proyecto)
			proyectos.AssertNotCalled(t, "CrearConIntegrantes",
				mock.Anything, mock.Anything, mock.Anything,
				"no debe alcanzarse la base: la duplicacion se detecta en memoria")
		})
	}
}

func TestListarProyectos(t *testing.T) {
	t.Parallel()

	esperados := []domain.Proyecto{
		*proyectoDePrueba(1, "Proyecto A", fechaInicio, &fechaFin, domain.EstadoActivo),
		*proyectoDePrueba(2, "Proyecto B", fechaInicio, nil, domain.EstadoActivo),
	}

	t.Run("E-22 devuelve los proyectos del repositorio", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, _ := nuevoServicio(t)
		proyectos.On("Listar", mock.Anything).Return(esperados, nil)

		lista, err := servicio.Listar(context.Background())

		require.NoError(t, err)
		assert.Equal(t, esperados, lista)
		proyectos.AssertExpectations(t)
	})

	t.Run("E-21 sin proyectos devuelve un slice vacio y no nil", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, _ := nuevoServicio(t)
		proyectos.On("Listar", mock.Anything).Return(nil, nil)

		lista, err := servicio.Listar(context.Background())

		require.NoError(t, err, "escenario BDD E-21")
		assert.NotNil(t, lista, "la respuesta nunca debe serializarse como null")
		assert.Empty(t, lista)
	})

	t.Run("propaga el error del repositorio", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, _ := nuevoServicio(t)
		proyectos.On("Listar", mock.Anything).Return(nil, errors.New("connection refused"))

		lista, err := servicio.Listar(context.Background())

		require.Error(t, err)
		assert.Nil(t, lista)
	})
}

func TestObtenerProyectoPorID(t *testing.T) {
	t.Parallel()

	t.Run("E-05 devuelve el proyecto con sus integrantes", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, _ := nuevoServicio(t)
		esperado := proyectoDePrueba(3, "Proyecto C", fechaInicio, &fechaFin, domain.EstadoCerrado)
		esperado.Integrantes = []domain.Integrante{{ID: 1, Nombre: "Ada", Email: "ada@utn.edu.ar"}}
		proyectos.On("ObtenerPorID", mock.Anything, uint(3)).Return(esperado, nil)

		proyecto, err := servicio.ObtenerPorID(context.Background(), 3)

		require.NoError(t, err)
		assert.Equal(t, esperado, proyecto)
		proyectos.AssertExpectations(t)
	})

	t.Run("E-13 propaga ErrProyectoNoEncontrado", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, _ := nuevoServicio(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(99)).Return(nil, domain.ErrProyectoNoEncontrado)

		proyecto, err := servicio.ObtenerPorID(context.Background(), 99)

		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrProyectoNoEncontrado, "escenario BDD E-13")
		assert.Nil(t, proyecto)
	})
}

func TestActualizarProyecto(t *testing.T) {
	t.Parallel()

	nuevaFin := time.Date(2026, 12, 20, 0, 0, 0, 0, time.UTC)

	casos := []struct {
		nombre          string
		id              uint
		actualizacion   domain.ActualizacionProyecto
		preparar        func(p *MockProyectoRepository)
		errEsperado     error
		escenario       string
		consultaLlamada bool
	}{
		{
			nombre: "E-10 modifica nombre y fechas de un proyecto activo",
			id:     1,
			actualizacion: domain.ActualizacionProyecto{
				Nombre:      "Nombre nuevo",
				FechaInicio: fechaInicio,
				FechaFin:    &nuevaFin,
			},
			preparar: func(p *MockProyectoRepository) {
				proyecto := proyectoDePrueba(1, "Nombre viejo", fechaInicio, &fechaFin, domain.EstadoActivo)
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyecto, nil)
				p.On("Actualizar", mock.Anything, proyecto).Return(nil)
			},
			escenario: "E-10",
		},
		{
			nombre: "RB-01 nombre vacio en la modificacion",
			id:     1,
			actualizacion: domain.ActualizacionProyecto{
				Nombre:      "  ",
				FechaInicio: fechaInicio,
			},
			errEsperado: domain.ErrNombreObligatorio,
			escenario:   "E-28",
		},
		{
			nombre: "RB-02 fecha de inicio ausente en la modificacion",
			id:     1,
			actualizacion: domain.ActualizacionProyecto{
				Nombre: "Sin fecha",
			},
			errEsperado: domain.ErrFechaInicioObligatoria,
			escenario:   "E-28",
		},
		{
			nombre: "RB-03 fecha de fin anterior en la modificacion",
			id:     1,
			actualizacion: domain.ActualizacionProyecto{
				Nombre:      "Invertido",
				FechaInicio: fechaInicio,
				FechaFin:    &fechaAntes,
			},
			errEsperado: domain.ErrFechaFinAnteriorAlInicio,
			escenario:   "E-29",
		},
		{
			nombre: "RB-10 E-32 proyecto cerrado no se puede modificar",
			id:     1,
			actualizacion: domain.ActualizacionProyecto{
				Nombre:      "Intento sobre cerrado",
				FechaInicio: fechaInicio,
				FechaFin:    &fechaFin,
			},
			preparar: func(p *MockProyectoRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).
					Return(proyectoDePrueba(1, "Cerrado", fechaInicio, &fechaFin, domain.EstadoCerrado), nil)
			},
			errEsperado: domain.ErrProyectoCerrado,
			escenario:   "E-32",
		},
		{
			nombre: "E-30 proyecto inexistente en la modificacion",
			id:     99,
			actualizacion: domain.ActualizacionProyecto{
				Nombre:      "No existe",
				FechaInicio: fechaInicio,
			},
			preparar: func(p *MockProyectoRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(99)).Return(nil, domain.ErrProyectoNoEncontrado)
			},
			errEsperado: domain.ErrProyectoNoEncontrado,
			escenario:   "E-30",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			servicio, proyectos, _ := nuevoServicio(t)
			if c.preparar != nil {
				c.preparar(proyectos)
			}

			proyecto, err := servicio.Actualizar(context.Background(), c.id, c.actualizacion)

			if c.errEsperado != nil {
				require.Error(t, err, "escenario BDD %s", c.escenario)
				assert.ErrorIs(t, err, c.errEsperado, "escenario BDD %s", c.escenario)
				assert.Nil(t, proyecto)
				proyectos.AssertNotCalled(t, "Actualizar", mock.Anything, mock.Anything)
				return
			}

			require.NoError(t, err, "escenario BDD %s", c.escenario)
			require.NotNil(t, proyecto)
			assert.Equal(t, c.actualizacion.Nombre, proyecto.Nombre)
			assert.Equal(t, c.actualizacion.FechaInicio, proyecto.FechaInicio)
			if c.actualizacion.FechaFin == nil {
				assert.Nil(t, proyecto.FechaFin)
			} else {
				require.NotNil(t, proyecto.FechaFin)
				assert.Equal(t, *c.actualizacion.FechaFin, *proyecto.FechaFin)
			}
			proyectos.AssertExpectations(t)
		})
	}
}

func TestCambiarEstadoProyecto(t *testing.T) {
	t.Parallel()

	t.Run("E-33 cierra un proyecto que tiene fecha de fin", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, _ := nuevoServicio(t)
		proyecto := proyectoDePrueba(1, "Proyecto", fechaInicio, &fechaFin, domain.EstadoActivo)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyecto, nil)
		proyectos.On("Actualizar", mock.Anything, proyecto).Return(nil)

		actualizado, err := servicio.CambiarEstado(context.Background(), 1, domain.EstadoCerrado)

		require.NoError(t, err, "escenario BDD E-33")
		assert.Equal(t, domain.EstadoCerrado, actualizado.Estado)
		proyectos.AssertExpectations(t)
	})

	t.Run("E-34 reabre un proyecto cerrado", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, _ := nuevoServicio(t)
		proyecto := proyectoDePrueba(1, "Proyecto", fechaInicio, &fechaFin, domain.EstadoCerrado)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyecto, nil)
		proyectos.On("Actualizar", mock.Anything, proyecto).Return(nil)

		actualizado, err := servicio.CambiarEstado(context.Background(), 1, domain.EstadoActivo)

		require.NoError(t, err, "escenario BDD E-34")
		assert.Equal(t, domain.EstadoActivo, actualizado.Estado)
		proyectos.AssertExpectations(t)
	})

	t.Run("E-35 transicion al mismo estado es idempotente y no escribe", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, _ := nuevoServicio(t)
		proyecto := proyectoDePrueba(1, "Proyecto", fechaInicio, &fechaFin, domain.EstadoActivo)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyecto, nil)

		actualizado, err := servicio.CambiarEstado(context.Background(), 1, domain.EstadoActivo)

		require.NoError(t, err, "escenario BDD E-35")
		assert.Equal(t, domain.EstadoActivo, actualizado.Estado)
		proyectos.AssertNotCalled(t, "Actualizar", mock.Anything, mock.Anything)
	})

	t.Run("E-37 no se puede cerrar un proyecto sin fecha de fin", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, _ := nuevoServicio(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).
			Return(proyectoDePrueba(1, "Proyecto", fechaInicio, nil, domain.EstadoActivo), nil)

		proyecto, err := servicio.CambiarEstado(context.Background(), 1, domain.EstadoCerrado)

		require.Error(t, err, "escenario BDD E-37")
		assert.ErrorIs(t, err, domain.ErrFechaFinRequerida)
		assert.Nil(t, proyecto)
		proyectos.AssertNotCalled(t, "Actualizar", mock.Anything, mock.Anything)
	})

	t.Run("E-36 estado invalido se rechaza antes de consultar el proyecto", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, _ := nuevoServicio(t)

		proyecto, err := servicio.CambiarEstado(context.Background(), 1, domain.EstadoProyecto("activo"))

		require.Error(t, err, "escenario BDD E-36")
		assert.ErrorIs(t, err, domain.ErrEstadoInvalido)
		assert.Nil(t, proyecto)
		proyectos.AssertNotCalled(t, "ObtenerPorID", mock.Anything, mock.Anything)
	})

	t.Run("E-38 proyecto inexistente al cambiar el estado", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, _ := nuevoServicio(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(99)).Return(nil, domain.ErrProyectoNoEncontrado)

		proyecto, err := servicio.CambiarEstado(context.Background(), 99, domain.EstadoCerrado)

		require.Error(t, err, "escenario BDD E-38")
		assert.ErrorIs(t, err, domain.ErrProyectoNoEncontrado)
		assert.Nil(t, proyecto)
	})
}

func TestEliminarProyecto(t *testing.T) {
	t.Parallel()

	t.Run("E-51 da de baja un proyecto sin historial", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, _ := nuevoServicio(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).
			Return(proyectoDePrueba(1, "Proyecto", fechaInicio, &fechaFin, domain.EstadoActivo), nil)
		proyectos.On("TieneHistorial", mock.Anything, uint(1)).Return(false, nil)
		proyectos.On("Eliminar", mock.Anything, uint(1)).Return(nil)

		err := servicio.Eliminar(context.Background(), 1)

		require.NoError(t, err, "escenario BDD E-51")
		proyectos.AssertExpectations(t)
	})

	t.Run("E-52 no se da de baja un proyecto con historial", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, _ := nuevoServicio(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).
			Return(proyectoDePrueba(1, "Proyecto", fechaInicio, &fechaFin, domain.EstadoActivo), nil)
		proyectos.On("TieneHistorial", mock.Anything, uint(1)).Return(true, nil)

		err := servicio.Eliminar(context.Background(), 1)

		require.Error(t, err, "escenario BDD E-52")
		assert.ErrorIs(t, err, domain.ErrProyectoConHistorial)
		proyectos.AssertNotCalled(t, "Eliminar", mock.Anything, mock.Anything)
	})

	t.Run("E-50 proyecto ya dado de baja responde como inexistente", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, _ := nuevoServicio(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(nil, domain.ErrProyectoNoEncontrado)

		err := servicio.Eliminar(context.Background(), 1)

		require.Error(t, err, "escenario BDD E-50")
		assert.ErrorIs(t, err, domain.ErrProyectoNoEncontrado)
		proyectos.AssertNotCalled(t, "Eliminar", mock.Anything, mock.Anything)
	})

	t.Run("propaga el error de TieneHistorial", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, _ := nuevoServicio(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).
			Return(proyectoDePrueba(1, "Proyecto", fechaInicio, &fechaFin, domain.EstadoActivo), nil)
		proyectos.On("TieneHistorial", mock.Anything, uint(1)).Return(false, errors.New("timeout"))

		err := servicio.Eliminar(context.Background(), 1)

		require.Error(t, err)
		proyectos.AssertNotCalled(t, "Eliminar", mock.Anything, mock.Anything)
	})
}

func TestAgregarIntegrante(t *testing.T) {
	t.Parallel()

	entrada := domain.IntegranteInput{Nombre: "Ada Lovelace", Email: "Ada@UTN.edu.ar"}

	t.Run("E-40 agrega un integrante nuevo", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, integrantes := nuevoServicio(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).
			Return(proyectoDePrueba(1, "Proyecto", fechaInicio, &fechaFin, domain.EstadoActivo), nil)
		integrantes.On("ObtenerPorEmail", mock.Anything, "ada@utn.edu.ar").Return(nil, domain.ErrIntegranteNoEncontrado)
		integrantes.On("Crear", mock.Anything, mock.Anything).Return(nil).Run(asignarIDIntegrante(4))
		integrantes.On("Vincular", mock.Anything, uint(1), uint(4)).Return(nil)

		agregado, err := servicio.AgregarIntegrante(context.Background(), 1, entrada)

		require.NoError(t, err, "escenario BDD E-40")
		assert.Equal(t, uint(4), agregado.ID, "el identificador lo genera el repositorio")
		assert.Equal(t, "ada@utn.edu.ar", agregado.Email, "el email se persiste normalizado")
		assert.Equal(t, "Ada Lovelace", agregado.Nombre)
		proyectos.AssertExpectations(t)
		integrantes.AssertExpectations(t)
	})

	t.Run("E-44 el integrante ya forma parte del proyecto", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, integrantes := nuevoServicio(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).
			Return(proyectoDePrueba(1, "Proyecto", fechaInicio, &fechaFin, domain.EstadoActivo), nil)
		integrantes.On("ObtenerPorEmail", mock.Anything, "ada@utn.edu.ar").
			Return(&domain.Integrante{ID: 4, Nombre: "Ada", Email: "ada@utn.edu.ar"}, nil)
		integrantes.On("EstaAsociado", mock.Anything, uint(1), uint(4)).Return(true, nil)

		agregado, err := servicio.AgregarIntegrante(context.Background(), 1, entrada)

		require.Error(t, err, "escenario BDD E-44")
		assert.ErrorIs(t, err, domain.ErrIntegranteYaAsociado)
		assert.Nil(t, agregado)
		integrantes.AssertNotCalled(t, "Vincular", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("E-53 proyecto dado de baja responde como inexistente", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, _ := nuevoServicio(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(nil, domain.ErrProyectoNoEncontrado)

		agregado, err := servicio.AgregarIntegrante(context.Background(), 1, entrada)

		require.Error(t, err, "escenario BDD E-53")
		assert.ErrorIs(t, err, domain.ErrProyectoNoEncontrado)
		assert.Nil(t, agregado)
	})

	t.Run("RB-10 E-48 no se agrega integrantes a un proyecto cerrado", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, _ := nuevoServicio(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).
			Return(proyectoDePrueba(1, "Cerrado", fechaInicio, &fechaFin, domain.EstadoCerrado), nil)

		agregado, err := servicio.AgregarIntegrante(context.Background(), 1, entrada)

		require.Error(t, err, "escenario BDD E-48")
		assert.ErrorIs(t, err, domain.ErrProyectoCerrado)
		assert.Nil(t, agregado)
	})

	t.Run("RB-07 E-42 email con formato Nombre <correo> se rechaza", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, _ := nuevoServicio(t)

		agregado, err := servicio.AgregarIntegrante(context.Background(), 1, domain.IntegranteInput{
			Nombre: "Ada",
			Email:  "Ada Lovelace <ada@utn.edu.ar>",
		})

		require.Error(t, err, "escenario BDD E-42")
		assert.ErrorIs(t, err, domain.ErrIntegranteEmailInvalido)
		assert.Nil(t, agregado)
		proyectos.AssertNotCalled(t, "ObtenerPorID", mock.Anything, mock.Anything)
	})

	t.Run("RB-07 E-43 email vacio se rechaza", func(t *testing.T) {
		t.Parallel()
		servicio, _, _ := nuevoServicio(t)

		agregado, err := servicio.AgregarIntegrante(context.Background(), 1, domain.IntegranteInput{
			Nombre: "Ada",
			Email:  "   ",
		})

		require.Error(t, err, "escenario BDD E-43")
		assert.ErrorIs(t, err, domain.ErrIntegranteEmailInvalido)
		assert.Nil(t, agregado)
	})

	t.Run("rechaza un nombre de integrante de mas de 100 caracteres", func(t *testing.T) {
		t.Parallel()
		servicio, _, _ := nuevoServicio(t)

		agregado, err := servicio.AgregarIntegrante(context.Background(), 1, domain.IntegranteInput{
			Nombre: strings.Repeat("A", 101),
			Email:  "ada@utn.edu.ar",
		})

		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrIntegranteNombreLargo)
		assert.Nil(t, agregado)
	})

	t.Run("propaga el error al buscar el proyecto", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, _ := nuevoServicio(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(nil, errors.New("connection reset"))

		agregado, err := servicio.AgregarIntegrante(context.Background(), 1, entrada)

		require.Error(t, err)
		assert.Nil(t, agregado)
	})

	t.Run("propaga el error al buscar el integrante por email", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, integrantes := nuevoServicio(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).
			Return(proyectoDePrueba(1, "Proyecto", fechaInicio, &fechaFin, domain.EstadoActivo), nil)
		integrantes.On("ObtenerPorEmail", mock.Anything, "ada@utn.edu.ar").Return(nil, errors.New("timeout"))

		agregado, err := servicio.AgregarIntegrante(context.Background(), 1, entrada)

		require.Error(t, err)
		assert.Nil(t, agregado)
	})

	t.Run("propaga el error al verificar la asociacion", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, integrantes := nuevoServicio(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).
			Return(proyectoDePrueba(1, "Proyecto", fechaInicio, &fechaFin, domain.EstadoActivo), nil)
		integrantes.On("ObtenerPorEmail", mock.Anything, "ada@utn.edu.ar").
			Return(&domain.Integrante{ID: 4, Nombre: "Ada", Email: "ada@utn.edu.ar"}, nil)
		integrantes.On("EstaAsociado", mock.Anything, uint(1), uint(4)).Return(false, errors.New("timeout"))

		agregado, err := servicio.AgregarIntegrante(context.Background(), 1, entrada)

		require.Error(t, err)
		assert.Nil(t, agregado)
	})

	t.Run("propaga el error al refrescar el nombre de un integrante existente", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, integrantes := nuevoServicio(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).
			Return(proyectoDePrueba(1, "Proyecto", fechaInicio, &fechaFin, domain.EstadoActivo), nil)
		integrantes.On("ObtenerPorEmail", mock.Anything, "ada@utn.edu.ar").
			Return(&domain.Integrante{ID: 4, Nombre: "Ada", Email: "ada@utn.edu.ar"}, nil)
		integrantes.On("EstaAsociado", mock.Anything, uint(1), uint(4)).Return(false, nil)
		integrantes.On("ActualizarNombre", mock.Anything, uint(4), "Ada Lovelace").Return(errors.New("timeout"))

		agregado, err := servicio.AgregarIntegrante(context.Background(), 1, entrada)

		require.Error(t, err)
		assert.Nil(t, agregado)
	})

	t.Run("propaga el error al vincular", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, integrantes := nuevoServicio(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).
			Return(proyectoDePrueba(1, "Proyecto", fechaInicio, &fechaFin, domain.EstadoActivo), nil)
		integrantes.On("ObtenerPorEmail", mock.Anything, "ada@utn.edu.ar").Return(nil, domain.ErrIntegranteNoEncontrado)
		integrantes.On("Crear", mock.Anything, mock.Anything).Return(nil).Run(asignarIDIntegrante(4))
		integrantes.On("Vincular", mock.Anything, uint(1), uint(4)).Return(errors.New("timeout"))

		agregado, err := servicio.AgregarIntegrante(context.Background(), 1, entrada)

		require.Error(t, err)
		assert.Nil(t, agregado)
	})
}

func TestQuitarIntegrante(t *testing.T) {
	t.Parallel()

	t.Run("E-46 desasocia un integrante del proyecto", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, integrantes := nuevoServicio(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).
			Return(proyectoDePrueba(1, "Proyecto", fechaInicio, &fechaFin, domain.EstadoActivo), nil)
		integrantes.On("EstaAsociado", mock.Anything, uint(1), uint(4)).Return(true, nil)
		integrantes.On("Desvincular", mock.Anything, uint(1), uint(4)).Return(nil)

		err := servicio.QuitarIntegrante(context.Background(), 1, 4)

		require.NoError(t, err, "escenario BDD E-46")
		proyectos.AssertExpectations(t)
		integrantes.AssertExpectations(t)
	})

	t.Run("E-47 el integrante no forma parte del proyecto", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, integrantes := nuevoServicio(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).
			Return(proyectoDePrueba(1, "Proyecto", fechaInicio, &fechaFin, domain.EstadoActivo), nil)
		integrantes.On("EstaAsociado", mock.Anything, uint(1), uint(9)).Return(false, nil)

		err := servicio.QuitarIntegrante(context.Background(), 1, 9)

		require.Error(t, err, "escenario BDD E-47")
		assert.ErrorIs(t, err, domain.ErrIntegranteNoAsociado)
		integrantes.AssertNotCalled(t, "Desvincular", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("RB-10 E-48 no se quitan integrantes de un proyecto cerrado", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, _ := nuevoServicio(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).
			Return(proyectoDePrueba(1, "Cerrado", fechaInicio, &fechaFin, domain.EstadoCerrado), nil)

		err := servicio.QuitarIntegrante(context.Background(), 1, 4)

		require.Error(t, err, "escenario BDD E-48")
		assert.ErrorIs(t, err, domain.ErrProyectoCerrado)
	})

	t.Run("propaga el error al obtener el proyecto", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, _ := nuevoServicio(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).Return(nil, errors.New("connection reset"))

		err := servicio.QuitarIntegrante(context.Background(), 1, 4)

		require.Error(t, err)
	})

	t.Run("propaga el error al verificar la asociacion", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, integrantes := nuevoServicio(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).
			Return(proyectoDePrueba(1, "Proyecto", fechaInicio, &fechaFin, domain.EstadoActivo), nil)
		integrantes.On("EstaAsociado", mock.Anything, uint(1), uint(4)).Return(false, errors.New("timeout"))

		err := servicio.QuitarIntegrante(context.Background(), 1, 4)

		require.Error(t, err)
		integrantes.AssertNotCalled(t, "Desvincular", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("propaga el error al desvincular", func(t *testing.T) {
		t.Parallel()
		servicio, proyectos, integrantes := nuevoServicio(t)
		proyectos.On("ObtenerPorID", mock.Anything, uint(1)).
			Return(proyectoDePrueba(1, "Proyecto", fechaInicio, &fechaFin, domain.EstadoActivo), nil)
		integrantes.On("EstaAsociado", mock.Anything, uint(1), uint(4)).Return(true, nil)
		integrantes.On("Desvincular", mock.Anything, uint(1), uint(4)).Return(errors.New("timeout"))

		err := servicio.QuitarIntegrante(context.Background(), 1, 4)

		require.Error(t, err)
	})
}
