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
	ErrUsuarioNombreVacio       = errors.New("el nombre del usuario es obligatorio")
	ErrUsuarioNombreLargo       = errors.New("el nombre del usuario no puede superar los 100 caracteres")
	ErrUsuarioEmailInvalido     = errors.New("el email del usuario es obligatorio y debe tener un formato válido")
	ErrUsuarioYaAsociado        = errors.New("el usuario ya forma parte del proyecto")
	ErrUsuarioNoAsociado        = errors.New("el usuario no forma parte del proyecto")
	ErrUsuarioNoEncontrado      = errors.New("el usuario no existe")
	ErrProyectoCerrado          = errors.New("el proyecto está cerrado: solo se permite cambiar su estado o reabrirlo")
	ErrProyectoConHistorial     = errors.New("el proyecto tiene historial asociado y no puede eliminarse")
	ErrErrorDePersistencia      = errors.New("error de persistencia")

	ErrHistoriaNoEncontrada         = errors.New("la historia no existe")
	ErrTituloHistoriaObligatorio    = errors.New("el título de la historia es obligatorio")
	ErrTituloHistoriaDemasiadoLargo = errors.New("el título de la historia no puede superar los 200 caracteres")
	ErrPrioridadHistoriaInvalida    = errors.New(`la prioridad de la historia debe ser "Alta", "Media" o "Baja"`)
	ErrEstadoHistoriaInvalido       = errors.New(`el estado de la historia debe ser "Backlog", "En Progreso" o "Hecho"`)
	ErrActualizacionHistoriaVacia   = errors.New("debés informar al menos un campo para modificar la historia")
	ErrCodigoHistoriaDuplicado      = errors.New("ya existe una historia con ese código en el proyecto")

	ErrSprintNoEncontrado                = errors.New("el sprint no existe")
	ErrIdentificadorSprintObligatorio    = errors.New("el identificador del sprint es obligatorio")
	ErrIdentificadorSprintDemasiadoLargo = errors.New(
		"el identificador del sprint no puede superar los 50 caracteres")
	ErrIdentificadorSprintDuplicado   = errors.New("ya existe un sprint con ese identificador en el proyecto")
	ErrSprintGoalObligatorio          = errors.New("el Sprint Goal del sprint es obligatorio")
	ErrSprintGoalDemasiadoLargo       = errors.New("el Sprint Goal del sprint no puede superar los 300 caracteres")
	ErrSprintFechaInicioObligatoria   = errors.New("la fecha de inicio del sprint es obligatoria")
	ErrSprintFechaFinAnteriorAlInicio = errors.New(
		"la fecha de fin del sprint no puede ser anterior a la fecha de inicio")
	ErrSprintFechasSolapadas    = errors.New("las fechas del sprint se solapan con otro sprint del proyecto")
	ErrTransicionSprintInvalida = errors.New("la transición de estado del sprint no está permitida")
	ErrSprintActivoExistente    = errors.New(
		"ya existe un sprint activo en el proyecto: cerrá ese sprint antes de activar otro")
	ErrEstadoSprintInvalido = errors.New(`el estado del sprint debe ser "Planificado", "Activo" o "Cerrado"`)
)
