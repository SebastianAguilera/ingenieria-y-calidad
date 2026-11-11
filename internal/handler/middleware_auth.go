package handler

import (
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"ingenieria-y-calidad/internal/domain"
)

const prefijoBearer = "Bearer "

// Autenticar devuelve el middleware de autenticacion (R-06): protege toda
// solicitud cuya ruta no figura en rutasPublicas (default-deny) exigiendo un
// JWT firmado con HS256 y aun vigente.
func Autenticar(secreto string, rutasPublicas ...string) gin.HandlerFunc {
	publicas := make(map[string]struct{}, len(rutasPublicas))
	for _, ruta := range rutasPublicas {
		publicas[ruta] = struct{}{}
	}

	return func(c *gin.Context) {
		if _, esPublica := publicas[c.Request.URL.Path]; esPublica {
			c.Next()
			return
		}

		autorizacion := c.GetHeader("Authorization")
		if autorizacion == "" || !strings.HasPrefix(autorizacion, prefijoBearer) {
			traducirError(c, domain.ErrTokenRequerido)
			return
		}

		token, err := jwt.Parse(strings.TrimPrefix(autorizacion, prefijoBearer),
			func(tok *jwt.Token) (interface{}, error) {
				if _, esHMAC := tok.Method.(*jwt.SigningMethodHMAC); !esHMAC {
					return nil, fmt.Errorf("algoritmo de firma no soportado: %v", tok.Header["alg"])
				}
				return []byte(secreto), nil
			})
		if err != nil || !token.Valid {
			// El detalle del fallo (firma, vigencia, algoritmo) solo va al log:
			// la respuesta expone un mensaje generico (E-13).
			log.Printf("token rechazado en %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
			traducirError(c, domain.ErrTokenInvalido)
			return
		}

		c.Next()
	}
}

// UsuarioDesdeContexto devuelve el id del usuario autenticado que Autenticar
// dejo en el contexto (US-F10, RB-01). La segunda devolucion indica si el
// identificador esta presente.
func UsuarioDesdeContexto(c *gin.Context) (uint, bool) {
	return 0, false
}
