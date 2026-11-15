# US-09 — Registro y Gestión de Defectos (Bugs)

## Objetivo

Permitir al equipo registrar defectos asociados a una historia de usuario y al Sprint donde fueron detectados, y actualizar su estado indicando el Sprint de resolución cuando corresponde. Esta información alimenta el control de calidad y las métricas de defectos del proyecto.

**Trazabilidad:** Issue #45 → Esta especificación → Criterios de aceptación → Escenarios BDD (`docs/features/US-09-registro-y-gestion-de-defectos.feature`) → Tests Go (TDD) → Código.

## Alcance

**Incluye**
- Registrar un defecto con descripción, severidad, historia de usuario y Sprint de detección.
- Crear todo defecto en estado `open`.
- Consultar los defectos de un proyecto.
- Actualizar el estado del defecto.
- Indicar el Sprint de resolución cuando el defecto pasa a `resolved` o `closed`.
- Validar severidad y estado contra los valores permitidos.

**No incluye**
- Asignación de responsables de resolución.
- Comentarios, adjuntos o historial de cambios.
- Cálculo de métricas de defectos (se consume en una historia posterior).
- Borrado físico o lógico de defectos.

## Entradas

### Endpoints

| Método | Path | Descripción |
| :--- | :--- | :--- |
| `POST` | `/api/proyectos/{id}/defectos` | Registra un defecto |
| `GET` | `/api/proyectos/{id}/defectos` | Lista los defectos del proyecto |
| `PATCH` | `/api/proyectos/{id}/defectos/{defectoId}` | Actualiza estado y Sprint de resolución |

### Alta de defecto

```json
{
  "descripcion": "Error en validación de fechas",
  "severidad": "major",
  "historia_id": 1,
  "sprint_deteccion_id": 2,
  "reportado_por": 3
}
```

| Campo | Tipo | Obligatorio | Regla |
| :--- | :--- | :---: | :--- |
| `descripcion` | string | Sí | No vacía; se recorta antes de persistir |
| `severidad` | string | Sí | `minor`, `major`, `critical` o `blocker` |
| `historia_id` | entero | Sí | Debe existir en el proyecto |
| `sprint_deteccion_id` | entero | Sí | Debe existir en el proyecto |
| `reportado_por` | entero | No | Si se informa, debe existir |

El cliente no puede informar `estado`: todo defecto nace en `open`.

### Actualización de defecto

```json
{
  "estado": "resolved",
  "sprint_resolucion_id": 3
}
```

| Campo | Tipo | Obligatorio | Regla |
| :--- | :--- | :---: | :--- |
| `estado` | string | No | `open`, `in_progress`, `resolved`, `closed` o `rejected` |
| `sprint_resolucion_id` | entero | No | Obligatorio si el estado resultante es `resolved` o `closed` |

Debe informarse al menos uno de los dos campos. El Sprint de resolución solo se admite cuando el estado resultante es `resolved` o `closed`.

## Salidas esperadas

| Caso | Estado HTTP | Cuerpo |
| :--- | :---: | :--- |
| Alta válida | 201 | Defecto creado con `estado: "open"` |
| Listado válido | 200 | `{"total": n, "defectos": [...]}` |
| Actualización válida | 200 | Defecto actualizado |
| Severidad inválida | 422 | `{"error":"VALIDACION","campo":"severidad"}` |
| Descripción vacía | 422 | `{"error":"VALIDACION","campo":"descripcion"}` |
| Estado inválido | 400 | `{"error":"ESTADO_INVALIDO","campo":"estado"}` |
| Sprint de resolución faltante o no permitido | 422 | `{"error":"VALIDACION","campo":"sprint_resolucion_id"}` |
| Historia, Sprint, usuario o defecto inexistente | 404 | Código correspondiente |

## Modelo de datos

### Dominio

```go
type Defecto struct {
    ID                 uint
    ProyectoID         uint
    HistoriaID         uint
    SprintDeteccionID  uint
    SprintResolucionID *uint
    ReportadoPor       *uint
    Descripcion        string
    Severidad          SeveridadDefecto
    Estado             EstadoDefecto
    CreadoEn           time.Time
    ResueltoEn         *time.Time
}
```

### Enumeraciones

- **Severidad:** `minor`, `major`, `critical`, `blocker`.
- **Estado:** `open`, `in_progress`, `resolved`, `closed`, `rejected`.

### Persistencia

Tabla `defectos`:

| Columna | Tipo | Restricción |
| :--- | :--- | :--- |
| `id` | serial | PK |
| `proyecto_id` | entero | FK → `proyectos`, NOT NULL |
| `historia_id` | entero | FK → `historias`, NOT NULL |
| `sprint_deteccion_id` | entero | FK → `sprints`, NOT NULL |
| `sprint_resolucion_id` | entero | FK → `sprints`, nullable |
| `reportado_por` | entero | FK → `usuarios`, nullable |
| `descripcion` | text | NOT NULL |
| `severidad` | varchar(20) | NOT NULL, default `minor` |
| `estado` | varchar(20) | NOT NULL, default `open` |
| `creado_en` | timestamptz | NOT NULL, default `now()` |
| `resuelto_en` | timestamptz | nullable |

Índices: `idx_defectos_proyecto`, `idx_defectos_historia`, `idx_defectos_sprint_deteccion`, `idx_defectos_sprint_resolucion`.

## Reglas de negocio

| ID | Regla |
| :--- | :--- |
| RN-01 | La descripción es obligatoria y se normaliza con `TrimSpace`. |
| RN-02 | La severidad debe ser `minor`, `major`, `critical` o `blocker`. |
| RN-03 | Todo defecto se crea con estado `open`; el estado no se acepta en el alta. |
| RN-04 | La historia y el Sprint de detección deben existir dentro del proyecto. |
| RN-05 | `reportado_por` es opcional; si se informa, el usuario debe existir. |
| RN-06 | La actualización exige al menos `estado` o `sprint_resolucion_id`. |
| RN-07 | El estado informado debe pertenecer al conjunto permitido. |
| RN-08 | Si el estado resultante es `resolved` o `closed`, el Sprint de resolución es obligatorio. |
| RN-09 | El Sprint de resolución debe existir en el proyecto y solo se admite para estados de resolución. |
| RN-10 | Al resolver o cerrar se registra `resuelto_en`; al reabrir se limpia junto con el Sprint de resolución. |
| RN-11 | Un defecto de otro proyecto se comporta como inexistente. |
| RN-12 | El listado devuelve un arreglo vacío, nunca `null`, cuando no hay defectos. |

## Casos límite

| ID | Caso | Resultado |
| :--- | :--- | :--- |
| CL-01 | Descripción vacía o solo espacios | 422 `VALIDACION` campo `descripcion` |
| CL-02 | Severidad vacía o con mayúsculas (`Major`) | 422 `VALIDACION` campo `severidad` |
| CL-03 | Historia inexistente | 404 `HISTORIA_NO_ENCONTRADA` |
| CL-04 | Sprint de detección inexistente | 404 `SPRINT_NO_ENCONTRADA` |
| CL-05 | `reportado_por` inexistente | 404 `USUARIO_NO_ENCONTRADO` |
| CL-06 | Actualización sin campos | 422 `VALIDACION` |
| CL-07 | Resolver sin Sprint de resolución | 422 `VALIDACION` campo `sprint_resolucion_id` |
| CL-08 | Informar Sprint de resolución con estado `open` | 422 `VALIDACION` campo `sprint_resolucion_id` |
| CL-09 | Sprint de resolución inexistente | 404 `SPRINT_NO_ENCONTRADA` |
| CL-10 | Reabrir un defecto resuelto | Limpia `sprint_resolucion_id` y `resuelto_en` |
| CL-11 | Listado de proyecto sin defectos | `{"total":0,"defectos":[]}` |
| CL-12 | Defecto de otro proyecto | 404 `DEFECTO_NO_ENCONTRADO` |

## Criterios de aceptación

- **CA-F1:** Se registra un defecto válido y queda con estado `open`.
- **CA-F2:** El defecto queda asociado a la historia y al Sprint de detección.
- **CA-F3:** Se actualiza el estado del defecto.
- **CA-F4:** Al resolver o cerrar se exige y persiste el Sprint de resolución.
- **CA-F5:** Se listan los defectos del proyecto.
- **CA-R1:** Severidad y estado se validan contra sus conjuntos permitidos.
- **CA-R2:** Las referencias (proyecto, historia, Sprint, usuario) deben existir.
- **CA-R3:** El Sprint de resolución solo se admite en estados de resolución.

## Mapeo de errores HTTP

| Error de dominio | Estado | Código | Campo |
| :--- | :---: | :--- | :--- |
| `ErrDefectoNoEncontrado` | 404 | `DEFECTO_NO_ENCONTRADO` | — |
| `ErrDescripcionDefectoObligatoria` | 422 | `VALIDACION` | `descripcion` |
| `ErrSeveridadDefectoInvalida` | 422 | `VALIDACION` | `severidad` |
| `ErrEstadoDefectoInvalido` | 400 | `ESTADO_INVALIDO` | `estado` |
| `ErrActualizacionDefectoVacia` | 422 | `VALIDACION` | — |
| `ErrSprintResolucionRequerido` | 422 | `VALIDACION` | `sprint_resolucion_id` |
| `ErrSprintResolucionNoPermitido` | 422 | `VALIDACION` | `sprint_resolucion_id` |
| `ErrHistoriaNoEncontrada` | 404 | `HISTORIA_NO_ENCONTRADA` | — |
| `ErrSprintNoEncontrado` | 404 | `SPRINT_NO_ENCONTRADA` | — |
| `ErrUsuarioNoEncontrado` | 404 | `USUARIO_NO_ENCONTRADO` | — |
| `ErrProyectoNoEncontrado` | 404 | `PROYECTO_NO_ENCONTRADO` | — |

## Arquitectura

- **domain:** `defecto.go` con enums, entidades, comandos de alta/actualización y `defecto_repository.go` con el puerto.
- **repository:** `defecto_repository.go` y `models/defecto_model.go` sobre GORM/PostgreSQL.
- **service:** `defecto_service.go` valida referencias, severidad, estado y Sprint de resolución.
- **handler:** `defecto_handler.go`, `dto_defecto.go` y registro de rutas en `cmd/api/router.go`.
