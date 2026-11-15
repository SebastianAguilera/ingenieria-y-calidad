package handler

import (
	"crypto/subtle"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

const (
	// cabeceraFrontendKey es la cabecera que agrega el proxy de Vercel (D9).
	cabeceraFrontendKey = "X-Frontend-Key"

	codigoFrontendKeyInvalida  = "FRONTEND_KEY_INVALIDA"
	mensajeFrontendKeyInvalida = "la solicitud no proviene del frontend autorizado"
)

// ExigirFrontendKey devuelve el middleware que solo acepta solicitudes con la
// clave compartida entre el proxy de Vercel y la API (D5). Si FRONTEND_KEY no
// esta definida, el middleware no bloquea: asi el desarrollo y los tests no
// necesitan configurarla. En produccion config.Load exige la clave, por lo que
// el arranque falla antes de llegar aca si falta (fail-closed).
//
// La comparacion usa subtle.ConstantTimeCompare para no filtrar la clave por
// diferencias de tiempo.
func ExigirFrontendKey() gin.HandlerFunc {
	claveEsperada := os.Getenv("FRONTEND_KEY")
	if claveEsperada == "" {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	claveEsperadaBytes := []byte(claveEsperada)

	return func(c *gin.Context) {
		recibida := c.GetHeader(cabeceraFrontendKey)
		if subtle.ConstantTimeCompare([]byte(recibida), claveEsperadaBytes) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, respuestaError{
				Error:   codigoFrontendKeyInvalida,
				Mensaje: mensajeFrontendKeyInvalida,
			})
			return
		}

		c.Next()
	}
}
