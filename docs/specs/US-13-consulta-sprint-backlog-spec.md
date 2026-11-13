# Especificación: US-13 - Consultar las historias asignadas a un sprint

## Objetivo

Exponer un read model del Sprint Backlog para que cualquier miembro del
proyecto consulte el trabajo planificado de un Sprint, con sus historias
anidadas, el orden y la fecha de asignación. Complementa a US-04, que ya
permite asignar y desasignar historias pero no consultarlas.

La funcionalidad se expone mediante un endpoint REST servido por Gin, con la
lógica de lectura en Go, persistencia en PostgreSQL mediante GORM y el esquema
generado por `AutoMigrate`.

**Trazabilidad:** Issue #71 → Esta especificación → Criterios de aceptación →
Escenarios BDD → Tests Go (TDD) → Código.

## Alcance

**Incluye:**
- `GET /api/proyectos/{id}/sprints/{sprintId}/historias`.
- Sprint solicitado con la representación de US-03.
- Historias anidadas con la representación de US-02.
- Orden de los items por `orden ASC`, desempatando por `id ASC`.
- Lectura de Sprints `Planificado`, `Activo` y `Cerrado`.
- Sprint sin historias devuelve `200` con `total: 0` e `items: []`.
- Sin paginación.

**No incluye:**
- Asignar o desasignar historias (US-04).
- Modificar el cierre de Sprint ni devolver historias al Product Backlog.
- Reordenamiento manual del Sprint Backlog.
- Métricas o reportes.

## Entradas

### Precondiciones
- Esquema de US-01 (proyectos), US-02 (historias), US-03 (sprints) y US-04
  (sprint_backlog_items) existente.
- El usuario autenticado es miembro del proyecto (guard de US-F10).
- El Sprint pertenece al proyecto indicado.

### Convenciones
| Convención | Regla | Consecuencia |
| :--- | :--- | :--- |
| Identificadores de recurso | Entero positivo en path | `400 PARAMETRO_INVALIDO` si no es numérico o 0 |
| Membresía | Cualquier miembro puede consultar | `404 PROYECTO_NO_ENCONTRADO` si no es miembro o el proyecto no existe |
| Sprint | Debe existir en el proyecto | `404 SPRINT_NO_ENCONTRADA` si no existe o es de otro proyecto |
| Estado del Sprint | `Planificado`, `Activo` o `Cerrado` | `200` en todos los casos |
| Orden | `orden ASC, id ASC` | Lista estable ante órdenes repetidos |

### Request
`GET /api/proyectos/{id}/sprints/{sprintId}/historias` sin cuerpo.

## Salidas esperadas

### Endpoint
| Método | Path | Descripción | Éxito | Errores |
| :--- | :--- | :--- | :---: | :---: |
| `GET` | `/api/proyectos/{id}/sprints/{sprintId}/historias` | Consulta el Sprint Backlog | `200 OK` | `400`, `401`, `404` |

### Respuesta exitosa
```json
{
  "sprint": {
    "id": 1,
    "identificador": "Sprint 1",
    "sprint_goal": "Entregar el MVP",
    "estado": "Activo",
    "fecha_inicio": "2026-10-06",
    "fecha_fin": "2026-10-20",
    "creado_en": "2026-10-01T09:15:00Z",
    "actualizado_en": "2026-10-06T09:00:00Z"
  },
  "total": 2,
  "items": [
    {
      "id": 10,
      "orden": 1,
      "fecha_asignacion": "2026-10-06T10:00:00Z",
      "historia": {
        "id": 100,
        "codigo": "US-01",
        "titulo": "Crear proyecto",
        "descripcion": "Como usuario...",
        "prioridad": "Alta",
        "estado": "En Progreso",
        "story_points": 3,
        "esfuerzo_real": null,
        "creado_en": "2026-10-01T09:00:00Z",
        "actualizado_en": "2026-10-06T11:00:00Z"
      }
    }
  ]
}
```

### Cuerpo estándar de error
```json
{
  "error": "CODIGO",
  "mensaje": "Descripción en español",
  "campo": "campo opcional"
}
```

## Modelo de datos

### Read model de dominio
```go
type SprintBacklogDetalle struct {
    Sprint *Sprint
    Items  []SprintBacklogItemDetalle
}

type SprintBacklogItemDetalle struct {
    Item     SprintBacklogItem
    Historia Historia
}
```

No se modifica el modelo `SprintBacklogItem`: la historia anidada se resuelve
con una consulta conjunta por lote (`HistoriaRepository.ObtenerPorIDs`) para
evitar el patrón N+1.

### Errores centinela reutilizados
```go
ErrProyectoNoEncontrado = errors.New("el proyecto no existe")
ErrSprintNoEncontrado   = errors.New("el sprint no existe")
```

## Reglas de negocio

| ID | Regla |
| :--- | :--- |
| RB-01 | El proyecto debe existir y el usuario debe ser miembro para consultar |
| RB-02 | El Sprint debe existir y pertenecer al proyecto indicado |
| RB-03 | Se pueden consultar Sprints `Planificado`, `Activo` y `Cerrado` |
| RB-04 | Los items se ordenan por `orden ASC` y desempatan por `id ASC` |
| RB-05 | Cada item incluye `id`, `orden`, `fecha_asignacion` e `historia` |
| RB-06 | Un Sprint sin historias responde `total: 0` e `items: []`, nunca `null` |
| RB-07 | La consulta no modifica el estado de las historias ni del Sprint |
| RB-08 | No hay paginación |

## Casos límite
| ID | Caso | Resultado |
| :--- | :--- | :--- |
| CL-01 | Sprint sin historias | `200` con `total: 0` e `items: []` |
| CL-02 | Dos items con el mismo `orden` | Se ordenan por `id ASC` |
| CL-03 | Sprint `Planificado` | `200` con su Sprint Backlog |
| CL-04 | Sprint `Activo` | `200` con su Sprint Backlog |
| CL-05 | Sprint `Cerrado` | `200` con su Sprint Backlog sin alterar asociaciones |
| CL-06 | Sprint de otro proyecto o inexistente | `404 SPRINT_NO_ENCONTRADA` |
| CL-07 | Proyecto inexistente o usuario no miembro | `404 PROYECTO_NO_ENCONTRADO` |
| CL-08 | `{id}` o `{sprintId}` inválido | `400 PARAMETRO_INVALIDO` |

## Criterios de aceptación

- **CA-01** Given un miembro del proyecto, When pide `GET
  /api/proyectos/{id}/sprints/{sprintId}/historias`, Then recibe `200` con
  `sprint`, `total` e `items`.
- **CA-02** Given un sprint sin historias, When se consulta, Then `total` es
  `0` e `items` es `[]`, no `null`.
- **CA-03** Given un sprint con historias, When se consulta, Then los items
  respetan `orden ASC, id ASC` e incluyen `orden`, `fecha_asignacion` e
  `historia`.
- **CA-04** Given un sprint `Cerrado`, When se consulta, Then se puede leer y
  las historias mantienen su estado y asociación.
- **CA-05** Given un sprint de otro proyecto o inexistente, When se consulta,
  Then `404 SPRINT_NO_ENCONTRADA`.
- **CA-06** Given un proyecto inexistente o un no miembro, When se consulta,
  Then `404 PROYECTO_NO_ENCONTRADO`.
- **CA-07** Given un `{id}` o `{sprintId}` inválido, When se consulta, Then
  `400 PARAMETRO_INVALIDO`.
- **CA-08** Given dos items con el mismo `orden`, When se consulta, Then se
  desempata por `id ASC`.
- **CA-09** Given US-04, When se asignan o desasignan historias, Then el
  listado refleja el cambio sin alterar las reglas existentes.

## Mapeo de errores HTTP
| Error | Código HTTP | Código API |
| :--- | :---: | :--- |
| `ErrSprintNoEncontrado` | 404 | `SPRINT_NO_ENCONTRADA` |
| `ErrProyectoNoEncontrado` | 404 | `PROYECTO_NO_ENCONTRADO` |
| Parámetro de ruta inválido | 400 | `PARAMETRO_INVALIDO` |
| Sin token | 401 | `TOKEN_REQUERIDO` / `TOKEN_INVALIDO` |

## Arquitectura

Respeta Clean Architecture: `domain` (read model/puertos) → `repository`
(GORM/Postgres) → `service` (reglas de lectura) → `handler` (HTTP/Gin). El
dominio no depende de infraestructura. No se modifican las reglas de US-04 ni
el cierre de Sprint.
