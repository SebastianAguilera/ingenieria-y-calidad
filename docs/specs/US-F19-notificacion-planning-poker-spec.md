# Especificación: US-F19 - Notificación de sesiones activas de Planning Poker

## Historia de usuario

**Como** miembro de un proyecto, **quiero** ver una campana cuando se abre un
Planning Poker, **para** enterarme y unirme sin depender de que alguien me avise.

## Descripción

US-F19 agrega un indicador tipo campana en `ProyectoLayout`, con badge, lista de
sesiones activas y un toast cuando aparece una sesión nueva. Usa un único polling
cada 3 segundos, compartido por todo el layout, y pausa cuando la pestaña está
oculta. La lista incluye sesiones `Votando` y `Revelada`; el badge cuenta solo
`Votando` donde `ya_vote = false`.

La historia satisface el issue `SebastianAguilera/ingenieria-y-calidad#85`.
**Requiere** US-14 (endpoint `activas`) y US-F15 (ruta y pantalla de Planning
Poker). **No requiere** US-11 para permisos.

**Trazabilidad:** Issue #85 → SDD → criterios de aceptación → escenarios BDD
(`docs/features/US-F19-notificacion-planning-poker.feature`) → tests Vitest/RTL
→ código React.

## Cross-referencias

- **US-14**: `GET /api/proyectos/{id}/poker-sessions/activas` devuelve las
  sesiones `Votando` y `Revelada` del proyecto. Nunca expone votos ajenos, solo
  conteos (`votos_registrados`, `total_miembros`) y `ya_vote`.
- **US-F15**: ruta `/proyectos/:id/backlog/:historiaId/poker` y distintivo
  `Planning Poker en curso` en el backlog, hoy con consulta propia.
- **US-F16/F17/F18**: estados de sesión `Votando`, `Revelada`, `Cerrada` y
  `Cancelada`. Solo `Votando` y `Revelada` son activos.
- **Sesión autenticada**: `useSesion()` de `session/SesionContext` identifica al
  usuario actual para excluir sus propios avisos.

## Objetivo

Que cualquier miembro del proyecto abierto en el navegador se entere de las
sesiones de Planning Poker activas sin recargar la página, pueda verlas en una
lista accesible desde la campana y unirse con un clic, sin duplicar polling en
cada pantalla.

## Alcance

- Crear `ProveedorSesionesPoker` en `ProyectoLayout` que consulte
  `GET /api/proyectos/{id}/poker-sessions/activas`.
- Un único polling cada 3 segundos para todo el layout; no un polling por
  componente.
- Pausar el polling cuando `document.hidden` sea `true`; reanudar al volver.
- Evitar solicitudes superpuestas con `AbortController`.
- Renderizar `CampanaPoker` en `ProyectoLayout`, alineada a la derecha.
- Badge: cantidad de sesiones `Votando` con `ya_vote = false`.
- Lista de sesiones con código, título, ronda, creador, estado y si ya votó.
- Para `Revelada`, mostrar `Revelada` y permitir navegar a ver resultados; no
  suma al badge ni genera toast.
- Click en una sesión navega a
  `/proyectos/:id/backlog/:historiaId/poker`.
- Toast `Se abrió Planning Poker para US-XX: título` cuando aparece una sesión
  nueva.
- No mostrar toast en la primera carga.
- No repetir toast para sesiones ya vistas.
- No mostrar toast al creador de la sesión.
- Si aparecen varias sesiones nuevas, mostrarlas una por una.
- `aria-live="polite"` para el toast; campana con `aria-label`, `aria-expanded`
  y `aria-controls`.
- Lista navegable por teclado; cierre con Escape y click fuera.
- Refactorizar el distintivo del backlog de `US-F15` para consumir el contexto
  compartido y eliminar su consulta propia.
- Tests Vitest/RTL con fake timers: polling, pausa por visibilidad, toast,
  badge, lista y navegación.

## Fuera de alcance

- Notificaciones entre proyectos.
- Notificaciones persistentes o históricas.
- Push notifications, email, WebSocket o SSE.
- Sonidos, vibración o badges del sistema operativo.
- Cambiar los contratos o el cálculo del backend de US-14.
- Permisos especiales: cualquier miembro del proyecto recibe las notificaciones.

## Entradas

- Proyecto activo en la URL: `/proyectos/:id/*`.
- `GET /api/proyectos/{id}/poker-sessions/activas` con JWT administrado por
  `api/client`.
- Respuesta: arreglo de `SesionPokerActiva` (`sesion_id`, `estado`,
  `ronda_actual`, `creado_por`, `creado_en`, `votos_registrados`,
  `total_miembros`, `ya_vote`, `historia`).
- Usuario actual desde `useSesion()`.
- Función de API existente:
  `listarSesionesActivas(proyectoId, signal?): Promise<SesionPokerActiva[]>`.

## Salidas esperadas

### `ProveedorSesionesPoker`

- Mantiene en estado la última lista de sesiones activas y expone por contexto:
  - `proyectoId`: número del proyecto activo o `null`.
  - `sesiones`: última lista recibida (se conserva ante errores).
  - `historiasConPoker`: conjunto de `historia.id` con sesión activa, para el
    distintivo del backlog.
- Consulta inmediatamente al montar y luego cada 3 segundos.
- Si la pestaña está oculta, no programa la siguiente consulta ni realiza
  solicitudes por intervalo.
- Al volver a ser visible, consulta de inmediato y reanuda el intervalo.
- Ante error de red o HTTP, conserva la última lista y reintenta en el próximo
  poll; no muestra toast de error.
- Aborta la solicitud anterior antes de iniciar una nueva.

### `CampanaPoker`

- Botón nativo con `aria-label="Sesiones de Planning Poker activas"`,
  `aria-expanded` y `aria-controls` apuntando a la lista.
- Badge visible solo cuando hay sesiones `Votando` con `ya_vote = false`; su
  número es la cantidad de esas sesiones.
- Al abrir, muestra la lista con `role="region"` o `role="list"` y nombre
  accesible. Si no hay sesiones, muestra `No hay sesiones de Planning Poker
  activas.`.
- Cada ítem es un botón navegable por teclado con:
  - `historia.codigo` y `historia.titulo`.
  - `Ronda {ronda_actual}`.
  - `Creada por {creado_por.nombre}`.
  - Estado: `Votando` o `Revelada`.
  - Indicador `Ya votaste` / `Falta tu voto` para `Votando`.
- Al hacer click en un ítem, cierra la lista y navega a
  `/proyectos/{proyectoId}/backlog/{historia.id}/poker`.
- Escape cierra la lista y devuelve el foco a la campana.
- Click fuera cierra la lista.
- El foco inicial al abrir cae en el primer ítem o, si no hay, en el contenedor.

### Toast de sesión nueva

- Texto exacto: `Se abrió Planning Poker para {codigo}: {titulo}`.
- Se anuncia con `aria-live="polite"` y `role="status"`.
- No aparece en la primera carga con sesiones existentes.
- No se repite para una `sesion_id` ya vista.
- No aparece cuando `creado_por.id` es el `usuarioActivo.id`.
- Solo se genera para sesiones `Votando`; `Revelada` no genera toast.
- Si hay varias sesiones nuevas, se encolan y se muestran una por una, con
  auto-cierre antes de pasar a la siguiente.

### Backlog

- `BacklogPage` deja de llamar a `listarSesionesActivas` por su cuenta y consume
  `historiasConPoker` del contexto compartido.
- El distintivo `Planning Poker en curso` / `En votación` se mantiene para las
  historias con sesión activa.
- Si el contexto no tiene sesiones, el backlog se muestra igual sin distintivos.

## Reglas de negocio

- **RB-01**: un único polling por layout de proyecto, cada 3000 ms.
- **RB-02**: el polling se pausa con `document.hidden === true` y se reanuda al
  volver a visible.
- **RB-03**: las solicitudes se cancelan con `AbortController`; no se superponen.
- **RB-04**: la lista muestra sesiones `Votando` y `Revelada`.
- **RB-05**: el badge cuenta solo sesiones `Votando` con `ya_vote = false`.
- **RB-06**: una sesión `Revelada` aparece como `Revelada`, no suma al badge y no
  genera toast.
- **RB-07**: solo se genera toast para sesiones nuevas `Votando`, nunca en la
  primera carga ni para sesiones ya vistas.
- **RB-08**: el creador de la sesión no recibe toast de su propia sesión.
- **RB-09**: los toasts se muestran de a uno, en orden de aparición.
- **RB-10**: un error de polling conserva la última lista y no interrumpe la UI.
- **RB-11**: la campana es accesible por teclado y anuncia su estado con
  `aria-expanded`.
- **RB-12**: el backlog consume el contexto compartido; no mantiene polling
  propio de sesiones activas.

## Condiciones de error

| Situación | Comportamiento de la UI |
| :--- | :--- |
| Error de red en el polling | Conserva la última lista, no muestra toast de error y reintenta en el próximo intervalo. |
| `401 TOKEN_REQUERIDO` / `TOKEN_INVALIDO` | Comportamiento global de `api/client`: cierra la sesión y redirige al login. |
| `404 PROYECTO_NO_ENCONTRADO` | Conserva la última lista y reintenta. El layout no rompe la pantalla del proyecto. |
| Respuesta nula o vacía | Lista vacía; sin badge y sin toasts. |
| La pestaña se oculta durante una solicitud | La solicitud en curso se conserva; al volver se cancela si quedó pendiente y se consulta de nuevo. |

## Casos límite

- Primera carga con varias sesiones activas: se listan y generan badge, pero no
  hay toast.
- Una sesión `Votando` que pasa a `Revelada`: sigue en la lista, deja de sumar al
  badge y no repite toast.
- Una sesión creada por el usuario actual: se lista, suma al badge si
  corresponde, pero no genera toast.
- Dos sesiones nuevas en el mismo poll: se muestran dos toasts consecutivos.
- Una sesión desaparece y vuelve con el mismo `sesion_id`: no se vuelve a
  notificar.
- El proyecto cambia de `:id`: se reinician lista, badge, sesiones vistas y
  toasts.
- La pestaña permanece oculta más de un intervalo: no se acumulan solicitudes.
- `total_miembros` puede ser 0: la lista se muestra con los conteos disponibles.
- La sesión no tiene `historia` válida: el backend no la devuelve; la UI no debe
  romperse ante una lista vacía.

## Criterios de aceptación

- [ ] **CA-1**: Given el layout de un proyecto, When se monta, Then hay un único
  polling cada 3 segundos a `activas`.
- [ ] **CA-2**: Given la pestaña oculta, When pasa el intervalo, Then no se hacen
  solicitudes.
- [ ] **CA-3**: Given una sesión nueva `Votando`, When aparece en el polling,
  Then se muestra el toast y el badge aumenta si no voté.
- [ ] **CA-4**: Given la primera carga con sesiones existentes, When se monta,
  Then no se muestra toast.
- [ ] **CA-5**: Given una sesión ya vista, When sigue activa en polls
  posteriores, Then no se repite el toast.
- [ ] **CA-6**: Given una sesión creada por el usuario actual, When aparece,
  Then no se muestra toast al creador.
- [ ] **CA-7**: Given una sesión `Revelada`, When se lista, Then aparece con
  estado `Revelada` y no suma al badge.
- [ ] **CA-8**: Given una sesión en la lista, When se hace clic, Then se navega a
  la ruta de esa historia.
- [ ] **CA-9**: Given el toast, When aparece, Then se anuncia con
  `aria-live="polite"`.
- [ ] **CA-10**: Given la campana, When se abre con teclado, Then la lista es
  navegable y se cierra con Escape.
- [ ] **CA-11**: Given el backlog, When hay sesión activa, Then el distintivo
  `Planning Poker en curso` usa el contexto compartido.

## Tareas de implementación

- [ ] Escribir spec SDD (`docs/specs/US-F19-notificacion-planning-poker-spec.md`)
  y BDD (`docs/features/US-F19-notificacion-planning-poker.feature`).
- [ ] RED: tests del provider, polling, visibilidad, toast, badge y campana.
- [ ] GREEN: provider, campana, lista, toast y navegación.
- [ ] Refactorizar el distintivo del backlog para consumir el provider.
- [ ] Verificación manual con dos pestañas y tests.

## Definición de terminado

- [ ] `npm test`, `npm run lint` y `npm run build` en verde.
- [ ] Spec SDD y escenarios BDD escritos antes del código.
- [ ] Verificación manual con dos pestañas: abrir sesión en una y ver
  toast/badge en la otra.
- [ ] Verificación de pausa al ocultar la pestaña.
- [ ] Accesibilidad revisada por teclado, foco, `aria-expanded`, `aria-controls`
  y `aria-live`.

## Dependencias

- **Requiere US-14** (bloqueante): endpoint `GET .../poker-sessions/activas`.
- **Requiere US-F15** (bloqueante): ruta y pantalla de Planning Poker.
- **No requiere US-11** para permisos.

## Notas

- El polling de la campana es independiente del polling de la sesión, pero debe
  haber uno solo por layout.
- Si el endpoint falla, mantener la última lista y reintentar en el próximo
  poll; no mostrar toast de error.
- Solo avisa con el proyecto abierto en el navegador; push o email quedan fuera
  de alcance.
- La duración de cada toast es de 5 segundos; con varias sesiones nuevas se
  encolan y se muestran una por una.
