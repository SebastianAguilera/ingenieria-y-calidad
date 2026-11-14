# Especificación: US-F15 - Entrada y sesión de Planning Poker en el backlog

## Objetivo

Que un miembro de un proyecto pueda **entrar al Planning Poker desde una
historia del backlog**, iniciar una sesión o unirse a la que ya está activa, sin
crear sesiones de forma implícita. La pantalla muestra el contexto de la
historia, el estado de la última sesión y el resultado cuando ya terminó.

La historia satisface el issue `SebastianAguilera/ingenieria-y-calidad#81`.
**Requiere** US-14 (sesiones y endpoints), US-F07 (backlog tabla/tablero) y
US-F11 (navegación del proyecto). **Bloquea a** US-F16 (votación) y US-F19
(contexto compartido de sesiones activas).

**Trazabilidad:** Issue #81 → SDD → criterios de aceptación → escenarios BDD
(`docs/features/US-F15-entrada-planning-poker.feature`) → tests Vitest/RTL
(`frontend/src/api/poker.test.ts`,
`frontend/src/pages/proyecto/SesionPokerPage.test.tsx`,
`frontend/src/pages/proyecto/BacklogPage.test.tsx`) → código React.

## Cross-referencias

- **US-14**: `GET .../historias/{historiaId}/poker-sessions`,
  `POST .../historias/{historiaId}/poker-sessions` y
  `GET .../poker-sessions/activas`; estados `Votando`, `Revelada`, `Cerrada`,
  `Cancelada`.
- **US-15**: la votación se implementa en US-F16 sobre la misma pantalla; F15
  solo muestra el estado de la sesión activa.
- **US-F07**: tabla y tablero del backlog con acciones por historia.
- **US-F11**: rutas anidadas bajo `/proyectos/:id`.
- **US-F19**: reemplazará la consulta de activas del backlog por un contexto
  compartido; F15 no implementa polling ni WebSocket.

## Entradas

- Ruta de entrada:
  `/proyectos/:id/backlog/:historiaId/poker`.
- `GET /api/proyectos/{id}/historias` para resolver la historia por `id`.
- `GET /api/proyectos/{id}/historias/{historiaId}/poker-sessions` para el
  historial (descendente por creación).
- `POST /api/proyectos/{id}/historias/{historiaId}/poker-sessions` sin cuerpo,
  solo al hacer clic explícito en “Iniciar Planning Poker”.
- `GET /api/proyectos/{id}/poker-sessions/activas` para el distintivo del
  backlog.
- Identidad: JWT administrado por `api/client`; la pantalla no maneja tokens.

## Salidas esperadas

### Pantalla `SesionPokerPage`

- **Cargando**: mensaje `Cargando Planning Poker…` con `role="status"`.
- **Historia sin sesión activa**: código y título de la historia, estado de la
  historia, botón `Iniciar Planning Poker` y, si existe, el resultado de la
  última sesión.
- **Última sesión `Cerrada`**: estado, ronda, fecha de cierre y
  `puntos_acordados` (o “Sin puntos acordados” si es `null`).
- **Última sesión `Cancelada`**: aviso de cancelación y botón para iniciar otra.
- **Sin sesiones**: aviso de que la historia todavía no tiene sesiones y botón
  para iniciar la primera.
- **Sesión activa `Votando`**: panel “Sesión en curso”, estado, ronda, creador y
  fecha de creación. No se muestra el botón de iniciar ni se crea otra sesión.
- **Sesión activa `Revelada`**: mismo panel, indicando que los votos fueron
  revelados y que la sesión sigue activa.
- **Historia inexistente o de otro proyecto**: mensaje de historia no
  encontrada y enlace para volver al backlog.
- **Proyecto inexistente o sin acceso**: mensaje
  `No tenés acceso a este proyecto o no existe`.
- **Error de red/carga**: alerta y botón `Reintentar`.

### Backlog

- Acción `Planning Poker` por historia en tabla y tablero; navega a
  `/proyectos/:id/backlog/:historiaId/poker`.
- Distintivo `Planning Poker en curso` en las historias con una sesión
  `Votando` o `Revelada`.
- Si la consulta de activas falla, el backlog sigue funcionando sin distintivos.

## Reglas de negocio

- **RB-01**: la pantalla nunca crea una sesión al montarse; la creación ocurre
  solo con clic explícito en “Iniciar Planning Poker”.
- **RB-02**: si el historial contiene una sesión `Votando` o `Revelada`, la
  pantalla entra directamente a esa sesión y no ofrece iniciar otra.
- **RB-03**: una sesión `Cerrada` o `Cancelada` no bloquea el inicio de una
  nueva; la pantalla muestra su resultado y habilita el botón.
- **RB-04**: si `POST` responde `409 HISTORIA_NO_ESTIMABLE`, se muestra el
  mensaje de que la historia está en `Hecho` y no se inicia ninguna sesión.
- **RB-05**: si `POST` responde `409 SESION_POKER_ACTIVA` (carrera entre
  usuarios), la pantalla vuelve a consultar el historial y se une a la sesión
  activa sin crear otra.
- **RB-06**: un `404 PROYECTO_NO_ENCONTRADO` se muestra como falta de acceso o
  proyecto inexistente; un `404 HISTORIA_NO_ENCONTRADA` como historia
  inexistente o de otro proyecto.
- **RB-07**: un error de red (`status 0`) muestra un mensaje genérico y permite
  reintentar la carga.
- **RB-08**: el distintivo del backlog se calcula con la lista de activas y se
  refresca al volver a montar el backlog (navegación de regreso desde la
  sesión).
- **RB-09**: el distintivo es informativo: no deshabilita la acción `Planning
  Poker`; al entrar, la pantalla se une a la sesión existente.
- **RB-10**: la pantalla es accesible por teclado: controles nativos, nombres
  accesibles en las acciones, encabezados jerárquicos y estados anunciados con
  `role="status"`/`role="alert"`.

## Casos límite

- Historia con varias sesiones cerradas: se muestra la más reciente por ser la
  primera del historial.
- Sesión activa `Revelada`: sigue contando como activa y no se puede iniciar
  otra.
- `puntos_acordados` nulo en una sesión cerrada: se informa “Sin puntos
  acordados”.
- `:id` o `:historiaId` no numéricos o menores a 1: se muestra el estado de
  historia no encontrada, sin llamar a endpoints inválidos.
- La lista de activas puede cambiar mientras el usuario está en la pantalla;
  un `409 SESION_POKER_ACTIVA` al iniciar se resuelve releyendo el historial.
- El historial puede llegar vacío (`[]`): se muestra el estado sin sesiones.

## Condiciones de error

| Situación | HTTP | `error` | Comportamiento de la UI |
| :--- | :---: | :--- | :--- |
| Proyecto inexistente o sin acceso | `404` | `PROYECTO_NO_ENCONTRADO` | Mensaje de acceso y enlace al backlog |
| Historia inexistente o de otro proyecto | `404` | `HISTORIA_NO_ENCONTRADA` | Mensaje de historia no encontrada |
| Historia en estado `Hecho` | `409` | `HISTORIA_NO_ESTIMABLE` | Alerta y no se crea la sesión |
| Ya existe una sesión activa | `409` | `SESION_POKER_ACTIVA` | Relee el historial y se une |
| Sin conexión o fallo de red | `0` | `NETWORK_ERROR` | Alerta y `Reintentar` |
| Cualquier otro error | varios | varios | Mensaje del backend o genérico |

## Criterios de aceptación

- **CA-1**: Given una historia sin sesión activa, When se abre la ruta, Then se
  ven el resumen de la historia y el botón “Iniciar Planning Poker”.
- **CA-2**: Given el botón “Iniciar Planning Poker”, When se hace clic, Then se
  envía el `POST`, se crea la sesión y se muestra el estado `Votando`.
- **CA-3**: Given una historia en estado `Hecho`, When se intenta iniciar, Then
  se muestra `HISTORIA_NO_ESTIMABLE` y no se muestra una sesión activa.
- **CA-4**: Given una historia con sesión `Votando` o `Revelada`, When se entra
  desde el backlog, Then se muestra esa sesión y no se envía el `POST`.
- **CA-5**: Given una historia con última sesión `Cerrada`, When se abre, Then
  se muestran los puntos acordados y el resultado.
- **CA-6**: Given una historia con última sesión `Cancelada`, When se abre,
  Then se muestra que fue cancelada y se permite iniciar otra.
- **CA-7**: Given una historia con sesión activa, When se lista el backlog,
  Then aparece “Planning Poker en curso”.
- **CA-8**: Given una historia inexistente o de otro proyecto, When se abre,
  Then se muestra el mensaje de `404`.
- **CA-9**: Given un error de red, When se carga o se inicia, Then se muestra el
  mensaje y se puede reintentar.
- **CA-10**: Given el backlog, When se hace clic en “Planning Poker”, Then se
  navega a `/proyectos/:id/backlog/:historiaId/poker`.
- **CA-11**: Given un `409 SESION_POKER_ACTIVA` al iniciar, When se relee el
  historial, Then la pantalla se une a la sesión activa sin crear otra.

## Fuera de alcance

- Votar, revelar votos, calcular dispersión, acordar puntos, cerrar o cancelar
  sesiones (US-F16 y siguientes).
- Contexto compartido, polling o notificaciones en tiempo real (US-F19).
- Cambios en el backend o en el contrato de US-14.
- Nuevas rutas distintas de `/proyectos/:id/backlog/:historiaId/poker`.

## Definición de terminado

- [ ] Spec SDD y escenarios BDD escritos antes del código.
- [ ] `npm test`, `npm run lint` y `npm run build` en verde.
- [ ] Verificación manual con sesión activa, sin sesión, cerrada y cancelada.
- [ ] Accesibilidad revisada por teclado y con nombres accesibles.
