package handler

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	codigoDemasiadasSolicitudes  = "DEMASIADAS_SOLICITUDES"
	mensajeDemasiadasSolicitudes = "demasiadas solicitudes, intentá de nuevo más tarde"

	prefijoRutasAuth = "/api/auth/"
)

// LimitarAuth limita los intentos de autenticacion por IP y endpoint. En
// produccion se registra sobre el grupo /api/auth, pero ademas ignora cualquier
// ruta ajena para que no pueda aplicarse por error a toda la API.
func LimitarAuth(max int, duracion time.Duration) gin.HandlerFunc {
	return limitarAuthConReloj(max, duracion, time.Now)
}

func limitarAuthConReloj(max int, duracion time.Duration, ahora func() time.Time) gin.HandlerFunc {
	if max <= 0 || duracion <= 0 {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	limitador := &limitadorAuth{
		max:            max,
		duracion:       duracion,
		ventanas:       make(map[string]*ventanaRateLimit),
		ahora:          ahora,
		ultimaLimpieza: ahora(),
	}

	return func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, prefijoRutasAuth) {
			c.Next()
			return
		}

		ip := c.ClientIP()
		if ip == "" {
			ip = "desconocido"
		}
		clave := ip + "|" + c.Request.URL.Path

		permitido, restantes, espera := limitador.permitir(clave)
		c.Header("X-RateLimit-Limit", strconv.Itoa(max))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(restantes))

		if !permitido {
			segundos := int(math.Ceil(espera.Seconds()))
			if segundos < 1 {
				segundos = 1
			}
			c.Header("Retry-After", strconv.Itoa(segundos))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, respuestaError{
				Error:   codigoDemasiadasSolicitudes,
				Mensaje: mensajeDemasiadasSolicitudes,
			})
			return
		}

		c.Next()
	}
}

type ventanaRateLimit struct {
	contador int
	reinicio time.Time
}

type limitadorAuth struct {
	mu             sync.Mutex
	max            int
	duracion       time.Duration
	ventanas       map[string]*ventanaRateLimit
	ahora          func() time.Time
	ultimaLimpieza time.Time
}

func (l *limitadorAuth) permitir(clave string) (bool, int, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	ahora := l.ahora()
	l.limpiar(ahora)

	ventana, existe := l.ventanas[clave]
	if !existe || !ahora.Before(ventana.reinicio) {
		ventana = &ventanaRateLimit{
			contador: 1,
			reinicio: ahora.Add(l.duracion),
		}
		l.ventanas[clave] = ventana
		return true, l.max - 1, 0
	}

	if ventana.contador >= l.max {
		return false, 0, ventana.reinicio.Sub(ahora)
	}

	ventana.contador++
	return true, l.max - ventana.contador, 0
}

// limpiar evita que el mapa crezca sin limite cuando cambian las IPs. Se
// ejecuta como maximo una vez por ventana, con el mutex ya tomado.
func (l *limitadorAuth) limpiar(ahora time.Time) {
	if ahora.Sub(l.ultimaLimpieza) < l.duracion {
		return
	}

	for clave, ventana := range l.ventanas {
		if !ahora.Before(ventana.reinicio) {
			delete(l.ventanas, clave)
		}
	}
	l.ultimaLimpieza = ahora
}
