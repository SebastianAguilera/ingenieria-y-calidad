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
