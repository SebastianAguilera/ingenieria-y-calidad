# US-22 — Reporte de proyecto y sprint

## Objetivo

Exponer, en solo lectura, un reporte consolidado de un proyecto o de un sprint, reuniendo información general, historias, métricas y defectos en una sola consulta.

El reporte se calcula al momento de la consulta a partir del estado actual de historias, Sprints, asignaciones al Sprint Backlog y defectos. Reutiliza los cálculos de US-20 (núcleo) y US-21 (extensiones). No se persisten resultados, no se crean tablas y no se modifica ningún dato.

**Trazabilidad:** Issue #109 → Esta especificación → Criterios de aceptación (CA-01 a CA-09) → Escenarios BDD (`docs/features/US-22-reporte.feature`) → Tests Go (TDD) → Código.

| Eslabón | Artefacto |
| :--- | :--- |
| Issue | [SebastianAguilera/ingenieria-y-calidad#109](https://github.com/SebastianAguilera/ingenieria-y-calidad/issues/109) |
| SDD | `docs/specs/US-22-reporte-spec.md` (este archivo) |
| Criterios de aceptación | CA-01 a CA-09 de este documento |
| BDD | `docs/features/US-22-reporte.feature` (E-01 a E-06) |
| Tests | `internal/service/reporte_service_test.go`, `internal/handler/reporte_handler_test.go` |
| Código | `internal/domain/reporte.go`, `internal/service/reporte_service.go`, `internal/handler/reporte_handler.go`, `internal/handler/dto_reporte.go`, `internal/handler/reporte_servicio.go`, `cmd/api/router.go` |

## Dependencias y relación con otras historias

| Historia | Relación |
| :--- | :--- |
| US-01 | Aporta el proyecto, sus campos y validación 404 `PROYECTO_NO_ENCONTRADO`. |
| US-02 | Aporta `Historia`, `Estado`, `StoryPoints`, `EsfuerzoReal`. |
| US-03 | Aporta `Sprint`, sus campos y validación 404 `SPRINT_NO_ENCONTRADA`. |
| US-04 | Aporta asignaciones `SprintBacklogItem` para filtrar historias por sprint. |
| US-09 / US-18 | Aporta `Defecto`, `Severidad`, `EstadoDefecto`, `SprintDeteccionID` y `SprintResolucionID`. |
| US-19 | Aporta `HorasEstimadas` por historia. |
| US-20 | Aporta el motor de métricas de proyecto y sprint (núcleo). |
| US-21 | Aporta las extensiones de métricas de proyecto (sprints restantes, wip, tasa de entrega, lead time, defectos abiertos y por severidad). |
| US-F10 | Guard de membresía: no miembro recibe `404 PROYECTO_NO_ENCONTRADO`. |

**Bloquea a:** US-24 (exportación a PDF) y US-F23 (vista del reporte).

## Alcance

**Incluye**
- `GET /api/proyectos/{id}/reporte`: reporte del proyecto completo.
- `GET /api/proyectos/{id}/sprints/{sprintId}/reporte`: reporte de un sprint.
- Encabezado con datos del proyecto o sprint, fechas, estado y `base: "estado_actual"`.
- Listado de historias (todas para proyecto, solo asignadas para sprint) con código, título, estado, story points, horas estimadas (`horas_estimadas`) y horas reales (`horas_reales`).
- Métricas consolidadas (núcleo US-20 + extensiones US-21).
- Listado de defectos (todos para proyecto, filtrados por detección o resolución para sprint) con código/id, descripción, severidad, estado, sprint de detección y sprint de resolución.
- Guard de membresía (US-F10) y validación de existencia de proyecto y sprint.
- Tests TDD en Go.

**No incluye**
- Exportación a PDF (US-24).
- Escritura de datos, reportes personalizados o histórico de cambios.
- Modificación de las fórmulas de métricas (se consumen directamente de US-20/US-21).

## Endpoints y parámetros

| Endpoint | Método | Cuerpo | Query |
| :--- | :---: | :---: | :---: |
| `/api/proyectos/{id}/reporte` | `GET` | No | No |
| `/api/proyectos/{id}/sprints/{sprintId}/reporte` | `GET` | No | No |

| Parámetro | Ubicación | Obligatorio | Tipo / formato | Límites |
| :--- | :--- | :---: | :--- | :--- |
| `id` | path | Sí | entero positivo (`uint`) | `> 0` |
| `sprintId` | path | Sí | entero positivo (`uint`) | `> 0` |

## Criterios de aceptación

- **CA-01**: Given un proyecto con historias y defectos, When se consulta el reporte, Then las métricas coinciden con las de `GET /metricas` (US-20 y US-21) para el mismo proyecto.
- **CA-02**: Given un sprint, When se consulta su reporte, Then incluye solo las historias asignadas a ese sprint y los defectos detectados o resueltos en él.
- **CA-03**: Given una métrica sin datos, When aparece en el reporte, Then se muestra como `null`, igual que en US-20/US-21.
- **CA-04**: Given un proyecto sin historias, When se consulta, Then responde `200` con listas vacías y no `null`.
- **CA-05**: Given una historia sin horas estimadas, When aparece en el reporte, Then `horas_estimadas` es `null`.
- **CA-06**: Given la respuesta, When se lee, Then incluye `base: "estado_actual"`.
- **CA-07**: Given un no miembro, When consulta, Then `404 PROYECTO_NO_ENCONTRADO`.
- **CA-08**: Given un sprint de otro proyecto o inexistente, When se consulta, Then `404 SPRINT_NO_ENCONTRADA`.
- **CA-09**: Given cualquier consulta, When se completa, Then no se modifica ningún dato.
