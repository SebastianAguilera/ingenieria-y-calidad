# Especificación: US-F14 - Detalle de sprint y Sprint Backlog

## Objetivo

Agregar la pantalla de **detalle de un sprint** (`/proyectos/:id/sprints/:sprintId`)
para hacer seguimiento del trabajo planificado: ver la cabecera del sprint y las
historias asignadas a su Sprint Backlog, asignar historias del Product Backlog y
quitarlas, sin usar la API directamente. La pantalla consume el `GET` de US-13 y
los endpoints de asignación/desasignación de US-04, y respeta sus reglas de
negocio (sprint `Activo` para asignar, sprint no `Cerrado` para quitar, ventana
temporal del sprint).

La historia satisface el issue
`SebastianAguilera/ingenieria-y-calidad#73`
(`US-F14 - Detalle de sprint y Sprint Backlog`), del milestone
`Sprint 3: Funcionalidad y Calidad`.

**Trazabilidad:** Issue #73 → esta especificación → criterios de aceptación →
escenarios BDD (`docs/features/US-F14-detalle-sprint-backlog.feature`) → tests
Vitest/RTL → código en `frontend/src`.

## Entradas

### Precondiciones

- **US-13**: `GET /api/proyectos/{id}/sprints/{sprintId}/historias` devuelve el
  sprint y sus `SprintBacklogItem` con la historia anidada, ordenados por
  `orden`. Ya está implementada; los tests de la pantalla usan mocks del
  contrato descrito abajo.
- US-04 implementada: `POST /api/proyectos/{id}/sprints/{sprintId}/historias` y
  `DELETE /api/proyectos/{id}/sprints/{sprintId}/historias/{historiaId}`, ambos
  con `204 No Content` en éxito.
- US-F13 operativa: `frontend/src/api/sprints.ts` y el listado de sprints desde
  el que se navega al detalle.
- US-F07 operativa: tabla del backlog (`TablaBacklog`) y sus tests.
- El guard de membresía (US-F10/US-11) responde `404 PROYECTO_NO_ENCONTRADO`
  para no miembros. Cualquier miembro puede asignar y quitar (no hay permisos
  de líder en esta pantalla).

### Contrato consumido (US-13)

`GET /api/proyectos/{id}/sprints/{sprintId}/historias` → `200`:

```json
{
  "sprint": {
    "id": 1, "identificador": "Sprint 1", "sprint_goal": "Entregar el MVP",
    "estado": "Activo", "fecha_inicio": "2026-10-06", "fecha_fin": "2026-10-20",
    "creado_en": "2026-10-01T09:15:00Z", "actualizado_en": "2026-10-06T09:00:00Z"
  },
  "total": 1,
  "items": [
    {
      "id": 10, "orden": 1, "fecha_asignacion": "2026-10-06T10:00:00Z",
      "historia": {
        "id": 100, "codigo": "US-01", "titulo": "Crear proyecto",
        "descripcion": "Como usuario...", "prioridad": "Alta",
        "estado": "En Progreso", "story_points": 3, "esfuerzo_real": null,
        "creado_en": "2026-10-01T09:00:00Z", "actualizado_en": "2026-10-06T11:00:00Z"
      }
    }
  ]
}
```

> Contrato confirmado contra `docs/specs/US-13-consulta-sprint-backlog-spec.md`
> (US-13 ya está en `develop`).

### Datos de entrada

| Acción | Endpoint | Cuerpo |
| :--- | :--- | :--- |
| Asignar historia | `POST /api/proyectos/{id}/sprints/{sprintId}/historias` | `{"historia_id": 100}` |
| Quitar historia | `DELETE /api/proyectos/{id}/sprints/{sprintId}/historias/{historiaId}` | sin cuerpo |
| Candidatas a asignar | `GET /api/proyectos/{id}/historias` (US-02, ya existe) | sin cuerpo |

## Salidas esperadas

- Ruta anidada `/proyectos/:id/sprints/:sprintId` con la página
  `SprintDetallePage`.
- **Cabecera**: identificador, Sprint Goal, fecha de inicio, fecha de fin y
  estado del sprint.
- **Tabla del Sprint Backlog** (`TablaSprintBacklog`) con columnas Código,
  Título, Estado, Story Points y la acción `Quitar del sprint`, respetando el
  orden devuelto por la API.
- Acción `Asignar historia` que abre un modal con las historias candidatas y
  actualiza la tabla al confirmar, sin recargar la página.
- Confirmación previa al quitar una historia.
- Aviso para sprint `Planificado` y modo lectura para sprint `Cerrado`.
- Estados de pantalla: carga, error con reintento, vacío, listo y sprint no
  encontrado.
- Navegación desde el listado de sprints (US-F13): el identificador de cada
  sprint enlaza a su detalle, y el detalle ofrece un enlace de vuelta a la lista.

## Reglas de negocio

- **RB-01** La página muestra la cabecera del sprint y sus historias asignadas
  tal como las devuelve el backend. → **CA-1**
- **RB-02** Un sprint sin historias muestra el estado vacío; la acción
  `Asignar historia` solo aparece si el sprint está `Activo`. → **CA-2**
- **RB-03** `Asignar historia` solo está disponible con el sprint `Activo`. El
  modal lista las historias del proyecto en estado `Backlog` que **no** estén ya
  en este sprint. → **CA-3, CA-5, CA-6**
- **RB-04** US-04 no cambia el estado de la historia al asignarla (RN-06), por lo
  que una historia en `Backlog` puede estar asignada a **otro** sprint y el
  frontend no puede saberlo. El modal la ofrece igual y, si el backend responde
  `409 HISTORIA_YA_ASIGNADA`, `400 HISTORIA_NO_PENDIENTE` o
  `400 FUERA_DE_VENTANA_TEMPORAL`, se muestra el mensaje del backend, la tabla
  no cambia y no se duplica nada. → **CA-4, CA-9**
- **RB-05** Tras asignar o quitar (ambos devuelven `204` sin cuerpo) la página
  vuelve a consultar el `GET` de US-13 para refrescar la tabla, sin recargar. →
  **CA-3, CA-7**
- **RB-06** Un sprint `Planificado` no permite asignar y muestra el aviso
  "Activá el sprint para poder asignar historias". → **CA-5**
- **RB-07** Un sprint `Cerrado` se muestra en modo lectura: sin `Asignar
  historia` y sin `Quitar del sprint`. Las historias incompletas quedan
  asociadas al sprint cerrado; no vuelven al Product Backlog desde esta
  pantalla. → **CA-6, CA-8**
- **RB-08** `Quitar del sprint` está disponible mientras el sprint no esté
  `Cerrado`, pide confirmación y envía el `DELETE`. Tras confirmar, la historia
  desaparece de la tabla. → **CA-7, CA-8**
- **RB-09** Un `404` (`SPRINT_NO_ENCONTRADA` por sprint inexistente o de otro
  proyecto, o `PROYECTO_NO_ENCONTRADO` por falta de acceso) muestra el estado
  "sprint no encontrado" con enlace de vuelta a la lista. → **CA-10**
- **RB-10** Un error de red al listar, asignar o quitar muestra un mensaje y
  permite reintentar. → **CA-11**
- **RB-11** Los modales (asignar y confirmar quitar) reciben foco, tienen
  etiquetas accesibles, se cierran con `Escape` y son operables por teclado. →
  **CA-13**
- **RB-12** El refactor de la tabla (`TablaHistorias` presentacional con
  columnas y acciones configurables, compuesta por `TablaBacklog` y
  `TablaSprintBacklog`) no cambia el comportamiento del Product Backlog: los
  tests de US-F07 siguen en verde. → **CA-12**
- **RB-13** La página no edita el Sprint Goal ni las fechas, no crea ni cierra
  el sprint, y no cambia el estado `Hecho` de las historias (solo lo muestra). →
  **CA-14**

## Restricciones

- **Sin cambios de backend**: US-13 se implementa en otra historia; aquí solo se
  consume. El hook de datos es específico del Sprint Backlog
  (`useSprintBacklog`): `useHistorias` es del Product Backlog.
- `SprintBacklogItem` se tipa en el frontend con la `historia` anidada.
- No se implementa el retorno de historias incompletas al Product Backlog ni el
  historial de sprints (US-F15 corresponde a Planning Poker y no forma parte de
  esta historia).
- Textos de interfaz en español; tests con Vitest + React Testing Library y
  mocks de `frontend/src/api/sprints.ts` y `frontend/src/api/historias.ts`.
- Verificación: `npm test`, `npm run lint` y `npm run build` en verde.

## Casos límite

- Sprint `Activo` sin historias: estado vacío con la acción de asignar.
- Sprint `Activo` y ninguna historia candidata (todas ya están en este sprint o
  ninguna está en `Backlog`): el modal muestra "No hay historias disponibles
  para asignar".
- Sprint `Planificado` o `Cerrado` sin historias: estado vacío sin acción de
  asignar.
- Historia con `story_points` nulo: la columna muestra un guion (`—`).
- Historia en estado `Hecho` dentro de un sprint cerrado: se muestra, sin
  acciones.
- `sprintId` o `id` no numérico en la URL: se trata como sprint no encontrado.
- Asignar dos veces la misma historia (doble clic): el segundo envío no se
  realiza mientras el primero está en curso.
- Sprint de otro proyecto: `404` del backend, sprint no encontrado.

## Condiciones de error

| Situación | Origen | Respuesta / UI |
| :--- | :--- | :--- |
| Sprint inexistente o de otro proyecto | US-13 | `404 SPRINT_NO_ENCONTRADA` → sprint no encontrado |
| No miembro o proyecto inexistente | Guard US-F10/US-11 | `404 PROYECTO_NO_ENCONTRADO` → `No tenés acceso a este proyecto o no existe` |
| Error de red al listar | Cliente HTTP | Mensaje de error + `Reintentar` |
| Error de red al asignar o quitar | Cliente HTTP | Mensaje en el modal / sobre la tabla + reintento |
| Historia ya asignada | US-04 | `409 HISTORIA_YA_ASIGNADA`, mensaje del backend, tabla sin cambios |
| Historia no pendiente | US-04 | `400 HISTORIA_NO_PENDIENTE`, mensaje del backend |
| Sprint no activo | US-04 | `400 SPRINT_NO_ACTIVO`, mensaje del backend |
| Fuera de la ventana temporal | US-04 | `400 FUERA_DE_VENTANA_TEMPORAL`, mensaje del backend, la historia no se agrega |
| Quitar en sprint cerrado | US-04 | `409 SPRINT_CERRADO`, mensaje del backend |
| Quitar historia que no pertenece al sprint | US-04 | `409 HISTORIA_NO_PERTENECE_A_SPRINT`, mensaje del backend |

## Criterios de aceptación

- **CA-1** Given un miembro, When entra a `/proyectos/:id/sprints/:sprintId`,
  Then ve la cabecera del sprint y sus historias asignadas.
- **CA-2** Given un sprint sin historias, When se abre, Then se ve el estado
  vacío y la acción de asignar si está `Activo`.
- **CA-3** Given un sprint `Activo`, When asigna una historia en `Backlog`, Then
  la historia aparece en la tabla y deja de estar disponible para asignar.
- **CA-4** Given una historia ya asignada o no pendiente, When intenta
  asignarla, Then se muestra el error correspondiente y no se duplica.
- **CA-5** Given un sprint `Planificado`, When se abre, Then no se permite
  asignar y se muestra el aviso.
- **CA-6** Given un sprint `Cerrado`, When se abre, Then se muestra en modo
  lectura, sin asignar ni quitar.
- **CA-7** Given una historia asignada y un sprint no cerrado, When se quita y
  confirma, Then desaparece de la tabla.
- **CA-8** Given un sprint cerrado, When se intenta quitar una historia, Then no
  hay acción disponible.
- **CA-9** Given una asignación fuera de la ventana temporal del sprint, When se
  envía, Then se muestra el error del backend `400 FUERA_DE_VENTANA_TEMPORAL` de
  US-04 y la historia no se agrega.
- **CA-10** Given un sprint de otro proyecto o inexistente, When se abre, Then
  se muestra `404`.
- **CA-11** Given un error de red, When se lista, asigna o quita, Then se
  muestra el mensaje y se puede reintentar.
- **CA-12** Given el refactor de tabla, When se ejecutan los tests de US-F07,
  Then siguen en verde.
- **CA-13** Given un modal abierto, When se usa el teclado, Then el foco queda en
  el modal, tiene etiquetas accesibles y `Escape` lo cierra.
- **CA-14** Given el detalle de un sprint, When se revisan las acciones, Then no
  hay edición de goal ni fechas, ni crear o cerrar el sprint.

## Arquitectura

Solo frontend (`frontend/src`):

- `api/sprints.ts`: `obtenerSprintBacklog`, `asignarHistoria`,
  `quitarHistoria` (se suman a las funciones de US-F13) y el tipo
  `SprintBacklogItem` en `api/types.ts`.
- `pages/proyecto/SprintDetallePage.tsx` y, en `pages/proyecto/sprints/`,
  `TablaSprintBacklog.tsx`, `ModalAsignarHistoria.tsx` y `useSprintBacklog.ts`.
- Refactor: `pages/proyecto/backlog/TablaHistorias.tsx` (presentacional, columnas
  y acciones por props), compuesto por `TablaBacklog` y `TablaSprintBacklog`.
- Ruta `sprints/:sprintId` en `App.tsx` y enlace desde `SprintsPage`.
