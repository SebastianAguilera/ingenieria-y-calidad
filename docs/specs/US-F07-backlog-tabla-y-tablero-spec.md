# Especificación: US-F07 - Backlog con vista de tabla y tablero

## Objetivo

Implementar la pantalla **Backlog** del detalle de un proyecto con dos vistas sincronizadas:

- **Tabla** estilo GitHub, con una fila por historia y columnas de código, título, estado y prioridad.
- **Tablero** kanban con tres columnas (`Backlog`, `En Progreso`, `Hecho`).

La pantalla consume los endpoints de US-02 (`GET`, `POST` y `PATCH /api/proyectos/{id}/historias`) y permite listar, crear, editar y mover historias sin recargar la página.

## Entradas

### Precondiciones

- El usuario está autenticado y dentro de un proyecto existente (`/proyectos/:id/backlog`).
- El backend de US-02 está disponible y devuelve historias con `id`, `codigo`, `titulo`, `descripcion`, `prioridad`, `estado`, `story_points`, `esfuerzo_real`, `creado_en` y `actualizado_en`.
- El `id` de la ruta es un entero positivo.

### Datos de entrada de la pantalla

| Acción | Entrada |
| :--- | :--- |
| Listar | `GET /api/proyectos/{id}/historias` |
| Crear | `POST /api/proyectos/{id}/historias` con `titulo`, `descripcion`, `prioridad` y `estado` |
| Editar | `PATCH /api/proyectos/{id}/historias/{historiaId}` con los campos modificados |
| Mover en tablero | `PATCH /api/proyectos/{id}/historias/{historiaId}` con `{ "estado": "<nuevo estado>" }` |

## Salidas esperadas

### Vista Tabla

- Encabezados: **Código**, **Título**, **Estado**, **Prioridad** y **Acciones**.
- Cada fila muestra el código `US-NN`, el título y dos selectores inline para estado y prioridad.
- El botón **Editar** abre el modal con los datos de la historia.
- La tabla tiene scroll horizontal en pantallas angostas.

### Vista Tablero

- Tres columnas con encabezado y contador: **Backlog**, **En Progreso** y **Hecho**.
- Cada tarjeta muestra código, título y prioridad.
- La tarjeta se puede arrastrar a otra columna o mover con un selector accesible por teclado.
- Las dos vistas comparten el mismo estado: un cambio en tabla se refleja en tablero y viceversa.

### Modal de alta/edición

- Campos: **Título**, **Descripción**, **Prioridad** y **Estado**.
- Validación cliente: título obligatorio, entre 1 y 200 caracteres después de `TrimSpace`.
- Los errores de validación del backend se muestran junto al campo indicado en `campo`.
- Se cierra con **Cancelar**, `Escape` o clic fuera del diálogo.

### Estados de pantalla

| Estado | Comportamiento |
| :--- | :--- |
| Cargando | Texto `Cargando historias…` con `role="status"`, sin tabla ni tablero |
| Error | `Alert` de error y botón **Reintentar** |
| Vacío | Mensaje `Todavía no hay historias` y botón **Nueva historia** |
| Listo | Barra con selector **Tabla/Tablero**, botón **Nueva historia** y la vista activa |

## Reglas de negocio

| ID | Regla |
| :--- | :--- |
| RB-01 | El título es obligatorio y se valida sobre el resultado de `TrimSpace`. |
| RB-02 | El título admite entre 1 y 200 caracteres. |
| RB-03 | La prioridad solo puede ser `Alta`, `Media` o `Baja`. |
| RB-04 | El estado solo puede ser `Backlog`, `En Progreso` o `Hecho`. |
| RB-05 | El alta usa `Media` y `Backlog` como valores por defecto. |
| RB-06 | Mover una tarjeta a la columna en la que ya está no genera una petición. |
| RB-07 | Un cambio inline es optimista: si el PATCH falla, la fila o tarjeta vuelve al valor anterior y se muestra el error. |
| RB-08 | Mientras una historia se guarda, sus controles se deshabilitan para evitar envíos duplicados. |
| RB-09 | El listado se muestra en el orden devuelto por el backend (orden de alta). |
| RB-10 | No se muestra ni se edita `story_points` ni `esfuerzo_real` en esta historia. |

## Casos límite

| ID | Caso | Resultado |
| :--- | :--- | :--- |
| CL-01 | Proyecto sin historias | Estado vacío con acción para crear |
| CL-02 | Título vacío o solo espacios | Error de validación, sin llamar al backend |
| CL-03 | Título de 200 caracteres | Válido |
| CL-04 | Título de 201 caracteres | Error de validación |
| CL-05 | PATCH falla por red | Se revierte el cambio optimista y se muestra error |
| CL-06 | PATCH devuelve error de validación con `campo` | El modal muestra el error junto al campo |
| CL-07 | Se suelta una tarjeta en su misma columna | No se llama a la API |
| CL-08 | `id` de ruta inválido | Se muestra el error de proyecto no encontrado |
| CL-09 | Backend devuelve 404 de proyecto | Se muestra `El proyecto no existe` |

## Condiciones de error

| Situación | Mensaje visible |
| :--- | :--- |
| Error de red al listar | `No se pudieron cargar las historias. Intentá nuevamente.` |
| Error de red al crear | `No se pudo crear la historia. Intentá nuevamente.` |
| Error de red al actualizar | `No se pudo actualizar la historia. Intentá nuevamente.` |
| Proyecto inexistente | `El proyecto no existe` |
| Historia inexistente | `La historia no existe` |
| Validación de campo | Mensaje del backend junto al campo |

## Criterios de aceptación

- **CA-I1.** La pantalla lista las historias del proyecto con estados de carga, error, vacío y listo.
- **CA-I2.** La vista Tabla muestra código, título, estado y prioridad por historia.
- **CA-I3.** Estado y prioridad se pueden modificar inline y el cambio persiste.
- **CA-I4.** Se puede crear una historia desde el modal y aparece en ambas vistas.
- **CA-I5.** Se puede editar título, descripción, prioridad y estado desde el modal.
- **CA-I6.** La vista Tablero agrupa las historias en las tres columnas del dominio.
- **CA-I7.** Una tarjeta se puede mover con drag & drop y con un control accesible por teclado.
- **CA-I8.** Los cambios optimistas se revierten si la API falla.
- **CA-I9.** La interfaz es navegable por teclado y usa etiquetas accesibles en los controles.
- **CA-A1.** `npm run test`, `npm run lint` y `npm run build` pasan sin errores.

## Trazabilidad

Esta especificación se vincula con los escenarios de `docs/features/US-F07-backlog-tabla-y-tablero.feature`, con los tests de `frontend/src/pages/proyecto/BacklogPage.test.tsx`, `ModalHistoria.test.tsx`, `TableroBacklog.test.tsx` y `frontend/src/api/historias.test.ts`, y con los endpoints de US-02.
