# Especificación: US-04 - Asignación de historias al Sprint Backlog

## Objetivo

Implementar la asignación y desasignación de historias de usuario desde el Product Backlog hacia el Sprint Backlog de un Sprint activo. Permite al Product Builder seleccionar el trabajo a realizar en el ciclo, garantizando integridad entre historias, Sprints y reglas de negocio. Constituye la base para el seguimiento de trabajo y métricas (US-06, US-07).

La funcionalidad se expone mediante endpoints REST servidos por Gin, con reglas de negocio en Go, persistencia en PostgreSQL mediante GORM y esquema generado por `AutoMigrate`.

**Trazabilidad:** Issue #31 → Esta especificación → Criterios de aceptación → Escenarios BDD → Tests Go (TDD) → Código.

## Alcance

**Incluye:**
- Asignar una historia en estado `Backlog` a un Sprint `Activo`
- Desasignar una historia del Sprint Backlog si el Sprint no está `Cerrado`
- Validación de ventana temporal del Sprint
- Garantía de que una historia solo pertenezca a un Sprint
- Registro de fecha/hora de asignación
- Respeto del orden dentro del Sprint Backlog

**No incluye:**
- Reordenamiento manual del Sprint Backlog (fuera de alcance)
- Cambio de estado de la historia al asignar/desasignar (se mantiene su estado)
- Asignación masiva
- Cierre de Sprint con métricas (US-07)

## Entradas

### Precondiciones
- Esquema de US-01 (proyectos), US-02 (historias) y US-03 (sprints) existentes
- Proyecto y Sprint referenciados existen
- Historia de usuario existe en el proyecto

### Convenciones
| Convención | Regla | Consecuencia |
| :--- | :--- | :--- |
| Identificadores de recurso | Entero positivo en path | `400 PARAMETRO_INVALIDO` si no es numérico o 0 |
| Estado de historia | Solo `Backlog` puede asignarse | `400 HISTORIA_NO_PENDIENTE` |
| Estado de Sprint | Debe estar `Activo` para asignar | `400 SPRINT_NO_ACTIVO` |
| Sprint cerrado | No permite desasignar | `409 SPRINT_CERRADO` |
| Historia ya asignada | No puede asignarse a otro Sprint | `409 HISTORIA_YA_ASIGNADA` |
| Ventana temporal | Asignación debe ocurrir dentro de `FechaInicio`-`FechaFin` del Sprint | `400 FUERA_DE_VENTANA_TEMPORAL` |

### Cuerpos de request

#### `POST /api/proyectos/{id}/sprints/{sprintId}/historias` - Asignar historia
```json
{
  "historia_id": 100
}
```

#### `DELETE /api/proyectos/{id}/sprints/{sprintId}/historias/{historiaId}` - Desasignar historia
Sin cuerpo.

## Salidas esperadas

### Endpoints
| Método | Path | Descripción | Éxito | Errores |
| :--- | :--- | :--- | :---: | :---: |
| `POST` | `/api/proyectos/{id}/sprints/{sprintId}/historias` | Asigna historia al Sprint Backlog | `204 No Content` | `400`, `404`, `409` |
| `DELETE` | `/api/proyectos/{id}/sprints/{sprintId}/historias/{historiaId}` | Desasigna historia del Sprint Backlog | `204 No Content` | `400`, `404`, `409` |

### Cuerpo estándar de error
```json
{
  "error": "CODIGO",
  "mensaje": "Descripción en español",
  "campo": "campo opcional"
}
```

## Modelo de datos

### Dominio
```go
type SprintBacklogItem struct {
    ID              uint
    ProyectoID      uint
    SprintID        uint
    HistoriaID      uint
    Orden           int
    FechaAsignacion time.Time
    CreadoEn        time.Time
    ActualizadoEn   time.Time
}
```

### Errores centinela
```go
ErrHistoriaYaAsignada         = errors.New("la historia ya está asignada a un Sprint")
ErrSprintNoActivo             = errors.New("el Sprint debe estar activo")
ErrSprintCerrado              = errors.New("no se pueden modificar Sprints cerrados")
ErrHistoriaNoPendiente        = errors.New("solo se pueden asignar historias en estado pendiente")
ErrFueraDeVentanaTemporal     = errors.New("la asignación está fuera de la ventana temporal del Sprint")
ErrHistoriaNoPerteneceASprint = errors.New("la historia no pertenece al sprint indicado")
ErrHistoriaYaEnSprintBacklog  = errors.New("la historia ya forma parte del Sprint Backlog")
```

### Persistencia
Tabla `sprint_backlog_items`:
- `id` (PK)
- `proyecto_id` (FK -> proyectos, NOT NULL)
- `sprint_id` (FK -> sprints, NOT NULL)
- `historia_id` (FK -> historias, NOT NULL, **UNIQUE**)
- `orden` (INT NOT NULL DEFAULT 0)
- `fecha_asignacion` (TIMESTAMPTZ NOT NULL)
- `creado_en`, `actualizado_en` (TIMESTAMPTZ)

Índices: `idx_sprint_backlog_proyecto`, `idx_sprint_backlog_sprint`, `ux_sprint_backlog_historia` (UNIQUE)

## Reglas de negocio

| ID | Regla |
| :--- | :--- |
| RN-01 | Solo se pueden asignar historias en estado `Backlog` |
| RN-02 | Una historia no puede estar asignada a más de un Sprint simultáneamente |
| RN-03 | Solo se puede asignar a Sprints en estado `Activo` |
| RN-04 | La asignación debe realizarse dentro de la ventana temporal del Sprint (entre FechaInicio y FechaFin) |
| RN-05 | Al asignar se conservan los StoryPoints de la historia |
| RN-06 | Al asignar NO se cambia el estado de la historia a "En Progreso" |
| RN-07 | No se puede desasignar historias de un Sprint `Cerrado` |
| RN-08 | Al desasignar, la historia vuelve al Product Backlog en estado `Backlog` |
| RN-09 | Se registra la fecha/hora de asignación para trazabilidad |
| RN-10 | El orden se asigna automáticamente (máximo + 1) |

## Casos límite
| ID | Caso | Resultado |
| :--- | :--- | :--- |
| CL-01 | Primer elemento en Sprint Backlog vacío | Orden = 1 |
| CL-02 | Historia ya asignada a otro Sprint | `409 HISTORIA_YA_ASIGNADA` |
| CL-03 | Historia en estado `En Progreso` | `400 HISTORIA_NO_PENDIENTE` |
| CL-04 | Historia en estado `Hecho` | `400 HISTORIA_NO_PENDIENTE` |
| CL-05 | Asignar a Sprint `Planificado` | `400 SPRINT_NO_ACTIVO` |
| CL-06 | Asignar a Sprint `Cerrado` | `400 SPRINT_NO_ACTIVO` |
| CL-07 | Asignar antes de FechaInicio | `400 FUERA_DE_VENTANA_TEMPORAL` |
| CL-08 | Asignar después de FechaFin | `400 FUERA_DE_VENTANA_TEMPORAL` |
| CL-09 | Desasignar de Sprint `Cerrado` | `409 SPRINT_CERRADO` |
| CL-10 | Desasignar historia que no pertenece al Sprint | `409 HISTORIA_NO_PERTENECE_A_SPRINT` |

## Criterios de aceptación

### Funcionales
- **CA-F1:** Se puede asignar una historia en estado `Backlog` a un Sprint `Activo` dentro de su ventana temporal
- **CA-F2:** La historia asignada forma parte del Sprint Backlog y deja de estar disponible para otros Sprints
- **CA-F3:** Se puede desasignar una historia de un Sprint no cerrado, devolviéndola al Product Backlog
- **CA-F4:** Se registra fecha de asignación y orden dentro del Sprint

### Reglas de negocio
- **CA-R1:** Solo historias en `Backlog` pueden asignarse (RN-01)
- **CA-R2:** Historia única por Sprint (RN-02)
- **CA-R3:** Sprint debe estar `Activo` (RN-03)
- **CA-R4:** Respeta ventana temporal (RN-04)
- **CA-R5:** Conserva StoryPoints y no cambia estado (RN-05, RN-06)
- **CA-R6:** No permite desasignar de Sprint cerrado (RN-07)

## Mapeo de errores HTTP
| Error | Código HTTP | Código API |
| :--- | :---: | :--- |
| `ErrHistoriaYaAsignada` / `ErrHistoriaYaEnSprintBacklog` | 409 | `HISTORIA_YA_ASIGNADA` |
| `ErrSprintNoActivo` | 400 | `SPRINT_NO_ACTIVO` |
| `ErrSprintCerrado` | 409 | `SPRINT_CERRADO` |
| `ErrHistoriaNoPendiente` | 400 | `HISTORIA_NO_PENDIENTE` |
| `ErrFueraDeVentanaTemporal` | 400 | `FUERA_DE_VENTANA_TEMPORAL` |
| `ErrHistoriaNoPerteneceASprint` | 409 | `HISTORIA_NO_PERTENECE_A_SPRINT` |
| `ErrHistoriaNoEncontrada` | 404 | `HISTORIA_NO_ENCONTRADA` |
| `ErrSprintNoEncontrado` | 404 | `SPRINT_NO_ENCONTRADA` |
| `ErrProyectoNoEncontrado` | 404 | `PROYECTO_NO_ENCONTRADO` |

## Arquitectura
Respeta Clean Architecture: `domain` (entidades/puertos/errores) → `repository` (GORM/Postgres) → `service` (reglas de negocio) → `handler` (HTTP/Gin). Dominio sin dependencias de infraestructura.
