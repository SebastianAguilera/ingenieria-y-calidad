# Especificación: US-F20 - Pantalla de defectos del proyecto

## Objetivo

Reemplazar el placeholder de **Defectos** por una pantalla operativa que permita
listar los defectos de un proyecto, registrar uno nuevo, cambiar su estado y
sprint de resolución, filtrarlos por sprint y editar su descripción y
severidad, sin usar la API directamente. La pantalla consume los endpoints de
US-09 (listado, alta y cambio de estado) y de US-18 (listado por sprint y
edición de `descripcion` y `severidad`) y respeta sus reglas de negocio.

La historia satisface el issue `SebastianAguilera/ingenieria-y-calidad#93`
(`US-F20 - Pantalla de defectos del proyecto`). Bloquea a US-F21 (métricas de
defectos en pantalla).

**Trazabilidad:** Issue #93 → esta especificación → criterios de aceptación →
escenarios BDD (`docs/features/US-F20-pantalla-defectos.feature`) → tests
Vitest/RTL → código en `frontend/src`.

Se implementa en dos tramos, cada uno con su PR:

- **Tramo 1** (sin dependencias): listado, alta y cambio de estado.
- **Tramo 2** (requiere US-18, ya mergeada): filtro por sprint, edición de
  `descripcion` y `severidad` y validación de largo. No se empieza hasta que el
  tramo 1 esté mergeado.

## Entradas

### Precondiciones

- US-09 implementada: `GET`/`POST /api/proyectos/{id}/defectos` y
  `PATCH /api/proyectos/{id}/defectos/{defectoId}`.
- US-18 implementada (tramo 2): `GET /api/proyectos/{id}/sprints/{sprintId}/defectos?tipo=deteccion|resolucion`
  y `PATCH` con `descripcion` y `severidad`.
- `listarHistorias` (US-02) y `listarSprints` (US-03) existentes en el cliente.
- El guard de membresía (US-F10/US-11) responde `404 PROYECTO_NO_ENCONTRADO`
  para no miembros.
- El usuario entra a la ruta existente `/proyectos/:id/defectos` (ya registrada
  en `App.tsx` y en la barra lateral).

### Datos de entrada

**Alta de defecto** (`POST`)

| Campo | Tipo | Obligatorio | Regla |
| :--- | :--- | :---: | :--- |
| `descripcion` | texto | sí | no vacía tras recortar; el alta no tiene límite de largo en el backend |
| `severidad` | `minor \| major \| critical \| blocker` | sí | lista cerrada |
| `historia_id` | entero | sí | historia del proyecto, elegida de `listarHistorias` |
| `sprint_deteccion_id` | entero | sí | sprint del proyecto, elegido de `listarSprints` |
| `reportado_por` | — | no | **no se muestra ni se envía** |

El estado inicial lo fija el backend (`open`); el cliente no lo informa.

**Cambio de estado** (`PATCH`)

| Campo | Tipo | Obligatorio | Regla |
| :--- | :--- | :---: | :--- |
| `estado` | `open \| in_progress \| resolved \| closed \| rejected` | sí | lista cerrada |
| `sprint_resolucion_id` | entero | solo al resolver o cerrar | se envía únicamente si el estado destino es `resolved` o `closed` |

**Edición** (`PATCH`, tramo 2)

| Campo | Tipo | Obligatorio | Regla |
| :--- | :--- | :---: | :--- |
| `descripcion` | texto | no | no vacía tras recortar, máximo 1000 caracteres (runas) |
| `severidad` | `minor \| major \| critical \| blocker` | no | lista cerrada |

El cliente **nunca** envía `sprint_deteccion_id` en el `PATCH` (el backend lo
rechaza con `422` por ser no modificable).

**Filtro** (tramo 2): sprint (de `listarSprints`) + tipo (`deteccion` o
`resolucion`).

## Salidas esperadas

- Tabla con **Código** (`DEF-{id}`), **Descripción**, **Historia**,
  **Severidad**, **Estado**, **Sprint de detección**, **Sprint de resolución** y
  **Acciones**.
- Descripción truncada a 80 caracteres con `…` en la tabla, con el texto
  completo disponible como `title`.
- Historia como `US-xx` (el `codigo` de la historia); `—` si el `historia_id`
  no está en la lista de historias.
- Sprint de detección y de resolución como el `identificador` del sprint; `—`
  si no hay sprint o no está en la lista.
- Severidad y estado como etiquetas con color y texto en español.
- Alta de defecto mediante modal; el defecto nuevo aparece en el listado sin
  recargar, en estado `Abierto`.
- Cambio de estado mediante modal con selector de estado y, al resolver o
  cerrar, selector de sprint de resolución.
- Estados de pantalla: carga, error con reintento, vacío y listo.
- Tramo 2: filtro por sprint y tipo, y edición de `descripcion` y `severidad`
  mediante modal.

## Reglas de negocio

- **RB-01** El listado muestra los defectos del proyecto devueltos por el
  backend, ordenados del más reciente al más antiguo (`id` descendente) en el
  cliente. Un proyecto sin defectos muestra el estado vacío con la acción de
  crear. → **CA-1, CA-3**
- **RB-02** El código `DEF-{id}` es solo de presentación: el backend no lo
  guarda ni lo devuelve. → **CA-1**
- **RB-03** La columna Historia cruza `historia_id` con `listarHistorias`. Si
  no se encuentra, muestra `—` y la página no falla. → **CA-2**
- **RB-04** El alta valida en el cliente: descripción obligatoria, severidad,
  historia y sprint de detección elegidos. Si hay errores no se llama a la API
  y el mensaje se muestra junto al campo. → **CA-5**
- **RB-05** El defecto creado se agrega al listado sin recargar la página. →
  **CA-4**
- **RB-06** El alta exige al menos una historia y un sprint. Si el proyecto no
  tiene historias o no tiene sprints, el modal avisa qué falta crear y no
  permite guardar. → **CA-5**
- **RB-07** Pasar a `resolved` o `closed` exige sprint de resolución (US-09
  RN-08). La regla se informa en el modal. Si falta, **no se bloquea en el
  cliente**: se envía el `PATCH` y se muestra el `422` del backend junto al
  campo `sprint_resolucion_id`. → **CA-6**
- **RB-08** Al pasar a `open`, `in_progress` o `rejected` no se envía
  `sprint_resolucion_id` (el backend lo limpia junto con `resuelto_en`). →
  **CA-7**
- **RB-09** Un cambio de estado exitoso reemplaza el defecto en la lista sin
  recargar. → **CA-7**
- **RB-10** Un `400 ESTADO_INVALIDO` u otro error del backend en el cambio de
  estado se muestra en el modal y no modifica la lista. → **CA-7**
- **RB-11** Un `404 PROYECTO_NO_ENCONTRADO` muestra
  `No tenés acceso a este proyecto o no existe`, mediante un helper compartido
  (`MENSAJE_PROYECTO_INEXISTENTE`). → **CA-9**
- **RB-12** Un error de red al listar, guardar o cambiar estado muestra un
  mensaje y permite reintentar. La carga necesita defectos, historias y
  sprints: si alguna falla se muestra el error con `Reintentar`. → **CA-8**
- **RB-13** Los modales reciben foco, tienen `aria-modal`, título y campos
  asociados con `useId`, `aria-invalid` y `aria-describedby` en los errores, y
  se cierran con `Escape` o `Cancelar`. → **CA-10**
- **RB-14** *(tramo 2)* El filtro elige un sprint y un tipo. Con un sprint
  elegido se consulta `GET .../sprints/{sprintId}/defectos?tipo=...`; sin sprint
  se vuelve al listado completo. Un filtro sin resultados muestra un aviso, no
  el estado vacío de alta. → **CA-11, CA-12**
- **RB-15** *(tramo 2)* La edición permite cambiar `descripcion` y `severidad`.
  Valida descripción no vacía y de hasta 1000 caracteres contados como runas
  (no bytes). Si hay errores no se llama a la API. Solo se envían los campos
  que cambiaron. → **CA-13, CA-14**
- **RB-16** *(tramo 2)* Un `422` del backend en la edición (`descripcion`,
  `severidad`) se muestra junto al campo. Editar no modifica el estado ni el
  sprint de resolución. → **CA-13, CA-14**

## Restricciones

- **Sin cambios de backend**: se reutilizan los endpoints y DTOs de US-09 y
  US-18.
- Cualquier miembro puede crear, cambiar estado y editar; no hay permisos
  adicionales por rol.
- No hay paginación (el backend no la ofrece para defectos y el issue no la
  pide).
- Textos de interfaz en español; tests con Vitest + React Testing Library y
  mocks de `frontend/src/api/defectos.ts`, `historias.ts` y `sprints.ts`.
- Los estados y severidades se muestran traducidos (`open` → Abierto,
  `in_progress` → En progreso, `resolved` → Resuelto, `closed` → Cerrado,
  `rejected` → Rechazado; `minor` → Menor, `major` → Mayor, `critical` →
  Crítica, `blocker` → Bloqueante) pero se envían con el valor de la API.
- El helper del mensaje 404 se usa solo en esta pantalla. Las demás páginas que
  repiten el texto no se tocan en esta historia.

## Casos límite

- Proyecto sin defectos: estado vacío con botón de creación.
- Defecto cuya historia no figura en `listarHistorias`: columna Historia `—`.
- Defecto con `sprint_resolucion_id` nulo: columna Sprint de resolución `—`.
- Descripción de exactamente 80 caracteres: no se trunca; de 81: se trunca.
- Proyecto sin historias o sin sprints: el alta avisa y no permite guardar.
- Resolver o cerrar sin sprint de resolución: el cliente envía y el backend
  responde `422` en `sprint_resolucion_id`.
- Reabrir un defecto resuelto: se envía solo `estado`.
- Respuesta `null` del listado (`defectos` ausente): se trata como arreglo
  vacío (`?? []`).
- `id` de proyecto inválido en la URL: se trata como proyecto inexistente.
- Tramo 2: descripción de exactamente 1000 caracteres es válida; de 1001, no;
  1000 caracteres multibyte (`ñ`) es válida.
- Tramo 2: descripción solo con espacios es inválida.
- Tramo 2: filtro por un sprint sin defectos del tipo elegido: aviso, no estado
  vacío.

## Condiciones de error

| Situación | Origen | Respuesta / UI |
| :--- | :--- | :--- |
| No miembro o proyecto inexistente | Guard US-F10/US-11 | `404 PROYECTO_NO_ENCONTRADO` → mensaje de acceso |
| Error de red al listar | Cliente HTTP | Mensaje de error + `Reintentar` |
| Error de red al crear o cambiar estado | Cliente HTTP | Mensaje en el modal, se puede reintentar |
| Validación de cliente | `validacionDefecto` | Error por campo, sin llamar a la API |
| Descripción o severidad inválida en el alta | US-09 | `422 VALIDACION`, campo `descripcion` o `severidad` |
| Historia o sprint inexistente | US-09 | `404 HISTORIA_NO_ENCONTRADA` / `SPRINT_NO_ENCONTRADA`, mensaje general |
| Resolver o cerrar sin sprint de resolución | US-09 | `422 VALIDACION`, campo `sprint_resolucion_id` |
| Estado fuera de la lista | US-09 | `400 ESTADO_INVALIDO`, campo `estado` |
| Defecto inexistente | US-09 | `404 DEFECTO_NO_ENCONTRADO`, mensaje general |
| Descripción vacía o > 1000 al editar | US-18 | `422 VALIDACION`, campo `descripcion` |
| Severidad inválida al editar | US-18 | `422 VALIDACION`, campo `severidad` |
| Sprint del filtro inexistente | US-18 | `404 SPRINT_NO_ENCONTRADA`, mensaje de error |

## Criterios de aceptación

**Tramo 1**

- [ ] **CA-1** Given un miembro, When entra a Defectos, Then ve el listado con
  código `DEF-{id}`, descripción, `US-xx`, severidad, estado y sprint de
  detección.
- [ ] **CA-2** Given un defecto cuya historia no está en la lista de historias,
  When se muestra, Then la columna Historia dice `—` y no falla la página.
- [ ] **CA-3** Given un proyecto sin defectos, When se entra, Then se ve el
  estado vacío y la acción de crear.
- [ ] **CA-4** Given un formulario de alta válido, When se guarda, Then el
  defecto aparece en la lista sin recargar la página.
- [ ] **CA-5** Given un formulario inválido, When se envía, Then se muestra el
  error en el campo y no se llama a la API.
- [ ] **CA-6** Given un cambio de estado a `resolved` o `closed` sin sprint de
  resolución, When se confirma, Then el backend responde 422 y la UI muestra el
  error en `sprint_resolucion_id`.
- [ ] **CA-7** Given un cambio de estado válido, When se confirma, Then la lista
  se actualiza.
- [ ] **CA-8** Given un error de red, When se lista, guarda o cambia, Then se
  muestra el mensaje y se puede reintentar.
- [ ] **CA-9** Given un no miembro, When entra al proyecto, Then ve
  `No tenés acceso a este proyecto o no existe`.
- [ ] **CA-10** Given un modal, When se abre, Then recibe foco, tiene etiquetas
  accesibles y se cierra con `Escape`.

**Tramo 2** (requiere US-18 mergeada)

- [ ] **CA-11** Given el filtro por sprint de detección, When se elige un
  sprint, Then solo se muestran los defectos detectados en él.
- [ ] **CA-12** Given el filtro por sprint de resolución, When se elige un
  sprint, Then solo se muestran los defectos resueltos en él.
- [ ] **CA-13** Given una edición de descripción o severidad válida, When se
  guarda, Then la lista refleja el cambio.
- [ ] **CA-14** Given una descripción de más de 1000 caracteres, When se edita,
  Then se muestra el error en el campo y no se llama a la API.

## Archivos afectados

| Archivo | Tramo | Cambio |
| :--- | :---: | :--- |
| `docs/specs/US-F20-pantalla-defectos-spec.md` | 1 | **Nuevo** esta especificación |
| `docs/features/US-F20-pantalla-defectos.feature` | 1 | **Nuevo** escenarios BDD |
| `frontend/src/api/defectos.ts` | 1 | **Nuevo** `listarDefectos`, `crearDefecto`, `actualizarDefecto` |
| `frontend/src/api/defectos.test.ts` | 1 | **Nuevo** tests de la API |
| `frontend/src/pages/proyecto/mensajes.ts` | 1 | **Nuevo** helper `MENSAJE_PROYECTO_INEXISTENTE` |
| `frontend/src/pages/proyecto/defectos/etiquetasDefecto.ts` | 1 | **Nuevo** textos y clases por estado y severidad |
| `frontend/src/pages/proyecto/defectos/validacionDefecto.ts` (+ test) | 1 | **Nuevo** validación de alta |
| `frontend/src/pages/proyecto/defectos/useDefectos.ts` | 1 | **Nuevo** carga, alta y cambio de estado |
| `frontend/src/pages/proyecto/defectos/ModalDefecto.tsx` | 1 | **Nuevo** modal de alta accesible |
| `frontend/src/pages/proyecto/defectos/ModalEstadoDefecto.tsx` | 1 | **Nuevo** modal de cambio de estado |
| `frontend/src/pages/proyecto/DefectosPage.tsx` | 1 | Reemplaza el placeholder |
| `frontend/src/pages/proyecto/DefectosPage.test.tsx` | 1 | **Nuevo** tests de la pantalla |
| `frontend/src/api/defectos.ts` | 2 | `listarDefectosPorSprint` |
| `frontend/src/api/types.ts` | 2 | `ActualizacionDefecto` gana `descripcion` y `severidad` |
| `frontend/src/pages/proyecto/defectos/ModalEdicionDefecto.tsx` | 2 | **Nuevo** modal de edición |
| `frontend/src/pages/proyecto/defectos/validacionDefecto.ts` | 2 | Validación de edición (límite 1000) |
| `frontend/src/pages/proyecto/defectos/useDefectos.ts` | 2 | Filtro y edición |

Los tipos `Defecto`, `AltaDefecto`, `ListadoDefectos`, `EstadoDefecto` y
`SeveridadDefecto` ya existen en `api/types.ts`. No se modifica el backend.

## Dependencias

- **US-09** (#45): endpoints y reglas de defectos, ya implementada.
- **US-18** (#92): filtro por sprint y edición, ya mergeada (#97). Requerida
  por el tramo 2.
- **US-02**: `listarHistorias`. **US-03**: `listarSprints`.
- **US-F10** (#66): guard de membresía.
- **US-F13** (#72): patrón de pantalla, modal y hook que se sigue.
- **US-F21** (#103) depende de esta pantalla.

## Fuera de alcance

- Métricas de defectos (US-F21).
- Eliminación de defectos e historial de cambios.
- El campo `reportado_por` (no se muestra ni se envía).
- Edición del sprint de detección (no modificable en el backend).
- Un código o título propio del defecto: es un cambio de backend y se abriría
  como historia aparte.
- Paginación y ordenamiento configurable.
- Refactor de las otras páginas que repiten el mensaje 404.
- Cambios de backend, esquema o migraciones.

## Supuestos a confirmar

- La columna **Sprint de resolución** se agrega a las columnas pedidas en el
  issue, para que el efecto de resolver y el filtro por resolución sean
  visibles en la tabla.
- Orden del listado: más reciente primero (`id` descendente); el issue no
  fija el orden.
- El límite de 1000 caracteres solo lo valida el backend al editar (US-18); el
  alta no lo valida, por lo que el cliente tampoco lo exige en el alta. Un
  defecto creado con más de 1000 caracteres deberá acortarse al editarlo.
- Al fallar la carga de historias o sprints se muestra el error de la pantalla
  completa con `Reintentar`, en lugar de degradar la tabla.

## Escenarios BDD

Ver `docs/features/US-F20-pantalla-defectos.feature`.
