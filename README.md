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
* **Respuesta:** `201 Created`. El proyecto se crea sin miembros asociados inicialmente.

#### **GET /api/proyectos** — Listar proyectos
* **Descripción:** Devuelve la lista de proyectos no dados de baja, incluyendo `cantidad_usuarios`.

#### **GET /api/proyectos/{id}** — Obtener detalle de un proyecto
* **Descripción:** Devuelve el detalle completo del proyecto junto con el array de `usuarios` asociados.

#### **PUT /api/proyectos/{id}** — Modificar proyecto
* **Descripción:** Actualiza nombre, fecha de inicio y fecha de fin de un proyecto activo.

#### **PATCH /api/proyectos/{id}/estado** — Cambiar estado del proyecto
* **Descripción:** Transiciona el estado del proyecto entre `"Activo"` y `"Cerrado"`.

#### **DELETE /api/proyectos/{id}** — Eliminar proyecto (Baja lógica)
* **Descripción:** Da de baja lógicamente el proyecto si no cuenta con historial asociado.

#### **POST /api/proyectos/{id}/usuarios** — Asociar usuario existente con rol (US-08)
* **Ejemplo de Request Body:**
```json
{
  "usuario_id": 5,
  "rol": "Constructor de Producto"
}
```
* **Roles válidos:** `"Constructor de Producto"`, `"Arquitecto de Producto"`, `"Facilitador Ágil"`.
* **Respuesta:** `200 OK`. El usuario debe estar registrado (`404` si no existe). Rol obligatorio (`422` si inválido).

#### **DELETE /api/proyectos/{id}/usuarios/{usuarioId}** — Quitar usuario del proyecto
* **Descripción:** Desasocia un usuario del equipo del proyecto.

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

#### **GET /api/proyectos/{id}/historias** — Listar el backlog del proyecto
* **Descripción:** Devuelve `{ "total": n, "historias": [...] }` en orden de alta.

#### **PATCH /api/proyectos/{id}/historias/{historiaId}** — Modificar historia
* **Descripción:** Actualización parcial de `titulo`, `descripcion`, `prioridad` y/o `estado`. Es obligatorio informar al menos un campo.

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
| **Scrum** | Sprints 0..4; roles: Product Architect, Agile Enabler, Product Builders. |
| **SDD** | Especificaciones versionadas en el repo (objetivo, entradas, salidas, reglas, restricciones, casos límite, errores, criterios de aceptación). |
| **BDD** | Escenarios `Given – When – Then` (normales, alternativos, límite, errores), automatizados cuando sea posible. |
| **TDD** | Ciclo RED → GREEN → REFACTOR, evidencia vía historial de commits. |
| **IA** | Soporte en análisis, especificaciones, código, pruebas y documentación (todo revisado por el equipo). |

### Trazabilidad

Historia de Usuario → Especificación SDD → Criterios de Aceptación → Escenarios BDD → Tests → Código Go.

## 👥 Integrantes del Grupo

* Aguilera Sebastián - Agile Enabler
* Aguilera Rocio - Product Builders
* Chang Yang Gabriela - Product Builders
* Choquevillca Celeste - Product Builders
* Perez Castro Jazmín - Product Builders
