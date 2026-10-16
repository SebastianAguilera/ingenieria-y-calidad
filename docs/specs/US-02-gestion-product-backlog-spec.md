# Especificación: US-02 - Gestión básica del Product Backlog

## Objetivo

Implementar el módulo de **gestión básica del Product Backlog** del *Software Metrics & Estimation Engine*: agregar, listar y modificar **historias de usuario** dentro de un proyecto. La funcionalidad satisface el requerimiento 2 de `docs/proyecto.md` (Product Backlog compuesto por historias con identificador, título, descripción, prioridad y estado) y constituye la base sobre la que se construirán los Sprints, la estimación por Story Points, el Planning Poker, los worklogs, los defectos y el motor de métricas.

El resultado es un conjunto de tres endpoints REST servidos por Gin, con las reglas de negocio en Go, persistencia en PostgreSQL mediante GORM y esquema generado por `AutoMigrate`. Las historias cuelgan de la entidad raíz definida en US-01 (`proyectos`): cada historia pertenece exactamente a un proyecto y su código visible `US-NN` es un consecutivo por proyecto.

**Alcance.** La historia cubre tres operaciones: agregar (alta), listar y modificar. No se especifica la eliminación de historias ni la reordenación manual del backlog: el orden del listado es el orden de alta.

**Trazabilidad:** la funcionalidad se desarrolla en la rama `feature/US-02-gestion-product-backlog` bajo el ciclo TDD (RED → GREEN → REFACTOR). Los escenarios BDD de `features/US-02-gestion-product-backlog.feature` (identificadores `E-01`..`E-34`) son el origen de los casos de prueba de cada capa y cada criterio de aceptación es verificable por un test concreto.

## Entradas

### Precondiciones

- La aplicación está configurada mediante `.env` (tomando `.env.example` como referencia) e `internal/repository.ConnectDB()` establece la conexión con PostgreSQL.
- El esquema de US-01 existe y el proyecto referenciado por la ruta fue creado previamente: las historias son un recurso anidado y no pueden existir de forma huérfana.
- El esquema de la tabla `historias` se crea automáticamente en el arranque mediante `AutoMigrate` (decisión D-02), incluida la clave foránea hacia `proyectos`.
- No se implementa autenticación ni autorización en US-02: el actor es implícitamente un administrador del sistema.

### Convenciones de formato de entrada

| Convención | Regla | Consecuencia si se viola |
| :--- | :--- | :--- |
| Identificadores de recurso | Entero positivo en el path (`/api/proyectos/{id}/historias/{historiaId}`), leído con `strconv.ParseUint` | `400 PARAMETRO_INVALIDO` si no es numérico o es `0` |
| Título | Texto de 1 a 200 caracteres después de `TrimSpace` | `422 VALIDACION` (campo `titulo`) |
| Descripción | Texto libre, opcional; puede ser vacío o ausente | — |
| Prioridad | Literal `"Alta"`, `"Media"` o `"Baja"` (distingue mayúsculas); ausente o vacío toma `"Media"` | `422 VALIDACION` (campo `prioridad`) |
| Estado | Literal `"Backlog"`, `"En Progreso"` o `"Hecho"` (distingue mayúsculas); ausente o vacío toma `"Backlog"` | `422 VALIDACION` (campo `estado`) |
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

| Campo | Tipo | Regla |
| :--- | :--- | :--- |
| `error` | string | Código estable en `UPPER_SNAKE_CASE`, para uso programático del cliente |
| `mensaje` | string | Mensaje en español, con la redacción del error de dominio que lo originó |
| `campo` | string | Presente solo cuando el error se refiere a un campo del body; se omite en errores de ruta, de recurso o de infraestructura |

### 2. Modelo de datos

#### `internal/domain/historia.go`

```go
type PrioridadHistoria string

const (
    PrioridadAlta  PrioridadHistoria = "Alta"
    PrioridadMedia PrioridadHistoria = "Media"
    PrioridadBaja  PrioridadHistoria = "Baja"
)

func (p PrioridadHistoria) Valido() bool

type EstadoHistoria string

const (
    EstadoHistoriaBacklog    EstadoHistoria = "Backlog"
    EstadoHistoriaEnProgreso EstadoHistoria = "En Progreso"
    EstadoHistoriaHecho      EstadoHistoria = "Hecho"
)

func (e EstadoHistoria) Valido() bool

type Historia struct {
    ID            uint
    ProyectoID    uint
    Numero        int
    Codigo        string
    Titulo        string
    Descripcion   string
    Prioridad     PrioridadHistoria
    Estado        EstadoHistoria
    CreadoEn      time.Time
    ActualizadoEn time.Time
}

type NuevaHistoria struct {
    Titulo      string
    Descripcion string
    Prioridad   PrioridadHistoria
    Estado      EstadoHistoria
}

type ActualizacionHistoria struct {
    Titulo      *string
    Descripcion *string
    Prioridad   *PrioridadHistoria
    Estado      *EstadoHistoria
}

func FormatearCodigoHistoria(numero int) string // fmt.Sprintf("US-%02d", numero)
```

`ActualizacionHistoria` usa punteros para distinguir "no informado" (nil) de "informado con valor".

#### `internal/domain/historia_repository.go`

```go
type HistoriaRepository interface {
    Crear(ctx context.Context, h *Historia) error
    SiguienteNumero(ctx context.Context, proyectoID uint) (int, error)
    ObtenerPorID(ctx context.Context, proyectoID, id uint) (*Historia, error)
    ListarPorProyecto(ctx context.Context, proyectoID uint) ([]Historia, error)
    Actualizar(ctx context.Context, h *Historia) error
}
```

#### `internal/domain/errores.go` (adiciones)

```go
ErrHistoriaNoEncontrada         = errors.New("la historia no existe")
ErrTituloHistoriaObligatorio    = errors.New("el título de la historia es obligatorio")
ErrTituloHistoriaDemasiadoLargo = errors.New("el título de la historia no puede superar los 200 caracteres")
ErrPrioridadHistoriaInvalida    = errors.New(`la prioridad de la historia debe ser "Alta", "Media" o "Baja"`)
ErrEstadoHistoriaInvalido       = errors.New(`el estado de la historia debe ser "Backlog", "En Progreso" o "Hecho"`)
ErrActualizacionHistoriaVacia   = errors.New("debés informar al menos un campo para modificar la historia")
ErrCodigoHistoriaDuplicado      = errors.New("ya existe una historia con ese código en el proyecto")
```

### 3. Persistencia

#### `internal/repository/models/historia_model.go`

Tabla `historias`, con índice único compuesto `(proyecto_id, numero)` que garantiza la unicidad del código `US-NN` dentro de cada proyecto, y clave foránea a `proyectos`.

```go
type HistoriaModel struct {
    ID            uint      `gorm:"column:id;primaryKey;autoIncrement"`
    ProyectoID    uint      `gorm:"column:proyecto_id;not null;uniqueIndex:ux_historias_proyecto_numero,priority:1"`
    Numero        int       `gorm:"column:numero;not null;uniqueIndex:ux_historias_proyecto_numero,priority:2"`
    Codigo        string    `gorm:"column:codigo;type:varchar(20);not null"`
    Titulo        string    `gorm:"column:titulo;type:varchar(200);not null"`
    Descripcion   string    `gorm:"column:descripcion;type:text"`
    Prioridad     string    `gorm:"column:prioridad;type:varchar(20);not null;default:'Media'"`
    Estado        string    `gorm:"column:estado;type:varchar(20);not null;default:'Backlog'"`
    CreadoEn      time.Time `gorm:"column:creado_en;type:timestamptz;not null;default:now()"`
    ActualizadoEn time.Time `gorm:"column:actualizado_en;type:timestamptz;not null;default:now()"`
    Proyecto      ProyectoModel `gorm:"foreignKey:ProyectoID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
```

#### `internal/repository/historia_repository.go` (resumen de comportamiento)

| Método | Comportamiento |
| :--- | :--- |
| `Crear` | Escribe `creado_en`/`actualizado_en` si faltan, inserta con `Omit("Proyecto")` y copia `id`, `creado_en`, `actualizado_en` generados por la base. Traduce `gorm.ErrDuplicatedKey` a `ErrCodigoHistoriaDuplicado`. |
| `SiguienteNumero` | `MAX(numero)` del proyecto + 1; devuelve `1` si el proyecto no tiene historias. |
| `ObtenerPorID` | Filtra por `id` **y** `proyecto_id`; una historia de otro proyecto se comporta como inexistente (`ErrHistoriaNoEncontrada`). |
| `ListarPorProyecto` | Devuelve el backlog ordenado por `numero ASC`; slice vacío (no nil) si no hay historias. |
| `Actualizar` | Actualiza `titulo`, `descripcion`, `prioridad`, `estado` y `actualizado_en`; si `RowsAffected == 0` devuelve `ErrHistoriaNoEncontrada`. |

### 4. DDL resultante de `AutoMigrate`

```sql
CREATE TABLE historias (
    id             bigserial   PRIMARY KEY,
    proyecto_id    bigint      NOT NULL,
    numero         bigint      NOT NULL,
    codigo         varchar(20) NOT NULL,
    titulo         varchar(200) NOT NULL,
    descripcion    text,
    prioridad      varchar(20) NOT NULL DEFAULT 'Media',
    estado         varchar(20) NOT NULL DEFAULT 'Backlog',
    creado_en      timestamptz NOT NULL DEFAULT now(),
    actualizado_en timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT fk_historias_proyecto
        FOREIGN KEY (proyecto_id) REFERENCES proyectos(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT ux_historias_proyecto_numero UNIQUE (proyecto_id, numero)
);
```

### 5. Archivos a crear y modificar

| Acción | Archivo |
| :--- | :--- |
| Crear | `internal/domain/historia.go` |
| Crear | `internal/domain/historia_repository.go` |
| Crear | `internal/domain/historia_test.go` |
| Crear | `internal/repository/models/historia_model.go` |
| Crear | `internal/repository/historia_repository.go` |
| Crear | `internal/repository/historia_repository_test.go` |
| Crear | `internal/service/historia_service.go` |
| Crear | `internal/service/historia_service_test.go` |
| Crear | `internal/handler/dto_historia.go` |
| Crear | `internal/handler/historia_handler.go` |
| Crear | `internal/handler/historia_handler_test.go` |
| Crear | `features/US-02-gestion-product-backlog.feature` |
| Crear | `docs/specs/US-02-gestion-product-backlog-spec.md` |
| Modificar | `internal/domain/errores.go` (errores centinela) |
| Modificar | `internal/repository/models/mappers.go` (`ToDomainHistoria`, `FromDomainHistoria`, `ToDomainHistorias`) |
| Modificar | `internal/repository/migraciones.go` (agregar `HistoriaModel`) |
| Modificar | `internal/repository/migraciones_test.go` (tabla `historias` en limpieza y verificación) |
| Modificar | `internal/service/mocks_test.go` (`MockHistoriaRepository`) |
| Modificar | `internal/handler/errores.go` (códigos y mapeos) |
| Modificar | `internal/handler/mocks_test.go` (`mockServicioHistorias`) |
| Modificar | `cmd/api/router.go` (`RegistrarRutasHistorias`) |
| Modificar | `cmd/api/main.go` (wiring del repositorio y servicio) |
| Modificar | `cmd/api/router_test.go` y `cmd/api/main_test.go` (nueva firma de `setupRouter`) |
| Modificar | `README.md` (endpoints de US-02) |

### 6. Plan de pruebas

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

## Decisiones de diseño y su justificación

| ID | Decisión | Justificación |
| :--- | :--- | :--- |
| D-01 | La modificación usa `PATCH` con actualización parcial mediante punteros. | El requerimiento indica modificar título, descripción, prioridad **y/o** estado; un reemplazo completo (`PUT`) obligaría al cliente a reenviar campos que no quiere cambiar. |
| D-02 | La tabla y la FK se generan con `AutoMigrate` en el arranque. | Coherente con US-01: el esquema es reproducible y se mantiene sin migraciones manuales. |
| D-03 | La representación de una historia no emite `proyecto_id`. | Va implícito en la ruta anidada; emitirlo sería redundante (mismo criterio que D-05 de US-01). |
| D-04 | El código `US-NN` se deriva del `numero` consecutivo y no se acepta del cliente. | Evita colisiones y hace predecible la nomenclatura; el índice único `(proyecto_id, numero)` es la garantía última. |
| D-05 | La unicidad se resuelve en la base y se traduce a `409 HISTORIA_DUPLICADA`. | En altas concurrentes el `SiguienteNumero` puede repetirse; el índice único convierte la carrera en un conflicto controlado. |
| D-06 | Los errores de validación se responden con `422` y `campo`. | Consistente con US-01: el cliente puede señalar el campo inválido de forma programática. |

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

### Errores centinela del dominio

| Error | Mensaje | `campo` |
| :--- | :--- | :--- |
| `ErrHistoriaNoEncontrada` | la historia no existe | — |
| `ErrTituloHistoriaObligatorio` | el título de la historia es obligatorio | `titulo` |
| `ErrTituloHistoriaDemasiadoLargo` | el título de la historia no puede superar los 200 caracteres | `titulo` |
| `ErrPrioridadHistoriaInvalida` | la prioridad de la historia debe ser "Alta", "Media" o "Baja" | `prioridad` |
| `ErrEstadoHistoriaInvalido` | el estado de la historia debe ser "Backlog", "En Progreso" o "Hecho" | `estado` |
| `ErrActualizacionHistoriaVacia` | debés informar al menos un campo para modificar la historia | — |
| `ErrCodigoHistoriaDuplicado` | ya existe una historia con ese código en el proyecto | `codigo` |

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

### Ejemplos de respuesta de error

`POST /api/proyectos/1/historias` con `{"titulo": ""}`:

```json
{ "error": "VALIDACION", "mensaje": "el título de la historia es obligatorio", "campo": "titulo" }
```

`PATCH /api/proyectos/1/historias/5` con `{}`:

```json
{ "error": "VALIDACION", "mensaje": "debés informar al menos un campo para modificar la historia" }
```

`GET /api/proyectos/99/historias`:

```json
{ "error": "PROYECTO_NO_ENCONTRADO", "mensaje": "el proyecto no existe" }
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

- **CA-A1.** Respeta las capas y el dominio no depende de infraestructura.
- **CA-A2.** Cobertura de pruebas ≥ 80 % en `internal/...`.
- **CA-A3.** `go build ./...`, `go vet ./...` y `go test ./...` sin errores.

## Escenarios BDD

Los escenarios completos viven en `features/US-02-gestion-product-backlog.feature` con identificadores `E-01`..`E-34` y se referencian como nombre de subtest en cada capa:

| Regla | Escenarios | Capa principal |
| :--- | :--- | :--- |
| Alta de historia | `E-01`..`E-17` | service / handler |
| Listado del backlog | `E-18`..`E-22` | service / handler / repository |
| Modificación de historia | `E-23`..`E-34` | service / handler / repository |

Trazabilidad: **Issue #29 → esta especificación → criterios de aceptación → escenarios BDD → tests Go → código**.
