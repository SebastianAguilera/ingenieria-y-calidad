package domain

import (
	"fmt"
	"time"
)

// PrioridadHistoria representa la prioridad de una historia de usuario.
// Solo admite los valores declarados en las constantes siguientes.
type PrioridadHistoria string

const (
	PrioridadAlta  PrioridadHistoria = "Alta"
	PrioridadMedia PrioridadHistoria = "Media"
	PrioridadBaja  PrioridadHistoria = "Baja"
)

// Valido indica si la prioridad corresponde a un valor aceptado por el dominio.
func (p PrioridadHistoria) Valido() bool {
	return p == PrioridadAlta || p == PrioridadMedia || p == PrioridadBaja
}

// EstadoHistoria representa el estado de una historia dentro del backlog.
// Solo admite los valores declarados en las constantes siguientes.
type EstadoHistoria string

const (
	EstadoHistoriaBacklog    EstadoHistoria = "Backlog"
	EstadoHistoriaEnProgreso EstadoHistoria = "En Progreso"
	EstadoHistoriaHecho      EstadoHistoria = "Hecho"
)

// Valido indica si el estado corresponde a un valor aceptado por el dominio.
func (e EstadoHistoria) Valido() bool {
	return e == EstadoHistoriaBacklog || e == EstadoHistoriaEnProgreso || e == EstadoHistoriaHecho
}

// Historia es una historia de usuario del Product Backlog de un proyecto.
type Historia struct {
	ID            uint
	ProyectoID    uint
	Numero        int
	Codigo        string
	Titulo        string
	Descripcion   string
	Prioridad     PrioridadHistoria
	Estado        EstadoHistoria
	StoryPoints   *int
	CreadoEn      time.Time
	ActualizadoEn time.Time
}

// NuevaHistoria describe el alta de una historia en el backlog. La prioridad y
// el estado son opcionales: si no se informan toman los valores por defecto.
type NuevaHistoria struct {
	Titulo      string
	Descripcion string
	Prioridad   PrioridadHistoria
	Estado      EstadoHistoria
	StoryPoints *int
}

// ActualizacionHistoria describe la modificacion parcial de una historia. Cada
// campo nil significa "no informado" y conserva el valor persistido.
type ActualizacionHistoria struct {
	Titulo      *string
	Descripcion *string
	Prioridad   *PrioridadHistoria
	Estado      *EstadoHistoria
}

// FormatearCodigoHistoria construye el codigo visible "US-NN" a partir del
// numero consecutivo de la historia dentro de su proyecto.
func FormatearCodigoHistoria(numero int) string {
	return fmt.Sprintf("US-%02d", numero)
}
