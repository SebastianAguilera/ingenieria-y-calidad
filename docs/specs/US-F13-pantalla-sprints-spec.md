# Especificación: US-F13 - Pantalla de gestión de sprints

## Objetivo

Reemplazar el placeholder de **Sprints** por una pantalla operativa que permita
listar los sprints de un proyecto, crear uno nuevo y activar o cerrar sprints
sin usar la API directamente. La pantalla consume los endpoints de US-03 y
respeta sus reglas de negocio (identificador, Sprint Goal, fechas y máquina de
estados `Planificado -> Activo -> Cerrado`).

La historia satisface el issue
`SebastianAguilera/ingenieria-y-calidad#72`
(`US-F13 - Pantalla de gestión de sprints`), del milestone
`Sprint 3: Funcionalidad y Calidad`.

**Trazabilidad:** Issue #72 → esta especificación → criterios de aceptación →
escenarios BDD (`docs/features/US-F13-pantalla-sprints.feature`) → tests
Vitest/RTL → código en `frontend/src`.

## Entradas

### Precondiciones

- US-03 implementada: `POST /api/proyectos/{id}/sprints`,
  `GET /api/proyectos/{id}/sprints` y
  `PATCH /api/proyectos/{id}/sprints/{sprintId}/estado`.
- US-F01 (listado y detalle de proyectos) y US-F03 (navegación) operativas.
- El guard de membresía (US-F10/US-11) responde `404 PROYECTO_NO_ENCONTRADO`
  para no miembros.
- El usuario entra a la ruta existente `/proyectos/:id/sprints`.

### Datos de entrada

| Campo | Tipo | Obligatorio | Regla |
| :--- | :--- | :---: | :--- |
| `identificador` | texto | sí | 1 a 50 caracteres después de recortar |
| `sprint_goal` | texto | sí | 1 a 300 caracteres después de recortar |
| `fecha_inicio` | fecha `YYYY-MM-DD` | sí | fecha calendario |
| `fecha_fin` | fecha `YYYY-MM-DD` | sí | igual o posterior a `fecha_inicio` |
| `estado` (destino) | `Planificado \| Activo \| Cerrado` | sí | solo transiciones válidas |

## Salidas esperadas

- Listado con **identificador, Sprint Goal, fecha de inicio, fecha de fin y
  estado** por cada sprint.
- Filtro por estado con las opciones `Todos`, `Planificado`, `Activo`,
  `Cerrado`.
- Alta de sprint mediante modal; todo sprint nuevo aparece como
  `Planificado`.
- Acciones `Activar` (desde `Planificado`) y `Cerrar` (desde `Activo`) con
  confirmación previa.
- Estados de pantalla: carga, error con reintento, vacío y listo.

## Reglas de negocio

- **RB-01** El listado muestra los sprints devueltos por el backend; un sprint
  sin sprints muestra el estado vacío con la acción de crear. → **CA-1, CA-2**
- **RB-02** El filtro por estado es local: `Todos` muestra todo y cada estado
  muestra solo ese estado; un filtro sin resultados muestra un aviso, no el
  estado vacío de alta. → **CA-3**
- **RB-03** El alta valida en el cliente las mismas reglas que US-03
  (identificador 1-50, goal 1-300, fechas obligatorias y `fecha_fin >=
  fecha_inicio`); si hay errores no se llama a la API y el mensaje se muestra
  junto al campo. → **CA-4, CA-5**
- **RB-04** El sprint creado se agrega al listado en estado `Planificado` sin
  recargar la página. → **CA-6**
- **RB-05** `Activar` solo aplica a sprints `Planificado`. Si el listado ya
  contiene un sprint `Activo`, el botón `Activar` de los demás queda
  deshabilitado. → **CA-7, CA-9**
- **RB-06** `Cerrar` solo aplica al sprint `Activo`. Ambas acciones piden
  confirmación antes de llamar a la API. La confirmación de cierre advierte
  que el sprint no se puede volver a abrir, porque `Cerrado` es terminal en
  US-03. → **CA-8, CA-10**
- **RB-07** Un sprint `Cerrado` se muestra en modo lectura, sin acciones. →
  **CA-10**
- **RB-08** Si la API responde `409 SPRINT_ACTIVO_EXISTENTE` (carrera entre
  usuarios), se muestra el mensaje del backend y no se cambia el estado. →
  **CA-11**
- **RB-09** Un `404 PROYECTO_NO_ENCONTRADO` muestra
  `No tenés acceso a este proyecto o no existe`. → **CA-12**
- **RB-10** Un error de red al listar o al guardar muestra un mensaje y
  permite reintentar. → **CA-13**
- **RB-11** El modal de alta recibe foco, tiene etiquetas accesibles y se
  cierra con `Escape` o con `Cancelar`. → **CA-14**

## Restricciones

- **Sin cambios de backend**: se reutilizan los endpoints y DTOs de US-03.
- La pantalla no edita goal ni fechas, no elimina sprints y no navega al
  detalle; el detalle y el Sprint Backlog pertenecen a US-F14.
- El filtro `Cerrado` cubre la consulta de sprints anteriores (no se crea una
  historia aparte).
- Cualquier miembro puede crear, activar y cerrar; no hay permisos adicionales.
- Textos de interfaz en español; tests con Vitest + React Testing Library y
  mocks de `frontend/src/api/sprints.ts`.

## Casos límite

- Proyecto sin sprints: estado vacío con botón de creación.
- Filtro por un estado sin coincidencias: aviso "No hay sprints con ese
  estado"; no se ofrece crear desde ahí.
- Identificador de exactamente 50 caracteres y goal de exactamente 300:
  válidos (límite inclusivo).
- Fechas de inicio y fin iguales: válidas.
- Sprint único en estado `Activo`: su acción disponible es `Cerrar`, no
  `Activar`.
- Sprint `Cerrado`: sin acciones aun cuando no haya otro activo.
- Alta con un identificador ya existente: `409 SPRINT_DUPLICADO` con el campo
  `identificador` marcado.
- Fechas solapadas con otro sprint: `409 SPRINT_FECHAS_SOLAPADAS`.
- `id` de proyecto inválido en la URL: se trata como proyecto inexistente.

## Condiciones de error

| Situación | Origen | Respuesta / UI |
| :--- | :--- | :--- |
| No miembro o proyecto inexistente | Guard US-F10/US-11 | `404 PROYECTO_NO_ENCONTRADO` → mensaje de acceso |
| Error de red al listar | Cliente HTTP | Mensaje de error + `Reintentar` |
| Error de red al crear/activar/cerrar | Cliente HTTP | Mensaje en el modal o sobre el listado + reintento |
| Validación de cliente | `validacionSprint` | Error por campo, sin llamar a la API |
| Identificador duplicado | US-03 | `409 SPRINT_DUPLICADO`, campo `identificador` |
| Fechas solapadas | US-03 | `409 SPRINT_FECHAS_SOLAPADAS`, mensaje general |
| Sprint activo existente (carrera) | US-03 | `409 SPRINT_ACTIVO_EXISTENTE`, mensaje general |
| Transición inválida | US-03 | `409 SPRINT_TRANSICION_INVALIDA`, mensaje general |

## Criterios de aceptación

- [ ] **CA-1** Given un proyecto sin sprints, When se entra a Sprints, Then se
  ve el estado vacío y la acción para crear.
- [ ] **CA-2** Given un proyecto con sprints, When se entra, Then se listan
  identificador, Sprint Goal, fechas y estado.
- [ ] **CA-3** Given el filtro `Cerrado`, When se selecciona, Then solo se
  muestran sprints cerrados.
- [ ] **CA-4** Given datos válidos, When se crea un sprint, Then aparece en la
  lista como `Planificado`.
- [ ] **CA-5** Given datos inválidos, When se envía el formulario, Then se
  muestra el error en el campo y no se llama a la API.
- [ ] **CA-6** Given el alta exitosa, When el modal se cierra, Then el nuevo
  sprint queda en el listado sin recargar.
- [ ] **CA-7** Given un sprint `Planificado` y sin otro activo, When se activa
  y confirma, Then pasa a `Activo`.
- [ ] **CA-8** Given un sprint `Activo`, When se cierra y confirma tras ver
  que no se puede volver a abrir, Then pasa a `Cerrado`.
- [ ] **CA-9** Given un sprint `Activo` en la lista, When se muestra `Activar`
  en los demás, Then el botón queda deshabilitado.
- [ ] **CA-10** Given un sprint `Cerrado`, When se muestra, Then no tiene
  acciones.
- [ ] **CA-11** Given `409 SPRINT_ACTIVO_EXISTENTE`, When se intenta activar,
  Then se muestra ese mensaje y no cambia el estado.
- [ ] **CA-12** Given un no miembro, When entra al proyecto, Then ve `404` con
  `No tenés acceso a este proyecto o no existe`.
- [ ] **CA-13** Given un error de red, When se lista o se guarda, Then se
  muestra el mensaje y se puede reintentar.
- [ ] **CA-14** Given el modal, When se abre, Then recibe foco, tiene etiquetas
  accesibles y se cierra con `Escape`.

## Archivos afectados

| Archivo | Cambio |
| :--- | :--- |
| `docs/specs/US-F13-pantalla-sprints-spec.md` | **Nuevo** esta especificación |
| `docs/features/US-F13-pantalla-sprints.feature` | **Nuevo** escenarios BDD |
| `frontend/src/api/sprints.ts` | **Nuevo** `listarSprints`, `crearSprint`, `cambiarEstadoSprint` |
| `frontend/src/api/sprints.test.ts` | **Nuevo** tests de la API |
| `frontend/src/pages/proyecto/sprints/validacionSprint.ts` | **Nuevo** validación de cliente |
| `frontend/src/pages/proyecto/sprints/validacionSprint.test.ts` | **Nuevo** tests de límites |
| `frontend/src/pages/proyecto/sprints/useSprints.ts` | **Nuevo** carga, alta y cambio de estado |
| `frontend/src/pages/proyecto/sprints/ModalSprint.tsx` | **Nuevo** modal de alta accesible |
| `frontend/src/pages/proyecto/sprints/ModalConfirmacion.tsx` | **Nuevo** confirmación de acciones |
| `frontend/src/pages/proyecto/SprintsPage.tsx` | Reemplaza el placeholder |
| `frontend/src/pages/proyecto/SprintsPage.test.tsx` | **Nuevo** tests de la pantalla |

No se modifican el backend ni los tipos de `api/types.ts` (ya existen
`Sprint`, `EstadoSprint`, `AltaSprint` y `ListadoSprints`).

## Dependencias

- **US-03** (`#30`): endpoints y reglas de sprints, ya implementada.
- **US-F01** (`#50`) y **US-F03** (`#52`): base del panel y navegación.
- **US-F10** (`#66`): guard de membresía.
- Puede ejecutarse en paralelo con **US-13** (`#71`).
- **US-F14** (`#73`) depende de esta pantalla.

## Fuera de alcance

- Detalle del sprint y Sprint Backlog (US-F14).
- Edición de goal o fechas, eliminación de sprints.
- Enlace a `/proyectos/:id/sprints/:sprintId` (se agrega en US-F14).
- Cambios de backend, esquema o nuevas migraciones.
- Permisos por rol dentro del proyecto.

## Escenarios BDD

Ver `docs/features/US-F13-pantalla-sprints.feature`.
