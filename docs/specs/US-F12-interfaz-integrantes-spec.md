# US-F12 — Interfaz de integrantes y roles del proyecto

## Objetivo

Implementar la pantalla **Equipo** del proyecto para que un integrante consulte a las
personas del equipo y sus roles, un `LIDER` gestione el alta por email, el cambio de
rol y la baja, y cualquier integrante pueda salir del proyecto.

Consume los endpoints de US-12:

- `GET /api/proyectos/{id}/usuarios`
- `POST /api/proyectos/{id}/usuarios`
- `PATCH /api/proyectos/{id}/usuarios/{usuarioId}`
- `DELETE /api/proyectos/{id}/usuarios/{usuarioId}`

## Entradas

- Ruta `/proyectos/:id/equipo` con sesión activa.
- `rol_actual` y `estado` del proyecto (detalle del proyecto).
- Listado de integrantes con `id`, `nombre`, `email` y `rol`.

## Salidas

- Tabla de integrantes con nombre, email y rol.
- Acciones visibles según `rol_actual`, estado del proyecto y fila.
- Mensajes de éxito y error accesibles.

## Reglas de negocio

- **RB-F12-01:** al entrar se cargan en paralelo el detalle del proyecto y sus
  integrantes. El primero aporta `rol_actual` y `estado`; el segundo, el equipo.
- **RB-F12-02:** un `LIDER` ve el botón "Agregar integrante", el selector de rol por
  fila y la acción "Quitar" sobre las demás filas.
- **RB-F12-03:** un `MIEMBRO` solo puede salir del proyecto (auto-baja); no ve el alta,
  ni el cambio de rol, ni la baja de otras personas.
- **RB-F12-04:** cualquier integrante ve "Salir del proyecto" en su propia fila.
  Si es el último líder, el backend responde `409 ULTIMO_LIDER` y se muestra el mensaje.
- **RB-F12-05:** con el proyecto en estado `Cerrado` la pantalla es de solo lectura y
  muestra un aviso para reabrirlo.
- **RB-F12-06:** el alta por email exige un email con formato válido; el rol por
  defecto es `MIEMBRO`. Si el email no está registrado o ya es integrante, el mensaje
  se muestra junto al campo `email`.
- **RB-F12-07:** el cambio de rol solo se envía si el rol elegido difiere del actual.
  Se actualiza la fila sin recargar la página.
- **RB-F12-08:** quitar a otra persona o salir del proyecto pide confirmación. Al
  completarse, la fila desaparece; la auto-baja navega a `/proyectos`.
- **RB-F12-09:** los errores de red y del servidor no pierden el estado del formulario.
- **RB-F12-10:** `404 PROYECTO_NO_ENCONTRADO` muestra "No tenés acceso a este proyecto
  o no existe" y ofrece reintentar solo ante fallos recuperables.
- **RB-F12-11:** el modal de alta es accesible: `role="dialog"`, `aria-modal`, foco
  inicial, cierre con Escape y clic en el fondo.
- **RB-F12-12:** los roles válidos del sistema son `LIDER` y `MIEMBRO`.

## Casos límite

- El proyecto no tiene integrantes además del usuario: se muestra el equipo existente.
- El líder intenta quitarse siendo el único líder: `409 ULTIMO_LIDER`.
- Cambiar un rol al mismo valor no dispara una petición.
- El email llega con espacios o mayúsculas: el backend lo normaliza; el frontend recorta.
- El proyecto se cierra mientras la pantalla está abierta: el backend responde
  `409 PROYECTO_CERRADO` y se muestra el error sin perder datos.

## Condiciones de error

| Situación | Respuesta esperada en la UI |
|---|---|
| `404 PROYECTO_NO_ENCONTRADO` | "No tenés acceso a este proyecto o no existe" |
| `403 ACCESO_DENEGADO` | mensaje de permisos del backend |
| `409 ULTIMO_LIDER` | "el proyecto debe conservar al menos un líder" |
| `409 USUARIO_DUPLICADO` | mensaje junto al campo `email` |
| `404 USUARIO_NO_ENCONTRADO` | mensaje junto al campo `email` |
| `409 PROYECTO_CERRADO` | mensaje de proyecto cerrado y vista de solo lectura |
| `422 VALIDACION` | mensaje junto al campo indicado (`email` o `rol`) |
| `status 0` | "No se pudo conectar con el servidor. Intentá nuevamente." |

## Criterios de aceptación

- **CA-F12-01:** la pantalla muestra todos los integrantes con su rol.
- **CA-F12-02:** un `LIDER` ve las acciones de alta, cambio de rol y baja.
- **CA-F12-03:** un `MIEMBRO` no ve acciones de gestión sobre otras personas.
- **CA-F12-04:** cualquier integrante puede salir del proyecto desde su fila.
- **CA-F12-05:** el alta por email agrega la persona y refresca el equipo.
- **CA-F12-06:** el cambio de rol actualiza la fila.
- **CA-F12-07:** la baja pide confirmación y quita la fila.
- **CA-F12-08:** los errores de email se muestran en el campo `email`.
- **CA-F12-09:** el último líder no puede degradarse ni quitarse.
- **CA-F12-10:** un proyecto cerrado no permite gestionar el equipo.
- **CA-F12-11:** el modal de alta cumple las pautas de accesibilidad indicadas.
- **CA-F12-12:** no hay roles antiguos en la interfaz: solo `LIDER` y `MIEMBRO`.

## Archivos afectados

- `frontend/src/api/types.ts`
- `frontend/src/api/proyectos.ts`
- `frontend/src/api/proyectos.test.ts`
- `frontend/src/pages/proyecto/EquipoPage.tsx`
- `frontend/src/pages/proyecto/equipo/*` (hook, modal, etiquetas)
- `frontend/src/pages/proyecto/EquipoPage.test.tsx`
- `frontend/src/pages/NuevoProyectoPage.tsx` (rol del creador `LIDER`)
- `frontend/src/pages/NuevoProyectoPage.test.tsx`
