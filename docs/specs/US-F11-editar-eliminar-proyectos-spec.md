# Especificación: US-F11 - Editar y eliminar proyectos desde la interfaz

## Objetivo

Permitir a los miembros de un proyecto editar sus datos (nombre, descripción, fechas y estado), consultar su información general y eliminarlo desde el listado o desde el detalle, manteniendo la información al día y quitando proyectos que ya no sirven.

Complementa US-01 (API de proyectos) y US-F10 (acceso por membresía). Reutiliza los endpoints existentes; no se crean endpoints nuevos.

**Trazabilidad:** Issue #67 → esta especificación → criterios de aceptación → `docs/features/US-F11-editar-eliminar-proyectos.feature` → tests Go/Vitest → código.

## Entradas

### Precondiciones

- US-F10 en marcha: guard de membresía y listado filtrado por usuario autenticado.
- Backend expone `GET/POST/PUT/PATCH/DELETE /api/proyectos` y `GET /api/proyectos/{id}`.
- `Proyecto` no tiene `descripcion`; `PUT` solo edita nombre y fechas. `TieneHistorial` es stub que devuelve `false`.

### Cambios de contrato

| Operación | Entrada |
| :--- | :--- |
| `POST /api/proyectos` | `{ nombre, descripcion?, fecha_inicio, fecha_fin? }` |
| `PUT /api/proyectos/{id}` | `{ nombre, descripcion?, fecha_inicio, fecha_fin? }` (proyecto `Activo`) |
| `PATCH /api/proyectos/{id}/estado` | `{ estado: "Activo" | "Cerrado" }` (cerrar exige `fecha_fin`) |
| `DELETE /api/proyectos/{id}` | Sin body |
| `GET /api/proyectos/{id}` | — |

## Salidas esperadas

| HTTP | `error` | Cuándo ocurre |
| :---: | :--- | :--- |
| `200` | — | `GET /api/proyectos/{id}` con proyecto completo (incluye `descripcion`) |
| `200` | — | `GET /api/proyectos` incluye `descripcion` en cada proyecto del listado |
| `200` | — | `PUT`/`PATCH` exitoso con proyecto completo |
| `201` | — | `POST` exitoso con proyecto completo |
| `204` | — | `DELETE` exitoso (baja lógica) |
| `400` | `PARAMETRO_INVALIDO` | `id` inválido o fechas con formato inválido |
| `400` | `JSON_INVALIDO` | Body no parseable |
| `404` | `PROYECTO_NO_ENCONTRADO` | Proyecto inexistente, ajeno o dado de baja (US-F10/RB-01) |
| `409` | `PROYECTO_CERRADO` | `PUT` sobre proyecto cerrado |
| `409` | `PROYECTO_CON_HISTORIAL` | `DELETE` con historial (hoy no emitido: stub `false`) |
| `422` | `VALIDACION` | Nombre, fechas o descripción inválidos; cierre sin fecha fin |

- `403 ACCESO_DENEGADO`: reservado como contrato defensivo; no se emite (proyecto ajeno → 404).

## Reglas de negocio

- **RB-01** Solo un miembro opera el proyecto. US-F10: proyecto ajeno responde `404 PROYECTO_NO_ENCONTRADO`.
- **RB-02** La descripción es opcional, se recorta (`TrimSpace`) y admite hasta 2000 caracteres; vacía se guarda como `""`.
- **RB-03** Editar datos usa `PUT` y aplica validaciones de US-01 más descripción. Un proyecto `Cerrado` no permite `PUT`.
- **RB-04** Cambiar estado usa `PATCH`. Cerrar exige `fecha_fin`. Transición al mismo estado es idempotente.
- **RB-05** Si cambian datos y estado: `PUT` y luego `PATCH`. Si `PUT` responde 200 y `PATCH` falla, se conserva el resultado del `PUT` y se permite reintentar el estado.
- **RB-06** `DELETE` realiza baja lógica (`borrado_en`); respuesta `204`. Si no quedan proyectos, el listado muestra estado vacío.

## Restricciones

- No crear endpoints nuevos.
- No modificar `tablas.sql` (legado) ni crear migraciones SQL manuales; la columna `descripcion` se agrega por `AutoMigrate` en GORM (`type:text`, nullable).
- No implementar `TieneHistorial` (stub). No cambiar estados (`Activo`/`Cerrado`). No hay roles/permisos por rol.
- UI accesible: modales con `role="dialog"`, `aria-modal="true"`, `aria-labelledby`, foco inicial, cierre con `Escape`, retorno de foco y botones con nombre accesible.
- Las acciones por fila no deben disparar la navegación de la fila.

## Casos límite

- Descripción de 2000 caracteres: válida. De 2001: `422 VALIDACION` (campo `descripcion`).
- Descripción solo espacios: se guarda como `""`.
- Proyecto cerrado: UI muestra "Reabrí el proyecto para editar sus datos" y acción **Reabrir proyecto** (solo `PATCH` a `Activo`, sin reapertura automática).
- `PUT` 200 + `PATCH` error: formulario abierto, datos guardados, mensaje y reintento de estado.
- `404 PROYECTO_NO_ENCONTRADO` en detalle: "No tenés acceso a este proyecto o no existe".
- Eliminación desde listado: la fila se quita al recibir `204`; si era la última, se muestra el estado vacío.
- Error de red: conservar datos del formulario y permitir reintentar.

## Condiciones de error

| Situación | Respuesta |
| :--- | :--- |
| `id` no numérico o 0 | `400 PARAMETRO_INVALIDO` |
| JSON inválido | `400 JSON_INVALIDO` |
| Fecha con formato inválido | `400 PARAMETRO_INVALIDO` (campo de fecha) |
| Proyecto inexistente/ajeno/baja | `404 PROYECTO_NO_ENCONTRADO` |
| `PUT` en proyecto cerrado | `409 PROYECTO_CERRADO` |
| Cerrar sin fecha de fin | `422 VALIDACION` |
| Nombre/fechas/descripción inválidos | `422 VALIDACION` (campo específico) |
| Historial asociado al eliminar | `409 PROYECTO_CON_HISTORIAL` |
| Fallo de persistencia | `500 ERROR_INTERNO` |

## Criterios de aceptación

- [ ] **CA-1** `POST/GET detalle/GET listado/PUT` incluyen `descripcion` (recortada, vacía como `""`, máx. 2000).
- [ ] **CA-2** Desde el listado se puede eliminar una fila con confirmación accesible; `204` quita la fila (estado vacío si no quedan), `409` conserva la fila y muestra mensaje; las acciones no navegan la fila.
- [ ] **CA-3** `/proyectos/:id` muestra **Información general** (no "Resumen") con datos completos y formulario precargado.
- [ ] **CA-4** Edición: solo datos → `PUT`; solo estado → `PATCH`; ambos → `PUT` luego `PATCH`. Proyecto cerrado exige reabrir explícitamente antes de editar datos.
- [ ] **CA-5** Fallo parcial (`PUT` ok, `PATCH` falla) conserva datos guardados y permite reintentar estado.
- [ ] **CA-6** Validaciones y errores en UI: campos obligatorios/límites, `404`, `422`, `409`, y error de red conservando el formulario.
- [ ] **CA-7** Eliminar desde información general con confirmación; al `204` navega a `/proyectos`.

## Archivos afectados (previstos)

| Capa | Archivos |
| :--- | :--- |
| Domain | `internal/domain/proyecto.go`, `internal/domain/errores.go` |
| Repository/model | `internal/repository/models/proyecto_model.go`, `internal/repository/models/mappers.go`, `internal/repository/proyecto_repository.go` |
| Service | `internal/service/proyecto_service.go` |
| Handler/DTO | `internal/handler/dto_proyecto.go`, `internal/handler/proyecto_handler.go`, `internal/handler/errores.go` |
| Frontend API | `frontend/src/api/types.ts`, `frontend/src/api/proyectos.ts` |
| Frontend UI | `frontend/src/pages/ProyectosPage.tsx`, `frontend/src/pages/proyecto/InformacionGeneralPage.tsx`, `frontend/src/pages/validacionProyecto.ts`, `frontend/src/layouts/BarraLateralProyecto.tsx`, `frontend/src/App.tsx` (eliminar `ResumenPage.tsx`) |
| Tests | `*_test.go` afectados, `frontend/src/**/*.test.ts(x)` afectados/nuevos |

## Escenarios BDD

Ver `docs/features/US-F11-editar-eliminar-proyectos.feature`.
