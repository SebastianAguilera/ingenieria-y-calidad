# Especificación: US-F26 - Mejora de pantalla de sprints y Sprint Backlog

## Objetivo
Mejorar la pantalla de listado de sprints y el detalle del Sprint Backlog para ofrecer una vista más completa y operativa: navegación clara al detalle, filtros por estado, cabecera con KPIs y una tabla de historias con las mismas columnas y acciones que el backlog. No modifica el backend; reutiliza contratos existentes (US-03, US-04, US-13, US-20, US-02).

**Trazabilidad:** Issue #132 → esta especificación → criterios de aceptación → escenarios BDD (`docs/features/US-F26-mejora-pantalla-sprints.feature`) → tests Vitest/RTL → código frontend.

## Entradas
- `GET /api/proyectos/{id}/sprints` (US-03/US-F13).
- `GET /api/proyectos/{id}/sprints/{sprintId}/historias` (US-13).
- `GET /api/proyectos/{id}/sprints/{sprintId}/metricas` (US-20).
- `GET /api/proyectos/{id}/historias` (US-02) para candidatas a asignar.
- `POST/DELETE .../historias` (US-04) para asignar/quitar.
- `PATCH /api/proyectos/{id}/historias/{historiaId}` (US-02) para estado/prioridad/edición.

## Salidas esperadas
### Listado de sprints (`/proyectos/:id/sprints`)
- Fila completa navegable (teclado/lector) al detalle; botón “Ver sprint” en todos los estados.
- Filtro por estado con chips `Todos`, `Planificado`, `Activo`, `Cerrado` y contador por estado.
- Badge de estado unificado y duración en días debajo del nombre/identificador.
- Sprint Cerrado: sin acciones Activar/Cerrar, con “Ver sprint”.
- Sin historias: estado vacío con CTA de creación (comportamiento existente).

### Detalle de sprint (`/proyectos/:id/sprints/:sprintId`)
- Cabecera: nombre/identificador, estado, fechas, duración y Sprint Goal; acciones `Asignar historia` (solo Activo), `Ver reporte`, `Ver métricas`.
- Fila de KPIs: SP planificados, SP completados, % completado, historias completadas, horas estimadas y horas reales. Fuente: métricas del sprint. Si falla la carga de métricas, tabla funciona y cada KPI muestra “—”.
- Tabla (columnas como backlog): Código, Título, Estado, Prioridad, Estimación, Horas estimadas, Acciones.
- Estado y Prioridad editables inline (PATCH historia) en sprints no Cerrados; en Cerrado chips de solo lectura.
- Acciones por historia: `Estimar` (Planning Poker), `Editar` (reutiliza `ModalHistoria`), `Quitar del sprint` (solo no Cerrado, con confirmación).
- Sprint Activo sin historias: estado vacío con CTA `Asignar historia`.

## Reglas de negocio
- **RB-01** La fila del listado y el botón “Ver sprint” navegan al detalle en todos los estados. → CA-1
- **RB-02** Los chips filtran localmente y muestran contador por estado; el contador/selección coincide con lo mostrado. → CA-2
- **RB-03** Un sprint Cerrado en el listado no muestra Activar/Cerrar, solo Ver sprint. → CA-3
- **RB-04** En detalle Activo se muestran cabecera completa y acción Asignar historia. → CA-4
- **RB-05** KPIs se muestran con datos de métricas; ante error de métricas la tabla sigue operativa y KPIs en “—”. → CA-5, CA-6
- **RB-06** Cambios inline de estado/prioridad se persisten con PATCH y se reflejan en tabla (sprint no Cerrado). → CA-7
- **RB-07** Sprint Cerrado: estado/prioridad solo lectura y sin Quitar del sprint. → CA-8
- **RB-08** Quitar del sprint requiere confirmación, usa DELETE existente y actualiza la tabla. → CA-9
- **RB-09** Sprint Activo vacío muestra CTA Asignar historia. → CA-10
- **RB-10** Sin cambios de backend; “Quitar del sprint” usa el endpoint existente de desasignación.

## Casos límite
- Duración mínima 1 día (misma fecha inicio/fin).
- Métricas con `porcentaje_completado` o `desviacion_esfuerzo` nulos: mostrar “—”.
- Historia sin story points u horas estimadas: mostrar “—”/“Sin estimar” según columna.
- Filtro sin resultados: aviso de ausencia (no estado vacío global).

## Condiciones de error
- Error al cargar métricas: KPIs “—”, tabla y acciones de historias disponibles.
- Errores de PATCH/quitar: mensaje visible, sin perder consistencia de tabla.

## Criterios de aceptación
- CA-1: pulsar fila o “Ver sprint” abre detalle.
- CA-2: chip “Activo” filtra y contador coincide.
- CA-3: Cerrado sin Activar/Cerrar y con “Ver sprint”.
- CA-4: detalle Activo muestra estado/fechas/duración/Sprint Goal y “Asignar historia”.
- CA-5: KPIs con SP planificados/completados, %, historias completadas y horas.
- CA-6: fallo métricas → tabla ok y KPIs “—”.
- CA-7: cambio inline estado/prioridad persiste y se refleja.
- CA-8: Cerrado → chips lectura y sin “Quitar del sprint”.
- CA-9: confirmar “Quitar del sprint” desasigna y actualiza tabla.
- CA-10: sprint sin historias y Activo muestra vacío con CTA “Asignar historia”.

## Arquitectura
Frontend (`frontend/src`): `pages/proyecto/SprintsPage.tsx`, `pages/proyecto/SprintDetallePage.tsx`, `pages/proyecto/sprints/TablaSprintBacklog.tsx`, `pages/proyecto/sprints/useSprintBacklog.ts` (y utilidades de duración/formato según convenga). Reutiliza `ModalHistoria`, `SelectPill`, `TablaHistorias`, `api/metricas`, `api/historias`.
