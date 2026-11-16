# Especificación: US-F22 - Extensiones de métricas en el dashboard

## Objetivo

Agregar a la pantalla de **Métricas** (US-F21) los indicadores de flujo
(sprints restantes, WiP, tasa de entrega y Lead Time estimado) y el desglose de
defectos por severidad, para que un miembro estime cuánto falta y dónde está el
trabajo sin leer la API. La pantalla consume los campos opcionales que
`GET /api/proyectos/{id}/metricas` devuelve desde US-21. Es solo frontend: no
modifica el backend.

La historia satisface el issue `SebastianAguilera/ingenieria-y-calidad#104`
(`US-F22 - Extensiones de métricas en el dashboard`). Es una historia separada
porque las extensiones están fuera de `docs/proyecto.md` §7: si el tiempo
aprieta, US-F21 se entrega sin esta.

**Trazabilidad:** Issue #104 → esta especificación → criterios de aceptación
(CA-01 a CA-12) → escenarios BDD
(`docs/features/US-F22-extensiones-metricas.feature`) → tests Vitest/RTL →
código en `frontend/src`.

## Dependencias y relación con otras historias

| Historia | Relación |
| :--- | :--- |
| US-20 | Aporta `GET /api/proyectos/{id}/metricas` y los campos base. **Mergeada.** |
| US-21 | Aporta los campos nuevos que esta historia muestra (`sprints_restantes`, `wip`, `tasa_entrega`, `lead_time_estimado`, `lead_time_aproximado`, `defectos_abiertos`, `defectos_por_severidad`). **Mergeada.** Spec: `docs/specs/US-21-extensiones-metricas-spec.md`. |
| US-F21 (tramo 1) | **Bloqueante.** Aporta `MetricasPage`, la ruta, el cliente de API de métricas, el componente de tarjeta KPI, el tratamiento de `null` ("Sin datos"), los estados de pantalla y la **librería de gráficos**. |
| US-F20 | Aporta el helper `MENSAJE_PROYECTO_INEXISTENTE` y `etiquetasDefecto.ts` (textos de severidad). **Mergeada.** |

**No bloquea** a ninguna historia.

## Entradas

### Precondiciones

- US-21 implementada: el endpoint devuelve los campos nuevos.
- US-F21 tramo 1 mergeada en `develop`: existe `MetricasPage` con las tarjetas
  base y el cliente `obtenerMetricasProyecto` (nombre a confirmar).
- El usuario entra a la ruta existente `/proyectos/:id/metricas`.
- Esta historia **no agrega llamadas a la API**: usa la misma respuesta que
  F21 ya carga.

### Contrato consumido

Campos nuevos de `GET /api/proyectos/{id}/metricas` (US-21). Todos son
opcionales en el tipo TypeScript: una versión vieja del backend puede no
devolverlos.

| Campo | Tipo | Nulo | Descripción |
| :--- | :--- | :---: | :--- |
| `sprints_restantes` | número | sí | `ceil(SP pendientes / velocidad)`; `null` si no hay velocidad. |
| `wip` | entero | no | Historias en `En Progreso`. |
| `tasa_entrega` | número | sí | Historias `Hecho` por sprint cerrado (promedio); `null` sin sprints cerrados. |
| `lead_time_estimado` | número | sí | `wip / tasa_entrega`, en sprints; `null` si no hay tasa. |
| `lead_time_aproximado` | booleano | no | `true` cuando hay `lead_time_estimado` (Ley de Little). |
| `defectos_abiertos` | entero | no | Defectos en estado `open` (no cuenta `in_progress`). |
| `defectos_por_severidad` | objeto | no | `minor`, `major`, `critical`, `blocker`, con `0` donde no hay. |

## Salidas esperadas

- **Cuatro tarjetas KPI nuevas** en la pantalla de métricas, con el mismo
  componente de tarjeta de US-F21:
  - **Sprints restantes**: `N sprints` (`1 sprint` en singular).
  - **WiP**: cantidad de historias en curso.
  - **Tasa de entrega**: historias por sprint.
  - **Lead Time estimado**: `N sprints` con la etiqueta **"aprox."** y una nota
    que explica que es una estimación según la Ley de Little (WiP / tasa de
    entrega), no una medición.
- **Tarjeta Defectos abiertos**: valor de `defectos_abiertos`, con una nota de
  que cuenta los defectos en estado `Abierto`.
- **Gráfico de barras horizontales** "Defectos por severidad", con las cuatro
  barras siempre presentes en el orden `minor`, `major`, `critical`, `blocker`
  y etiquetas en español (Menor, Mayor, Crítica, Bloqueante). `blocker` se ve
  aunque valga `0`.
- **Resumen en texto** del gráfico de severidad, con los mismos valores, para
  lector de pantalla.
- Un valor `null` se muestra como **"Sin datos"** con una explicación breve,
  nunca como `0`.

## Reglas de negocio

- **RB-01** Cada tarjeta nueva muestra el valor tal como lo devuelve la API, sin
  recalcularlo en el cliente. → **CA-01, CA-05**
- **RB-02** `sprints_restantes` se muestra como `N sprints`; con `1`, `1 sprint`.
  Con `null` muestra "Sin datos" y la explicación "Hace falta la velocidad de
  al menos un sprint cerrado." → **CA-01, CA-02**
- **RB-03** `lead_time_estimado` con `lead_time_aproximado: true` muestra el
  valor con la etiqueta "aprox." y la nota de la Ley de Little. Si el valor es
  `null`, muestra "Sin datos" con la explicación "Hace falta la tasa de entrega
  de al menos un sprint cerrado." y **sin** etiqueta "aprox.". → **CA-03,
  CA-04**
- **RB-04** `tasa_entrega` `null` muestra "Sin datos" con la misma explicación
  que RB-03. Un `0` real (sprints cerrados sin historias hechas) se muestra
  como `0`, no como "Sin datos". → **CA-04**
- **RB-05** `wip` y `defectos_abiertos` nunca son `null`: `0` se muestra como
  `0`. → **CA-05**
- **RB-06** El gráfico de severidad muestra siempre las cuatro severidades en
  el orden fijo `minor`, `major`, `critical`, `blocker`, sin reordenar por
  valor. Si una clave falta dentro del objeto, se toma como `0`. → **CA-06,
  CA-07**
- **RB-07** El gráfico no usa solo el color para distinguir severidades: cada
  barra lleva su etiqueta y su valor. → **CA-08**
- **RB-08** El resumen en texto lista las cuatro severidades con sus valores,
  por ejemplo "Menor: 1, Mayor: 1, Crítica: 1, Bloqueante: 0". Con todos en
  `0`, el resumen dice "Sin defectos registrados." → **CA-08**
- **RB-09** **Compatibilidad con backend viejo:** si la respuesta no trae un
  campo (`undefined`, distinto de `null`), la tarjeta o el gráfico que lo usa
  **no se muestra**. `null` y ausente son casos distintos. → **CA-09**
- **RB-10** Las tarjetas nuevas comparten el estado de la pantalla de US-F21:
  carga, vacío, error y 404 se resuelven una sola vez en `MetricasPage`. Esta
  historia no define estados propios. → **CA-10**
- **RB-11** Las tarjetas y el gráfico nuevos aparecen **debajo** de los KPI y
  gráficos de US-F21 y no alteran su orden ni su contenido. → **CA-10**
- **RB-12** El aviso de US-F21 "Calculado sobre el estado actual de las
  historias" también aplica a estas métricas; no se duplica. → **CA-10**
- **RB-13** Los números se muestran con formato local (`es-AR`), con hasta dos
  decimales y sin ceros sobrantes (`2` y `2,5`, no `2,00`). → **CA-11**

## Restricciones

- **Sin cambios de backend** ni llamadas nuevas a la API.
- Se usa la **librería de gráficos y el componente de tarjeta de US-F21**; esta
  historia no incorpora dependencias nuevas.
- Fuera de alcance: coste en días-persona, desglose por sprint de estas
  métricas, diagramas de flujo acumulado, scatterplot, histograma y Monte Carlo.
- Textos de interfaz en español; tests con Vitest + React Testing Library,
  mockeando el cliente de métricas.
- Severidades traducidas con las etiquetas de `etiquetasDefecto.ts` (US-F20):
  `minor` → Menor, `major` → Mayor, `critical` → Crítica, `blocker` →
  Bloqueante. A la API no se envía nada.
- Estilos con los tokens del sistema visual existente; legibles en tema claro y
  oscuro.

## Casos límite

- Proyecto sin sprints cerrados: `sprints_restantes`, `tasa_entrega` y
  `lead_time_estimado` en `null` → tres tarjetas "Sin datos"; `wip` y defectos
  siguen con su valor.
- `lead_time_aproximado: false` con `lead_time_estimado: null`: sin etiqueta
  "aprox.".
- `tasa_entrega` en `0` real: se muestra `0`; el Lead Time queda `null` (el
  backend no divide por cero).
- `sprints_restantes` en `1`: "1 sprint".
- `sprints_restantes` en `0` (sin historias pendientes): "0 sprints".
- Todos los defectos por severidad en `0`: las cuatro barras en `0` y resumen
  "Sin defectos registrados."
- `defectos_por_severidad` ausente en la respuesta: no se muestra el gráfico.
- `defectos_por_severidad` con una clave faltante: esa severidad vale `0`.
- Respuesta de US-20 sin ningún campo de US-21 (backend viejo): se ve la
  pantalla de F21 sin tarjetas nuevas y sin errores.
- Lead Time no entero (`2,5`): se muestra `2,5 sprints`.

## Condiciones de error

Esta historia no agrega condiciones de error propias; hereda las de US-F21.

| Situación | Origen | Respuesta / UI |
| :--- | :--- | :--- |
| No miembro o proyecto inexistente | Guard US-F10/US-11 | `404 PROYECTO_NO_ENCONTRADO` → `MENSAJE_PROYECTO_INEXISTENTE` (de F21) |
| Error de red o de servidor | Cliente HTTP | Mensaje con `Reintentar` (de F21) |
| Campo de US-21 ausente | Backend viejo | La tarjeta o el gráfico no se muestra (RB-09) |

## Criterios de aceptación

- [ ] **CA-01** Given `sprints_restantes` en `7`, When se muestra, Then la
  tarjeta dice `7 sprints`.
- [ ] **CA-02** Given `sprints_restantes` en `null`, When se muestra, Then
  aparece "Sin datos" con su explicación y no `0`.
- [ ] **CA-03** Given Lead Time estimado con `lead_time_aproximado: true`, When
  se muestra, Then aparece la etiqueta "aprox." y la nota de la Ley de Little.
- [ ] **CA-04** Given Lead Time `null` (o tasa de entrega `null`), When se
  muestra, Then aparece "Sin datos" y no `0`, sin etiqueta "aprox.".
- [ ] **CA-05** Given `wip` y `defectos_abiertos` con valores de la API, When se
  muestran, Then cada tarjeta coincide con el valor de la API, incluido `0`.
- [ ] **CA-06** Given defectos sin ningún `blocker`, When se muestra el gráfico,
  Then `blocker` aparece con valor `0`.
- [ ] **CA-07** Given defectos de distintas severidades, When se muestra el
  gráfico, Then las barras van en el orden Menor, Mayor, Crítica, Bloqueante.
- [ ] **CA-08** Given el gráfico de severidad, When se consulta con lector de
  pantalla, Then existe un resumen en texto con los mismos valores.
- [ ] **CA-09** Given una respuesta sin los campos de US-21, When se muestra la
  pantalla, Then no aparecen las tarjetas ni el gráfico nuevos y no hay error.
- [ ] **CA-10** Given la pantalla de métricas, When carga, Then las tarjetas
  nuevas aparecen debajo de las de US-F21 sin modificarlas y comparten sus
  estados de carga, error y 404.
- [ ] **CA-11** Given una tasa de entrega de `2.5`, When se muestra, Then dice
  `2,5` y un `2` se ve como `2`.
- [ ] **CA-12** Given tema claro u oscuro, When se muestra el gráfico, Then
  etiquetas y valores son legibles y no dependen solo del color.

## Archivos afectados

| Archivo | Cambio |
| :--- | :--- |
| `docs/specs/US-F22-extensiones-metricas-spec.md` | **Nuevo** esta especificación |
| `docs/features/US-F22-extensiones-metricas.feature` | **Nuevo** escenarios BDD |
| `frontend/src/api/types.ts` | Los campos de US-21 de `MetricasProyecto` pasan a opcionales |
| `frontend/src/pages/proyecto/metricas/formato.ts` (+ test) | `formatearSprints` (singular y plural) |
| `frontend/src/pages/proyecto/metricas/severidad.ts` (+ test) | **Nuevo** desglose en orden fijo y resumen en texto |
| `frontend/src/pages/proyecto/metricas/TarjetaKpi.tsx` | Prop opcional `etiqueta` (para "aprox.") |
| `frontend/src/pages/proyecto/metricas/GraficoSeveridad.tsx` | **Nuevo** barras horizontales por severidad con tabla y resumen |
| `frontend/src/pages/proyecto/metricas/SeccionFlujo.tsx` | **Nuevo** tarjetas de flujo y defectos abiertos; oculta lo que la API no devuelve |
| `frontend/src/pages/proyecto/MetricasPage.tsx` | Monta `SeccionFlujo` debajo de los gráficos de US-F21 |
| `frontend/src/pages/proyecto/MetricasPage.extensiones.test.tsx` | **Nuevo** tests E-01 a E-20 |
| `frontend/src/pages/proyecto/MetricasPage.test.tsx` | El test E-13 de US-F21 cuenta 3 gráficos en vez de 2 |
