# Especificación: US-F17 - Revelado y dispersión en Planning Poker

## Historia de usuario

**Como** miembro de una sesión de Planning Poker, **quiero** revelar los votos y
ver la dispersión, **para** discutir las diferencias del equipo.

## Descripción

US-F17 implementa el revelado de votos, la visualización de votos con nombre y el
panel de dispersión. También permite abrir una nueva ronda desde una sesión
`Revelada`, conservando el historial anterior.

## Objetivo

Que cualquier miembro de un proyecto pueda revelar los votos de una sesión en
estado `Votando`, ver quién votó qué y analizar la dispersión del equipo. Desde
`Revelada`, podrá abrir una nueva ronda para seguir estimando sin perder los
votos ni las estadísticas de las rondas anteriores.

La historia satisface el issue `SebastianAguilera/ingenieria-y-calidad#83`.
**Requiere** US-16 (backend de revelado, dispersión, historial y nuevas rondas) y
US-F16 (cartas, voto y polling). **Bloquea a** US-F18 (acuerdo, cierre y
cancelación desde la UI).

**Trazabilidad:** Issue #83 → SDD → criterios de aceptación → escenarios BDD
(`docs/features/US-F17-revelado-dispersion-planning-poker.feature`) → tests
Vitest/RTL → código React.

## Alcance

- Botón `Revelar votos` visible solo en `Votando`; llama a
  `POST /api/proyectos/{id}/poker-sessions/{sesionId}/revelar`.
- Si no hay votos, mostrar el error `422 SESION_POKER_SIN_VOTOS`.
- Al revelar, mostrar todos los votos con `usuario_nombre` y `puntos`.
- Crear `PanelDispersion` con `promedio`, `mediana`, `minimo`, `maximo`, `rango`,
  `desviacion_estandar`, `total_votos` y `votos_por_valor`.
- Mostrar el badge `Consenso` cuando `hay_dispersion = false` y
  `Hay diferencias` cuando `hay_dispersion = true`, según lo calculado por el
  backend.
- Resaltar el mínimo y el máximo entre los votos revelados.
- Botón `Nueva ronda` visible en `Revelada`; llama a
  `POST /api/proyectos/{id}/poker-sessions/{sesionId}/rondas`.
- Al abrir una nueva ronda, limpiar la selección local, volver a `Votando` y
  mostrar la nueva ronda.
- Crear `HistorialRondas` con número, votos y estadísticas de rondas anteriores.
- Polling cada 3 segundos del detalle mientras la sesión esté activa.
- Estados de carga, error, sesión cerrada/cancelada y error de red.
- Tests Vitest/RTL de revelado, panel, badge, nueva ronda e historial.
- Accesibilidad: botones etiquetados, panel con texto alternativo y navegación
  por teclado.

## Fuera de alcance

- Acordar puntos, cerrar o cancelar una sesión desde la UI. Corresponde a US-F18.
- Modificar el cálculo de dispersión o los contratos del backend de US-16.
- Gráficos, exportaciones, comentarios por voto o animaciones.
- Notificaciones globales de sesiones activas (US-F19).

## Entradas

- Ruta existente: `/proyectos/:id/backlog/:historiaId/poker`.
- `GET /api/proyectos/{id}/poker-sessions/{sesionId}` para el detalle.
- `POST /api/proyectos/{id}/poker-sessions/{sesionId}/revelar`, sin cuerpo.
- `POST /api/proyectos/{id}/poker-sessions/{sesionId}/rondas`, sin cuerpo.
- JWT administrado por `api/client`; la pantalla no maneja tokens.
- Funciones de API previstas:
  - `revelarSesionPoker(proyectoId, sesionId): Promise<DetalleSesionPoker>`.
  - `abrirNuevaRondaSesionPoker(proyectoId, sesionId): Promise<DetalleSesionPoker>`.

### Contrato del detalle

`DetalleSesionPoker` incluye, entre otros campos:

- `id`, `historia_id`, `estado`, `ronda_actual`.
- `total_miembros`, `votos_registrados`, `ya_vote`, `mi_voto`.
- `votos`: `{ usuario_id, usuario_nombre, puntos, ronda }[]`.
- `dispersion`: `{ promedio, mediana, minimo, maximo, rango,
  desviacion_estandar, hay_dispersion, total_votos, votos_por_valor }`.
- `historial_rondas`: `{ numero, estado, revelada_en, votos, dispersion? }[]`.

## Salidas esperadas

### Sesión `Votando`

- Se mantiene el comportamiento de US-F16: cartas habilitadas, voto propio,
  conteo `"X de Y votaron"` y votos ajenos ocultos.
- Se muestra el botón `Revelar votos`, visible únicamente en este estado.
- El botón no se deshabilita por tener cero votos: el backend responde
  `422 SESION_POKER_SIN_VOTOS` y la UI muestra el error.
- Mientras la petición está en curso, el botón queda deshabilitado y con
  `aria-busy="true"`.

### Sesión `Revelada`

- Las cartas siguen visibles pero deshabilitadas.
- Se listan todos los votos de la ronda con nombre y puntos, por ejemplo
  `Ana: 1 punto`, `Ben: 3 puntos`, `Cris: 8 puntos`.
- Los votos que coinciden con `dispersion.minimo` llevan la marca `Mínimo`; los
  que coinciden con `dispersion.maximo` llevan `Máximo`. Si
  `minimo === maximo`, se muestra una única marca `Mínimo y máximo`.
- Se muestra `PanelDispersion` con todas las estadísticas y la distribución.
- El badge indica `Consenso` o `Hay diferencias` según `hay_dispersion`.
- Se muestra el botón `Nueva ronda`, visible únicamente en este estado.
- No se muestran botones de acuerdo, cierre o cancelación (US-F18).

### Panel de dispersión

- Región accesible con nombre `Dispersión de votos`.
- Estadísticas visibles con etiquetas y valores:
  - `Promedio: N`
  - `Mediana: N`
  - `Mínimo: N`
  - `Máximo: N`
  - `Rango: N`
  - `Desviación estándar: N`
  - `Total de votos: N`
- Distribución visible bajo el título `Distribución de votos`, por ejemplo
  `1 punto: 1 voto`, `3 puntos: 1 voto`, `8 puntos: 1 voto`.
- La distribución se ordena numéricamente por valor.
- Los números se formatean con coma decimal y hasta 2 decimales. Una desviación
  de `2.9439202887759492` se muestra como `2,94`.

### Nueva ronda

- Al confirmar `Nueva ronda`, se llama a `POST .../rondas`.
- La respuesta reemplaza el detalle actual: estado `Votando`, `ronda_actual`
  incrementado, `mi_voto` nulo y votos de la ronda nueva vacíos.
- Se limpian la selección local, los errores y los avisos de revelado/ronda.
- Se vuelve a mostrar el conteo de la nueva ronda y las cartas habilitadas.
- El botón `Nueva ronda` desaparece y reaparece `Revelar votos`.
- La ronda anterior permanece visible en `HistorialRondas`.

### Historial de rondas

- Región accesible con nombre `Historial de rondas`.
- Solo se muestran rondas anteriores: `numero < ronda_actual`.
- La ronda actual no se duplica en el historial.
- Si no hay rondas anteriores, la sección no se renderiza.
- Cada ronda muestra:
  - Encabezado `Ronda N`.
  - Fecha `Revelada el dd/mm/aaaa` cuando `revelada_en` existe.
  - Votos revelados con nombre y puntos.
  - Estadísticas de dispersión de esa ronda.
- La ronda anterior conserva sus votos y estadísticas después de abrir una nueva
  ronda.

### Sesión `Cerrada` o `Cancelada`

- Se mantiene el resultado de US-F15 cuando la última sesión ya terminó.
- No se muestran `Revelar votos`, `Nueva ronda`, cartas ni acciones de US-F18.
- Si el polling detecta el paso a `Cerrada` o `Cancelada`, se detiene y se
  deshabilitan las acciones.

## Reglas de negocio

- **RB-01**: `Revelar votos` solo se muestra y se ejecuta si el detalle está en
  `Votando`.
- **RB-02**: `Nueva ronda` solo se muestra y se ejecuta si el detalle está en
  `Revelada`.
- **RB-03**: en `Votando` no se renderizan nombres, puntos ajenos, dispersión ni
  historial de la ronda actual.
- **RB-04**: revelar reemplaza el detalle con la respuesta del backend. El
  frontend no calcula votos ni estadísticas.
- **RB-05**: `hay_dispersion` se muestra tal como lo devuelve el backend, que lo
  define como `maximo != minimo`.
- **RB-06**: la desviación estándar es poblacional y se muestra tal como la
  devuelve el backend; el frontend solo la formatea.
- **RB-07**: las marcas `Mínimo`, `Máximo` o `Mínimo y máximo` se derivan por
  igualdad con los valores del backend y no reemplazan el texto del voto.
- **RB-08**: abrir una nueva ronda reemplaza el detalle y limpia cualquier
  selección, error o aviso local asociado a la ronda anterior.
- **RB-09**: el historial solo incluye rondas con `numero < ronda_actual`.
- **RB-10**: el polling consulta el detalle cada 3 segundos mientras el estado
  sea `Votando` o `Revelada`. Se detiene en `Cerrada` o `Cancelada`, evita
  solicitudes solapadas y aborta al desmontar o cambiar de sesión.
- **RB-11**: el polling en `Revelada` permite detectar que otro miembro abrió una
  nueva ronda o que la sesión fue cerrada/cancelada.
- **RB-12**: un error de polling conserva los últimos datos y muestra el aviso no
  bloqueante `No se pudo actualizar la sesión. Reintentando…`; el próximo
  polling exitoso lo limpia.
- **RB-13**: los botones son nativos, tienen nombres accesibles, foco visible y
  evitan doble envío mientras la petición está en curso.
- **RB-14**: el panel y el historial usan regiones, encabezados, listas y texto
  visible; la información no depende solo del color.
- **RB-15**: una sesión `Cerrada` o `Cancelada` nunca ofrece revelar ni abrir
  ronda.

## Condiciones de error

| Situación | HTTP | `error` | Comportamiento de la UI |
| :--- | :---: | :--- | :--- |
| Revelar sin votos | `422` | `SESION_POKER_SIN_VOTOS` | Muestra `No se puede revelar sin votos.` y conserva `Votando`. |
| Revelar fuera de `Votando` | `409` | `SESION_POKER_NO_VOTANDO` | Refresca el detalle y muestra `La votación ya cerró`. |
| Abrir ronda fuera de `Revelada` | `409` | `SESION_POKER_NO_REVELADA` | Refresca el detalle y muestra `La sesión ya no está revelada.` |
| Proyecto inexistente o sin acceso | `404` | `PROYECTO_NO_ENCONTRADO` | Mensaje de acceso/proyecto inexistente. |
| Sesión inexistente o de otro proyecto | `404` | `SESION_POKER_NO_ENCONTRADA` | Mensaje de sesión no encontrada. |
| Error de red al revelar | `0` | `NETWORK_ERROR` | Muestra `No se pudieron revelar los votos. Intentá nuevamente.` y permite reintentar. |
| Error de red al abrir ronda | `0` | `NETWORK_ERROR` | Muestra `No se pudo abrir la nueva ronda. Intentá nuevamente.` y permite reintentar. |
| Error de carga inicial | `0` | `NETWORK_ERROR` | Mantiene el error de US-F16 con `Reintentar`. |
| Error de polling | `0` | `NETWORK_ERROR` | Aviso no bloqueante y reintento automático. |
| Token inválido/expirado | `401` | `TOKEN_*` | Comportamiento global de sesión de `api/client`. |

## Casos límite

- Una sesión con un único voto produce `hay_dispersion = false`.
- Votos `3`, `3`, `3` producen `Consenso`.
- Votos `1`, `3`, `8` producen `Hay diferencias`.
- Con cantidad par de votos, la mediana es el promedio de los dos centrales.
- Varios votos pueden compartir el mínimo o el máximo; todos reciben su marca.
- Si `minimo === maximo`, se usa `Mínimo y máximo`.
- Los valores `0` y `0.5` son válidos y deben formatearse correctamente.
- La distribución puede contener valores repetidos, por ejemplo `3: 2`.
- La ronda actual puede estar `Revelada` sin que exista una ronda anterior.
- Otro miembro puede revelar o abrir una ronda mientras el polling está activo.
- El usuario puede navegar fuera de la pantalla con polling o una acción en
  curso; las solicitudes deben abortarse o ignorar su resultado.

## Criterios de aceptación

- [ ] **CA-1**: Given una sesión `Votando` con votos, When se revela, Then se
  muestran nombres y puntos de todos.
- [ ] **CA-2**: Given una sesión `Votando` sin votos, When se intenta revelar,
  Then se muestra `422 SESION_POKER_SIN_VOTOS`.
- [ ] **CA-3**: Given una sesión `Revelada`, When se muestra el panel, Then
  aparecen promedio, mediana, mínimo, máximo, rango, desviación estándar y
  distribución.
- [ ] **CA-4**: Given votos `1`, `3` y `8`, When se revela, Then el panel muestra
  los valores esperados, resalta `1` y `8` y el badge dice `Hay diferencias`.
- [ ] **CA-5**: Given votos `3`, `3` y `3`, When se revela, Then el badge dice
  `Consenso`.
- [ ] **CA-6**: Given una sesión `Revelada`, When se abre una nueva ronda, Then
  la sesión vuelve a `Votando`, el número de ronda aumenta y la selección local
  se limpia.
- [ ] **CA-7**: Given una nueva ronda, When se consulta el historial, Then la
  ronda anterior sigue visible.
- [ ] **CA-8**: Given una sesión `Cerrada` o `Cancelada`, When se abre, Then no
  se puede revelar ni abrir ronda.
- [ ] **CA-9**: Given un error de red, When se revela o se abre ronda, Then se
  muestra el mensaje y se puede reintentar.
- [ ] **CA-10**: Given una sesión `Votando` o `Revelada`, When pasan 3 segundos,
  Then el detalle se actualiza y el polling se detiene en `Cerrada`/`Cancelada`.
- [ ] **CA-11**: Given una sesión `Revelada`, When otro miembro abre una nueva
  ronda, Then el polling vuelve a `Votando`, limpia la selección y habilita las
  cartas.
- [ ] **CA-12**: Given los botones, el panel y el historial, When se navega con
  teclado o lector de pantalla, Then hay nombres accesibles, regiones y texto
  alternativo suficiente.

## Tareas de implementación

- [ ] Escribir spec SDD y escenarios BDD de US-F17.
- [ ] RED: tests de revelado, `PanelDispersion`, badge, `HistorialRondas` y nueva
  ronda.
- [ ] GREEN: botones, visualización de votos, panel e historial.
- [ ] Integrar polling de 3 segundos en `Votando` y `Revelada`.
- [ ] Verificación manual y tests automatizados.

## Definición de terminado

- [ ] Spec SDD y escenarios BDD escritos antes del código.
- [ ] Tests RED de API, `PanelDispersion`, `HistorialRondas` y
  `SesionPokerPage` escritos antes de la implementación.
- [ ] `npm test`, `npm run lint` y `npm run build` en verde.
- [ ] Verificación manual de revelado, panel, badge, nueva ronda e historial.
- [ ] Accesibilidad revisada por teclado, regiones y etiquetas.

## Dependencias

- **Requiere US-16** (bloqueante): revelado, dispersión, nuevas rondas e
  historial en backend.
- **Requiere US-F16** (bloqueante): cartas, voto propio y polling.
- **Bloquea a US-F18**: acuerdo, cierre y cancelación desde la UI.

## Notas

- El revelado no cierra la sesión; el acuerdo/cierre es US-F18.
- La desviación estándar y `hay_dispersion` se muestran tal como los calcula el
  backend (poblacional; `maximo != minimo`).
- La nueva ronda no borra el historial anterior.
- El frontend no recalcula estadísticas ni decide por sí mismo si hay consenso.
