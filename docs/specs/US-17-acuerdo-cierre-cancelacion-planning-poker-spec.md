# Especificación: US-17 - Acuerdo, cierre y cancelación de Planning Poker

## Objetivo

Que cualquier miembro de un proyecto pueda **acordar los puntos** de una sesión de
Planning Poker en estado `Revelada`, dejando la historia estimada y la sesión
cerrada de forma atómica, o **cancelar** una sesión en `Votando`/`Revelada` sin
estimar la historia.

La historia satisface el issue `SebastianAguilera/ingenieria-y-calidad#79`.
**Requiere** US-14 (sesiones y rondas), US-15 (votación oculta) y US-16
(revelado, dispersión y nuevas rondas).
**Bloquea a** US-F17 (frontend de acuerdo y cierre).

**Trazabilidad:** Issue #79 → SDD → criterios de aceptación → escenarios BDD
(`docs/features/US-17-acuerdo-cierre-cancelacion-planning-poker.feature`) → tests Go
(`internal/domain`, `internal/service`, `internal/handler`,
`internal/repository`) → código.

## Cross-referencias

- **US-05**: escala de puntos `0, 0.5, 1, 2, 3, 5, 8, 13, 21` en `float64`.
- **US-14**: tablas `poker_sessions`, `poker_rounds`, `poker_votes` y
  `story_estimates`; estados `Votando`, `Revelada`, `Cerrada` y `Cancelada`.
- **US-15**: voto por ronda y usuario con upsert; solo se vota en `Votando`.
- **US-16**: solo se acuerda desde `Revelada`; la ronda actual está revelada.
- **US-11**: guard de membresía; cualquier miembro puede acordar o cancelar. Un
  no miembro recibe `404`.
- **US-F17**: consume estos endpoints. Fuera de alcance en US-17.

## Entradas

- `POST /api/proyectos/{id}/poker-sessions/{sesionId}/acuerdo`, cuerpo:

  ```json
  { "puntos": 5 }
  ```

- `POST /api/proyectos/{id}/poker-sessions/{sesionId}/cancelar`, sin cuerpo.
- Identidad del request: claim `sub` del JWT.
- `:id` y `:sesionId` se parsean como `uint`; un valor no numérico o cero es
  `400 PARAMETRO_INVALIDO`.
- `puntos` se lee de forma laxa para distinguir JSON malformado de un valor
  ausente o no numérico, igual que en US-15.

## Reglas de negocio

- **RB-01**: solo se puede acordar desde `Revelada`. Desde `Votando` la respuesta
  es `409 SESION_POKER_NO_REVELADA` y no se modifica nada.
- **RB-02**: no se puede acordar desde `Cerrada` ni `Cancelada`:
  `409 SESION_POKER_CERRADA`, sin modificar nada.
- **RB-03**: `puntos` debe pertenecer a la escala de US-05. Un valor ausente, no
  numérico o fuera de escala produce `422 PUNTOS_POKER_INVALIDOS` con
  `campo = "puntos"`, sin llamar al servicio ni tocar la base.
- **RB-04**: acordar ejecuta en **una única transacción**:
  1. bloquear la fila de `poker_sessions` (`SELECT ... FOR UPDATE`);
  2. validar que el estado sea `Revelada`;
  3. actualizar `historias.story_points` con los puntos acordados;
  4. actualizar la sesión a `Cerrada`, fijar `puntos_acordados` y `cerrado_en`;
  5. insertar una fila en `story_estimates` con `historia_id`, `sesion_id`,
     `ronda_id` (la ronda actual), `puntos`, `fuente = "planning_poker"`,
     `estimado_por = usuarioID` y `estimado_en = now()`;
  6. confirmar. Cualquier error revierte **todo**: historia, sesión y estimación.
- **RB-05**: el acuerdo conserva votos y rondas; no los borra ni los modifica.
- **RB-06**: `puntos_acordados`, `historias.story_points` y
  `story_estimates.puntos` deben coincidir exactamente.
- **RB-07**: solo se puede cancelar desde `Votando` o `Revelada`. Desde `Cerrada`
  o `Cancelada` la respuesta es `409 SESION_POKER_CERRADA`, sin modificar nada.
- **RB-08**: cancelar ejecuta en una transacción con la sesión bloqueada y fija
  `estado = Cancelada` y `cancelado_en = now()`.
- **RB-09**: cancelar **no** fija `puntos_acordados`, **no** fija `cerrado_en`,
  **no** inserta en `story_estimates` y **no** modifica
  `historias.story_points`.
- **RB-10**: cancelar conserva votos y rondas. La historia queda libre para una
  nueva sesión porque el índice único parcial solo cubre `Votando` y `Revelada`.
- **RB-11**: cualquier miembro del proyecto puede acordar o cancelar. El creador
  de la sesión no tiene privilegios extra.
- **RB-12**: sesión inexistente o de otro proyecto: `404
  SESION_POKER_NO_ENCONTRADA`. Proyecto inexistente o solicitante no miembro:
  `404 PROYECTO_NO_ENCONTRADO`.
- **RB-13**: el detalle de una sesión `Cerrada` incluye `story_estimate` como
  objeto completo. En sesiones sin estimación el campo es `null` o está ausente.
- **RB-14**: el historial de sesiones de una historia incluye `story_estimate`
  en cada item `Cerrada` y `cancelado_en` en cada item `Cancelada`.
- **RB-15**: la respuesta de acordar y cancelar es el detalle de la sesión ya
  actualizada, con el mismo contrato que `GET .../poker-sessions/{sesionId}`.
- **RB-16**: precedencia de errores de negocio: proyecto/sesión inexistente
  (`404`) → puntos inválidos (`422`) → estado inválido (`409`). El JSON
  malformado y los parámetros de ruta inválidos (`400`) se resuelven en el
  handler antes de invocar al servicio.
- **RB-17**: el acuerdo y la cancelación compiten bajo el mismo bloqueo de fila.
  Si dos operaciones llegan a la vez, solo una gana; la perdedora recibe
  `409 SESION_POKER_CERRADA` y no deja escrituras parciales.

## Cambios de esquema y modelos (GORM)

No se modifica `tablas.sql`; el esquema se mantiene con `AutoMigrate` en
`internal/repository/migraciones.go`.

- `PokerSessionModel`: agregar `CanceladoEn *time.Time`
  (`column:cancelado_en;type:timestamptz`). El dominio `SesionPoker` agrega el
  mismo campo.
- `StoryEstimateModel`: agregar `SesionID uint` (`column:sesion_id;not null`) y
  `RondaID uint` (`column:ronda_id;not null`). El dominio `StoryEstimate` agrega
  los mismos campos.
- `SesionPoker` agrega `StoryEstimate *StoryEstimate` para transportar la
  estimación en el detalle y en el historial.
- `DetalleSesionPoker` obtiene la estimación a través de `Sesion.StoryEstimate`.

## Salidas esperadas

### `POST .../poker-sessions/{sesionId}/acuerdo` → `200`

Devuelve el detalle de la sesión cerrada:

```json
{
  "id": 9,
  "historia_id": 5,
  "estado": "Cerrada",
  "ronda_actual": 1,
  "puntos_acordados": 5,
  "creado_por": { "id": 2, "nombre": "Ana" },
  "creado_en": "2026-11-13T10:00:00Z",
  "cerrado_en": "2026-11-13T18:00:00Z",
  "cancelado_en": null,
  "total_miembros": 3,
  "votos_registrados": 3,
  "ya_vote": true,
  "mi_voto": 3,
  "votos": [
    { "usuario_id": 2, "usuario_nombre": "Ana", "puntos": 3, "ronda": 1 }
  ],
  "dispersion": { "...": "..." },
  "historial_rondas": [ { "...": "..." } ],
  "story_estimate": {
    "id": 12,
    "historia_id": 5,
    "sesion_id": 9,
    "ronda_id": 3,
    "puntos": 5,
    "fuente": "planning_poker",
    "estimado_por": 2,
    "estimado_en": "2026-11-13T18:00:00Z"
  }
}
```

### `POST .../poker-sessions/{sesionId}/cancelar` → `200`

Devuelve el detalle de la sesión cancelada. No incluye `story_estimate`:

```json
{
  "id": 9,
  "historia_id": 5,
  "estado": "Cancelada",
  "ronda_actual": 1,
  "puntos_acordados": null,
  "cerrado_en": null,
  "cancelado_en": "2026-11-13T18:00:00Z",
  "story_estimate": null,
  "votos_registrados": 3,
  "historial_rondas": [ { "...": "..." } ]
}
```

### `GET .../poker-sessions/{sesionId}` y `GET .../historias/{historiaId}/poker-sessions`

Incluyen `story_estimate` cuando existe y `cancelado_en` cuando la sesión fue
cancelada. En el historial, cada item es independiente y conserva su propia
estimación.

## Condiciones de error

| Situación | Código HTTP | `error` |
| :--- | :--- | :--- |
| `id` o `sesionId` inválidos | `400` | `PARAMETRO_INVALIDO` |
| JSON malformado en acuerdo | `400` | `JSON_INVALIDO` |
| Sin token / token inválido | `401` | `TOKEN_REQUERIDO` / `TOKEN_INVALIDO` |
| No miembro / proyecto inexistente | `404` | `PROYECTO_NO_ENCONTRADO` |
| Sesión inexistente o de otro proyecto | `404` | `SESION_POKER_NO_ENCONTRADA` |
| `puntos` ausente, no numérico o fuera de escala | `422` | `PUNTOS_POKER_INVALIDOS` |
| Acordar en `Votando` | `409` | `SESION_POKER_NO_REVELADA` |
| Acordar o cancelar en `Cerrada`/`Cancelada` | `409` | `SESION_POKER_CERRADA` |

## Restricciones y casos límite

- Acordar `0` y `0.5` es válido; `historias.story_points` queda en `0`/`0.5` y
  `puntos_acordados` no queda `null`.
- Acordar `4`, `-1`, `22`, `0.25` o `100` es inválido.
- Un fallo al insertar `story_estimates` revierte la actualización de la historia
  y el cierre de la sesión.
- Cancelar una sesión `Revelada` conserva sus votos y permite crear una sesión
  nueva para la misma historia.
- Cancelar no borra rondas ni votos; el historial los sigue mostrando.
- La estimación se asocia a la ronda actual al momento del acuerdo, no a una
  ronda arbitraria del request.
- No se agregan endpoints de reapertura ni de edición de una estimación cerrada.
- Fuera de alcance: notificaciones, edición posterior del `story_points` desde
  este endpoint y frontend.

## Criterios de aceptación

- [ ] **CA-1**: Given una sesión `Revelada`, When un miembro acuerda `5`, Then la
  sesión queda `Cerrada` con `puntos_acordados = 5` y `cerrado_en` informado.
- [ ] **CA-2**: Given el acuerdo anterior, Then `historias.story_points = 5` y
  existe una fila en `story_estimates` con `sesion_id`, `ronda_id`, `puntos = 5`,
  `fuente = planning_poker` y `estimado_por` del miembro.
- [ ] **CA-3**: Given `puntos` `0` o `0.5`, When se acuerda, Then el acuerdo es
  válido y se persiste el mismo valor.
- [ ] **CA-4**: Given `puntos` ausente, no numérico o fuera de escala, When se
  acuerda, Then `422 PUNTOS_POKER_INVALIDOS` y no se modifica nada.
- [ ] **CA-5**: Given una sesión `Votando`, When se intenta acordar, Then `409
  SESION_POKER_NO_REVELADA` y no se modifica nada.
- [ ] **CA-6**: Given una sesión `Cerrada` o `Cancelada`, When se intenta acordar,
  Then `409 SESION_POKER_CERRADA` y no se modifica nada.
- [ ] **CA-7**: Given un fallo forzado al insertar `story_estimates`, When se
  acuerda, Then se revierten `story_points`, el estado y los puntos de la sesión.
- [ ] **CA-8**: Given una sesión `Votando` o `Revelada`, When un miembro cancela,
  Then queda `Cancelada` con `cancelado_en` y sin `story_estimate`.
- [ ] **CA-9**: Given una sesión cancelada, Then `story_points` no cambia, no se
  fija `puntos_acordados`/`cerrado_en` y se conservan votos y rondas.
- [ ] **CA-10**: Given una sesión cancelada, When se crea otra sesión para la
  misma historia, Then la creación es exitosa.
- [ ] **CA-11**: Given una sesión `Cerrada` o `Cancelada`, When se intenta
  cancelar, Then `409 SESION_POKER_CERRADA` y no se modifica nada.
- [ ] **CA-12**: Given un acuerdo o cancelación concurrentes, Then solo uno gana y
  el perdedor recibe `409 SESION_POKER_CERRADA` sin escrituras parciales.
- [ ] **CA-13**: Given una sesión cerrada, When se consulta el detalle o el
  historial, Then `story_estimate` aparece como objeto completo.
- [ ] **CA-14**: Given un no miembro o una sesión de otro proyecto, When intenta
  acordar o cancelar, Then `404` con el código correspondiente.

## Definición de terminado

- [ ] Spec SDD y escenarios BDD escritos antes del código.
- [ ] Tests RED de dominio, repositorio, servicio y handler para acuerdo,
  cancelación, atomicidad, historial y errores.
- [ ] `go test ./...` en verde, `gofmt -l .` vacío y `go vet ./...` sin avisos.
- [ ] E2E por curl de acordar, cancelar, detalle e historial.
- [ ] `AutoMigrate` agrega `poker_sessions.cancelado_en`,
  `story_estimates.sesion_id` y `story_estimates.ronda_id` sin modificar
  `tablas.sql`.
