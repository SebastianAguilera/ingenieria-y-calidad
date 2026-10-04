package main

import (
	"log"

	"ingenieria-y-calidad/internal/config"
	"ingenieria-y-calidad/internal/handler"
	"ingenieria-y-calidad/internal/repository"
	"ingenieria-y-calidad/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error cargando la configuración: %v", err)
	}

	db, err := repository.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Error conectando a la base de datos (%s / db: %s): %v", cfg.AppEnv, cfg.DBName, err)
	}
	log.Printf("✅ Conexión a la base de datos establecida exitosamente (Entorno: %s, DB: %s)", cfg.AppEnv, cfg.DBName)

	if err := repository.MigrarEsquema(db); err != nil {
		log.Fatalf("Error creando el esquema de la base de datos: %v", err)
	}
	log.Println("Esquema de la base de datos verificado")

	proyectoRepo := repository.NuevoProyectoRepository(db)
	usuarioRepo := repository.NuevoUsuarioRepository(db)
	proyectoService := service.NewProyectoService(proyectoRepo, usuarioRepo)

	r := setupRouter(handler.ServicioProyectos(proyectoService))

	log.Printf("🚀 Servidor iniciado en el puerto %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}
