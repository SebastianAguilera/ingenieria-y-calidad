# US-18 — Defectos por Sprint y Edición de Defectos

## Objetivo

Permitir consultar los defectos asociados a un Sprint, filtrados por Sprint de detección o de resolución, y corregir los datos `descripcion` y `severidad` de un defecto existente, sin crear un endpoint de edición nuevo.

**Trazabilidad:** Issue #92 → Esta especificación → Escenarios BDD (`docs/features/US-18-defectos-por-sprint.feature`) → Tests Go (TDD) → Código.

**Extiende:** US-09 (registro y gestión de defectos). **Bloquea:** US-19, US-20, US-F20.

## Alcance

**Incluye**
- Listado de defectos de un Sprint con filtro `tipo=deteccion|resolucion`.
- Extensión del `PATCH /api/proyectos/{id}/defectos/{defectoId}` de US-09 para aceptar `descripcion` y `severidad`.
- Guard de membresía (US-F10). Cualquier miembro puede consultar y editar (US-11).

**No incluye**
- Pantallas (US-F20), métricas (US-19), reporte de sprint (US-20).
- Eliminación de defectos.
- Historial de cambios.
- Cambios en la máquina de estados de US-09.
- Cambio del Sprint de detección.

## Entradas

### Endpoint 1: listado por Sprint

`GET /api/proyectos/{id}/sprints/{sprintId}/defectos?tipo=deteccion|resolucion`

| Parámetro | Ubicación | Obligatorio | Regla |
| :--- | :--- | :---: | :--- |
| `id` | path | Sí | Proyecto del usuario autenticado |
| `sprintId` | path | Sí | Debe existir en el proyecto |
| `tipo` | query | Sí | `deteccion` o `resolucion` |

- `deteccion`: defectos con `sprint_deteccion_id = sprintId`.
- `resolucion`: defectos con `sprint_resolucion_id = sprintId`. Usa el índice `idx_defectos_sprint_resolucion`.

### Endpoint 2: edición extendida

`PATCH /api/proyectos/{id}/defectos/{defectoId}`

```json
{
  "descripcion": "Error en validación de fechas al cerrar sprint",
  "severidad": "critical"
}
```

| Campo | Tipo | Obligatorio | Regla |
| :--- | :--- | :---: | :--- |
| `descripcion` | string | No | Si se envía: no vacía tras `TrimSpace`, máximo 1000 caracteres (runas) |
| `severidad` | string | No | `minor`, `major`, `critical` o `blocker` |
| `estado` | string | No | Comportamiento de US-09, sin cambios |
| `sprint_resolucion_id` | entero | No | Comportamiento de US-09, sin cambios |

Debe enviarse al menos un campo del PATCH (US-09 para `estado`/`sprint_resolucion_id`, esta historia para `descripcion`/`severidad`).

## Salidas esperadas

| Caso | Estado HTTP | Cuerpo |
| :--- | :---: | :--- |
| Listado válido | 200 | `{"total": n, "defectos": [...]}` |
| Listado sin resultados | 200 | `{"total": 0, "defectos": []}` |
| `tipo` ausente o inválido | 422 | `{"error":"VALIDACION","campo":"tipo"}` |
| Sprint inexistente o de otro proyecto | 404 | `{"error":"SPRINT_NO_ENCONTRADA"}` |
| Proyecto inexistente o usuario no miembro | 404 | `{"error":"PROYECTO_NO_ENCONTRADO"}` |
| Edición válida | 200 | Defecto actualizado |
| `descripcion` vacía o > 1000 | 422 | `{"error":"VALIDACION","campo":"descripcion"}` |
| `severidad` fuera de la lista | 422 | `{"error":"VALIDACION","campo":"severidad"}` |
| `sprint_deteccion_id` presente en el body | 422 | `{"error":"VALIDACION","mensaje":"sprint_deteccion_id no es modificable","campo":"sprint_deteccion_id"}` |
| Defecto inexistente o de otro proyecto | 404 | `{"error":"DEFECTO_NO_ENCONTRADO"}` |

## Modelo de datos

Sin cambios de esquema. `defectos.descripcion` es `text NOT NULL` y `defectos.severidad` es `varchar(20)`; el límite de 1000 caracteres se valida en la aplicación. Los índices `idx_defectos_sprint_deteccion` e `idx_defectos_sprint_resolucion` ya existen (creados por `AutoMigrate`).

## Reglas de negocio

| ID | Regla |
| :--- | :--- |
| RN-01 | `tipo` es obligatorio y solo admite `deteccion` o `resolucion`. |
| RN-02 | El listado por `deteccion` filtra por `sprint_deteccion_id`. |
| RN-03 | El listado por `resolucion` filtra por `sprint_resolucion_id`. |
| RN-04 | El Sprint consultado debe existir dentro del proyecto. |
| RN-05 | El listado devuelve `[]`, nunca `null`, cuando no hay coincidencias. |
| RN-06 | `descripcion` se normaliza con `TrimSpace`; no puede quedar vacía. |
| RN-07 | `descripcion` admite como máximo 1000 caracteres (runas Unicode, no bytes). |
| RN-08 | `severidad` debe pertenecer al conjunto `minor`, `major`, `critical`, `blocker` (sensible a mayúsculas). |
| RN-09 | El sprint de detección no es editable. Si el body del PATCH incluye la clave `sprint_deteccion_id`, se responde `422 VALIDACION` con campo `sprint_deteccion_id` y mensaje "sprint_deteccion_id no es modificable". Se rechaza siempre que la clave esté presente, aunque su valor sea igual al actual o `null`. Ninguna edición lo altera. |
| RN-10 | Editar `descripcion` o `severidad` no modifica `estado`, `sprint_resolucion_id` ni `resuelto_en`. |
| RN-11 | Un defecto de otro proyecto se comporta como inexistente. |
| RN-12 | Cualquier miembro puede consultar y editar. |

## Casos límite

| ID | Caso | Resultado |
| :--- | :--- | :--- |
| CL-01 | `tipo` ausente | 422 `VALIDACION` campo `tipo` |
| CL-02 | `tipo=Deteccion` (mayúscula) | 422 `VALIDACION` campo `tipo` |
| CL-03 | Sprint sin defectos para el tipo | 200 `[]` |
| CL-04 | `descripcion` de exactamente 1000 caracteres | 200, se guarda |
| CL-05 | `descripcion` de 1001 caracteres | 422 campo `descripcion` |
| CL-06 | `descripcion` con 1000 caracteres multibyte (ej. "ñ") | 200 (se cuentan runas, no bytes) |
| CL-07 | `descripcion` solo espacios | 422 campo `descripcion` |
| CL-08 | `severidad` vacía o `Major` | 422 campo `severidad` |
| CL-09 | PATCH con solo `severidad` | 200, no toca estado ni Sprint de resolución |
| CL-10 | PATCH con `sprint_deteccion_id` (cualquier valor, incluido el actual) | `422 VALIDACION` campo `sprint_deteccion_id`; no se modifica el defecto |

## Criterios de aceptación

- **CA-1:** `GET .../defectos?tipo=deteccion` devuelve solo los defectos detectados en el Sprint.
- **CA-2:** `GET .../defectos?tipo=resolucion` devuelve solo los defectos resueltos en el Sprint.
- **CA-3:** Un Sprint sin coincidencias responde `200` con `[]`.
- **CA-4:** `tipo` ausente o inválido responde `422` con campo `tipo`.
- **CA-5:** Sprint inexistente o de otro proyecto responde `404 SPRINT_NO_ENCONTRADA`.
- **CA-6:** Un no miembro recibe `404 PROYECTO_NO_ENCONTRADO` en listado y edición.
- **CA-7:** Se edita `descripcion` válida (hasta 1000 caracteres) y la respuesta la devuelve.
- **CA-8:** `descripcion` vacía o de más de 1000 caracteres responde `422` con campo `descripcion`.
- **CA-9:** Se edita `severidad` con un valor de la lista.
- **CA-10:** `severidad` fuera de la lista responde `422` con campo `severidad`.
- **CA-11:** Un PATCH con `sprint_deteccion_id` responde `422` con campo `sprint_deteccion_id`, aunque el valor sea el actual, y el defecto no cambia.
- **CA-12:** Defecto inexistente o de otro proyecto responde `404 DEFECTO_NO_ENCONTRADO`.

## Mapeo de errores HTTP

| Error de dominio | Estado | Código | Campo |
| :--- | :---: | :--- | :--- |
| `ErrTipoListadoDefectoInvalido` | 422 | `VALIDACION` | `tipo` |
| `ErrSprintNoEncontrado` | 404 | `SPRINT_NO_ENCONTRADA` | — |
| `ErrDescripcionDefectoObligatoria` | 422 | `VALIDACION` | `descripcion` |
| `ErrDescripcionDefectoDemasiadoLarga` | 422 | `VALIDACION` | `descripcion` |
| `ErrSeveridadDefectoInvalida` | 422 | `VALIDACION` | `severidad` |
| `ErrDefectoNoEncontrado` | 404 | `DEFECTO_NO_ENCONTRADO` | — |
| `ErrProyectoNoEncontrado` | 404 | `PROYECTO_NO_ENCONTRADO` | — |
| `ErrSprintDeteccionNoModificable` | 422 | `VALIDACION` | `sprint_deteccion_id` |

## Arquitectura

- **domain:** nuevo tipo `TipoListadoDefecto` (`deteccion`/`resolucion`) con `Valido()`; nuevo error `ErrDescripcionDefectoDemasiadoLarga`; `ActualizacionDefecto` gana `Descripcion` y `Severidad` opcionales; puerto `DefectoRepository` gana `ListarPorSprint(ctx, proyectoID, sprintID, tipo)`.
- **repository:** implementación de `ListarPorSprint` sobre GORM; `Actualizar` persiste `descripcion` y `severidad`.
- **service:** `DefectoService.ListarPorSprint` valida tipo y existencia del Sprint; `Actualizar` valida y aplica `descripcion`/`severidad`.
- **handler:** nueva ruta `GET /api/proyectos/:id/sprints/:sprintId/defectos` en `defecto_handler.go`; `dto_defecto.go` acepta los campos nuevos del PATCH.

## Supuestos a confirmar

- Decisión confirmada: `sprint_deteccion_id` en el body del PATCH se rechaza con `422 VALIDACION` (campo `sprint_deteccion_id`). Es una excepción a la regla general de "campos desconocidos se ignoran" de US-01/02/03, porque el campo es conocido y no modificable.
- El listado por Sprint no incluye paginación (la issue no la pide).

## Escenarios BDD

Archivo: `docs/features/US-18-defectos-por-sprint.feature`. Total: **22 escenarios (E-01 a E-22)**. Cada ID se reutiliza como nombre de subtest en Go.

| Grupo | IDs | Criterios |
| :--- | :--- | :--- |
| Listado por Sprint | E-01 a E-08 | CA-1 a CA-6 |
| Edición de descripción y severidad | E-09 a E-16, E-18, E-19 | CA-7 a CA-10, CA-12 |
| Rechazo de `sprint_deteccion_id` | E-17, E-21, E-22 | CA-11 |
| No miembro en edición | E-20 | CA-6 |

Guardas de regresión (pasan desde el primer RED porque ya estaba implementado en US-09 / US-F10 o solo traducen un error del service): **E-15** (handler), **E-20** (`cmd/api`), y en el handler **E-11, E-12, E-14, E-16, E-18, E-19** (`[GUARDA DE REGRESION]`).
