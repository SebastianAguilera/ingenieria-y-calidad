# Deploy: API en Render + frontend en Vercel

## Arquitectura

```
Navegador
  │  https://ingenieria-y-calidad.vercel.app  (SPA estática)
  │  llamadas directas con CORS
  ▼
Render (Docker, Go)  ──red privada──▶  Render Postgres "metricas-db"
  https://<api>.onrender.com/api/...
```

- El frontend llama directamente a la API. La URL base sale de `VITE_API_BASE_URL`
  (variable de build en Vercel). Si no está definida, usa `/api` (proxy de Vite en
  desarrollo).
- La API aplica CORS con allowlist exacta (`CORS_ALLOWED_ORIGINS`) en producción.
  Fuera de producción (`APP_ENV != production`) acepta cualquier origen.
- Rate limiting de `POST /api/auth/login` y `POST /api/auth/registro`: 20 solicitudes
  por minuto por IP y endpoint (configurable).
- La base existente **no** se declara en el Blueprint: se referencia por nombre
  (`render.yaml`).

## Requisitos previos

- Repositorio en GitHub.
- Cuenta de Render con la instancia Postgres **`metricas-db`** ya creada en Virginia.
- Cuenta de Vercel.

## Render (API)

1. En Render: **New > Blueprint**, conectar el repositorio y aplicar `render.yaml`.
2. El Blueprint crea el servicio web `metricas-api` y referencia la base existente
   (`fromDatabase.name: metricas-db`). Si el recurso no existe con ese nombre exacto,
   el sync falla.
3. `JWT_SECRET` se genera automáticamente (`generateValue: true`).
4. `CORS_ALLOWED_ORIGINS` = `https://ingenieria-y-calidad.vercel.app`.
5. Health check: `GET /health` (público, sin CORS ni rate limit).

## Vercel (frontend)

1. Importar el repositorio y configurar **Root Directory = `frontend`**.
2. Framework preset: **Vite**. Build y output están en `frontend/vercel.json`.
3. Variable de entorno de build:
   - `VITE_API_BASE_URL` = URL pública del servicio de Render, sin barra final
     (por ejemplo, `https://<api>.onrender.com`).
4. **Importante:** `VITE_API_BASE_URL` se incrusta en el build. Después de cargarla o
   cambiarla hay que hacer **Redeploy** en Vercel; un deploy anterior no la toma.

## Variables (solo nombres)

| Variable | Servicio | ¿Secreto? | Descripción |
| --- | --- | --- | --- |
| `APP_ENV` | Render | No | `production` activa la allowlist estricta de CORS. |
| `PORT` | Render | No | Puerto HTTP del contenedor. |
| `DB_HOST` / `DB_PORT` / `DB_USER` | Render | No | Desde `fromDatabase`. |
| `DB_PASSWORD` | Render | Sí | Desde `fromDatabase`. |
| `DB_NAME_PROD` | Render | No | Desde `fromDatabase`. |
| `DB_SSLMODE` | Render | No | Conexión interna por red privada. |
| `JWT_SECRET` | Render | Sí | Generado por Render. |
| `CORS_ALLOWED_ORIGINS` | Render | No | Orígenes exactos separados por coma. Obligatoria en producción. |
| `AUTH_RATE_LIMIT_MAX` | Render | No | Solicitudes por ventana en login/registro (20 por defecto). |
| `AUTH_RATE_LIMIT_WINDOW` | Render | No | Duración de la ventana, formato Go (`1m` por defecto). |
| `VITE_API_BASE_URL` | Vercel (build) | No | URL pública de la API. Requiere Redeploy al cambiar. |

Ningún valor secreto debe escribirse en el repositorio ni en `render.yaml`.
`VITE_*` es público por diseño: no poner secretos ahí.

## Reglas de CORS

- Producción: solo se refleja el origen exacto de `CORS_ALLOWED_ORIGINS`.
  Un `Origin` presente y no permitido recibe `403 ORIGEN_NO_PERMITIDO`.
- Sin cabecera `Origin` (curl, health checks, herramientas) la solicitud se procesa
  normalmente. CORS solo controla qué puede leer el navegador; no autentica.
- Preflight `OPTIONS` permitido: `204` con `Access-Control-Allow-Methods`,
  `Access-Control-Allow-Headers: Authorization, Content-Type, Accept` y
  `Access-Control-Max-Age: 600`. No se envía `Access-Control-Allow-Credentials`.

## Rate limiting de autenticación

- Límite por IP (`c.ClientIP()`) y endpoint, dentro de una ventana fija.
- Respuesta al exceder: `429 DEMASIADAS_SOLICITUDES` con cabecera `Retry-After`.
- Cabeceras `X-RateLimit-Limit` y `X-RateLimit-Remaining` en las respuestas de auth.
- El servidor confía en el IP reportado por el edge de Render (`SetTrustedProxies`).
  Si el proxy no está configurado correctamente, `X-Forwarded-For` puede falsificarse
  y evadir el límite. Para este proyecto escolar se acepta ese riesgo; un entorno real
  requeriría un almacenamiento compartido y proxies confiables fijos.

## Smoke test post-deploy

1. `GET https://<api>.onrender.com/health` → `200`.
2. `OPTIONS /api/auth/login` con `Origin: https://ingenieria-y-calidad.vercel.app` → `204`
   con `Access-Control-Allow-Origin` igual a ese origen.
3. `OPTIONS /api/auth/login` con `Origin: https://evil.example` → `403 ORIGEN_NO_PERMITIDO`.
4. `POST /api/auth/login` sin `Origin` (curl) → no es `403` (se procesa).
5. Más de `AUTH_RATE_LIMIT_MAX` intentos en la ventana → `429` con `Retry-After`.
6. Login y una operación autenticada desde el frontend en Vercel → `200`.
7. Refrescar una ruta interna de la SPA (por ejemplo `/proyectos`) → sin `404`.

## Problemas frecuentes

- Error de CORS en el navegador: el origen del frontend no coincide exactamente con
  `CORS_ALLOWED_ORIGINS` (esquema, host o barra final).
- Las llamadas van a `/api` en producción: falta `VITE_API_BASE_URL` o no se hizo
  Redeploy en Vercel.
- `502/504` en la primera llamada: cold start del plan Free de Render.
- Health check fallido: la app no escucha en el puerto de `PORT`.
- Sync del Blueprint fallido: `fromDatabase.name` no coincide con el recurso Render.

## Límites conocidos

- El plan Free de Postgres vence a los 30 días (con 14 días de gracia).
- El plan Free del servicio web se suspende por inactividad y tarda en despertar.
- El rate limit es en memoria por instancia; con varias instancias el límite no es global.
