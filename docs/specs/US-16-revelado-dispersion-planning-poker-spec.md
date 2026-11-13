# Especificación: US-16 - Revelado, dispersión y nuevas rondas

## Objetivo

Que cualquier miembro de un proyecto pueda **revelar los votos** de una sesión de
Planning Poker en estado `Votando`, ver la **dispersión** de la ronda y **abrir
una nueva ronda** para volver a estimar conservando el historial anterior.

La historia satisface el issue `SebastianAguilera/ingenieria-y-calidad#78`.
**Requiere** US-14 (sesiones y rondas) y US-15 (votación oculta).
**Bloquea a** US-17 (acuerdo y cierre) y US-F17 (frontend).

**Trazabilidad:** Issue #78 → SDD → criterios de aceptación → escenarios BDD
(`docs/features/US-16-revelado-dispersion-planning-poker.feature`) → tests Go
(`internal/domain`, `internal/service`, `internal/handler`,
`internal/repository`, `cmd/api`) → código.

## Cross-referencias

- **US-14**: tablas `poker_sessions`, `poker_rounds` y `poker_votes`; estados
  `Votando`, `Revelada`, `Cerrada` y `Cancelada`; detalle de sesión.
- **US-15**: voto por ronda y usuario con upsert; solo se vota en `Votando`.
- **US-05**: escala de puntos `0, 0.5, 1, 2, 3, 5, 8, 13, 21` en `float64`.
- **US-F10/US-11**: guard de membresía; cualquier miembro puede revelar y abrir
  una nueva ronda. Un no miembro recibe `404`.
- **US-17**: acordar puntos, cerrar y cancelar. Fuera de alcance en US-16.

## Entradas

- `POST /api/proyectos/{id}/poker-sessions/{sesionId}/revelar`, sin cuerpo.
- `POST /api/proyectos/{id}/poker-sessions/{sesionId}/rondas`, sin cuerpo.
- `GET /api/proyectos/{id}/poker-sessions/{sesionId}`.
- Identidad del request: claim `sub` del JWT.
- `:id` y `:sesionId` se parsean como `uint`; un valor no numérico o cero es
  `400 PARAMETRO_INVALIDO`.

## Reglas de negocio

- **RB-01**: solo se puede revelar desde `Votando`. En otro estado la respuesta
  es `409 SESION_POKER_NO_VOTANDO` y no se modifica nada.
- **RB-02**: no se puede revelar una ronda sin votos:
  `422 SESION_POKER_SIN_VOTOS`, sin modificar nada.
- **RB-03**: revelar pasa la sesión a `Revelada`, marca `poker_rounds.revelada_en`
  de la ronda actual y actualiza `poker_sessions.actualizado_en`. No cierra la
  sesión ni fija `puntos_acordados`.
- **RB-04**: solo se puede abrir una nueva ronda desde `Revelada`. En otro estado
  la respuesta es `409 SESION_POKER_NO_REVELADA` y no se modifica nada.
- **RB-05**: abrir una nueva ronda incrementa `ronda_actual`, crea una fila nueva
  en `poker_rounds` con el número siguiente y devuelve la sesión a `Votando`.
  Los votos de rondas anteriores **no se borran**.
- **RB-06**: la verificación de estado y la escritura de revelar/abrir ronda
  ocurren en una transacción con la fila de la sesión bloqueada
  (`SELECT ... FOR UPDATE`), igual que el voto de US-15.
- **RB-07**: el detalle en `Revelada` o `Cerrada` devuelve los votos de la ronda
  actual con `usuario_id`, `usuario_nombre`, `puntos` y `ronda`.
- **RB-08**: en `Votando` el detalle no expone nombres, puntos ajenos, votos ni
  estadísticas de dispersión. Solo `votos_registrados`, `ya_vote` y `mi_voto`.
- **RB-09**: la dispersión se calcula sobre todos los votos de la ronda revelada:
  `promedio`, `mediana`, `minimo`, `maximo`, `rango`, `desviacion_estandar`,
  `hay_dispersion`, `total_votos` y `votos_por_valor`.
- **RB-10**: `desviacion_estandar` es **poblacional**:
  `sqrt(sum((x - promedio)^2) / n)`.
- **RB-11**: `hay_dispersion` es `true` si `maximo != minimo` y `false` si todos
  los votos son iguales. No es un umbral configurable.
- **RB-12**: `votos_por_valor` es un objeto `{"<punto>": cantidad}`. Las claves
  usan la representación decimal mínima (`"0.5"`, `"1"`, `"3"`, `"8"`).
- **RB-13**: el detalle incluye `historial_rondas` con todas las rondas de la
  sesión ordenadas por número ascendente. Cada ítem informa `numero`, `estado`,
  `revelada_en`, `votos` y `dispersion`. Una ronda no revelada devuelve `votos`
  vacíos y `dispersion` ausente.
- **RB-14**: el detalle mantiene `puntos_acordados` si existe.
- **RB-15**: sesión inexistente o de otro proyecto: `404
  SESION_POKER_NO_ENCONTRADA`. Proyecto inexistente o solicitante no miembro:
  `404 PROYECTO_NO_ENCONTRADO`.
- **RB-16**: precedencia de errores: proyecto/sesión inexistente (`404`) →
  sin votos (`422`) → estado inválido (`409`).

## Salidas esperadas

### `POST .../poker-sessions/{sesionId}/revelar` → `200`

Devuelve el detalle de la sesión ya revelada, con los votos y la dispersión:

```json
{
  "id": 7,
  "historia_id": 5,
  "estado": "Revelada",
  "ronda_actual": 1,
  "puntos_acordados": null,
  "total_miembros": 3,
  "votos_registrados": 3,
  "ya_vote": true,
  "mi_voto": 1,
  "votos": [
    { "usuario_id": 2, "usuario_nombre": "Ana", "puntos": 1, "ronda": 1 },
    { "usuario_id": 7, "usuario_nombre": "Ben", "puntos": 3, "ronda": 1 },
    { "usuario_id": 9, "usuario_nombre": "Cris", "puntos": 8, "ronda": 1 }
  ],
  "dispersion": {
    "promedio": 4,
    "mediana": 3,
    "minimo": 1,
    "maximo": 8,
    "rango": 7,
    "desviacion_estandar": 2.9439202887759492,
    "hay_dispersion": true,
    "total_votos": 3,
    "votos_por_valor": { "1": 1, "3": 1, "8": 1 }
  },
  "historial_rondas": [
    {
      "numero": 1,
      "estado": "Revelada",
      "revelada_en": "2026-11-13T18:00:00Z",
      "votos": [
        { "usuario_id": 2, "usuario_nombre": "Ana", "puntos": 1, "ronda": 1 }
      ],
      "dispersion": { "...": "..." }
    }
  ]
}
```

### `POST .../poker-sessions/{sesionId}/rondas` → `200`

Devuelve el detalle de la sesión en `Votando` con `ronda_actual` incrementado. El
historial conserva la ronda revelada anterior con sus votos y estadísticas.

### `GET .../poker-sessions/{sesionId}` → `200`

Mismo contrato que el detalle anterior. En `Votando`, `votos` queda vacío y
`dispersion`/estadísticas no se exponen; `historial_rondas` solo muestra votos de
rondas ya reveladas.

## Condiciones de error

| Situación | Código HTTP | `error` |
| :--- | :--- | :--- |
| `id` o `sesionId` inválidos | `400` | `PARAMETRO_INVALIDO` |
| Sin token / token inválido | `401` | `TOKEN_REQUERIDO` / `TOKEN_INVALIDO` |
| No miembro / proyecto inexistente | `404` | `PROYECTO_NO_ENCONTRADO` |
| Sesión inexistente o de otro proyecto | `404` | `SESION_POKER_NO_ENCONTRADA` |
| Revelar una ronda sin votos | `422` | `SESION_POKER_SIN_VOTOS` |
| Revelar fuera de `Votando` | `409` | `SESION_POKER_NO_VOTANDO` |
| Abrir ronda fuera de `Revelada` | `409` | `SESION_POKER_NO_REVELADA` |

## Restricciones y casos límite

- Un voto único produce `rango = 0`, `desviacion_estandar = 0` y
  `hay_dispersion = false`.
- Votos iguales (`3`, `3`, `3`) producen `hay_dispersion = false`.
- Votos distintos (`1`, `2`, `5`) producen `hay_dispersion = true`.
- Con cantidad par de votos, la mediana es el promedio de los dos centrales.
- No se agregan tablas ni columnas: se reutilizan `poker_rounds.revelada_en`,
  `poker_sessions.ronda_actual` y `poker_sessions.estado`.
- El revelado no cierra la sesión; la nueva ronda no borra votos anteriores.
- Fuera de alcance: acordar puntos, cierre, cancelación, notificaciones y
  frontend.

## Criterios de aceptación

- [ ] **CA-1**: Given una sesión `Votando` con al menos un voto, When se revela,
  Then pasa a `Revelada` y todos ven nombre y puntos.
- [ ] **CA-2**: Given una sesión `Votando` sin votos, When se intenta revelar,
  Then `422 SESION_POKER_SIN_VOTOS`.
- [ ] **CA-3**: Given una sesión `Revelada`, `Cerrada` o `Cancelada`, When se
  intenta revelar, Then `409 SESION_POKER_NO_VOTANDO`.
- [ ] **CA-4**: Given una sesión `Revelada`, When se abre una nueva ronda, Then
  `ronda_actual` aumenta y la sesión vuelve a `Votando`.
- [ ] **CA-5**: Given una sesión `Votando`, `Cerrada` o `Cancelada`, When se
  intenta abrir una nueva ronda, Then `409 SESION_POKER_NO_REVELADA`.
- [ ] **CA-6**: Given votos `1`, `3` y `8`, When se revela, Then `promedio = 4`,
  `mediana = 3`, `minimo = 1`, `maximo = 8`, `rango = 7`,
  `desviacion_estandar` es la poblacional y `hay_dispersion = true`.
- [ ] **CA-7**: Given votos `3`, `3` y `3`, When se revela, Then
  `hay_dispersion = false`.
- [ ] **CA-8**: Given votos `1`, `2` y `5`, When se revela, Then
  `hay_dispersion = true`.
- [ ] **CA-9**: Given una sesión `Votando`, When se consulta el detalle, Then no
  se exponen nombres ni puntos ajenos.
- [ ] **CA-10**: Given una nueva ronda, When se consulta el historial, Then la
  ronda anterior sigue visible con sus votos y estadísticas.

## Definición de terminado

- [ ] `go test ./...` en verde, `gofmt -l .` vacío y `go vet ./...` sin avisos.
- [ ] Spec SDD y escenarios BDD escritos antes del código.
- [ ] Tests de dispersión con casos de un voto, votos iguales y votos distintos.
- [ ] Tests de transiciones, permisos y no filtración antes del revelado.
- [ ] E2E por curl de revelar, nueva ronda e historial.
