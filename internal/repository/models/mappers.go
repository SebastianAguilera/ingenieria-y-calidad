package models

import "ingenieria-y-calidad/internal/domain"

func ToDomainUsuario(m UsuarioModel) domain.Usuario {
	return domain.Usuario{
		ID:       m.ID,
		Nombre:   m.Nombre,
		Email:    m.Email,
		CreadoEn: m.CreadoEn,
	}
}

func FromDomainUsuario(u domain.Usuario) UsuarioModel {
	return UsuarioModel{
		ID:       u.ID,
		Nombre:   u.Nombre,
		Email:    u.Email,
		CreadoEn: u.CreadoEn,
	}
}

func ToDomainUsuarios(models []UsuarioModel) []domain.Usuario {
	usuarios := make([]domain.Usuario, len(models))
	for idx, m := range models {
		usuarios[idx] = ToDomainUsuario(m)
	}
	return usuarios
}

func FromDomainUsuarios(usuarios []domain.Usuario) []UsuarioModel {
	models := make([]UsuarioModel, len(usuarios))
	for idx, u := range usuarios {
		models[idx] = FromDomainUsuario(u)
	}
	return models
}

func ToDomainProyecto(m ProyectoModel) domain.Proyecto {
	return domain.Proyecto{
		ID:            m.ID,
		Nombre:        m.Nombre,
		FechaInicio:   m.FechaInicio,
		FechaFin:      m.FechaFin,
		Estado:        domain.EstadoProyecto(m.Estado),
		BorradoEn:     m.BorradoEn,
		CreadoEn:      m.CreadoEn,
		ActualizadoEn: m.ActualizadoEn,
		Usuarios:      ToDomainUsuarios(m.Usuarios),
	}
}

func FromDomainProyecto(p domain.Proyecto) ProyectoModel {
	return ProyectoModel{
		ID:            p.ID,
		Nombre:        p.Nombre,
		FechaInicio:   p.FechaInicio,
		FechaFin:      p.FechaFin,
		Estado:        string(p.Estado),
		BorradoEn:     p.BorradoEn,
		CreadoEn:      p.CreadoEn,
		ActualizadoEn: p.ActualizadoEn,
		Usuarios:      FromDomainUsuarios(p.Usuarios),
	}
}

func ToDomainProyectos(models []ProyectoModel) []domain.Proyecto {
	proyectos := make([]domain.Proyecto, len(models))
	for idx, m := range models {
		proyectos[idx] = ToDomainProyecto(m)
	}
	return proyectos
}

func ToDomainHistoria(m HistoriaModel) domain.Historia {
	return domain.Historia{
		ID:            m.ID,
		ProyectoID:    m.ProyectoID,
		Numero:        m.Numero,
		Codigo:        m.Codigo,
		Titulo:        m.Titulo,
		Descripcion:   m.Descripcion,
		Prioridad:     domain.PrioridadHistoria(m.Prioridad),
		Estado:        domain.EstadoHistoria(m.Estado),
		CreadoEn:      m.CreadoEn,
		ActualizadoEn: m.ActualizadoEn,
	}
}

func FromDomainHistoria(h domain.Historia) HistoriaModel {
	return HistoriaModel{
		ID:            h.ID,
		ProyectoID:    h.ProyectoID,
		Numero:        h.Numero,
		Codigo:        h.Codigo,
		Titulo:        h.Titulo,
		Descripcion:   h.Descripcion,
		Prioridad:     string(h.Prioridad),
		Estado:        string(h.Estado),
		CreadoEn:      h.CreadoEn,
		ActualizadoEn: h.ActualizadoEn,
	}
}

func ToDomainHistorias(models []HistoriaModel) []domain.Historia {
	historias := make([]domain.Historia, len(models))
	for idx, m := range models {
		historias[idx] = ToDomainHistoria(m)
	}
	return historias
}

func ToDomainSprint(m SprintModel) domain.Sprint {
	return domain.Sprint{
		ID:            m.ID,
		ProyectoID:    m.ProyectoID,
		Identificador: m.Identificador,
		SprintGoal:    m.SprintGoal,
		Estado:        domain.EstadoSprint(m.Estado),
		FechaInicio:   m.FechaInicio,
		FechaFin:      m.FechaFin,
		CreadoEn:      m.CreadoEn,
		ActualizadoEn: m.ActualizadoEn,
	}
}

func FromDomainSprint(s domain.Sprint) SprintModel {
	return SprintModel{
		ID:            s.ID,
		ProyectoID:    s.ProyectoID,
		Identificador: s.Identificador,
		SprintGoal:    s.SprintGoal,
		Estado:        string(s.Estado),
		FechaInicio:   s.FechaInicio,
		FechaFin:      s.FechaFin,
		CreadoEn:      s.CreadoEn,
		ActualizadoEn: s.ActualizadoEn,
	}
}

func ToDomainSprints(modelos []SprintModel) []domain.Sprint {
	sprints := make([]domain.Sprint, len(modelos))
	for idx, m := range modelos {
		sprints[idx] = ToDomainSprint(m)
	}
	return sprints
}
