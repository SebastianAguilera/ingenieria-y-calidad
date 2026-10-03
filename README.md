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
| Persistencia | **PostgreSQL 16** + **GORM** (docker-compose) |
| Configuración | `.env` + godotenv |
| Pruebas | `go test` (table-driven, TDD) + testify |

### Arquitectura (Clean Architecture)

Dependencias apuntan siempre hacia adentro: `handler → service → repository → domain`.

| Paquete | Responsabilidad |
| :--- | :--- |
| `cmd/api/` | Punto de entrada (`main.go`): bootstrap, wire-up de dependencias y arranque del servidor. |
| `internal/domain/` | Modelos puros (structs + reglas de dominio sin dependencias externas): `Project`, `Story`, `Sprint`, `Estimation`, `Worklog`, `Defect`, `Metric`. |
| `internal/service/` | Lógica de negocio: Planning Poker, cierre de Sprints y **motor de métricas**. |
| `internal/repository/` | Persistencia PostgreSQL/GORM e interfaces de acceso a datos (patrón repositorio). |
| `internal/handler/` | Controladores REST: parseo, validación de entrada y respuesta JSON. |

---

## 🚀 Guía de Uso y Endpoints (US-01)

### 1. Ejecución con Docker
Para levantar la base de datos PostgreSQL y la API automáticamente:
```bash
docker compose up --build
```
La aplicación quedará disponible en `http://localhost:8080`. Las tablas (`proyectos`, `usuarios`, `proyecto_usuarios`) se crean automáticamente en el arranque mediante `AutoMigrate` de GORM.

---

### 2. Referencia de Endpoints REST

#### **POST /api/proyectos** — Crear proyecto con usuarios iniciales
* **Descripción:** Da de alta un proyecto con fecha de inicio, fecha de fin opcional y lista inicial de usuarios.
* **Ejemplo de Request Body:**
```json
{
  "nombre": "Proyecto Alfa",
  "fecha_inicio": "2026-10-01",
  "fecha_fin": "2026-12-31",
  "usuarios": [
    { "nombre": "Usuario Uno", "email": "user1@example.com" },
    { "nombre": "Usuario Dos", "email": "user2@example.com" }
  ]
}
```
* **Respuesta exitosa:** `201 Created` con el detalle del proyecto creado.

---

#### **GET /api/proyectos** — Listar proyectos
* **Descripción:** Devuelve la lista de proyectos no dados de baja, incluyendo `cantidad_usuarios` (sin anidar el array de usuarios por diseño D-05).
* **Ejemplo de Respuesta:** `200 OK`
```json
{
  "total": 1,
  "proyectos": [
    {
      "id": 1,
      "nombre": "Proyecto Alfa",
      "fecha_inicio": "2026-10-01",
      "fecha_fin": "2026-12-31",
      "estado": "Activo",
      "cantidad_usuarios": 2,
      "creado_en": "2026-10-03T20:00:00Z",
      "actualizado_en": "2026-10-03T20:00:00Z"
    }
  ]
}
```

---

#### **GET /api/proyectos/{id}** — Obtener detalle de un proyecto
* **Descripción:** Devuelve el detalle completo del proyecto junto con el array de `usuarios` asociados.
* **Respuesta exitosa:** `200 OK`

---

#### **PUT /api/proyectos/{id}** — Modificar proyecto
* **Descripción:** Actualiza nombre, fecha de inicio y fecha de fin de un proyecto activo.
* **Ejemplo de Request Body:**
```json
{
  "nombre": "Proyecto Alfa Actualizado",
  "fecha_inicio": "2026-10-01",
  "fecha_fin": "2027-01-15"
}
```

---

#### **PATCH /api/proyectos/{id}/estado** — Cambiar estado del proyecto
* **Descripción:** Transiciona el estado del proyecto entre `"Activo"` y `"Cerrado"`.
* **Ejemplo de Request Body:**
```json
{
  "estado": "Cerrado"
}
```

---

#### **DELETE /api/proyectos/{id}** — Eliminar proyecto (Baja lógica)
* **Descripción:** Da de baja lógicamente el proyecto si no cuenta con historial asociado.
* **Respuesta exitosa:** `204 No Content`

---

#### **POST /api/proyectos/{id}/usuarios** — Agregar usuario al proyecto
* **Descripción:** Asocia un usuario al equipo del proyecto. Si el email ya existe en el sistema, reutiliza el registro.
* **Ejemplo de Request Body:**
```json
{
  "nombre": "Usuario Tres",
  "email": "user3@example.com"
}
```
* **Respuesta exitosa:** `201 Created`

---

#### **DELETE /api/proyectos/{id}/usuarios/{usuarioId}** — Quitar usuario del proyecto
* **Descripción:** Desasocia un usuario del equipo del proyecto.
* **Respuesta exitosa:** `204 No Content`

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

## 👥 Usuarios del Grupo

* Aguilera Sebastián - Agile Enabler
* Aguilera Rocio - Product Builders
* Chang Yang Gabriela - Product Builders
* Choquevillca Celeste - Product Builders
* Perez Castro Jazmín - Product Builders
