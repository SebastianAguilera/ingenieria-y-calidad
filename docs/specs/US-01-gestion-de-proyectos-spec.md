# Especificación: US-01 - Creación y gestión de proyectos

## Objetivo

Implementar el módulo de **gestión de proyectos** del *Software Metrics & Estimation Engine*: alta, consulta, modificación, baja y cambio de estado de proyectos, junto con la administración del conjunto de **usuarios** asociados a cada proyecto. La funcionalidad satisface el requerimiento 1 de `docs/proyecto.md` (crear y modificar proyectos, registrar usuarios, registrar fecha de inicio y finalización, consultar el estado de un proyecto) y constituye la base sobre la que se construirán el Product Backlog, la gestión de Sprints, el registro de esfuerzo, la gestión de defectos y el motor de métricas.

El resultado es una API REST servida por Gin, con las reglas de negocio implementadas en Go, persistencia en PostgreSQL mediante GORM y esquema generado por `AutoMigrate` en el arranque de la aplicación. El proyecto es la entidad raíz del dominio: su identificador será la clave foránea de todas las entidades futuras (historias, Sprints, worklogs, defectos y métricas), por lo que el modelo de datos definido en esta especificación es un contrato estable para el resto de las historias de usuario.

**Trazabilidad:** la funcionalidad se desarrolla en la rama `feature/US-01-gestion-proyectos` bajo el ciclo TDD (RED → GREEN → REFACTOR). Los escenarios BDD de esta especificación son el origen de los casos de prueba de la capa de servicio y de la capa HTTP, y cada criterio de aceptación es verificable por un test concreto.

## Entradas

### Precondiciones

- La aplicación está configurada mediante `.env` (tomando `.env.example` como referencia) e `internal/repository.ConnectDB()` establece la conexión con PostgreSQL.
- El esquema de las tablas de US-01 existe: se crea automáticamente en el arranque de la API mediante `AutoMigrate` (decisión D-02).
- No se implementa autenticación ni autorización en US-01: el actor es implícitamente un administrador del sistema.

### Convenciones de formato de entrada

| Convención | Regla | Consecuencia si se viola |
| :--- | :--- | :--- |
| Fechas de calendario | Texto en ISO 8601, exactamente `YYYY-MM-DD`, interpretado con `time.Parse(time.DateOnly, s)` | `400 PARAMETRO_INVALIDO` (capa handler: error sintáctico) |
| Fechas de auditoría | `YYYY-MM-DD` para `fecha_inicio` y `fecha_fin`; RFC 3339 UTC (`2026-09-28T14:03:11Z`) para `creado_en` y `actualizado_en` | — |
| Email | Texto con exactamente un `@`, sin espacios, longitud máxima 150; se normaliza con `strings.ToLower(strings.TrimSpace(...))` antes de persistir | `422 VALIDACION` (capa service: error semántico) |
| Identificadores de recurso | Entero positivo en el path (`/api/proyectos/1`), leído con `strconv.ParseUint` | `400 PARAMETRO_INVALIDO` si no es numérico, es `0` o es negativo |
| JSON | Un único objeto JSON por request; el handler usa `ShouldBindJSON` | `400 JSON_INVALIDO` si el cuerpo está vacío, malformado o con tipos incorrectos |
| Campos desconocidos | Se ignoran silenciosamente en el body | — |

### Cuerpos de request por operación

#### `POST /api/proyectos` — crear proyecto

| Campo | Tipo JSON | Requerido | Formato / rango | Descripción |
| :--- | :--- | :---: | :--- | :--- |
| `nombre` | string | Sí | 1 a 150 caracteres después de `TrimSpace` | Nombre del proyecto |
| `fecha_inicio` | string | Sí | `YYYY-MM-DD` | Fecha de inicio |
| `fecha_fin` | string o `null` | No | `YYYY-MM-DD`, `null` o ausente | Fecha de finalización; si se informa, no puede ser anterior a `fecha_inicio` |
| `usuarios` | array de objetos | No | Puede enviarse `[]`, omitirse o `null` | Listado inicial de usuarios |

Objeto `usuarios` (cada elemento del array):

| Campo | Tipo JSON | Requerido | Formato / rango | Descripción |
| :--- | :--- | :---: | :--- | :--- |
| `nombre` | string | Sí | 1 a 100 caracteres después de `TrimSpace` | Nombre y apellido del usuario |
| `email` | string | Sí | Formato de email, máximo 150 | Identificador natural de la persona (decisión D-01) |

```json
{
  "nombre": "Software Metrics & Estimation",
  "fecha_inicio": "2026-09-28",
  "fecha_fin": "2026-12-18",
  "usuarios": [
    { "nombre": "Aguilera Sebastián", "email": "sebas.aguilera@utn.edu.ar" },
    { "nombre": "Choquevillca Celeste", "email": "celeste.choque@utn.edu.ar" }
  ]
}
```

#### `PUT /api/proyectos/{id}` — modificar proyecto

| Campo | Tipo JSON | Requerido | Formato / rango | Descripción |
| :--- | :--- | :---: | :--- | :--- |
| `nombre` | string | Sí | 1 a 150 caracteres después de `TrimSpace` | Nuevo nombre |
| `fecha_inicio` | string | Sí | `YYYY-MM-DD` | Nueva fecha de inicio |
| `fecha_fin` | string o `null` | No | `YYYY-MM-DD`, `null` o ausente | Nueva fecha de finalización |

```json
{
  "nombre": "Software Metrics & Estimation Engine",
  "fecha_inicio": "2026-09-28",
  "fecha_fin": "2026-12-20"
}
```

El body de `PUT` **no** incluye `usuarios`: la composición del equipo se modifica exclusivamente mediante los endpoints de usuarios (decisión D-06). Esto evita la semántica ambigua de "reemplazar el conjunto completo" y garantiza que cada cambio de composición quede registrado como una operación explícita y auditable.

#### `PATCH /api/proyectos/{id}/estado` — cambiar estado

| Campo | Tipo JSON | Requerido | Formato / rango | Descripción |
| :--- | :--- | :---: | :--- | :--- |
| `estado` | string | Sí | Literal `"Activo"` o `"Cerrado"` (distingue mayúsculas) | Nuevo estado del proyecto |

```json
{ "estado": "Cerrado" }
```

#### `POST /api/proyectos/{id}/usuarios` — agregar usuario

| Campo | Tipo JSON | Requerido | Formato / rango | Descripción |
| :--- | :--- | :---: | :--- | :--- |
| `nombre` | string | Sí | 1 a 100 caracteres después de `TrimSpace` | Nombre y apellido |
| `email` | string | Sí | Formato de email, máximo 150 | Email identificatorio de la persona |

```json
{ "nombre": "Perez Castro Jazmín", "email": "jazmin.perez@utn.edu.ar" }
```

#### `DELETE /api/proyectos/{id}/usuarios/{usuarioId}` — quitar usuario

Sin cuerpo de request. El parámetro de path `usuarioId` es el identificador del usuario tal como fue devuelto por el endpoint de alta o por el detalle del proyecto; el cliente nunca necesita conocer el email para desvincular a una persona.

## Salidas esperadas

### 1. Diseño de la API REST

Todas las rutas se registran bajo el grupo `/api` desde `cmd/api/router.go`, sobre el `*gin.Engine` creado en `main.go`. La ruta `/health` existente se conserva sin cambios.

| # | Método | Path | Descripción | Éxito | Errores |
| :---: | :--- | :--- | :--- | :---: | :--- |
| 1 | `POST` | `/api/proyectos` | Crea un proyecto y, opcionalmente, sus usuarios iniciales | `201 Created` | `400`, `422`, `500` |
| 2 | `GET` | `/api/proyectos` | Lista los proyectos no dados de baja | `200 OK` | `500` |
| 3 | `GET` | `/api/proyectos/{id}` | Devuelve el detalle de un proyecto con sus usuarios | `200 OK` | `400`, `404`, `500` |
| 4 | `PUT` | `/api/proyectos/{id}` | Modifica nombre y fechas de un proyecto activo | `200 OK` | `400`, `404`, `409`, `422`, `500` |
| 5 | `PATCH` | `/api/proyectos/{id}/estado` | Cambia el estado entre `Activo` y `Cerrado` | `200 OK` | `400`, `404`, `422`, `500` |
| 6 | `DELETE` | `/api/proyectos/{id}` | Da de baja lógicamente el proyecto | `204 No Content` | `400`, `404`, `409`, `500` |
| 7 | `POST` | `/api/proyectos/{id}/usuarios` | Asocia un usuario al proyecto | `201 Created` | `400`, `404`, `409`, `422`, `500` |
| 8 | `DELETE` | `/api/proyectos/{id}/usuarios/{usuarioId}` | Desasocia un usuario del proyecto | `204 No Content` | `400`, `404`, `500` |

No se especifica ningún endpoint adicional en US-01. En particular, **no** existe un endpoint independiente para listar los usuarios de un proyecto: la composición del equipo se devuelve embebida en `GET /api/proyectos/{id}` (decisión D-05).

#### Respuestas de ejemplo

`POST /api/proyectos` — `201 Created`:

```json
{
  "id": 1,
  "nombre": "Software Metrics & Estimation",
  "fecha_inicio": "2026-09-28",
  "fecha_fin": "2026-12-18",
  "estado": "Activo",
  "cantidad_usuarios": 2,
  "usuarios": [
    { "id": 1, "nombre": "Aguilera Sebastián", "email": "sebas.aguilera@utn.edu.ar" },
    { "id": 2, "nombre": "Choquevillca Celeste", "email": "celeste.choque@utn.edu.ar" }
  ],
  "creado_en": "2026-09-28T14:03:11Z",
  "actualizado_en": "2026-09-28T14:03:11Z"
}
```

Los identificadores de los usuarios pueden corresponder a registros preexistentes si el email ya estaba registrado en el sistema (decisión D-01): el usuario con `id` 1 podría tener un `creado_en` de una operación anterior.

`GET /api/proyectos` — `200 OK`:

```json
{
  "total": 2,
  "proyectos": [
    {
      "id": 1,
      "nombre": "Software Metrics & Estimation",
      "fecha_inicio": "2026-09-28",
      "fecha_fin": "2026-12-18",
      "estado": "Activo",
      "cantidad_usuarios": 2,
      "creado_en": "2026-09-28T14:03:11Z",
      "actualizado_en": "2026-09-28T14:03:11Z"
    },
    {
      "id": 2,
      "nombre": "Portal de Autoatencion",
      "fecha_inicio": "2026-10-01",
      "fecha_fin": null,
      "estado": "Activo",
      "cantidad_usuarios": 0,
      "creado_en": "2026-10-01T09:15:00Z",
      "actualizado_en": "2026-10-01T09:15:00Z"
    }
  ]
}
```

Con cero proyectos, la respuesta es `{"total": 0, "proyectos": []}`: el array vacío se serializa explícitamente para no producir `null`.

`GET /api/proyectos/{id}`, `PUT /api/proyectos/{id}` y `PATCH /api/proyectos/{id}/estado` — `200 OK`: devuelven el mismo cuerpo que el ejemplo de alta (`201`), es decir, el recurso completo ya persistido y con la composición de usuarios cargada.

`POST /api/proyectos/{id}/usuarios` — `201 Created`:

```json
{
  "id": 7,
  "nombre": "Perez Castro Jazmín",
  "email": "jazmin.perez@utn.edu.ar",
  "proyecto_id": 1
}
```

El recurso creado es la **asociación** entre el proyecto y la persona; por eso el código de éxito es siempre `201`, incluso cuando el registro de `usuarios` ya existía y solo se agregó el vínculo (decisión D-01).

`DELETE /api/proyectos/{id}` y `DELETE /api/proyectos/{id}/usuarios/{usuarioId}` — `204 No Content`: sin cuerpo de respuesta.

#### Cuerpo estándar de error

```json
{
  "error": "PROYECTO_CERRADO",
  "mensaje": "el proyecto está cerrado: solo se permite cambiar su estado o reabrirlo",
  "campo": "estado"
}
```

| Campo | Tipo | Regla |
| :--- | :--- | :--- |
| `error` | string | Código estable en `UPPER_SNAKE_CASE`, pensado para uso programático del cliente |
| `mensaje` | string | Mensaje en español, con la misma redacción que el error de dominio que lo originó (ver **Condiciones de error**) |
| `campo` | string | Presente únicamente cuando el error se refiere a un campo concreto del body; se omite en errores de ruta, de estado o de infraestructura |

El mapeo completo `error de dominio` → `código HTTP` → `mensaje` está en **Condiciones de error**.

### 2. Modelo de datos

#### `internal/domain/proyecto.go`

```go
package domain

import "time"

// EstadoProyecto representa el estado del ciclo de vida de un proyecto.
// Solo admite los valores declarados en las constantes siguientes.
type EstadoProyecto string

const (
	EstadoActivo  EstadoProyecto = "Activo"
	EstadoCerrado EstadoProyecto = "Cerrado"
)

// Valido indica si el estado corresponde a un valor aceptado por el dominio.
func (e EstadoProyecto) Valido() bool {
	return e == EstadoActivo || e == EstadoCerrado
}

// Proyecto es la entidad raíz del dominio.
type Proyecto struct {
	ID            uint
	Nombre        string
	FechaInicio   time.Time
	FechaFin      *time.Time
	Estado        EstadoProyecto
	BorradoEn     *time.Time
	CreadoEn      time.Time
	ActualizadoEn time.Time

	Usuarios []Usuario
}

// NuevoProyecto describe el alta de un proyecto con su equipo inicial.
type NuevoProyecto struct {
	Nombre      string
	FechaInicio time.Time
	FechaFin    *time.Time
	Usuarios []UsuarioInput
}

// ActualizacionProyecto describe la modificación de los datos de un proyecto.
// No incluye Usuarios: la composición se gestiona por separado (decisión D-06).
type ActualizacionProyecto struct {
	Nombre      string
	FechaInicio time.Time
	FechaFin    *time.Time
}
```

#### `internal/domain/usuario.go`

```go
package domain

import (
	"net/mail"
	"strings"
	"time"
)

// Usuario identifica a una persona. El email es su clave natural
// (decisión D-01) y es único en todo el sistema.
type Usuario struct {
	ID       uint
	Nombre   string
	Email    string
	CreadoEn time.Time

	// Relación inversa, usada por las métricas agregadas por persona.
	// No se expone en la API de US-01.
	Proyectos []Proyecto
}

// UsuarioInput es el dato de entrada de un usuario en las operaciones
// de alta de proyecto y de asociación.
type UsuarioInput struct {
	Nombre string
	Email  string
}

// NormalizarEmail canonicaliza el email para que las comparaciones y el
// índice único sean insensibles a mayúsculas y espacios.
func NormalizarEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// EmailValido valida la forma del email sin exigir dominios reales y
// rechaza la forma "Nombre <correo>" que net/mail aceptaría.
func EmailValido(email string) bool {
	normalizado := NormalizarEmail(email)
	if normalizado == "" || len(normalizado) > 150 || strings.ContainsAny(normalizado, " \t") {
		return false
	}
	direccion, err := mail.ParseAddress(normalizado)
	if err != nil {
		return false
	}
	return direccion.Address == normalizado
}
```

#### `internal/domain/proyecto_repository.go`

```go
package domain

import "context"

// ProyectoRepository es el puerto de persistencia de proyectos.
// Lo implementa internal/repository con GORM.
type ProyectoRepository interface {
	// CrearConUsuarios inserta el proyecto y todos sus vínculos en una
	// única transacción: o se persiste todo, o no se persiste nada.
	CrearConUsuarios(ctx context.Context, p *Proyecto, usuarios []Usuario) error
	ObtenerPorID(ctx context.Context, id uint) (*Proyecto, error)
	// Listar devuelve los proyectos no dados de baja, con usuarios precargados.
	Listar(ctx context.Context) ([]Proyecto, error)
	Actualizar(ctx context.Context, p *Proyecto) error
	// Eliminar realiza la baja lógica: informa la hora en borrado_en.
	Eliminar(ctx context.Context, id uint) error
	// TieneHistorial informa si el proyecto ya tiene entidades dependientes.
	TieneHistorial(ctx context.Context, id uint) (bool, error)
}

// UsuarioRepository es el puerto de persistencia de usuarios.
type UsuarioRepository interface {
	ObtenerPorEmail(ctx context.Context, email string) (*Usuario, error)
	Crear(ctx context.Context, i *Usuario) error
	EstaAsociado(ctx context.Context, proyectoID, usuarioID uint) (bool, error)
	// Vincular agrega el vínculo; si el vínculo ya existe devuelve
	// domain.ErrUsuarioYaAsociado (la PK compuesta es la red de seguridad).
	Vincular(ctx context.Context, proyectoID, usuarioID uint) error
	Desvincular(ctx context.Context, proyectoID, usuarioID uint) error
	ListarPorProyecto(ctx context.Context, proyectoID uint) ([]Usuario, error)
}
```

`ObtenerPorID`, `ObtenerPorEmail` y `ListarPorProyecto` devuelven `ErrProyectoNoEncontrado` o `ErrUsuarioNoEncontrado` cuando no hay coincidencias: el repository traduce internamente `gorm.ErrRecordNotFound`, de modo que las capas superiores nunca conocen GORM.

### 3. Persistencia de la relación proyecto–usuario

Se adopta una **tabla puente relacional `proyecto_usuarios`** (relación muchos a muchos declarada con `many2many` de GORM) en lugar de almacenar la lista de usuarios como columna `jsonb` dentro de `proyectos`.

Justificación:

1. **Integridad referencial y métricas.** En las historias de worklogs, defectos y Planning Poker el esfuerzo y los votos se atribuyen a una persona mediante clave foránea. Con `jsonb` no se puede declarar una FK, y las métricas por persona (velocidad individual, desviación de esfuerzo) obligarían a consultar con operadores `jsonb` y a deserializar en Go, con riesgo de inconsistencia entre consultas.
2. **Restricción de duplicados a nivel de motor.** La clave primaria compuesta `(proyecto_id, usuario_id)` impide físicamente que una persona figure dos veces en el mismo equipo, incluso ante peticiones concurrentes; la alternativa `jsonb` no puede expresar esa restricción.
3. **Mutaciones puntuales.** Agregar o quitar un usuario afecta una sola fila, en lugar de reescribir el array completo y perder escrituras concurrentes.
4. **Evolución del dominio.** Permite consultar "proyectos en los que participa la persona X" sin duplicar el dato y deja preparado el terreno para asignar historias a usuarios.

El campo `Usuarios` es la parte autoritativa de la relación; la tabla `proyectos` no almacena ninguna copia de la lista en `jsonb`, lo que garantiza una única fuente de verdad.

### 4. DDL resultante de `AutoMigrate`

El esquema se genera en el arranque de la aplicación mediante `db.AutoMigrate(&domain.Proyecto{}, &domain.Usuario{})` (que crea también la tabla puente de `many2many`). El DDL resultante en PostgreSQL es equivalente a:

```sql
CREATE TABLE IF NOT EXISTS "proyectos" (
    "id"             bigserial,
    "nombre"         varchar(150) NOT NULL,
    "fecha_inicio"   date         NOT NULL,
    "fecha_fin"      date,
    "estado"         varchar(20)  DEFAULT 'Activo' NOT NULL,
    "borrado_en"     timestamptz,
    "creado_en"      timestamptz  DEFAULT now() NOT NULL,
    "actualizado_en" timestamptz  NOT NULL,
    PRIMARY KEY ("id")
);
CREATE INDEX IF NOT EXISTS "idx_proyectos_nombre"     ON "proyectos" ("nombre");
CREATE INDEX IF NOT EXISTS "idx_proyectos_estado"     ON "proyectos" ("estado");
CREATE INDEX IF NOT EXISTS "idx_proyectos_borrado_en" ON "proyectos" ("borrado_en");

CREATE TABLE IF NOT EXISTS "usuarios" (
    "id"        bigserial,
    "nombre"    varchar(100) NOT NULL,
    "email"     varchar(150) NOT NULL,
    "creado_en" timestamptz  DEFAULT now() NOT NULL,
    PRIMARY KEY ("id")
);
CREATE UNIQUE INDEX IF NOT EXISTS "idx_usuarios_email" ON "usuarios" ("email");

CREATE TABLE IF NOT EXISTS "proyecto_usuarios" (
    "proyecto_id"   bigint NOT NULL,
    "usuario_id" bigint NOT NULL,
    PRIMARY KEY ("proyecto_id", "usuario_id"),
    CONSTRAINT "fk_proyecto_usuarios_proyecto"
        FOREIGN KEY ("proyecto_id") REFERENCES "proyectos" ("id")
        ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT "fk_proyecto_usuarios_usuario"
        FOREIGN KEY ("usuario_id") REFERENCES "usuarios" ("id")
        ON DELETE RESTRICT ON UPDATE CASCADE
);
```

Observaciones sobre el DDL:

- `borrado_en` se indexa porque **todas** las consultas de lectura filtran por `borrado_en IS NULL`. Un índice parcial (`WHERE borrado_en IS NULL`) sería más eficiente, pero no es declarable mediante tags GORM y no justifica complejidad adicional en US-01.
- `ON DELETE CASCADE` sobre `proyecto_id` solo surte efecto ante un borrado físico, que US-01 no expone; queda como garantía de integridad. `ON DELETE RESTRICT` sobre `usuarios` impide borrar una persona que todavía participa de un proyecto.
- El directorio `sql/` (montado en `docker-entrypoint-initdb.d`) permanece sin archivos `.sql` para US-01, ya que el esquema lo genera la aplicación.
- `AutoMigrate` es idempotente: ejecutarlo en cada arranque no destruye datos y no requiere scripts de migración manuales para esta historia.

### 5. Archivos a crear y modificar

| Acción | Ruta | Contenido |
| :--- | :--- | :--- |
| Crear | `internal/domain/proyecto.go` | `EstadoProyecto`, constantes, `Valido`, `Proyecto`, `NuevoProyecto`, `ActualizacionProyecto` |
| Crear | `internal/domain/usuario.go` | `Usuario`, `UsuarioInput`, `NormalizarEmail`, `EmailValido` |
| Crear | `internal/domain/errores.go` | Errores centinela del dominio (ver **Condiciones de error**) |
| Crear | `internal/domain/proyecto_repository.go` | Interfaces `ProyectoRepository` e `UsuarioRepository` |
| Crear | `internal/domain/proyecto_test.go` | Tests de las funciones puras: `EstadoProyecto.Valido`, `NormalizarEmail`, `EmailValido` |
| Crear | `internal/repository/migraciones.go` | `MigrarEsquema(db *gorm.DB) error` con el `AutoMigrate` de las entidades de US-01 |
| Crear | `internal/repository/migraciones_test.go` | Test de integración: `AutoMigrate` es idempotente y crea las tablas esperadas (omitido con `testing.Short()`) |
| Crear | `internal/repository/proyecto_repository.go` | Implementación GORM de `ProyectoRepository`, incluida la transacción de `CrearConUsuarios` y el filtro `borrado_en IS NULL` |
| Crear | `internal/repository/usuario_repository.go` | Implementación GORM de `UsuarioRepository` |
| Crear | `internal/service/proyecto_service.go` | `ProyectoService` con las ocho operaciones de negocio y la validación de las reglas de negocio |
| Crear | `internal/service/mocks_test.go` | Mocks de `ProyectoRepository` e `UsuarioRepository` con `testify/mock` |
| Crear | `internal/service/proyecto_service_test.go` | Tests table-driven de cada regla de negocio, con casos felices y de error |
| Crear | `internal/handler/dto_proyecto.go` | DTOs de request y response, con conversión de fechas ISO 8601 ↔ `time.Time` |
| Crear | `internal/handler/errores.go` | Traducción de errores de dominio a respuesta HTTP (`errors.Is` + `switch`) |
| Crear | `internal/handler/proyecto_handler.go` | `ProyectoHandler`, interfaz `ServicioProyectos` (definida por el consumidor, para poder testear con mock) y los ocho handlers |
| Crear | `internal/handler/proyecto_handler_test.go` | Tests del router con `httptest`: códigos de estado, cuerpos de error y parseo de path params |
| Crear | `cmd/api/router.go` | `registrarRutas(r *gin.Engine, srv handler.ServicioProyectos)` con el grupo `/api` |
| Crear | `cmd/api/router_test.go` | Test de que las ocho rutas de `/api` quedan registradas y no colisionan con `/health` |
| Crear | `docs/specs/US-01-gestion-de-proyectos-spec.md` | Este documento de especificación |
| Modificar | `cmd/api/main.go` | Invocar `MigrarEsquema` antes de arrancar el servidor, cablear `ProyectoRepository`, `UsuarioRepository`, `ProyectoService` y `ProyectoHandler`, y registrar el router |

No se modifica `internal/repository/db.go` (la conexión ya existe), ni `go.mod` ni `go.sum`, ni `.env` / `.env.example`, ni `docker-compose.yml`, ni `cmd/api/main_test.go` (que sigue probando `/health` de forma independiente).

### 6. Plan de pruebas

| Capa | Archivo de prueba | Técnica | Casos cubiertos |
| :--- | :--- | :--- | :--- |
| Domain | `internal/domain/proyecto_test.go` | Table-driven | `EstadoProyecto.Valido` con `"Activo"`, `"Cerrado"`, `"activo"`, `""`; `NormalizarEmail` con mayúsculas, espacios y cadena vacía; `EmailValido` con email válido, sin `@`, con dos `@`, con espacios, con la forma `Nombre <correo>` y con longitud 151 |
| Service | `internal/service/proyecto_service_test.go` | Table-driven con mocks `testify/mock` | Una fila por cada regla RB-01 a RB-12, más el orden de evaluación de errores ante entradas múltiples inválidas y la propagación de `ErrProyectoNoEncontrado` del repository |
| Repository | `internal/repository/migraciones_test.go` | Integración con PostgreSQL, `testing.Short()` la omite | `AutoMigrate` dos veces consecutivas no falla; existen las tablas `proyectos`, `usuarios` y `proyecto_usuarios`; el índice único de `usuarios.email` rechaza duplicados |
| Handler | `internal/handler/proyecto_handler_test.go` | `httptest` + mock del service | JSON malformado → `400`; `id` no numérico → `400`; fecha `28/09/2026` → `400`; proyecto inexistente → `404`; cada regla de negocio → su código; error inesperado del service → `500`; desvinculación de un usuario no asociado → `404` |
| Router | `cmd/api/router_test.go` | `httptest` | Las ocho rutas de `/api` resuelven (no devuelven el `404` propio de Gin) y `/health` sigue respondiendo `200` |

Mocks requeridos, todos con `testify/mock`, que ya figura en `go.mod`:

- `MockProyectoRepository` y `MockUsuarioRepository` en `internal/service/mocks_test.go`.
- `MockServicioProyectos` en `internal/handler/proyecto_handler_test.go`, que implementa la interfaz `handler.ServicioProyectos`.

Orden de implementación bajo TDD (RED → GREEN → REFACTOR), respetando el flujo de capas de `AGENTS.md`:

1. **RED:** `internal/domain/proyecto_test.go` (funciones puras, sin mocks).
2. **GREEN:** `internal/domain/proyecto.go`, `usuario.go`, `errores.go`, `proyecto_repository.go`.
3. **RED:** `internal/service/mocks_test.go` y `proyecto_service_test.go` con una fila por regla de negocio.
4. **GREEN:** `internal/service/proyecto_service.go`, con la mínima implementación que pasa los tests.
5. **REFACTOR:** `internal/repository/migraciones.go` y `migraciones_test.go`, `proyecto_repository.go`, `usuario_repository.go`.
6. **RED → GREEN:** `internal/handler/dto_proyecto.go`, `errores.go`, `proyecto_handler_test.go`, `proyecto_handler.go` y, por último, `cmd/api/router.go` junto con el cableado de `main.go`.

Cada commit sigue el formato de `docs/methodology.md` (`feat:`, `test:`, `docs:`) y deja evidencia del ciclo RED → GREEN en el historial del repositorio, según exige `docs/proyecto.md` §Trazabilidad.

## Reglas de negocio

Las reglas se evalúan **en la capa de servicio** y en el orden indicado, de modo que ante varias entradas inválidas el error devuelto sea determinista y el test sea inequívoco. Cada regla declara su criterio de validación y el mensaje exacto que se devuelve al cliente.

| ID | Regla | Criterio de validación | Mensaje exacto |
| :--- | :--- | :--- | :--- |
| RB-01 | El nombre del proyecto es obligatorio | `strings.TrimSpace(nombre) != ""` y su longitud está entre 1 y 150 caracteres después del recorte | `el nombre del proyecto es obligatorio` |
| RB-02 | La fecha de inicio del proyecto es obligatoria | `!fechaInicio.IsZero()`; el handler garantiza además el formato `YYYY-MM-DD` | `la fecha de inicio del proyecto es obligatoria` |
| RB-03 | Si se informa fecha de fin, no puede ser anterior a la fecha de inicio | `fechaFin == nil` o `!fechaFin.Before(fechaInicio)`. El caso `fechaFin == fechaInicio` es **válido** | `la fecha de fin no puede ser anterior a la fecha de inicio` |
| RB-04 | Todo proyecto nace `Activo` | El service fija `Estado = domain.EstadoActivo`; el cliente no puede enviar el estado al crear | — |
| RB-05 | El estado solo admite `Activo` y `Cerrado` | `EstadoProyecto.Valido()`, con comparación exacta del literal recibido | `el estado del proyecto debe ser "Activo" o "Cerrado"` |
| RB-06 | Cerrar un proyecto exige fecha de fin informada | Al pasar a `Cerrado`, `fechaFin != nil` | `no se puede cerrar un proyecto sin fecha de fin: informá la fecha de fin antes de cerrarlo` |
| RB-07 | Cada usuario debe tener nombre y email válidos | `TrimSpace(nombre) != ""` (entre 1 y 100 caracteres) y `domain.EmailValido(email)` | `el nombre del usuario es obligatorio` / `el email del usuario es obligatorio y debe tener un formato válido` |
| RB-08 | No se puede agregar dos veces al mismo usuario a un proyecto | Identidad por **email normalizado**: se rechaza si el email ya está asociado al proyecto o si se repite dentro del mismo request | `el usuario ya forma parte del proyecto` |
| RB-09 | Un email identifica a una sola persona en todo el sistema | `UsuarioRepository.ObtenerPorEmail`: si ya existe, se reutiliza el registro y se actualiza su nombre, en lugar de crear uno nuevo | — |
| RB-10 | Un proyecto cerrado no admite modificaciones de datos | `Actualizar`, `AgregarUsuario` y `QuitarUsuario` exigen `Estado == Activo`. La única operación admitida sobre un proyecto cerrado es el cambio de estado | `el proyecto está cerrado: solo se permite cambiar su estado o reabrirlo` |
| RB-11 | No se puede dar de baja un proyecto con historial asociado | `ProyectoRepository.TieneHistorial` se consulta antes de la baja; mientras no existan Sprints, historias ni worklogs devuelve `false` | `el proyecto tiene historial asociado y no puede eliminarse` |
| RB-12 | No se puede quitar un usuario que no forma parte del proyecto | `EstaAsociado(proyectoID, usuarioID)` debe devolver `true` | `el usuario no forma parte del proyecto` |

**Orden de evaluación en el alta** (`POST /api/proyectos`): RB-01 → RB-02 → RB-03 → RB-07 (por cada usuario, en el orden recibido) → RB-08 (normalización y detección de duplicados) → RB-04.
**Orden de evaluación en la modificación** (`PUT`): RB-01 → RB-02 → RB-03 → RB-10 (el estado del proyecto se verifica antes de escribir).
**Orden de evaluación en el cambio de estado** (`PATCH /estado`): RB-05 (validez del literal) → RB-06 (exigencia de fecha de fin) → RB-10 (no aplica: el cambio de estado es la excepción).

### Decisiones de diseño y su justificación

**D-01 · Identidad del usuario: email normalizado, único en todo el sistema.**
La clave natural de una persona es su email y no su nombre, porque los nombres se repiten, se acortan o cambian de formato entre proyectos. El email se normaliza a minúsculas y sin espacios antes de cualquier comparación o persistencia, y el índice único `idx_usuarios_email` garantiza una sola fila por persona en todo el sistema. Consecuencias: (a) agregar el mismo email desde dos proyectos reutiliza el registro existente y solo agrega el vínculo, evitando duplicar el historial de esfuerzo de una persona; (b) la desvinculación se expone por identificador numérico, de modo que el cliente no necesita conocer el email; (c) la duplicación dentro de un mismo proyecto (RB-08) es un caso particular de la unicidad global y por eso se responde con `409 Conflict` y no con `400`.

**D-02 · El esquema se gestiona con `AutoMigrate` de GORM, sin archivos `.sql`.**
El esquema se crea en el arranque de la aplicación a partir de los structs de dominio, en lugar de mantener scripts SQL versionados. Justificación: (a) el modelo y el esquema no pueden divergir, porque ambos provienen de la misma declaración; (b) `AutoMigrate` es idempotente, por lo que no requiere un gestor de versiones en una etapa del proyecto donde el esquema todavía cambia; (c) reduce el Swarm de configuración de la base para el desarrollo del MVP. El riesgo asumido es que `AutoMigrate` no resuelve cambios destructivos ni renombres de columna: cuando las historias de usuario stabilicen el modelo, esa necesidad se cubrirá con migraciones versionadas en una historia posterior.

**D-03 · `Cerrado` no es un estado terminal: la transición `Cerrado → Activo` está permitida.**
El enunciado no exige irreversibilidad y las métricas de `docs/proyecto.md` §7 se calculan sobre el histórico de Sprints e historias, no sobre proyectos cerrados, por lo que reabrir un proyecto no altera ninguna métrica ya calculada. En la práctica, cerrar un proyecto por error es un incidente frecuente y obligar a recrear el proyecto perdería todo su historial, lo que contradice el requisito de trazabilidad de la entrega. La reapertura no contradice RB-10 porque esa regla congela los **datos** del proyecto y no su ciclo de vida: `PATCH /api/proyectos/{id}/estado` es la única operación admitida sobre un proyecto cerrado. La transición al mismo estado en que ya se encuentra es una operación idempotente que responde `200` sin escribir en la base.

**D-04 · La baja de un proyecto es lógica (soft delete) mediante `borrado_en`.**
La baja marca `borrado_en` con la hora actual y el registro permanece en la tabla. Justificación: (a) las métricas de `docs/proyecto.md` §7 se calculan sobre el histórico de Sprints, historias y worklogs, y un borrado físico destruiría evidencia ya registrada, incumpliendo la cadena de trazabilidad exigida en los entregables; (b) permite auditar qué proyectos se dieron de baja y cuándo se hizo; (c) mantiene válida la clave foránea de las entidades futuras. El filtrado `borrado_en IS NULL` es responsabilidad del repository, de modo que el dominio expone el campo pero no conoce la técnica de baja. El borrado físico queda deliberadamente **fuera** de US-01: no se expone ningún endpoint de borrado duro, y cuando exista, RB-11 exigirá que el proyecto no tenga historial asociado.

**D-05 · La composición de usuarios se devuelve embebida en el proyecto; no hay endpoint propio de listado.**
El detalle (`GET /api/proyectos/{id}`) devuelve el array `usuarios` completo y el listado (`GET /api/proyectos`) devuelve solo `cantidad_usuarios`. Justificación: (a) evita el problema N+1, porque el repository precarga la relación con una única consulta adicional para toda la colección; (b) evita un endpoint redundante cuya respuesta ya está disponible dentro del recurso que la origina; (c) la composición del equipo es un atributo del proyecto, no una entidad consultable de forma independiente en US-01.

Dos precisiones que hacen la decisión verificable:

- **En el listado la clave `usuarios` no se emite, ni siquiera vacía.** El listado se construye con un tipo de respuesta propio (`respuestaProyectoListado`), no con el del detalle. Compartir un único tipo con un flag `conUsuarios` no alcanza: el flag decide si el array se puebla, pero no si la clave aparece, y resolverlo con `omitempty` rompería CL-06, que exige `"usuarios": []` en el detalle de un proyecto sin equipo.
- **`cantidad_usuarios` no depende de la clave `usuarios`.** Se calcula con `len(p.Usuarios)` sobre el objeto de dominio, así que el listado sigue informando el conteo aunque no serialice el array. La precarga de la relación en `Listar` existe precisamente para poder calcularlo.

**D-06 · `PUT` no modifica la composición del equipo.**
Modificar la lista de usuarios desde el `PUT` obligaría a definir una semántica de reemplazo total, con el riesgo de perder vínculos por omisión accidental de un campo en el body. La composición se modifica únicamente con `POST` y `DELETE` sobre `/api/proyectos/{id}/usuarios`, que son operaciones explícitas y auditables.

**D-07 · Un proyecto cerrado bloquea `PUT`, el alta de usuarios y la baja de usuarios, pero no el cambio de estado.**
La regla se define sobre el conjunto completo de operaciones mutantes para evitar un comportamiento inconsistente (permitir `PUT` pero no `POST` de usuarios, o viceversa). El congelamiento responde a que un proyecto cerrado se considera un snapshot finalizado para el cálculo de métricas; la única excepción es la reapertura (D-03). El mismo criterio se aplica a los usuarios: quitarlos de un proyecto cerrado alteraría los promedios de esfuerzo de la persona.

**D-08 · Cerrar un proyecto exige `fecha_fin` informada (RB-06).**
Sin fecha de finalización, un proyecto cerrado no puede apelarse a ninguna serie temporal: ni el dashboard, ni los reportes, ni las métricas por período podrían ubicar el cierre en el tiempo. La regla se valida en el service y por eso responde `422`; el cliente puede resolverla inmediatamente enviando un `PUT` con la `fecha_fin` antes de cerrar.

## Restricciones

1. **Clean Architecture obligatoria**: `handler → service → repository → domain`. Ninguna capa apunta hacia afuera.
2. **`internal/domain/` no importa nada externo**: solo biblioteca estándar (`context`, `time`, `net/mail`, `strings`, `errors`). Los tags de las structs son metadatos de texto y **no** constituyen una dependencia: el paquete sigue compilando sin GORM.
3. **Sin dependencias nuevas**: se usan únicamente las que ya figuran en `go.mod` (gin v1.9.1, gorm v1.31.2, gorm.io/driver/postgres v1.6.3, stretchr/testify v1.11.1, joho/godotenv v1.5.1) y la biblioteca estándar.
4. **El repository es la única capa que interactúa con la persistencia** y no contiene reglas de negocio: traduce `gorm.ErrRecordNotFound` a errores de dominio, aplica el filtro de baja lógica y ejecuta `AutoMigrate`, nada más.
5. **El service depende solo de las interfaces de `internal/domain/`** (inyección por constructor) y es la única capa que valida reglas de negocio.
6. **El handler solo valida la capa sintáctica** (JSON bien formado, tipos, `id` numérico, formato de fecha) y traduce errores; toda regla semántica vive en el service. Corolario de diseño: el `400` proviene del handler y el `422` del service.
7. **Atomicidad en la creación**: el proyecto y sus usuarios iniciales se persisten en una única transacción (`db.Transaction` dentro de `CrearConUsuarios`). La atomicidad es una preocupación de persistencia; las validaciones permanecen en el service.
8. **Errores con contexto**: todo error se envuelve con `fmt.Errorf("contexto: %w", err)`; los errores de dominio son centinela y se comparan con `errors.Is`.
9. **Fuera de alcance en US-01**: autenticación, autorización, paginación o filtrado del listado, exportación, historial de cambios de un proyecto, unicidad del nombre de proyecto (se deja libre a propósito) y cualquier endpoint que no figure en la tabla de la sección **Salidas esperadas**.
10. **Idioma y convención de nombres**: entidades, servicios, handlers y mensajes al usuario en español; identificadores JSON en `snake_case`; códigos de error en `UPPER_SNAKE_CASE`.
11. **Sin emojis** en el código ni en la documentación; los mensajes de log se escriben en español con `%v` sobre el error.
12. **TDD obligatorio**: ningún archivo de producción de esta US se entrega sin su test escrito previamente, en el commit o en el commit inmediatamente anterior, siguiendo RED → GREEN → REFACTOR.
13. **Nota operativa**: `docker-compose.yml` publica PostgreSQL en el puerto `5433` del host, mientras que la red interna de contenedores usa `5432`. Para ejecutar la API en el host contra el contenedor hay que usar `DB_PORT=5433`; dentro de la red de contenedores, `DB_PORT=5432`.

## Casos límite

| # | Caso | Comportamiento esperado |
| :--- | :--- | :--- |
| CL-01 | `nombre` formado solo por espacios o tabuladores (`"   "`) | `422` con `el nombre del proyecto es obligatorio`, porque RB-01 valida sobre el resultado de `TrimSpace` |
| CL-02 | `nombre` con exactamente 150 caracteres | Se acepta: es el máximo de la columna `varchar(150)` |
| CL-03 | `nombre` con 151 caracteres | `422` con `el nombre del proyecto no puede superar los 150 caracteres` |
| CL-04 | `fecha_fin` igual a `fecha_inicio` (proyecto de un solo día) | **Válido**: RB-03 compara con `Before`, por lo que la igualdad no se rechaza |
| CL-05 | `fecha_fin` ausente, `null` o cadena vacía | Se interpreta como proyecto sin fecha de finalización y se persiste en `NULL` |
| CL-06 | Proyecto creado sin el campo `usuarios` | Se persiste con cero usuarios; el detalle responde `"usuarios": []` y `"cantidad_usuarios": 0` |
| CL-07 | `usuarios: null` o `usuarios: []` | Comportamiento idéntico a CL-06; la respuesta nunca devuelve `null` en el array |
| CL-08 | El mismo email con distinta capitalización dentro de un mismo request (`Ada@utn.edu.ar` y `ada@utn.edu.ar`) | `409` con `el usuario ya forma parte del proyecto`, porque RB-08 compara emails normalizados |
| CL-09 | El mismo email ya asociado previamente al proyecto, enviado con distinta capitalización | `409` con `el usuario ya forma parte del proyecto` |
| CL-10 | Un email que ya existe en el sistema pero asociado a **otro** proyecto | Se reutiliza el registro existente, se actualiza su nombre y se crea el vínculo; responde `201` con el mismo `id` de usuario (RB-09) |
| CL-11 | Email con la forma `Nombre <correo@dominio>` | `422` con `el email del usuario es obligatorio y debe tener un formato válido`: el validador exige que la dirección parseada coincida exactamente con la entrada |
| CL-12 | Listado sin proyectos | `200` con `{"total": 0, "proyectos": []}` |
| CL-13 | `DELETE` sobre un proyecto ya dado de baja | `404` con `el proyecto no existe`: la baja lógica es invisible para el cliente, igual que un identificador inexistente (D-04) |
| CL-14 | Cambio de estado con distinta capitalización (`"activo"`, `"CERRADO"`) | `400` con `el estado del proyecto debe ser "Activo" o "Cerrado"` (RB-05) |
| CL-15 | Transición al mismo estado en que ya se encuentra (Activo → Activo) | `200` idempotente: no escribe en la base ni dispara el bloqueo de RB-10 |
| CL-16 | Agregar un usuario a un proyecto que ya fue dado de baja | `404` con `el proyecto no existe`, porque el filtrado de baja lógica se aplica también a las operaciones de composición |
| CL-17 | Dos peticiones concurrentes que agregan el mismo email al mismo proyecto | Una responde `201` y la otra `409`; la clave primaria compuesta de la tabla puente actúa como garantía final frente a la condición de carrera |
| CL-18 | Falla la persistencia de un usuario durante el alta del proyecto | La transacción revierte: no queda ni el proyecto ni los usuarios parciales, y la respuesta es `500` |

## Condiciones de error

### Errores centinela del dominio (`internal/domain/errores.go`)

```go
package domain

import "errors"

var (
	ErrProyectoNoEncontrado     = errors.New("el proyecto no existe")
	ErrNombreObligatorio        = errors.New("el nombre del proyecto es obligatorio")
	ErrNombreDemasiadoLargo     = errors.New("el nombre del proyecto no puede superar los 150 caracteres")
	ErrFechaInicioObligatoria   = errors.New("la fecha de inicio del proyecto es obligatoria")
	ErrFechaFinAnteriorAlInicio = errors.New("la fecha de fin no puede ser anterior a la fecha de inicio")
	ErrFechaFinRequerida        = errors.New("no se puede cerrar un proyecto sin fecha de fin: informá la fecha de fin antes de cerrarlo")
	ErrEstadoInvalido           = errors.New(`el estado del proyecto debe ser "Activo" o "Cerrado"`)
	ErrUsuarioNombreVacio    = errors.New("el nombre del usuario es obligatorio")
	ErrUsuarioEmailInvalido  = errors.New("el email del usuario es obligatorio y debe tener un formato válido")
	ErrUsuarioYaAsociado     = errors.New("el usuario ya forma parte del proyecto")
	ErrUsuarioNoAsociado     = errors.New("el usuario no forma parte del proyecto")
	ErrProyectoCerrado          = errors.New("el proyecto está cerrado: solo se permite cambiar su estado o reabrirlo")
	ErrProyectoConHistorial     = errors.New("el proyecto tiene historial asociado y no puede eliminarse")
	ErrErrorDePersistencia      = errors.New("error de persistencia")
)
```

`ErrProyectoNoEncontrado` e `ErrUsuarioNoAsociado` cubren también el caso de un recurso dado de baja lógicamente, porque el repository filtra por `borrado_en IS NULL` (CL-13 y CL-16).

### Mapeo de error de dominio a respuesta HTTP

| Error de dominio | Código HTTP | `error` en el JSON | `campo` | Operación afectada |
| :--- | :---: | :--- | :--- | :--- |
| Error de `ShouldBindJSON` (JSON malformado, tipo incorrecto, body vacío) | `400 Bad Request` | `JSON_INVALIDO` | — | Todas las que reciben body |
| `strconv.ParseUint` fallido o `id` no positivo | `400 Bad Request` | `PARAMETRO_INVALIDO` | — | Todas las rutas con `{id}` |
| `time.Parse` fallido (por ejemplo `28/09/2026`) | `400 Bad Request` | `PARAMETRO_INVALIDO` | `fecha_inicio`, `fecha_fin` o `fecha` según el caso | Todas las que reciben fechas |
| `ErrEstadoInvalido` | `400 Bad Request` | `ESTADO_INVALIDO` | `estado` | `PATCH /api/proyectos/{id}/estado` |
| `ErrProyectoNoEncontrado` | `404 Not Found` | `PROYECTO_NO_ENCONTRADO` | — | `GET`, `PUT`, `PATCH`, `DELETE` y composición de usuarios |
| `ErrUsuarioNoAsociado` | `404 Not Found` | `INTEGRANTE_NO_ENCONTRADO` | — | `DELETE /api/proyectos/{id}/usuarios/{usuarioId}` |
| `ErrUsuarioYaAsociado` | `409 Conflict` | `USUARIO_DUPLICADO` | `email` | `POST` de usuarios y alta con usuarios |
| `ErrProyectoCerrado` | `409 Conflict` | `PROYECTO_CERRADO` | — | `PUT`, alta y baja de usuarios |
| `ErrProyectoConHistorial` | `409 Conflict` | `PROYECTO_CON_HISTORIAL` | — | `DELETE /api/proyectos/{id}` |
| `ErrNombreObligatorio` | `422 Unprocessable Entity` | `VALIDACION` | `nombre` | `POST`, `PUT` |
| `ErrNombreDemasiadoLargo` | `422 Unprocessable Entity` | `VALIDACION` | `nombre` | `POST`, `PUT` |
| `ErrFechaInicioObligatoria` | `422 Unprocessable Entity` | `VALIDACION` | `fecha_inicio` | `POST`, `PUT` |
| `ErrFechaFinAnteriorAlInicio` | `422 Unprocessable Entity` | `VALIDACION` | `fecha_fin` | `POST`, `PUT` |
| `ErrFechaFinRequerida` | `422 Unprocessable Entity` | `VALIDACION` | `fecha_fin` | `PATCH /api/proyectos/{id}/estado` |
| `ErrUsuarioNombreVacio` | `422 Unprocessable Entity` | `VALIDACION` | `usuarios[].nombre` en el alta del proyecto; `nombre` en el alta de un usuario suelto | `POST` de proyecto y de usuarios |
| `ErrUsuarioEmailInvalido` | `422 Unprocessable Entity` | `VALIDACION` | `usuarios[].email` en el alta del proyecto; `email` en el alta de un usuario suelto | `POST` de proyecto y de usuarios |
| Cualquier otro error (violación de índice único, fallo de conexión, panic recovery) | `500 Internal Server Error` | `ERROR_INTERNO` | — | Todas |

El handler implementa el mapeo con un `switch` sobre `errors.Is`; todo error no reconocido se registra con `log.Printf` y se responde con un mensaje genérico que no filtra detalles de infraestructura.

**Regla de `campo` para los errores de usuario.** El campo `campo` replica la ruta JSON que el campo ocupa en el body que el cliente envió, para que sepa dónde corregir. Como el body cambia según el endpoint, el mismo error de negocio se informa de dos maneras:

| Endpoint | Forma del body | `campo` para nombre y email |
| :--- | :--- | :--- |
| `POST /api/proyectos` | `{"nombre":…,"usuarios":[{"nombre":…,"email":…}]}` | `usuarios[].nombre` y `usuarios[].email` |
| `POST /api/proyectos/{id}/usuarios` | `{"nombre":…,"email":…}` | `nombre` y `email` |

Por eso el mapeador no puede ser una función del error solamente: recibe además un contexto que indica si los usuarios venían anidados (`contextoCampo` en `internal/handler/errores.go`). `ErrUsuarioYaAsociado` es la excepción y siempre informa `email`, porque el conflicto se describe por el email duplicado y no por su ubicación en el body (E-17 y E-44).

### Ejemplos de respuesta de error

`422` por regla de negocio:

```json
{
  "error": "VALIDACION",
  "mensaje": "la fecha de fin no puede ser anterior a la fecha de inicio",
  "campo": "fecha_fin"
}
```

`409` por conflicto de estado:

```json
{
  "error": "PROYECTO_CERRADO",
  "mensaje": "el proyecto está cerrado: solo se permite cambiar su estado o reabrirlo"
}
```

`500` por fallo de persistencia (mensaje genérico; el detalle queda en el log del servidor):

```json
{
  "error": "ERROR_INTERNO",
  "mensaje": "ocurrió un error inesperado al procesar la solicitud"
}
```

### Fallas de infraestructura

| Situación | Manejo |
| :--- | :--- |
| PostgreSQL no está disponible al arrancar | `MigrarEsquema` devuelve error y `main.go` termina con `log.Fatalf` antes de levantar el servidor, en lugar de exponer una API que fallará en cada request |
| Violación de unicidad de `usuarios.email` por una carrera entre dos altas | Se traduce a `409 USUARIO_DUPLICADO`; el índice único es la garantía final de RB-09 |
| Falla la transacción de `CrearConUsuarios` | `db.Transaction` revierte automáticamente; el service envuelve el error con `fmt.Errorf("creando proyecto: %w", err)` y el handler responde `500` |
| Panic dentro de un handler | `gin.Recovery()` (incluido en `gin.Default()`) devuelve `500` sin cerrar el proceso |
| Base de datos inaccesible durante una request | Error de driver no reconocido por el `switch`: se responde `500 ERROR_INTERNO` y se registra el error real en el log |

## Criterios de aceptación

Cada criterio declara su veredicto y la evidencia que lo respalda. `CUMPLE` significa que
existe un test que falla si el comportamiento se rompe. `CUMPLE PARCIAL` significa que el
comportamiento está implementado y probado, pero falta cubrir alguna arista de la frase.
`NO CUMPLE` se reserva para desviaciones de proceso que no se pueden corregir de forma
retroactiva, y se explican en lugar de marcarse como verde.

### Funcionales

- [x] `POST /api/proyectos` crea un proyecto con nombre, `fecha_inicio` y `fecha_fin` opcional, y responde `201` con el proyecto persistido. — CUMPLE. `TestCrearProyecto/E-01`, `handler.TestCrearProyecto`, `TestLaAltaDevuelveLosDatosPersistidos`.
- [x] El `id` del proyecto es autogenerado por el sistema y es único; el cliente no puede enviarlo. — CUMPLE. `TestElIdLoAsignaElServidorYNoElCliente` envía `{"id":999}` y comprueba que la respuesta trae el id del servidor y que el DTO de dominio no tiene campo `ID`.
- [x] Un proyecto creado sin `fecha_fin` queda con la fecha en `NULL` y puede consultarse sin error. — CUMPLE. `TestCrearProyecto/E-02`; la lectura sin `fecha_fin` se cubre en `TestDetalleEmiteArrayUsuariosVacio` y `TestLasFechasSeDevuelvenEnISO8601`.
- [x] Todo proyecto se crea en estado `Activo` sin importar el body recibido. — CUMPLE. `TestCrearProyecto` afirma `EstadoActivo` en todos sus casos (`RB-04`), incluido el que no manda estado.
- [x] `GET /api/proyectos` lista los proyectos no dados de baja y devuelve `{"total": n, "proyectos": [...]}`. — CUMPLE. `TestListarProyectos`, `handler.TestListarProyectos`, `TestBajaLogicaInvisibleParaElCliente` (el filtro de baja lógica se verifica contra la base real) y `TestListadoOmiteArrayUsuariosPorD05`.
- [x] `GET /api/proyectos/{id}` devuelve el detalle con el array `usuarios` y `cantidad_usuarios` coherente. — CUMPLE. `TestObtenerProyectoPorID/E-05`, `TestCantidadUsuariosCoincideConElDetalle` (tabla de 0, 1 y 3 usuarios, incluido el orden) y `TestDetalleEmiteArrayUsuariosVacio`.
- [x] `PUT /api/proyectos/{id}` modifica nombre y fechas y responde `200` con el recurso actualizado. — CUMPLE. `TestActualizarProyecto/E-10`, `handler.TestActualizarProyecto`.
- [x] `PATCH /api/proyectos/{id}/estado` cambia el estado entre `Activo` y `Cerrado`, en ambos sentidos, y responde `200`. — CUMPLE. `TestCambiarEstadoProyecto/E-33` (cierre), `/E-34` (reapertura) y `/E-35` (idempotencia); `handler.TestCambiarEstadoProyecto`.
- [x] `DELETE /api/proyectos/{id}` realiza la baja lógica: el proyecto deja de aparecer en el listado y en el detalle, y responde `204`. — CUMPLE. `TestBajaLogicaInvisibleParaElCliente` comprueba contra la base real que desaparece de `Listar` y de `ObtenerPorID`; `TestEliminarProyecto/E-51` y `handler.TestEliminarProyecto` cubren el `204`.
- [x] `POST /api/proyectos/{id}/usuarios` agrega un usuario y responde `201`. — CUMPLE. `TestAgregarUsuario/E-40`, `handler.TestAgregarUsuario`, `TestAltaUsuarioSiInformaProyectoId`.
- [x] `DELETE /api/proyectos/{id}/usuarios/{usuarioId}` quita un usuario del proyecto y responde `204`. — CUMPLE. `TestQuitarUsuario/E-46`, `handler.TestQuitarUsuario`.
- [x] Las fechas se aceptan y se devuelven en formato ISO 8601 `YYYY-MM-DD`. — CUMPLE. `TestLasFechasSeDevuelvenEnISO8601` parsea `fecha_inicio` y `fecha_fin` con `time.DateOnly` y `creado_en` con `time.RFC3339`. La aceptación del formato de entrada se cubre en `TestCrearProyecto` y en el binding de los handlers.

### Reglas de negocio

- [x] Un nombre vacío o compuesto solo por espacios se rechaza con `422` y el mensaje `el nombre del proyecto es obligatorio`. — CUMPLE. `TestCrearProyecto/RB-01` y `/E-09`; el mensaje exacto se contrasta contra `domain.ErrNombreObligatorio` en `TestCampoDelErrorDeUsuarioSegunLaRuta`.
- [x] Una fecha de inicio ausente o vacía se rechaza con `422` y el mensaje `la fecha de inicio del proyecto es obligatoria`. — CUMPLE. `TestCrearProyecto/RB-02` y `/E-19` (el nombre se valida antes que la fecha).
- [x] Una `fecha_fin` anterior a la `fecha_inicio` se rechaza con `422`; una `fecha_fin` igual a la `fecha_inicio` se acepta. — CUMPLE. `TestCrearProyecto/E-11` (rechazo) y `/E-04` (aceptación); en la modificación, `TestActualizarProyecto/RB-03`.
- [x] Un proyecto puede crearse sin usuarios. — CUMPLE. `TestCrearProyecto/E-03` y `handler.TestListarProyectos` sobre la lista vacía.
- [x] No se puede agregar al mismo usuario (mismo email normalizado) dos veces al mismo proyecto: se responde `409`. — CUMPLE. `TestAgregarUsuario/E-44` (alta suelta, `409`), `TestCrearProyectoRechazaEmailsDuplicadosEnElMismoAlta` (tres variantes: idéntico, distinta capitalización, espacios sobrantes) y `TestLaAltaConUsuariosRevierteSiFallaUnUsuario` (el índice único de la base es la garantía final).
- [x] Cada usuario debe tener nombre no vacío y email con formato válido. — CUMPLE. `TestCrearProyecto/RB-07` y `/E-15`, `TestAgregarUsuario/RB-07 E-42` y `/RB-07 E-43`, más el límite de 100 caracteres del nombre.
- [x] No se puede modificar un proyecto cerrado: `PUT` responde `409`; el cambio de estado sigue disponible. — CUMPLE. `TestActualizarProyecto/RB-10 E-32` cubre el `409`; la segunda mitad la cubre `TestCambiarEstadoProyecto/E-34`, que reabre un proyecto cerrado.
- [x] Cerrar un proyecto sin `fecha_fin` informada se rechaza con `422`. — CUMPLE. `TestCambiarEstadoProyecto/E-37`.
- [x] No se puede quitar un usuario que no forma parte del proyecto: se responde `404`. — CUMPLE. `TestQuitarUsuario/E-47`.

### Arquitectura y calidad

- [x] `internal/domain/` compila sin GORM ni ninguna dependencia externa; los errores de dominio son centinela comparables con `errors.Is`. — CUMPLE. Verificado por inspección de imports: los cuatro archivos de producción (`errores.go`, `usuario.go`, `proyecto.go`, `proyecto_repository.go`) no importan ningún módulo externo. Los 16 centinelas se construyen con `errors.New` en `internal/domain/errores.go` y se comparan con `assert.ErrorIs` en toda la suite.
- [x] El service depende solo de las interfaces definidas en `internal/domain/` y se testea con mocks de `testify/mock`. — CUMPLE. `internal/service` importa `internal/domain` y nada más de la capa interna; `internal/service/proyecto_service_test.go` construye los dobles con `testify/mock`.
- [x] El repository no contiene reglas de negocio y traduce los errores de GORM a errores de dominio. — CUMPLE. Los repositorios solo mapean `gorm.ErrRecordNotFound` a `ErrProyectoNoEncontrado` / `ErrUsuarioNoEncontrado` y `gorm.ErrDuplicatedKey` a `ErrUsuarioYaAsociado`; las reglas viven en el service. `TestLaAltaConUsuariosRevierteSiFallaUnUsuario` verifica la traducción en el camino real.
- [x] El handler valida únicamente la capa sintáctica y devuelve el cuerpo de error estandarizado. — CUMPLE. `TestCampoDelErrorDeUsuarioSegunLaRuta` recorre los seis errores de validación y su `campo`, y `TestContextoCanceladoSePropaga` comprueba que el handler no inventa reglas de negocio.
- [x] El esquema se crea con `AutoMigrate` en el arranque, es idempotente y no requiere archivos `.sql`. — CUMPLE. `TestMigrarEsquemaEsIdempotente` ejecuta `MigrarEsquema` dos veces contra la base real. No hay archivos `.sql` en el repositorio.
- [x] La creación de un proyecto con usuarios es atómica: ante un fallo no queda persistencia parcial. — CUMPLE. `TestLaAltaConUsuariosRevierteSiFallaUnUsuario` provoca un email duplicado y comprueba que no sobreviven ni el proyecto, ni el primer usuario ya insertado. `TestCrearConUsuariosEsAtomico` cubre la construcción del equipo sin duplicados.
- [x] Todos los errores se manejan explícitamente y se envuelven con `fmt.Errorf("contexto: %w", err)`. — CUMPLE. `go vet ./...` no reporta nada y los tests de propagación (`TestCrearProyectoPropagaErrorDePersistencia`, y los casos "propaga el error a…" del service) fijan el comportamiento esperado.
- [x] `go build ./...`, `go vet ./...` y `go test ./...` finalizan sin errores. — CUMPLE. Ejecutados con PostgreSQL real (`localhost:5434`) el resultado fue `ok` en los cinco paquetes: `cmd/api`, `internal/domain`, `internal/handler`, `internal/repository`, `internal/service`.
- [ ] Los tests se escribieron antes que la implementación, con evidencia en el historial de commits. — **NO CUMPLE.** Tests e implementación entraron juntos en el commit `c9e8b60`, así que el historial no demuestra la secuencia RED → GREEN. No se puede reconstruir sin reescribir la historia, que no es una opción: se registra la desviación. Mitigación: las correcciones de esta revisión se hicieron al revés (test primero, y el test de atomicidad fallenó en rojo y reveló el defecto de normalización antes de corregirlo).
- [x] La especificación, los escenarios BDD y los tests permanecen versionados en el repositorio, según la cadena de trazabilidad de `docs/proyecto.md`. — CUMPLE. `docs/specs/US-01-gestion-de-proyectos-spec.md`, `features/US-01-gestion-de-proyectos.feature` y `docs/proyecto.md` están trackeados por git.

## Escenarios BDD

### Caso normal

Given no existen proyectos en el sistema
When creo un proyecto con nombre "Software Metrics & Estimation", `fecha_inicio` "2026-09-28", `fecha_fin` "2026-12-18" y dos usuarios
Then la API responde `201 Created` con el proyecto persistido
And el `id` es autogenerado por el sistema
And el estado del proyecto es `Activo`
And `cantidad_usuarios` es `2` y el array `usuarios` contiene a ambas personas
And el proyecto aparece en `GET /api/proyectos`
And el detalle de `GET /api/proyectos/{id}` devuelve las mismas fechas en formato `YYYY-MM-DD`
And modifico el proyecto con `PUT` y la respuesta es `200 OK` con los datos actualizados
And doy de baja el proyecto con `DELETE` y la respuesta es `204 No Content`
And el proyecto deja de aparecer en el listado y su detalle responde `404`

### Caso alternativo

Given un proyecto activo con `fecha_fin` "2026-12-18" y dos usuarios
When envío `PATCH /api/proyectos/{id}/estado` con `{"estado": "Cerrado"}`
Then la API responde `200 OK` con el proyecto en estado `Cerrado`
And envío `PATCH /api/proyectos/{id}/estado` con `{"estado": "Activo"}`
And la API responde `200 OK` con el proyecto en estado `Activo`, porque el cierre es reversible
And envío `POST /api/proyectos/{id}/usuarios` con un email que ya existe en otro proyecto
And la API responde `201 Created` reutilizando el identificador del usuario existente
And envío `DELETE /api/proyectos/{id}/usuarios/{usuarioId}`
And la API responde `204 No Content` y `cantidad_usuarios` disminuye en uno

### Caso límite

Given un proyecto que comienza y termina el mismo día
When creo el proyecto con `fecha_inicio` "2026-11-02" y `fecha_fin` "2026-11-02"
Then la API responde `201 Created`, porque una fecha de fin igual a la de inicio es válida
And cuando creo el proyecto enviando `"usuarios": []`
And la API responde `201 Created` con `"usuarios": []` y `"cantidad_usuarios": 0`
And cuando consulto el listado sin ningún proyecto visible
And la API responde `200 OK` con `{"total": 0, "proyectos": []}` y no con `null`
And cuando envío dos usuarios con el mismo email en distinta capitalización dentro del mismo request
And la API responde `409 Conflict` con el mensaje `el usuario ya forma parte del proyecto`

### Caso de error

Given un proyecto activo con `fecha_inicio` "2026-09-28"
When intento cerrarlo sin informar `fecha_fin`
Then la API responde `422 Unprocessable Entity` con el mensaje `no se puede cerrar un proyecto sin fecha de fin: informá la fecha de fin antes de cerrarlo`
And cuando intento crear un proyecto con `fecha_fin` "2026-09-01"
And la API responde `422 Unprocessable Entity` con el mensaje `la fecha de fin no puede ser anterior a la fecha de inicio`
And cuando creo un proyecto cuyo `nombre` es solo espacios
And la API responde `422 Unprocessable Entity` con el mensaje `el nombre del proyecto es obligatorio`
And cuando envío el body con una fecha en formato `28/09/2026`
And la API responde `400 Bad Request` con el código `PARAMETRO_INVALIDO`
And cuando consulto un proyecto que fue dado de baja lógicamente
And la API responde `404 Not Found` con el mensaje `el proyecto no existe`
And cuando intento modificar un proyecto ya cerrado con `PUT`
And la API responde `409 Conflict` con el código `PROYECTO_CERRADO`
And cuando la persistencia falla durante el alta de un proyecto con usuarios
And la API responde `500 Internal Server Error` sin dejar registros parciales en la base de datos

## Notas de implementación

### Los tests de integración se omitían en silencio (corregido)

`internal/repository/migraciones_test.go` exige que la variable `DB_HOST` esté
definida; si no lo está, cada test se salta con `t.Skip` y el paquete se reporta
como `ok`. En un entorno sin PostgreSQL —incluido el pipeline— la suite pasaba
sin ejecutar una sola prueba de persistencia, lo que daba una falsa sensación de
cobertura.

**Resuelto en `be8cbbd`** con dos medidas en `.github/workflows/go.yml`:

1. Un servicio `postgres:15.4-bullseye` —la misma imagen que usa
   `docker-compose.yml`— con las variables `DB_*` exportadas al job, de modo que
   la integración se ejecuta de verdad y un fallo de persistencia rompe el
   pipeline.
2. Un paso `Verify que la integracion se ejecuto` que corre el paquete entero con
   `-v` y falla si aparece cualquier `--- SKIP`. Se revisa el paquete completo en
   lugar de enumerar tests, para que el control no se quede viejo al agregar
   casos nuevos. Sin este paso, un corte futuro de la configuración del servicio
   volvería a dejar el pipeline en verde sin probar la base.

Para replicar la validación en local:

```bash
docker compose up -d postgresql
export DB_HOST=localhost DB_PORT=5433 DB_USER=postgres
export DB_PASSWORD=postgres DB_NAME=metrics_db DB_SSLMODE=disable
go test ./internal/repository/... -v
```

### El volumen de desarrollo local estaba corrupto (causa encontrada y corregida)

Durante la validación del 2026-09-28, el volumen `data/pgdata` del
`docker-compose.yml` entró en un ciclo de recuperación sin fin
(`FATAL: the database system is in recovery mode`, SQLSTATE `57P03`) y Postgres
rechazaba todas las conexiones. El código de la US-01 no intervenía.

**Causa raíz**: `docker-compose.yml` montaba `./data` —el directorio de datos
completo— como *bind mount* del proyecto. PostgreSQL no soporta su directorio de
datos sobre el sistema de archivos de Windows. El síntoma era que los procesos
backend morían con `server process (PID N) exited with exit code 2` en el
momento exacto en que una consulta tocaba el disco, y el cluster entraba a
recuperar WAL indefinidamente.

El `healthcheck` del compose reporta `healthy` durante la recuperación, porque
`pg_isready` también responde afirmativo mientras la base está recovering. Por
eso el contenedor figuraba sano mientras la base rechazaba conexiones: **el
healthcheck no sirve para detectar este problema**.

**Corrección**: `PGDATA` pasa a un volumen nombrado de Docker
(`pgdata:/var/lib/postgresql/data`) y `./sql` se mantiene como bind mount de
solo lectura, porque los scripts de inicialización solo se leen al crear el
cluster. Con el volumen nombrado el arranque de `initdb` pasó de más de diez
minutos a veinte segundos, y la suite completa de integración pasa sin que el
servidor caiga.

Los datos que había en `./data` se respaldaron antes de vaciarlo; no eran
recuperables de todos modos, porque el cluster no llegaba a consolidar.

Para levantar el entorno:

```bash
docker compose up -d postgresql
# desde Windows, el puerto publicado es 5433 y las credenciales salen del .env:
export DB_HOST=localhost DB_PORT=5433
export DB_USER=<POSTGRES_USER> DB_PASSWORD=<POSTGRES_PASSWORD>
export DB_NAME=<POSTGRES_DB> DB_SSLMODE=disable
go test ./internal/repository/... -v
```

Ojo con `DB_PORT`: en el `.env` figura `5432`, que es el puerto **dentro** de la
red de Docker. Desde el host hay que usar el puerto publicado, `5433`.


### `PUT` es un reemplazo completo, no una actualización parcial

`PUT /api/proyectos/:id` exige `nombre` y `fecha_inicio`. Omitir cualquiera de
los dos responde `422`, porque la semántica REST de `PUT` es sustituir el
recurso. Las actualizaciones parciales de negocio no están expuestas en esta
historia de usuario.
