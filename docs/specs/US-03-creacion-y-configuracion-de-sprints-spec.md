# Especificación: US-03 - Creación y configuración de Sprints

## Objetivo

Implementar el módulo de **Sprints** del *Software Metrics & Estimation Engine*: crear un Sprint dentro de un proyecto especificando su identificador, su Sprint Goal y sus fechas, listar los Sprints del proyecto y cambiar su estado a lo largo de la máquina de estados `Planificado → Activo → Cerrado`. La funcionalidad satisface el requerimiento 3 de `docs/proyecto.md` (Sprints con objetivo y fechas, estado gestionable) y constituye la base sobre la que se construirán US-04 (Sprint Backlog), US-06 (worklogs), US-07 (cierre de Sprints) y el motor de métricas.

El resultado es un conjunto de tres endpoints REST servidos por Gin, con las reglas de negocio en Go, persistencia en PostgreSQL mediante GORM y esquema generado por `AutoMigrate`. Los Sprints cuelgan de la entidad raíz definida en US-01 (`proyectos`): cada Sprint pertenece exactamente a un proyecto y su identificador visible (`Sprint 1`, `Sprint 2`, …) es único por proyecto.

**Alcance.** La historia cubre tres operaciones: crear, listar y cambiar el estado. No se especifica la eliminación de Sprints, la modificación de sus fechas o de su Sprint Goal una vez creado (la configuración se hace al momento del alta), ni la asignación de historias al Sprint Backlog (US-04), ni el cierre automático con métricas (US-07).

**Trazabilidad:** la funcionalidad se desarrolla bajo el ciclo TDD (RED → GREEN → REFACTOR). Los escenarios BDD de `features/US-03-creacion-y-configuracion-de-sprints.feature` (identificadores `E-01`..`E-41`) son el origen de los casos de prueba de cada capa y cada criterio de aceptación es verificable por un test concreto.

## Entradas

### Precondiciones

- La aplicación está configurada mediante `.env` (tomando `.env.example` como referencia) e `internal/repository.ConnectDB()` establece la conexión con PostgreSQL.
- El esquema de US-01 existe y el proyecto referenciado por la ruta fue creado previamente: los Sprints son un recurso anidado y no pueden existir de forma huérfana.
- El esquema de la tabla `sprints` se crea automáticamente en el arranque mediante `AutoMigrate` (decisión D-02), incluida la clave foránea hacia `proyectos`.
- No se implementa autenticación ni autorización en US-03: el actor es implícitamente un administrador del sistema.

### Convenciones de formato de entrada

| Convención | Regla | Consecuencia si se viola |
| :--- | :--- | :--- |
| Identificadores de recurso | Entero positivo en el path (`/api/proyectos/{id}/sprints/{sprintId}`), leído con `strconv.ParseUint` | `400 PARAMETRO_INVALIDO` si no es numérico o es `0` |
| Identificador del Sprint | Texto de 1 a 50 caracteres después de `TrimSpace`; único dentro del proyecto | `422 VALIDACION` (campo `identificador`) / `409 SPRINT_DUPLICADO` |
| Sprint Goal | Texto de 1 a 300 caracteres después de `TrimSpace` | `422 VALIDACION` (campo `sprint_goal`) |
| Fechas de Sprint | Fecha calendario en formato `YYYY-MM-DD` (RFC 3339 simplificado, mismo formato que US-01); ambas obligatorias | `400 PARAMETRO_INVALIDO` si no parsean o faltan |
| Orden de fechas | `fecha_fin >= fecha_inicio`; un Sprint de un solo día (`fecha_fin == fecha_inicio`) es válido | `422 VALIDACION` (campo `fecha_fin`) si `fecha_fin < fecha_inicio` |
| Estado | Literal `"Planificado"`, `"Activo"` o `"Cerrado"` (distingue mayúsculas); ausente o vacío en el cambio de estado no está admitido | `400 ESTADO_INVALIDO` si el literal no pertenece al conjunto |
| Timestamps de auditoría | RFC 3339 UTC (`2026-10-01T09:15:00Z`) para `creado_en` y `actualizado_en` | — |
| JSON | Un único objeto JSON por request; el handler usa `ShouldBindJSON` | `400 JSON_INVALIDO` si el cuerpo está vacío, malformado o con tipos incorrectos |
| Campos desconocidos | Se ignoran silenciosamente en el body | — |
| Campos calculados | `id`, `estado`, `creado_en` y `actualizado_en` son de solo lectura: si el cliente los envía al crear, se ignoran | — |

### Cuerpos de request por operación

#### `POST /api/proyectos/{id}/sprints` — crear Sprint

| Campo | Tipo JSON | Requerido | Formato / rango | Descripción |
| :--- | :--- | :---: | :--- | :--- |
| `identificador` | string | Sí | 1 a 50 caracteres después de `TrimSpace` | Identificador visible del Sprint (ej. `"Sprint 1"`) |
| `sprint_goal` | string | Sí | 1 a 300 caracteres después de `TrimSpace` | Objetivo del Sprint |
| `fecha_inicio` | string | Sí | `YYYY-MM-DD` | Fecha de inicio |
| `fecha_fin` | string | Sí | `YYYY-MM-DD`, `>= fecha_inicio` | Fecha de fin |

```json
{
  "identificador": "Sprint 1",
  "sprint_goal": "Entregar el MVP del motor de métricas",
  "fecha_inicio": "2026-10-06",
  "fecha_fin": "2026-10-20"
}
```

Los cuatro campos son obligatorios: no existe un body mínimo válido. El `estado` no se acepta del cliente: el alta siempre produce un Sprint en estado `Planificado` (decisión D-03).

#### `GET /api/proyectos/{id}/sprints` — listar los Sprints del proyecto

Sin cuerpo de request. El listado es siempre por proyecto, ordenado por `fecha_inicio ASC` y desempatado por `id ASC`, y no admite paginación en US-03.

#### `PATCH /api/proyectos/{id}/sprints/{sprintId}` — cambiar el estado del Sprint

La ruta es `PATCH /api/proyectos/{id}/sprints/{sprintId}/estado`.

| Campo | Tipo JSON | Requerido | Formato / rango | Descripción |
| :--- | :--- | :---: | :--- | :--- |
| `estado` | string | Sí | `"Activo"` o `"Cerrado"` (el estado actual del Sprint debe permitir la transición) | Nuevo estado |

```json
{ "estado": "Activo" }
```

Solo se admite el cambio de estado en US-03: no existe un `PATCH` de identificador, goal ni fechas.

## Salidas esperadas

### 1. Diseño de la API REST

Las rutas se registran bajo el grupo `/api` desde `cmd/api/router.go`, sobre el mismo `*gin.Engine` que ya sirve US-01, US-02 y `/health`.

| # | Método | Path | Descripción | Éxito | Errores |
| :---: | :--- | :--- | :--- | :---: | :---: |
| 1 | `POST` | `/api/proyectos/{id}/sprints` | Crea un Sprint en estado `Planificado` | `201 Created` | `400`, `404`, `409`, `422`, `500` |
| 2 | `GET` | `/api/proyectos/{id}/sprints` | Lista los Sprints del proyecto | `200 OK` | `400`, `404`, `500` |
| 3 | `PATCH` | `/api/proyectos/{id}/sprints/{sprintId}/estado` | Aplica una transición de estado válida | `200 OK` | `400`, `404`, `409`, `500` |

No se especifica ningún endpoint adicional en US-03. En particular no existe `DELETE /sprints/{sprintId}` ni `GET /sprints/{sprintId}` (el detalle se obtiene del listado).

#### Representación de un Sprint

```json
{
  "id": 1,
  "identificador": "Sprint 1",
  "sprint_goal": "Entregar el MVP del motor de métricas",
  "estado": "Planificado",
  "fecha_inicio": "2026-10-06",
  "fecha_fin": "2026-10-20",
  "creado_en": "2026-10-01T09:15:00Z",
  "actualizado_en": "2026-10-01T09:15:00Z"
}
```

Por la decisión D-04 no se emite `proyecto_id`: el proyecto ya está identificado por la ruta.

#### Respuestas de ejemplo

`POST /api/proyectos/1/sprints` — `201 Created`: devuelve la representación anterior con `estado` `"Planificado"`.

`GET /api/proyectos/1/sprints` — `200 OK`:

```json
{
  "total": 2,
  "sprints": [
    { "id": 1, "identificador": "Sprint 1", "sprint_goal": "MVP", "estado": "Cerrado", "fecha_inicio": "2026-09-08", "fecha_fin": "2026-09-22", "creado_en": "2026-09-01T10:00:00Z", "actualizado_en": "2026-09-22T18:00:00Z" },
    { "id": 2, "identificador": "Sprint 2", "sprint_goal": "Métricas", "estado": "Activo", "fecha_inicio": "2026-09-23", "fecha_fin": "2026-10-06", "creado_en": "2026-09-22T18:05:00Z", "actualizado_en": "2026-09-23T09:00:00Z" }
  ]
}
```

Con cero Sprints la respuesta es `{"total": 0, "sprints": []}`: el array vacío se serializa explícitamente para no producir `null`.

`PATCH /api/proyectos/1/sprints/2/estado` con `{"estado": "Cerrado"}` — `200 OK`: devuelve la representación del Sprint ya modificado, con `actualizado_en` refrescado.

#### Cuerpo estándar de error

```json
{
  "error": "VALIDACION",
  "mensaje": "el identificador del sprint es obligatorio",
  "campo": "identificador"
}
```

| Campo | Tipo | Regla |
| :--- | :--- | :--- |
| `error` | string | Código estable en `UPPER_SNAKE_CASE`, para uso programático del cliente |
| `mensaje` | string | Mensaje en español, con la redacción del error de dominio que lo originó |
| `campo` | string | Presente solo cuando el error se refiere a un campo del body o de la ruta; se omite en errores de recurso o de infraestructura |

### 2. Modelo de datos

#### `internal/domain/sprint.go`

```go
type EstadoSprint string

const (
    EstadoSprintPlanificado EstadoSprint = "Planificado"
    EstadoSprintActivo      EstadoSprint = "Activo"
    EstadoSprintCerrado     EstadoSprint = "Cerrado"
)

func (e EstadoSprint) Valido() bool
func (e EstadoSprint) TransicionValidaA(destino EstadoSprint) bool

type Sprint struct {
    ID            uint
    ProyectoID    uint
    Identificador string
    SprintGoal    string
    Estado        EstadoSprint
    FechaInicio   time.Time
    FechaFin      time.Time
    CreadoEn      time.Time
    ActualizadoEn time.Time
}

type NuevoSprint struct {
    Identificador string
    SprintGoal    string
    FechaInicio   time.Time
    FechaFin      time.Time
}
```

La máquina de estados (RB-05) es exclusivamente `Planificado → Activo → Cerrado`: `Cerrado` es terminal y no se admiten transiciones al mismo estado.

#### `internal/domain/sprint_repository.go`

```go
type SprintRepository interface {
    Crear(ctx context.Context, s *Sprint) error
    ObtenerPorID(ctx context.Context, proyectoID, id uint) (*Sprint, error)
    ListarPorProyecto(ctx context.Context, proyectoID uint) ([]Sprint, error)
    ActualizarEstado(ctx context.Context, s *Sprint) error
    ObtenerActivo(ctx context.Context, proyectoID uint) (*Sprint, error)
    HaySolapamiento(ctx context.Context, proyectoID uint, inicio, fin time.Time) (bool, error)
}
```

#### `internal/domain/errores.go` (adiciones)

```go
ErrSprintNoEncontrado             = errors.New("el sprint no existe")
ErrIdentificadorSprintObligatorio = errors.New("el identificador del sprint es obligatorio")
ErrIdentificadorSprintDemasiadoLargo = errors.New("el identificador del sprint no puede superar los 50 caracteres")
ErrIdentificadorSprintDuplicado   = errors.New("ya existe un sprint con ese identificador en el proyecto")
ErrSprintGoalObligatorio          = errors.New("el Sprint Goal del sprint es obligatorio")
ErrSprintGoalDemasiadoLargo       = errors.New("el Sprint Goal del sprint no puede superar los 300 caracteres")
ErrSprintFechaInicioObligatoria   = errors.New("la fecha de inicio del sprint es obligatoria")
ErrSprintFechaFinAnteriorAlInicio = errors.New("la fecha de fin del sprint no puede ser anterior a la fecha de inicio")
ErrSprintFechasSolapadas          = errors.New("las fechas del sprint se solapan con otro sprint del proyecto")
ErrTransicionSprintInvalida       = errors.New("la transición de estado del sprint no está permitida")
ErrSprintActivoExistente          = errors.New("ya existe un sprint activo en el proyecto: cerrá ese sprint antes de activar otro")
ErrEstadoSprintInvalido           = errors.New(`el estado del sprint debe ser "Planificado", "Activo" o "Cerrado"`)
```

### 3. Persistencia

#### `internal/repository/models/sprint_model.go`

Tabla `sprints`, con índice único compuesto `(proyecto_id, identificador)` que garantiza la unicidad del identificador dentro de cada proyecto, y clave foránea a `proyectos`.

```go
type SprintModel struct {
    ID            uint      `gorm:"column:id;primaryKey;autoIncrement"`
    ProyectoID    uint      `gorm:"column:proyecto_id;not null;uniqueIndex:ux_sprints_proyecto_identificador,priority:1"`
    Identificador string    `gorm:"column:identificador;type:varchar(50);not null;uniqueIndex:ux_sprints_proyecto_identificador,priority:2"`
    SprintGoal    string    `gorm:"column:sprint_goal;type:text;not null"`
    Estado        string    `gorm:"column:estado;type:varchar(20);not null;default:'Planificado'"`
    FechaInicio   time.Time `gorm:"column:fecha_inicio;type:timestamptz;not null"`
    FechaFin      time.Time `gorm:"column:fecha_fin;type:timestamptz;not null"`
    CreadoEn      time.Time `gorm:"column:creado_en;type:timestamptz;not null;default:now()"`
    ActualizadoEn time.Time `gorm:"column:actualizado_en;type:timestamptz;not null;default:now()"`
    Proyecto      ProyectoModel `gorm:"foreignKey:ProyectoID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
```

#### `internal/repository/sprint_repository.go` (resumen de comportamiento)

| Método | Comportamiento |
| :--- | :--- |
| `Crear` | Escribe `estado` `Planificado` si falta, `creado_en`/`actualizado_en` si faltan, inserta con `Omit("Proyecto")` y copia `id`, `creado_en`, `actualizado_en` generados por la base. Traduce `gorm.ErrDuplicatedKey` a `ErrIdentificadorSprintDuplicado`. |
| `ObtenerPorID` | Filtra por `id` **y** `proyecto_id`; un Sprint de otro proyecto se comporta como inexistente (`ErrSprintNoEncontrado`). |
| `ListarPorProyecto` | Devuelve los Sprints ordenados por `fecha_inicio ASC, id ASC`; slice vacío (no nil) si no hay Sprints. |
| `ActualizarEstado` | Actualiza `estado` y `actualizado_en`; si `RowsAffected == 0` devuelve `ErrSprintNoEncontrado`. |
| `ObtenerActivo` | Primer Sprint del proyecto con `estado = 'Activo'`; `(nil, nil)` si no existe. |
| `HaySolapamiento` | `TRUE` si existe un Sprint del proyecto cuyo rango se interseca estrictamente con `[inicio, fin]`: `fecha_inicio < finNuevo AND fecha_fin > inicioNuevo`. Los Sprints contiguos (la fin de uno es igual a la inicio del siguiente) **no** se consideran solapados. |

### 4. DDL resultante de `AutoMigrate`

```sql
CREATE TABLE sprints (
    id             bigserial   PRIMARY KEY,
    proyecto_id    bigint      NOT NULL,
    identificador  varchar(50) NOT NULL,
    sprint_goal    text        NOT NULL,
    estado         varchar(20) NOT NULL DEFAULT 'Planificado',
    fecha_inicio   timestamptz NOT NULL,
    fecha_fin      timestamptz NOT NULL,
    creado_en      timestamptz NOT NULL DEFAULT now(),
    actualizado_en timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT fk_sprints_proyecto
        FOREIGN KEY (proyecto_id) REFERENCES proyectos(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT ux_sprints_proyecto_identificador UNIQUE (proyecto_id, identificador)
);
```

### 5. Archivos a crear y modificar

| Acción | Archivo |
| :--- | :--- |
| Crear | `internal/domain/sprint.go` |
| Crear | `internal/domain/sprint_repository.go` |
| Crear | `internal/domain/sprint_test.go` |
| Crear | `internal/repository/models/sprint_model.go` |
| Crear | `internal/repository/sprint_repository.go` |
| Crear | `internal/repository/sprint_repository_test.go` |
| Crear | `internal/service/sprint_service.go` |
| Crear | `internal/service/sprint_service_test.go` |
| Crear | `internal/handler/dto_sprint.go` |
| Crear | `internal/handler/sprint_handler.go` |
| Crear | `internal/handler/sprint_handler_test.go` |
| Crear | `features/US-03-creacion-y-configuracion-de-sprints.feature` |
| Crear | `docs/specs/US-03-creacion-y-configuracion-de-sprints-spec.md` |
| Modificar | `internal/domain/errores.go` (errores centinela) |
| Modificar | `internal/repository/models/mappers.go` (`ToDomainSprint`, `FromDomainSprint`, `ToDomainSprints`) |
| Modificar | `internal/repository/migraciones.go` (agregar `SprintModel`) |
| Modificar | `internal/repository/migraciones_test.go` (tabla `sprints` en limpieza y verificación) |
| Modificar | `internal/service/mocks_test.go` (`MockSprintRepository`) |
| Modificar | `internal/handler/errores.go` (códigos y mapeos) |
| Modificar | `internal/handler/mocks_test.go` (`mockServicioSprints`) |
| Modificar | `cmd/api/router.go` (`RegistrarRutasSprints`) |
| Modificar | `cmd/api/main.go` (wiring del repositorio y servicio) |
| Modificar | `cmd/api/router_test.go` y `cmd/api/main_test.go` (nueva firma de `setupRouter`) |
| Modificar | `README.md` (endpoints de US-03) |

### 6. Plan de pruebas

| Capa | Archivo | Qué verifica |
| :--- | :--- | :--- |
| Dominio | `internal/domain/sprint_test.go` | `Valido()` y `TransicionValidaA()` (matriz completa de la máquina de estados). |
| Servicio | `internal/service/sprint_service_test.go` | Validaciones, TrimSpace, límites 50/300, unicidad, solapamiento, máquina de estados, un solo sprint activo, propagación de errores. |
| Handler | `internal/handler/sprint_handler_test.go` | Códigos HTTP, cuerpos de respuesta, mapeo de errores, validación de path, de fechas y de JSON. |
| Repositorio | `internal/repository/sprint_repository_test.go` | Integración real: índice único, ObtenerPorID por proyecto, ObtenerActivo, HaySolapamiento (incl. contiguos), ActualizarEstado, auditoría. |
| Router | `cmd/api/router_test.go` | Las tres rutas están registradas y alcanzan el servicio inyectado. |

## Reglas de negocio

| ID | Regla |
| :--- | :--- |
| RB-01 | El identificador del Sprint es obligatorio: después de `TrimSpace` debe tener entre 1 y 50 caracteres. |
| RB-02 | El Sprint Goal es obligatorio: después de `TrimSpace` debe tener entre 1 y 300 caracteres. |
| RB-03 | Ambas fechas son obligatorias y `fecha_fin` no puede ser anterior a `fecha_inicio` (un Sprint de un solo día es válido). |
| RB-04 | El identificador es único dentro del proyecto: no puede existir otro Sprint con el mismo identificador en el mismo proyecto (en otro proyecto sí está permitido). |
| RB-05 | La máquina de estados admite únicamente `Planificado → Activo` y `Activo → Cerrado`; `Cerrado` es estado terminal y no se admite la transición al mismo estado. |
| RB-06 | Solo un Sprint puede estar `Activo` a la vez por proyecto: activar un Sprint mientras otro está `Activo` se rechaza. |
| RB-07 | Las fechas de un Sprint no pueden solaparse con las de otro Sprint del mismo proyecto; los Sprints contiguos (una fecha de fin igual a la fecha de inicio siguiente) están permitidos. |
| RB-08 | El estado inicial por defecto de un Sprint recién creado es `Planificado` y no es elegible por el cliente. |
| RB-09 | El listado de Sprints es por proyecto, ordenado por `fecha_inicio ASC` y desempatado por `id ASC`. |
| RB-10 | Una operación sobre un proyecto inexistente se rechaza con `PROYECTO_NO_ENCONTRADO`. |
| RB-11 | Una operación sobre un Sprint inexistente (o perteneciente a otro proyecto) se rechaza con `SPRINT_NO_ENCONTRADA`. |

## Decisiones de diseño y su justificación

| ID | Decisión | Justificación |
| :--- | :--- | :--- |
| D-01 | El cambio de estado usa `PATCH /sprints/{sprintId}/estado` con `{"estado": "..."}`, en lugar de un `PUT` del Sprint completo. | El único campo mutable en US-03 es el estado; un reemplazo completo obligaría a reenviar fechas y goal que no cambian. El patrón replica `PATCH /proyectos/{id}/estado` de US-01 (consistencia de API). |
| D-02 | La tabla y la FK se generan con `AutoMigrate` en el arranque. | Coherente con US-01 y US-02: el esquema es reproducible y se mantiene sin migraciones manuales. |
| D-03 | El estado no se acepta del cliente en el alta. | El requerimiento fija `Planificado` como estado inicial por defecto; aceptar el estado en el alta abriría la puerta a crear un Sprint `Activo` sin cumplir RB-06 ni la máquina de estados. |
| D-04 | La representación de un Sprint no emite `proyecto_id`. | Va implícito en la ruta anidada; mismo criterio que D-03 de US-02 y D-05 de US-01. |
| D-05 | La unicidad del identificador se resuelve en la base y se traduce a `409 SPRINT_DUPLICADO`. | El índice único `(proyecto_id, identificador)` convierte cualquier carrera en un conflicto controlado, sin depender de una consulta previa. |
| D-06 | Los errores de validación de campos responden `422` con `campo`; el literal de estado inválido responde `400 ESTADO_INVALIDO`. | Consistente con US-01/US-02: un literal fuera del conjunto cerrado es un error de forma (400), mientras que un valor válido pero con defecto de negocio (vacío, demasiado largo, orden de fechas) es un error de semántica (422). |
| D-07 | El solapamiento se define con desigualdades estrictas (`fecha_inicio < finNuevo AND fecha_fin > inicioNuevo`). | Los Sprints son intervalos semiabiertos en la práctica: un Sprint termina el mismo día en que comienza el siguiente sin solaparse. La regla queda explícita en RB-07 y cubierta por el escenario límite E-17. |
| D-08 | La verificación de "un solo Sprint activo" se hace en el service consultando `ObtenerActivo` antes de persistir. | Es una regla de negocio cruzada entre filas, no una restricción de columna; la consulta la encapsula el repositorio y el service decide. |

## Restricciones

1. Arquitectura por capas: `domain → repository → service → handler`; las dependencias apuntan hacia el dominio.
2. El `domain` no importa Gin, GORM ni ningún paquete de infraestructura.
3. Los mensajes de error replican literalmente los errores centinela definidos en `internal/domain/errores.go`.
4. El handler no contiene reglas de negocio: solo traduce HTTP ↔ dominio, parsea las fechas `YYYY-MM-DD` y aplica `TrimSpace` vía el service.
5. El service no conoce `*gorm.DB` ni `*gin.Context`; recibe `context.Context` y las interfaces de `domain`.
6. Los escenarios BDD usan el dialecto `# language: es` y el identificador `E-nn` como nombre de subtest.
7. Cada caso de prueba se anota con el escenario BDD que lo origina.

## Casos límite

| ID | Caso | Resultado |
| :--- | :--- | :--- |
| CL-01 | Identificador de exactamente 50 caracteres | Válido (`201`). |
| CL-02 | Identificador de 51 caracteres | `422 VALIDACION` campo `identificador`. |
| CL-03 | Identificador vacío o solo espacios | `422 VALIDACION` campo `identificador`. |
| CL-04 | Sprint Goal de exactamente 300 caracteres | Válido (`201`). |
| CL-05 | Sprint Goal de 301 caracteres | `422 VALIDACION` campo `sprint_goal`. |
| CL-06 | Sprint Goal vacío o solo espacios | `422 VALIDACION` campo `sprint_goal`. |
| CL-07 | `fecha_fin == fecha_inicio` (Sprint de un solo día) | Válido (`201`). |
| CL-08 | `fecha_fin < fecha_inicio` | `422 VALIDACION` campo `fecha_fin`. |
| CL-09 | Identificador ya tomado en el mismo proyecto (carrera) | `409 SPRINT_DUPLICADO`. |
| CL-10 | Mismo identificador en otro proyecto | Válido (`201`). |
| CL-11 | Solapamiento parcial de rangos | `409 SPRINT_FECHAS_SOLAPADAS`. |
| CL-12 | Un Sprint contenido completamente en el rango de otro | `409 SPRINT_FECHAS_SOLAPADAS`. |
| CL-13 | Sprints contiguos: `fecha_fin` de uno == `fecha_inicio` del otro | Válido (`201`). |
| CL-14 | `PATCH` a `Planificado → Activo` sin ningún Sprint activo previo | `200`, el Sprint queda como activo del proyecto. |
| CL-15 | `PATCH` `Activo → Activo` (misma estado) | `409 SPRINT_TRANSICION_INVALIDA`. |
| CL-16 | `PATCH` `Planificado → Cerrado` (saltear `Activo`) | `409 SPRINT_TRANSICION_INVALIDA`. |
| CL-17 | Activar un Sprint existiendo otro `Activo` | `409 SPRINT_ACTIVO_EXISTENTE`. |
| CL-18 | `GET` de un proyecto sin Sprints | `200` con `{"total":0,"sprints":[]}`. |
| CL-19 | Sprint de otro proyecto: `PATCH`/`ObtenerPorID` | `404 SPRINT_NO_ENCONTRADA`. |
| CL-20 | `{id}` o `{sprintId}` no numérico o `0` | `400 PARAMETRO_INVALIDO`. |
| CL-21 | `fecha_inicio` ausente o no parseable | `400 PARAMETRO_INVALIDO` campo `fecha_inicio`. |
| CL-22 | `fecha_fin` ausente o no parseable | `400 PARAMETRO_INVALIDO` campo `fecha_fin`. |

## Condiciones de error

### Errores centinela del dominio

| Error | Mensaje | `campo` |
| :--- | :--- | :--- |
| `ErrSprintNoEncontrado` | el sprint no existe | — |
| `ErrIdentificadorSprintObligatorio` | el identificador del sprint es obligatorio | `identificador` |
| `ErrIdentificadorSprintDemasiadoLargo` | el identificador del sprint no puede superar los 50 caracteres | `identificador` |
| `ErrIdentificadorSprintDuplicado` | ya existe un sprint con ese identificador en el proyecto | `identificador` |
| `ErrSprintGoalObligatorio` | el Sprint Goal del sprint es obligatorio | `sprint_goal` |
| `ErrSprintGoalDemasiadoLargo` | el Sprint Goal del sprint no puede superar los 300 caracteres | `sprint_goal` |
| `ErrSprintFechaInicioObligatoria` | la fecha de inicio del sprint es obligatoria | `fecha_inicio` |
| `ErrSprintFechaFinAnteriorAlInicio` | la fecha de fin del sprint no puede ser anterior a la fecha de inicio | `fecha_fin` |
| `ErrSprintFechasSolapadas` | las fechas del sprint se solapan con otro sprint del proyecto | — |
| `ErrTransicionSprintInvalida` | la transición de estado del sprint no está permitida | — |
| `ErrSprintActivoExistente` | ya existe un sprint activo en el proyecto: cerrá ese sprint antes de activar otro | — |
| `ErrEstadoSprintInvalido` | el estado del sprint debe ser "Planificado", "Activo" o "Cerrado" | `estado` (HTTP 400) |

### Mapeo de error de dominio a respuesta HTTP

| Error de dominio | `error` | HTTP | `campo` |
| :--- | :--- | :---: | :--- |
| `ErrProyectoNoEncontrado` | `PROYECTO_NO_ENCONTRADO` | `404` | — |
| `ErrSprintNoEncontrado` | `SPRINT_NO_ENCONTRADA` | `404` | — |
| `ErrIdentificadorSprintDuplicado` | `SPRINT_DUPLICADO` | `409` | `identificador` |
| `ErrSprintFechasSolapadas` | `SPRINT_FECHAS_SOLAPADAS` | `409` | — |
| `ErrSprintActivoExistente` | `SPRINT_ACTIVO_EXISTENTE` | `409` | — |
| `ErrTransicionSprintInvalida` | `SPRINT_TRANSICION_INVALIDA` | `409` | — |
| `ErrIdentificadorSprintObligatorio` / `ErrIdentificadorSprintDemasiadoLargo` | `VALIDACION` | `422` | `identificador` |
| `ErrSprintGoalObligatorio` / `ErrSprintGoalDemasiadoLargo` | `VALIDACION` | `422` | `sprint_goal` |
| `ErrSprintFechaInicioObligatoria` | `VALIDACION` | `422` | `fecha_inicio` |
| `ErrSprintFechaFinAnteriorAlInicio` | `VALIDACION` | `422` | `fecha_fin` |
| `ErrEstadoSprintInvalido` | `ESTADO_INVALIDO` | `400` | `estado` |
| Path no numérico o `0` | `PARAMETRO_INVALIDO` | `400` | — |
| Fecha ausente o mal formada | `PARAMETRO_INVALIDO` | `400` | `fecha_inicio` / `fecha_fin` |
| JSON malformado | `JSON_INVALIDO` | `400` | — |

### Ejemplos de respuesta de error

`POST /api/proyectos/1/sprints` con `{"identificador":"Sprint 1","sprint_goal":"Goal","fecha_inicio":"2026-10-06","fecha_fin":"2026-10-05"}`:

```json
{ "error": "VALIDACION", "mensaje": "la fecha de fin del sprint no puede ser anterior a la fecha de inicio", "campo": "fecha_fin" }
```

`PATCH /api/proyectos/1/sprints/5/estado` con `{"estado":"Planificado"}` sobre un Sprint `Cerrado`:

```json
{ "error": "SPRINT_TRANSICION_INVALIDA", "mensaje": "la transición de estado del sprint no está permitida" }
```

`GET /api/proyectos/99/sprints`:

```json
{ "error": "PROYECTO_NO_ENCONTRADO", "mensaje": "el proyecto no existe" }
```

## Criterios de aceptación

### Funcionales

- **CA-I1.** Cada Sprint expone `id`, `identificador`, `sprint_goal`, `estado`, `fecha_inicio`, `fecha_fin`, `creado_en` y `actualizado_en`.
- **CA-I2.** Se puede crear un Sprint con identificador, goal y fechas, y queda registrado con estado `Planificado`.
- **CA-I3.** Se pueden listar los Sprints de un proyecto con `{"total": n, "sprints": [...]}`.
- **CA-I4.** Se puede cambiar el estado de un Sprint según la máquina de estados, y al activarlo pasa a ser el Sprint activo del proyecto.

### Reglas de negocio

- **CA-R1.** Identificador obligatorio de 1 a 50 caracteres y único por proyecto (RB-01, RB-04).
- **CA-R2.** Sprint Goal obligatorio de 1 a 300 caracteres (RB-02).
- **CA-R3.** `fecha_fin >= fecha_inicio` y sin solapamiento con otros Sprints del proyecto (RB-03, RB-07).
- **CA-R4.** Transiciones válidas únicamente `Planificado → Activo` y `Activo → Cerrado`; `Cerrado` es terminal (RB-05).
- **CA-R5.** Solo un Sprint activo por proyecto (RB-06).
- **CA-R6.** Alta o cambio de estado sobre proyecto inexistente → `404` (RB-10); Sprint inexistente o de otro proyecto → `404` (RB-11).

### Arquitectura y calidad

- **CA-A1.** Respeta las capas y el dominio no depende de infraestructura.
- **CA-A2.** Cobertura de pruebas ≥ 80 % en `internal/...`.
- **CA-A3.** `go build ./...`, `go vet ./...` y `go test ./...` sin errores.

## Escenarios BDD

Los escenarios completos viven en `features/US-03-creacion-y-configuracion-de-sprints.feature` con identificadores `E-01`..`E-41` y se referencian como nombre de subtest en cada capa:

| Regla | Escenarios | Capa principal |
| :--- | :--- | :--- |
| Alta de Sprint | `E-01`..`E-24` | service / handler |
| Listado de Sprints | `E-25`..`E-28` | service / handler |
| Cambio de estado | `E-29`..`E-41` | service / handler / repository |

Trazabilidad: **Issue #30 → esta especificación → criterios de aceptación → escenarios BDD → tests Go → código**.
