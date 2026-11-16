package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// errJWTSecretFaltante se devuelve cuando JWT_SECRET no esta definida: sin
// secreto no se pueden firmar ni verificar tokens (US-10).
var errJWTSecretFaltante = errors.New("JWT_SECRET es obligatoria: definila en el archivo .env para poder emitir tokens")

// errCORSAllowedOriginsFaltante se devuelve cuando APP_ENV=production y no hay
// origenes permitidos: en produccion el navegador solo debe poder leer las
// respuestas desde el frontend autorizado (fail-closed).
var errCORSAllowedOriginsFaltante = errors.New("CORS_ALLOWED_ORIGINS es obligatoria cuando APP_ENV=production: definí el origen exacto del frontend")

// errCORSAllowedOriginsInvalida se devuelve cuando la lista de origenes tiene
// un valor que no es un origen HTTP(S) exacto.
var errCORSAllowedOriginsInvalida = errors.New("CORS_ALLOWED_ORIGINS invalida: usá orígenes exactos con esquema y host, sin path, query, fragmento ni comodines")

// errAuthRateLimitMaxInvalido se devuelve cuando AUTH_RATE_LIMIT_MAX no es un
// entero positivo.
var errAuthRateLimitMaxInvalido = errors.New("AUTH_RATE_LIMIT_MAX invalida: usá un entero positivo")

// errAuthRateLimitWindowInvalida se devuelve cuando AUTH_RATE_LIMIT_WINDOW no
// es una duracion valida de time.ParseDuration.
var errAuthRateLimitWindowInvalida = errors.New("AUTH_RATE_LIMIT_WINDOW invalida: usá una duración válida, por ejemplo 1m")

// expiracionJWTPorDefecto es la vigencia del token cuando JWT_EXPIRY no esta
// definida.
const expiracionJWTPorDefecto = 24 * time.Hour

const (
	limiteAuthPorDefecto       = 20
	ventanaRateLimitPorDefecto = time.Minute
)

type Config struct {
	AppEnv              string
	Port                string
	DBHost              string
	DBPort              string
	DBUser              string
	DBPassword          string
	DBName              string
	DBSSLMode           string
	JWTSecret           string
	JWTExpiry           time.Duration
	CORSAllowedOrigins  []string
	AuthRateLimitMax    int
	AuthRateLimitWindow time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}

	var dbName string
	switch appEnv {
	case "test":
		dbName = os.Getenv("DB_NAME_TEST")
		if dbName == "" {
			dbName = "metrics_db_test"
		}
	case "production":
		dbName = os.Getenv("DB_NAME_PROD")
		if dbName == "" {
			dbName = "metrics_db"
		}
	default: // development
		dbName = os.Getenv("DB_NAME_DEV")
		if dbName == "" {
			dbName = "metrics_db"
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, errJWTSecretFaltante
	}

	jwtExpiry := expiracionJWTPorDefecto
	if valor := os.Getenv("JWT_EXPIRY"); valor != "" {
		d, err := time.ParseDuration(valor)
		if err != nil || d <= 0 {
			return nil, fmt.Errorf("JWT_EXPIRY invalida (%q): use una duracion valida de time.ParseDuration, por ejemplo 24h", valor)
		}
		jwtExpiry = d
	}

	corsAllowedOrigins, err := parseCORSAllowedOrigins(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if err != nil {
		return nil, err
	}
	if appEnv == "production" && len(corsAllowedOrigins) == 0 {
		return nil, errCORSAllowedOriginsFaltante
	}

	authRateLimitMax := limiteAuthPorDefecto
	if valor := strings.TrimSpace(os.Getenv("AUTH_RATE_LIMIT_MAX")); valor != "" {
		n, err := strconv.Atoi(valor)
		if err != nil || n <= 0 {
			return nil, errAuthRateLimitMaxInvalido
		}
		authRateLimitMax = n
	}

	authRateLimitWindow := ventanaRateLimitPorDefecto
	if valor := strings.TrimSpace(os.Getenv("AUTH_RATE_LIMIT_WINDOW")); valor != "" {
		d, err := time.ParseDuration(valor)
		if err != nil || d <= 0 {
			return nil, errAuthRateLimitWindowInvalida
		}
		authRateLimitWindow = d
	}

	cfg := &Config{
		AppEnv:              appEnv,
		Port:                port,
		DBHost:              getEnvOrDefault("DB_HOST", "localhost"),
		DBPort:              getEnvOrDefault("DB_PORT", "5432"),
		DBUser:              getEnvOrDefault("DB_USER", "postgres"),
		DBPassword:          os.Getenv("DB_PASSWORD"),
		DBName:              dbName,
		DBSSLMode:           getEnvOrDefault("DB_SSLMODE", "disable"),
		JWTSecret:           jwtSecret,
		JWTExpiry:           jwtExpiry,
		CORSAllowedOrigins:  corsAllowedOrigins,
		AuthRateLimitMax:    authRateLimitMax,
		AuthRateLimitWindow: authRateLimitWindow,
	}

	return cfg, nil
}

// parseCORSAllowedOrigins separa la lista por comas y normaliza cada origen.
// Un origen vacio se ignora para que APP_ENV=development pueda arrancar sin la
// variable; en produccion Load exige que quede al menos uno.
func parseCORSAllowedOrigins(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	vistos := make(map[string]struct{})
	origenes := make([]string, 0)
	for _, parte := range strings.Split(raw, ",") {
		origen := strings.TrimSpace(parte)
		if origen == "" {
			continue
		}

		normalizado, err := normalizarOrigen(origen)
		if err != nil {
			return nil, fmt.Errorf("%w: %q", errCORSAllowedOriginsInvalida, origen)
		}
		if _, repetido := vistos[normalizado]; repetido {
			continue
		}
		vistos[normalizado] = struct{}{}
		origenes = append(origenes, normalizado)
	}

	return origenes, nil
}

// normalizarOrigen valida que el valor sea un origen exacto y lo devuelve en
// minusculas y sin barra final, para compararlo con la cabecera Origin.
func normalizarOrigen(origen string) (string, error) {
	if strings.Contains(origen, "*") {
		return "", errCORSAllowedOriginsInvalida
	}

	u, err := url.Parse(origen)
	if err != nil {
		return "", err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", errCORSAllowedOriginsInvalida
	}
	if u.Host == "" || u.User != nil {
		return "", errCORSAllowedOriginsInvalida
	}
	if u.Path != "" && u.Path != "/" {
		return "", errCORSAllowedOriginsInvalida
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return "", errCORSAllowedOriginsInvalida
	}

	return strings.ToLower(u.Scheme) + "://" + strings.ToLower(u.Host), nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
