# Especificación: US-14 - Sesiones de Planning Poker por historia

## Objetivo

Que los miembros de un proyecto puedan **iniciar y consultar sesiones de
Planning Poker por historia** para estimar en equipo sin perder el contexto del
backlog. Esta historia implementa la base backend: modelo de sesión, ronda, voto
y estimación; creación, historial, detalle y listado de sesiones activas.

La historia satisface el issue `SebastianAguilera/ingenieria-y-calidad#74`.
**Bloquea a US-15** (votar), **US-16** (revelar/acordar), **US-17** (cerrar y
cancelar), **US-F15** y **US-F19**. No incluye votar, revelar, acordar, cerrar
ni cancelar.

**Trazabilidad:** Issue #74 → SDD → criterios de aceptación → escenarios BDD
(`docs/features/US-14-sesiones-planning-poker.feature`) → tests Go
(`internal/domain`, `internal/service`, `internal/handler`,
`internal/repository`) → código.

## Cross-referencias

- **US-01**: existencia y baja lógica del proyecto.
- **US-02**: la historia debe existir dentro del proyecto consultado.
- **US-05**: escala de story points `0, 0.5, 1, 2, 3, 5, 8, 13, 21`; por eso
  los puntos se modelan en `float64` (admite `0.5`).
- **US-F10/US-11**: guard de membresía (`ExigirAccesoAProyecto`) y contexto de
  identidad (`UsuarioDesdeContexto`). Cualquier miembro puede crear y consultar;
  un no miembro recibe `404`.
- **US-15/16/17**: consumirán los modelos y las rutas creados aquí.

## Entradas

- `POST /api/proyectos/{id}/historias/{historiaId}/poker-sessions` sin cuerpo.
- `GET /api/proyectos/{id}/historias/{historiaId}/poker-sessions`.
- `GET /api/proyectos/{id}/poker-sessions/{sesionId}`.
- `GET /api/proyectos/{id}/poker-sessions/activas`.
- Identidad del request: claim `sub` del JWT (`creado_por` y `ya_vote`).
- `:id`, `:historiaId` y `:sesionId` se parsean como `uint`; un valor no
  numérico o cero es `400 PARAMETRO_INVALIDO`.

## Reglas de negocio

- **RB-01**: al crear una sesión se persiste en estado `Votando`, con
  `ronda_actual = 1`, una `PokerRound` número 1 y `creado_por` tomado del JWT.
- **RB-02**: solo puede existir **una sesión activa por historia**. Se consideran
  activas `Votando` y `Revelada`. Si ya existe una, la creación responde
  `409 SESION_POKER_ACTIVA` y no se persiste nada.
- **RB-03**: `Cerrada` y `Cancelada` no bloquean: se puede crear una nueva
  sesión después de ellas.
- **RB-04**: una historia en estado `Hecho` no es estimable:
  `409 HISTORIA_NO_ESTIMABLE`, sin persistir nada.
- **RB-05**: la historia debe existir dentro del proyecto de la ruta. Historia
  inexistente o de otro proyecto: `404 HISTORIA_NO_ENCONTRADA`. Proyecto
  inexistente: `404 PROYECTO_NO_ENCONTRADO`.
- **RB-06**: un no miembro recibe `404 PROYECTO_NO_ENCONTRADO` (guard de US-F10),
  tanto al crear como al consultar. Cualquier miembro puede crear y consultar.
- **RB-07**: el historial devuelve todas las sesiones de la historia, incluidas
  `Cerrada` y `Cancelada`, ordenadas por creación descendente.
- **RB-08**: mientras la sesión está en `Votando`, el detalle **no expone votos
  ajenos ni puntos individuales** de otros. Solo informa `votos_registrados`,
  `ya_vote` y `mi_voto` (puntos propios). La lista de votos queda vacía.
- **RB-09**: en estado `Revelada`, el detalle devuelve los votos con `usuario_id`,
  `nombre` y `puntos`.
- **RB-10**: el detalle y las activas informan `total_miembros` del proyecto.
- **RB-11**: `GET .../activas` devuelve solo sesiones `Votando` o `Revelada` del
  proyecto. Cada ítem incluye `sesion_id`, `estado`, `ronda_actual`,
  `creado_por` (`id`, `nombre`), `creado_en`, `votos_registrados`,
  `total_miembros`, `ya_vote` y resumen de historia (`id`, `codigo`, `titulo`).
  **Nunca** incluye votos, ni siquiera en `Revelada`; solo conteos.
- **RB-12**: sin sesiones activas, la respuesta es `200` con `[]` (nunca `null`).
- **RB-13**: el índice único parcial de "una sesión activa por historia" aplica
  solo a `Votando`/`Revelada` y es la red de seguridad ante concurrencia.
- **RB-14**: las respuestas no emiten `proyecto_id` (va implícito en la ruta).

## Modelo de datos

### `poker_sessions`

| Campo | Tipo | Notas |
| :--- | :--- | :--- |
| `id` | `uint` PK | |
| `proyecto_id` | `uint` FK `proyectos.id` | `ON DELETE CASCADE` |
| `historia_id` | `uint` FK `historias.id` | `ON DELETE CASCADE` |
| `creado_por` | `uint` FK `usuarios.id` | `ON DELETE RESTRICT` |
| `ronda_actual` | `int` | `NOT NULL DEFAULT 1` |
| `estado` | `varchar(20)` | `Votando`, `Revelada`, `Cerrada`, `Cancelada` |
| `puntos_acordados` | `*float64` | admite `0.5` |
| `creado_en` / `actualizado_en` | `timestamptz` | |
| `cerrado_en` | `*timestamptz` | |

Índice parcial:

```sql
CREATE UNIQUE INDEX IF NOT EXISTS ux_poker_sesiones_activas_historia
ON poker_sessions (historia_id) WHERE estado IN ('Votando', 'Revelada');
```

### `poker_rounds`

| Campo | Tipo | Notas |
| :--- | :--- | :--- |
| `id` | `uint` PK | |
| `sesion_id` | `uint` FK `poker_sessions.id` | `ON DELETE CASCADE` |
| `numero_ronda` | `int` | |
| `revelada_en` | `*timestamptz` | |
| `creado_en` | `timestamptz` | |

Índice único `(sesion_id, numero_ronda)`.

### `poker_votes`

| Campo | Tipo | Notas |
| :--- | :--- | :--- |
| `id` | `uint` PK | |
| `ronda_id` | `uint` FK `poker_rounds.id` | `ON DELETE CASCADE` |
| `usuario_id` | `uint` FK `usuarios.id` | `ON DELETE RESTRICT` |
| `puntos` | `float64` | admite `0.5` |
| `votado_en` | `timestamptz` | |

Índice único `(ronda_id, usuario_id)`.

### `story_estimates`

Modelo mínimo para US-16: `id`, `historia_id` FK, `puntos float64`,
`fuente` (`direct`/`planning_poker`), `estimado_por` FK, `estimado_en`. Sin
endpoints en esta historia.

## Salidas esperadas

### `POST .../poker-sessions` → `201`

```json
{
  "id": 1,
  "historia_id": 5,
  "estado": "Votando",
  "ronda_actual": 1,
  "creado_por": { "id": 2, "nombre": "Ana" },
  "creado_en": "2026-11-13T10:00:00Z",
  "puntos_acordados": null,
  "total_miembros": 3,
  "votos_registrados": 0,
  "ya_vote": false,
  "mi_voto": null,
  "votos": []
}
```

### `GET .../historias/{historiaId}/poker-sessions` → `200`

Array de sesiones (incluye `Cerrada`/`Cancelada`), descendente por creación:

```json
[
  {
    "id": 3, "estado": "Cerrada", "ronda_actual": 2,
    "puntos_acordados": 5, "creado_por": { "id": 2, "nombre": "Ana" },
    "creado_en": "2026-11-10T09:00:00Z"
  }
]
```

### `GET .../poker-sessions/{sesionId}` → `200`

```json
{
  "id": 1, "historia_id": 5, "estado": "Revelada", "ronda_actual": 1,
  "creado_por": { "id": 2, "nombre": "Ana" },
  "creado_en": "2026-11-13T10:00:00Z",
  "puntos_acordados": null,
  "total_miembros": 3, "votos_registrados": 2, "ya_vote": true,
  "mi_voto": 3,
  "votos": [
    { "usuario_id": 2, "nombre": "Ana", "puntos": 3 },
    { "usuario_id": 7, "nombre": "Ben", "puntos": 5 }
  ]
}
```

### `GET .../poker-sessions/activas` → `200`

```json
[
  {
    "sesion_id": 1, "estado": "Votando", "ronda_actual": 1,
    "creado_por": { "id": 2, "nombre": "Ana" },
    "creado_en": "2026-11-13T10:00:00Z",
    "votos_registrados": 1, "total_miembros": 3, "ya_vote": true,
    "historia": { "id": 5, "codigo": "US-05", "titulo": "Estimar" }
  }
]
```

## Condiciones de error

| HTTP | `error` | Cuándo ocurre |
| :---: | :--- | :--- |
| `400` | `PARAMETRO_INVALIDO` | `:id`, `:historiaId` o `:sesionId` no numérico/cero |
| `401` | `TOKEN_REQUERIDO` / `TOKEN_INVALIDO` | Sin JWT o inválido (US-10) |
| `404` | `PROYECTO_NO_ENCONTRADO` | Proyecto inexistente o solicitante no miembro |
| `404` | `HISTORIA_NO_ENCONTRADA` | Historia inexistente o de otro proyecto |
| `404` | `SESION_POKER_NO_ENCONTRADA` | Sesión inexistente o de otro proyecto |
| `409` | `SESION_POKER_ACTIVA` | Ya hay una sesión `Votando`/`Revelada` en la historia |
| `409` | `HISTORIA_NO_ESTIMABLE` | La historia está en estado `Hecho` |
| `500` | `ERROR_INTERNO` | Fallo de persistencia |

## Restricciones y casos límite

- Fuera de alcance: votar, revelar, calcular dispersión, acordar puntos, cerrar,
  cancelar, notificaciones push/email/WebSocket y frontend.
- Una sesión en `Revelada` sigue contando como activa (bloquea nuevas sesiones).
- Una sesión `Cerrada`/`Cancelada` con `puntos_acordados` no bloquea la creación
  de la siguiente.
- `total_miembros` cuenta vínculos en `proyecto_usuarios` del proyecto.
- La lista de activas no expone votos aunque la sesión esté `Revelada`.
- `votos_registrados` cuenta votos de la ronda actual.
- IDs enteros sin signo; `historiaId`/`sesionId` de otro proyecto se tratan como
  `404` sin filtrar información.

## Criterios de aceptación

- **CA-1**: un miembro crea una sesión sobre una historia sin sesión activa y
  queda `Votando`, ronda 1 y `creado_por` correcto.
- **CA-2**: con sesión `Votando`/`Revelada`, otro miembro recibe
  `409 SESION_POKER_ACTIVA` y no se crea nada.
- **CA-3**: tras una sesión `Cerrada`/`Cancelada` se permite crear otra.
- **CA-4**: una historia en `Hecho` produce `409 HISTORIA_NO_ESTIMABLE` sin crear.
- **CA-5**: un no miembro obtiene `404` al crear y al consultar.
- **CA-6**: historia inexistente o de otro proyecto produce `404`.
- **CA-7**: en `Votando` el detalle no expone votos ajenos ni puntos individuales.
- **CA-8**: el detalle incluye `total_miembros`.
- **CA-9**: cada ítem de `activas` incluye `total_miembros`.
- **CA-10**: sin activas, `activas` responde `200 []`.
- **CA-11**: `activas` excluye `Cerrada`/`Cancelada`.
- **CA-12**: una sesión `Revelada` aparece en `activas` con `estado: "Revelada"`
  y sin votos (solo conteos y `total_miembros`).
- **CA-13**: en `Revelada` el detalle devuelve votos con nombre y puntos.
- **CA-14**: el historial incluye `Cerrada`/`Cancelada` y ordena descendente.

## Definición de terminado

- [ ] Spec SDD y escenarios BDD escritos antes del código.
- [ ] `go test ./...` en verde.
- [ ] `AutoMigrate` verificado en base limpia y existente, incluido el índice
  parcial.
- [ ] E2E por curl de crear, listar, detalle y activas.
- [ ] Verificado que no se filtran votos ajenos en `Votando`.
