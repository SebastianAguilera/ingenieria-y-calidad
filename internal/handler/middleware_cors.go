package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	codigoOrigenNoPermitido  = "ORIGEN_NO_PERMITIDO"
	mensajeOrigenNoPermitido = "el origen de la solicitud no está autorizado"
)

// CORS aplica la politica de acceso del navegador. En produccion solo refleja
// el origen exacto configurado en CORS_ALLOWED_ORIGINS; en development, test o
// cualquier otro entorno permite cualquier origen para facilitar el trabajo
// local. La ausencia de Origin nunca bloquea: CORS solo define que puede leer
// el navegador, no autentica la solicitud.
func CORS(appEnv string, allowedOrigins []string) gin.HandlerFunc {
	esProduccion := appEnv == "production"
	permitidos := make(map[string]struct{}, len(allowedOrigins))
	for _, origen := range allowedOrigins {
		permitidos[origen] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		if esProduccion {
			if origin != "" {
				if _, permitido := permitidos[origin]; !permitido {
					c.AbortWithStatusJSON(http.StatusForbidden, respuestaError{
						Error:   codigoOrigenNoPermitido,
						Mensaje: mensajeOrigenNoPermitido,
					})
					return
				}
				c.Header("Access-Control-Allow-Origin", origin)
				c.Writer.Header().Add("Vary", "Origin")
			}
		} else if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Add("Vary", "Origin")
		} else {
			c.Header("Access-Control-Allow-Origin", "*")
		}

		if c.Request.Method == http.MethodOptions {
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept")
			c.Header("Access-Control-Max-Age", "600")
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
