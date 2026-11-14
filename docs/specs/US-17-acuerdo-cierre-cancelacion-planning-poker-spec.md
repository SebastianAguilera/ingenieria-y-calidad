# Especificación: US-17 - Acuerdo, cierre y cancelación de Planning Poker

## Objetivo

Que cualquier miembro de un proyecto pueda **registrar la estimación acordada**
de una sesión de Planning Poker `Revelada` (cerrándola y dejando la historia
estimada) o **cancelar** una sesión `Votando`/`Revelada` (liberando la historia
sin cambios).

La historia satisface el issue `SebastianAguilera/ingenieria-y-calidad#79`.
**Requiere** US-14 (sesiones), US-15 (votación), US-16 (revelado) y US-05
(escala de story points y actualización de historias).
**Bloquea a** US-F18 (frontend de acuerdo y cierre).

**Trazabilidad:** Issue #79 → SDD → criterios de aceptación → escenarios BDD
(`docs/features/US-17-acuerdo-cierre-cancelacion-planning-poker.feature`) →
tests Go (`internal/domain`, `internal/service`, `internal/handler`,
`internal/repository`, `cmd/api`) → código.

## Cross-referencias

- **US-14**: tablas `poker_sessions`, `poker_rounds`, `poker_votes` y
  `story_estimates`; estados `Votando`, `Revelada`, `Cerrada` y `Cancelada`;
  índice único parcial de sesión activa por historia (`Votando`/`Revelada`).
- **US-15**: votos por ronda; se conservan al cerrar o cancelar.
- **US-16**: revelado y rondas; el acuerdo solo es posible tras revelar.
- **US-05**: escala `0, 0.5, 1, 2, 3, 5, 8, 13, 21` (`domain.StoryPointsValidos`)
  y escritura de `historias.story_points` mediante `HistoriaRepository.Actualizar`.
- **US-F10/US-11**: guard de membresía; cualquier miembro puede acordar o
  cancelar. Un no miembro recibe `404`.

## Entradas

- `POST /api/proyectos/{id}/poker-sessions/{sesionId}/acuerdo` con cuerpo
  `{ "puntos": number }`.
- `POST /api/proyectos/{id}/poker-sessions/{sesionId}/cancelar`, sin cuerpo.
- `GET /api/proyectos/{id}/poker-sessions/{sesionId}` (detalle).
- `GET /api/proyectos/{id}/historias/{historiaId}/poker-sessions` (historial).
- Identidad del request: claim `sub` del JWT.
- `:id` y `:sesionId` se parsean como `uint`; un valor no numérico o cero es
  `400 PARAMETRO_INVALIDO`. Un cuerpo JSON malformado es `400`; `puntos` ausente
  o no numérico es `422 PUNTOS_POKER_INVALIDOS` (mismo criterio que el voto de
  US-15).

## Reglas de negocio

### Acuerdo y cierre

- **RB-01**: solo se puede acordar desde `Revelada`. Si la sesión está `Votando`
  la respuesta es `409 SESION_POKER_NO_REVELADA`; si está `Cerrada` o
  `Cancelada`, `409 SESION_POKER_CERRADA`. En ambos casos no se modifica nada.
- **RB-02**: `puntos` debe pertenecer a la escala de US-05
  (`domain.StoryPointsValidos`). Fuera de escala: `422 PUNTOS_POKER_INVALIDOS`,
  sin modificar nada. `puntos` no tiene por qué coincidir con ningún voto.
- **RB-03**: el acuerdo fija `poker_sessions.puntos_acordados`, pasa la sesión a
  `Cerrada` y setea `cerrado_en` y `actualizado_en`.
- **RB-04**: el acuerdo actualiza `historias.story_points` con los puntos
  acordados a través de `HistoriaRepository.Actualizar` (camino de US-05). No se
  hace un `UPDATE historias` directo fuera de la transacción.
- **RB-05**: el acuerdo crea un `story_estimate` con `historia_id`, `sesion_id`,
  `ronda_id` (la ronda actual, ya revelada), `puntos`,
  `fuente = "planning_poker"`, `estimado_por` (usuario que acuerda) y
  `estimado_en`.
- **RB-06 (atomicidad)**: cerrar la sesión, actualizar la historia y crear el
  `story_estimate` ocurren en **una única transacción**. Si cualquiera de las
  tres operaciones falla se hace rollback total: la sesión sigue `Revelada`, la
  historia conserva sus `story_points` previos y no queda `story_estimate`.
- **RB-07 (concurrencia)**: la transacción bloquea la fila de la sesión
  (`SELECT ... FOR UPDATE`) y verifica el estado dentro de ella, igual que
  US-15/US-16. Dos acuerdos simultáneos producen un único cierre; el segundo
  recibe `409 SESION_POKER_CERRADA`.

### Cancelación

- **RB-08**: solo se puede cancelar desde `Votando` o `Revelada`. Si la sesión
  está `Cerrada` o `Cancelada`: `409 SESION_POKER_CERRADA`, sin modificar nada.
- **RB-09**: cancelar pasa la sesión a `Cancelada` y setea `cancelado_en` y
  `actualizado_en`. No setea `puntos_acordados`, no crea `story_estimate` y no
  modifica la historia.
- **RB-10**: cancelar conserva las rondas y los votos de la sesión (historial).
- **RB-11**: al quedar fuera de `Votando`/`Revelada`, el índice único parcial
  deja de cubrir la sesión: la historia queda libre para una sesión nueva.
- **RB-12**: la verificación de estado y la escritura de la cancelación ocurren
  en una transacción con la fila de la sesión bloqueada.

### Consultas y permisos

- **RB-13**: el detalle devuelve `estado`, `puntos_acordados`, `cerrado_en`,
  `cancelado_en` y, si la sesión fue acordada, `story_estimate`. En `Cerrada`
  los votos de la ronda actual siguen visibles. En `Cancelada` solo se ven si
  esa ronda llegó a revelarse: una sesión cancelada mientras se votaba **no
  expone los votos ocultos** (se mantiene US-16 RB-08). El historial de rondas
  reveladas siempre se conserva.
- **RB-14**: el historial por historia incluye para cada sesión `estado`,
  `puntos_acordados`, `cerrado_en`, `cancelado_en` y `story_estimate` cuando
  corresponde.
- **RB-15**: cualquier miembro del proyecto puede acordar o cancelar. Proyecto
  inexistente o solicitante no miembro: `404 PROYECTO_NO_ENCONTRADO`. Sesión
  inexistente o de otro proyecto: `404 SESION_POKER_NO_ENCONTRADA`.
- **RB-16**: precedencia de errores: parámetros (`400`) → proyecto/sesión
  inexistente (`404`) → puntos inválidos (`422`) → estado inválido (`409`).

## Diseño técnico

### Transacción compartida (UnitOfWork)

Se agrega en `internal/domain` un puerto de unidad de trabajo:

```go
// RepositoriosTransaccion expone los repositorios ligados a una transacción.
type RepositoriosTransaccion interface {
    Historias() HistoriaRepository
    Poker() SesionPokerRepository
    Estimaciones() StoryEstimateRepository
}

// UnidadDeTrabajo ejecuta fn dentro de una transacción. Si fn devuelve error
// (o entra en pánico) se hace rollback; si no, commit.
type UnidadDeTrabajo interface {
    Ejecutar(ctx context.Context, fn func(repos RepositoriosTransaccion) error) error
}
```

`internal/repository` la implementa con `gorm.DB.Transaction`, construyendo los
repositorios existentes sobre el `tx`. El service orquesta el acuerdo dentro de
`Ejecutar`, de modo que la regla vive en el service y la atomicidad en el
repositorio. Los tests de service usan un fake de `UnidadDeTrabajo` que invoca
`fn` con mocks; el test de rollback del repositorio fuerza un fallo al crear el
`story_estimate`.

### Cambios de esquema

- `poker_sessions.cancelado_en timestamptz NULL` (no estaba en `tablas.sql`).
- `story_estimates.sesion_id` y `story_estimates.ronda_id` (`NULL`, FK a
  `poker_sessions` y `poker_rounds`). Nulos para estimaciones que no provengan
  de Planning Poker.
- `poker_sessions.puntos_acordados` y `story_estimates.puntos` siguen siendo
  `double precision` para admitir `0.5`.

### Puertos nuevos o ampliados

- `SesionPokerRepository.CerrarSesion(ctx, sesionID, puntos)`: bloquea, valida
  `Revelada`, fija `puntos_acordados`, `Cerrada`, `cerrado_en`; devuelve la
  ronda actual para el `story_estimate`.
- `SesionPokerRepository.CancelarSesion(ctx, sesionID)`: bloquea, valida
  `Votando`/`Revelada`, fija `Cancelada` y `cancelado_en`.
- `StoryEstimateRepository`: `Crear` y consulta por sesión(es).
- Error nuevo `domain.ErrSesionPokerCerrada` → `409 SESION_POKER_CERRADA` en
  `buscarMapeo`.

## Salidas esperadas

### `POST .../poker-sessions/{sesionId}/acuerdo` → `200`

Devuelve el detalle de la sesión cerrada:

```json
{
  "id": 7,
  "historia_id": 5,
  "estado": "Cerrada",
  "ronda_actual": 2,
  "puntos_acordados": 5,
  "cerrado_en": "2026-11-14T18:30:00Z",
  "total_miembros": 3,
  "votos_registrados": 3,
  "ya_vote": true,
  "mi_voto": 5,
  "votos": [
    { "usuario_id": 2, "usuario_nombre": "Ana", "puntos": 5, "ronda": 2 }
  ],
  "dispersion": { "...": "..." },
  "historial_rondas": [ { "numero": 1, "...": "..." }, { "numero": 2, "...": "..." } ],
  "story_estimate": {
    "id": 11,
    "historia_id": 5,
    "sesion_id": 7,
    "ronda_id": 14,
    "puntos": 5,
    "fuente": "planning_poker",
    "estimado_por": 2,
    "estimado_en": "2026-11-14T18:30:00Z"
  }
}
```

### `POST .../poker-sessions/{sesionId}/cancelar` → `200`

Devuelve el detalle de la sesión con `estado = "Cancelada"`, `cancelado_en`
informado, `puntos_acordados = null` y sin `story_estimate`.

### `GET` de detalle e historial → `200`

Mismo contrato de US-14/US-16 más `cancelado_en` y `story_estimate` (omitidos
cuando no corresponden).

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
| Fallo de persistencia (rollback) | `500` | `ERROR_INTERNO` |

## Restricciones y casos límite

- `puntos = 0` y `puntos = 0.5` son válidos; `4`, `-1` y `34` no.
- Acordar puntos distintos de todos los votos es válido (decisión del equipo).
- Cancelar una sesión `Votando` sin votos es válido.
- Tras cancelar, crear una sesión nueva para la misma historia responde `201`.
- Tras cerrar, crear una sesión nueva para la misma historia también es posible
  (US-14 RB-03) y su acuerdo genera otro `story_estimate`.
- Fuera de alcance: frontend, auditoría `cancelado_por`, reabrir sesiones.

## Criterios de aceptación

- [ ] **CA-1**: Given una sesión `Revelada` con votos, When se acuerda un punto
  válido, Then la sesión pasa a `Cerrada`, se guarda `puntos_acordados`, se
  actualiza `historias.story_points` y se crea `story_estimate`.
- [ ] **CA-2**: Given una sesión `Votando`, When se intenta acordar, Then
  `409 SESION_POKER_NO_REVELADA`.
- [ ] **CA-3**: Given una sesión `Cerrada` o `Cancelada`, When se intenta
  acordar, Then `409 SESION_POKER_CERRADA`.
- [ ] **CA-4**: Given un punto fuera de la escala de US-05, When se acuerda,
  Then `422 PUNTOS_POKER_INVALIDOS` y no se actualiza la historia.
- [ ] **CA-5**: Given una falla al cerrar la sesión o al crear `story_estimate`,
  When se acuerda, Then no se actualiza `historias.story_points` y no queda
  ningún cambio parcial.
- [ ] **CA-6**: Given un acuerdo exitoso, When finaliza, Then
  `historias.story_points`, el estado `Cerrada` y `story_estimate` quedan
  confirmados en la misma transacción.
- [ ] **CA-7**: Given una sesión `Votando` o `Revelada`, When se cancela, Then
  pasa a `Cancelada`, no cambia `story_points` y la historia queda libre para
  otra sesión.
- [ ] **CA-8**: Given una sesión `Cerrada` o `Cancelada`, When se intenta
  cancelar, Then `409 SESION_POKER_CERRADA`.
- [ ] **CA-9**: Given una cancelación, When se consulta el historial, Then la
  sesión aparece como `Cancelada` y conserva sus votos/rondas.
- [ ] **CA-10**: Given un no miembro, When acuerda o cancela, Then `404`.

## Definición de terminado

- [ ] `go test ./...` en verde, `gofmt -l` vacío en los archivos tocados y
  `go vet ./...` sin avisos.
- [ ] Spec SDD y escenarios BDD escritos antes del código.
- [ ] Tests de integración con la historia y `story_estimate`.
- [ ] Test de rollback forzando fallo en `story_estimate`.
- [ ] E2E por curl de acuerdo, cierre, cancelación y reapertura de una nueva
  sesión.
