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
	ID            uint           `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Nombre        string         `gorm:"column:nombre;type:varchar(150);not null;index" json:"nombre"`
	FechaInicio   time.Time      `gorm:"column:fecha_inicio;type:date;not null" json:"fecha_inicio"`
	FechaFin      *time.Time     `gorm:"column:fecha_fin;type:date" json:"fecha_fin"`
	Estado        EstadoProyecto `gorm:"column:estado;type:varchar(20);not null;default:'Activo';index" json:"estado"`
	BorradoEn     *time.Time     `gorm:"column:borrado_en;type:timestamptz;index" json:"-"`
	CreadoEn      time.Time      `gorm:"column:creado_en;type:timestamptz;not null;default:now()" json:"creado_en"`
	ActualizadoEn time.Time      `gorm:"column:actualizado_en;type:timestamptz;not null;default:now()" json:"actualizado_en"`

	Integrantes []Integrante `gorm:"many2many:proyecto_integrantes;joinForeignKey:ProyectoID;joinReferences:IntegranteID" json:"integrantes"`
}

// NuevoProyecto describe el alta de un proyecto con su equipo inicial.
type NuevoProyecto struct {
	Nombre      string
	FechaInicio time.Time
	FechaFin    *time.Time
	Integrantes []IntegranteInput
}

// ActualizacionProyecto describe la modificación de los datos de un proyecto.
// No incluye Integrantes: la composición se gestiona por separado.
type ActualizacionProyecto struct {
	Nombre      string
	FechaInicio time.Time
	FechaFin    *time.Time
}
