package domain

import "time"

type EstadoProyecto string

const (
	EstadoActivo  EstadoProyecto = "Activo"
	EstadoCerrado EstadoProyecto = "Cerrado"
)

func (e EstadoProyecto) Valido() bool {
	return e == EstadoActivo || e == EstadoCerrado
}

type Proyecto struct {
	ID            uint           `json:"id"`
	Nombre        string         `json:"nombre"`
	FechaInicio   time.Time      `json:"fecha_inicio"`
	FechaFin      *time.Time     `json:"fecha_fin"`
	Estado        EstadoProyecto `json:"estado"`
	BorradoEn     *time.Time     `json:"-"`
	CreadoEn      time.Time      `json:"creado_en"`
	ActualizadoEn time.Time      `json:"actualizado_en"`

	Usuarios []Usuario `json:"usuarios"`
}

type NuevoProyecto struct {
	Nombre      string
	FechaInicio time.Time
	FechaFin    *time.Time
	Usuarios    []UsuarioInput
}

type ActualizacionProyecto struct {
	Nombre      string
	FechaInicio time.Time
	FechaFin    *time.Time
}
