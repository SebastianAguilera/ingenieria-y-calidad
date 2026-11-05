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

| Operación | Campos requeridos y formato |
| :--- | :--- |
| `POST /api/proyectos` | `nombre` (1–150 tras `TrimSpace`), `fecha_inicio` (`YYYY-MM-DD`); opcionales: `fecha_fin` (`YYYY-MM-DD` o `null`, no anterior a inicio) y `usuarios` (array; cada elemento con `nombre` 1–100 y `email` válido ≤150) |
| `PUT /api/proyectos/{id}` | `nombre`, `fecha_inicio`; opcional `fecha_fin`. No incluye `usuarios` (D-06) |
| `PATCH /api/proyectos/{id}/estado` | `estado`: `"Activo"` o `"Cerrado"` (distingue mayúsculas) |
| `POST /api/proyectos/{id}/usuarios` | `nombre` (1–100), `email` (válido ≤150) |
| `DELETE /api/proyectos/{id}/usuarios/{usuarioId}` | Sin body; `usuarioId` es el id devuelto por el alta o el detalle |

Ejemplo de alta completa:

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

El body de `PUT` no incluye `usuarios`: la composición se modifica con `POST`/`DELETE` sobre `/usuarios` (D-06).

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

Recurso proyecto (respuesta de `POST`, `GET /{id}`, `PUT` y `PATCH /estado`):

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

- `GET /api/proyectos`: `{"total": n, "proyectos": [...]}` sin `usuarios` por proyecto; con cero proyectos, `{"total": 0, "proyectos": []}`.
- `POST /api/proyectos/{id}/usuarios`: `201` con la asociación `{ "id", "nombre", "email", "proyecto_id" }`; el `id` puede ser de un usuario preexistente (D-01).
- `DELETE`: `204` sin cuerpo.
- Los usuarios pueden corresponder a registros preexistentes si el email ya existía (D-01).

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

| Tipo | Campos / operaciones | Reglas |
| :--- | :--- | :--- |
| `Proyecto` | `ID`, `Nombre`, `FechaInicio`, `FechaFin *time.Time`, `Estado`, `BorradoEn *time.Time`, `CreadoEn`, `ActualizadoEn`, `Usuarios []Usuario` | `Estado` solo `Activo`/`Cerrado`; `FechaFin` opcional; `BorradoEn` implementa baja lógica. |
| `NuevoProyecto` / `ActualizacionProyecto` | `Nombre`, `FechaInicio`, `FechaFin *time.Time` (+ `Usuarios []UsuarioInput` solo en el alta) | La actualización no incluye usuarios (D-06). |
| `Usuario` / `UsuarioInput` | `ID`, `Nombre`, `Email`, `CreadoEn` (+ `Proyectos []Proyecto` inverso) | El email es la clave natural; se normaliza con `NormalizarEmail` y se valida con `EmailValido`. |
| `ProyectoRepository` | `CrearConUsuarios`, `ObtenerPorID`, `Listar`, `Actualizar`, `Eliminar`, `TieneHistorial` | `CrearConUsuarios` es atómico; `Eliminar` es baja lógica; traduce `gorm.ErrRecordNotFound` a errores de dominio. |
| `UsuarioRepository` | `ObtenerPorEmail`, `Crear`, `EstaAsociado`, `Vincular`, `Desvincular`, `ListarPorProyecto` | `Vincular` devuelve `ErrUsuarioYaAsociado` si el vínculo existe. |

`EstadoProyecto.Valido()` acepta solo `Activo` y `Cerrado`; `NormalizarEmail` aplica `strings.ToLower(strings.TrimSpace(...))`; `EmailValido` rechaza vacíos, más de 150 caracteres, espacios y la forma `Nombre <correo>`. Las firmas completas y los tags están en `internal/domain/proyecto.go`, `usuario.go` y `proyecto_repository.go`.

### 3. Plan de pruebas

| Capa | Archivo | Qué verifica |
| :--- | :--- | :--- |
| Dominio | `internal/domain/proyecto_test.go` | Validaciones puras de estado, email y normalización. |
| Servicio | `internal/service/proyecto_service_test.go` | Una fila por RB-01..RB-12, orden de evaluación y propagación de errores. |
| Repositorio | `internal/repository/migraciones_test.go` | `AutoMigrate` idempotente, tablas e índice único de email (integración). |
| Handler | `internal/handler/proyecto_handler_test.go` | `400/404/409/422/500`, mapeo de `campo` y JSON inválido. |
| Router | `cmd/api/router_test.go` | Las ocho rutas y `/health`. |

Mocks: `MockProyectoRepository`/`MockUsuarioRepository` (servicio) y `MockServicioProyectos` (handler), con `testify/mock`. Cada capa se implementa en orden `domain → service → repository → handler` bajo RED → GREEN → REFACTOR, con commits `feat:`, `test:` y `docs:` según `docs/methodology.md`.

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

## Decisiones de diseño

| ID | Decisión |
| :--- | :--- |
| D-01 | El email normalizado identifica al usuario en todo el sistema; si ya existe, se reutiliza el registro y solo se agrega el vínculo. |
| D-02 | El esquema se crea con `AutoMigrate` al arrancar; no hay scripts SQL versionados. |
| D-03 | `Cerrado` no es terminal: la reapertura (`Cerrado → Activo`) está permitida y es idempotente. |
| D-04 | La baja de proyecto es lógica (`borrado_en`); el cliente la ve como `404`. |
| D-05 | La composición de usuarios se devuelve embebida en el detalle; el listado solo informa `cantidad_usuarios` y no existe endpoint de listado de usuarios. |
| D-06 | El `PUT` no modifica la composición del equipo; se usan `POST`/`DELETE` sobre `/usuarios`. |
| D-07 | Un proyecto cerrado bloquea `PUT` y las altas/bajas de usuarios; solo admite el cambio de estado. |
| D-08 | Cerrar exige `fecha_fin`; sin ella se responde `422`. |

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

El cuerpo de error sigue el formato estándar definido en **Salidas esperadas**.

## Criterios de aceptación

Todos los criterios están cubiertos por tests automatizados; la evidencia detallada vive en los propios tests. Veredicto global: **CUMPLE**, salvo la desviación de proceso registrada al final.

| Grupo | Criterios | Evidencia principal |
| :--- | :--- | :--- |
| Funcionales | Alta, consulta, listado, modificación, cambio de estado, baja lógica y composición de usuarios (8 endpoints) | `TestCrearProyecto`, `TestListarProyectos`, `TestObtenerProyectoPorID`, `TestActualizarProyecto`, `TestCambiarEstadoProyecto`, `TestEliminarProyecto`, `TestAgregarUsuario`, `TestQuitarUsuario` |
| Reglas de negocio | RB-01..RB-12 y sus mensajes exactos | `TestCrearProyecto`, `TestActualizarProyecto`, `TestCambiarEstadoProyecto`, `TestAgregarUsuario`, `TestQuitarUsuario` |
| Arquitectura | Capas, dominio sin dependencias externas, `AutoMigrate` idempotente, atomicidad y errores envueltos | `TestMigrarEsquemaEsIdempotente`, `TestLaAltaConUsuariosRevierteSiFallaUnUsuario`, `go vet ./...` |
| Trazabilidad BDD | E-01, E-02, E-03, E-04, E-05, E-09, E-10, E-11, E-15, E-17, E-19, E-32, E-33, E-34, E-35, E-37, E-40, E-42, E-43, E-44, E-46, E-47, E-51 | Subtests Go y `docs/features/US-01-gestion-de-proyectos.feature` |

- [ ] **Desviación de proceso:** los tests de US-01 se commitearon junto con la implementación (`c9e8b60`), por lo que el historial no demuestra RED → GREEN. Se registra como `NO CUMPLE`; las correcciones posteriores sí se hicieron test-first.

## Escenarios BDD

Los escenarios Gherkin completos (normal, alternativo, límite y error) están en
[`docs/features/US-01-gestion-de-proyectos.feature`](../features/US-01-gestion-de-proyectos.feature).
Cada escenario se automatiza como subtest Go con su ID `E-nn`; la lista de IDs está en **Criterios de aceptación**.
