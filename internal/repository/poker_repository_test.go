package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"ingenieria-y-calidad/internal/domain"
)

// prepararPoker siembra un proyecto con dos miembros y una historia para los
// tests de persistencia de US-14.
func prepararPoker(t *testing.T, db *gorm.DB) (proyectoID, historiaID, anaID, benID uint) {
	t.Helper()
	ctx := context.Background()

	proyectos := NuevoProyectoRepository(db)
	usuarios := NuevoUsuarioRepository(db)
	historias := NuevoHistoriaRepository(db)

	ana := &domain.Usuario{Nombre: "Ana", Email: "ana.poker@utn.edu.ar"}
	ben := &domain.Usuario{Nombre: "Ben", Email: "ben.poker@utn.edu.ar"}
	require.NoError(t, usuarios.Crear(ctx, ana))
	require.NoError(t, usuarios.Crear(ctx, ben))

	inicio := time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC)
	proyecto := &domain.Proyecto{Nombre: "Poker", FechaInicio: inicio, Estado: domain.EstadoActivo}
	require.NoError(t, proyectos.CrearConUsuarios(ctx, proyecto, nil))
	require.NoError(t, usuarios.Vincular(ctx, proyecto.ID, ana.ID, domain.RolLider))
	require.NoError(t, usuarios.Vincular(ctx, proyecto.ID, ben.ID, domain.RolMiembro))

	historia := &domain.Historia{
		ProyectoID: proyecto.ID,
		Numero:     1,
		Codigo:     "US-01",
		Titulo:     "Estimar en equipo",
		Prioridad:  domain.PrioridadMedia,
		Estado:     domain.EstadoHistoriaBacklog,
	}
	require.NoError(t, historias.Crear(ctx, historia))

	return proyecto.ID, historia.ID, ana.ID, ben.ID
}

func TestCrearSesionPersisteEstadoYRondaUno(t *testing.T) {
	db := abrirBaseDePruebas(t)
	repo := NuevoPokerRepository(db)
	proyectoID, historiaID, anaID, _ := prepararPoker(t, db)
	ctx := context.Background()

	sesion := &domain.SesionPoker{ProyectoID: proyectoID, HistoriaID: historiaID, CreadoPor: anaID}
	require.NoError(t, repo.CrearSesion(ctx, sesion))

	assert.NotZero(t, sesion.ID)
	assert.Equal(t, domain.EstadoSesionVotando, sesion.Estado)
	assert.Equal(t, 1, sesion.RondaActual)
	assert.False(t, sesion.CreadoEn.IsZero())

	ronda, err := repo.ObtenerRonda(ctx, sesion.ID, 1)
	require.NoError(t, err)
	assert.Equal(t, 1, ronda.NumeroRonda)
	assert.Equal(t, sesion.ID, ronda.SesionID)
}

func TestIndiceParcialBloqueaSesionActivaYPermiteTrasCerrar(t *testing.T) {
	db := abrirBaseDePruebas(t)
	repo := NuevoPokerRepository(db)
	proyectoID, historiaID, anaID, _ := prepararPoker(t, db)
	ctx := context.Background()

	primera := &domain.SesionPoker{ProyectoID: proyectoID, HistoriaID: historiaID, CreadoPor: anaID}
	require.NoError(t, repo.CrearSesion(ctx, primera))

	segunda := &domain.SesionPoker{ProyectoID: proyectoID, HistoriaID: historiaID, CreadoPor: anaID}
	err := repo.CrearSesion(ctx, segunda)
	assert.ErrorIs(t, err, domain.ErrSesionPokerActiva)

	var total int64
	require.NoError(t, db.Table("poker_sessions").
		Where("historia_id = ?", historiaID).
		Count(&total).Error)
	assert.Equal(t, int64(1), total, "la creación rechazada no debe persistir nada")

	require.NoError(t, db.Exec("UPDATE poker_sessions SET estado = 'Cerrada' WHERE id = ?", primera.ID).Error)

	tercera := &domain.SesionPoker{ProyectoID: proyectoID, HistoriaID: historiaID, CreadoPor: anaID}
	require.NoError(t, repo.CrearSesion(ctx, tercera))
	assert.NotZero(t, tercera.ID)
}

func TestDetalleActivasYVotos(t *testing.T) {
	db := abrirBaseDePruebas(t)
	repo := NuevoPokerRepository(db)
	proyectoID, historiaID, anaID, benID := prepararPoker(t, db)
	ctx := context.Background()

	sesion := &domain.SesionPoker{ProyectoID: proyectoID, HistoriaID: historiaID, CreadoPor: anaID}
	require.NoError(t, repo.CrearSesion(ctx, sesion))

	ronda, err := repo.ObtenerRonda(ctx, sesion.ID, 1)
	require.NoError(t, err)
	require.NoError(t, db.Exec(
		"INSERT INTO poker_votes (ronda_id, usuario_id, puntos) VALUES (?, ?, ?), (?, ?, ?)",
		ronda.ID, anaID, 3.0, ronda.ID, benID, 5.0).Error)

	obtenida, err := repo.ObtenerPorID(ctx, proyectoID, sesion.ID)
	require.NoError(t, err)
	assert.Equal(t, "Ana", obtenida.CreadorNombre)

	totalMiembros, err := repo.ContarMiembros(ctx, proyectoID)
	require.NoError(t, err)
	assert.Equal(t, 2, totalMiembros)

	totalVotos, err := repo.ContarVotosPorRonda(ctx, ronda.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, totalVotos)

	votos, err := repo.ListarVotosPorRonda(ctx, ronda.ID)
	require.NoError(t, err)
	require.Len(t, votos, 2)
	assert.Equal(t, "Ana", votos[0].Nombre)
	assert.Equal(t, 3.0, votos[0].Puntos)
	assert.Equal(t, "Ben", votos[1].Nombre)

	votoBen, err := repo.ObtenerVotoUsuario(ctx, ronda.ID, benID)
	require.NoError(t, err)
	require.NotNil(t, votoBen)
	assert.Equal(t, 5.0, votoBen.Puntos)

	sinVoto, err := repo.ObtenerVotoUsuario(ctx, ronda.ID, anaID+benID+1000)
	require.NoError(t, err)
	assert.Nil(t, sinVoto)

	activa, err := repo.ObtenerActivaPorHistoria(ctx, historiaID)
	require.NoError(t, err)
	require.NotNil(t, activa)
	assert.Equal(t, sesion.ID, activa.ID)

	activas, err := repo.ListarActivasPorProyecto(ctx, proyectoID)
	require.NoError(t, err)
	require.Len(t, activas, 1)

	require.NoError(t, db.Exec("UPDATE poker_sessions SET estado = 'Revelada' WHERE id = ?", sesion.ID).Error)
	activas, err = repo.ListarActivasPorProyecto(ctx, proyectoID)
	require.NoError(t, err)
	assert.Len(t, activas, 1, "Revelada sigue siendo activa")

	require.NoError(t, db.Exec("UPDATE poker_sessions SET estado = 'Cancelada' WHERE id = ?", sesion.ID).Error)
	activas, err = repo.ListarActivasPorProyecto(ctx, proyectoID)
	require.NoError(t, err)
	assert.Empty(t, activas)

	historial, err := repo.ListarPorHistoria(ctx, historiaID)
	require.NoError(t, err)
	require.Len(t, historial, 1, "el historial incluye Cancelada")
}

func TestObtenerSesionDeOtroProyectoNoExiste(t *testing.T) {
	db := abrirBaseDePruebas(t)
	repo := NuevoPokerRepository(db)
	proyectoID, historiaID, anaID, _ := prepararPoker(t, db)
	ctx := context.Background()

	sesion := &domain.SesionPoker{ProyectoID: proyectoID, HistoriaID: historiaID, CreadoPor: anaID}
	require.NoError(t, repo.CrearSesion(ctx, sesion))

	_, err := repo.ObtenerPorID(ctx, proyectoID+1, sesion.ID)
	assert.ErrorIs(t, err, domain.ErrSesionPokerNoEncontrada)
}
