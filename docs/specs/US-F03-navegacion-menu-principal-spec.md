# Especificación: US-F03 - Navegación con el menú principal

## Objetivo

Especificar la **estructura fija de navegación del panel** (frontend React + TypeScript +
Tailwind + React Router v7 + Vitest): una barra superior con la aplicación, el usuario
activo y el botón "Salir", y una barra lateral con los accesos principales del panel.
Al entrar a un proyecto, la barra lateral cambia a las secciones de ese proyecto
(Resumen, Backlog, Sprints, Equipo, Defectos, Métricas), como en Jira.

Se trata de una **historia de frontend**: no se define ningún endpoint ni lógica de
backend Go. La navegación se resuelve con React Router (`NavLink`, rutas anidadas,
`<Outlet />`), el usuario activo proviene de `SesionContext` (US-F02) y no se agregan
librerías de navegación nuevas.

La funcionalidad satisface el issue `SebastianAguilera/ingenieria-y-calidad#52`
(milestone *Sprint 2: Interfaz Objetivo*) y es prerrequisito de las historias de
contenido de sección (US-F04, US-F06 y demás) que se renderizarán dentro de las rutas
anidadas definidas aquí.

**Trazabilidad:** Issue #52 → esta especificación →
`docs/features/US-F03-navegacion-menu-principal.feature` → tests Vitest
(`frontend/src/**/*.test.tsx`) → código. Ciclo TDD (RED → GREEN → REFACTOR).

## Entradas

### Precondiciones

- Sesión activa de US-F02: `SesionContext` expone `usuarioActivo` (`id`, `nombre`,
  `email`, `creado_en`) y `salir()`.
- `RutaProtegida` envuelve las rutas del panel: sin usuario activo se redirige a
  `/login`.
- Cliente HTTP tipado de US-F01 (`frontend/src/api/client.ts`): **solo** se usa si la
  barra o las secciones muestran datos obtenidos de la API; en esta historia las
  páginas de sección son placeholders, por lo que la navegación no depende de llamadas
  a la API.
- Tokens de diseño de `frontend/tailwind.config.js`: `bg-page`, `bg-surface`,
  `text-ink`, `text-ink-soft`, `border-border`, `rounded-pill`, `bg-primary`,
  `text-[var(--on-primary)]`, `font-display`. **No se usan colores hexadecimales
  hardcodeados**.
- Punto de corte responsive: **768px** (breakpoint `md` de Tailwind).
- Existe `frontend/src/layouts/AppShell.tsx` a refactorizar y
  `frontend/src/pages/NotFoundPage.tsx` a ajustar.

### Elementos de UI (entradas del usuario)

| Elemento | Tipo | Obligatorio | Regla |
| :--- | :--- | :---: | :--- |
| Enlace "Proyectos" (barra lateral global) | `NavLink` → `/proyectos` | Sí | Visible en todo el panel |
| Enlace "Usuarios" (barra lateral global) | `NavLink` → `/usuarios` | Sí | Visible en todo el panel |
| Secciones de proyecto (barra lateral de proyecto) | `NavLink` → `/proyectos/:id/...` | Sí | Resumen, Backlog, Sprints, Equipo, Defectos, Métricas |
| Botón "Salir" (barra superior) | `button` | Sí | Ejecuta `salir()` de `SesionContext` y navega a `/login` |
| Usuario activo (barra superior) | texto | Sí | `usuarioActivo.nombre` desde `SesionContext` |
| Botón de menú (barra lateral colapsable) | `button` | Sí | Solo en pantallas < 768px; `aria-expanded`, `aria-controls` y texto alternativo accesible |
| Tema (modo claro/oscuro) | `ModeSwitch` | Sí | Se conserva el componente actual en la barra superior |

### Rutas de entrada

| Ruta | Componente | Barra lateral |
| :--- | :--- | :--- |
| `/proyectos` | `ProyectosPage` | Global |
| `/usuarios` | página de usuarios (placeholder o existente) | Global |
| `/proyectos/:id` | `ResumenPage` (placeholder) | De proyecto |
| `/proyectos/:id/backlog` | `BacklogPage` (placeholder) | De proyecto |
| `/proyectos/:id/sprints` | `SprintsPage` (placeholder) | De proyecto |
| `/proyectos/:id/equipo` | `EquipoPage` (placeholder) | De proyecto |
| `/proyectos/:id/defectos` | `DefectosPage` (placeholder) | De proyecto |
| `/proyectos/:id/metricas` | `MetricasPage` (placeholder) | De proyecto |
| cualquier otra ruta del panel | `NotFoundPage` | Global |

## Salidas esperadas

| # | Acción | Resultado | Efecto |
| :---: | :--- | :--- | :--- |
| 1 | Ver cualquier pantalla del panel | Barra superior con nombre de la aplicación, nombre del usuario activo y botón "Salir" | Ninguno |
| 2 | Ver cualquier pantalla del panel | Barra lateral con "Proyectos" y "Usuarios" (o las secciones del proyecto si la ruta es de proyecto) | Ninguno |
| 3 | Clic en "Usuarios" desde `/proyectos` | URL pasa a `/usuarios` y el enlace queda activo (`aria-current="page"`) | Navegación |
| 4 | Entrar a `/proyectos/:id` | Barra lateral muestra Resumen, Backlog, Sprints, Equipo, Defectos y Métricas | Navegación |
| 5 | Estar en `/proyectos/:id/backlog` | El enlace "Backlog" aparece marcado como activo | Ninguno |
| 6 | Clic en "Proyectos" desde una ruta de proyecto | Se vuelve al listado y la barra lateral vuelve a las secciones globales | Navegación |
| 7 | Pantalla < 768px | Barra lateral colapsada; se abre/cierra con el botón de menú | Ninguno |
| 8 | Recorrer el menú con Tab y Enter | Se puede navegar entre secciones y abrir/cerrar la barra lateral | Navegación |
| 9 | Ingresar a una ruta inexistente | Pantalla 404 con enlace a `/proyectos` | Navegación |
| 10 | Clic en "Salir" | Sesión eliminada y redirigido a `/login` (comportamiento de US-F02) | Navegación |

## Reglas de negocio

- **RB-01** `AppShell` es el layout fijo del panel: barra superior + barra lateral +
  área de contenido (`<Outlet />`); todas las rutas protegidas se renderizan dentro.
- **RB-02** La barra superior muestra el nombre de la aplicación, el nombre del usuario
  activo tomado de `SesionContext` y el botón "Salir" (US-F02: limpia la sesión y
  navega a `/login`).
- **RB-03** La barra lateral global contiene exactamente los accesos "Proyectos"
  (`/proyectos`) y "Usuarios" (`/usuarios`).
- **RB-04** La barra lateral de proyecto contiene exactamente las secciones "Resumen",
  "Backlog", "Sprints", "Equipo", "Defectos" y "Métricas" bajo `/proyectos/:id`.
- **RB-05** Las rutas anidadas bajo `/proyectos/:id` se resuelven con un layout
  `ProyectoLayout` que renderiza `<Outlet />`:
  `/proyectos/:id` → Resumen, `/proyectos/:id/backlog` → Backlog,
  `/proyectos/:id/sprints` → Sprints, `/proyectos/:id/equipo` → Equipo,
  `/proyectos/:id/defectos` → Defectos, `/proyectos/:id/metricas` → Métricas.
- **RB-06** El enlace de la sección que coincide con la ruta actual se marca como
  activo mediante `NavLink` (o `useLocation`) y lleva `aria-current="page"`.
- **RB-07** En pantallas con menos de 768px la barra lateral inicia colapsada y se
  abre/cierra con un botón de menú accesible (`aria-expanded`, `aria-controls` y
  texto alternativo).
- **RB-08** Toda la navegación del menú es operable con teclado (Tab para recorrer,
  Enter para activar) y el foco es visible.
- **RB-09** Las páginas de sección son placeholders en esta historia; el contenido
  real se implementa en historias posteriores.
- **RB-10** Cualquier ruta inexistente dentro del panel muestra `NotFoundPage` con un
  enlace a `/proyectos`.
- **RB-11** No se agrega ninguna librería de navegación: se mantiene React Router v7.
- **RB-12** Los estilos usan tokens de Tailwind del proyecto; no se hardcodean colores
  hexadecimales.

## Restricciones

- **Frontend puro:** React + TypeScript + Tailwind + React Router v7 + Vitest
  (React Testing Library + `userEvent`). No se toca el backend Go.
- Clean Architecture del frontend: `AppShell` y `ProyectoLayout` son layouts en
  `src/layouts/`; las secciones son páginas en `src/pages/`; la sesión sigue en
  `src/session/`.
- `AppShell` no debe contener lógica de negocio de proyectos ni llamadas a la API;
  solo composición de layout y lectura de `SesionContext`.
- Reutilizar los componentes existentes (`ModeSwitch`, `PageHeader`, `Card`) sin
  duplicar estilos; colores y radios solo desde tokens de `tailwind.config.js`.
- El botón de menú debe usar `aria-expanded` + `aria-controls` apuntando al id del
  panel lateral, y un nombre accesible (texto o `aria-label`) en español.
- Los textos de interfaz van en español.
- Los tests se ejecutan con `npm test`; además deben pasar `npm run lint` y
  `npm run build`.

## Casos límite

- Ruta `/proyectos/:id` con `id` no numérico o inexistente: el layout de proyecto
  sigue renderizando (el detalle/validación del proyecto es de US-F06); en esta
  historia el placeholder no depende de que el proyecto exista.
- Sin sesión activo: `RutaProtegida` redirige a `/login` antes de renderizar
  `AppShell` (comportamiento heredado de US-F02).
- Pantalla exactamente de 768px: se considera ancho ≥ 768px, es decir la barra
  lateral visible (el corte es "menores a 768px").
- Barra lateral colapsada y cambio de ruta: el estado abierto/cerrado no altera la
  ruta activa; el enlace activo se recalcula según la URL.
- Navegación con el teclado en la barra lateral colapsada: el botón de menú debe ser
  alcanzable por Tab para poder abrir el panel.
- Usuario activo nulo dentro de `AppShell` (no debería ocurrir por `RutaProtegida`):
  la barra superior no muestra nombre y "Salir" sigue funcionando.
- Rutas anidadas de proyecto abiertas directamente por URL (deep link): deben
  renderizar `ProyectoLayout` con el enlace de la sección correspondiente activo.
- Varias secciones abiertas en distintos `id`: el enlace activo se calcula por
  pathname completo, no solo por el nombre de la sección.

## Condiciones de error

| Situación | Origen | Comportamiento esperado |
| :--- | :--- | :--- |
| Ruta inexistente dentro del panel | React Router (`path="*"`) | `NotFoundPage` con enlace a `/proyectos` |
| Sin usuario activo | `RutaProtegida` | Redirección a `/login` (US-F02) |
| Sesión inválida en `localStorage` | `leerSesionGuardada` | Se ignora y se redirige a `/login` (US-F02) |
| `id` de proyecto inválido en la URL | Rutas anidadas | Se muestra el placeholder de la sección; sin pantalla de error en esta historia (validación real en US-F06) |
| Botón "Salir" presionado | `SesionContext.salir` | Sesión limpiada y navegación a `/login` |
| Errores de API de las secciones | Historias posteriores | Fuera de alcance: las páginas son placeholders |

## Criterios de aceptación

- [ ] Given estoy logueado, When miro cualquier pantalla del panel, Then veo la barra superior con mi nombre y la barra lateral con "Proyectos" y "Usuarios".
- [ ] Given estoy en `/proyectos`, When hago clic en "Usuarios", Then la URL cambia a `/usuarios` y ese enlace queda marcado como activo.
- [ ] Given entro a `/proyectos/:id`, When se muestra la barra lateral, Then aparecen las secciones Resumen, Backlog, Sprints, Equipo, Defectos y Métricas del proyecto.
- [ ] Given estoy en `/proyectos/:id/backlog`, When miro la barra lateral, Then "Backlog" queda marcado como activo.
- [ ] Given estoy en una ruta de proyecto, When hago clic en "Proyectos", Then vuelvo al listado y la barra lateral vuelve a las secciones globales.
- [ ] Given la pantalla tiene menos de 768px de ancho, When miro la barra lateral, Then está colapsada y se abre con un ícono de menú.
- [ ] Given uso solo el teclado, When recorro el menú con Tab y Enter, Then puedo navegar entre secciones y abrir/cerrar la barra lateral.
- [ ] Given ingreso a una ruta inexistente, When la app carga, Then veo una pantalla 404 con un enlace a `/proyectos`.

## Modelo de datos de dominio y plan de pruebas

### Modelo de datos

No se crean ni modifican modelos de dominio nuevos. Los datos que consume el layout ya
existen:

| Dato | Origen | Uso en esta historia |
| :--- | :--- | :--- |
| `Usuario { id, nombre, email, creado_en }` | `frontend/src/api/types.ts` | Nombre del usuario activo en la barra superior |
| `usuarioActivo`, `salir()` | `SesionContext` / `useSesion()` | Barra superior y botón "Salir" |
| `params.id` | `useParams()` en `ProyectoLayout` | Prefijo de las rutas de sección del proyecto |
| `pathname` | `useLocation()` / `NavLink` | Estado activo de los enlaces (`aria-current="page"`) |

Constantes propuestas (en `src/layouts/` o junto a las rutas): listas inmutables de
secciones del proyecto (`{ etiqueta, ruta }`) para generar la barra lateral y sus
tests sin duplicar cadenas.

### Archivos a crear

| Archivo | Contenido |
| :--- | :--- |
| `frontend/src/layouts/ProyectoLayout.tsx` | Layout anidado con barra lateral de proyecto y `<Outlet />` |
| `frontend/src/pages/proyecto/ResumenPage.tsx` | Placeholder de Resumen |
| `frontend/src/pages/proyecto/BacklogPage.tsx` | Placeholder de Backlog |
| `frontend/src/pages/proyecto/SprintsPage.tsx` | Placeholder de Sprints |
| `frontend/src/pages/proyecto/EquipoPage.tsx` | Placeholder de Equipo |
| `frontend/src/pages/proyecto/DefectosPage.tsx` | Placeholder de Defectos |
| `frontend/src/pages/proyecto/MetricasPage.tsx` | Placeholder de Métricas |
| `frontend/src/layouts/AppShell.test.tsx` | Tests de barra superior, barra lateral global/activa, colapso responsive, teclado y 404 |
| `frontend/src/layouts/ProyectoLayout.test.tsx` | Tests de secciones de proyecto y enlace activo por ruta (o integrados en `AppShell.test.tsx`) |
| `docs/features/US-F03-navegacion-menu-principal.feature` | Escenarios BDD (Gherkin, idioma `es`) |

### Archivos a modificar

| Archivo | Cambio |
| :--- | :--- |
| `frontend/src/layouts/AppShell.tsx` | Refactor a layout fijo: barra superior (app + usuario + Salir), barra lateral colapsable y `<Outlet />` |
| `frontend/src/App.tsx` | Rutas anidadas `/proyectos/:id` con `ProyectoLayout`; rutas de placeholders |
| `frontend/src/pages/NotFoundPage.tsx` | Verificar/añadir el enlace a `/proyectos` dentro del layout |
| `frontend/src/pages/ProyectosPage.tsx` | Ajustes mínimos de contenido si el layout cambia (sin lógica nueva) |

### Tests previstos (Vitest + React Testing Library)

| Test | Archivo | Verifica |
| :--- | :--- | :--- |
| Barra superior visible en cualquier ruta del panel | `AppShell.test.tsx` | RB-01, RB-02, CA-01 |
| Barra lateral global: "Proyectos" y "Usuarios" | `AppShell.test.tsx` | RB-03, CA-01 |
| Clic en "Usuarios" cambia la URL y marca el activo (`aria-current="page"`) | `AppShell.test.tsx` | RB-06, CA-02 |
| Secciones de proyecto en `/proyectos/:id` | `ProyectoLayout.test.tsx` | RB-04, RB-05, CA-03 |
| "Backlog" activo en `/proyectos/:id/backlog` | `ProyectoLayout.test.tsx` | RB-06, CA-04 |
| Clic en "Proyectos" vuelve al listado y a la barra global | `AppShell.test.tsx` | CA-05 |
| Barra lateral colapsada por defecto y botón con `aria-expanded`/`aria-controls` | `AppShell.test.tsx` | RB-07, CA-06 |
| Navegación por teclado (Tab + Enter) y foco visible | `AppShell.test.tsx` | RB-08, CA-07 |
| Ruta inexistente muestra 404 con enlace a `/proyectos` | `AppShell.test.tsx` | RB-10, CA-08 |
| Botón "Salir" limpia la sesión y navega a `/login` | `AppShell.test.tsx` | RB-02 (regresión US-F02) |

El ancho < 768px se simula en jsdom redefiniendo `window.innerWidth` y disparando
`window.resizeEvent` (o con `matchMedia` según el patrón adoptado en el repo); no se
requiere un navegador real.

### Checklist de implementación

- [ ] Refactorizar `AppShell` con barra superior, barra lateral y área de contenido.
- [ ] Mostrar el usuario activo y el botón "Salir" en la barra superior.
- [ ] Crear la barra lateral global con "Proyectos" y "Usuarios".
- [ ] Crear `ProyectoLayout` y la barra lateral de proyecto con sus secciones.
- [ ] Definir las rutas anidadas bajo `/proyectos/:id`.
- [ ] Implementar el estado activo de los enlaces y `aria-current="page"`.
- [ ] Implementar el colapso responsive de la barra lateral (<768px) con botón accesible.
- [ ] Verificar navegación por teclado y foco visible.
- [ ] Actualizar `NotFoundPage` con enlace a `/proyectos`.
- [ ] Escribir tests de enlace activo, secciones de proyecto y 404.
- [ ] Verificar `npm run build`, `npm run lint` y `npm test`.

## Fuera de alcance

- Contenido real de las secciones (Resumen, Backlog, Sprints, Equipo, Defectos,
  Métricas): historias posteriores (US-F04, US-F06, etc.).
- Listado y detalle de proyectos (US-F04 / US-F06).
- Módulo de gestión de usuarios (US-08) y usuarios de un proyecto ("Equipo").
- Backend Go: sin endpoints nuevos ni cambios en la API.
- Autenticación real, permisos por rol o guardas de autorización por sección.
- Editor de apariencia y gráficos del design system.
- Cambios en el comportamiento de sesión de US-F02 (login, persistencia, "Salir").
- Agregar librerías de navegación o de iconos nuevas (salvo que el repo ya las
  tenga; los íconos pueden ser SVG inline o los ya existentes).

## Escenarios BDD

Ver `docs/features/US-F03-navegacion-menu-principal.feature`
(E-01 a E-08: barra global, enlace activo, secciones de proyecto, colapso < 768px,
teclado y 404). Criterio de aceptación ↔ escenario ↔ test Vitest: misma numeración
base.
