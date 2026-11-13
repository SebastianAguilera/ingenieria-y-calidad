# Especificación: US-15 - Votación oculta en Planning Poker

## Objetivo

Que un miembro del proyecto pueda **registrar y modificar su voto** dentro de una
sesión de Planning Poker en estado `Votando`, sin que el resto del equipo vea su
valor antes del revelado. Mientras la sesión sigue en votación, el detalle solo
informa **cuántos votos hay** (`votos_registrados`) y **el voto propio**
(`mi_voto`).

La historia satisface el issue `SebastianAguilera/ingenieria-y-calidad#75`.
**Requiere** US-14 (modelos y rutas de sesión) y US-05 (escala de story points).
**Bloquea a** US-16 (revelar/acordar) y US-F16. No incluye revelar, calcular
dispersión, nuevas rondas, acuerdo ni frontend.

**Trazabilidad:** Issue #75 → SDD → criterios de aceptación → escenarios BDD
(`docs/features/US-15-votacion-planning-poker.feature`) → tests Go
(`internal/domain`, `internal/service`, `internal/handler`,
`internal/repository`, `cmd/api`) → código.

## Cross-referencias

- **US-05**: escala `0, 0.5, 1, 2, 3, 5, 8, 13, 21`, validada con
  `domain.StoryPointsValidos`.
- **US-14**: tablas `poker_sessions`, `poker_rounds` y `poker_votes` (índice
  único `(ronda_id, usuario_id)`), detalle de sesión y estados.
- **US-F10/US-11**: guard `ExigirAccesoAProyecto` e identidad del JWT
  (`UsuarioDesdeContexto`). Cualquier miembro puede votar; un no miembro recibe
  `404`.
- **US-16**: revelará la sesión (`Votando` → `Revelada`). La carrera
  votar/revelar de esta historia debe quedar resuelta antes.

## Entradas

- `POST /api/proyectos/{id}/poker-sessions/{sesionId}/votos` con cuerpo
  `{ "puntos": number }`.
- Identidad del request: claim `sub` del JWT (el votante).
- `:id` y `:sesionId` se parsean como `uint`; un valor no numérico o cero es
  `400 PARAMETRO_INVALIDO`.
- Un cuerpo que no es JSON válido es `400 JSON_INVALIDO`.

## Reglas de negocio

- **RB-01**: el voto se guarda por **ronda actual y usuario**. Votar por primera
  vez inserta; votar de nuevo en la misma ronda **actualiza** el voto existente
  (upsert) sin duplicar filas.
- **RB-02**: `puntos` debe pertenecer a la escala de US-05. Un valor fuera de
  escala, negativo, ausente o no numérico (por ejemplo `"abc"` o `null`) es
  `422 PUNTOS_POKER_INVALIDOS`, con `campo: "puntos"`, y no modifica nada.
- **RB-03**: solo se puede votar si la sesión está en `Votando`. En `Revelada`,
  `Cerrada` o `Cancelada` la respuesta es `409 SESION_POKER_NO_VOTANDO` y no se
  inserta ni actualiza ningún voto.
- **RB-04**: la verificación de estado y la escritura ocurren **en la misma
  transacción**, con la fila de la sesión bloqueada
  (`SELECT ... FOR UPDATE`). No alcanza con leer el estado y luego escribir.
  Si el revelado gana la carrera, el voto tardío no se inserta ni se actualiza y
  recibe `409 SESION_POKER_NO_VOTANDO`.
- **RB-05**: sesión inexistente o de otro proyecto: `404
  SESION_POKER_NO_ENCONTRADA`. Proyecto inexistente: `404
  PROYECTO_NO_ENCONTRADO`.
- **RB-06**: un no miembro recibe `404 PROYECTO_NO_ENCONTRADO` (guard de US-F10).
- **RB-07**: precedencia de errores: proyecto/sesión inexistente (`404`) →
  puntos inválidos (`422`) → sesión no `Votando` (`409`).
- **RB-08**: mientras la sesión está en `Votando`, ni la respuesta del voto ni el
  detalle exponen nombres, votantes ni puntos ajenos. Solo `votos_registrados`
  y, para el propio usuario, `mi_voto` (contrato ya cubierto por US-14, RB-08).
- **RB-09**: el voto es editable mientras la sesión siga `Votando`; no se crea un
  voto nuevo por cada cambio.

## Salidas esperadas

### `POST .../poker-sessions/{sesionId}/votos` → `200`

```json
{
  "sesion_id": 7,
  "mi_voto": 5,
  "votado_en": "2026-11-13T18:00:00Z",
  "votos_registrados": 2,
  "total_miembros": 3
}
```

La respuesta es `200` tanto al registrar como al modificar el voto. Nunca incluye
votos de otros usuarios.

## Condiciones de error

| Situación | Código HTTP | `error` |
| :--- | :--- | :--- |
| JSON malformado | `400` | `JSON_INVALIDO` |
| `id` o `sesionId` inválidos | `400` | `PARAMETRO_INVALIDO` |
| Sin token / token inválido | `401` | `TOKEN_REQUERIDO` / `TOKEN_INVALIDO` |
| No miembro / proyecto inexistente | `404` | `PROYECTO_NO_ENCONTRADO` |
| Sesión inexistente o de otro proyecto | `404` | `SESION_POKER_NO_ENCONTRADA` |
| Puntos fuera de escala, negativos, ausentes o no numéricos | `422` | `PUNTOS_POKER_INVALIDOS` |
| Sesión `Revelada`, `Cerrada` o `Cancelada` | `409` | `SESION_POKER_NO_VOTANDO` |

## Restricciones y casos límite

- `0` y `0.5` son votos válidos (el cero no cuenta como "ausente").
- Dos votos simultáneos del mismo usuario convergen en una única fila (índice
  único `(ronda_id, usuario_id)` + upsert).
- La concurrencia votar/revelar se prueba con dos operaciones simultáneas contra
  PostgreSQL real; como el revelado (US-16) aún no existe, el test lo simula con
  una transición directa de estado bajo el mismo bloqueo de fila.
- No se agregan tablas ni columnas: se reutiliza `poker_votes` de US-14.

## Criterios de aceptación

- [ ] **CA-1**: Given una sesión `Votando` y un miembro, When vota con un valor de
  la escala, Then el voto queda registrado para la ronda actual.
- [ ] **CA-2**: Given un miembro que ya votó, When vuelve a votar mientras sigue
  `Votando`, Then se actualiza su voto y no se duplica.
- [ ] **CA-3**: Given una sesión `Votando`, When otro miembro consulta el detalle,
  Then solo ve `votos_registrados` y su propio voto, nunca los ajenos.
- [ ] **CA-4**: Given una sesión `Revelada`, `Cerrada` o `Cancelada`, When se
  intenta votar, Then `409 SESION_POKER_NO_VOTANDO` y no se modifica nada.
- [ ] **CA-5**: Given un valor fuera de escala, negativo o no numérico, When se
  vota, Then `422 PUNTOS_POKER_INVALIDOS`.
- [ ] **CA-6**: Given un no miembro, When vota, Then `404`.
- [ ] **CA-7**: Given una carrera entre votar y revelar, When el revelado gana,
  Then el voto tardío no se inserta ni actualiza y recibe
  `409 SESION_POKER_NO_VOTANDO`.

## Definición de terminado

- [ ] `go test ./...` en verde, `gofmt -l .` vacío y `go vet ./...` sin avisos.
- [ ] Spec SDD y escenarios BDD escritos antes del código.
- [ ] Test explícito de concurrencia votar/revelar con dos operaciones
  simultáneas.
- [ ] E2E por curl de voto, cambio de voto y voto fuera de estado.
