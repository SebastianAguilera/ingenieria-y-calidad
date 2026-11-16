# Especificación: US-F21 - Dashboard de métricas y carga de horas estimadas

## Objetivo

Reemplazar el placeholder de **Métricas** por un dashboard que muestre el avance
del proyecto y de cada sprint con tarjetas de KPI y gráficos, y agregar al
backlog la carga de **horas estimadas** por historia, para que la desviación de
esfuerzo tenga datos. La pantalla **no calcula métricas**: solo muestra lo que
devuelven los endpoints de US-20. La única excepción es la distribución de
historias por estado (dona), que se cuenta en el frontend a partir del backlog.

La historia satisface el issue `SebastianAguilera/ingenieria-y-calidad#103`
(`US-F21 - Dashboard de métricas y carga de horas estimadas`). No modifica el
backend.

**Trazabilidad:** Issue #103 → esta especificación → criterios de aceptación
(CA-1 a CA-15) → escenarios BDD
(`docs/features/US-F21-dashboard-metricas.feature`) → tests Vitest/RTL → código
en `frontend/src`.

| Eslabón | Artefacto |
| :--- | :--- |
| Issue | [SebastianAguilera/ingenieria-y-calidad#103](https://github.com/SebastianAguilera/ingenieria-y-calidad/issues/103) |
| SDD | `docs/specs/US-F21-dashboard-metricas-spec.md` (este archivo) |
| BDD | `docs/features/US-F21-dashboard-metricas.feature` (E-01 a E-24) |
| Tests | `api/metricas.test.ts`, `pages/proyecto/metricas/formato.test.ts`, `pages/proyecto/metricas/distribucion.test.ts`, `pages/proyecto/MetricasPage.test.tsx`, `pages/proyecto/backlog/ModalHistoria.test.tsx`, `pages/proyecto/backlog/TablaBacklog.test.tsx`, `pages/proyecto/BacklogPage.test.tsx` |
| Código | ver [Archivos afectados](#archivos-afectados) |

Se implementa en dos tramos:

- **Tramo 1** (requiere US-20, mergeada): dashboard de métricas.
- **Tramo 2** (requiere US-19, mergeada): campo y columna de horas estimadas.

Ambas dependencias están en `develop`, así que los dos tramos se entregan en la
misma rama (`feature/US-F21-dashboard-metricas`).

## Entradas

### Precondiciones

- US-20: `GET /api/proyectos/{id}/metricas` y
  `GET /api/proyectos/{id}/sprints/{sprintId}/metricas`.
- US-19: `horas_estimadas` en `POST`/`PATCH /api/proyectos/{id}/historias` y en
  el listado de historias.
- `listarHistorias` (US-02) y `listarSprints` (US-03) existentes en el cliente.
- Guard de membresía (US-F10): no miembro → `404 PROYECTO_NO_ENCONTRADO`.
- Ruta `/proyectos/:id/metricas` y entrada "Métricas" de la barra lateral del
  proyecto ya registradas (US-F11/US-F03); esta historia reemplaza el
  contenido del placeholder.

### Contrato consumido (US-20)

`GET /api/proyectos/{id}/metricas` → `MetricasProyecto`. Campos usados:

| Campo | Tipo | Nulo | Uso en la UI |
| :--- | :--- | :---: | :--- |
| `velocidad` | número | sí | KPI "Velocidad" (SP por sprint cerrado) |
| `porcentaje_historias_completadas` | número `0..1` | sí | KPI "Historias completadas" |
| `desviacion_esfuerzo` | número | sí | KPI "Desviación de esfuerzo" |
| `horas_estimadas`, `horas_reales` | número | no | detalle del KPI de desviación |
| `ratio_defectos` | número `≥ 0` | sí | KPI "Ratio de defectos" |
| `defectos_detectados`, `defectos_resueltos` | entero | no | KPI "Defectos" |

Los campos de US-21 (`sprints_restantes`, `wip`, `tasa_entrega`,
`lead_time_*`, `defectos_abiertos`, `defectos_por_severidad`) se tipan en
`api/types.ts` pero **no se muestran**: van en US-F22.

`GET /api/proyectos/{id}/sprints/{sprintId}/metricas` → `MetricasSprint`:
`story_points_planificados`, `story_points_completados`,
`porcentaje_completado` (nulo), `historias_completadas`, `horas_estimadas`,
`horas_reales`, `desviacion_esfuerzo` (nulo), `defectos_detectados`,
`defectos_resueltos`.

### Datos de entrada del tramo 2

| Campo | Tipo | Obligatorio | Regla |
| :--- | :--- | :---: | :--- |
| `horas_estimadas` | número (horas) | no | vacío → sin valor; si se completa, número finito `> 0` |

## Salidas esperadas

### Tramo 1: `MetricasPage`

1. **Encabezado** "Métricas" y **aviso** visible: "Calculado sobre el estado
   actual de las historias" con la aclaración de que el sistema no guarda
   histórico.
2. **Resumen (tarjetas KPI)**, en este orden:
   - Velocidad: `"{n} SP"` por sprint cerrado.
   - Historias completadas: porcentaje.
   - Desviación de esfuerzo: porcentaje con signo y detalle
     `"{reales} h reales / {estimadas} h estimadas"`.
   - Ratio de defectos: número con hasta 2 decimales y "por historia completada".
   - Defectos detectados.
   - Defectos resueltos.
   La tarjeta de ratio de defectos usa la variante de alerta (acento) cuando el
   valor es mayor que 0.
3. **Gráfico de barras** "Story points por sprint": por cada sprint `Cerrado` y
   el `Activo`, dos barras (SP planificados y SP completados), con leyenda y
   valores. Si no hay sprints cerrados ni activo, muestra un estado vacío.
4. **Gráfico de dona** "Historias por estado": tres segmentos (`Backlog`,
   `En Progreso`, `Hecho`) con cantidad y etiqueta en la leyenda.
5. **Panel de sprint**: selector con todos los sprints del proyecto y sus
   métricas: SP planificados, SP completados, porcentaje completado, historias
   completadas, horas estimadas, horas reales y desviación de esfuerzo.
   Selección inicial: el sprint `Activo`; si no hay, el último `Cerrado`; si no,
   el primero de la lista. Sin sprints, el panel muestra "Todavía no hay
   sprints".
6. **Estados**: carga (`role="status"`), vacío (proyecto sin historias), error
   con botón "Reintentar" y `404` con el mensaje compartido.

### Tramo 2

- `ModalHistoria` con el campo "Horas estimadas" (opcional), precargado al
  editar.
- `TablaBacklog` con la columna "Horas estimadas" (`—` cuando es `null`).
- `Historia` en `api/types.ts` con `horas_estimadas: number | null`;
  `AltaHistoria` y `ActualizacionHistoria` con `horas_estimadas?: number | null`.

## Reglas de negocio

| ID | Regla |
| :--- | :--- |
| RN-01 | La pantalla no calcula métricas; solo formatea lo que devuelve la API. Excepción: la distribución de la dona se cuenta del backlog cargado. |
| RN-02 | Un valor `null` se muestra como **"Sin datos"** con una explicación breve, nunca como `0`. Un `0` real se muestra como `0`. |
| RN-03 | Los porcentajes llegan en `0..1`; la UI los multiplica por 100, usa hasta 1 decimal con coma (`es-AR`) y agrega `%`. `0.75` → `75%`; `0.6153…` → `61,5%`. |
| RN-04 | La desviación de esfuerzo se muestra como porcentaje con signo: `0.2` → `+20%`, `-0.1` → `−10%`, `0` → `0%`. Positiva = se trabajó más de lo estimado. |
| RN-05 | Los números (SP, horas, ratio) se muestran con hasta 2 decimales, sin ceros de relleno. |
| RN-06 | Explicaciones de `null`: velocidad → "Todavía no hay sprints cerrados"; historias completadas → "El proyecto no tiene historias"; desviación → "Ninguna historia tiene horas estimadas y reales"; ratio → "Todavía no hay historias completadas"; porcentaje del sprint → "El sprint no tiene story points planificados". |
| RN-07 | El gráfico de barras incluye los sprints `Cerrado` y el `Activo`, en el orden de `listarSprints`. Los `Planificado` no entran. |
| RN-08 | El aviso "Calculado sobre el estado actual de las historias" se muestra siempre que haya datos. |
| RN-09 | `404 PROYECTO_NO_ENCONTRADO` → "No tenés acceso a este proyecto o no existe" (constante compartida `MENSAJE_PROYECTO_INEXISTENTE`, US-F20). |
| RN-10 | `horas_estimadas` es opcional. Vacío → en el alta se envía `null`; en la edición no se envía el campo. Completo → número finito `> 0`, si no, error en el campo y no se llama a la API. |
| RN-11 | Un `422` con `campo: "horas_estimadas"` se muestra en el campo, no como error general. |

### Sistema visual (skill `teal-dashboard-ui`)

- Solo tokens de `theme.css`; nada de hex ni `rgb()` en componentes ni en la
  configuración de los gráficos.
- Se agregan los tokens de gráficos de la skill: `--chart-1` (primario),
  `--chart-2` (acento), `--chart-3` (primario suave), `--chart-grid` y
  `--accent-300`, con sus valores de modo oscuro.
- Tarjetas KPI con caja de icono de 40 px (`--primary-100` + icono primario),
  valor grande en `--font-secondary` y etiqueta chica; variante de alerta en
  `--accent-100`.
- Barras: SP planificados en `--chart-3`, SP completados en `--chart-1`. Dona:
  `Hecho` en `--chart-1`, `En Progreso` en `--chart-2`, `Backlog` en `--chart-3`.
- Rejilla solo horizontal y punteada en `--chart-grid`; ejes sin línea y texto
  de 11 px en `--ink-soft`.
- Íconos de `bootstrap-icons`, que ya usa el proyecto (la skill sugiere Lucide,
  pero no se agrega una segunda librería de íconos).

### Librería de gráficos

No había ninguna. Se elige **Recharts 3** (`recharts@^3.10.1`), la que indica la
skill: componentes React declarativos, SVG y colores por prop (acepta
`var(--chart-n)`).

Para no depender de `ResponsiveContainer` (que necesita `ResizeObserver` y
medidas reales, ausentes en jsdom), cada gráfico mide su contenedor con un hook
propio (`useAnchoContenedor`) y solo dibuja cuando el ancho es mayor a 0. En los
tests los valores se verifican en la tabla accesible, que siempre se renderiza.

`MetricasPage` se carga con `React.lazy` para que Recharts (≈395 kB) no entre en
el bundle inicial.

## Restricciones

- Sin cambios de backend.
- Sin exportación a PDF, comparación entre proyectos ni actualización en tiempo
  real.
- Las extensiones de US-21 no se muestran (US-F22).

## Accesibilidad

- Cada gráfico tiene una tabla con `caption` y los mismos valores. Está oculta
  visualmente (`sr-only`) y se muestra con el botón "Ver como tabla"
  (`aria-expanded`). El SVG del gráfico tiene `aria-hidden`.
- Los colores siempre van con etiqueta y valor (leyenda con texto y valores en
  la tabla).
- El selector de sprint es un `<select>` nativo con `<label>`, navegable por
  teclado.
- El campo "Horas estimadas" tiene etiqueta, `aria-invalid` y
  `aria-describedby` hacia el mensaje de error (`TextField`).

## Casos límite

| ID | Caso | Resultado |
| :--- | :--- | :--- |
| CL-01 | Todas las métricas nulas (proyecto con historias y sin sprints) | Tarjetas con "Sin datos" y explicación; gráfico de barras vacío. |
| CL-02 | `velocidad: 0` (cerrados sin SP) | Muestra `0 SP`, no "Sin datos". |
| CL-03 | `ratio_defectos > 1` | Se muestra tal cual (ej. `1,5`), sin acotar. |
| CL-04 | Proyecto sin historias | Estado vacío "Todavía no hay historias" con enlace al backlog; no se dibujan KPI ni gráficos. |
| CL-05 | Proyecto sin sprints | Panel de sprint con "Todavía no hay sprints". |
| CL-06 | Sprint `Planificado` elegido | Se piden sus métricas igual; con `porcentaje_completado: null` muestra "Sin datos". |
| CL-07 | `horas_estimadas: 0.25` | Válido. |
| CL-08 | `horas_estimadas: 0` o negativo | Error "Las horas estimadas deben ser mayores a 0" y no se llama a la API. |
| CL-09 | Campo de horas vaciado al editar una historia con horas | No se envía el campo; el valor guardado se mantiene (US-19 RN-05: el `PATCH` no borra). La ayuda del campo lo aclara. |

## Condiciones de error

| Situación | Resultado |
| :--- | :--- |
| `404 PROYECTO_NO_ENCONTRADO` al cargar | "No tenés acceso a este proyecto o no existe", sin botón de reintento útil (se muestra igual). |
| Error de red o 5xx al cargar | "No se pudieron cargar las métricas. Intentá nuevamente." + "Reintentar". |
| Error al cargar las métricas de un sprint del gráfico de barras | Se trata como error de carga de la pantalla (mensaje general + "Reintentar"). |
| Error al cargar las métricas de un sprint del panel | Mensaje en el panel "No se pudieron cargar las métricas del sprint." + "Reintentar". |
| `422 VALIDACION` en `horas_estimadas` | Mensaje del backend en el campo. |

## Criterios de aceptación

Tramo 1:

- **CA-1** Given un miembro, When entra a Métricas, Then ve las tarjetas de KPI
  con los valores de `GET /metricas`.
- **CA-2** Given una métrica con valor `null`, When se muestra, Then aparece
  "Sin datos" con una explicación y no `0`.
- **CA-3** Given un porcentaje de `0.75` de la API, When se muestra, Then la UI
  dice `75%`.
- **CA-4** Given sprints cerrados, When se muestra el gráfico de barras, Then
  cada sprint tiene dos barras: SP planificados y SP completados.
- **CA-5** Given historias en distintos estados, When se muestra la dona, Then
  cada segmento tiene su cantidad y su etiqueta.
- **CA-6** Given un sprint elegido en el selector, When se carga, Then se
  muestran sus métricas de `GET .../sprints/{sprintId}/metricas`.
- **CA-7** Given un proyecto sin historias, When se entra, Then se muestra el
  estado vacío con un mensaje claro.
- **CA-8** Given cualquier gráfico, When se consulta con lector de pantalla,
  Then existe un resumen en texto con los mismos valores.
- **CA-9** Given el aviso de estado actual, When se entra al dashboard, Then se
  muestra visible.
- **CA-10** Given un no miembro, When entra al proyecto, Then ve "No tenés
  acceso a este proyecto o no existe".

Tramo 2:

- **CA-11** Given el formulario de historia, When se completa "Horas estimadas"
  con un valor mayor a 0, Then se guarda y la tabla lo muestra.
- **CA-12** Given el campo vacío, When se guarda, Then se envía sin valor y la
  tabla muestra `—`.
- **CA-13** Given `0` o un valor negativo, When se envía, Then el error aparece
  en el campo y no se llama a la API.
- **CA-14** Given un 422 del backend en `horas_estimadas`, When se guarda, Then
  el mensaje aparece en el campo.
- **CA-15** Given una historia con horas estimadas, When se abre para editar,
  Then el campo viene precargado.

## Archivos afectados

| Archivo | Cambio |
| :--- | :--- |
| `frontend/package.json` | Dependencia `recharts`. |
| `frontend/src/theme/theme.css` | Tokens de gráficos (`--chart-*`, `--accent-300`). |
| `frontend/src/api/types.ts` | `MetricasProyecto`, `MetricasSprint`, `horas_estimadas` en `Historia`, `AltaHistoria` y `ActualizacionHistoria`. |
| `frontend/src/api/metricas.ts` | `obtenerMetricasProyecto`, `obtenerMetricasSprint`. |
| `frontend/src/pages/proyecto/MetricasPage.tsx` | Dashboard. |
| `frontend/src/pages/proyecto/metricas/formato.ts` | Formato de porcentajes, desviación, números y horas. |
| `frontend/src/pages/proyecto/metricas/distribucion.ts` | Conteo de historias por estado y sprints del gráfico. |
| `frontend/src/pages/proyecto/metricas/useMetricas.ts` | Carga de métricas, historias, sprints y métricas por sprint. |
| `frontend/src/pages/proyecto/metricas/useMetricasSprint.ts` | Métricas del sprint elegido. |
| `frontend/src/pages/proyecto/metricas/TarjetaKpi.tsx` | Tarjeta de KPI. |
| `frontend/src/pages/proyecto/metricas/GraficoSprints.tsx` | Barras SP planificados/completados. |
| `frontend/src/pages/proyecto/metricas/GraficoEstados.tsx` | Dona por estado. |
| `frontend/src/pages/proyecto/metricas/TarjetaGrafico.tsx` | Tarjeta de gráfico con la tabla alternativa y "Ver como tabla". |
| `frontend/src/pages/proyecto/metricas/Leyenda.tsx` | Leyenda con cuadrado de color y texto. |
| `frontend/src/pages/proyecto/metricas/estilosGrafico.ts` | Colores y estilos de los gráficos (solo tokens). |
| `frontend/src/pages/proyecto/metricas/PanelSprint.tsx` | Selector y métricas del sprint. |
| `frontend/src/pages/proyecto/metricas/useAnchoContenedor.ts` | Ancho del contenedor para los gráficos. |
| `frontend/src/App.tsx` | `MetricasPage` con `React.lazy`: Recharts queda en un chunk aparte. |
| `frontend/src/ui/TextField.tsx` | Prop `ayuda` (texto de ayuda enlazado con `aria-describedby`). |
| `frontend/src/pages/proyecto/backlog/ModalHistoria.tsx` | Campo "Horas estimadas". |
| `frontend/src/pages/proyecto/backlog/tipos.ts` | `horas_estimadas` en `DatosHistoria`. |
| `frontend/src/pages/proyecto/backlog/TablaBacklog.tsx` | Columna "Horas estimadas". |
| `frontend/src/pages/proyecto/backlog/useHistorias.ts` | Aplica `horas_estimadas` en la actualización optimista. |

## Dependencias

| Historia | Relación |
| :--- | :--- |
| US-20 | Endpoints de métricas (bloqueante del tramo 1). |
| US-19 | `horas_estimadas` (bloqueante del tramo 2). |
| US-F11 / US-F03 | Ruta y navegación del proyecto. |
| US-F20 | Constante compartida del mensaje `404`. |
| US-F22, US-F23 | Bloqueadas por esta historia. |

## Fuera de alcance

- Exportar a PDF, comparación entre proyectos y actualización en tiempo real.
- Mostrar las extensiones de US-21 (US-F22).
- Cambios de backend.
- Borrar `horas_estimadas` de una historia (el `PATCH` de US-19 no lo permite).

## Supuestos

- "Horas de desviación" del panel de sprint se muestra como horas reales y
  estimadas del sprint más la desviación porcentual que devuelve la API; la UI
  no resta horas por su cuenta (RN-01).
- Las métricas de los sprints del gráfico se piden una por una al endpoint de
  sprint (no hay endpoint agregado).
