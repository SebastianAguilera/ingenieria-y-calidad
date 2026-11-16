# Software Metrics & Estimation

## 📋 Descripción del proyecto

**Software Metrics & Estimation** es el Trabajo Práctico Integrador de la asignatura
**Ingeniería y Calidad de Software (UTN – Facultad Regional San Rafael, 2026)**.

Aplicación para **estimar, planificar, dar seguimiento y medir** proyectos de software.
El núcleo de la solución y sus reglas de negocio están desarrollados en **Go (Golang)**,
con API **REST** expuesta vía HTTP y persistencia.

### Stack tecnológico

| Capa | Tecnología |
| :--- | :--- |
| Lenguaje | Go (go 1.27.1) |
| API | REST / JSON con **Gin** |
| Frontend | **React + TypeScript + Vite + Tailwind CSS + React Router** |
| Persistencia | **PostgreSQL 15** + **GORM** (container `postgres:15.4`) |
| Configuración | `.env` + godotenv (`internal/config`) |
| Pruebas | `go test` (table-driven, TDD) + testify |

### Arquitectura (Clean Architecture)

Dependencias apuntan siempre hacia adentro: `handler → service → repository → domain`.

| Paquete | Responsabilidad |
| :--- | :--- |
| `cmd/api/` | Punto de entrada (`main.go`): bootstrap, wire-up de dependencias y arranque del servidor. |
| `internal/domain/` | Modelos puros (structs + reglas de dominio sin dependencias externas): `Project`, `Story`, `Sprint`, `Estimation`, `Worklog`, `Defect`, `Metric`. |
| `internal/service/` | Lógica de negocio: Planning Poker, cierre de Sprints y **motor de métricas**. |
| `internal/repository/` | Persistencia PostgreSQL/GORM e interfaces de acceso a datos (patrón repositorio). |
| `internal/config/` | Carga de configuración desde `.env` (entorno + credenciales de DB). |
| `internal/handler/` | Controladores REST: parseo, validación de entrada y respuesta JSON. |

---

## 🚀 Cómo correr la app

Antes de arrancar, creá los archivos de entorno a partir de los ejemplos y la red de Docker
(solo la primera vez):
```bash
cp .env.example .env
cp docker/.env.example docker/.env
docker network create mired   # solo si la red no existe
```

### 1. Base de datos (Docker)
La base de datos corre en un contenedor definido en `docker/docker-compose.yml`:
```bash
cd docker
docker compose up -d postgresql
```
Esto levanta solo PostgreSQL en `localhost:5433` con la base `metrics_db`.

> Para desarrollo local levantá únicamente el servicio `postgresql`. Si usás
> `docker compose up -d`, también se inicia `app-servidor` en el puerto `8080` y
> choca con la API local del paso 2.

### 2. Correr la API
Con la base de datos de Docker corriendo, desde la raíz del repo:
```bash
go run ./cmd/api
```
La configuración se toma del `.env` raíz. Para desarrollo contra el contenedor de Docker,
los valores deben coincidir con `docker/.env`:

```env
APP_ENV=development
DB_HOST=localhost
DB_PORT=5433
DB_NAME_DEV=metrics_db
DB_SSLMODE=disable
JWT_SECRET=cambiame-por-un-secreto-largo-y-aleatorio
JWT_EXPIRY=24h
```

`DB_NAME_DEV` debe ser igual a `POSTGRES_DB` (`metrics_db` por defecto) y `DB_PORT` al
puerto publicado por el compose (`5433`). Las tablas se crean automáticamente en el
arranque mediante `AutoMigrate`.

`JWT_SECRET` es obligatoria: sin ella la API no arranca. Firma los tokens de
`POST /api/auth/login` y los que valida el middleware de `/api/*`. `JWT_EXPIRY`
es opcional y define la vigencia de cada token (`24h` por defecto).

### 3. Correr el frontend
```bash
cd frontend
npm install
npm run dev
```
El frontend queda en `http://localhost:5173` y proxya las peticiones `/api` hacia `http://localhost:8080`.

#### Tests del frontend
Los tests usan **Vitest** con entorno `jsdom` y **React Testing Library**. La configuración vive en `frontend/vite.config.ts` (bloque `test`) y el setup global en `frontend/src/test/setup.ts`.
```bash
cd frontend
npm test               # corre la suite una vez (falla si algún test falla)
npm run test:watch     # modo watch para desarrollo
npm run test:coverage  # suite + informe de cobertura (v8)
```
Convenciones:
- Los tests van junto al código, con el sufijo `.test.ts` o `.test.tsx`.
- El módulo de API se mockea con `vi.mock` o `vi.spyOn`; no se llama al backend real.
- Cada historia escribe sus propios tests; no se duplica la configuración de Vitest.
- Antes de subir cambios: `npm test`, `npm run build` y `npm run lint` deben pasar.

### Acceso al panel (login simulado)

La pantalla `/login` pide un **email o nombre** y busca al usuario con `GET /api/usuarios`.
Si existe, lo guarda como usuario activo en `localStorage` y navega a `/proyectos`; las
rutas del panel quedan protegidas por `RutaProtegida`. Si no existe, muestra una tarjeta
para crearlo con `POST /api/usuarios` y luego ingresa automáticamente. El botón "Salir"
del encabezado limpia la sesión y vuelve a `/login`.

> **Importante:** no es autenticación real. No hay contraseña, token ni verificación de
> identidad; el email es único en el backend y el nombre se usa solo como comodidad.
> Cuando exista autenticación real solo cambia cómo se obtiene el usuario activo.

Flujo de prueba manual:
1. Levantar PostgreSQL (`cd docker && docker compose up -d postgresql`) y la API (`go run ./cmd/api`).
2. Levantar el frontend (`cd frontend && npm run dev`) y abrir `http://localhost:5173/login`.
3. Como al principio no hay usuarios, ingresar un email inexistente y crear el usuario.
4. Verificar en la base de datos que el usuario quedó guardado y volver a ingresar con él.

---

## 📌 Guía de Uso y Endpoints (US-01, US-02, US-03 y US-08)

### Referencia de Endpoints REST

#### **POST /api/usuarios** — Registrar usuario independiente (US-08)
* **Ejemplo de Request Body:**
```json
{
  "nombre": "Ana Pérez",
  "email": "dev@example.com"
}
```
* **Respuesta:** `201 Created`. Email único en todo el sistema.

#### **GET /api/usuarios** — Listar usuarios (US-08)
* **Descripción:** Devuelve la lista de usuarios registrados.

#### **POST /api/proyectos** — Crear proyecto (US-08: sin usuarios anidados)
* **Ejemplo de Request Body:**
```json
{
  "nombre": "Proyecto Alfa",
  "fecha_inicio": "2026-10-01",
  "fecha_fin": "2026-12-31"
}
```
* **Respuesta:** `201 Created`. El creador queda vinculado al proyecto como `"LIDER"` (US-11) en la misma transacción: `cantidad_usuarios` vuelve en `1`.

#### **GET /api/proyectos** — Listar proyectos
* **Descripción:** Devuelve la lista de proyectos no dados de baja **del usuario autenticado** (solo donde es miembro), incluyendo `cantidad_usuarios`.

#### **GET /api/proyectos/{id}** — Obtener detalle de un proyecto
* **Descripción:** Devuelve el detalle completo del proyecto junto con el array de `usuarios` asociados. Si el proyecto no existe, está dado de baja o **el usuario no es miembro**, responde `404 PROYECTO_NO_ENCONTRADO` (misma respuesta en ambos casos). Lo mismo vale para todas las rutas anidadas bajo `/api/proyectos/{id}/`.

#### **PUT /api/proyectos/{id}** — Modificar proyecto
* **Descripción:** Actualiza nombre, fecha de inicio y fecha de fin de un proyecto activo.

#### **PATCH /api/proyectos/{id}/estado** — Cambiar estado del proyecto
* **Descripción:** Transiciona el estado del proyecto entre `"Activo"` y `"Cerrado"`.

#### **DELETE /api/proyectos/{id}** — Eliminar proyecto (Baja lógica)
* **Descripción:** Da de baja lógicamente el proyecto si no cuenta con historial asociado.

#### **GET /api/proyectos/{id}/usuarios** — Listar integrantes (US-12)
* **Descripción:** Devuelve el equipo con `id`, `nombre`, `email` y `rol` de cada integrante. Cualquier miembro puede consultarlo; un no miembro recibe `404 PROYECTO_NO_ENCONTRADO`. Un proyecto sin integrantes responde `[]`.
* **Respuesta:** `200 OK` con un array.

#### **POST /api/proyectos/{id}/usuarios** — Agregar integrante por email (US-12)
* **Ejemplo de Request Body:**
```json
{
  "email": "ada@utn.edu.ar",
  "rol": "MIEMBRO"
}
```
* **Roles válidos:** `"LIDER"` y `"MIEMBRO"` (si se omite `rol`, entra como `"MIEMBRO"`). Solo el `LIDER` agrega, cambia roles o quita a otros; un `MIEMBRO` recibe `403 ACCESO_DENEGADO`, salvo para quitarse a sí mismo. No se puede quitar ni degradar al último líder (`409 ULTIMO_LIDER`). Ver `docs/specs/US-11-roles-y-permisos-spec.md` y `docs/specs/US-12-integrantes-por-email-spec.md`.
* **Respuesta:** `200 OK`. El email debe pertenecer a un usuario registrado (`404 USUARIO_NO_ENCONTRADO` si no existe, sin crearlo); si ya es integrante, `409 USUARIO_DUPLICADO`.

#### **PATCH /api/proyectos/{id}/usuarios/{usuarioId}** — Cambiar rol (US-12)
* **Ejemplo de Request Body:** `{ "rol": "LIDER" }`
* **Respuesta:** `200 OK` con `{ "usuario_id": 5, "rol": "LIDER" }`. Solo `LIDER`; rol inválido `422`; último líder `409 ULTIMO_LIDER`.

#### **DELETE /api/proyectos/{id}/usuarios/{usuarioId}** — Quitar usuario del proyecto
* **Descripción:** Desasocia un usuario del equipo del proyecto. Solo `LIDER` o auto-baja (`204`).

### Product Backlog (US-02)

#### **POST /api/proyectos/{id}/historias** — Agregar historia al backlog
* **Descripción:** Crea una historia de usuario. El código `US-NN` es consecutivo dentro del proyecto y no editable.
* **Ejemplo de Request Body:**
```json
{
  "titulo": "Como Product Builder quiero gestionar el backlog",
  "descripcion": "Cubrir el alcance del proyecto",
  "prioridad": "Alta",
  "estado": "Backlog"
}
```
* `prioridad` (`"Alta"`, `"Media"`, `"Baja"`) y `estado` (`"Backlog"`, `"En Progreso"`, `"Hecho"`) son opcionales; por defecto `"Media"` y `"Backlog"`.
* **Estimación y priorización (US-23):** `story_points` usa la escala `0, 0.5, 1, 2, 3, 5, 8, 13, 21`. `valor_negocio`, `criticidad_tiempo` y `reduccion_riesgo` son opcionales y usan la escala `1, 2, 3, 5, 8, 13, 21`; `null` significa "no informado". Los tres valores se aceptan tanto en el alta como en el PATCH.
  ```json
  {
    "titulo": "Como Product Builder quiero priorizar el backlog",
    "story_points": 5,
    "valor_negocio": 8,
    "criticidad_tiempo": 3,
    "reduccion_riesgo": 2
  }
  ```
* **Errores:** un valor fuera de escala o no numérico responde `422 VALIDACION` con `campo` igual a `story_points`, `valor_negocio`, `criticidad_tiempo` o `reduccion_riesgo`.

#### **GET /api/proyectos/{id}/historias** — Listar el backlog del proyecto
* **Descripción:** Devuelve `{ "total": n, "historias": [...] }` en orden de alta. Cada historia incluye `story_points`, `valor_negocio`, `criticidad_tiempo` y `reduccion_riesgo`; los no informados se emiten como `null`.

#### **PATCH /api/proyectos/{id}/historias/{historiaId}** — Modificar historia
* **Descripción:** Actualización parcial de `titulo`, `descripcion`, `prioridad`, `estado`, `story_points`, `valor_negocio`, `criticidad_tiempo` y/o `reduccion_riesgo`. Es obligatorio informar al menos un campo.

#### **GET /api/proyectos/{id}/backlog/priorizado** — Priorizar backlog con ROI y WSJF (US-23)
* **Descripción:** Devuelve las historias del proyecto excluyendo las que están en estado `"Hecho"`, ordenadas por `wsjf` descendente. Las historias sin `wsjf` quedan al final; los empates se resuelven por `id` ascendente.
* **Fórmulas:**
  * `roi = valor_negocio / story_points`
  * `wsjf = (valor_negocio + criticidad_tiempo + reduccion_riesgo) / story_points`
* **Nulos:** los valores ausentes cuentan como `0` en el numerador de WSJF. `roi` y `wsjf` se emiten como `null` si falta algún dato necesario o si `story_points` es `null` o `0`.
* **Respuesta:** `200 OK` con `{ "total": n, "historias": [...] }`. Cada elemento incluye los campos habituales de la historia más `roi` y `wsjf`.
  ```json
  {
    "total": 1,
    "historias": [
      {
        "id": 15,
        "codigo": "US-01",
        "titulo": "Como Product Builder quiero priorizar el backlog",
        "estado": "Backlog",
        "story_points": 5,
        "valor_negocio": 8,
        "criticidad_tiempo": 3,
        "reduccion_riesgo": 2,
        "roi": 1.6,
        "wsjf": 2.6
      }
    ]
  }
  ```
* **Errores:** `404 PROYECTO_NO_ENCONTRADO` si el proyecto no existe o el usuario no es miembro.

### Sprints (US-03)

#### **POST /api/proyectos/{id}/sprints** — Crear sprint
* **Descripción:** Da de alta un sprint en estado `Planificado`. El `identificador` (1–50 caracteres) debe ser único dentro del proyecto y las fechas (`YYYY-MM-DD`, ambas obligatorias) no pueden solaparse con otros sprints del mismo proyecto (fechas contiguas sí están permitidas).
* **Ejemplo de Request Body:**
```json
{
  "identificador": "Sprint 1",
  "sprint_goal": "Entregar el MVP del motor de métricas",
  "fecha_inicio": "2026-10-06",
  "fecha_fin": "2026-10-20"
}
```
* **Respuesta:** `201 Created` con el sprint creado (`id`, `proyecto_id`, `identificador`, `sprint_goal`, `estado`, `fecha_inicio`, `fecha_fin`, `creado_en`, `actualizado_en`).

#### **GET /api/proyectos/{id}/sprints** — Listar sprints del proyecto
* **Descripción:** Devuelve `{ "total": n, "sprints": [...] }` ordenados por `fecha_inicio` ascendente.

#### **PATCH /api/proyectos/{id}/sprints/{sprintId}/estado** — Cambiar estado del sprint
* **Descripción:** Transiciones válidas: `Planificado → Activo → Cerrado`. Solo puede haber un sprint `Activo` por proyecto y `Cerrado` es un estado terminal.
* **Ejemplo de Request Body:**
```json
{
  "estado": "Activo"
}
```
* **Errores típicos:** `400 ESTADO_INVALIDO` (literal fuera del conjunto), `404 SPRINT_NO_ENCONTRADA`, `409 SPRINT_TRANSICION_INVALIDA` / `SPRINT_ACTIVO_EXISTENTE`, `409 SPRINT_FECHAS_SOLAPADAS` / `SPRINT_DUPLICADO`, `422 VALIDACION` (identificador o sprint_goal inválidos).

### Defectos (US-09)

#### **POST /api/proyectos/{id}/defectos** — Registrar defecto
* **Descripción:** Registra un defecto asociado a una historia y a un Sprint de detección. Todo defecto nace en estado `open`.
* **Ejemplo de Request Body:**
```json
{
  "descripcion": "El cálculo de velocidad falla con sprints vacíos",
  "severidad": "major",
  "historia_id": 3,
  "sprint_deteccion_id": 1,
  "reportado_por": 5
}
```
* **Severidades válidas:** `"minor"`, `"major"`, `"critical"`, `"blocker"`. `reportado_por` es opcional.
* **Respuesta:** `201 Created` con el defecto creado.

#### **GET /api/proyectos/{id}/defectos** — Listar defectos del proyecto
* **Descripción:** Devuelve `{ "total": n, "defectos": [...] }` en orden de alta. Un proyecto sin defectos devuelve `"defectos": []`.

#### **PATCH /api/proyectos/{id}/defectos/{defectoId}** — Actualizar defecto
* **Descripción:** Actualización parcial de `estado` y/o `sprint_resolucion_id`. Es obligatorio informar al menos un campo.
* **Ejemplo de Request Body:**
```json
{
  "estado": "resolved",
  "sprint_resolucion_id": 2
}
```
* **Estados válidos:** `"open"`, `"in_progress"`, `"resolved"`, `"closed"`, `"rejected"`. Los estados `resolved` y `closed` exigen `sprint_resolucion_id`; reabrir un defecto limpia su resolución.
* **Errores típicos:** `400 ESTADO_INVALIDO` / `PARAMETRO_INVALIDO` / `JSON_INVALIDO`, `404 DEFECTO_NO_ENCONTRADO` / `PROYECTO_NO_ENCONTRADO` / `HISTORIA_NO_ENCONTRADA` / `SPRINT_NO_ENCONTRADA` / `USUARIO_NO_ENCONTRADO`, `422 VALIDACION`.

---

### Motor de métricas (core en Go)

| Métrica | Fórmula |
| :--- | :--- |
| Velocidad del equipo | `Σ StoryPoints completados` / `nro. de Sprints cerrados` |
| Desviación de esfuerzo | `(Horas reales − Horas estimadas) / Horas estimadas × 100` |
| % historias completadas | `Historias completadas / Historias planificadas × 100` |
| Ratio de defectos | `Defectos resueltos / Defectos detectados` (al Sprint de detección) |

### Metodologías de desarrollo (obligatorias)

| Metodología | Uso |
| :--- | :--- |
| **Scrum** | Sprints 0..4; roles del sistema: Líder (`LIDER`) y Miembro (`MIEMBRO`). |
| **SDD** | Especificaciones versionadas en el repo (objetivo, entradas, salidas, reglas, restricciones, casos límite, errores, criterios de aceptación). |
| **BDD** | Escenarios `Given – When – Then` (normales, alternativos, límite, errores), automatizados cuando sea posible. |
| **TDD** | Ciclo RED → GREEN → REFACTOR, evidencia vía historial de commits. |
| **IA** | Soporte en análisis, especificaciones, código, pruebas y documentación (todo revisado por el equipo). |

### Trazabilidad

Historia de Usuario → Especificación SDD → Criterios de Aceptación → Escenarios BDD → Tests → Código Go.

## ⚙️ Integración continua (CI)

El workflow vive en [`.github/workflows/ci.yml`](.github/workflows/ci.yml) y se ejecuta en
cada `push` y `pull_request` hacia `develop` o `master`. Tiene dos jobs independientes.

### Job `backend` (Go)

Corre sobre `ubuntu-latest` con un servicio **PostgreSQL 16** (`postgres:16`) en el puerto
`5432`, con usuario `test`, contraseña `test` y base `metrics_db_test`.

Pasos, en orden:

1. `go vet ./...`
2. `go test ./... -v`, con la salida guardada en `test.log`. Si el log contiene `--- SKIP`,
   el paso **falla**: un test de integración saltado no cuenta como verde.
3. Subida de `test.log` como artifact `reporte-tests-go` (14 días). Corre aunque los tests
   fallen (`if: always()`), así el reporte queda disponible igual.
4. `go build ./cmd/api`

Variables del job: `APP_ENV=development`, `DB_HOST=localhost`, `DB_PORT=5432`,
`DB_USER=test`, `DB_PASSWORD=test`, `DB_NAME_TEST=metrics_db_test`, `DB_SSLMODE=disable`,
`JWT_SECRET=ci-only-secret`. Son valores de CI, no credenciales reales.

> Los tests de `internal/repository` usan `DB_NAME_TEST`: el helper exige que el nombre
> termine en `_test`, porque hace `DROP TABLE`. Por eso el servicio crea `metrics_db_test`.

### Job `frontend` (React)

Corre en `frontend/` con Node 22 y la caché de npm:

1. `npm ci`
2. `npm run lint`
3. `npm test`
4. `npm run build`

### Correr lo mismo en local

Con el contenedor de Docker arriba (`docker compose up -d postgresql` desde `docker/`),
y con la base `metrics_db_test` creada, el backend se prueba así:

```powershell
$env:DB_HOST="localhost"; $env:DB_PORT="5433"; $env:DB_USER="postgres"; $env:DB_PASSWORD="postgres"; $env:DB_SSLMODE="disable"; $env:DB_NAME_TEST="metrics_db_test"
go test ./... -v -count=1 -timeout 30m 2>&1 | Tee-Object -FilePath test.log
Select-String -Path test.log -Pattern "--- SKIP"   # no debe devolver nada
```

Con las mismas variables pero en `bash`, usá `VAR=valor` delante de `go test`. Verificado:
106 PASS, 0 SKIP en `internal/repository`.

> Si la base no existe o el contenedor está apagado, los tests de integración se saltan
> (`--- SKIP`). El CI lo convierte en fallo; en local conviene revisar la salida.

## 🚀 Deploy (Render + Vercel)

La API se despliega en **Render** (Docker + Postgres existente `metricas-db`) y el
frontend en **Vercel**. El navegador llama directamente a la API en Render
(`VITE_API_BASE_URL`), con CORS restringido a `https://ingenieria-y-calidad.vercel.app`
en producción. Login y registro tienen rate limiting (20 solicitudes/min por IP).

- Guía completa: [`docs/deploy.md`](docs/deploy.md).
- Blueprint de la API: [`render.yaml`](render.yaml) (no declara la base; la referencia
  con `fromDatabase`).
- Configuración de Vercel: [`frontend/vercel.json`](frontend/vercel.json).
- `VITE_API_BASE_URL` es variable de build: tras cargarla en Vercel hay que hacer
  **Redeploy**.

Variables (solo nombres):

| Variable | Servicio | ¿Secreto? |
| --- | --- | --- |
| `APP_ENV` | Render | No |
| `PORT` | Render | No |
| `DB_HOST` | Render | No |
| `DB_PORT` | Render | No |
| `DB_USER` | Render | No |
| `DB_PASSWORD` | Render | Sí |
| `DB_NAME_PROD` | Render | No |
| `DB_SSLMODE` | Render | No |
| `JWT_SECRET` | Render | Sí |
| `CORS_ALLOWED_ORIGINS` | Render | No |
| `AUTH_RATE_LIMIT_MAX` | Render | No |
| `AUTH_RATE_LIMIT_WINDOW` | Render | No |
| `VITE_API_BASE_URL` | Vercel (build) | No |

> No escribir valores secretos en el repositorio ni en `render.yaml`. `VITE_*` es
> público por diseño.

Checklist de smoke test post-deploy:

1. `GET https://<api>.onrender.com/health` responde `200`.
2. Preflight desde `https://ingenieria-y-calidad.vercel.app` responde `204`; desde otro
   origen responde `403 ORIGEN_NO_PERMITIDO`.
3. Login desde el frontend responde `200`.
4. Exceder el límite de login responde `429` con `Retry-After`.
5. Refrescar una ruta interna de la SPA no devuelve `404`.

## 👥 Integrantes del Grupo

* Aguilera Sebastián - Agile Enabler
* Aguilera Rocio - Product Builders
* Chang Yang Gabriela - Product Builders
* Choquevillca Celeste - Product Builders
* Perez Castro Jazmín - Product Builders
