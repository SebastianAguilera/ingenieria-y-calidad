package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

func TestUsuarioMappers(t *testing.T) {
	creado := time.Date(2026, 10, 1, 12, 30, 0, 0, time.UTC)
	modelo := UsuarioModel{
		ID:       7,
		Nombre:   "Ada Lovelace",
		Email:    "ada@utn.edu.ar",
		CreadoEn: creado,
	}

	usuario := ToDomainUsuario(modelo)

	assert.Equal(t, uint(7), usuario.ID)
	assert.Equal(t, "Ada Lovelace", usuario.Nombre)
	assert.Equal(t, "ada@utn.edu.ar", usuario.Email)
	assert.Equal(t, creado, usuario.CreadoEn)
	assert.Equal(t, modelo, FromDomainUsuario(usuario))
}

func TestProyectoMappersConEquipo(t *testing.T) {
	inicio := time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC)
	fin := time.Date(2026, 12, 18, 0, 0, 0, 0, time.UTC)
	borrado := time.Date(2026, 12, 20, 8, 0, 0, 0, time.UTC)
	creado := time.Date(2026, 9, 20, 10, 0, 0, 0, time.UTC)
	actualizado := time.Date(2026, 10, 2, 15, 45, 0, 0, time.UTC)

	modelo := ProyectoModel{
		ID:            3,
		Nombre:        "Motor de metricas",
		FechaInicio:   inicio,
		FechaFin:      &fin,
		Estado:        string(domain.EstadoCerrado),
		BorradoEn:     &borrado,
		CreadoEn:      creado,
		ActualizadoEn: actualizado,
		Usuarios: []UsuarioModel{
			{ID: 1, Nombre: "Ada Lovelace", Email: "ada@utn.edu.ar", CreadoEn: creado},
		},
	}

	proyecto := ToDomainProyecto(modelo)

	assert.Equal(t, uint(3), proyecto.ID)
	assert.Equal(t, "Motor de metricas", proyecto.Nombre)
	assert.Equal(t, inicio, proyecto.FechaInicio)
	require.NotNil(t, proyecto.FechaFin)
	assert.Equal(t, fin, *proyecto.FechaFin)
	assert.Equal(t, domain.EstadoCerrado, proyecto.Estado)
	require.NotNil(t, proyecto.BorradoEn)
	assert.Equal(t, borrado, *proyecto.BorradoEn)
	assert.Equal(t, creado, proyecto.CreadoEn)
	assert.Equal(t, actualizado, proyecto.ActualizadoEn)
	require.Len(t, proyecto.Usuarios, 1)
	assert.Equal(t, "Ada Lovelace", proyecto.Usuarios[0].Nombre)

	assert.Equal(t, modelo, FromDomainProyecto(proyecto))
}

func TestProyectoMappersSinFechaFinNiEquipo(t *testing.T) {
	inicio := time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC)
	proyecto := domain.Proyecto{
		ID:          4,
		Nombre:      "Sin fechas",
		FechaInicio: inicio,
		Estado:      domain.EstadoActivo,
		Usuarios:    []domain.Usuario{},
	}

	modelo := FromDomainProyecto(proyecto)

	assert.Nil(t, modelo.FechaFin)
	assert.Nil(t, modelo.BorradoEn)
	assert.Empty(t, modelo.Usuarios)

	vuelta := ToDomainProyecto(modelo)
	assert.Nil(t, vuelta.FechaFin)
	assert.Nil(t, vuelta.BorradoEn)
	assert.Empty(t, vuelta.Usuarios)
	assert.Equal(t, proyecto, vuelta)
}

func TestHistoriaMappersConStoryPoints(t *testing.T) {
	puntos := 8.0
	creado := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	actualizado := time.Date(2026, 10, 2, 9, 0, 0, 0, time.UTC)

	modelo := HistoriaModel{
		ID:            10,
		ProyectoID:    3,
		Numero:        2,
		Codigo:        "US-02",
		Titulo:        "Estimar historias",
		Descripcion:   "Como equipo quiero estimar",
		Prioridad:     string(domain.PrioridadAlta),
		Estado:        string(domain.EstadoHistoriaEnProgreso),
		StoryPoints:   &puntos,
		CreadoEn:      creado,
		ActualizadoEn: actualizado,
	}

	historia := ToDomainHistoria(modelo)

	assert.Equal(t, uint(10), historia.ID)
	assert.Equal(t, uint(3), historia.ProyectoID)
	assert.Equal(t, 2, historia.Numero)
	assert.Equal(t, "US-02", historia.Codigo)
	assert.Equal(t, "Estimar historias", historia.Titulo)
	assert.Equal(t, "Como equipo quiero estimar", historia.Descripcion)
	assert.Equal(t, domain.PrioridadAlta, historia.Prioridad)
	assert.Equal(t, domain.EstadoHistoriaEnProgreso, historia.Estado)
	require.NotNil(t, historia.StoryPoints)
	assert.Equal(t, 8.0, *historia.StoryPoints)
	assert.Equal(t, creado, historia.CreadoEn)
	assert.Equal(t, actualizado, historia.ActualizadoEn)

	assert.Equal(t, modelo, FromDomainHistoria(historia))
}

func TestHistoriaMappersSinStoryPoints(t *testing.T) {
	historia := domain.Historia{
		ID:          11,
		ProyectoID:  3,
		Numero:      3,
		Codigo:      "US-03",
		Titulo:      "Sin estimar",
		Prioridad:   domain.PrioridadBaja,
		Estado:      domain.EstadoHistoriaBacklog,
		StoryPoints: nil,
	}

	modelo := FromDomainHistoria(historia)

	assert.Nil(t, modelo.StoryPoints)
	assert.Equal(t, string(domain.PrioridadBaja), modelo.Prioridad)
	assert.Equal(t, string(domain.EstadoHistoriaBacklog), modelo.Estado)
	assert.Equal(t, historia, ToDomainHistoria(modelo))
}

func TestSprintMappers(t *testing.T) {
	inicio := time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC)
	fin := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
	creado := time.Date(2026, 9, 30, 11, 0, 0, 0, time.UTC)
	actualizado := time.Date(2026, 10, 1, 11, 0, 0, 0, time.UTC)

	modelo := SprintModel{
		ID:            5,
		ProyectoID:    3,
		Identificador: "Sprint 1",
		SprintGoal:    "Cerrar el backlog",
		Estado:        string(domain.EstadoSprintActivo),
		FechaInicio:   inicio,
		FechaFin:      fin,
		CreadoEn:      creado,
		ActualizadoEn: actualizado,
	}

	sprint := ToDomainSprint(modelo)

	assert.Equal(t, uint(5), sprint.ID)
	assert.Equal(t, uint(3), sprint.ProyectoID)
	assert.Equal(t, "Sprint 1", sprint.Identificador)
	assert.Equal(t, "Cerrar el backlog", sprint.SprintGoal)
	assert.Equal(t, domain.EstadoSprintActivo, sprint.Estado)
	assert.Equal(t, inicio, sprint.FechaInicio)
	assert.Equal(t, fin, sprint.FechaFin)
	assert.Equal(t, creado, sprint.CreadoEn)
	assert.Equal(t, actualizado, sprint.ActualizadoEn)

	assert.Equal(t, modelo, FromDomainSprint(sprint))
}

func TestListMappersPreservanOrden(t *testing.T) {
	usuarios := ToDomainUsuarios([]UsuarioModel{{ID: 2}, {ID: 1}})
	require.Len(t, usuarios, 2)
	assert.Equal(t, uint(2), usuarios[0].ID)
	assert.Equal(t, uint(1), usuarios[1].ID)

	modelosUsuarios := FromDomainUsuarios([]domain.Usuario{{ID: 2}, {ID: 1}})
	require.Len(t, modelosUsuarios, 2)
	assert.Equal(t, uint(2), modelosUsuarios[0].ID)
	assert.Equal(t, uint(1), modelosUsuarios[1].ID)

	proyectos := ToDomainProyectos([]ProyectoModel{
		{ID: 2, Usuarios: []UsuarioModel{{ID: 9}}},
		{ID: 1},
	})
	require.Len(t, proyectos, 2)
	assert.Equal(t, uint(2), proyectos[0].ID)
	assert.Equal(t, uint(9), proyectos[0].Usuarios[0].ID)
	assert.Equal(t, uint(1), proyectos[1].ID)

	historias := ToDomainHistorias([]HistoriaModel{{ID: 2}, {ID: 1}})
	require.Len(t, historias, 2)
	assert.Equal(t, uint(2), historias[0].ID)
	assert.Equal(t, uint(1), historias[1].ID)

	sprints := ToDomainSprints([]SprintModel{{ID: 2}, {ID: 1}})
	require.Len(t, sprints, 2)
	assert.Equal(t, uint(2), sprints[0].ID)
	assert.Equal(t, uint(1), sprints[1].ID)
}

func TestListMappersVaciosNoSonNil(t *testing.T) {
	assert.NotNil(t, ToDomainUsuarios(nil))
	assert.Empty(t, ToDomainUsuarios(nil))

	assert.NotNil(t, FromDomainUsuarios(nil))
	assert.Empty(t, FromDomainUsuarios(nil))

	assert.NotNil(t, ToDomainProyectos(nil))
	assert.Empty(t, ToDomainProyectos(nil))

	assert.NotNil(t, ToDomainHistorias(nil))
	assert.Empty(t, ToDomainHistorias(nil))

	assert.NotNil(t, ToDomainSprints(nil))
	assert.Empty(t, ToDomainSprints(nil))
}
