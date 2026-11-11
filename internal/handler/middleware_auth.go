package handler

import (
	"github.com/gin-gonic/gin"
)

// Autenticar stub de la fase RED: no valida el token y deja pasar toda la
// solicitud, por eso el middleware todavia no protege ninguna ruta.
func Autenticar(secreto string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
