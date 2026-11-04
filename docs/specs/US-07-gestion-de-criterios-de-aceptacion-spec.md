# US-07 — Gestión de Criterios de Aceptación

## Objetivo

Permitir al Product Builder agregar, consultar, actualizar y eliminar criterios de aceptación estructurados para cada historia de usuario del Product Backlog, asegurando la trazabilidad y la definición formal de los requerimientos según el diseño de `tablas.sql` y `docs/proyecto.md`.

**Trazabilidad:** Issue #49 → Esta especificación → Criterios de aceptación → Escenarios BDD (`features/US-07-gestion-de-criterios-de-aceptacion.feature`) → Tests Go (TDD) → Código.

## Alcance

**Incluye**
- Crear un criterio de aceptación asociado a una historia de usuario (`POST /api/proyectos/{id}/historias/{historiaId}/criterios`).
- Listar los criterios de aceptación de una historia de usuario (`GET /api/proyectos/{id}/historias/{historiaId}/criterios`).
- Actualizar la descripción o marcar como cumplido (`is_met`) un criterio de aceptación (`PATCH /api/proyectos/{id}/historias/{historiaId}/criterios/{criterioId}`).
- Eliminar un criterio de aceptación (`DELETE /api/proyectos/{id}/historias/{historiaId}/criterios/{criterioId}`).
- Validar que la descripción sea obligatoria (no vacía ni compuesta solo por espacios) y que no supere los 500 caracteres.

**No incluye**
- Criterios de aceptación a nivel de proyecto o sprint.
- Reordenamiento masivo de criterios (la posición se asigna automáticamente de forma secuencial).

## Entradas

### Precondiciones
- Proyecto e historia de usuario existentes.

### Endpoints y Campos

| Endpoint | Parámetro / Campo | Tipo JSON | Obligatorio | Validación |
|---|---|---|---|---|
| `POST /api/proyectos/{id}/historias/{historiaId}/criterios` | `descripcion` | string | Sí | No vacío, máx. 500 chars |
| `PATCH /api/proyectos/{id}/historias/{historiaId}/criterios/{criterioId}` | `descripcion` | string | No | Si se informa, no vacío |
| `PATCH /api/proyectos/{id}/historias/{historiaId}/criterios/{criterioId}` | `is_met` | boolean | No | Booleano (`true`/`false`) |

### Ejemplo de Request (POST)
```json
{
  "descripcion": "El sistema valida que el formato de email sea correcto"
}
```

### Ejemplo de Request (PATCH)
```json
{
  "is_met": true
}
```

## Salidas esperadas

| Caso | Estado HTTP | Cuerpo |
|---|---|---|
| Creación exitosa de criterio | 201 | objeto del criterio creado (`id`, `descripcion`, `position`, `is_met`) |
| Listado de criterios | 200 | `{"total": N, "criterios": [...]}` |
| Modificación exitosa | 200 | objeto del criterio actualizado |
| Eliminación exitosa | 204 | sin contenido |
| Descripción vacía o inválida | 422 | `{"error":"VALIDACION","mensaje":"...","campo":"descripcion"}` |
| Historia o criterio inexistente | 404 | `{"error":"HISTORIA_NO_ENCONTRADA"}` o `{"error":"CRITERIO_NO_ENCONTRADO"}` |

## Modelo de datos

### Dominio (`internal/domain/acceptance_criterion.go`)
- Estructura `AcceptanceCriterion`:
  - `ID` uint
  - `HistoriaID` uint
  - `Descripcion` string
  - `Position` int
  - `IsMet` bool
- Errores centinela:
  - `ErrCriterioNoEncontrado`
  - `ErrDescripcionCriterioObligatoria`
  - `ErrDescripcionCriterioDemasiadoLarga`
  - `ErrActualizacionCriterioVacia`

### Persistencia (`internal/repository/models/acceptance_criterion_model.go`)
- Mapeo de la tabla `acceptance_criterion` (`tablas.sql`):
  - `id` (primaryKey, autoIncrement)
  - `backlog_item_id` (foreignKey a `historias.id`, onDelete CASCADE)
  - `description` (varchar/text, not null)
  - `position` (integer, default 1)
  - `is_met` (boolean, default false)
- Se incluye en `MigrarEsquema` mediante `AutoMigrate`.

## Reglas de negocio

- **RN-01** Toda historia de usuario puede tener cero o más criterios de aceptación.
- **RN-02** La descripción del criterio es obligatoria y debe recortarse (`strings.TrimSpace`). No puede estar vacía.
- **RN-03** La longitud máxima de la descripción es de 500 caracteres.
- **RN-04** La posición (`position`) se asigna automáticamente como el siguiente número consecutivo dentro de la historia (máx. posición actual + 1, o 1 si es el primero).
- **RN-05** El estado de cumplimiento (`is_met`) por defecto es `false` al crear el criterio.
- **RN-06** En la modificación (`PATCH`), cada campo es opcional (puntero); si no se informa, conserva su valor anterior. Al menos un campo debe ser informado.

## Casos límite

- **CL-01** Crear el primer criterio asigna `position: 1`.
- **CL-02** Crear un segundo criterio asigna `position: 2` automáticamente.
- **CL-03** Enviar una descripción compuesta solo por espacios se rechaza con error de validación.
- **CL-04** Intentar operar sobre un criterio que pertenece a otro proyecto o historia devuelve error de no encontrado (404).

## Criterios de aceptación

- **CA-F1** Se crea un criterio de aceptación válido y la API responde 201 con los datos asignados.
- **CA-F2** Se listan correctamente todos los criterios de una historia en orden de posición.
- **CA-F3** Se actualiza el estado `is_met` a `true` y la API responde 200 con el cambio reflejado.
- **CA-F4** Se elimina un criterio y deja de aparecer en el listado.
- **CA-R1** Descripciones vacías o con más de 500 caracteres devuelven error 422 con campo `descripcion`.

## Mapeo de errores HTTP

| Error de dominio | Estado | Código | Campo |
|---|---|---|---|
| `ErrCriterioNoEncontrado` | 404 | `CRITERIO_NO_ENCONTRADO` | — |
| `ErrDescripcionCriterioObligatoria` | 422 | `VALIDACION` | `descripcion` |
| `ErrDescripcionCriterioDemasiadoLarga` | 422 | `VALIDACION` | `descripcion` |
| `ErrActualizacionCriterioVacia` | 422 | `VALIDACION` | — |

## Arquitectura

- **domain:** Estructuras, validaciones de dominio y errores centinela.
- **repository:** Modelo GORM `AcceptanceCriterionModel`, métodos de persistencia (`Crear`, `ListarPorHistoria`, `ObtenerPorID`, `Actualizar`, `Eliminar`, `SiguientePosicion`).
- **service:** `AcceptanceCriterionService` con reglas de negocio y validación de existencia de proyecto/historia.
- **handler:** `AcceptanceCriterionHandler` con DTOs, parsing y traducción de errores HTTP.
- **router:** Nuevas rutas bajo `/api/proyectos/{id}/historias/{historiaId}/criterios`.
