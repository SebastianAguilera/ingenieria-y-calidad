# Especificación: US-F16 - Votación con cartas en Planning Poker

## Objetivo

Que un miembro de una sesión de Planning Poker pueda **registrar y cambiar su
voto** eligiendo una carta de la escala, sin ver nombres ni valores ajenos
mientras la sesión permanece en `Votando`. La pantalla muestra el voto propio y
el conteo `"X de Y votaron"`, y se mantiene actualizada con polling cada 3
segundos.

La historia satisface el issue `SebastianAguilera/ingenieria-y-calidad#82`.
**Requiere** US-15 (backend de voto oculto) y US-F15 (entrada y sesión desde el
backlog). **Bloquea a** US-F17 (revelado y acuerdo en la UI).

**Trazabilidad:** Issue #82 → SDD → criterios de aceptación → escenarios BDD
(`docs/features/US-F16-votacion-planning-poker.feature`) → tests Vitest/RTL
(`frontend/src/api/pokerVotacion.test.ts`,
`frontend/src/pages/proyecto/poker/CartasVoto.test.tsx`,
`frontend/src/pages/proyecto/SesionPokerPage.votacion.test.tsx`) → código React.

## Cross-referencias

- **US-05**: escala válida `0, 0.5, 1, 2, 3, 5, 8, 13, 21`.
- **US-14**: detalle de sesión con `estado`, `ronda_actual`, `mi_voto`,
  `votos_registrados`, `total_miembros`, `ya_vote` y `votos`.
- **US-15**: `POST .../poker-sessions/{sesionId}/votos` con
  `{ "puntos": number }`; upsert por ronda/usuario; `409
  SESION_POKER_NO_VOTANDO`; `422 PUNTOS_POKER_INVALIDOS`.
- **US-F15**: pantalla `SesionPokerPage`, entrada desde el backlog y estados
  `Votando`/`Revelada`/`Cerrada`/`Cancelada`.
- **US-F17**: mostrará los votos ajenos revelados y permitirá acordar puntos.
  US-F16 no renderiza nombres ni valores ajenos, ni siquiera en `Revelada`.
- **US-F19**: polling de la campana de sesiones activas. No se reutiliza ni se
  mezcla con el polling de esta pantalla.

## Entradas

- Ruta existente: `/proyectos/:id/backlog/:historiaId/poker`.
- `GET /api/proyectos/{id}/poker-sessions/{sesionId}` para obtener el detalle de
  la sesión activa.
- `POST /api/proyectos/{id}/poker-sessions/{sesionId}/votos` con
  `{ "puntos": number }` al elegir una carta.
- Identidad: JWT administrado por `api/client`; la pantalla no maneja tokens.

## Salidas esperadas

### Sesión activa `Votando`

- Panel “Sesión en curso” con estado, ronda, creador y fecha de creación.
- Grupo accesible “Tu estimación” con las cartas `0, 0.5, 1, 2, 3, 5, 8, 13,
  21`.
- La carta correspondiente a `mi_voto` aparece seleccionada.
- Texto `Tu voto: N` cuando `mi_voto` no es nulo; en caso contrario,
  `Todavía no votaste`.
- Texto `"X de Y votaron"` usando `votos_registrados` y `total_miembros`.
- Las cartas se pueden cambiar mientras la sesión siga `Votando`.

### Sesión activa `Revelada`

- Las cartas se muestran deshabilitadas.
- Se mantiene el voto propio y el conteo.
- Se informa que los votos fueron revelados y que la sesión sigue activa.
- No se muestran nombres ni valores ajenos. El arreglo `votos` no se renderiza.

### Sesión `Cerrada` o `Cancelada` al abrir la pantalla

- Se mantiene el comportamiento de US-F15: resultado de la última sesión y
  botón para iniciar otra cuando corresponda.
- No se muestran cartas si no hay una sesión activa.

### Sesión que sale de `Votando` con la pantalla abierta

- El polling actualiza el estado.
- Si pasa a `Revelada`, las cartas se deshabilitan y se informa el revelado.
- Si pasa a `Cerrada` o `Cancelada`, las cartas se deshabilitan y se muestra el
  nuevo estado.

### Errores de voto

- `409 SESION_POKER_NO_VOTANDO`: refrescar el detalle y mostrar
  `La votación ya cerró`.
- Error de red: mostrar `No se pudo registrar el voto. Intentá nuevamente.` y
  permitir reintentar.
- `422 PUNTOS_POKER_INVALIDOS`: mostrar un mensaje de valor inválido; las cartas
  no permiten generar valores fuera de escala.
- `404 PROYECTO_NO_ENCONTRADO` / `404 SESION_POKER_NO_ENCONTRADA`: mostrar el
  mensaje de acceso o sesión no encontrada.

### Errores de carga y polling

- Error inicial del detalle: alerta
  `No se pudo cargar el Planning Poker. Intentá nuevamente.` y botón
  `Reintentar`.
- Error de polling: conservar los últimos datos y mostrar el aviso no bloqueante
  `No se pudo actualizar la sesión. Reintentando…`.
- El próximo polling exitoso limpia el aviso.

## Reglas de negocio

- **RB-01**: la escala de cartas es exactamente `0, 0.5, 1, 2, 3, 5, 8, 13,
  21`; no se agregan ni omiten valores.
- **RB-02**: una carta solo se puede seleccionar si el detalle está en
  `Votando` y no hay una petición de voto en curso.
- **RB-03**: elegir una carta envía `POST` con `{ "puntos": number }`. El valor
  seleccionado se confirma visualmente cuando la respuesta es exitosa.
- **RB-04**: si el usuario ya votó, elegir otra carta actualiza el voto; no se
  crea un voto adicional ni se duplica visualmente.
- **RB-05**: mientras el estado sea `Votando`, la UI no muestra nombres,
  identificadores ni puntos ajenos. Solo muestra el conteo y `mi_voto`.
- **RB-06**: `votos_registrados` y `total_miembros` provienen del backend; no se
  calculan en frontend.
- **RB-07**: el polling consulta el detalle cada 3 segundos mientras el estado
  sea `Votando`. Debe evitar solicitudes solapadas y abortar al desmontar o
  cambiar de sesión.
- **RB-08**: al recibir `Revelada`, `Cerrada` o `Cancelada`, el polling se
  detiene y las cartas quedan deshabilitadas.
- **RB-09**: un `409 SESION_POKER_NO_VOTANDO` obliga a refrescar el detalle
  antes de volver a habilitar cualquier interacción.
- **RB-10**: un error de polling no invalida el voto propio ni el último conteo
  conocido; se reintenta automáticamente.
- **RB-11**: una sesión recién creada debe obtener su detalle para mostrar
  `0 de Y votaron` y habilitar las cartas.
- **RB-12**: los controles son accesibles por teclado, tienen foco visible y
  etiquetas como `"5 puntos"`. El conteo y los avisos se anuncian con
  `aria-live`/roles apropiados.

## Condiciones de error

| Situación | HTTP | `error` | Comportamiento de la UI |
| :--- | :---: | :--- | :--- |
| Sesión ya no está `Votando` | `409` | `SESION_POKER_NO_VOTANDO` | Refresca detalle y muestra “La votación ya cerró”. |
| Puntos fuera de escala | `422` | `PUNTOS_POKER_INVALIDOS` | Muestra valor inválido y conserva el voto anterior. |
| Proyecto inexistente o sin acceso | `404` | `PROYECTO_NO_ENCONTRADO` | Mensaje de acceso/proyecto inexistente. |
| Sesión inexistente o de otro proyecto | `404` | `SESION_POKER_NO_ENCONTRADA` | Mensaje de sesión no encontrada. |
| Sin conexión al votar | `0` | `NETWORK_ERROR` | Mensaje de error y reintento habilitado. |
| Sin conexión al cargar detalle | `0` | `NETWORK_ERROR` | Error de carga con `Reintentar`. |
| Sin conexión durante polling | `0` | `NETWORK_ERROR` | Aviso no bloqueante y reintento automático. |
| Token inválido/expirado | `401` | `TOKEN_*` | Comportamiento global de sesión de `api/client`. |

## Casos límite

- `0` y `0.5` son votos válidos y deben poder seleccionarse.
- `mi_voto` puede ser `null` antes del primer voto.
- Cambiar el voto por el mismo valor sigue siendo una operación válida.
- El conteo puede alcanzar `total_miembros`; no debe superarlo por cálculos de
  frontend.
- Una sesión puede pasar de `Votando` a `Revelada`, `Cerrada` o `Cancelada`
  entre dos polls.
- El detalle puede responder con `votos` vacío mientras está `Votando`; la UI
  no depende de ese arreglo.
- Si la sesión tiene más de una ronda, el voto propio mostrado corresponde a la
  ronda actual informada por el backend.
- Si el usuario navega fuera de la pantalla, se cancelan polling y solicitudes
  pendientes.

## Criterios de aceptación

- **CA-1**: Given una sesión `Votando`, When se abre la pantalla, Then se ven
  las 9 cartas, el conteo y el voto propio si existe.
- **CA-2**: Given una sesión `Votando`, When se elige una carta, Then se envía
  el voto y la carta queda seleccionada tras la respuesta exitosa.
- **CA-3**: Given un voto ya registrado, When se elige otra carta, Then el voto
  se actualiza y no se duplica.
- **CA-4**: Given una sesión `Votando`, When se muestra el detalle, Then solo se
  ve el conteo y el voto propio, nunca nombres o valores ajenos.
- **CA-5**: Given una sesión `Votando`, When se muestra el conteo, Then el texto
  es `"X de Y votaron"` usando los valores del backend.
- **CA-6**: Given una sesión `Revelada`, When se abre o transiciona, Then las
  cartas están deshabilitadas y no se renderizan votos ajenos.
- **CA-7**: Given una sesión `Votando`, When pasan 3 segundos, Then el detalle y
  el conteo se actualizan sin recargar la página.
- **CA-8**: Given un voto con `409 SESION_POKER_NO_VOTANDO`, When falla, Then se
  refresca el detalle y se muestra “La votación ya cerró”.
- **CA-9**: Given un error de red al votar, When falla, Then se muestra un
  mensaje y se puede reintentar.
- **CA-10**: Given un error de detalle inicial, When se abre la pantalla, Then
  se muestra el error y `Reintentar` vuelve a consultar.
- **CA-11**: Given un error de polling, When falla, Then se conserva el último
  conteo y se muestra un aviso no bloqueante.
- **CA-12**: Given una sesión que pasa a `Cerrada` o `Cancelada`, When el
  polling la detecta, Then las cartas se deshabilitan y se informa el estado.

## Fuera de alcance

- Mostrar nombres, identificadores o puntos ajenos, incluso en `Revelada`
  (US-F17).
- Revelar votos, calcular dispersión, acordar puntos, cerrar o cancelar sesiones
  desde la UI.
- Nueva ronda desde la UI.
- Contexto compartido de sesiones activas o notificaciones globales (US-F19).
- Cambios en el backend, contrato de US-14/US-15 o esquema de base de datos.

## Definición de terminado

- [ ] Spec SDD y escenarios BDD escritos antes del código.
- [ ] Tests RED de API, `CartasVoto` y `SesionPokerPage` escritos antes de la
  implementación.
- [ ] `npm test`, `npm run lint` y `npm run build` en verde.
- [ ] Verificación manual de voto, cambio de voto, polling, revelado y error de
  red.
- [ ] Accesibilidad revisada por teclado y con etiquetas accesibles.
