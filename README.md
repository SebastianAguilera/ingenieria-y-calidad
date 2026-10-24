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

### 1. Base de datos (Docker)
La base de datos corre en un contenedor definido en `docker/docker-compose.yml`:
```bash
cd docker
docker compose up -d
```

### 2. Correr la API
Con la base de datos de Docker corriendo:
```bash
go run cmd/api/main.go
```
La aplicación se conecta según el entorno indicado en `APP_ENV` en el `.env` raíz (por defecto `development` usando `metrics_db_dev`). Las tablas se crean automáticamente en el arranque mediante `AutoMigrate`.

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
