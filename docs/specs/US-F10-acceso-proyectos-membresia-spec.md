# Especificación: US-F10 - Acceso a proyectos por membresía

## Objetivo

Que cada usuario **solo vea y opere los proyectos del que es miembro**. Hoy
cualquier usuario autenticado lista, abre, edita y borra todos los proyectos, e
incluso puede unirse a cualquiera con `POST /api/proyectos/{id}/usuarios`; con
esta historia el alta de un proyecto **vincula al creador como miembro** en la
misma transacción y toda ruta anidada bajo `/api/proyectos/{id}/...` exige la
membresía antes de llegar al handler.

La historia se resuelve con tres mecanismos:

1. El middleware `Autenticar` (US-10) deja de descartar los claims: expone el
   `sub` del JWT como identidad del request.
2. Un guard `ExigirAccesoAProyecto`, registrado en `setupRouter` justo después
   de `Autenticar`, responde `404 PROYECTO_NO_ENCONTRADO` a quien no es miembro
   (reutiliza el centinela y el mapeo existentes; **no** se crea un `403`).
3. El listado pasa a filtrarse por membresía y el alta vincula al creador.

**Decisión de proceso:** no se creó una issue en GitHub (el usuario scrum master
pidió no tocar GitHub); el nombre `US-F10` se usa únicamente en la
documentación y en los mensajes de commit.

La historia complementa a US-08 (asociar/quitar miembros) y a US-10 (JWT y
middleware); el modelo `proyecto_usuarios` con `rol` ya existe desde US-01/US-08
y **no hay cambios de esquema**.

**Trazabilidad:** ciclo TDD (RED → GREEN → REFACTOR) en backend Go y frontend.
Los escenarios BDD de `docs/features/US-F10-acceso-proyectos-membresia.feature`
son el origen de los tests de `internal/handler`, `internal/service`,
`internal/repository` y de los tests Vitest del frontend.

## Entradas

### Precondiciones

- US-10 en marcha: middleware default-deny y JWT HS256 con claims `sub` (id en
  texto), `email`, `iat`, `exp`.
- Tabla `proyecto_usuarios` (`proyecto_id`, `usuario_id`, `rol`) con clave
  primaria compuesta, generada por `AutoMigrate`.
- `UsuarioRepository.EstaAsociado(ctx, proyectoID, usuarioID)` disponible
  (US-08).
- Las rutas de US-01, US-02, US-03, US-04, US-05, US-06, US-07 y US-09 están
  registradas **después** del `r.Use(...)` del router (Gin copia la cadena de
  middleware al registrar la ruta).

### Quién pide qué

| Actor | Operación |
| :--- | :--- |
| Usuario autenticado | `GET /api/proyectos` (solo los suyos), `POST /api/proyectos` (crea y queda vinculado), `GET /api/usuarios` |
| Miembro del proyecto | `GET/PUT/PATCH/DELETE /api/proyectos/{id}`, `POST/DELETE /api/proyectos/{id}/usuarios` y todas las rutas anidadas (historias, sprints, sprint backlog, defectos, criterios) |
| No miembro | Recibe `404 PROYECTO_NO_ENCONTRADO` en todo lo anterior |

## Salidas esperadas

### Contrato de la historia

| HTTP | `error` | Cuándo ocurre | Uso en el frontend |
| :---: | :--- | :--- | :--- |
| `200` | — | `GET /api/proyectos` con al menos un proyecto del usuario | Filas del listado |
| `200` | — | `GET /api/proyectos` sin proyectos | `{"total": 0, "proyectos": []}` → estado vacío |
| `201` | — | `POST /api/proyectos`: el creador queda vinculado | `cantidad_usuarios: 1` |
| `401` | `TOKEN_INVALIDO` | Token sin claim `sub` o `sub` no numérico | Cierre de sesión (US-F08) |
| `401` | `TOKEN_REQUERIDO` | Sin token (US-10, sin cambios) | Cierre de sesión |
| `400` | `PARAMETRO_INVALIDO` | `id` no numérico o `0`: el guard **no** intercepta | Mensaje de parámetro |
| `404` | `PROYECTO_NO_ENCONTRADO` | El proyecto no existe, está dado de baja **o el usuario no es miembro** | `No tenés acceso a este proyecto o no existe` |
| `500` | `ERROR_INTERNO` | Fallo de persistencia en el guard | Error genérico + reintento |

- **No existe `403`**: la respuesta para un proyecto ajeno es idéntica a la de
  un proyecto inexistente, para no revelar la existencia del recurso (misma
  filosofía anti-enumeración que US-10/US-F08).
- El guard responde **antes** que el handler: un no miembro nunca llega al
  service del recurso anidado.

### Orden de evaluación por request

1. Autenticación (`Autenticar`): sin token → `401 TOKEN_REQUERIDO`; token
   inválido o sin `sub` numérico → `401 TOKEN_INVALIDO`.
2. Guard `ExigirAccesoAProyecto`: si la ruta no tiene `:id` de proyecto no hace
   nada; si tiene y el usuario no es miembro → `404 PROYECTO_NO_ENCONTRADO`.
3. Handler: sintaxis (`400 JSON_INVALIDO`, `400 PARAMETRO_INVALIDO`, fechas).
4. Service: reglas del recurso (`422`, `409`, `404` propios).

## Reglas de negocio

- **RB-01** El middleware de autenticación expone la identidad del token
  (`sub`) como `uint` en el contexto de Gin. Un token sin `sub` o con `sub` no
  numérico se rechaza con `401 TOKEN_INVALIDO`. → **CA-1**
- **RB-02** `GET /api/proyectos` devuelve **solo** los proyectos no dados de
  baja donde el usuario es miembro; sin proyectos responde
  `{"total": 0, "proyectos": []}`. → **CA-2, CA-3**
- **RB-03** `POST /api/proyectos` vincula al creador como miembro (rol
  `"Constructor de Producto"`) **en la misma transacción** que el alta: o existe
  el proyecto con su creador, o no existe nada. → **CA-4**
- **RB-04** Toda ruta cuyo path incluye `:id` de proyecto exige membresía:
  quien no es miembro recibe `404 PROYECTO_NO_ENCONTRADO` sin consultar el
  service del recurso. Aplica a detalle, modificación, cambio de estado, baja,
  alta/baja de miembros e historias, sprints, sprint backlog, defectos y
  criterios. → **CA-5, CA-6**
- **RB-05** El guard no interviene en `/health`, `/api/auth/*`,
  `GET /api/proyectos`, `POST /api/proyectos` ni `/api/usuarios`: esas rutas
  no llevan `:id` de proyecto. → **CA-7**
- **RB-06** Un `id` no numérico o `0` responde `400 PARAMETRO_INVALIDO` desde
  el handler: el guard no intercepta parámetros malformados. → **CA-7**
- **RB-07** Dentro de un proyecto **cualquier miembro** puede leer, editar,
  cerrar, eliminar y gestionar el equipo (opción A). No hay roles de dueño:
  los permisos por rol y la columna `creado_por` quedan fuera de alcance. →
  **CA-8**
- **RB-08** Ser miembro es también requisito para agregar o quitar integrantes:
  el propio guard cubre `POST/DELETE /api/proyectos/{id}/usuarios`, de modo que
  un usuario no puede auto-insertarse en un proyecto ajeno. → **CA-6**

## Restricciones

- **Sin cambios de esquema**: se reutiliza `proyecto_usuarios` y
  `UsuarioRepository.EstaAsociado`/`Vincular`.
- **Sin código nuevo de errores**: se reutiliza `domain.ErrProyectoNoEncontrado`
  y su rama en `handler/buscarMapeo` (`404`).
- El guard vive en `internal/handler` junto a `Autenticar` (la autorización es
  una preocupación transversal, igual que la autenticación en US-10); la regla
  de negocio "¿es miembro?" se implementa en `ProyectoService.PerteneceA`.
- El guard se registra en `setupRouter` **inmediatamente después** de
  `handler.Autenticar(...)` y **antes** de registrar las rutas de proyectos y
  recursos anidados: en Gin, las rutas registradas antes del `Use` no reciben
  el middleware.
- Textos de interfaz en español; tests con Vitest + React Testing Library y
  mocks de la API.

## Casos límite

- Proyecto inexistente y proyecto ajeno: **misma** respuesta
  (`404 PROYECTO_NO_ENCONTRADO` con el mismo `mensaje`).
- Proyecto dado de baja lógicamente: `404` (el filtro `borrado_en IS NULL`
  sigue aplicando, también para miembros).
- Usuario sin proyectos: `200` con `total: 0`; el frontend muestra el estado
  vacío "Todavía no participás en ningún proyecto".
- Usuario recién registrado (US-F09): listado vacío hasta que alguien lo
  agregue con `POST /api/proyectos/{id}/usuarios` (US-08).
- Usuario quitado del proyecto: a partir de ese momento recibe `404` en todas
  las rutas de ese proyecto.
- Token emitido antes de esta historia (con `sub`): sigue siendo válido.
- Dos usuarios crean proyectos distintos: cada uno solo ve el suyo.
- `id` con forma `"abc"`, `""` o `0` → `400 PARAMETRO_INVALIDO`, nunca `404`.
- El guard hace una consulta de membresía por request a ruta con `:id`; se
  acepta por volumen de este proyecto (sin caché).

## Condiciones de error

| Situación | Origen | Respuesta |
| :--- | :--- | :--- |
| Sin `Authorization` o sin prefijo `Bearer` | `Autenticar` (US-10) | `401 TOKEN_REQUERIDO` |
| Token con firma/algoritmo/expiración inválidos | `Autenticar` (US-10) | `401 TOKEN_INVALIDO` |
| Token sin claim `sub` o `sub` no numérico | `Autenticar` (RB-01) | `401 TOKEN_INVALIDO` |
| No miembro en ruta con `:id` | `ExigirAccesoAProyecto` (RB-04) | `404 PROYECTO_NO_ENCONTRADO` |
| `:id` no numérico o `0` | handler (RB-06) | `400 PARAMETRO_INVALIDO` |
| Fallo de membresía en la base | `ProyectoService.PerteneceA` | `500 ERROR_INTERNO` |
| Alta de proyecto con fallo de persistencia | US-01 (sin cambios) | `500 ERROR_INTERNO` y sin proyecto parcial (RB-03) |

## Criterios de aceptación

- [ ] **CA-1** Given un token válido, When el request pasa por `Autenticar`,
  Then el handler ve el `uint` de `sub` en el contexto; Given un token sin `sub`
  o con `sub` no numérico, When llega a la ruta, Then responde
  `401 TOKEN_INVALIDO`.
- [ ] **CA-2** Given el usuario A es miembro de los proyectos 1 y 3, When pide
  `GET /api/proyectos`, Then solo recibe 1 y 3 en ese orden.
- [ ] **CA-3** Given el usuario sin proyectos, When pide `GET /api/proyectos`,
  Then responde `200` con `{"total": 0, "proyectos": []}`.
- [ ] **CA-4** Given un usuario autenticado, When crea un proyecto, Then
  responde `201` con `cantidad_usuarios: 1` y aparece en su listado (misma
  transacción: si falla el vínculo no queda el proyecto).
- [ ] **CA-5** Given el proyecto 1 pertenece a A, When B (autenticado) pide
  `GET /api/proyectos/1`, Then recibe `404 PROYECTO_NO_ENCONTRADO` con el
  mensaje `el proyecto no existe`.
- [ ] **CA-6** Given lo mismo para B, When pide una ruta anidada
  (`GET /api/proyectos/1/historias`, `POST .../usuarios`, etc.), Then también
  recibe `404` y el service del recurso **no** es invocado; cuando B ya es
  miembro, Then la misma ruta responde como antes.
- [ ] **CA-7** Given cualquier usuario, When pide `GET /api/proyectos/abc` o
  `GET /health` o `POST /api/proyectos`, Then el guard no interviene
  (`400 PARAMETRO_INVALIDO` o la respuesta normal).
- [ ] **CA-8** Given A es miembro del proyecto, When A lo modifica, lo cierra,
  lo elimina o gestiona su equipo, Then todas las operaciones responden como
  hasta ahora (opción A: sin roles de dueño).

## Archivos afectados

| Archivo | Cambio |
| :--- | :--- |
| `internal/handler/middleware_auth.go` | `Autenticar` parsea `sub` y hace `c.Set`; `401 TOKEN_INVALIDO` si falta o no es numérico |
| `internal/handler/middleware_auth_test.go` | Tests de identidad en el contexto (RED/GREEN) |
| `internal/handler/middleware_acceso.go` | **Nuevo** guard `ExigirAccesoAProyecto` |
| `internal/handler/middleware_acceso_test.go` | **Nuevo** tests del guard (miembro, ajeno, sin `:id`, id malformado) |
| `internal/service/proyecto_service.go` | `PerteneceA`, `Listar(ctx, usuarioID)`, `Crear(ctx, usuarioID, …)` vincula al creador |
| `internal/service/proyecto_service_test.go` | Firma nueva + tests de `PerteneceA` y del vínculo del creador |
| `internal/domain/proyecto_repository.go` | `Listar` → `ListarPorUsuario(ctx, usuarioID)` |
| `internal/repository/proyecto_repository.go` | Implementación con `JOIN proyecto_usuarios` |
| `internal/repository/proyecto_repository_test.go` | **Nuevo** tests de integración del listado filtrado |
| `internal/handler/proyecto_handler.go` | Interface `ServicioProyectos` gana `PerteneceA`; `Listar`/`Crear` leen el actor del contexto |
| `internal/handler/mocks_test.go`, `proyecto_handler_test.go` | Dobles y router de prueba con usuario en contexto |
| `cmd/api/router.go` | `r.Use(handler.ExigirAccesoAProyecto(...))` después de `Autenticar` |
| `cmd/api/router_test.go` | Doble con `PerteneceA` + test del guard en el router real |
| `frontend/src/pages/proyecto/*.tsx` | `404 PROYECTO_NO_ENCONTRADO` → "No tenés acceso a este proyecto o no existe" |
| `frontend/src/pages/ProyectosPage.tsx` | Copy del estado vacío (ya existente) |
| `frontend/src/pages/proyecto/*.test.tsx`, `ProyectosPage.test.tsx` | Tests de los mensajes anteriores |

No se modifica el esquema, `internal/repository/migraciones.go` ni
`internal/domain/errores.go`.

## Dependencias

- **US-10** (`#61`): JWT, middleware y claims `sub`.
- **US-01**: modelo de proyectos y `CrearConUsuarios` (atómica).
- **US-08** (`#39`): `EstaAsociado`, `Vincular` y el endpoint de asociación.
- **US-F08** (`#62`): manejo de `401` en el cliente.

## Fuera de alcance

- Permisos por rol dentro del proyecto (opción B: `creado_por` o rol líder).
- Restringir `GET /api/usuarios` (sigue abierto a cualquier autenticado).
- Recuperación de contraseña, verificación de email, MFA.
- Caché de membresías o índices adicionales.
- El bug preexistente de `ProyectoRepository.TieneHistorial` (siempre `false`).

## Escenarios BDD

Ver `docs/features/US-F10-acceso-proyectos-membresia.feature`.
