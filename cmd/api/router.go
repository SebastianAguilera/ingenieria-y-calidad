package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"ingenieria-y-calidad/internal/config"
	"ingenieria-y-calidad/internal/handler"
)

// setupRouter conserva la firma historica para los tests de enrutamiento: usa
// una configuracion de prueba sin restricciones de CORS. main usa
// setupRouterConConfig para inyectar la configuracion real del entorno.
func setupRouter(servicioProyectos handler.ServicioProyectos, servicioHistorias handler.ServicioHistorias, servicioSprints handler.ServicioSprints, servicioUsuarios handler.ServicioUsuarios, servicioSprintBacklog handler.ServicioSprintBacklog, servicioDefectos handler.ServicioDefectos, servicioCriterios handler.ServicioAcceptanceCriteria, servicioPoker handler.ServicioPoker, servicioMetricas handler.ServicioMetricas, servicioAuth handler.ServicioAuth, jwtSecret string) *gin.Engine {
	return setupRouterConConfig(
		servicioProyectos,
		servicioHistorias,
		servicioSprints,
		servicioUsuarios,
		servicioSprintBacklog,
		servicioDefectos,
		servicioCriterios,
		servicioPoker,
		servicioMetricas,
		servicioAuth,
		config.Config{
			AppEnv:              "test",
			JWTSecret:           jwtSecret,
			AuthRateLimitMax:    20,
			AuthRateLimitWindow: time.Minute,
		},
	)
}

// setupRouterConConfig construye el router de la aplicacion. Vive fuera de main
// para que los tests ejerciten exactamente las mismas rutas que se sirven en
// produccion. R-06: todo lo que no figura en la lista publica queda tras el
// middleware de autenticacion (default-deny).
func setupRouterConConfig(servicioProyectos handler.ServicioProyectos, servicioHistorias handler.ServicioHistorias, servicioSprints handler.ServicioSprints, servicioUsuarios handler.ServicioUsuarios, servicioSprintBacklog handler.ServicioSprintBacklog, servicioDefectos handler.ServicioDefectos, servicioCriterios handler.ServicioAcceptanceCriteria, servicioPoker handler.ServicioPoker, servicioMetricas handler.ServicioMetricas, servicioAuth handler.ServicioAuth, cfg config.Config) *gin.Engine {
	r := gin.Default()
	// Render no publica un rango fijo de proxies: se confia en el edge de
	// Render/Cloudflare para que c.ClientIP() lea el IP del cliente. Ver la
	// advertencia sobre X-Forwarded-For en el README.
	_ = r.SetTrustedProxies([]string{"0.0.0.0/0", "::/0"})
	r.GET("/health", healthHandler)
	r.Use(handler.CORS(cfg.AppEnv, cfg.CORSAllowedOrigins))
	if servicioAuth != nil {
		handler.RegistrarRutasAuth(r, servicioAuth,
			handler.LimitarAuth(cfg.AuthRateLimitMax, cfg.AuthRateLimitWindow))
	}
	r.Use(handler.Autenticar(cfg.JWTSecret, "/health", "/api/auth/login", "/api/auth/registro"))
	// US-F10: el guard de membresia se registra justo despues del middleware de
	// autenticacion y antes de cualquier ruta de proyectos: en Gin las rutas
	// registradas antes del Use no reciben la cadena de middlewares.
	r.Use(handler.ExigirAccesoAProyecto(servicioProyectos))
	handler.RegistrarRutasProyectos(r, servicioProyectos)
	handler.RegistrarRutasHistorias(r, servicioHistorias)
	handler.RegistrarRutasSprints(r, servicioSprints)
	if servicioUsuarios != nil {
		handler.RegistrarRutasUsuarios(r, servicioUsuarios)
	}
	if servicioSprintBacklog != nil {
		handler.RegistrarRutasSprintBacklog(r, servicioSprintBacklog)
	}
	if servicioDefectos != nil {
		handler.RegistrarRutasDefectos(r, servicioDefectos)
	}
	if servicioCriterios != nil {
		handler.RegistrarRutasAcceptanceCriteria(r, servicioCriterios)
	}
	if servicioPoker != nil {
		handler.RegistrarRutasPoker(r, servicioPoker)
	}
	if servicioMetricas != nil {
		handler.RegistrarRutasMetricas(r, servicioMetricas)
	}
	return r
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Software Metrics & Estimation Engine API running",
	})
}
