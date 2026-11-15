package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"ingenieria-y-calidad/internal/domain"
)

// ServicioPoker es el puerto que el handler necesita del service de Planning
// Poker. Se define en el consumidor para que los tests inyecten un mock.
type ServicioPoker interface {
	CrearSesion(ctx context.Context, proyectoID, historiaID, usuarioID uint) (*domain.DetalleSesionPoker, error)
	ListarPorHistoria(ctx context.Context, proyectoID, historiaID uint) ([]domain.SesionPoker, error)
	Detalle(ctx context.Context, proyectoID, sesionID, usuarioID uint) (*domain.DetalleSesionPoker, error)
	Activas(ctx context.Context, proyectoID, usuarioID uint) ([]domain.SesionPokerActiva, error)
	Votar(ctx context.Context, proyectoID, sesionID, usuarioID uint, puntos float64) (*domain.ResultadoVotoPoker, error)
	Revelar(ctx context.Context, proyectoID, sesionID, usuarioID uint) (*domain.DetalleSesionPoker, error)
	NuevaRonda(ctx context.Context, proyectoID, sesionID, usuarioID uint) (*domain.DetalleSesionPoker, error)
	Acordar(ctx context.Context, proyectoID, sesionID, usuarioID uint, puntos float64) (*domain.DetalleSesionPoker, error)
	Cancelar(ctx context.Context, proyectoID, sesionID, usuarioID uint) (*domain.DetalleSesionPoker, error)
}

// PokerHandler expone las sesiones de Planning Poker de US-14 por HTTP.
type PokerHandler struct {
	servicio ServicioPoker
}

// NuevoPokerHandler construye el controlador con inyección de dependencias.
func NuevoPokerHandler(servicio ServicioPoker) *PokerHandler {
	return &PokerHandler{servicio: servicio}
}

// RegistrarRutasPoker registra las rutas de US-14, US-15 y US-16 bajo
// /api/proyectos/:id.
func RegistrarRutasPoker(r *gin.Engine, servicio ServicioPoker) {
	h := NuevoPokerHandler(servicio)
	api := r.Group("/api")
	proyectos := api.Group("/proyectos")
	{
		proyectos.POST("/:id/historias/:historiaId/poker-sessions", h.Crear)
		proyectos.GET("/:id/historias/:historiaId/poker-sessions", h.ListarPorHistoria)
		proyectos.GET("/:id/poker-sessions/activas", h.Activas)
		proyectos.GET("/:id/poker-sessions/:sesionId", h.Detalle)
		proyectos.POST("/:id/poker-sessions/:sesionId/votos", h.Votar)
		proyectos.POST("/:id/poker-sessions/:sesionId/revelar", h.Revelar)
		proyectos.POST("/:id/poker-sessions/:sesionId/rondas", h.NuevaRonda)
		proyectos.POST("/:id/poker-sessions/:sesionId/acuerdo", h.Acordar)
		proyectos.POST("/:id/poker-sessions/:sesionId/cancelar", h.Cancelar)
	}
}

// Crear maneja POST /api/proyectos/:id/historias/:historiaId/poker-sessions.
func (h *PokerHandler) Crear(c *gin.Context) {
	proyectoID, ok := parsearID(c, "id")
	if !ok {
		return
	}
	historiaID, ok := parsearID(c, "historiaId")
	if !ok {
		return
	}
	usuarioID, ok := usuarioDelContexto(c)
	if !ok {
		return
	}

	detalle, err := h.servicio.CrearSesion(c.Request.Context(), proyectoID, historiaID, usuarioID)
	if err != nil {
		traducirError(c, err)
		return
	}

	c.JSON(http.StatusCreated, construirRespuestaDetalleSesion(detalle))
}

// ListarPorHistoria maneja
// GET /api/proyectos/:id/historias/:historiaId/poker-sessions.
func (h *PokerHandler) ListarPorHistoria(c *gin.Context) {
	proyectoID, ok := parsearID(c, "id")
	if !ok {
		return
	}
	historiaID, ok := parsearID(c, "historiaId")
	if !ok {
		return
	}

	sesiones, err := h.servicio.ListarPorHistoria(c.Request.Context(), proyectoID, historiaID)
	if err != nil {
		traducirError(c, err)
		return
	}

	respuestas := make([]respuestaHistorialSesion, 0, len(sesiones))
	for _, sesion := range sesiones {
		respuestas = append(respuestas, construirRespuestaHistorialSesion(&sesion))
	}
	c.JSON(http.StatusOK, respuestas)
}

// Detalle maneja GET /api/proyectos/:id/poker-sessions/:sesionId.
func (h *PokerHandler) Detalle(c *gin.Context) {
	proyectoID, ok := parsearID(c, "id")
	if !ok {
		return
	}
	sesionID, ok := parsearID(c, "sesionId")
	if !ok {
		return
	}
	usuarioID, ok := usuarioDelContexto(c)
	if !ok {
		return
	}

	detalle, err := h.servicio.Detalle(c.Request.Context(), proyectoID, sesionID, usuarioID)
	if err != nil {
		traducirError(c, err)
		return
	}

	c.JSON(http.StatusOK, construirRespuestaDetalleSesion(detalle))
}

// Activas maneja GET /api/proyectos/:id/poker-sessions/activas.
func (h *PokerHandler) Activas(c *gin.Context) {
	proyectoID, ok := parsearID(c, "id")
	if !ok {
		return
	}
	usuarioID, ok := usuarioDelContexto(c)
	if !ok {
		return
	}

	activas, err := h.servicio.Activas(c.Request.Context(), proyectoID, usuarioID)
	if err != nil {
		traducirError(c, err)
		return
	}

	respuestas := make([]respuestaSesionActiva, 0, len(activas))
	for _, activa := range activas {
		respuestas = append(respuestas, construirRespuestaSesionActiva(&activa))
	}
	c.JSON(http.StatusOK, respuestas)
}

// Votar maneja POST /api/proyectos/:id/poker-sessions/:sesionId/votos. El
// cuerpo se lee de forma laxa para distinguir un JSON malformado (400) de un
// valor de puntos ausente o no numérico (422 PUNTOS_POKER_INVALIDOS).
func (h *PokerHandler) Votar(c *gin.Context) {
	proyectoID, ok := parsearID(c, "id")
	if !ok {
		return
	}
	sesionID, ok := parsearID(c, "sesionId")
	if !ok {
		return
	}
	usuarioID, ok := usuarioDelContexto(c)
	if !ok {
		return
	}

	var cuerpo struct {
		Puntos any `json:"puntos"`
	}
	if err := c.ShouldBindJSON(&cuerpo); err != nil {
		responderJSONInvalido(c)
		return
	}
	puntos, esNumero := cuerpo.Puntos.(float64)
	if !esNumero {
		traducirError(c, domain.ErrPuntosPokerInvalidos)
		return
	}

	resultado, err := h.servicio.Votar(c.Request.Context(), proyectoID, sesionID, usuarioID, puntos)
	if err != nil {
		traducirError(c, err)
		return
	}

	c.JSON(http.StatusOK, construirRespuestaVotoRegistrado(resultado))
}

// Revelar maneja POST /api/proyectos/:id/poker-sessions/:sesionId/revelar.
func (h *PokerHandler) Revelar(c *gin.Context) {
	proyectoID, ok := parsearID(c, "id")
	if !ok {
		return
	}
	sesionID, ok := parsearID(c, "sesionId")
	if !ok {
		return
	}
	usuarioID, ok := usuarioDelContexto(c)
	if !ok {
		return
	}

	detalle, err := h.servicio.Revelar(c.Request.Context(), proyectoID, sesionID, usuarioID)
	if err != nil {
		traducirError(c, err)
		return
	}

	c.JSON(http.StatusOK, construirRespuestaDetalleSesion(detalle))
}

// NuevaRonda maneja POST /api/proyectos/:id/poker-sessions/:sesionId/rondas.
func (h *PokerHandler) NuevaRonda(c *gin.Context) {
	proyectoID, ok := parsearID(c, "id")
	if !ok {
		return
	}
	sesionID, ok := parsearID(c, "sesionId")
	if !ok {
		return
	}
	usuarioID, ok := usuarioDelContexto(c)
	if !ok {
		return
	}

	detalle, err := h.servicio.NuevaRonda(c.Request.Context(), proyectoID, sesionID, usuarioID)
	if err != nil {
		traducirError(c, err)
		return
	}

	c.JSON(http.StatusOK, construirRespuestaDetalleSesion(detalle))
}

// Acordar maneja POST /api/proyectos/:id/poker-sessions/:sesionId/acuerdo. El
// cuerpo se lee de forma laxa para distinguir JSON malformado (400) de puntos
// ausentes o no numéricos (422 PUNTOS_POKER_INVALIDOS).
func (h *PokerHandler) Acordar(c *gin.Context) {
	proyectoID, ok := parsearID(c, "id")
	if !ok {
		return
	}
	sesionID, ok := parsearID(c, "sesionId")
	if !ok {
		return
	}
	usuarioID, ok := usuarioDelContexto(c)
	if !ok {
		return
	}

	var cuerpo struct {
		Puntos any `json:"puntos"`
	}
	if err := c.ShouldBindJSON(&cuerpo); err != nil {
		responderJSONInvalido(c)
		return
	}
	puntos, esNumero := cuerpo.Puntos.(float64)
	if !esNumero {
		traducirError(c, domain.ErrPuntosPokerInvalidos)
		return
	}

	detalle, err := h.servicio.Acordar(c.Request.Context(), proyectoID, sesionID, usuarioID, puntos)
	if err != nil {
		traducirError(c, err)
		return
	}

	c.JSON(http.StatusOK, construirRespuestaDetalleSesion(detalle))
}

// Cancelar maneja POST /api/proyectos/:id/poker-sessions/:sesionId/cancelar.
func (h *PokerHandler) Cancelar(c *gin.Context) {
	proyectoID, ok := parsearID(c, "id")
	if !ok {
		return
	}
	sesionID, ok := parsearID(c, "sesionId")
	if !ok {
		return
	}
	usuarioID, ok := usuarioDelContexto(c)
	if !ok {
		return
	}

	detalle, err := h.servicio.Cancelar(c.Request.Context(), proyectoID, sesionID, usuarioID)
	if err != nil {
		traducirError(c, err)
		return
	}

	c.JSON(http.StatusOK, construirRespuestaDetalleSesion(detalle))
}

// usuarioDelContexto recupera la identidad que dejó el middleware Autenticar.
func usuarioDelContexto(c *gin.Context) (uint, bool) {
	usuarioID, existe := UsuarioDesdeContexto(c)
	if !existe {
		traducirError(c, domain.ErrTokenInvalido)
		return 0, false
	}
	return usuarioID, true
}
