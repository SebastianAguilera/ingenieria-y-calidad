package domain

import "errors"

var (
	ErrProyectoNoEncontrado      = errors.New("el proyecto no existe")
	ErrNombreObligatorio         = errors.New("el nombre del proyecto es obligatorio")
	ErrNombreDemasiadoLargo      = errors.New("el nombre del proyecto no puede superar los 150 caracteres")
	ErrDescripcionDemasiadoLarga = errors.New("la descripción del proyecto no puede superar los 2000 caracteres")
	ErrFechaInicioObligatoria    = errors.New("la fecha de inicio del proyecto es obligatoria")
	ErrFechaFinAnteriorAlInicio  = errors.New("la fecha de fin no puede ser anterior a la fecha de inicio")
	ErrFechaFinRequerida         = errors.New("no se puede cerrar un proyecto sin fecha de fin: informá la fecha de fin antes de cerrarlo")
	ErrEstadoInvalido            = errors.New(`el estado del proyecto debe ser "Activo" o "Cerrado"`)
	ErrUsuarioNombreVacio        = errors.New("el nombre del usuario es obligatorio")
	ErrUsuarioNombreLargo        = errors.New("el nombre del usuario no puede superar los 100 caracteres")
	ErrUsuarioEmailInvalido      = errors.New("el email del usuario es obligatorio y debe tener un formato válido")
	ErrUsuarioYaAsociado         = errors.New("el usuario ya forma parte del proyecto")
	ErrUsuarioNoAsociado         = errors.New("el usuario no forma parte del proyecto")
	ErrUsuarioNoEncontrado       = errors.New("el usuario no existe")
	// US-12: alta por email. El mensaje es contrato de la API.
	ErrUsuarioEmailNoRegistrado = errors.New("No existe un usuario registrado con ese email")
	ErrUsuarioYaEsIntegrante    = errors.New("Ese usuario ya es integrante")
	ErrProyectoCerrado          = errors.New("el proyecto está cerrado: solo se permite cambiar su estado o reabrirlo")
	ErrProyectoConHistorial     = errors.New("el proyecto tiene historial asociado y no puede eliminarse")
	ErrErrorDePersistencia      = errors.New("error de persistencia")

	ErrCredencialesInvalidas = errors.New("email o contraseña incorrectos")
	ErrTokenInvalido         = errors.New("el token de autenticación no es válido o ha expirado")
	ErrTokenRequerido        = errors.New("la solicitud no incluye un token de autenticación")
	ErrPasswordVacio         = errors.New("la contraseña es obligatoria")

	ErrHistoriaNoEncontrada         = errors.New("la historia no existe")
	ErrTituloHistoriaObligatorio    = errors.New("el título de la historia es obligatorio")
	ErrTituloHistoriaDemasiadoLargo = errors.New("el título de la historia no puede superar los 200 caracteres")
	ErrPrioridadHistoriaInvalida    = errors.New(`la prioridad de la historia debe ser "Alta", "Media" o "Baja"`)
	ErrEstadoHistoriaInvalido       = errors.New(`el estado de la historia debe ser "Backlog", "En Progreso" o "Hecho"`)
	ErrActualizacionHistoriaVacia   = errors.New("debés informar al menos un campo para modificar la historia")
	ErrCodigoHistoriaDuplicado      = errors.New("ya existe una historia con ese código en el proyecto")
	ErrStoryPointsInvalido          = errors.New("los story points deben ser un número de la escala 0, 0.5, 1, 2, 3, 5, 8, 13 o 21")
	ErrEsfuerzoRealInvalido         = errors.New("el esfuerzo real debe ser un número mayor a 0")

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
	ErrUsuarioDuplicado     = errors.New("el email ya está registrado")
	ErrRolInvalido          = errors.New("el rol especificado no es válido")
	ErrAccesoDenegado       = errors.New("no tenés permisos de líder para realizar esta operación")
	ErrUltimoLider          = errors.New("el proyecto debe conservar al menos un líder")

	ErrHistoriaYaAsignada         = errors.New("la historia ya está asignada a un Sprint")
	ErrSprintNoActivo             = errors.New("el Sprint debe estar activo")
	ErrSprintCerrado              = errors.New("no se pueden modificar Sprints cerrados")
	ErrHistoriaNoPendiente        = errors.New("solo se pueden asignar historias en estado pendiente")
	ErrFueraDeVentanaTemporal     = errors.New("la asignación está fuera de la ventana temporal del Sprint")
	ErrHistoriaNoPerteneceASprint = errors.New("la historia no pertenece al sprint indicado")
	ErrHistoriaYaEnSprintBacklog  = errors.New("la historia ya forma parte del Sprint Backlog")

	ErrSesionPokerActiva       = errors.New("ya existe una sesión de Planning Poker activa para esta historia")
	ErrSesionPokerNoEncontrada = errors.New("la sesión de Planning Poker no existe")
	ErrHistoriaNoEstimable     = errors.New("la historia está en estado Hecho y no se puede estimar")
	ErrSesionPokerNoVotando    = errors.New("la sesión de Planning Poker no está en votación")
	ErrPuntosPokerInvalidos    = errors.New("los puntos deben pertenecer a la escala 0, 0.5, 1, 2, 3, 5, 8, 13 o 21")

	ErrDefectoNoEncontrado           = errors.New("el defecto no existe")
	ErrDescripcionDefectoObligatoria = errors.New("la descripción del defecto es obligatoria")
	ErrSeveridadDefectoInvalida      = errors.New(
		`la severidad del defecto debe ser "minor", "major", "critical" o "blocker"`)
	ErrEstadoDefectoInvalido = errors.New(
		`el estado del defecto debe ser "open", "in_progress", "resolved", "closed" o "rejected"`)

	ErrCriterioNoEncontrado              = errors.New("el criterio de aceptación no existe")
	ErrDescripcionCriterioObligatoria    = errors.New("la descripción del criterio de aceptación es obligatoria")
	ErrDescripcionCriterioDemasiadoLarga = errors.New("la descripción del criterio de aceptación no puede superar los 500 caracteres")
	ErrActualizacionCriterioVacia        = errors.New("debés informar al menos un campo para modificar el criterio de aceptación")
	ErrActualizacionDefectoVacia         = errors.New("debés informar al menos un campo para modificar el defecto")
	ErrDescripcionDefectoDemasiadoLarga  = errors.New("la descripción del defecto no puede superar los 1000 caracteres")
	ErrSprintDeteccionNoModificable      = errors.New("sprint_deteccion_id no es modificable")
	ErrTipoListadoDefectoInvalido        = errors.New("tipo es obligatorio y debe ser deteccion o resolucion")
	ErrSprintResolucionRequerido         = errors.New("debés indicar el Sprint de resolución para resolver o cerrar el defecto")
	ErrSprintResolucionNoPermitido       = errors.New(
		"el Sprint de resolución solo se admite cuando el defecto está resuelto o cerrado")
)
