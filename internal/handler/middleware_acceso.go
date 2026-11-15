package handler

import (
	"context"
	"log"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"ingenieria-y-calidad/internal/domain"
)

// ServicioAcceso es el puerto que el guard de membresia necesita del service
// de proyectos. Se define en el consumidor (Clean Architecture).
type ServicioAcceso interface {
	PerteneceA(ctx context.Context, proyectoID, usuarioID uint) (bool, error)
}

// ExigirAccesoAProyecto devuelve el guard de membresia de US-F10 (RB-04,
// RB-05, RB-08): toda ruta registrada bajo /api/proyectos/:id exige que el
// usuario autenticado sea miembro del proyecto. Quien no lo es recibe el mismo
// 404 que un proyecto inexistente, para no revelar la existencia del recurso.
// Las rutas sin :id de proyecto (listado, alta, /api/usuarios, /health) quedan
// fuera del guard.
func ExigirAccesoAProyecto(servicio ServicioAcceso) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !esRutaDeProyecto(c) {
			c.Next()
			return
		}

		proyectoID, esValido := parsearParametroProyecto(c)
		if !esValido {
			// Sintaxis del :id: la resuelve el handler con 400 (RB-06 / E-12).
			c.Next()
			return
		}

		usuarioID, tieneIdentidad := UsuarioDesdeContexto(c)
		if !tieneIdentidad {
			// Defense en profundidad: Autenticar ya garantiza el sub, pero sin
			// identidad no se puede decidir (CA-1).
			traducirError(c, domain.ErrTokenInvalido)
			return
		}

		pertenece, err := servicio.PerteneceA(c.Request.Context(), proyectoID, usuarioID)
		if err != nil {
			log.Printf("verificando membresia del proyecto %d en %s %s: %v",
				proyectoID, c.Request.Method, c.Request.URL.Path, err)
			traducirError(c, domain.ErrErrorDePersistencia)
			return
		}
		if !pertenece {
			traducirError(c, domain.ErrProyectoNoEncontrado)
			return
		}

		c.Next()
	}
}

// esRutaDeProyecto informa si la ruta registrada (FullPath, no la pedida)
// pertenece al arbol de /api/proyectos/:id.
func esRutaDeProyecto(c *gin.Context) bool {
	return strings.HasPrefix(c.FullPath(), "/api/proyectos/:id")
}

// parsearParametroProyecto traduce el :id del proyecto a un uint. Un valor no
// numerico o cero no se considera id de proyecto: el handler responde 400.
func parsearParametroProyecto(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		return 0, false
	}
	return uint(id), true
}
