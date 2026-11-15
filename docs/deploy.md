# Deploy: API en Render + frontend en Vercel

## Arquitectura

```
Navegador
  │  https://<app>.vercel.app/api/...
  ▼
Vercel (static + función api/[...path].ts)
  │  https://<api>.onrender.com/api/...  +  X-Frontend-Key
  ▼
Render (Docker, Go)  ──red privada──▶  Render Postgres "metricas-db"
```

- El navegador siempre llama a `/api/...` en el mismo origen (`frontend/src/api/client.ts:4`),
  por lo que no hace falta CORS.
- La función de Vercel agrega `X-Frontend-Key`; la API la valida cuando `FRONTEND_KEY`
  está definida (`internal/handler/middleware_frontend_key.go`).
- La base existente **no** se declara en el Blueprint: se referencia por nombre
  (`render.yaml`).

## Requisitos previos

- Repositorio en GitHub con la rama `master`.
- Cuenta de Render con la instancia Postgres **`metricas-db`** ya creada en Virginia.
- Cuenta de Vercel.

## Render (API)

1. En Render: **New > Blueprint**, conectar el repositorio y aplicar `render.yaml`.
2. El Blueprint crea el servicio web `metricas-api` y referencia la base existente:
   - `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME_PROD` usan
     `fromDatabase.name: metricas-db`.
   - Si el recurso no existe en el workspace con ese nombre exacto, el sync falla.
3. `JWT_SECRET` se genera automáticamente (`generateValue: true`).
4. `FRONTEND_KEY` queda en `sync: false`: Render la pide al aplicar el Blueprint.
5. Health check: `GET /health` (público, registrado antes de los middlewares).
6. Puerto: el Blueprint define `PORT=8080`; la app escucha en `cfg.Port`
   (`cmd/api/main.go:63`) y el `Dockerfile` declara `EXPOSE 8080`.

Validación del Blueprint (requiere sesión/workspace de Render):

```bash
render blueprints validate render.yaml
```

> Sin autenticación, este repositorio valida `render.yaml` contra el JSON Schema
> oficial de Render (`https://render.com/schema/render.yaml.json`).

## Vercel (frontend)

1. Importar el repositorio y configurar **Root Directory = `frontend`**.
2. Framework preset: **Vite**. El build y el output ya están en `frontend/vercel.json`
   (`npm run build`, `dist`).
3. Variables de entorno (solo del lado servidor, **nunca** con prefijo `VITE_`):
   - `API_ORIGIN`: URL pública del servicio de Render, sin barra final
     (por ejemplo, `https://<servicio>.onrender.com`).
   - `FRONTEND_KEY`: el mismo valor configurado en Render.
4. La función `frontend/api/[...path].ts` es un catch-all:
   - reenvía método, query, `Authorization` y body;
   - descarta cabeceras hop-by-hop y cualquier `X-Frontend-Key` entrante;
   - agrega la clave desde el entorno del servidor;
   - responde `500` si falta configuración, `504` si la API no responde a tiempo.
5. `maxDuration` = 120 s (`frontend/vercel.json`), suficiente para el cold start
   del plan Free de Render.

## Variables (solo nombres)

| Variable | Servicio | ¿Secreto? | Descripción |
| --- | --- | --- | --- |
| `APP_ENV` | Render | No | Activa la selección de `DB_NAME_PROD`. |
| `PORT` | Render | No | Puerto HTTP del contenedor. |
| `DB_HOST` | Render | No | Desde `fromDatabase` (`host`). |
| `DB_PORT` | Render | No | Desde `fromDatabase` (`port`). |
| `DB_USER` | Render | No | Desde `fromDatabase` (`user`). |
| `DB_PASSWORD` | Render | Sí | Desde `fromDatabase` (`password`). |
| `DB_NAME_PROD` | Render | No | Desde `fromDatabase` (`database`). |
| `DB_SSLMODE` | Render | No | Conexión interna por red privada. |
| `JWT_SECRET` | Render | Sí | Generado por Render. |
| `FRONTEND_KEY` | Render y Vercel | Sí | Mismo valor en ambos servicios. |
| `API_ORIGIN` | Vercel | No | URL pública de la API en Render. |

Ningún valor secreto debe escribirse en el repositorio, en `render.yaml` ni en
`vercel.json`.

## Smoke test post-deploy

1. `GET https://<api>.onrender.com/health` → `200` con `status: ok` (sin cabecera).
2. Llamada directa a la API sin `X-Frontend-Key` → `401 FRONTEND_KEY_INVALIDA`.
3. Login a través del frontend (`POST /api/auth/login` vía Vercel) → `200` y token.
4. Una operación autenticada a través del frontend (por ejemplo, listar proyectos)
   → `200`.
5. Verificar que la respuesta del proxy no expone la clave y que el navegador no
   incluye `X-Frontend-Key` en las solicitudes salientes.
6. Refrescar una ruta interna del frontend (por ejemplo `/proyectos`) → sirve la SPA
   sin `404`.

## Problemas frecuentes

- `401 FRONTEND_KEY_INVALIDA`: `FRONTEND_KEY` distinta entre Render y Vercel.
- `500 PROXY_MAL_CONFIGURADO`: falta `API_ORIGIN` o `FRONTEND_KEY` en Vercel.
- `502/504` en la primera llamada: cold start del plan Free de Render.
- Health check fallido: la app no escucha en el puerto de `PORT`.
- Sync del Blueprint fallido: `fromDatabase.name` no coincide con el recurso Render.
- Error de conexión a la base: verificar `DB_NAME_PROD` (la base real puede tener
  sufijo) y que el servicio esté en la misma región que la base.

## Límites conocidos

- El plan Free de Postgres vence a los 30 días (con 14 días de gracia).
- El plan Free del servicio web se suspende por inactividad y tarda en despertar.
- El proxy no es un control de acceso: quien conozca la URL de Vercel puede usar
  `/api/...`. La clave solo bloquea el acceso directo a Render.
