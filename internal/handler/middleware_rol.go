package handler

import (
	"context"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"

	"ingenieria-y-calidad/internal/domain"
)

// ServicioRoles es el puerto que el guard de permisos necesita del service de
// proyectos. Se define en el consumidor (Clean Architecture).
type ServicioRoles interface {
	EsLider(ctx context.Context, proyectoID, usuarioID uint) (bool, error)
}

// ExigirLider devuelve el guard de permisos de US-11 (RB-03): la operación solo
// la puede realizar un LIDER del proyecto; un MIEMBRO recibe 403
// ACCESO_DENEGADO. Debe registrarse despues de ExigirAccesoAProyecto, que ya
// respondio 404 a quien no es miembro (RB-07).
//
// Con permitirAutoBaja el guard deja pasar a quien se esta quitando a si mismo
// (RB-05): el :usuarioId de la ruta coincide con la identidad del request.
func ExigirLider(servicio ServicioRoles, permitirAutoBaja bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		proyectoID, esValido := parsearParametroProyecto(c)
		if !esValido {
			// Sintaxis del :id: la resuelve el handler con 400.
			c.Next()
			return
		}

		usuarioID, tieneIdentidad := UsuarioDesdeContexto(c)
		if !tieneIdentidad {
			traducirError(c, domain.ErrTokenInvalido)
			return
		}

		if permitirAutoBaja && esElMismoUsuario(c, usuarioID) {
			c.Next()
			return
		}

		esLider, err := servicio.EsLider(c.Request.Context(), proyectoID, usuarioID)
		if err != nil {
			log.Printf("verificando rol de lider del proyecto %d en %s %s: %v",
				proyectoID, c.Request.Method, c.Request.URL.Path, err)
			traducirError(c, domain.ErrErrorDePersistencia)
			return
		}
		if !esLider {
			traducirError(c, domain.ErrAccesoDenegado)
			return
		}

		c.Next()
	}
}

// esElMismoUsuario informa si el :usuarioId de la ruta es el usuario autenticado.
func esElMismoUsuario(c *gin.Context, usuarioID uint) bool {
	objetivo, err := strconv.ParseUint(c.Param("usuarioId"), 10, 32)
	return err == nil && uint(objetivo) == usuarioID
}
