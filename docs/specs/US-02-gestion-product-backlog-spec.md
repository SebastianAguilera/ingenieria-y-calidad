# Especificación: US-02 - Gestión básica del Product Backlog

## Objetivo

Implementar el módulo de **gestión básica del Product Backlog** del *Software Metrics & Estimation Engine*: agregar, listar y modificar **historias de usuario** dentro de un proyecto. La funcionalidad satisface el requerimiento 2 de `docs/proyecto.md` (Product Backlog compuesto por historias con identificador, título, descripción, prioridad y estado) y constituye la base sobre la que se construirán los Sprints, la estimación por Story Points, el Planning Poker, los worklogs, los defectos y el motor de métricas.

El resultado es un conjunto de tres endpoints REST servidos por Gin, con las reglas de negocio en Go, persistencia en PostgreSQL mediante GORM y esquema generado por `AutoMigrate`. Las historias cuelgan de la entidad raíz definida en US-01 (`proyectos`): cada historia pertenece exactamente a un proyecto y su código visible `US-NN` es un consecutivo por proyecto.

**Alcance.** La historia cubre tres operaciones: agregar (alta), listar y modificar. No se especifica la eliminación de historias ni la reordenación manual del backlog: el orden del listado es el orden de alta.

**Trazabilidad:** la funcionalidad se desarrolla en la rama `feature/US-02-gestion-product-backlog` bajo el ciclo TDD (RED → GREEN → REFACTOR). Los escenarios BDD de `docs/features/US-02-gestion-product-backlog.feature` (identificadores `E-01`..`E-34`) son el origen de los casos de prueba de cada capa y cada criterio de aceptación es verificable por un test concreto.

## Entradas

### Precondiciones

- La aplicación está configurada mediante `.env` (tomando `.env.example` como referencia) e `internal/repository.ConnectDB()` establece la conexión con PostgreSQL.
- El esquema de US-01 existe y el proyecto referenciado por la ruta fue creado previamente: las historias son un recurso anidado y no pueden existir de forma huérfana.
- El actor está autenticado (US-10) y **es miembro del proyecto de la ruta** (US-F10): quien no es miembro recibe `404 PROYECTO_NO_ENCONTRADO` en el guard, antes de llegar a este servicio.

### Convenciones de formato de entrada

| Convención | Regla | Consecuencia si se viola |
| :--- | :--- | :--- |
| Identificadores de recurso | Entero positivo en el path (`/api/proyectos/{id}/historias/{historiaId}`), leído con `strconv.ParseUint` | `400 PARAMETRO_INVALIDO` si no es numérico o es `0` |
| Timestamps de auditoría | RFC 3339 UTC (`2026-10-01T09:15:00Z`) para `creado_en` y `actualizado_en` | — |
| JSON | Un único objeto JSON por request; el handler usa `ShouldBindJSON` | `400 JSON_INVALIDO` si el cuerpo está vacío, malformado o con tipos incorrectos |
| Campos desconocidos | Se ignoran silenciosamente en el body | — |
| Campos calculados | `id`, `codigo`, `proyecto_id`, `creado_en` y `actualizado_en` son de solo lectura: si el cliente los envía, se ignoran | — |

### Cuerpos de request por operación

#### `POST /api/proyectos/{id}/historias` — agregar historia

| Campo | Tipo JSON | Requerido | Formato / rango | Descripción |
| :--- | :--- | :---: | :--- | :--- |
| `titulo` | string | Sí | 1 a 200 caracteres después de `TrimSpace` | Título de la historia |
| `descripcion` | string | No | Texto libre (por defecto `""`) | Descripción de la historia |
| `prioridad` | string | No | `"Alta"`, `"Media"` o `"Baja"`; por defecto `"Media"` | Prioridad |
| `estado` | string | No | `"Backlog"`, `"En Progreso"` o `"Hecho"`; por defecto `"Backlog"` | Estado |

```json
{
  "titulo": "Como Product Builder quiero gestionar el backlog",
  "descripcion": "Cubrir el alcance del proyecto",
  "prioridad": "Alta",
  "estado": "Backlog"
}
```

El body mínimo válido es `{"titulo": "..."}`: prioridad, estado y descripción son opcionales.

#### `GET /api/proyectos/{id}/historias` — listar el backlog

Sin cuerpo de request. El listado es siempre por proyecto y no admite paginación en US-02.

#### `PATCH /api/proyectos/{id}/historias/{historiaId}` — modificar historia

| Campo | Tipo JSON | Requerido | Formato / rango | Descripción |
| :--- | :--- | :---: | :--- | :--- |
| `titulo` | string | No | 1 a 200 caracteres después de `TrimSpace` | Nuevo título |
| `descripcion` | string | No | Texto libre | Nueva descripción |
| `prioridad` | string | No | `"Alta"`, `"Media"` o `"Baja"` | Nueva prioridad |
| `estado` | string | No | `"Backlog"`, `"En Progreso"` o `"Hecho"` | Nuevo estado |

```json
{ "prioridad": "Alta" }
```

Es una **actualización parcial** (decisión D-01): solo los campos informados se modifican; los ausentes conservan su valor. Es obligatorio informar al menos un campo, de lo contrario se responde `422 VALIDACION`.

## Salidas esperadas

### 1. Diseño de la API REST

Las rutas se registran bajo el grupo `/api` desde `cmd/api/router.go`, sobre el mismo `*gin.Engine` que ya sirve US-01 y `/health`.

| # | Método | Path | Descripción | Éxito | Errores |
| :---: | :--- | :--- | :--- | :---: | :--- |
| 1 | `POST` | `/api/proyectos/{id}/historias` | Agrega una historia al backlog del proyecto | `201 Created` | `400`, `404`, `409`, `422`, `500` |
| 2 | `GET` | `/api/proyectos/{id}/historias` | Lista el backlog del proyecto en orden de alta | `200 OK` | `400`, `404`, `500` |
| 3 | `PATCH` | `/api/proyectos/{id}/historias/{historiaId}` | Modifica parcialmente una historia existente | `200 OK` | `400`, `404`, `409`, `422`, `500` |

No se especifica ningún endpoint adicional en US-02. En particular no existe un `DELETE /historias/{historiaId}` ni un endpoint para reordenar el backlog.

#### Representación de una historia

```json
{
  "id": 10,
  "codigo": "US-01",
  "titulo": "Como Product Builder quiero gestionar el backlog",
  "descripcion": "Cubrir el alcance del proyecto",
  "prioridad": "Alta",
  "estado": "Backlog",
  "creado_en": "2026-10-01T09:15:00Z",
  "actualizado_en": "2026-10-01T09:15:00Z"
}
```

Por la decisión D-03 no se emite `proyecto_id`: el proyecto ya está identificado por la ruta.

#### Respuestas de ejemplo

`POST /api/proyectos/1/historias` — `201 Created`: devuelve la representación anterior con `codigo` `"US-01"` cuando es la primera historia del proyecto.

`GET /api/proyectos/1/historias` — `200 OK`:

```json
{
  "total": 2,
  "historias": [
    { "id": 10, "codigo": "US-01", "titulo": "Primera", "descripcion": "", "prioridad": "Media", "estado": "Backlog", "creado_en": "2026-10-01T09:15:00Z", "actualizado_en": "2026-10-01T09:15:00Z" },
    { "id": 11, "codigo": "US-02", "titulo": "Segunda", "descripcion": "", "prioridad": "Alta", "estado": "En Progreso", "creado_en": "2026-10-01T09:20:00Z", "actualizado_en": "2026-10-01T09:20:00Z" }
  ]
}
```

Con cero historias la respuesta es `{"total": 0, "historias": []}`: el array vacío se serializa explícitamente para no producir `null`.

`PATCH /api/proyectos/1/historias/10` — `200 OK`: devuelve la representación de la historia ya modificada, con `actualizado_en` refrescado.

#### Cuerpo estándar de error

```json
{
  "error": "VALIDACION",
  "mensaje": "el título de la historia es obligatorio",
  "campo": "titulo"
}
```

El cuerpo usa el formato estándar de US-01: `error` (código estable), `mensaje` y `campo` opcional cuando el error se refiere a un campo del body.

### 2. Modelo de datos

| Tipo | Campos / operaciones | Reglas |
| :--- | :--- | :--- |
| `Historia` | `ID`, `ProyectoID`, `Numero`, `Codigo`, `Titulo`, `Descripcion`, `Prioridad`, `Estado`, `CreadoEn`, `ActualizadoEn` | `Codigo` se deriva de `Numero` (`US-%02d`) y no es editable. |
| `NuevaHistoria` | `Titulo`, `Descripcion`, `Prioridad`, `Estado` | Prioridad y estado opcionales, con defaults `Media`/`Backlog`. |
| `ActualizacionHistoria` | `Titulo`, `Descripcion`, `Prioridad`, `Estado` (punteros) | Distingue "no informado" (`nil`) de "informado con valor". |
| `PrioridadHistoria` / `EstadoHistoria` | `Valido()` | Conjuntos cerrados `{Alta, Media, Baja}` y `{Backlog, En Progreso, Hecho}`. |
| `HistoriaRepository` | `Crear`, `SiguienteNumero`, `ObtenerPorID`, `ListarPorProyecto`, `Actualizar` | `ObtenerPorID` filtra por proyecto; traduce `gorm.ErrRecordNotFound` a errores de dominio. |

Las firmas completas, los tags y los errores centinela están en `internal/domain/historia.go`, `historia_repository.go` y `errores.go`.

#### Decisiones de diseño

| ID | Decisión |
| :--- | :--- |
| D-01 | La modificación es parcial: solo cambian los campos informados; si no se informa ninguno, se responde `422`. |
| D-02 | El esquema se crea con `AutoMigrate` al arrancar; no hay scripts SQL versionados. |
| D-03 | La representación no emite `proyecto_id`; el proyecto ya está identificado por la ruta. |
| D-04 | El `codigo` (`US-NN`) lo deriva el service del `numero`; el cliente no lo envía. |
| D-05 | El índice único `(proyecto_id, codigo)` garantiza la unicidad; la violación se traduce a `409 HISTORIA_DUPLICADA`. |
| D-06 | Los errores de validación responden `422` con `campo` para ubicar el dato inválido. |

### 3. Plan de pruebas

| Capa | Archivo | Qué verifica |
| :--- | :--- | :--- |
| Dominio | `internal/domain/historia_test.go` | `Valido()` de prioridad y estado; `FormatearCodigoHistoria`. |
| Servicio | `internal/service/historia_service_test.go` | Validaciones, valores por defecto, consecutivo, existencia del proyecto, actualización parcial, propagación de errores. |
| Handler | `internal/handler/historia_handler_test.go` | Códigos HTTP, cuerpos de respuesta, mapeo de errores, validación de path y JSON. |
| Repositorio | `internal/repository/historia_repository_test.go` | Integración real: consecutivo, índice único, historia de otro proyecto, listado vacío, auditoría. |
| Router | `cmd/api/router_test.go` | Las tres rutas están registradas y alcanzan el servicio inyectado. |

## Reglas de negocio

| ID | Regla |
| :--- | :--- |
| RB-01 | El título es obligatorio: después de `TrimSpace` debe tener entre 1 y 200 caracteres. |
| RB-02 | La descripción es opcional y puede ser vacía; no tiene límite máximo en US-02. |
| RB-03 | La prioridad debe pertenecer al conjunto cerrado `{Alta, Media, Baja}`. |
| RB-04 | El estado debe pertenecer al conjunto cerrado `{Backlog, En Progreso, Hecho}`. |
| RB-05 | Si no se informa prioridad toma `Media`; si no se informa estado toma `Backlog`. |
| RB-06 | El código `US-NN` es un consecutivo por proyecto, autogenerado y no editable; `NN` se formatea con al menos dos dígitos. |
| RB-07 | El listado de historias es por proyecto y respeta el orden de alta (`numero ASC`). |
| RB-08 | La modificación es parcial: solo cambia los campos informados y exige al menos uno. |
| RB-09 | Una operación sobre un proyecto inexistente se rechaza con `PROYECTO_NO_ENCONTRADO`. |
| RB-10 | Una operación sobre una historia inexistente (o perteneciente a otro proyecto) se rechaza con `HISTORIA_NO_ENCONTRADA`. |

## Restricciones

1. Arquitectura por capas: `domain → repository → service → handler`; las dependencias apuntan hacia el dominio.
2. El `domain` no importa Gin, GORM ni ningún paquete de infraestructura.
3. Los mensajes de error replican literalmente los errores centinela definidos en `internal/domain/errores.go`.
4. El handler no contiene reglas de negocio: solo traduce HTTP ↔ dominio.
5. El service no conoce `*gorm.DB` ni `*gin.Context`; recibe `context.Context` y las interfaces de `domain`.
6. Los escenarios BDD usan el dialecto `# language: es` y el identificador `E-nn` como nombre de subtest.
7. Cada caso de prueba se anota con el escenario BDD que lo origina.

## Casos límite

| ID | Caso | Resultado |
| :--- | :--- | :--- |
| CL-01 | Título de exactamente 200 caracteres | Válido (`201` / `200`). |
| CL-02 | Título de 201 caracteres | `422 VALIDACION` campo `titulo`. |
| CL-03 | Título vacío o solo espacios | `422 VALIDACION` campo `titulo`. |
| CL-04 | Alta sin prioridad ni estado | Defaults `Media` / `Backlog`. |
| CL-05 | Alta sin descripción | `descripcion` `""`. |
| CL-06 | Segunda historia del proyecto | `codigo` `US-02`. |
| CL-07 | Décima historia del proyecto | `codigo` `US-10`. |
| CL-08 | Historia con `numero` ya tomado (carrera) | `409 HISTORIA_DUPLICADA`. |
| CL-09 | `PATCH` sin campos | `422 VALIDACION`. |
| CL-10 | `PATCH` de una historia inexistente | `404 HISTORIA_NO_ENCONTRADA`. |
| CL-11 | `GET` de un proyecto sin historias | `200` con `{"total":0,"historias":[]}`. |
| CL-12 | Historia de otro proyecto: `ObtenerPorID`/`PATCH` | `404 HISTORIA_NO_ENCONTRADA`. |
| CL-13 | `{id}` o `{historiaId}` no numérico o `0` | `400 PARAMETRO_INVALIDO`. |
| CL-14 | `PATCH` idempotente (mismo valor ya persistido) | `200`, sin cambios efectivos. |

## Condiciones de error

Los mensajes exactos son los de los errores centinela de `internal/domain/errores.go`.

### Mapeo de error de dominio a respuesta HTTP

| Error de dominio | `error` | HTTP | `campo` |
| :--- | :--- | :---: | :--- |
| `ErrProyectoNoEncontrado` | `PROYECTO_NO_ENCONTRADO` | `404` | — |
| `ErrHistoriaNoEncontrada` | `HISTORIA_NO_ENCONTRADA` | `404` | — |
| `ErrCodigoHistoriaDuplicado` | `HISTORIA_DUPLICADA` | `409` | `codigo` |
| `ErrTituloHistoriaObligatorio` / `ErrTituloHistoriaDemasiadoLargo` | `VALIDACION` | `422` | `titulo` |
| `ErrPrioridadHistoriaInvalida` | `VALIDACION` | `422` | `prioridad` |
| `ErrEstadoHistoriaInvalido` | `VALIDACION` | `422` | `estado` |
| `ErrActualizacionHistoriaVacia` | `VALIDACION` | `422` | — |
| Path no numérico o `0` | `PARAMETRO_INVALIDO` | `400` | — |
| JSON malformado | `JSON_INVALIDO` | `400` | — |

### Ejemplo de respuesta de error

`POST /api/proyectos/1/historias` con `{"titulo": ""}`:

```json
{ "error": "VALIDACION", "mensaje": "el título de la historia es obligatorio", "campo": "titulo" }
```

## Criterios de aceptación

### Funcionales

- **CA-I1.** Cada historia expone `id`, `codigo`, `titulo`, `descripcion`, `prioridad`, `estado`, `creado_en` y `actualizado_en`.
- **CA-I2.** Se puede listar el backlog completo de un proyecto con `{"total": n, "historias": [...]}`.
- **CA-I3.** Se puede modificar una historia (título, descripción, prioridad y/o estado).
- **CA-I4.** El código `US-NN` es único y consecutivo por proyecto, y no es editable por el cliente.
- **CA-I5.** La prioridad y el estado por defecto al alta son `Media` y `Backlog`.

### Reglas de negocio

- **CA-R1.** Título obligatorio de 1 a 200 caracteres (RB-01).
- **CA-R2.** Prioridad y estado en conjuntos cerrados (RB-03, RB-04).
- **CA-R3.** Alta y modificación sobre proyecto inexistente → `404` (RB-09).
- **CA-R4.** Modificación/consulta de historia inexistente o de otro proyecto → `404` (RB-10).

### Arquitectura y calidad

- **CA-A1..A3.** Capas y dominio sin infraestructura; cobertura ≥ 80 %; `go build`, `go vet` y `go test` sin errores.

## Escenarios BDD

Los escenarios completos viven en `docs/features/US-02-gestion-product-backlog.feature` (`E-01`..`E-34`) y se usan como nombre de subtest en cada capa. Trazabilidad: **Issue #29 → esta especificación → criterios de aceptación → escenarios BDD → tests Go → código**.
