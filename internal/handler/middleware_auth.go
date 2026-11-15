package handler

import (
	"fmt"
	"log"
	"strconv"
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

		identidad, esValida := identidadDesdeClaims(token.Claims)
		if !esValida {
			// Token firmado y vigente pero sin un sub numerico usable: no hay
			// identidad para las capas siguientes (US-F10, RB-01).
			log.Printf("token sin identidad valida en %s %s", c.Request.Method, c.Request.URL.Path)
			traducirError(c, domain.ErrTokenInvalido)
			return
		}
		c.Set(claveContextoUsuario, identidad)

		c.Next()
	}
}

// claveContextoUsuario es la clave con la que Autenticar guarda el id del
// usuario en el contexto de Gin (los Set de gin solo aceptan string).
const claveContextoUsuario = "usuario_id"

// identidadDesdeClaims traduce el claim "sub" al id del usuario. Acepta el
// formato de texto que emite el login ("7") y un numero JSON entero (7); el
// resto de las formas no identifica a nadie (US-F10, RB-01).
func identidadDesdeClaims(claims jwt.Claims) (uint, bool) {
	claimsDeMapa, esMapa := claims.(jwt.MapClaims)
	if !esMapa {
		return 0, false
	}

	sub, existe := claimsDeMapa["sub"]
	if !existe || sub == nil {
		return 0, false
	}

	texto, esTexto := sub.(string)
	if !esTexto {
		numero, esNumero := sub.(float64)
		if !esNumero {
			return 0, false
		}
		texto = strconv.FormatFloat(numero, 'f', -1, 64)
	}

	// 32 bits: los ids de usuario no superan 4294967295 y asi cualquier valor
	// fuera de rango se rechaza en lugar de truncarse.
	id, err := strconv.ParseUint(texto, 10, 32)
	if err != nil {
		return 0, false
	}
	return uint(id), true
}

// UsuarioDesdeContexto devuelve el id del usuario autenticado que Autenticar
// dejo en el contexto (US-F10, RB-01). La segunda devolucion indica si el
// identificador esta presente.
func UsuarioDesdeContexto(c *gin.Context) (uint, bool) {
	usuarioID, existe := c.Get(claveContextoUsuario)
	if !existe {
		return 0, false
	}
	id, esNumero := usuarioID.(uint)
	if !esNumero {
		return 0, false
	}
	return id, true
}
