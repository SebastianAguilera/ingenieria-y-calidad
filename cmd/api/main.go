package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"ingenieria-y-calidad/internal/config"
	"ingenieria-y-calidad/internal/repository"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error cargando la configuración: %v", err)
	}

	_, err = repository.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Error conectando a la base de datos (%s / db: %s): %v", cfg.AppEnv, cfg.DBName, err)
	}
	log.Printf("✅ Conexión a la base de datos establecida exitosamente (Entorno: %s, DB: %s)", cfg.AppEnv, cfg.DBName)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Software Metrics & Estimation Engine API running",
			"env":     cfg.AppEnv,
		})
	})

	log.Printf("🚀 Servidor iniciado en el puerto %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}
