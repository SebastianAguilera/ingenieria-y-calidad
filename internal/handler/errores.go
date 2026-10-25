package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"ingenieria-y-calidad/internal/domain"
)

const (
	codigoJSONInvalido             = "JSON_INVALIDO"
	codigoParametroInvalido        = "PARAMETRO_INVALIDO"
	codigoEstadoInvalido           = "ESTADO_INVALIDO"
	codigoProyectoNoEncontrado     = "PROYECTO_NO_ENCONTRADO"
	codigoUsuarioNoEncontrado      = "USUARIO_NO_ENCONTRADO"
	codigoUsuarioDuplicado         = "USUARIO_DUPLICADO"
	codigoProyectoCerrado          = "PROYECTO_CERRADO"
	codigoProyectoConHistorial     = "PROYECTO_CON_HISTORIAL"
	codigoHistoriaNoEncontrada     = "HISTORIA_NO_ENCONTRADA"
	codigoHistoriaDuplicada        = "HISTORIA_DUPLICADA"
	codigoSprintNoEncontrada       = "SPRINT_NO_ENCONTRADA"
	codigoSprintDuplicado          = "SPRINT_DUPLICADO"
	codigoSprintFechasSolapadas    = "SPRINT_FECHAS_SOLAPADAS"
	codigoSprintActivoExistente      = "SPRINT_ACTIVO_EXISTENTE"
	codigoTransicionSprintInvalida   = "SPRINT_TRANSICION_INVALIDA"
	codigoHistoriaYaAsignada         = "HISTORIA_YA_ASIGNADA"
	codigoSprintNoActivo             = "SPRINT_NO_ACTIVO"
	codigoSprintCerrado              = "SPRINT_CERRADO"
	codigoFueraDeVentanaTemporal     = "FUERA_DE_VENTANA_TEMPORAL"
	codigoHistoriaNoPerteneceASprint = "HISTORIA_NO_PERTENECE_A_SPRINT"
	codigoHistoriaNoPendiente        = "HISTORIA_NO_PENDIENTE"
	codigoValidacion                 = "VALIDACION"
	codigoErrorInterno               = "ERROR_INTERNO"

	mensajeErrorInterno      = "ocurrió un error inesperado al procesar la solicitud"
	mensajeJSONInvalido      = "el cuerpo de la solicitud debe ser un objeto JSON válido"
	mensajeParametroInvalido = "el parámetro de la ruta no es válido"
)

type respuestaError struct {
	Error   string `json:"error"`
	Mensaje string `json:"mensaje"`
	Campo   string `json:"campo,omitempty"`
}

type mapeoError struct {
	codigo     string
	mensaje    string
	campo      string
	estadoHTTP int
}

type contextoCampo struct {
	prefijoUsuario string
}

var (
	contextoRaiz          = contextoCampo{}
	contextoArrayUsuarios = contextoCampo{prefijoUsuario: "usuarios[]."}
)

func (ctx contextoCampo) campoUsuario(nombre string) string {
	return ctx.prefijoUsuario + nombre
}

func traducirError(c *gin.Context, err error) {
	traducirErrorCon(c, err, contextoRaiz)
}

func traducirErrorDeAltaConEquipo(c *gin.Context, err error) {
	traducirErrorCon(c, err, contextoArrayUsuarios)
}

func traducirErrorDeActualizacion(c *gin.Context, err error) {
	traducirErrorCon(c, err, contextoRaiz)
}

func traducirErrorDeCambioEstado(c *gin.Context, err error) {
	traducirErrorCon(c, err, contextoRaiz)
}

func traducirErrorDeEliminacion(c *gin.Context, err error) {
	traducirErrorCon(c, err, contextoRaiz)
}

func traducirErrorDeAsociacion(c *gin.Context, err error) {
	traducirErrorCon(c, err, contextoRaiz)
}

func traducirErrorDeDesvinculacion(c *gin.Context, err error) {
	traducirErrorCon(c, err, contextoRaiz)
}

func traducirErrorCon(c *gin.Context, err error, ctx contextoCampo) {
	mapeo, ok := buscarMapeo(err, ctx)
	if !ok {
		log.Printf("error no controlado en %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		mapeo = mapeoError{
			codigo:     codigoErrorInterno,
			mensaje:    mensajeErrorInterno,
			estadoHTTP: http.StatusInternalServerError,
		}
	}
	c.AbortWithStatusJSON(mapeo.estadoHTTP, respuestaError{
		Error:   mapeo.codigo,
		Mensaje: mapeo.mensaje,
		Campo:   mapeo.campo,
	})
}

func buscarMapeo(err error, ctx contextoCampo) (mapeoError, bool) {
	switch {
	case errors.Is(err, domain.ErrEstadoInvalido):
		return mapeoError{codigoEstadoInvalido, err.Error(), "estado", http.StatusBadRequest}, true
	case errors.Is(err, domain.ErrProyectoNoEncontrado):
		return mapeoError{codigoProyectoNoEncontrado, err.Error(), "", http.StatusNotFound}, true
	case errors.Is(err, domain.ErrUsuarioNoEncontrado), errors.Is(err, domain.ErrUsuarioNoAsociado):
		return mapeoError{codigoUsuarioNoEncontrado, err.Error(), "", http.StatusNotFound}, true
	case errors.Is(err, domain.ErrUsuarioDuplicado):
		return mapeoError{codigoUsuarioDuplicado, err.Error(), "email", http.StatusUnprocessableEntity}, true
	case errors.Is(err, domain.ErrUsuarioYaAsociado):
		return mapeoError{codigoUsuarioDuplicado, err.Error(), "email", http.StatusConflict}, true
	case errors.Is(err, domain.ErrRolInvalido):
		return mapeoError{codigoValidacion, err.Error(), "rol", http.StatusUnprocessableEntity}, true
	case errors.Is(err, domain.ErrProyectoCerrado):
		return mapeoError{codigoProyectoCerrado, err.Error(), "", http.StatusConflict}, true
	case errors.Is(err, domain.ErrProyectoConHistorial):
		return mapeoError{codigoProyectoConHistorial, err.Error(), "", http.StatusConflict}, true
	case errors.Is(err, domain.ErrNombreObligatorio):
		return mapeoError{codigoValidacion, err.Error(), "nombre", http.StatusUnprocessableEntity}, true
	case errors.Is(err, domain.ErrNombreDemasiadoLargo):
		return mapeoError{codigoValidacion, err.Error(), "nombre", http.StatusUnprocessableEntity}, true
	case errors.Is(err, domain.ErrFechaInicioObligatoria):
		return mapeoError{codigoValidacion, err.Error(), "fecha_inicio", http.StatusUnprocessableEntity}, true
	case errors.Is(err, domain.ErrFechaFinAnteriorAlInicio):
		return mapeoError{codigoValidacion, err.Error(), "fecha_fin", http.StatusUnprocessableEntity}, true
	case errors.Is(err, domain.ErrFechaFinRequerida):
		return mapeoError{codigoValidacion, err.Error(), "fecha_fin", http.StatusUnprocessableEntity}, true
	case errors.Is(err, domain.ErrUsuarioNombreVacio),
		errors.Is(err, domain.ErrUsuarioNombreLargo):
		return validacion(err, ctx.campoUsuario("nombre")), true
	case errors.Is(err, domain.ErrUsuarioEmailInvalido):
		return validacion(err, ctx.campoUsuario("email")), true
	case errors.Is(err, domain.ErrHistoriaNoEncontrada):
		return mapeoError{codigoHistoriaNoEncontrada, err.Error(), "", http.StatusNotFound}, true
	case errors.Is(err, domain.ErrCodigoHistoriaDuplicado):
		return mapeoError{codigoHistoriaDuplicada, err.Error(), "codigo", http.StatusConflict}, true
	case errors.Is(err, domain.ErrTituloHistoriaObligatorio),
		errors.Is(err, domain.ErrTituloHistoriaDemasiadoLargo):
		return validacion(err, "titulo"), true
	case errors.Is(err, domain.ErrPrioridadHistoriaInvalida):
		return validacion(err, "prioridad"), true
	case errors.Is(err, domain.ErrEstadoHistoriaInvalido):
		return validacion(err, "estado"), true
	case errors.Is(err, domain.ErrActualizacionHistoriaVacia):
		return validacion(err, ""), true
	case errors.Is(err, domain.ErrSprintNoEncontrado):
		return mapeoError{codigoSprintNoEncontrada, err.Error(), "", http.StatusNotFound}, true
	case errors.Is(err, domain.ErrIdentificadorSprintDuplicado):
		return mapeoError{codigoSprintDuplicado, err.Error(), "identificador", http.StatusConflict}, true
	case errors.Is(err, domain.ErrSprintFechasSolapadas):
		return mapeoError{codigoSprintFechasSolapadas, err.Error(), "", http.StatusConflict}, true
	case errors.Is(err, domain.ErrSprintActivoExistente):
		return mapeoError{codigoSprintActivoExistente, err.Error(), "", http.StatusConflict}, true
	case errors.Is(err, domain.ErrTransicionSprintInvalida):
		return mapeoError{codigoTransicionSprintInvalida, err.Error(), "", http.StatusConflict}, true
	case errors.Is(err, domain.ErrEstadoSprintInvalido):
		return mapeoError{codigoEstadoInvalido, err.Error(), "estado", http.StatusBadRequest}, true
	case errors.Is(err, domain.ErrIdentificadorSprintObligatorio),
		errors.Is(err, domain.ErrIdentificadorSprintDemasiadoLargo):
		return validacion(err, "identificador"), true
	case errors.Is(err, domain.ErrSprintGoalObligatorio),
		errors.Is(err, domain.ErrSprintGoalDemasiadoLargo):
		return validacion(err, "sprint_goal"), true
	case errors.Is(err, domain.ErrSprintFechaInicioObligatoria):
		return validacion(err, "fecha_inicio"), true
	case errors.Is(err, domain.ErrSprintFechaFinAnteriorAlInicio):
		return validacion(err, "fecha_fin"), true
	case errors.Is(err, domain.ErrHistoriaYaAsignada), errors.Is(err, domain.ErrHistoriaYaEnSprintBacklog):
		return mapeoError{codigoHistoriaYaAsignada, err.Error(), "", http.StatusConflict}, true
	case errors.Is(err, domain.ErrSprintNoActivo):
		return mapeoError{codigoSprintNoActivo, err.Error(), "", http.StatusBadRequest}, true
	case errors.Is(err, domain.ErrSprintCerrado):
		return mapeoError{codigoSprintCerrado, err.Error(), "", http.StatusConflict}, true
	case errors.Is(err, domain.ErrHistoriaNoPendiente):
		return mapeoError{codigoHistoriaNoPendiente, err.Error(), "", http.StatusBadRequest}, true
	case errors.Is(err, domain.ErrFueraDeVentanaTemporal):
		return mapeoError{codigoFueraDeVentanaTemporal, err.Error(), "", http.StatusBadRequest}, true
	case errors.Is(err, domain.ErrHistoriaNoPerteneceASprint):
		return mapeoError{codigoHistoriaNoPerteneceASprint, err.Error(), "", http.StatusConflict}, true
	default:
		return mapeoError{}, false
	}
}

func validacion(err error, campo string) mapeoError {
	return mapeoError{codigoValidacion, err.Error(), campo, http.StatusUnprocessableEntity}
}

func responderJSONInvalido(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusBadRequest, respuestaError{
		Error:   codigoJSONInvalido,
		Mensaje: mensajeJSONInvalido,
	})
}

func responderParametroInvalido(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusBadRequest, respuestaError{
		Error:   codigoParametroInvalido,
		Mensaje: mensajeParametroInvalido,
	})
}

func responderFechaInvalida(c *gin.Context, campo string) {
	c.AbortWithStatusJSON(http.StatusBadRequest, respuestaError{
		Error:   codigoParametroInvalido,
		Mensaje: mensajeParametroInvalido,
		Campo:   campo,
	})
}
