# Especificación: US-F06 - Ver el detalle de un proyecto en formato tabla

## Objetivo

Especificar la **pantalla de detalle de un proyecto** (`/proyectos/:id` y sus rutas
hijas) del frontend (React + TypeScript + Tailwind + React Router v7 + Vitest):
la carga del proyecto con un único fetch compartido a todas las secciones, el
encabezado del detalle con el nombre del proyecto y la etiqueta "Cerrado" cuando
corresponde, la sección **Resumen** como tabla de dos columnas (campo/valor), la
sección **Equipo** como tabla de integrantes con su estado vacío, y los estados de
carga, error genérico y proyecto inexistente.

Se trata de una **historia de frontend**: el endpoint `GET /api/proyectos/:id` ya
está implementado en el backend Go (`internal/handler/proyecto_handler.go`,
`ObtenerPorID`), por lo que no se define ningún cambio de backend. Tampoco se
tocan las rutas ni el menú de secciones, que quedaron cableados en US-F03.

La funcionalidad satisface el issue `SebastianAguilera/ingenieria-y-calidad#55`
(milestone *Sprint 2: Interfaz Objetivo*) y depende de US-F01 (cliente HTTP
tipado), US-F02 (sesión y rutas protegidas), US-F03 (rutas y menú del proyecto) y
US-F04 (listado que navega a `/proyectos/:id`).

**Trazabilidad:** Issue #55 → esta especificación →
`docs/features/US-F06-detalle-proyecto.feature` → tests Vitest
(`frontend/src/**/*.test.{ts,tsx}`) → código. Ciclo TDD (RED → GREEN → REFACTOR).

## Alcance ya cubierto por US-F03

US-F03 (mergeada en `develop`, commit `7f86429`) dejó el esqueleto del detalle
listo; esta historia **no vuelve a definir** nada de esto, solo lo consume:

- `frontend/src/layouts/ProyectoLayout.tsx` existe, pero hoy es solo
  `<Outlet />` (sin encabezado ni fetch). US-F06 lo completa.
- El menú de secciones ya existe: `BarraLateralProyecto`
  (`frontend/src/layouts/BarraLateralProyecto.tsx`), renderizado por `AppShell`
  cuando la ruta empieza con `/proyectos/:id`, con `NavLink` a Resumen (index),
  Backlog, Sprints, Equipo, Defectos y Métricas. **No se modifica.**
- Las rutas hijas ya están cableadas en `frontend/src/App.tsx`:
  `proyectos/:id` con `ProyectoLayout` y las seis secciones (index →
  `ResumenPage`, `backlog`, `sprints`, `equipo`, `defectos`, `metricas`).
  **No se agregan ni cambian rutas.**
- Los placeholders de las 6 páginas ya existen en
  `frontend/src/pages/proyecto/` (`ResumenPage`, `EquipoPage`, `BacklogPage`,
  `SprintsPage`, `DefectosPage`, `MetricasPage`). US-F06 reemplaza el contenido
  de **Resumen** y **Equipo**; los otros cuatro quedan placeholder hasta sus
  propias historias.
- Los tests de navegación, enlace activo (`aria-current="page"`), colapso
  responsive y 404 de ruta ya están cubiertos por US-F03; esta historia no los
  duplica (solo agrega los tests de contenido y estados propios).

## Entradas

### Precondiciones

- Sesión activa de US-F02: `RutaProtegida` redirige a `/login` sin usuario
  activo; el detalle siempre se renderiza dentro de `AppShell`.
- Cliente HTTP tipado de US-F01 (`frontend/src/api/client.ts`): convierte los
  errores HTTP en `ApiError` (`frontend/src/api/errors.ts`) con `status`, `code`
  y `message`, y soporta `AbortSignal`.
- Tipos ya existentes en `frontend/src/api/types.ts`: `Proyecto` (incluye
  `usuarios: UsuarioProyecto[]`) y `UsuarioProyecto` (`rol` es opcional y **no**
  viene en el detalle).
- El listado de US-F04 (`ProyectosPage`) ya navega a `/proyectos/:id` con
  `irAlDetalle(id)`; esta historia es el destino de esa navegación.
- Componentes y tokens existentes: `PageHeader`, `Card`, `Alert`, `Button`,
  `bg-page`, `bg-surface`, `text-ink`, `text-ink-soft`, `border-border`,
  `rounded-tile`, `overflow-x-auto`. **No se usan colores hexadecimales
  hardcodeados.**

### Contrato con el backend (entrada de datos)

`GET /api/proyectos/:id` (ya implementado) responde:

| Campo | Tipo | Observaciones |
| :--- | :--- | :--- |
| `id` | `number` | Coincide con el `:id` pedido |
| `nombre` | `string` | |
| `fecha_inicio` | `string` `YYYY-MM-DD` | |
| `fecha_fin` | `string` `YYYY-MM-DD` o `null` | `null` = sin fecha de fin |
| `estado` | `'Activo' \| 'Cerrado'` | |
| `cantidad_usuarios` | `number` | |
| `usuarios` | `UsuarioProyecto[]` | `{ id, nombre, email, creado_en }`; **sin `rol`** |
| `creado_en`, `actualizado_en` | `string` ISO | No se muestran en esta historia |

- **No existe el campo `descripcion`** en la respuesta: la tabla de Resumen no
  tiene fila "Descripción".
- Error 404: `{ "error": "PROYECTO_NO_ENCONTRADO", "mensaje": "el proyecto no existe" }`,
  que el cliente tipado convierte en `ApiError` con `status: 404` y
  `code: 'PROYECTO_NO_ENCONTRADO'`.
- Cualquier otro fallo (red caída, HTTP 5xx, JSON inválido) llega como `ApiError`
  con `status` distinto de 404 (o `status: 0` / `NETWORK_ERROR`).

### Elementos de UI (entradas del usuario)

| Elemento | Tipo | Obligatorio | Regla |
| :--- | :--- | :---: | :--- |
| URL `/proyectos/:id` (y sus secciones) | navegación / deep link | Sí | Dispara la carga del proyecto |
| Parámetro `:id` | `string` de `useParams()` | Sí | Se parsea y valida antes de llamar a la API (RB-03) |
| Menú de secciones (`BarraLateralProyecto`) | `NavLink` existente de US-F03 | Sí | Sin cambios; cambiar de sección no refetchea (RB-19) |
| Botón "Reintentar" | `Button` | Solo en estado error | RB-07, RB-08 |
| Enlace "Volver a proyectos" (texto propuesto) | `Link` → `/proyectos` | Solo en estado no encontrado | RB-09 |

### Textos de interfaz (exactos para los tests)

| Texto | Dónde | Rol accesible |
| :--- | :--- | :--- |
| `Cargando proyecto…` | Estado cargando | `role="status"` |
| `No se pudo cargar el proyecto. Intentá nuevamente.` | Estado error (`Alert` tono `error`) | `role="alert"` |
| `Reintentar` | Botón del estado error | `button` |
| `Proyecto no encontrado` | Estado no encontrado | mensaje principal |
| `Volver a proyectos` (propuesto) | Enlace del estado no encontrado | `link` → `/proyectos` |
| `Cerrado` | Chip del encabezado cuando `estado === 'Cerrado'` | texto visible |
| `Sin fecha` | Fila "Fecha de fin" cuando `fecha_fin` es `null` | celda |
| `Este proyecto todavía no tiene integrantes` | Equipo sin integrantes | estado vacío |

Los encabezados de tabla ("Campo", "Valor", "Nombre", "Email", "Registrado el")
y las filas de Resumen se detallan en las reglas de negocio. Los textos marcados
"propuesto" pueden ajustarse en implementación sin cambiar la spec, siempre que
se actualicen tests y feature file en el mismo commit.

## Salidas esperadas

| # | Acción | Resultado | Efecto |
| :---: | :--- | :--- | :--- |
| 1 | Abrir `/proyectos/:id` | "Cargando proyecto…" y, al resolver, Resumen con la tabla campo/valor (Nombre, Estado, Fecha de inicio, Fecha de fin, Cantidad de integrantes) | Lectura |
| 2 | Abrir el detalle de un proyecto sin `fecha_fin` | La fila "Fecha de fin" muestra "Sin fecha" | Lectura |
| 3 | Abrir el detalle de un proyecto cerrado | El encabezado muestra el nombre y el chip "Cerrado" | Lectura |
| 4 | Clic en "Backlog", "Sprints", "Equipo", "Defectos" o "Métricas" | La URL cambia a la sección, el contenido se actualiza sin recargar la página y **sin un nuevo fetch** del proyecto | Navegación |
| 5 | Abrir Equipo con integrantes | Tabla con una fila por integrante: Nombre, Email y "Registrado el" | Lectura |
| 6 | Abrir Equipo sin integrantes | "Este proyecto todavía no tiene integrantes", sin tabla | Lectura |
| 7 | Ingresar el `id` de un proyecto inexistente (API 404) | "Proyecto no encontrado" con enlace a `/proyectos`; sin contenido de sección | Error |
| 8 | Ingresar un `:id` no numérico en la URL | Mismo resultado que el 404, sin llamar a la API | Error |
| 9 | Falla de API distinta de 404 | `Alert` de error con botón "Reintentar" | Error |
| 10 | Clic en "Reintentar" con la API recuperada | Vuelve a "Cargando proyecto…" y termina mostrando el detalle | Error → Lectura |
| 11 | Deep link a `/proyectos/:id/equipo` con proyecto válido | El layout carga el proyecto y muestra Equipo ya en la tabla | Navegación |
| 12 | Deep link a `/proyectos/:id/equipo` con proyecto inexistente | "Proyecto no encontrado" y **sin** el contenido de Equipo debajo | Error |

## Reglas de negocio

- **RB-01** `ProyectoLayout` es el único responsable de pedir el proyecto: hace
  un **único fetch** con `obtenerProyecto` (envuelto en un `ProyectoProvider`)
  y comparte el resultado a las páginas hijas vía React Context con el hook
  `useProyecto()`. Ninguna página de sección llama a la API por su cuenta.
- **RB-02** `obtenerProyecto(proyectoId: number, signal?: AbortSignal):
  Promise<Proyecto>` vive en `frontend/src/api/proyectos.ts`, consume
  `GET /api/proyectos/:id` con el cliente tipado de US-F01 y propaga el
  `signal` para poder cancelar.
- **RB-03** El `:id` llega como `string` por `useParams()`; se parsea a número
  antes de llamar a la API. Si no representa un **entero positivo**
  (`Number.isInteger(Number(id))` y `> 0`), se entra directamente en el estado
  "no encontrado" **sin** llamar a la API.
- **RB-04** El estado del detalle es una unión discriminada
  `{ tipo: 'cargando' } | { tipo: 'error' } | { tipo: 'noEncontrado' } |
  { tipo: 'listo'; proyecto: Proyecto }`, siguiendo el patrón de
  `EstadoCarga` en `ProyectosPage.tsx`.
- **RB-05** En estado `cargando` se muestra el texto "Cargando proyecto…" con
  `role="status"`; no se renderiza ninguna tabla ni contenido previo.
- **RB-06** El fetch se cancela con `AbortController` al desmontar el layout y
  al relanzarlo (cambio de `intento` o de `id`); un `AbortError` se ignora y no
  cambia el estado visible.
- **RB-07** Un error de API distinto de 404 (red, HTTP 5xx, respuesta inválida)
  produce el estado `error`: `Alert` con tono `error` y el texto "No se pudo
  cargar el proyecto. Intentá nuev." acompañado del botón "Reintentar".
- **RB-08** "Reintentar" resetea el estado a `cargando` e incrementa `intento`,
  disparando exactamente **un** nuevo fetch por clic (patrón existente de
  `ProyectosPage`).
- **RB-09** Un 404 de la API (`ApiError` con `status === 404`, típicamente
  `code === 'PROYECTO_NO_ENCONTRADO'`) o un `:id` inválido (RB-03) produce el
  estado `noEncontrado`: se muestra "Proyecto no encontrado" con un enlace a
  `/proyectos`. En este estado **no** se ofrece el botón "Reintentar".
- **RB-10** El `<Outlet />` (contenido de la sección activa) solo se renderiza
  cuando `tipo === 'listo'`; en `cargando`, `error` y `noEncontrado` el layout
  muestra únicamente el estado correspondiente, sin el placeholder de la sección
  debajo.
- **RB-11** El encabezado del detalle (visible en todas las secciones) muestra
  el `nombre` del proyecto y, solo cuando `estado === 'Cerrado'`, la etiqueta
  "Cerrado" al lado del nombre. Los proyectos activos **no** muestran chip en el
  encabezado (el estado "Activo" se lee en la tabla de Resumen).
- **RB-12** Resumen es una tabla de dos columnas (campo/valor) con exactamente,
  y en este orden, las filas: **Nombre, Estado, Fecha de inicio, Fecha de fin,
  Cantidad de integrantes**. No hay fila "Descripción" porque el backend no la
  devuelve.
- **RB-13** Las fechas se muestran en formato `dd/mm/yyyy` con formateo manual
  (el mismo criterio que `formatearFecha` en `ProyectosPage.tsx`: toma los
  primeros 10 caracteres y separa por `-`); la fila "Fecha de fin" muestra
  "Sin fecha" cuando `fecha_fin` es `null`.
- **RB-14** La fila "Cantidad de integrantes" muestra `cantidad_usuarios` tal
  como la devuelve la API; el frontend **no** la recalcula a partir de
  `usuarios.length`.
- **RB-15** Equipo es una tabla con las columnas **Nombre, Email** y
  **"Registrado el"**, con una fila por integrante de `proyecto.usuarios`;
  "Registrado el" es el `creado_en` del usuario formateado con `dd/mm/yyyy`
  (fecha de alta del usuario en el sistema, no de unión al proyecto). No hay
  columna "Rol" porque el endpoint del detalle no la devuelve.
- **RB-16** Si el proyecto no tiene integrantes (`usuarios` vacío), Equipo
  muestra "Este proyecto todavía no tiene integrantes" y **no** renderiza la
  tabla (ni sus encabezados).
- **RB-17** Accesibilidad: ambas tablas usan `<table>`, `<thead>` y `<th>` con
  `scope`; en Resumen los encabezados de columna son "Campo" y "Valor" y cada
  nombre de fila va en `<th scope="row">`; en Equipo los encabezados llevan
  `scope="col"`. Cada pantalla conserva **una sola `<h1>`** (la del `PageHeader`
  de la sección); el encabezado del proyecto del layout se renderiza en un
  nivel inferior (por ejemplo `<h2>` o elemento no-encabezado), sin competir
  con ella.
- **RB-18** Responsive: las tablas van dentro de un contenedor con scroll
  horizontal (`overflow-x-auto` + ancho mínimo, patrón de `ProyectosPage`) para
  que en pantallas chicas se desplace la tabla y no el layout.
- **RB-19** Cambiar de sección (Resumen, Backlog, Sprints, Equipo, Defectos,
  Métricas) dentro del mismo proyecto no dispara un nuevo fetch: el proyecto
  persiste en el contexto mientras `ProyectoLayout` siga montado; la URL cambia
  sin recargar la página (rutas anidadas de US-F03).
- **RB-20** Estilos y dependencias: solo tokens de Tailwind del proyecto y
  componentes existentes (`PageHeader`, `Card`, `Alert`, `Button`); sin colores
  hex, sin librerías nuevas (el formateo de fechas es manual, sin `date-fns` ni
  similar).

## Restricciones

- **Frontend puro:** React + TypeScript + Tailwind + React Router v7 + Vitest
  (React Testing Library + `userEvent`). No se toca el backend Go ni el
  contrato de la API.
- Capas del frontend: el acceso a datos solo en `src/api/` (`obtenerProyecto`);
  el fetch y el estado del detalle viven en el contexto del layout
  (`ProyectoProvider` + `useProyecto`, archivo nuevo propuesto en
  `frontend/src/pages/proyecto/proyectoContext.tsx`; ubicación final a criterio
  del implementador, ej. junto al layout en `src/layouts/`); las páginas de
  `src/pages/proyecto/` **no** importan la API, solo consumen `useProyecto()`.
- `ProyectoLayout` no contiene lógica de presentación de las secciones (solo
  encabezado, estados y `ProyectoProvider`); `ResumenPage` y `EquipoPage` no
  gestionan fetch ni errores (eso lo resuelve el layout).
- No se modifican rutas ni menús de US-F03; no se agregan rutas nuevas.
- `useProyecto()` usado fuera de `ProyectoProvider` debe fallar con un error
  descriptivo (error de programación, no un estado de UI).
- Los textos de interfaz van en español (rioplatense).
- Los tests se ejecutan con `npm test`; además deben pasar `npm run lint` y
  `npm run build`.
- Tests con el patrón del repo: `vi.mock('../api/proyectos')`, `MemoryRouter`
  con rutas anidadas (`/proyectos/:id` → layout → sección), como en
  `frontend/src/pages/ProyectosPage.test.tsx`.

## Casos límite

- `:id` no numérico (`abc`, `1.5`, vacío) o no positivo (`0`, `-3`): se trata
  como no encontrado **sin** llamada a la API (RB-03).
- `fecha_fin` en `null`: la fila muestra "Sin fecha" (RB-13). `fecha_inicio`
  nunca es `null` según el contrato del backend; si lo fuera, el formateador
  devuelve "Sin fecha" por defensiva.
- `creado_en` de los usuarios es un timestamp ISO (`2026-01-01T00:00:00Z`): el
  formateador toma los primeros 10 caracteres y muestra `01/01/2026`.
- `cantidad_usuarios` distinto de `usuarios.length` (inconsistencia del
  backend): Resumen muestra `cantidad_usuarios` y Equipo las filas de
  `usuarios`, sin recalcular nada en el frontend (RB-14).
- Proyecto con `usuarios: []` y `cantidad_usuarios: 0`: estado vacío de Equipo
  (RB-16) y "0" en la fila "Cantidad de integrantes" de Resumen.
- Nombre de proyecto muy largo: el encabezado ajusta sin romper el layout
  (wrap o `truncate`); la tabla mantiene su ancho mínimo con scroll.
- Muchos integrantes en Equipo: la tabla crece hacia abajo y mantiene el scroll
  horizontal en pantallas chicas (RB-18).
- Navegación entre secciones **mientras carga**: el layout no se desmonta al
  cambiar de sección, así que el fetch en curso no se cancela ni se reinicia.
- Doble clic rápido en "Reintentar": cada clic resetea a `cargando` e incrementa
  `intento`; el `AbortController` cancela el fetch anterior antes de relanzar
  (RB-06, RB-08).
- Deep link directo a una sección (ej. `/proyectos/7/equipo`): el layout hace
  el fetch y la sección ya se renderiza con su contenido; el enlace activo del
  menú es el de esa sección (US-F03).
- Proyecto cerrado sin integrantes: chip "Cerrado" en el encabezado y estado
  vacío en Equipo, combinables sin conflicto.
- Respuesta de la API sin `usuarios` (no debería ocurrir por el contrato):
  Equipo cae en el estado vacío en lugar de romper.

## Condiciones de error

| Situación | Origen | Comportamiento esperado |
| :--- | :--- | :--- |
| Proyecto inexistente | API 404 → `ApiError { status: 404, code: 'PROYECTO_NO_ENCONTRADO' }` | "Proyecto no encontrado" + enlace a `/proyectos`; sin "Reintentar" y sin contenido de sección |
| `:id` inválido en la URL | `useParams` + RB-03 | Igual que el 404, sin llamada a la API |
| Fallo de red | `ApiError { status: 0, code: 'NETWORK_ERROR' }` | `Alert` de error + botón "Reintentar" |
| Otro error HTTP (500, 401, 403, …) | `ApiError` con `status ≠ 404` | `Alert` de error + botón "Reintentar" (mensaje genérico, sin exponer detalles) |
| Respuesta JSON inválida | `ApiError { code: 'RESPUESTA_INVALIDA' }` | `Alert` de error + botón "Reintentar" |
| Desmontaje durante la carga | `AbortError` | Se ignora: no cambia el estado visible |
| "Reintentar" presionado y la API sigue fallando | RB-08 | Vuelve a `cargando` y termina de nuevo en el estado error |
| `useProyecto()` fuera de `ProyectoProvider` | Error de programación | Lanza error descriptivo (no es estado de UI) |
| Sin sesión activa | `RutaProtegida` (US-F02) | Redirección a `/login` antes de montar el detalle (heredado, fuera de alcance) |

## Criterios de aceptación

- [ ] Given abro `/proyectos/:id`, When la página carga, Then veo en Resumen una tabla con nombre, estado, fechas y cantidad de integrantes.
- [ ] Given el proyecto no tiene fecha de fin, When veo la fila "Fecha de fin" en Resumen, Then el valor es "Sin fecha".
- [ ] Given el proyecto está cerrado, When miro el encabezado, Then veo la etiqueta "Cerrado".
- [ ] Given estoy en el detalle, When hago clic en "Backlog", "Sprints", "Equipo", "Defectos" o "Métricas", Then la URL cambia a la sección correspondiente y el contenido se actualiza sin recargar la página.
- [ ] Given el proyecto tiene integrantes, When abro Equipo, Then veo una fila por integrante con nombre, email y fecha de registro.
- [ ] Given el proyecto no tiene integrantes, When abro Equipo, Then veo "Este proyecto todavía no tiene integrantes" en lugar de una tabla vacía.
- [ ] Given ingreso el id de un proyecto que no existe, When la API responde 404, Then veo "Proyecto no encontrado" con un enlace a `/proyectos`.
- [ ] Given la API falla por otro motivo, When termina la carga, Then veo un mensaje de error con el botón "Reintentar".

Cada criterio es comprobable con los tests de la sección "Tests previstos" y con
los escenarios del feature file (misma numeración base: CA-n ↔ E-n ↔ test).

## Modelo de datos de dominio y plan de pruebas

### Modelo de datos

No se crean ni modifican tipos de dominio de la API. Los datos ya existen:

| Dato | Origen | Uso en esta historia |
| :--- | :--- | :--- |
| `Proyecto { id, nombre, fecha_inicio, fecha_fin, estado, cantidad_usuarios, usuarios, creado_en, actualizado_en }` | `frontend/src/api/types.ts` / `GET /api/proyectos/:id` | Encabezado, Resumen y Equipo |
| `UsuarioProyecto { id, nombre, email, rol?, creado_en? }` | `frontend/src/api/types.ts` | Filas de la tabla de Equipo |
| `ApiError { status, code, message }` | `frontend/src/api/errors.ts` | Distinguir el 404 del resto de los errores |
| `params.id` | `useParams()` en el layout | Parseo/validación (RB-03) |

Artefactos nuevos **solo de frontend** (no son dominio):

| Artefacto | Forma |
| :--- | :--- |
| `EstadoProyectoDetalle` | Unión discriminada de estados (RB-04) |
| `ProyectoProvider` + `useProyecto()` | React Context: expone `estado` y `reintentar()`; quedó implementado en `frontend/src/pages/proyecto/ProyectoProvider.tsx` (provider con el fetch único) y `frontend/src/pages/proyecto/proyectoContext.ts` (contexto, tipos y hook), siguiendo el patrón de `session/sesion.ts` + `session/SesionContext.tsx` |
| `formatearFecha` compartida | En REFACTOR se extrae de `ProyectosPage.tsx` a una util compartida (propuesto `frontend/src/utils/fechas.ts`, carpeta nueva), con sus propios tests |

### Archivos a crear

| Archivo | Contenido |
| :--- | :--- |
| `frontend/src/pages/proyecto/proyectoContext.ts` | `ProyectoContext`, `EstadoProyectoDetalle`, `ValorProyectoContexto`, `useProyecto` |
| `frontend/src/pages/proyecto/ProyectoProvider.tsx` | `ProyectoProvider`: fetch único con `AbortController` e `intento`, remontado por `key={id}` en el layout |
| `frontend/src/utils/fechas.ts` (propuesto, REFACTOR) | `formatearFecha(fecha: string \| null): string` compartida |
| `frontend/src/utils/fechas.test.ts` | Tests de formateo: fecha, timestamp ISO y `null` |
| `frontend/src/layouts/ProyectoLayout.test.tsx` | Estados, encabezado/chip, gating del `Outlet`, 404, error/reintentar, no-refetch |
| `frontend/src/pages/proyecto/ResumenPage.test.tsx` | Filas de la tabla de Resumen y "Sin fecha" |
| `frontend/src/pages/proyecto/EquipoPage.test.tsx` | Filas/columnas de Equipo y estado vacío |
| `docs/features/US-F06-detalle-proyecto.feature` | Escenarios BDD (Gherkin, idioma `es`) |

### Archivos a modificar

| Archivo | Cambio |
| :--- | :--- |
| `frontend/src/api/proyectos.ts` | Agregar `obtenerProyecto(proyectoId, signal)` (RB-02) |
| `frontend/src/api/proyectos.test.ts` | Tests de `obtenerProyecto`: llamada, `signal`, mapeo a `Proyecto`, propagación de `ApiError` 404 |
| `frontend/src/layouts/ProyectoLayout.tsx` | Encabezado con nombre + chip "Cerrado", `ProyectoProvider`, estados y `<Outlet />` solo en `listo` |
| `frontend/src/pages/proyecto/ResumenPage.tsx` | Reemplazar placeholder por la tabla campo/valor |
| `frontend/src/pages/proyecto/EquipoPage.tsx` | Reemplazar placeholder por la tabla de integrantes + estado vacío |
| `frontend/src/pages/ProyectosPage.tsx` | REFACTOR: extraer `formatearFecha` a la util compartida (sin cambio de comportamiento) |

### Tests previstos (Vitest + React Testing Library)

| Test | Archivo | Verifica |
| :--- | :--- | :--- |
| `obtenerProyecto` llama a `GET /proyectos/:id` con `signal` y devuelve el `Proyecto` | `api/proyectos.test.ts` | RB-02, CA-01 |
| "Cargando proyecto…" con `role="status"` mientras resuelve y sin tabla | `ProyectoLayout.test.tsx` | RB-04, RB-05 |
| Resumen muestra las cinco filas con sus valores (nombre, estado, fechas, cantidad) | `ResumenPage.test.tsx` | RB-12, CA-01 |
| Fila "Fecha de fin" en `null` muestra "Sin fecha" | `ResumenPage.test.tsx` | RB-13, CA-02 |
| Fechas visibles como `dd/mm/yyyy` (incluido `creado_en` ISO de usuarios) | `ResumenPage.test.tsx` / `EquipoPage.test.tsx` | RB-13, RB-15 |
| Chip "Cerrado" en el encabezado solo cuando el proyecto está cerrado | `ProyectoLayout.test.tsx` | RB-11, CA-03 |
| Cambiar a "Equipo" actualiza el contenido sin recargar y **sin** un segundo fetch (mock llamado una vez) | `ProyectoLayout.test.tsx` | RB-19, CA-04 |
| Equipo muestra una fila por integrante con nombre, email y "Registrado el" | `EquipoPage.test.tsx` | RB-15, CA-05 |
| Equipo sin integrantes muestra el estado vacío y no la tabla | `EquipoPage.test.tsx` | RB-16, CA-06 |
| 404 de la API muestra "Proyecto no encontrado" con enlace a `/proyectos` | `ProyectoLayout.test.tsx` | RB-09, CA-07 |
| `:id` no numérico muestra "Proyecto no encontrado" sin llamar a la API | `ProyectoLayout.test.tsx` | RB-03, CA-07 |
| Error genérico muestra `Alert` + "Reintentar"; al reintentar con éxito muestra el detalle | `ProyectoLayout.test.tsx` | RB-07, RB-08, CA-08 |
| En `noEncontrado` no se renderiza el contenido de la sección (deep link a `/equipo`) | `ProyectoLayout.test.tsx` | RB-10, CA-07 |
| Tablas con `table`, `thead`, `columnheader` y `rowheader` accesibles | `ResumenPage.test.tsx` / `EquipoPage.test.tsx` | RB-17, DoD |
| `formatearFecha`: fecha, timestamp ISO y `null` | `utils/fechas.test.ts` | RB-13 (REFACTOR) |

Los tests renderizan el árbol de rutas anidadas (`MemoryRouter` con
`/proyectos/:id` → `ProyectoLayout` → sección) y `vi.mock('../api/proyectos')`,
siguiendo el patrón de `frontend/src/pages/ProyectosPage.test.tsx`. El scroll
horizontal (RB-18) se verifica por clases del contenedor (`overflow-x-auto` +
ancho mínimo) y queda además como escenario manual del feature file.

### Checklist de implementación

- [ ] RED: tests de `obtenerProyecto` en `api/proyectos.test.ts`.
- [ ] GREEN: implementar `obtenerProyecto` en `api/proyectos.ts`.
- [ ] RED: tests del contexto/layout (estados, 404, error, chip, gating).
- [ ] GREEN: crear `proyectoContext.tsx` y completar `ProyectoLayout`.
- [ ] RED: tests de `ResumenPage` (filas, "Sin fecha").
- [ ] GREEN: implementar la tabla campo/valor de Resumen.
- [ ] RED: tests de `EquipoPage` (filas y estado vacío).
- [ ] GREEN: implementar la tabla de Equipo y su estado vacío.
- [ ] REFACTOR: extraer `formatearFecha` a `frontend/src/utils/fechas.ts` con
      sus tests, sin cambiar comportamiento.
- [ ] Escribir `docs/features/US-F06-detalle-proyecto.feature` (escenarios de
      esta spec).
- [ ] Verificar accesibilidad (`table`, `thead`, `th` con `scope`, `role`).
- [ ] Verificar `npm run build`, `npm run lint` y `npm test`.

## Fuera de alcance

- Contenido real de Backlog, Sprints, Defectos y Métricas: quedan placeholder
  hasta sus propias historias.
- Columna "Rol" en Equipo: el endpoint del detalle no la devuelve; requiere un
  ajuste de backend → **A confirmar** para una historia futura.
- Alta/baja de integrantes desde Equipo, edición del proyecto o cierre del
  proyecto desde el detalle.
- Refetch o revalidación automática del proyecto (al volver del navegador, al
  cambiar de pestaña, etc.): el fetch ocurre una vez por montaje del layout.
- Orden, búsqueda, filtros o paginación en la tabla de Equipo.
- Cambios en el backend Go: el endpoint `GET /api/proyectos/:id` ya existe.
- Rutas, menú de secciones y barra lateral (US-F03); listado y navegación al
  detalle (US-F04); sesión y rutas protegidas (US-F02).
- Métricas y gráficos del proyecto.

## Escenarios BDD

Ver `docs/features/US-F06-detalle-proyecto.feature` (E-01 a E-15: resumen en
tabla, "Sin fecha", chip "Cerrado", navegación sin recarga ni refetch, equipo
con y sin integrantes, id inválido, 404 con enlace, error genérico con
"Reintentar" y reintento exitoso; incluye casos límite y de error). Criterio de
aceptación ↔ escenario ↔ test Vitest: misma numeración base, y cada regla de
negocio (RB-01 a RB-20) mapea a al menos un escenario de tipo normal,
alternativo, límite o error.
