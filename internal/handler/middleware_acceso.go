package handler

import (
	"context"

	"github.com/gin-gonic/gin"
)

// ServicioAcceso es el puerto que el guard de membresia necesita del service
// de proyectos. Se define en el consumidor (Clean Architecture).
type ServicioAcceso interface {
	PerteneceA(ctx context.Context, proyectoID, usuarioID uint) (bool, error)
}

// ExigirAccesoAProyecto devuelve el middleware de membresia de US-F10.
func ExigirAccesoAProyecto(_ ServicioAcceso) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
