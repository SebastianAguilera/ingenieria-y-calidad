# Especificación: US-F23 - Vista del reporte de proyecto y sprint

## Objetivo

Agregar al proyecto la pantalla **Reporte**, que muestra el reporte consolidado
del proyecto completo o de un sprint y permite descargarlo en **JSON** y en
**PDF**. La pantalla **no calcula métricas**: muestra lo que devuelven los
endpoints de US-22 y descarga lo que devuelven los de US-24. No modifica el
backend.

**Trazabilidad:** Issue #112 → esta especificación → criterios de aceptación
(CA-1 a CA-12) → escenarios BDD (`docs/features/US-F23-reporte.feature`) →
tests Vitest/RTL → código en `frontend/src`.

| Eslabón | Artefacto |
| :--- | :--- |
| Issue | [SebastianAguilera/ingenieria-y-calidad#112](https://github.com/SebastianAguilera/ingenieria-y-calidad/issues/112) |
| SDD | `docs/specs/US-F23-reporte-spec.md` (este archivo) |
| BDD | `docs/features/US-F23-reporte.feature` (E-01 a E-16) |
| Tests | `api/client.descarga.test.ts`, `api/reportes.test.ts`, `pages/proyecto/reporte/nombreArchivo.test.ts`, `pages/proyecto/ReportePage.test.tsx`, `layouts/AppShell.test.tsx` |
| Código | ver [Archivos afectados](#archivos-afectados) |

## Dependencias

| Historia | Relación |
| :--- | :--- |
| US-22 (#109) | **Bloqueante, mergeada.** `GET /api/proyectos/{id}/reporte` y `GET /api/proyectos/{id}/sprints/{sprintId}/reporte`. |
| US-24 (#111) | **Mergeada.** `GET .../reporte.pdf` para proyecto y sprint. Ver [D-01](#decisiones). |
| US-F21 (#103) | Componentes de KPI (`TarjetaKpi`), formato de números y regla "Sin datos". |
| US-F11 | Navegación del proyecto (`ProyectoLayout`, barra lateral). |
| US-F20 | Mensaje compartido de `404`: "No tenés acceso a este proyecto o no existe". |

## Decisiones

| Id | Decisión | Motivo |
| :--- | :--- | :--- |
| D-01 | **Se incluye el botón "Descargar PDF"**, aunque el issue lo dejaba fuera de alcance. | El issue lo excluía "hasta que se decida US-24"; US-24 ya está mergeada en `develop`. Cambio de alcance acordado con el equipo. |
| D-02 | El nombre de los archivos descargados lo arma el frontend. | En producción el frontend llama a la API en otro origen (Render) y el CORS no expone `Content-Disposition`, así que el nombre que manda el backend no es legible desde el navegador. |
| D-03 | Formato del nombre: `reporte-{nombre-proyecto}-proyecto-{fecha}.{ext}` y `reporte-proyecto-{id}-sprint-{sprint}-{fecha}.{ext}`. | Mismo esquema que `pdf_service.go` (US-24); el reporte de sprint no trae el nombre del proyecto, por eso usa su id. Nombre e identificador en slug ASCII (minúsculas, sin tildes, guiones); `{fecha}` es el día local `AAAA-MM-DD`. |
| D-04 | El JSON descargado es la respuesta de la API sin transformar, con sangría de 2 espacios. | CA del issue: "el mismo contenido que devuelve la API". |
| D-05 | Las tarjetas de métricas se reutilizan de US-F21 extrayendo `ResumenKpi` y `DatosSprint` a archivos propios. | Así el reporte no importa `MetricasPage` ni Recharts, y el bundle inicial no crece. |

## Entradas

- Ruta `/proyectos/:id/reporte`; `:id` entero positivo.
- Selector de alcance: `Proyecto completo` (valor por defecto) o un sprint del
  proyecto (lista de `GET /api/proyectos/{id}/sprints`).
- Botones `Descargar JSON` y `Descargar PDF`.

## Salidas

| Alcance | Endpoint del reporte | Endpoint del PDF |
| :--- | :--- | :--- |
| Proyecto | `GET /api/proyectos/{id}/reporte` | `GET /api/proyectos/{id}/reporte.pdf` |
| Sprint | `GET /api/proyectos/{id}/sprints/{sprintId}/reporte` | `GET /api/proyectos/{id}/sprints/{sprintId}/reporte.pdf` |

Secciones de la pantalla:

1. **Encabezado.** Proyecto: nombre, descripción, fechas (fin `null` → "Sin
   fecha de fin") y estado. Sprint: identificador, sprint goal, fechas y estado.
2. **Métricas.** Proyecto: las seis tarjetas de US-F21 (velocidad, historias
   completadas, desviación, ratio de defectos, defectos detectados y
   resueltos). Sprint: los ocho datos del panel de sprint de US-F21.
3. **Historias.** Código, título, estado, SP, horas estimadas y horas reales.
4. **Defectos.** Código, descripción, severidad, estado, sprint de detección y
   sprint de resolución (nombre del sprint; `null` → `—`).

## Reglas de negocio

| Id | Regla |
| :--- | :--- |
| RN-01 | La pantalla no recalcula métricas: muestra las de la respuesta. |
| RN-02 | Una métrica `null` se muestra como "Sin datos", nunca como `0`. En las tablas, un valor `null` se muestra como `—`. |
| RN-03 | Al cambiar el alcance se pide el reporte correspondiente; mientras carga, el selector y el título siguen visibles. |
| RN-04 | Lista vacía de historias o defectos: la tabla muestra un mensaje de vacío y no falla. |
| RN-05 | Los botones de descarga solo están habilitados con un reporte cargado. Mientras se genera el PDF, su botón se deshabilita y dice "Generando PDF…". |
| RN-06 | `404 PROYECTO_NO_ENCONTRADO` → "No tenés acceso a este proyecto o no existe". Error de red o de servidor → mensaje con botón "Reintentar". |
| RN-07 | Si falla la descarga del PDF, se muestra el error y el reporte sigue visible. |

## Criterios de aceptación

- **CA-1:** Given un miembro, When entra a Reporte, Then ve el reporte del proyecto completo.
- **CA-2:** Given el selector en un sprint, When se elige, Then se carga `GET .../sprints/{sprintId}/reporte` y se muestran solo sus historias y defectos.
- **CA-3:** Given un proyecto sin historias, When se consulta, Then las tablas muestran su estado vacío y no fallan.
- **CA-4:** Given una métrica `null`, When se muestra, Then aparece "Sin datos".
- **CA-5:** Given el botón "Descargar JSON", When se presiona, Then se descarga el reporte con el mismo contenido que devuelve la API.
- **CA-6:** Given un error de red, When se carga el reporte, Then se muestra el mensaje y se puede reintentar.
- **CA-7:** Given un no miembro, When entra al proyecto, Then ve "No tenés acceso a este proyecto o no existe".
- **CA-8 (D-01):** Given el botón "Descargar PDF", When se presiona, Then se descarga el PDF del alcance elegido.
- **CA-9:** Given un PDF que falla, When se presiona el botón, Then se ve el error y el reporte sigue en pantalla.
- **CA-10:** Given los defectos de un reporte, When se muestran, Then los sprints de detección y resolución aparecen por nombre.
- **CA-11:** Given cualquier descarga, When se guarda, Then el archivo se llama según D-03.
- **CA-12:** Given la barra lateral del proyecto, When se mira, Then existe la entrada "Reporte".

## Casos límite

- Proyecto sin sprints: el selector solo ofrece "Proyecto completo".
- Proyecto sin fecha de fin: "Sin fecha de fin".
- Defecto sin sprint de resolución: `—`.
- Nombre de proyecto con tildes o símbolos: el slug los quita (`Gestión & Co` → `gestion-co`).
- Nombre que queda vacío como slug: se usa `reporte`.

## Archivos afectados

| Archivo | Cambio |
| :--- | :--- |
| `api/types.ts` | Tipos `ReporteProyecto`, `ReporteSprint`, `ReporteHistoria`, `ReporteDefecto`. |
| `api/client.ts` | `descargarArchivo`: GET con token que devuelve un `Blob` y traduce errores igual que `request`. |
| `api/reportes.ts` | `obtenerReporteProyecto`, `obtenerReporteSprint`, `descargarPdfReporte`. |
| `pages/proyecto/ReportePage.tsx` | Página nueva. |
| `pages/proyecto/reporte/*` | Hook `useReporte`, nombre de archivo, guardado, encabezado, tablas y botones. |
| `pages/proyecto/metricas/ResumenKpi.tsx`, `DatosSprint.tsx` | Extraídos de `MetricasPage` y `PanelSprint` (D-05), sin cambiar su comportamiento. |
| `App.tsx`, `layouts/BarraLateralProyecto.tsx`, `layouts/MigasDePan.tsx` | Ruta, entrada "Reporte" en la barra lateral y en las migas de pan. |

## Fuera de alcance

Edición de datos desde el reporte, reportes personalizados, cambios de backend
(incluido exponer `Content-Disposition` por CORS).
