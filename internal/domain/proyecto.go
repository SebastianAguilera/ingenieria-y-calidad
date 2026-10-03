package domain

import "time"

// EstadoProyecto representa el estado del ciclo de vida de un proyecto.
// Solo admite los valores declarados en las constantes siguientes.
type EstadoProyecto string

const (
	EstadoActivo  EstadoProyecto = "Activo"
	EstadoCerrado EstadoProyecto = "Cerrado"
)

// Valido indica si el estado corresponde a un valor aceptado por el dominio.
func (e EstadoProyecto) Valido() bool {
	return e == EstadoActivo || e == EstadoCerrado
}

// Proyecto es la entidad raíz del dominio.
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

// NuevoProyecto describe el alta de un proyecto con su equipo inicial.
type NuevoProyecto struct {
	Nombre      string
	FechaInicio time.Time
	FechaFin    *time.Time
	Usuarios    []UsuarioInput
}

// ActualizacionProyecto describe la modificación de los datos de un proyecto.
// No incluye Usuarios: la composición se gestiona por separado.
type ActualizacionProyecto struct {
	Nombre      string
	FechaInicio time.Time
	FechaFin    *time.Time
}
