# Especificación: US-F18 - Acuerdo, cierre y cancelación en Planning Poker

## Historia de usuario

**Como** miembro de una sesión de Planning Poker, **quiero** registrar el acuerdo,
cerrar la sesión o cancelarla, **para** dejar la historia estimada o liberarla y
ver el resultado de la estimación en el backlog.

## Descripción

US-F18 implementa el modal de acuerdo, el cierre de la sesión y la cancelación
desde la UI. El acuerdo usa la escala de US-05, actualiza la historia y deja la
sesión en modo lectura. La cancelación pide confirmación, deja la sesión
`Cancelada` y permite iniciar otra. El resultado de la estimación se muestra en el
backlog, en una columna propia.

## Objetivo

Que cualquier miembro pueda acordar los puntos de una sesión `Revelada` y cerrarla,
o cancelar una sesión `Votando`/`Revelada`. Al cerrar, la historia queda estimada y
la pantalla pasa a modo lectura mostrando los puntos acordados. Al cancelar, la
historia queda libre para una nueva sesión. El backlog refleja los `story_points`
actualizados en una columna `Estimación`.

La historia satisface el issue `SebastianAguilera/ingenieria-y-calidad#84`.
**Requiere** US-17 (acuerdo, cierre y cancelación en backend) y US-F17 (revelado,
dispersión y nuevas rondas). **Requiere** US-05 (escala de puntos).

**Trazabilidad:** Issue #84 → SDD → criterios de aceptación → escenarios BDD
(`docs/features/US-F18-acuerdo-cierre-planning-poker.feature`) → tests Vitest/RTL
→ código React.

## Cross-referencias

- **US-05**: escala de puntos `0, 0.5, 1, 2, 3, 5, 8, 13, 21` en `float64`.
- **US-17**: endpoints `POST .../acuerdo` y `POST .../cancelar`; solo se acuerda
  desde `Revelada`; cancelar desde `Votando`/`Revelada`; errores `422
  PUNTOS_POKER_INVALIDOS`, `409 SESION_POKER_NO_REVELADA`, `409
  SESION_POKER_CERRADA`, `404`.
- **US-F17**: revelado, dispersión (con `mediana` y `votos_por_valor`) y nuevas
  rondas; aporta el detalle que consume el modal de acuerdo.
- **US-F15/F16**: estados `Cerrada`/`Cancelada` de solo lectura y cartas de voto.

## Alcance

- Crear `ModalAcuerdo` accesible.
- Mostrar el botón `Registrar acuerdo` solo en `Revelada`.
- Prefill propuesto: la `mediana` si es un valor válido de la escala; si no, el
  valor más votado; si no, vacío.
- Permitir elegir cualquier valor de la escala.
- Enviar `POST /api/proyectos/{id}/poker-sessions/{sesionId}/acuerdo` con
  `{ "puntos": number }`.
- Al cerrar, mostrar los puntos acordados, estado `Cerrada` y modo lectura.
- Mostrar el error `422 PUNTOS_POKER_INVALIDOS` si el backend rechaza el valor.
- Botón `Cancelar sesión` visible en `Votando` y `Revelada`, con confirmación.
- Enviar `POST /api/proyectos/{id}/poker-sessions/{sesionId}/cancelar`.
- Al cancelar, mostrar estado `Cancelada`, sin puntos acordados y con opción de
  iniciar otra.
- Si la sesión ya está `Cerrada` o `Cancelada`, mostrar modo lectura y deshabilitar
  acciones.
- Backlog: columna `Estimación` en `TablaBacklog`, junto a `Prioridad`, que muestra
  `story_points` o `–` si la historia no está estimada.
- Refrescar el backlog al volver para reflejar los `story_points` actualizados.
- Tests Vitest/RTL de modal, acuerdo, cierre, cancelación, errores y columna
  `Estimación`.
- Accesibilidad: foco en el modal, cierre con Escape, etiquetas y navegación por
  teclado.

## Fuera de alcance

- Modificar el cálculo o los contratos del backend de US-17.
- Notificaciones globales de sesiones activas (US-F19).
- Gráficos, exportaciones, comentarios por voto o animaciones.
- Reapertura o edición de una estimación ya cerrada.
- Permisos especiales: cualquier miembro puede acordar o cancelar (US-11).

## Entradas

- Ruta existente: `/proyectos/:id/backlog/:historiaId/poker`.
- `DetalleSesionPoker` de US-F17 (incluye `dispersion` con `mediana` y
  `votos_por_valor`).
- `POST /api/proyectos/{id}/poker-sessions/{sesionId}/acuerdo`, cuerpo
  `{ "puntos": number }`.
- `POST /api/proyectos/{id}/poker-sessions/{sesionId}/cancelar`, sin cuerpo.
- JWT administrado por `api/client`; la pantalla no maneja tokens.
- Funciones de API previstas:
  - `acordarSesionPoker(proyectoId, sesionId, puntos): Promise<DetalleSesionPoker>`.
  - `cancelarSesionPoker(proyectoId, sesionId): Promise<DetalleSesionPoker>`.

### Contrato del detalle actualizado

La respuesta de `acuerdo` y `cancelar` es el detalle de la sesión ya actualizada,
con el mismo contrato que `GET .../poker-sessions/{sesionId}`:

- `estado`: `Cerrada` (acuerdo) o `Cancelada` (cancelación).
- `puntos_acordados`: `number` en `Cerrada`, `null` en `Cancelada`.
- `cerrado_en` / `cancelado_en`: `string | null`.
- `story_estimate`: objeto en `Cerrada`, `null` en `Cancelada`.
- `votos`, `dispersion` e `historial_rondas`: se conservan.

## Salidas esperadas

### Sesión `Revelada`

- Se mantiene el comportamiento de US-F17 (votos, dispersión, `Nueva ronda`).
- Se muestra el botón `Registrar acuerdo`, visible únicamente en este estado.
- Se muestra el botón `Cancelar sesión`, visible en `Votando` y `Revelada`.

### `ModalAcuerdo`

- Diálogo accesible (`role="dialog"`, `aria-modal`, título asociado).
- Título `Registrar acuerdo`.
- Selector con los nueve valores de la escala como tarjetas de radio etiquetadas
  (`0 puntos`, `0.5 puntos`, …, `21 puntos`).
- Prefill: la opción de la `mediana` queda marcada cuando la mediana pertenece a la
  escala; si no, se marca el valor más votado; si no, ninguna.
- Foco inicial en la opción prefijada o, si no hay, en la primera de la escala.
- Botones `Cancelar` y `Confirmar acuerdo`; el segundo queda deshabilitado y con
  `aria-busy="true"` mientras la petición está en curso.
- Escape y clic en el fondo cierran sin enviar (salvo mientras se procesa).

### Acuerdo

- Confirmar con un valor válido llama a `POST .../acuerdo` con `{ puntos }`.
- Al confirmar sin selección, la UI muestra `Elegí una estimación de la escala.` y
  no envía la petición.
- En caso de éxito, la sesión pasa a `Cerrada`: se muestran los puntos acordados y
  la pantalla queda en modo lectura.
- El botón `Registrar acuerdo`, las cartas y las acciones de votación desaparecen.
- La opción `Iniciar Planning Poker` queda disponible para una nueva sesión.

### Cancelación

- El botón `Cancelar sesión` abre un diálogo de confirmación con el mensaje
  `¿Querés cancelar la sesión? No se estimará la historia y podrás iniciar otra.`
- Confirmar llama a `POST .../cancelar`.
- En caso de éxito, la sesión pasa a `Cancelada`: se informa
  `La última sesión fue cancelada.`, no se muestran puntos acordados y se habilita
  `Iniciar Planning Poker`.
- Cancelar el diálogo (botón, Escape o fondo) no envía la petición.

### Columna `Estimación` en el backlog

- Se agrega entre `Prioridad` y `Acciones`, con encabezado `Estimación`.
- Muestra el `story_points` de la historia; si es `null`, muestra `–`.
- Se ordena después de `Prioridad`, respetando el orden de columnas de la tabla.

### Sesión `Cerrada` o `Cancelada`

- Se mantiene el resultado de US-F15 cuando la última sesión ya terminó.
- No se muestran `Registrar acuerdo`, `Cancelar sesión`, `Revelar votos` ni `Nueva
  ronda`.
- Si el polling detecta el paso a `Cerrada`/`Cancelada`, se detiene y se
  deshabilitan las acciones.

## Reglas de negocio

- **RB-01**: `Registrar acuerdo` solo se muestra y se ejecuta si el detalle está en
  `Revelada`.
- **RB-02**: `Cancelar sesión` solo se muestra y se ejecuta si el detalle está en
  `Votando` o `Revelada`.
- **RB-03**: el valor acordado debe pertenecer a la escala de US-05. El frontend
  valida el vacío antes de enviar; un valor fuera de escala lo rechaza el backend.
- **RB-04**: acordar reemplaza el detalle con la respuesta del backend. El frontend
  no calcula puntos ni actualiza la historia localmente.
- **RB-05**: cancelar reemplaza el detalle con la respuesta del backend y deja
  `puntos_acordados` en `null`.
- **RB-06**: el prefill usa la `mediana` cuando pertenece a la escala; si no, el
  valor con más votos; si hay empate o no hay votos, queda vacío.
- **RB-07**: tras el acuerdo o la cancelación, la sesión deja de considerarse activa
  y la pantalla pasa a modo lectura con la opción de iniciar otra.
- **RB-08**: el polling solo consulta el detalle mientras el estado sea `Votando` o
  `Revelada`; se detiene en `Cerrada` o `Cancelada`.
- **RB-09**: los botones son nativos, tienen nombres accesibles, foco visible y
  evitan doble envío mientras la petición está en curso.
- **RB-10**: el modal atrapa el foco inicial, se cierra con Escape y con clic en el
  fondo, y no envía si no hay selección.
- **RB-11**: una sesión `Cerrada` o `Cancelada` nunca ofrece acordar ni cancelar.
- **RB-12**: el backlog se refresca al volver a la pantalla, porque el listado de
  historias se consulta al montar; no se recalcula `story_points` en el cliente.

## Condiciones de error

| Situación | HTTP | `error` | Comportamiento de la UI |
| :--- | :---: | :--- | :--- |
| Acordar con valor inválido | `422` | `PUNTOS_POKER_INVALIDOS` | Muestra `El valor elegido no es válido. Elegí otro de la escala.` en el modal y conserva `Revelada`. |
| Acordar fuera de `Revelada` | `409` | `SESION_POKER_NO_REVELADA` | Refresca el detalle, cierra el modal y muestra `La sesión ya no está revelada.` |
| Acordar o cancelar en `Cerrada`/`Cancelada` | `409` | `SESION_POKER_CERRADA` | Refresca el detalle y muestra `La sesión ya fue cerrada o cancelada.` |
| Proyecto inexistente o sin acceso | `404` | `PROYECTO_NO_ENCONTRADO` | Mensaje de acceso/proyecto inexistente. |
| Sesión inexistente o de otro proyecto | `404` | `SESION_POKER_NO_ENCONTRADA` | Mensaje de sesión no encontrada. |
| Error de red al acordar | `0` | `NETWORK_ERROR` | Muestra `No se pudo registrar el acuerdo. Intentá nuevamente.` en el modal y permite reintentar. |
| Error de red al cancelar | `0` | `NETWORK_ERROR` | Muestra `No se pudo cancelar la sesión. Intentá nuevamente.` y permite reintentar. |
| Token inválido/expirado | `401` | `TOKEN_*` | Comportamiento global de sesión de `api/client`. |

## Casos límite

- La mediana de votos `1`, `3`, `8` es `3`, que pertenece a la escala: prefill `3`.
- La mediana de votos `1`, `3`, `5`, `8` es `4`, que no pertenece a la escala: se
  usa el valor más votado.
- Si dos o más valores comparten la mayor cantidad de votos, el prefill queda vacío.
- Sin votos o sin `dispersion`, el prefill queda vacío.
- Acordar `0` y `0.5` es válido y se envía tal cual.
- Confirmar sin selección muestra el error local y no envía la petición.
- Otro miembro puede cerrar o cancelar mientras el modal está abierto; el `409` se
  resuelve refrescando el detalle.
- El usuario puede navegar fuera con el modal abierto o una acción en curso; las
  solicitudes deben abortarse o ignorar su resultado.

## Criterios de aceptación

- [ ] **CA-1**: Given una sesión `Revelada`, When se registra un punto válido, Then
  la sesión pasa a `Cerrada`, se muestran los puntos acordados y la historia queda
  actualizada.
- [ ] **CA-2**: Given una sesión `Votando`, When se abre la pantalla, Then no hay
  acción de acordar disponible.
- [ ] **CA-3**: Given un punto inválido, When se confirma, Then se muestra `422
  PUNTOS_POKER_INVALIDOS` y la sesión no se cierra.
- [ ] **CA-4**: Given una sesión `Votando` o `Revelada`, When se cancela y se
  confirma, Then pasa a `Cancelada`, no muestra puntos acordados y permite iniciar
  otra.
- [ ] **CA-5**: Given una sesión `Cerrada` o `Cancelada`, When se abre, Then está en
  modo lectura.
- [ ] **CA-6**: Given un error de red, When se acuerda o se cancela, Then se muestra
  el mensaje y se puede reintentar.
- [ ] **CA-7**: Given el modal abierto, When se presiona Escape, Then se cierra sin
  enviar.
- [ ] **CA-8**: Given una historia con `story_points = 5` tras cerrar la sesión,
  When se vuelve al backlog, Then la columna `Estimación` muestra `5`.
- [ ] **CA-9**: Given una historia sin estimar, When se muestra el backlog, Then la
  columna `Estimación` muestra `–`.

## Tareas de implementación

- [ ] Escribir spec SDD y escenarios BDD de US-F18.
- [ ] RED: tests de `ModalAcuerdo`, acuerdo, cierre, cancelación y columna
  `Estimación`.
- [ ] GREEN: modal, envío de acuerdo, cancelación, estados finales y columna en
  `TablaBacklog`.
- [ ] Refrescar el backlog al volver.
- [ ] Verificación manual y tests automatizados.

## Definición de terminado

- [ ] Spec SDD y escenarios BDD escritos antes del código.
- [ ] Tests RED de API, `ModalAcuerdo` y `SesionPokerPage` escritos antes de la
  implementación.
- [ ] `npm test`, `npm run lint` y `npm run build` en verde.
- [ ] Verificación manual de acuerdo, cierre, cancelación, reapertura y columna
  `Estimación` en el backlog.
- [ ] Accesibilidad revisada por teclado, foco y etiquetas.

## Dependencias

- **Requiere US-17** (bloqueante): endpoints de acuerdo y cancelación en backend.
- **Requiere US-F17** (bloqueante): revelado, dispersión y nuevas rondas.
- **Requiere US-05**: escala de puntos.

## Notas

- El acuerdo puede ser cualquier valor de la escala, no necesariamente uno de los
  votados.
- `story_points` lo actualiza el backend; el frontend solo refresca el backlog.
- La cancelación pide confirmación porque cualquier miembro puede cancelar una
  sesión ajena.
- El resultado de la estimación sigue estando disponible para una nueva sesión de
  la misma historia.
