package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"ingenieria-y-calidad/internal/domain"
)

const (
	codigoJSONInvalido         = "JSON_INVALIDO"
	codigoParametroInvalido    = "PARAMETRO_INVALIDO"
	codigoEstadoInvalido       = "ESTADO_INVALIDO"
	codigoProyectoNoEncontrado = "PROYECTO_NO_ENCONTRADO"
	codigoUsuarioNoEncontrado  = "USUARIO_NO_ENCONTRADO"
	codigoUsuarioDuplicado     = "USUARIO_DUPLICADO"
	codigoProyectoCerrado      = "PROYECTO_CERRADO"
	codigoProyectoConHistorial = "PROYECTO_CON_HISTORIAL"
	codigoValidacion           = "VALIDACION"
	codigoErrorInterno         = "ERROR_INTERNO"

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
	case errors.Is(err, domain.ErrUsuarioYaAsociado):
		return mapeoError{codigoUsuarioDuplicado, err.Error(), "email", http.StatusConflict}, true
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
