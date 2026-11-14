package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

func nuevoServicioDefectos(t *testing.T) (*DefectoService, *MockProyectoRepository, *MockHistoriaRepository, *MockSprintRepository, *MockUsuarioRepository, *MockDefectoRepository) {
	t.Helper()
	proyectos := nuevoMockProyectoRepository(t)
	historias := nuevoMockHistoriaRepository(t)
	sprints := nuevoMockSprintRepository(t)
	usuarios := nuevoMockUsuarioRepository(t)
	defectos := nuevoMockDefectoRepository(t)
	return NewDefectoService(proyectos, historias, sprints, usuarios, defectos), proyectos, historias, sprints, usuarios, defectos
}

func proyectoActivoDePrueba(id uint) *domain.Proyecto {
	return &domain.Proyecto{ID: id, Estado: domain.EstadoActivo}
}

func defectoDePrueba(id uint, estado domain.EstadoDefecto) *domain.Defecto {
	return &domain.Defecto{
		ID:                id,
		ProyectoID:        1,
		HistoriaID:        10,
		SprintDeteccionID: 20,
		Descripcion:       "Error en validación",
		Severidad:         domain.SeveridadMajor,
		Estado:            estado,
		CreadoEn:          time.Now().Add(-24 * time.Hour),
	}
}

func TestCrearDefecto(t *testing.T) {
	t.Parallel()

	descripcionValida := "  Error en validación  "

	casos := []struct {
		nombre      string
		nuevo       domain.NuevoDefecto
		preparar    func(p *MockProyectoRepository, h *MockHistoriaRepository, sp *MockSprintRepository, u *MockUsuarioRepository, d *MockDefectoRepository)
		errEsperado error
		verificar   func(t *testing.T, d *MockDefectoRepository)
		escenario   string
	}{
		{
			nombre: "BDD-01 registro exitoso con estado open",
			nuevo: domain.NuevoDefecto{
				Descripcion:       descripcionValida,
				Severidad:         domain.SeveridadMajor,
				HistoriaID:        10,
				SprintDeteccionID: 20,
			},
			preparar: func(p *MockProyectoRepository, h *MockHistoriaRepository, sp *MockSprintRepository, u *MockUsuarioRepository, d *MockDefectoRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyectoActivoDePrueba(1), nil)
				h.On("ObtenerPorID", mock.Anything, uint(1), uint(10)).Return(historiaBacklogDePrueba(10, 1), nil)
				sp.On("ObtenerPorID", mock.Anything, uint(1), uint(20)).Return(sprintActivoDePrueba(20, 1), nil)
				d.On("Crear", mock.Anything, mock.MatchedBy(func(defecto *domain.Defecto) bool {
					return defecto.Estado == domain.EstadoDefectoAbierto &&
						defecto.Severidad == domain.SeveridadMajor &&
						defecto.Descripcion == "Error en validación" &&
						defecto.HistoriaID == 10 &&
						defecto.SprintDeteccionID == 20 &&
						defecto.ProyectoID == 1
				})).Run(func(args mock.Arguments) {
					args.Get(1).(*domain.Defecto).ID = 1
				}).Return(nil)
			},
			verificar: func(t *testing.T, d *MockDefectoRepository) {
				d.AssertCalled(t, "Crear", mock.Anything, mock.AnythingOfType("*domain.Defecto"))
			},
			escenario: "BDD-01",
		},
		{
			nombre: "BDD-01 reportante existente",
			nuevo: domain.NuevoDefecto{
				Descripcion:       "Error",
				Severidad:         domain.SeveridadCritical,
				HistoriaID:        10,
				SprintDeteccionID: 20,
				ReportadoPor:      func() *uint { v := uint(30); return &v }(),
			},
			preparar: func(p *MockProyectoRepository, h *MockHistoriaRepository, sp *MockSprintRepository, u *MockUsuarioRepository, d *MockDefectoRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyectoActivoDePrueba(1), nil)
				h.On("ObtenerPorID", mock.Anything, uint(1), uint(10)).Return(historiaBacklogDePrueba(10, 1), nil)
				sp.On("ObtenerPorID", mock.Anything, uint(1), uint(20)).Return(sprintActivoDePrueba(20, 1), nil)
				u.On("ObtenerPorID", mock.Anything, uint(30)).Return(&domain.Usuario{ID: 30}, nil)
				d.On("Crear", mock.Anything, mock.MatchedBy(func(defecto *domain.Defecto) bool {
					return defecto.ReportadoPor != nil && *defecto.ReportadoPor == 30
				})).Return(nil)
			},
			escenario: "BDD-01",
		},
		{
			nombre: "BDD-02 severidad invalida",
			nuevo: domain.NuevoDefecto{
				Descripcion:       "Error",
				Severidad:         domain.SeveridadDefecto("urgente"),
				HistoriaID:        10,
				SprintDeteccionID: 20,
			},
			errEsperado: domain.ErrSeveridadDefectoInvalida,
			escenario:   "BDD-02",
		},
		{
			nombre: "BDD-02 severidad vacia",
			nuevo: domain.NuevoDefecto{
				Descripcion:       "Error",
				HistoriaID:        10,
				SprintDeteccionID: 20,
			},
			errEsperado: domain.ErrSeveridadDefectoInvalida,
			escenario:   "BDD-02",
		},
		{
			nombre: "BDD-03 descripcion vacia",
			nuevo: domain.NuevoDefecto{
				Descripcion:       "   ",
				Severidad:         domain.SeveridadMinor,
				HistoriaID:        10,
				SprintDeteccionID: 20,
			},
			errEsperado: domain.ErrDescripcionDefectoObligatoria,
			escenario:   "BDD-03",
		},
		{
			nombre: "BDD-04 historia inexistente",
			nuevo: domain.NuevoDefecto{
				Descripcion:       "Error",
				Severidad:         domain.SeveridadMajor,
				HistoriaID:        99,
				SprintDeteccionID: 20,
			},
			preparar: func(p *MockProyectoRepository, h *MockHistoriaRepository, sp *MockSprintRepository, u *MockUsuarioRepository, d *MockDefectoRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyectoActivoDePrueba(1), nil)
				h.On("ObtenerPorID", mock.Anything, uint(1), uint(99)).Return(nil, domain.ErrHistoriaNoEncontrada)
			},
			errEsperado: domain.ErrHistoriaNoEncontrada,
			escenario:   "BDD-04",
		},
		{
			nombre: "BDD-05 sprint de deteccion inexistente",
			nuevo: domain.NuevoDefecto{
				Descripcion:       "Error",
				Severidad:         domain.SeveridadMajor,
				HistoriaID:        10,
				SprintDeteccionID: 99,
			},
			preparar: func(p *MockProyectoRepository, h *MockHistoriaRepository, sp *MockSprintRepository, u *MockUsuarioRepository, d *MockDefectoRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyectoActivoDePrueba(1), nil)
				h.On("ObtenerPorID", mock.Anything, uint(1), uint(10)).Return(historiaBacklogDePrueba(10, 1), nil)
				sp.On("ObtenerPorID", mock.Anything, uint(1), uint(99)).Return(nil, domain.ErrSprintNoEncontrado)
			},
			errEsperado: domain.ErrSprintNoEncontrado,
			escenario:   "BDD-05",
		},
		{
			nombre: "BDD-06 usuario reportante inexistente",
			nuevo: domain.NuevoDefecto{
				Descripcion:       "Error",
				Severidad:         domain.SeveridadMajor,
				HistoriaID:        10,
				SprintDeteccionID: 20,
				ReportadoPor:      func() *uint { v := uint(99); return &v }(),
			},
			preparar: func(p *MockProyectoRepository, h *MockHistoriaRepository, sp *MockSprintRepository, u *MockUsuarioRepository, d *MockDefectoRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyectoActivoDePrueba(1), nil)
				h.On("ObtenerPorID", mock.Anything, uint(1), uint(10)).Return(historiaBacklogDePrueba(10, 1), nil)
				sp.On("ObtenerPorID", mock.Anything, uint(1), uint(20)).Return(sprintActivoDePrueba(20, 1), nil)
				u.On("ObtenerPorID", mock.Anything, uint(99)).Return(nil, domain.ErrUsuarioNoEncontrado)
			},
			errEsperado: domain.ErrUsuarioNoEncontrado,
			escenario:   "BDD-06",
		},
		{
			nombre: "proyecto inexistente",
			nuevo: domain.NuevoDefecto{
				Descripcion:       "Error",
				Severidad:         domain.SeveridadMajor,
				HistoriaID:        10,
				SprintDeteccionID: 20,
			},
			preparar: func(p *MockProyectoRepository, h *MockHistoriaRepository, sp *MockSprintRepository, u *MockUsuarioRepository, d *MockDefectoRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(nil, domain.ErrProyectoNoEncontrado)
			},
			errEsperado: domain.ErrProyectoNoEncontrado,
			escenario:   "BDD-04",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			servicio, p, h, sp, u, d := nuevoServicioDefectos(t)
			if c.preparar != nil {
				c.preparar(p, h, sp, u, d)
			}

			defecto, err := servicio.Crear(context.Background(), 1, c.nuevo)

			if c.errEsperado != nil {
				require.Error(t, err, "escenario BDD %s", c.escenario)
				assert.ErrorIs(t, err, c.errEsperado, "escenario BDD %s", c.escenario)
				assert.Nil(t, defecto)
			} else {
				require.NoError(t, err, "escenario BDD %s", c.escenario)
				require.NotNil(t, defecto)
				if c.verificar != nil {
					c.verificar(t, d)
				}
			}
			p.AssertExpectations(t)
			h.AssertExpectations(t)
			sp.AssertExpectations(t)
			u.AssertExpectations(t)
			d.AssertExpectations(t)
		})
	}
}

func TestActualizarDefecto(t *testing.T) {
	t.Parallel()

	estado := func(e domain.EstadoDefecto) *domain.EstadoDefecto { return &e }
	sprintID := func(v uint) *uint { return &v }

	casos := []struct {
		nombre      string
		defecto     *domain.Defecto
		actualizar  domain.ActualizacionDefecto
		preparar    func(p *MockProyectoRepository, sp *MockSprintRepository, d *MockDefectoRepository)
		errEsperado error
		verificar   func(t *testing.T, d *MockDefectoRepository)
		escenario   string
	}{
		{
			nombre:     "BDD-07 actualizar a en progreso",
			defecto:    defectoDePrueba(1, domain.EstadoDefectoAbierto),
			actualizar: domain.ActualizacionDefecto{Estado: estado(domain.EstadoDefectoEnProgreso)},
			preparar: func(p *MockProyectoRepository, sp *MockSprintRepository, d *MockDefectoRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyectoActivoDePrueba(1), nil)
				d.On("ObtenerPorID", mock.Anything, uint(1), uint(1)).Return(defectoDePrueba(1, domain.EstadoDefectoAbierto), nil)
				d.On("Actualizar", mock.Anything, mock.MatchedBy(func(defecto *domain.Defecto) bool {
					return defecto.Estado == domain.EstadoDefectoEnProgreso && defecto.SprintResolucionID == nil && defecto.ResueltoEn == nil
				})).Return(nil)
			},
			escenario: "BDD-07",
		},
		{
			nombre:     "BDD-08 resolver con sprint de resolucion",
			defecto:    defectoDePrueba(1, domain.EstadoDefectoAbierto),
			actualizar: domain.ActualizacionDefecto{Estado: estado(domain.EstadoDefectoResuelto), SprintResolucionID: sprintID(30)},
			preparar: func(p *MockProyectoRepository, sp *MockSprintRepository, d *MockDefectoRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyectoActivoDePrueba(1), nil)
				d.On("ObtenerPorID", mock.Anything, uint(1), uint(1)).Return(defectoDePrueba(1, domain.EstadoDefectoAbierto), nil)
				sp.On("ObtenerPorID", mock.Anything, uint(1), uint(30)).Return(sprintActivoDePrueba(30, 1), nil)
				d.On("Actualizar", mock.Anything, mock.MatchedBy(func(defecto *domain.Defecto) bool {
					return defecto.Estado == domain.EstadoDefectoResuelto &&
						defecto.SprintResolucionID != nil && *defecto.SprintResolucionID == 30 &&
						defecto.ResueltoEn != nil
				})).Return(nil)
			},
			escenario: "BDD-08",
		},
		{
			nombre:     "BDD-09 cerrar con sprint de resolucion",
			defecto:    defectoDePrueba(1, domain.EstadoDefectoResuelto),
			actualizar: domain.ActualizacionDefecto{Estado: estado(domain.EstadoDefectoCerrado), SprintResolucionID: sprintID(30)},
			preparar: func(p *MockProyectoRepository, sp *MockSprintRepository, d *MockDefectoRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyectoActivoDePrueba(1), nil)
				d.On("ObtenerPorID", mock.Anything, uint(1), uint(1)).Return(defectoDePrueba(1, domain.EstadoDefectoResuelto), nil)
				sp.On("ObtenerPorID", mock.Anything, uint(1), uint(30)).Return(sprintActivoDePrueba(30, 1), nil)
				d.On("Actualizar", mock.Anything, mock.MatchedBy(func(defecto *domain.Defecto) bool {
					return defecto.Estado == domain.EstadoDefectoCerrado && defecto.ResueltoEn != nil
				})).Return(nil)
			},
			escenario: "BDD-09",
		},
		{
			nombre:     "BDD-10 resolver sin sprint de resolucion",
			defecto:    defectoDePrueba(1, domain.EstadoDefectoAbierto),
			actualizar: domain.ActualizacionDefecto{Estado: estado(domain.EstadoDefectoResuelto)},
			preparar: func(p *MockProyectoRepository, sp *MockSprintRepository, d *MockDefectoRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyectoActivoDePrueba(1), nil)
				d.On("ObtenerPorID", mock.Anything, uint(1), uint(1)).Return(defectoDePrueba(1, domain.EstadoDefectoAbierto), nil)
			},
			errEsperado: domain.ErrSprintResolucionRequerido,
			escenario:   "BDD-10",
		},
		{
			nombre:     "BDD-11 sprint de resolucion con estado abierto",
			defecto:    defectoDePrueba(1, domain.EstadoDefectoAbierto),
			actualizar: domain.ActualizacionDefecto{Estado: estado(domain.EstadoDefectoAbierto), SprintResolucionID: sprintID(30)},
			preparar: func(p *MockProyectoRepository, sp *MockSprintRepository, d *MockDefectoRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyectoActivoDePrueba(1), nil)
				d.On("ObtenerPorID", mock.Anything, uint(1), uint(1)).Return(defectoDePrueba(1, domain.EstadoDefectoAbierto), nil)
			},
			errEsperado: domain.ErrSprintResolucionNoPermitido,
			escenario:   "BDD-11",
		},
		{
			nombre:      "BDD-12 estado invalido",
			defecto:     defectoDePrueba(1, domain.EstadoDefectoAbierto),
			actualizar:  domain.ActualizacionDefecto{Estado: estado(domain.EstadoDefecto("fixed"))},
			errEsperado: domain.ErrEstadoDefectoInvalido,
			escenario:   "BDD-12",
		},
		{
			nombre:      "BDD-13 actualizacion vacia",
			defecto:     defectoDePrueba(1, domain.EstadoDefectoAbierto),
			actualizar:  domain.ActualizacionDefecto{},
			errEsperado: domain.ErrActualizacionDefectoVacia,
			escenario:   "BDD-13",
		},
		{
			nombre:     "BDD-14 defecto inexistente",
			defecto:    nil,
			actualizar: domain.ActualizacionDefecto{Estado: estado(domain.EstadoDefectoEnProgreso)},
			preparar: func(p *MockProyectoRepository, sp *MockSprintRepository, d *MockDefectoRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyectoActivoDePrueba(1), nil)
				d.On("ObtenerPorID", mock.Anything, uint(1), uint(1)).Return(nil, domain.ErrDefectoNoEncontrado)
			},
			errEsperado: domain.ErrDefectoNoEncontrado,
			escenario:   "BDD-14",
		},
		{
			nombre: "BDD-15 reabrir limpia sprint y fecha de resolucion",
			defecto: func() *domain.Defecto {
				d := defectoDePrueba(1, domain.EstadoDefectoResuelto)
				id := uint(30)
				ahora := time.Now().Add(-time.Hour)
				d.SprintResolucionID = &id
				d.ResueltoEn = &ahora
				return d
			}(),
			actualizar: domain.ActualizacionDefecto{Estado: estado(domain.EstadoDefectoAbierto)},
			preparar: func(p *MockProyectoRepository, sp *MockSprintRepository, d *MockDefectoRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyectoActivoDePrueba(1), nil)
				resuelto := defectoDePrueba(1, domain.EstadoDefectoResuelto)
				id := uint(30)
				ahora := time.Now().Add(-time.Hour)
				resuelto.SprintResolucionID = &id
				resuelto.ResueltoEn = &ahora
				d.On("ObtenerPorID", mock.Anything, uint(1), uint(1)).Return(resuelto, nil)
				d.On("Actualizar", mock.Anything, mock.MatchedBy(func(defecto *domain.Defecto) bool {
					return defecto.Estado == domain.EstadoDefectoAbierto &&
						defecto.SprintResolucionID == nil &&
						defecto.ResueltoEn == nil
				})).Return(nil)
			},
			escenario: "BDD-15",
		},
		{
			nombre:     "BDD-09 sprint de resolucion inexistente",
			defecto:    defectoDePrueba(1, domain.EstadoDefectoAbierto),
			actualizar: domain.ActualizacionDefecto{Estado: estado(domain.EstadoDefectoResuelto), SprintResolucionID: sprintID(99)},
			preparar: func(p *MockProyectoRepository, sp *MockSprintRepository, d *MockDefectoRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyectoActivoDePrueba(1), nil)
				d.On("ObtenerPorID", mock.Anything, uint(1), uint(1)).Return(defectoDePrueba(1, domain.EstadoDefectoAbierto), nil)
				sp.On("ObtenerPorID", mock.Anything, uint(1), uint(99)).Return(nil, domain.ErrSprintNoEncontrado)
			},
			errEsperado: domain.ErrSprintNoEncontrado,
			escenario:   "BDD-09",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			servicio, p, _, sp, _, d := nuevoServicioDefectos(t)
			if c.preparar != nil {
				c.preparar(p, sp, d)
			}

			defecto, err := servicio.Actualizar(context.Background(), 1, 1, c.actualizar)

			if c.errEsperado != nil {
				require.Error(t, err, "escenario BDD %s", c.escenario)
				assert.ErrorIs(t, err, c.errEsperado, "escenario BDD %s", c.escenario)
				assert.Nil(t, defecto)
			} else {
				require.NoError(t, err, "escenario BDD %s", c.escenario)
				require.NotNil(t, defecto)
				if c.verificar != nil {
					c.verificar(t, d)
				}
			}
			p.AssertExpectations(t)
			sp.AssertExpectations(t)
			d.AssertExpectations(t)
		})
	}
}

func TestListarDefectos(t *testing.T) {
	t.Parallel()

	t.Run("BDD-16 listar defectos del proyecto", func(t *testing.T) {
		t.Parallel()
		servicio, p, _, _, _, d := nuevoServicioDefectos(t)
		p.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyectoActivoDePrueba(1), nil)
		d.On("ListarPorProyecto", mock.Anything, uint(1)).Return([]domain.Defecto{
			*defectoDePrueba(1, domain.EstadoDefectoAbierto),
			*defectoDePrueba(2, domain.EstadoDefectoResuelto),
		}, nil)

		defectos, err := servicio.Listar(context.Background(), 1)

		require.NoError(t, err)
		assert.Len(t, defectos, 2)
		d.AssertExpectations(t)
	})

	t.Run("BDD-17 listado vacio no es nil", func(t *testing.T) {
		t.Parallel()
		servicio, p, _, _, _, d := nuevoServicioDefectos(t)
		p.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyectoActivoDePrueba(1), nil)
		d.On("ListarPorProyecto", mock.Anything, uint(1)).Return(nil, nil)

		defectos, err := servicio.Listar(context.Background(), 1)

		require.NoError(t, err)
		assert.NotNil(t, defectos)
		assert.Empty(t, defectos)
		d.AssertExpectations(t)
	})

	t.Run("proyecto inexistente", func(t *testing.T) {
		t.Parallel()
		servicio, p, _, _, _, _ := nuevoServicioDefectos(t)
		p.On("ObtenerPorID", mock.Anything, uint(1)).Return(nil, domain.ErrProyectoNoEncontrado)

		defectos, err := servicio.Listar(context.Background(), 1)

		assert.ErrorIs(t, err, domain.ErrProyectoNoEncontrado)
		assert.Nil(t, defectos)
		p.AssertExpectations(t)
	})
}

// TestListarDefectosPorSprint cubre el listado por Sprint de US-18.
// Escenarios: docs/features/US-18-defectos-por-sprint.feature (E-01 a E-08).
func TestListarDefectosPorSprint(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre      string
		tipo        domain.TipoListadoDefecto
		sprintID    uint
		preparar    func(p *MockProyectoRepository, sp *MockSprintRepository, d *MockDefectoRepository)
		esperados   int
		errEsperado error
		escenario   string
	}{
		{
			nombre: "E-01 listado por deteccion devuelve solo los del sprint",
			tipo:   domain.TipoListadoDeteccion,
			preparar: func(p *MockProyectoRepository, sp *MockSprintRepository, d *MockDefectoRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyectoActivoDePrueba(1), nil)
				sp.On("ObtenerPorID", mock.Anything, uint(1), uint(1)).Return(sprintActivoDePrueba(1, 1), nil)
				d.On("ListarPorSprint", mock.Anything, uint(1), uint(1), domain.TipoListadoDeteccion).Return([]domain.Defecto{*defectoDePrueba(1, domain.EstadoDefectoAbierto)}, nil)
			},
			esperados: 1,
			escenario: "E-01",
		},
		{
			nombre: "E-02 listado por resolucion devuelve solo los resueltos en el sprint",
			tipo:   domain.TipoListadoResolucion,
			preparar: func(p *MockProyectoRepository, sp *MockSprintRepository, d *MockDefectoRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyectoActivoDePrueba(1), nil)
				sp.On("ObtenerPorID", mock.Anything, uint(1), uint(1)).Return(sprintActivoDePrueba(1, 1), nil)
				d.On("ListarPorSprint", mock.Anything, uint(1), uint(1), domain.TipoListadoResolucion).Return([]domain.Defecto{*defectoDePrueba(2, domain.EstadoDefectoResuelto)}, nil)
			},
			esperados: 1,
			escenario: "E-02",
		},
		{
			nombre: "E-03 sprint sin defectos devuelve slice vacio y no nil",
			tipo:   domain.TipoListadoResolucion,
			preparar: func(p *MockProyectoRepository, sp *MockSprintRepository, d *MockDefectoRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyectoActivoDePrueba(1), nil)
				sp.On("ObtenerPorID", mock.Anything, uint(1), uint(1)).Return(sprintActivoDePrueba(1, 1), nil)
				d.On("ListarPorSprint", mock.Anything, uint(1), uint(1), domain.TipoListadoResolucion).Return(nil, nil)
			},
			esperados: 0,
			escenario: "E-03",
		},
		{
			nombre:      "E-04 tipo vacio se rechaza sin consultar el repositorio",
			tipo:        "",
			preparar:    func(p *MockProyectoRepository, sp *MockSprintRepository, d *MockDefectoRepository) {},
			errEsperado: domain.ErrTipoListadoDefectoInvalido,
			escenario:   "E-04",
		},
		{
			nombre:      "E-05 tipo con mayusculas se rechaza (sensible a mayusculas)",
			tipo:        domain.TipoListadoDefecto("Deteccion"),
			preparar:    func(p *MockProyectoRepository, sp *MockSprintRepository, d *MockDefectoRepository) {},
			errEsperado: domain.ErrTipoListadoDefectoInvalido,
			escenario:   "E-05",
		},
		{
			nombre:   "E-06 sprint inexistente",
			tipo:     domain.TipoListadoDeteccion,
			sprintID: 99,
			preparar: func(p *MockProyectoRepository, sp *MockSprintRepository, d *MockDefectoRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyectoActivoDePrueba(1), nil)
				sp.On("ObtenerPorID", mock.Anything, uint(1), uint(99)).Return(nil, domain.ErrSprintNoEncontrado)
			},
			errEsperado: domain.ErrSprintNoEncontrado,
			escenario:   "E-06",
		},
		{
			nombre:   "E-07 sprint de otro proyecto se trata como inexistente",
			tipo:     domain.TipoListadoDeteccion,
			sprintID: 5,
			preparar: func(p *MockProyectoRepository, sp *MockSprintRepository, d *MockDefectoRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyectoActivoDePrueba(1), nil)
				sp.On("ObtenerPorID", mock.Anything, uint(1), uint(5)).Return(nil, domain.ErrSprintNoEncontrado)
			},
			errEsperado: domain.ErrSprintNoEncontrado,
			escenario:   "E-07",
		},
		{
			nombre: "E-08 proyecto no accesible",
			tipo:   domain.TipoListadoDeteccion,
			preparar: func(p *MockProyectoRepository, sp *MockSprintRepository, d *MockDefectoRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(nil, domain.ErrProyectoNoEncontrado)
			},
			errEsperado: domain.ErrProyectoNoEncontrado,
			escenario:   "E-08",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			servicio, p, _, sp, _, d := nuevoServicioDefectos(t)
			c.preparar(p, sp, d)

			sprintID := c.sprintID
			if sprintID == 0 {
				sprintID = 1
			}
			defectos, err := servicio.ListarPorSprint(context.Background(), 1, sprintID, c.tipo)

			if c.errEsperado != nil {
				assert.ErrorIs(t, err, c.errEsperado, "escenario BDD %s", c.escenario)
				assert.Nil(t, defectos)
			} else {
				require.NoError(t, err, "escenario BDD %s", c.escenario)
				require.NotNil(t, defectos, "escenario BDD %s: el listado nunca debe ser nil", c.escenario)
				assert.Len(t, defectos, c.esperados, "escenario BDD %s", c.escenario)
			}
			p.AssertExpectations(t)
			sp.AssertExpectations(t)
			d.AssertExpectations(t)
		})
	}
}

// TestActualizarDefectoDescripcionYSeveridad cubre la extensión del PATCH de
// US-18: descripción, severidad y rechazo de sprint_deteccion_id.
// Escenarios: docs/features/US-18-defectos-por-sprint.feature (E-09 a E-21).
func TestActualizarDefectoDescripcionYSeveridad(t *testing.T) {
	t.Parallel()

	texto := func(v string) *string { return &v }
	severidad := func(v domain.SeveridadDefecto) *domain.SeveridadDefecto { return &v }
	sprintResolucion := func(v uint) *uint { return &v }

	casos := []struct {
		nombre      string
		defecto     *domain.Defecto
		actualizar  domain.ActualizacionDefecto
		preparar    func(p *MockProyectoRepository, d *MockDefectoRepository)
		errEsperado error
		verificar   func(t *testing.T, defecto *domain.Defecto)
		escenario   string
	}{
		{
			nombre:     "E-09 editar descripcion valida (se normaliza con TrimSpace)",
			defecto:    defectoDePrueba(1, domain.EstadoDefectoAbierto),
			actualizar: domain.ActualizacionDefecto{Descripcion: texto("  Error corregido  ")},
			preparar: func(p *MockProyectoRepository, d *MockDefectoRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyectoActivoDePrueba(1), nil)
				d.On("ObtenerPorID", mock.Anything, uint(1), uint(1)).Return(defectoDePrueba(1, domain.EstadoDefectoAbierto), nil)
				d.On("Actualizar", mock.Anything, mock.MatchedBy(func(defecto *domain.Defecto) bool {
					return defecto.Descripcion == "Error corregido"
				})).Return(nil)
			},
			verificar: func(t *testing.T, defecto *domain.Defecto) {
				assert.Equal(t, "Error corregido", defecto.Descripcion)
			},
			escenario: "E-09",
		},
		{
			nombre:     "E-10 editar severidad con valor de la lista",
			defecto:    defectoDePrueba(1, domain.EstadoDefectoAbierto),
			actualizar: domain.ActualizacionDefecto{Severidad: severidad(domain.SeveridadCritical)},
			preparar: func(p *MockProyectoRepository, d *MockDefectoRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyectoActivoDePrueba(1), nil)
				d.On("ObtenerPorID", mock.Anything, uint(1), uint(1)).Return(defectoDePrueba(1, domain.EstadoDefectoAbierto), nil)
				d.On("Actualizar", mock.Anything, mock.MatchedBy(func(defecto *domain.Defecto) bool {
					return defecto.Severidad == domain.SeveridadCritical
				})).Return(nil)
			},
			verificar: func(t *testing.T, defecto *domain.Defecto) {
				assert.Equal(t, domain.SeveridadCritical, defecto.Severidad)
			},
			escenario: "E-10",
		},
		{
			nombre:     "E-11 descripcion de exactamente 1000 caracteres es valida",
			defecto:    defectoDePrueba(1, domain.EstadoDefectoAbierto),
			actualizar: domain.ActualizacionDefecto{Descripcion: texto(strings.Repeat("a", 1000))},
			preparar: func(p *MockProyectoRepository, d *MockDefectoRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyectoActivoDePrueba(1), nil)
				d.On("ObtenerPorID", mock.Anything, uint(1), uint(1)).Return(defectoDePrueba(1, domain.EstadoDefectoAbierto), nil)
				d.On("Actualizar", mock.Anything, mock.Anything).Return(nil)
			},
			escenario: "E-11",
		},
		{
			nombre:     "E-12 descripcion de 1000 runas multibyte es valida (se cuentan runas, no bytes)",
			defecto:    defectoDePrueba(1, domain.EstadoDefectoAbierto),
			actualizar: domain.ActualizacionDefecto{Descripcion: texto(strings.Repeat("ñ", 1000))},
			preparar: func(p *MockProyectoRepository, d *MockDefectoRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyectoActivoDePrueba(1), nil)
				d.On("ObtenerPorID", mock.Anything, uint(1), uint(1)).Return(defectoDePrueba(1, domain.EstadoDefectoAbierto), nil)
				d.On("Actualizar", mock.Anything, mock.Anything).Return(nil)
			},
			escenario: "E-12",
		},
		{
			nombre:     "E-13 descripcion de 1001 caracteres se rechaza",
			defecto:    defectoDePrueba(1, domain.EstadoDefectoAbierto),
			actualizar: domain.ActualizacionDefecto{Descripcion: texto(strings.Repeat("a", 1001))},
			preparar: func(p *MockProyectoRepository, d *MockDefectoRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyectoActivoDePrueba(1), nil)
			},
			errEsperado: domain.ErrDescripcionDefectoDemasiadoLarga,
			escenario:   "E-13",
		},
		{
			nombre:     "E-14 descripcion formada solo por espacios se rechaza",
			defecto:    defectoDePrueba(1, domain.EstadoDefectoAbierto),
			actualizar: domain.ActualizacionDefecto{Descripcion: texto("   ")},
			preparar: func(p *MockProyectoRepository, d *MockDefectoRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyectoActivoDePrueba(1), nil)
			},
			errEsperado: domain.ErrDescripcionDefectoObligatoria,
			escenario:   "E-14",
		},
		{
			nombre:     "E-15 severidad fuera de la lista (mayusculas) se rechaza",
			defecto:    defectoDePrueba(1, domain.EstadoDefectoAbierto),
			actualizar: domain.ActualizacionDefecto{Severidad: severidad("Major")},
			preparar: func(p *MockProyectoRepository, d *MockDefectoRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyectoActivoDePrueba(1), nil)
			},
			errEsperado: domain.ErrSeveridadDefectoInvalida,
			escenario:   "E-15",
		},
		{
			nombre: "E-16 editar severidad no modifica estado ni sprint de resolucion",
			defecto: func() *domain.Defecto {
				d := defectoDePrueba(1, domain.EstadoDefectoResuelto)
				d.SprintResolucionID = sprintResolucion(2)
				ahora := time.Now().Add(-time.Hour)
				d.ResueltoEn = &ahora
				return d
			}(),
			actualizar: domain.ActualizacionDefecto{Severidad: severidad(domain.SeveridadBlocker)},
			preparar: func(p *MockProyectoRepository, d *MockDefectoRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyectoActivoDePrueba(1), nil)
				resuelto := defectoDePrueba(1, domain.EstadoDefectoResuelto)
				resuelto.SprintResolucionID = sprintResolucion(2)
				ahora := time.Now().Add(-time.Hour)
				resuelto.ResueltoEn = &ahora
				d.On("ObtenerPorID", mock.Anything, uint(1), uint(1)).Return(resuelto, nil)
				d.On("Actualizar", mock.Anything, mock.MatchedBy(func(defecto *domain.Defecto) bool {
					return defecto.Estado == domain.EstadoDefectoResuelto &&
						defecto.SprintResolucionID != nil && *defecto.SprintResolucionID == 2 &&
						defecto.Severidad == domain.SeveridadBlocker
				})).Return(nil)
			},
			escenario: "E-16",
		},
		{
			nombre:     "E-17 sprint_deteccion_id en el body se rechaza aunque incluya otros campos validos",
			defecto:    defectoDePrueba(1, domain.EstadoDefectoAbierto),
			actualizar: domain.ActualizacionDefecto{SprintDeteccionPresente: true, Severidad: severidad(domain.SeveridadMajor)},
			preparar: func(p *MockProyectoRepository, d *MockDefectoRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyectoActivoDePrueba(1), nil)
			},
			errEsperado: domain.ErrSprintDeteccionNoModificable,
			escenario:   "E-17",
		},
		{
			nombre:     "E-21 sprint_deteccion_id con el valor actual tambien se rechaza",
			defecto:    defectoDePrueba(1, domain.EstadoDefectoAbierto),
			actualizar: domain.ActualizacionDefecto{SprintDeteccionPresente: true},
			preparar: func(p *MockProyectoRepository, d *MockDefectoRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyectoActivoDePrueba(1), nil)
			},
			errEsperado: domain.ErrSprintDeteccionNoModificable,
			escenario:   "E-21",
		},
		{
			nombre:     "E-18 defecto inexistente con descripcion",
			defecto:    nil,
			actualizar: domain.ActualizacionDefecto{Descripcion: texto("Intento")},
			preparar: func(p *MockProyectoRepository, d *MockDefectoRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyectoActivoDePrueba(1), nil)
				d.On("ObtenerPorID", mock.Anything, uint(1), uint(1)).Return(nil, domain.ErrDefectoNoEncontrado)
			},
			errEsperado: domain.ErrDefectoNoEncontrado,
			escenario:   "E-18",
		},
		{
			nombre:     "E-19 defecto de otro proyecto se trata como inexistente",
			defecto:    nil,
			actualizar: domain.ActualizacionDefecto{Descripcion: texto("Intento")},
			preparar: func(p *MockProyectoRepository, d *MockDefectoRepository) {
				p.On("ObtenerPorID", mock.Anything, uint(1)).Return(proyectoActivoDePrueba(1), nil)
				d.On("ObtenerPorID", mock.Anything, uint(1), uint(7)).Return(nil, domain.ErrDefectoNoEncontrado)
			},
			errEsperado: domain.ErrDefectoNoEncontrado,
			escenario:   "E-19",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			t.Parallel()
			servicio, p, _, _, _, d := nuevoServicioDefectos(t)
			c.preparar(p, d)

			defectoID := uint(1)
			if c.escenario == "E-19" {
				defectoID = 7
			}
			defecto, err := servicio.Actualizar(context.Background(), 1, defectoID, c.actualizar)

			if c.errEsperado != nil {
				require.Error(t, err, "escenario BDD %s", c.escenario)
				assert.ErrorIs(t, err, c.errEsperado, "escenario BDD %s", c.escenario)
				assert.Nil(t, defecto)
			} else {
				require.NoError(t, err, "escenario BDD %s", c.escenario)
				require.NotNil(t, defecto)
				if c.verificar != nil {
					c.verificar(t, defecto)
				}
			}
			p.AssertExpectations(t)
			d.AssertExpectations(t)
		})
	}
}
