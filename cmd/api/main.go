package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"ingenieria-y-calidad/internal/handler"
	"ingenieria-y-calidad/internal/repository"
	"ingenieria-y-calidad/internal/service"
)

const puertoPorDefecto = "8080"

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system environment")
	}

	db, err := repository.ConnectDB()
	if err != nil {
		log.Fatalf("Error conectando a la base de datos: %v", err)
	}
	log.Println("Conexión a la base de datos establecida")

	if err := repository.MigrarEsquema(db); err != nil {
		log.Fatalf("Error creando el esquema de la base de datos: %v", err)
	}
	log.Println("Esquema de la base de datos verificado")

	proyectoRepo := repository.NuevoProyectoRepository(db)
	integranteRepo := repository.NuevoIntegranteRepository(db)
	proyectoService := service.NewProyectoService(proyectoRepo, integranteRepo)

	puerto := os.Getenv("PORT")
	if puerto == "" {
		puerto = puertoPorDefecto
	}

	r := setupRouter(handler.ServicioProyectos(proyectoService))

	log.Printf("Servidor iniciado en el puerto %s", puerto)
	if err := r.Run(":" + puerto); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}
