package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"ingenieria-y-calidad/internal/domain"
)

const (
	codigoJSONInvalido           = "JSON_INVALIDO"
	codigoParametroInvalido      = "PARAMETRO_INVALIDO"
	codigoEstadoInvalido         = "ESTADO_INVALIDO"
	codigoProyectoNoEncontrado   = "PROYECTO_NO_ENCONTRADO"
	codigoIntegranteNoEncontrado = "INTEGRANTE_NO_ENCONTRADO"
	codigoIntegranteDuplicado    = "INTEGRANTE_DUPLICADO"
	codigoProyectoCerrado        = "PROYECTO_CERRADO"
	codigoProyectoConHistorial   = "PROYECTO_CON_HISTORIAL"
	codigoValidacion             = "VALIDACION"
	codigoErrorInterno           = "ERROR_INTERNO"

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

// contextoCampo describe donde viven los campos en el body que el cliente
// envió, para que "campo" en la respuesta de error replique la ruta JSON real.
// El mismo error de negocio necesita nombres distintos segun la ruta: en el alta
// del proyecto el integrante llega dentro del array "integrantes" y se informa
// como "integrantes[].nombre"; en el alta de un integrante suelto el body tiene
// "nombre" en la raiz. Sin este contexto el mapeador no puede distinguirlos.
type contextoCampo struct {
	prefijoIntegrante string
}

var (
	// contextoRaiz es el caso de los endpoints cuyo body tiene los campos del
	// integrante en la raiz.
	contextoRaiz = contextoCampo{}
	// contextoArrayIntegrantes es el caso de POST /api/proyectos, donde el
	// equipo viaja dentro del array "integrantes".
	contextoArrayIntegrantes = contextoCampo{prefijoIntegrante: "integrantes[]."}
)

func (ctx contextoCampo) campoIntegrante(nombre string) string {
	return ctx.prefijoIntegrante + nombre
}

// traducirError convierte un error de dominio en la respuesta HTTP
// estandarizada para los endpoints cuyo body tiene los campos en la raiz.
// Los errores no reconocidos se registran y se responden con un mensaje
// generico que no filtra detalles de infraestructura.
func traducirError(c *gin.Context, err error) {
	traducirErrorCon(c, err, contextoRaiz)
}

// traducirErrorDeAltaConEquipo es la variante para POST /api/proyectos, donde
// los integrantes llegan anidados en el array "integrantes".
func traducirErrorDeAltaConEquipo(c *gin.Context, err error) {
	traducirErrorCon(c, err, contextoArrayIntegrantes)
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
	case errors.Is(err, domain.ErrIntegranteNoEncontrado), errors.Is(err, domain.ErrIntegranteNoAsociado):
		return mapeoError{codigoIntegranteNoEncontrado, err.Error(), "", http.StatusNotFound}, true
	case errors.Is(err, domain.ErrIntegranteYaAsociado):
		// El BDD exige "email" con y sin anidamiento (E-17 y E-44), porque el
		// conflicto siempre se describe por el email duplicado y no por su
		// ubicacion en el body.
		return mapeoError{codigoIntegranteDuplicado, err.Error(), "email", http.StatusConflict}, true
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
	case errors.Is(err, domain.ErrIntegranteNombreVacio),
		errors.Is(err, domain.ErrIntegranteNombreLargo):
		return validacion(err, ctx.campoIntegrante("nombre")), true
	case errors.Is(err, domain.ErrIntegranteEmailInvalido):
		return validacion(err, ctx.campoIntegrante("email")), true
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

func responderParametroInvalido(c *gin.Context, campo string) {
	c.AbortWithStatusJSON(http.StatusBadRequest, respuestaError{
		Error:   codigoParametroInvalido,
		Mensaje: mensajeParametroInvalido,
		Campo:   campo,
	})
}

func responderFechaInvalida(c *gin.Context, campo string) {
	c.AbortWithStatusJSON(http.StatusBadRequest, respuestaError{
		Error:   codigoParametroInvalido,
		Mensaje: mensajeParametroInvalido,
		Campo:   campo,
	})
}
