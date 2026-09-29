package domain

import "errors"

var (
	ErrProyectoNoEncontrado     = errors.New("el proyecto no existe")
	ErrNombreObligatorio        = errors.New("el nombre del proyecto es obligatorio")
	ErrNombreDemasiadoLargo     = errors.New("el nombre del proyecto no puede superar los 150 caracteres")
	ErrFechaInicioObligatoria   = errors.New("la fecha de inicio del proyecto es obligatoria")
	ErrFechaFinAnteriorAlInicio = errors.New("la fecha de fin no puede ser anterior a la fecha de inicio")
	ErrFechaFinRequerida        = errors.New("no se puede cerrar un proyecto sin fecha de fin: informá la fecha de fin antes de cerrarlo")
	ErrEstadoInvalido           = errors.New(`el estado del proyecto debe ser "Activo" o "Cerrado"`)
	ErrIntegranteNombreVacio    = errors.New("el nombre del integrante es obligatorio")
	ErrIntegranteNombreLargo    = errors.New("el nombre del integrante no puede superar los 100 caracteres")
	ErrIntegranteEmailInvalido  = errors.New("el email del integrante es obligatorio y debe tener un formato válido")
	ErrIntegranteYaAsociado     = errors.New("el integrante ya forma parte del proyecto")
	ErrIntegranteNoAsociado     = errors.New("el integrante no forma parte del proyecto")
	ErrIntegranteNoEncontrado   = errors.New("el integrante no existe")
	ErrProyectoCerrado          = errors.New("el proyecto está cerrado: solo se permite cambiar su estado o reabrirlo")
	ErrProyectoConHistorial     = errors.New("el proyecto tiene historial asociado y no puede eliminarse")
	ErrErrorDePersistencia      = errors.New("error de persistencia")
)
