# Software Metrics & Estimation

##  Descripcion del proyecto

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
| `internal/config/` | Carga de configuración desde `.env` (entorno + credenciales de DB). |
| `internal/handler/` | Controladores REST: parseo, validación de entrada y respuesta JSON. |

### Motor de métricas (core en Go)

| Métrica | Fórmula |
| :--- | :--- |
| Velocidad del equipo | `Σ StoryPoints completados` / `nro. de Sprints cerrados` |
| Desviación de esfuerzo | `(Horas reales − Horas estimadas) / Horas estimadas × 100` |
| % historias completadas | `Historias completadas / Historias planificadas × 100` |
| Ratio de defectos | `Defectos resueltos / Defectos detectados` (al Sprint de detección) |

### Modelo de dominio (entidades clave)

- **Project** – Nombre, estado, integrantes, fechas de inicio/fin.
- **Story** – Título, descripción, prioridad, estado, Story Points, criterios de aceptación; asociada a Sprint y Backlog.
- **Sprint** – Sprint Goal, historias asignadas, estado (abierto/cerrado), fechas.
- **Estimation / PlanningPokerRound** – Votos ocultos por integrante, detección de dispersión, rondas y estimación acordada.
- **Worklog** – Integrante, fecha, actividad, horas reales.
- **Defect** – Descripción, severidad, estado, historia relacionada, Sprints de detección y resolución.

### Funcionalidades principales

1. **Gestión de proyectos** – CRUD, integrantes, fechas y estado.
2. **Product Backlog** – Historias con identificador, prioridad, estado, Story Points y criterios de aceptación.
3. **Gestión de Sprints** – Sprint Goal, asignación/completado de historias, cierre y consulta de Sprints anteriores.
4. **Estimación** – Story Points y **Planning Poker** (votos ocultos → revelado → detección de dispersión → rondas → acuerdo).
5. **Registro de esfuerzo** – Worklogs para comparar esfuerzo estimado vs. real.
6. **Gestión de defectos** – Registro y seguimiento por severidad/estado/Sprint.
7. **Métricas** – Cálculo automático sobre los datos registrados.
8. **Dashboard** – Estado del proyecto y representaciones gráficas de métricas.
9. **Reportes** – Reporte de proyecto/Sprint exportable a **PDF**.

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

## Cómo correr la app

### 1. Base de datos (Docker)

La base de datos corre en un contenedor definido en `docker/docker-compose.yml`:

```bash
cd docker
docker compose up -d
```

Queda disponible en `localhost:5433` y se enciende sola al abrir Docker Desktop.

### 2. Correr la API

**a) En Docker:** ya quedó levantada con el paso 1 (servicio `app`).
Navegador → http://localhost:8080/health

**b) En local:** con la DB de Docker corriendo:

```bash
go run cmd/api/main.go
```

Se conecta a `localhost:5433` y usa la base que indica `APP_ENV` en el `.env` raíz.

### Entornos

| `APP_ENV` | Base de datos |
| :--- | :--- |
| development | `metrics_db_dev` |
| test | `metrics_db_test` |
| production | `metrics_db` |

## Integrantes del Grupo

* Aguilera Sebastián - Agile Enabler
* Aguilera Rocio - Product Builders
* Chang Yang Gabriela - Product Builders
* Choquevillca Celeste - Product Builders
* Perez Castro Jazmín - Product Builders

