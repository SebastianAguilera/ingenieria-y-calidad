# Especificación SDD: US-10 - Autenticación de usuarios con JWT

> **Issue:** `#61` · **Estado:** especificación (sin implementar) · **Metodología:** SDD → BDD → TDD
> **Dependencias:** US-08 (usuarios) · Frontend: US-F08 / US-F02

## 1. Objetivo

Autenticar usuarios con email + contraseña: guardar la contraseña solo como hash bcrypt, emitir un JWT en `POST /api/auth/login` y proteger todas las rutas `/api/*` con un middleware (públicas: `GET /health` y `POST /api/auth/login`).

## 2. Entradas

- **`.env`:** `JWT_SECRET` (obligatoria; sin ella la app no arranca) y `JWT_EXPIRY` (default `24h`, formato `time.ParseDuration`).
- **`POST /api/auth/login`:** `{ "email": "...", "password": "..." }` (ambos obligatorios y no vacíos; el email se normaliza con `domain.NormalizarEmail`).
- **Rutas protegidas:** header `Authorization: Bearer <token>`.

## 3. Salidas

| Caso | Status | JSON |
|---|---|---|
| Login OK | 200 | `{ "token": "<JWT HS256>", "usuario": {id, nombre, email, creado_en} }` — nunca `password_hash` |
| Credenciales inválidas (contraseña mala, email inexistente o sin hash) | 401 | `{ "error": "CREDENCIALES_INVALIDAS", "mensaje": "email o contraseña incorrectos" }` (idéntico en ambos casos) |
| Ruta protegida sin `Authorization` / sin `Bearer ` | 401 | `{ "error": "TOKEN_REQUERIDO", "mensaje": "la solicitud no incluye un token de autenticación" }` |
| Token malformado, firma ajena, `alg` distinto o expirado | 401 | `{ "error": "TOKEN_INVALIDO", "mensaje": "el token de autenticación no es válido o ha expirado" }` |
| Campo obligatorio vacío | 422 | `{ "error": "VALIDACION", "campo": "email"|"password", ... }` |
| Body no JSON | 400 | `{ "error": "JSON_INVALIDO", ... }` (patrón existente) |
| `JWT_SECRET` ausente | — | la app no arranca (`config.Load` falla) |

## 4. Reglas de negocio

- **R-01** La contraseña se persiste solo como hash bcrypt (`golang.org/x/crypto/bcrypt`); comparación con `bcrypt.CompareHashAndPassword` (nunca `==`).
- **R-02** `domain.Usuario` lleva `PasswordHash` con `json:"-"`; ningún endpoint lo devuelve.
- **R-03** Email normalizado (trim + minúsculas) antes de buscar con `ObtenerPorEmail`.
- **R-04** Email inexistente, usuario sin hash y contraseña incorrecta → **el mismo** `401 CREDENCIALES_INVALIDAS` (no se revela existencia; no se usa `ErrUsuarioNoEncontrado` en login).
- **R-05** JWT: algoritmo **HS256** fijo, claims `sub` (string del ID), `email`, `iat`, `exp = iat + JWT_EXPIRY`; firma solo con `JWT_SECRET`.
- **R-06** Middleware (default-deny en `/api/*`): valida firma + `alg` + `exp`. Sin header/sin prefijo → `TOKEN_REQUERIDO`; con `Bearer ` pero fallo de parseo/firma/expiración → `TOKEN_INVALIDO`.
- **R-06a** (US-F10) El middleware además expone el `sub` como identidad del request (`c.Set`): un token válido sin `sub` o con `sub` no numérico se rechaza con `401 TOKEN_INVALIDO`. El guard de membresía (`ExigirAccesoAProyecto`) consume esa identidad justo después de este middleware.
- **R-07** Sin logout server-side ni blacklist: el cliente borra su token (un token válido vive hasta `exp`).
- **R-08** Orden de errores en login: JSON (400) → campos (422) → credenciales (401) → éxito (200).
- **R-09** Usuarios sin hash (p. ej. `seba` pre-backfill) → `401 CREDENCIALES_INVALIDAS`.

## 5. Restricciones

- La contraseña en claro solo vive en el request de login: no se loguea ni se refleja en respuestas.
- Mensaje de credenciales genérico e idéntico para "no existe" y "contraseña mala" (anti user-enumeration).
- `alg` distinto de HS256 (incluido `none`) → rechazado.
- `JWT_SECRET` nunca en el repo (solo `.env`, con placeholder en `.env.example`).
- Clean Architecture: reglas en `service`, tipos/errores sentinelas en `domain`, JWT/middleware en `handler`, config en `internal/config` (inyectada en el constructor, testable sin `.env`).
- Dependencias nuevas: `github.com/golang-jwt/jwt/v5` y `golang.org/x/crypto` (directo).
- Migración por `AutoMigrate` existente (columna `password_hash`, `varchar(72)`).

## 6. Casos límite

| Caso | Comportamiento |
|---|---|
| Email con mayúsculas/espacios (`" Seba@Example.COM "`) | normaliza → 200 |
| `email: ""` / `password: ""` | 422 con el `campo` correspondiente (no llega a bcrypt) |
| Usuario sin `password_hash` | 401 `CREDENCIALES_INVALIDAS` (igual que contraseña mala) |
| Token con otra firma / expirado / `alg:none` | 401 `TOKEN_INVALIDO` |
| Header ausente o sin `Bearer ` | 401 `TOKEN_REQUERIDO`; `Bearer basura` → `TOKEN_INVALIDO` |
| Hash corrupto en DB | 401 `CREDENCIALES_INVALIDAS` (detalle solo en logs) |
| `JWT_SECRET` vacía o `JWT_EXPIRY` inválido | no arranca la app |
| Dos sesiones simultáneas | ambas válidas hasta su `exp` (sin blacklist) |
| `/health` y login sin token | responden normal, nunca 401 |

## 7. Condiciones de error

Errores sentinelas nuevos en `internal/domain/errores.go` + ramas en `handler/errores.go` (`buscarMapeo`), sin solapar con `ErrUsuarioNoEncontrado` (que sigue siendo 404 en el resto de endpoints):

```go
ErrCredencialesInvalidas = errors.New("email o contraseña incorrectos")   // 401
ErrTokenInvalido         = errors.New("el token ... no es válido o ha expirado") // 401
ErrTokenRequerido        = errors.New("la solicitud no incluye un token de autenticación") // 401
ErrPasswordVacio         = errors.New("la contraseña es obligatoria")     // 422
```

## 8. Criterios de aceptación

1. **Given** usuario con hash válido, **When** login correcto, **Then** 200 con `token` (HS256, sin expirar) y `usuario` sin `password_hash`.
2. **Given** contraseña incorrecta, **When** login, **Then** 401 `CREDENCIALES_INVALIDAS` sin `token`.
3. **Given** email inexistente, **When** login, **Then** 401 con **exactamente el mismo** `mensaje` que el CA-2.
4. **Given** ruta `/api/*` protegida, **When** la llamo sin `Authorization`, **Then** 401 `TOKEN_REQUERIDO` (no llega al servicio).
5. **Given** token con firma ajena o expirado, **When** la llamo, **Then** 401 `TOKEN_INVALIDO`.
6. **Given** token válido del login, **When** `GET /api/proyectos`, **Then** 200 con el listado normal.
7. **Given** un usuario con contraseña, **When** inspecciono DB y respuestas, **Then** solo hay hash bcrypt y ningún JSON contiene `password_hash`.

## Capas a tocar

| Capa | Archivo | Cambio |
|---|---|---|
| domain | `usuario.go`, `errores.go` | `PasswordHash json:"-"`; errores sentinelas |
| config | `config.go` | `JWTSecret`, `JWTExpiry` + validación |
| repository | `models/usuario_model.go`, `mappers.go` | columna `password_hash` |
| service | `auth_service.go` (nuevo) | `Login`: normalizar → buscar → compare → JWT |
| handler | `auth_handler.go`, `middleware_auth.go` (nuevos), `errores.go` | DTO login, rutas, middleware, mapeo de errores |
| router | `cmd/api/router.go` | registrar auth + middleware con lista pública explícita |
| entrypoint | `cmd/api/main.go` | inyectar repo + secret; fallar si falta secret |
| config ejemplo | `.env.example` | `JWT_SECRET`, `JWT_EXPIRY=24h` |
| backfill | script SQL | contraseña inicial para `seba` (id 1) |

## Plan de tests (RED → GREEN)

1. `config_test`: `JWT_SECRET` obligatoria, `JWT_EXPIRY` default.
2. `domain`: errores sentinelas + `PasswordHash` no serializado.
3. `migraciones_test`: columna `password_hash` presente (skip sin DB).
4. `service/auth_service_test` (nuevo): CA-1/2/3, email normalizado, sin hash → mismo error, claims del token (round-trip).
5. `handler/auth_handler_test` (nuevo): 200/401/422/400.
6. `handler/middleware_auth_test` (nuevo): CA-4/5, variantes de header, rutas públicas.
7. `cmd/api/router_test` (extender): wiring — login/health sin token OK; `/api/proyectos` sin token 401; con token OK; firma ajena 401.
8. Test de serialización: ningún JSON con `password_hash` (CA-7).

Convención: `t.Parallel()`, testify, mocks al estilo de `service/mocks_test.go`, comentarios que citan la CA.

## Fuera de alcance

Refresh tokens · recuperación de contraseña · registro público (el alta sigue en US-08) · roles/permisos por proyecto (el middleware solo autentica) · logout server-side · rate limiting · MFA/SSO · frontend (US-F08).

## Decisiones abiertas

| # | Decisión | Recomendación |
|---|---|---|
| D-1 | `POST /api/usuarios` público o protegido | protegido por JWT (el alta se hace con sesión) |
| D-2 | Backfill de `seba` | script SQL puntual |
| D-3 | Dummy compare anti-timing cuando no existe el email | sí (hardening) |
| D-4 | Validar formato de email en login | no: solo normalizar (el formato estricto ya aplica en alta) |
| D-5 | Límite de largo de `password` (anti-DoS) | 72 bytes máx. |
