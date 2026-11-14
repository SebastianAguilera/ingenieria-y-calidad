# US-20 — Métricas de proyecto y sprint (núcleo §7)

## Objetivo

Exponer, en solo lectura, las métricas mínimas de `docs/proyecto.md` §7 para que
cualquier miembro de un proyecto sepa cuánto se planificó, cuánto se completó y
qué tan bien se estimó, tanto a nivel de un Sprint como del proyecto completo.

Las métricas se calculan al momento de la consulta a partir del estado actual de
historias, Sprints, asignaciones al Sprint Backlog, defectos y horas. No se
persisten resultados, no se crean tablas y no se modifica ningún dato.

**Trazabilidad:** Issue #101 → Esta especificación → Criterios de aceptación
(CA-01 a CA-22) → Escenarios BDD
(`docs/features/US-20-metricas.feature`) → Tests Go (TDD) → Código.

| Eslabón | Artefacto |
| :--- | :--- |
| Issue | [SebastianAguilera/ingenieria-y-calidad#101](https://github.com/SebastianAguilera/ingenieria-y-calidad/issues/101) |
| SDD | `docs/specs/US-20-metricas-spec.md` (este archivo) |
| Criterios de aceptación | CA-01 a CA-22 de este documento |
| BDD | `docs/features/US-20-metricas.feature` (E-01 a E-12) |
| Tests | `internal/domain/metricas_test.go`, `internal/repository/sprint_backlog_repository_test.go`, `internal/service/metricas_service_test.go`, `internal/handler/metricas_handler_test.go`, `cmd/api/router_test.go` |
| Código | `internal/domain/metricas.go`, `internal/service/metricas_service.go`, `internal/handler/metricas_handler.go`, `internal/handler/dto_metricas.go`, `internal/handler/metricas_servicio.go`, `cmd/api/router.go`, `cmd/api/main.go` |

## Dependencias y relación con otras historias

| Historia | Relación |
| :--- | :--- |
| US-01 | Aporta el proyecto y su existencia (404 `PROYECTO_NO_ENCONTRADO`). |
| US-02 | Aporta `Historia`, `Estado`, `StoryPoints` y `EsfuerzoReal`. |
| US-03 | Aporta `Sprint` y sus estados `Planificado`, `Activo` y `Cerrado`. |
| US-04 | Aporta `sprint_backlog_items`: la asignación historia↔Sprint. Define el conjunto de historias planificadas de un Sprint. |
| US-05 | Define la escala de `story_points` en `float64` (incluye `0` y `0.5`). |
| US-06 | Aporta `esfuerzo_real` (horas reales). |
| US-09 | Aporta `Defecto`, `EstadoDefecto`, `SprintDeteccionID` y `SprintResolucionID`. |
| US-18 | Confirma el filtrado de defectos por Sprint de detección/resolución y los estados de resolución `resolved`/`closed`. |
| US-19 | Aporta `horas_estimadas` por historia. |
| US-11 | Guards compartidos y criterio "cualquier miembro consulta". |
| US-F10 | Guard de membresía: no miembro recibe `404 PROYECTO_NO_ENCONTRADO`. |
| US-21 | Extiende estas métricas (sprints restantes, WiP, tasa de entrega, Lead Time, defectos abiertos, severidad). **No se implementa aquí.** |
| US-22 | Consume estas métricas para el reporte. **No se implementa aquí.** |
| US-F21 | Dashboard que consumirá estos endpoints. **No se implementa aquí.** |

**Bloquea a:** US-21, US-22 y US-F21.

## Alcance

**Incluye**

- `GET /api/proyectos/{id}/sprints/{sprintId}/metricas`.
- `GET /api/proyectos/{id}/metricas`.
- Cálculo de SP planificados/completados, porcentaje completado, historias
  completadas, horas estimadas/reales, desviación de esfuerzo, defectos
  detectados/resueltos, velocidad, porcentaje de historias completadas y ratio de
  defectos.
- Funciones puras de cálculo en `internal/domain`, con tests de tabla.
- Nuevo método de lectura `ListarPorProyecto` en `SprintBacklogRepository` para
  traer todas las asignaciones del proyecto en una sola consulta.
- Service de métricas, handler, DTO snake_case y rutas detrás del guard de
  membresía (US-F10).
- Indicador `base: "estado_actual"` y advertencia de histórico.

**No incluye**

- Sprints restantes, WiP, tasa de entrega, Lead Time, defectos abiertos ni
  desglose por severidad (US-21).
- Reporte de proyecto/Sprint (US-22).
- Diagramas de flujo, Monte Carlo, ROI ni WSJF (US-23).
- Dashboard o frontend (US-F21).
- Snapshots históricos, cacheo o persistencia de resultados.
- Tablas nuevas, migraciones, cambios en `tablas.sql` o escrituras de cualquier
  tipo.
- Filtros, paginación o parámetros de consulta.

## Entradas y precondiciones

### Precondiciones

- Esquema existente de US-01 a US-06, US-09, US-18 y US-19 (`AutoMigrate`).
- Request autenticado con JWT válido (US-10); el `sub` identifica al usuario.
- El usuario es miembro del proyecto (US-F10); cualquier rol alcanza.
- Para el endpoint de Sprint, el Sprint pertenece al proyecto indicado.

### Endpoints y parámetros

| Endpoint | Método | Cuerpo | Query |
| :--- | :---: | :---: | :---: |
| `/api/proyectos/{id}/sprints/{sprintId}/metricas` | `GET` | No | No |
| `/api/proyectos/{id}/metricas` | `GET` | No | No |

| Parámetro | Ubicación | Obligatorio | Tipo / formato | Límites |
| :--- | :--- | :---: | :--- | :--- |
| `id` | path | Sí | entero positivo (`uint`) | `> 0`; no numérico o `0` → `400 PARAMETRO_INVALIDO` |
| `sprintId` | path | Sí | entero positivo (`uint`) | `> 0`; no numérico o `0` → `400 PARAMETRO_INVALIDO` |

### Request

```http
GET /api/proyectos/1/sprints/10/metricas
Authorization: Bearer <jwt>
```

```http
GET /api/proyectos/1/metricas
Authorization: Bearer <jwt>
```

Sin cuerpo y sin parámetros de consulta. Los query params desconocidos se
ignoran, como en el resto de los `GET` del proyecto.

## Salidas esperadas y contrato JSON

Ambos endpoints responden `200 OK` con un objeto JSON de claves en
`snake_case`. Los porcentajes y ratios son `float64` en el rango `0..1` salvo
`ratio_defectos` (ver RN-12). **La API no redondea**: los ejemplos muestran el
valor completo de `float64`; las tablas explicativas usan 2 decimales solo para
lectura.

### `GET .../sprints/{sprintId}/metricas` → `200`

Claves mínimas y tipos:

| Clave | Tipo JSON | Nulo | Descripción |
| :--- | :--- | :---: | :--- |
| `sprint_id` | entero | No | `sprintId` consultado. |
| `base` | string | No | Siempre `"estado_actual"`. |
| `story_points_planificados` | número | No | Suma de SP de las historias asignadas al Sprint. |
| `story_points_completados` | número | No | Suma de SP de las asignadas en `Hecho`. |
| `porcentaje_completado` | número | Sí | `completados / planificados` en `0..1`; `null` si planificados es `0`. |
| `historias_completadas` | entero | No | Cantidad de asignadas en estado `Hecho`. |
| `horas_estimadas` | número | No | Suma de `horas_estimadas` no nulas de las asignadas; `0` si no hay. |
| `horas_reales` | número | No | Suma de `esfuerzo_real` no nulos de las asignadas; `0` si no hay. |
| `desviacion_esfuerzo` | número | Sí | Desviación con historias que tienen ambas horas; `null` sin base. |
| `defectos_detectados` | entero | No | Defectos con `sprint_deteccion_id = sprintId`. |
| `defectos_resueltos` | entero | No | Defectos con `sprint_resolucion_id = sprintId` y estado `resolved`/`closed`. |

```json
{
  "sprint_id": 10,
  "base": "estado_actual",
  "story_points_planificados": 13,
  "story_points_completados": 8,
  "porcentaje_completado": 0.6153846153846154,
  "historias_completadas": 2,
  "horas_estimadas": 20,
  "horas_reales": 24,
  "desviacion_esfuerzo": 0.2,
  "defectos_detectados": 2,
  "defectos_resueltos": 1
}
```

### `GET .../metricas` → `200`

Claves mínimas y tipos:

| Clave | Tipo JSON | Nulo | Descripción |
| :--- | :--- | :---: | :--- |
| `proyecto_id` | entero | No | `id` consultado. |
| `base` | string | No | Siempre `"estado_actual"`. |
| `velocidad` | número | Sí | Promedio de SP completados de Sprints `Cerrado`; `null` sin cerrados. |
| `porcentaje_historias_completadas` | número | Sí | Historias `Hecho` / total de historias; `null` sin historias. |
| `horas_estimadas` | número | No | Suma de `horas_estimadas` no nulas de todas las historias; `0` si no hay. |
| `horas_reales` | número | No | Suma de `esfuerzo_real` no nulos de todas las historias; `0` si no hay. |
| `desviacion_esfuerzo` | número | Sí | Desviación agregada con historias que tienen ambas horas; `null` sin base. |
| `defectos_detectados` | entero | No | Todos los defectos del proyecto (incluye `rejected`). |
| `defectos_resueltos` | entero | No | Defectos en estado `resolved` o `closed`. |
| `ratio_defectos` | número | Sí | Detectados **sin `rejected`** / historias completadas; `null` si no hay completadas. Sin cota superior. |

```json
{
  "proyecto_id": 1,
  "base": "estado_actual",
  "velocidad": 6,
  "porcentaje_historias_completadas": 0.5,
  "horas_estimadas": 40,
  "horas_reales": 48,
  "desviacion_esfuerzo": 0.2,
  "defectos_detectados": 3,
  "defectos_resueltos": 2,
  "ratio_defectos": 0.3333333333333333
}
```

> **Decisión de contrato:** no se agregan `historias_completadas` ni
> `historias_totales` al endpoint de proyecto. El issue pide el porcentaje y el
> ratio, y §7 no exige los conteos absolutos a nivel proyecto. Mantener el
> contrato mínimo evita inventar campos. Si US-21/US-F21 los necesita, se
> agregarán en una historia posterior. El conjunto de claves emitido es
> exactamente el de las tablas anteriores.

### Casos sin datos (no son error)

Proyecto sin datos:

```json
{
  "proyecto_id": 1,
  "base": "estado_actual",
  "velocidad": null,
  "porcentaje_historias_completadas": null,
  "horas_estimadas": 0,
  "horas_reales": 0,
  "desviacion_esfuerzo": null,
  "defectos_detectados": 0,
  "defectos_resueltos": 0,
  "ratio_defectos": null
}
```

Sprint sin historias asignadas:

```json
{
  "sprint_id": 10,
  "base": "estado_actual",
  "story_points_planificados": 0,
  "story_points_completados": 0,
  "porcentaje_completado": null,
  "historias_completadas": 0,
  "horas_estimadas": 0,
  "horas_reales": 0,
  "desviacion_esfuerzo": null,
  "defectos_detectados": 0,
  "defectos_resueltos": 0
}
```

### Cuerpo estándar de error

```json
{
  "error": "CODIGO",
  "mensaje": "Descripción en español",
  "campo": "campo opcional"
}
```

`campo` se omite cuando está vacío. Las respuestas de error no incluyen `base`.

## Fórmulas y modelo de datos

### Notación

- `H(s)`: historias asignadas al Sprint `s` mediante `sprint_backlog_items`
  (`proyecto_id = P`, `sprint_id = s`).
- `SP(h)`: `h.story_points`; si es `null`, vale `0`.
- `Hecho(h)`: `h.estado == "Hecho"`.
- `E(h)`: `h.horas_estimadas`; `null` si no está registrada.
- `R(h)`: `h.esfuerzo_real`; `null` si no está registrado.
- `Pares(X)`: subconjunto de historias de `X` con `E(h) != null` **y**
  `R(h) != null` simultáneamente.
- `Cerrados(P)`: Sprints del proyecto con `estado == "Cerrado"`.

### Métricas de Sprint

```text
SP_plan(s)      = Σ_{h ∈ H(s)} SP(h)
SP_comp(s)      = Σ_{h ∈ H(s) ∧ Hecho(h)} SP(h)
pct_comp(s)     = SP_plan(s) == 0 ? null : SP_comp(s) / SP_plan(s)
hist_comp(s)    = |{ h ∈ H(s) : Hecho(h) }|
horas_est(s)    = Σ_{h ∈ H(s) ∧ E(h) != null} E(h)
horas_real(s)   = Σ_{h ∈ H(s) ∧ R(h) != null} R(h)
desv(s)         = Pares(H(s)) == ∅ ∨ Σ_{Pares(H(s))} E == 0
                    ? null
                    : (Σ_{Pares(H(s))} R − Σ_{Pares(H(s))} E) / Σ_{Pares(H(s))} E
def_det(s)      = |{ d ∈ Defectos(P) : d.sprint_deteccion_id == s }|
def_res(s)      = |{ d ∈ Defectos(P) : d.sprint_resolucion_id == s
                                      ∧ d.estado ∈ {resolved, closed} }|
```

### Métricas de proyecto

```text
Historias(P)    = todas las historias del proyecto
vel(P)          = Cerrados(P) == ∅
                    ? null
                    : (Σ_{s ∈ Cerrados(P)} SP_comp(s)) / |Cerrados(P)|
pct_hist(P)     = |Historias(P)| == 0
                    ? null
                    : |{ h ∈ Historias(P) : Hecho(h) }| / |Historias(P)|
horas_est(P)    = Σ_{h ∈ Historias(P) ∧ E(h) != null} E(h)
horas_real(P)   = Σ_{h ∈ Historias(P) ∧ R(h) != null} R(h)
desv(P)         = Pares(Historias(P)) == ∅ ∨ Σ_{Pares(Historias(P))} E == 0
                    ? null
                    : (Σ_{Pares(Historias(P))} R − Σ_{Pares(Historias(P))} E)
                      / Σ_{Pares(Historias(P))} E
def_det(P)      = |Defectos(P)|                       // incluye rejected
def_res(P)      = |{ d ∈ Defectos(P) : d.estado ∈ {resolved, closed} }|
hist_comp(P)    = |{ h ∈ Historias(P) : Hecho(h) }|
ratio_def(P)    = hist_comp(P) == 0
                    ? null
                    : |{ d ∈ Defectos(P) : d.estado != rejected }| / hist_comp(P)
```

> **Totales vs. desviación.** `horas_estimadas` y `horas_reales` son totales
> independientes de los valores disponibles: una historia con horas estimadas y
> sin horas reales suma al primero pero no al segundo. La desviación, en cambio,
> usa solo el subconjunto pareado (`Pares`) y por eso puede no coincidir con
> `(horas_reales − horas_estimadas) / horas_estimadas`. Ejemplo: A (est 5, real
> 8), B (est 5, real 4), C (est 3, real `null`) → totales est `13`, real `12`,
> desviación `(12 − 10) / 10 = 0.2`.

### Estructuras de dominio (referencia)

Los nombres son indicativos; el requisito es que sean funciones **puras** en
`internal/domain`, sin dependencias de Gin, GORM ni contexto, y que los tests
sean de tabla.

```go
type MetricasSprint struct {
    StoryPointsPlanificados float64
    StoryPointsCompletados  float64
    PorcentajeCompletado    *float64
    HistoriasCompletadas    int
    HorasEstimadas          float64
    HorasReales             float64
    DesviacionEsfuerzo      *float64
    DefectosDetectados      int
    DefectosResueltos       int
}

type MetricasProyecto struct {
    Velocidad                      *float64
    PorcentajeHistoriasCompletadas *float64
    HorasEstimadas                 float64
    HorasReales                    float64
    DesviacionEsfuerzo             *float64
    DefectosDetectados             int
    DefectosResueltos              int
    RatioDefectos                  *float64
}
```

Funciones puras sugeridas (o equivalentes):

```go
func CalcularSPPlanificados(historias []Historia) float64
func CalcularSPCompletados(historias []Historia) float64
func CalcularPorcentajeCompletado(completados, planificados float64) *float64
func ContarHistoriasCompletadas(historias []Historia) int
func SumarHorasEstimadas(historias []Historia) float64
func SumarHorasReales(historias []Historia) float64
func CalcularDesviacionEsfuerzo(historias []Historia) *float64
func CalcularVelocidad(sprints []Sprint, asignaciones []SprintBacklogItem,
    historiasPorID map[uint]Historia) *float64
func CalcularRatioDefectos(defectos []Defecto, historiasCompletadas int) *float64
func CalcularMetricasSprint(historias []Historia, defectos []Defecto,
    sprintID uint) MetricasSprint
func CalcularMetricasProyecto(sprints []Sprint, asignaciones []SprintBacklogItem,
    historias []Historia, defectos []Defecto) MetricasProyecto
```

`SprintID` y `ProyectoID` se agregan en el DTO del handler junto con `base`; no
son parte del cálculo puro.

### Campos existentes utilizados (verificados en el código)

| Struct | Campos |
| :--- | :--- |
| `domain.Historia` | `ID`, `ProyectoID`, `Estado` (`Backlog`/`En Progreso`/`Hecho`), `StoryPoints *float64`, `EsfuerzoReal *float64`, `HorasEstimadas *float64` |
| `domain.Sprint` | `ID`, `ProyectoID`, `Estado` (`Planificado`/`Activo`/`Cerrado`) |
| `domain.SprintBacklogItem` | `ID`, `ProyectoID`, `SprintID`, `HistoriaID`, `Orden` |
| `domain.Defecto` | `ID`, `ProyectoID`, `SprintDeteccionID`, `SprintResolucionID *uint`, `Estado` (`open`/`in_progress`/`resolved`/`closed`/`rejected`) |

No se requieren campos nuevos ni cambios de esquema. El índice único
`ux_sprint_backlog_historia` garantiza que una historia tenga como máximo una
asignación vigente, por lo que no hay doble conteo dentro de un Sprint.

### Consultas y ausencia de N+1

| Operación | Consultas fijas |
| :--- | :--- |
| Métricas de Sprint | `proyectos.ObtenerPorID` + `sprints.ObtenerPorID` + `sprintBacklog.ObtenerPorSprint` + `historias.ObtenerPorIDs` (una sola vez por lote) + `defectos.ListarPorProyecto` |
| Métricas de proyecto | `proyectos.ObtenerPorID` + `sprints.ListarPorProyecto` + `sprintBacklog.ListarPorProyecto` (método nuevo) + `historias.ListarPorProyecto` + `defectos.ListarPorProyecto` |

El número de consultas no depende de la cantidad de Sprints ni de historias. Si
el Sprint no tiene asignaciones, se evita la consulta de historias y se calcula
con listas vacías. El guard de membresía agrega su propia consulta de
`PerteneceA`, ya existente en US-F10.

## Reglas de negocio

| ID | Regla |
| :--- | :--- |
| RN-01 | Los endpoints son de solo lectura: no crean, actualizan ni eliminan filas, no ejecutan transacciones de escritura y no modifican `tablas.sql`. |
| RN-02 | Cualquier miembro del proyecto puede consultar ambos endpoints; no se exige rol `LIDER`. Un no miembro recibe `404 PROYECTO_NO_ENCONTRADO`, igual que un proyecto inexistente. |
| RN-03 | El Sprint debe existir dentro del proyecto. Un Sprint inexistente o de otro proyecto produce `404 SPRINT_NO_ENCONTRADA`. |
| RN-04 | `H(s)` se obtiene de `sprint_backlog_items` filtrando por `proyecto_id` y `sprint_id`. No se incluyen historias del Product Backlog no asignadas. |
| RN-05 | `story_points_planificados` suma `story_points` de todas las historias asignadas. Una historia con `story_points = null` aporta `0`. |
| RN-06 | `story_points_completados` suma `story_points` de las asignadas en estado `Hecho`. Una historia `Hecho` con `story_points = null` aporta `0`. |
| RN-07 | `porcentaje_completado` = completados / planificados, en `0..1`; si planificados es `0` es `null`. Con planificados `> 0` y completados `0` es `0`, no `null`. |
| RN-08 | `historias_completadas` cuenta las historias asignadas en `Hecho`, sin importar si tienen o no `story_points`. |
| RN-09 | `horas_estimadas` y `horas_reales` son sumas de los valores no nulos de las historias consideradas (asignadas en Sprint; todas en proyecto). Una lista sin valores produce `0`, nunca `null`. |
| RN-10 | `desviacion_esfuerzo` = `(ΣR − ΣE) / ΣE` usando solo historias con `E` y `R` no nulos. Si no hay historias pareadas o `ΣE == 0`, es `null`. Aplica igual a Sprint y proyecto. Puede ser negativa. |
| RN-11 | `defectos_detectados` de un Sprint cuenta los defectos con `sprint_deteccion_id = sprintId`, **incluidos los `rejected`** (la exclusión de RN-13 aplica solo al ratio). `defectos_resueltos` cuenta los de `sprint_resolucion_id = sprintId` cuyo estado actual sea `resolved` o `closed`. |
| RN-12 | `defectos_detectados` del proyecto cuenta todos los defectos del proyecto, incluidos los `rejected`. `defectos_resueltos` cuenta los que estén en `resolved` o `closed`. |
| RN-13 | `ratio_defectos` = detectados **excluyendo `rejected`** / historias completadas del proyecto. Si no hay historias completadas es `null`. No se acota a `1`: puede ser `0` o mayor que `1`. |
| RN-14 | `velocidad` = promedio aritmético de `SP_comp` de los Sprints `Cerrado`. Los Sprints `Planificado` y `Activo` no entran. Sin Sprints cerrados es `null`. |
| RN-15 | Un Sprint cerrado sin historias o sin SP aporta `0` al promedio; por eso un proyecto con cerrados y SP total `0` devuelve `0`, que no se confunde con `null` (sin cerrados). |
| RN-16 | El divisor de `velocidad` es la cantidad de Sprints `Cerrado`, incluidos los que aportan `0`. |
| RN-17 | `porcentaje_historias_completadas` = historias `Hecho` / total de historias del proyecto, en `0..1`. Sin historias es `null`; con historias y ninguna `Hecho` es `0`. |
| RN-18 | Porcentajes y ratios se devuelven como `float64` en `0..1` (salvo el ratio, sin cota superior) y **no se redondean** en la API. |
| RN-19 | Ambas respuestas exitosas incluyen `base: "estado_actual"`. El sistema no guarda snapshots: un Sprint cerrado refleja el estado actual de sus historias y defectos. Si una historia cambia de estado, de SP o de horas después del cierre, la métrica cambia. |
| RN-20 | Un defecto detectado en el Sprint 2 y resuelto (`resolved`/`closed`) en el Sprint 3 suma como detectado en el Sprint 2 y como resuelto en el Sprint 3 (y en ninguno de los dos como lo contrario). Si detección y resolución coinciden en el mismo Sprint, suma en ambos conteos de ese Sprint. |
| RN-21 | No se agregan claves al contrato más allá de las listadas; en particular el endpoint de proyecto no emite `historias_completadas` ni `historias_totales`. |
| RN-22 | No hay paginación, filtros ni parámetros de consulta. Los resultados son únicos por recurso. |

> **Nota de nomenclatura:** el error de dominio real es
> `domain.ErrSprintNoEncontrado` (masculino), ya existente. El código HTTP de la
> API es `SPRINT_NO_ENCONTRADA` (femenino), definido en `handler/errores.go`. No
> se crea un error nuevo con el nombre `ErrSprintNoEncontrada`.

## Restricciones

### Técnicas

- Solo se ejecutan sentencias `SELECT`; ninguna operación de escritura, bloqueo
  de filas ni transacción de escritura.
- No se crean tablas, columnas, índices ni migraciones. No se modifica
  `docs/tablas.sql` ni `internal/repository/migraciones.go`.
- No se agregan errores centinela, endpoints, campos ni dependencias externas.
- Se reutilizan `parsearID`, `traducirError` y el guard `ExigirAccesoAProyecto`.
- El número de consultas es fijo y no depende de la cantidad de Sprints ni de
  historias (sin N+1). El método nuevo filtra por `proyecto_id` usando los
  índices existentes de `sprint_backlog_items`.
- Las respuestas no se cachean; cada request recalcula.
- No se abre una transacción de lectura larga ni se garantiza una foto atómica
  entre tablas (D-08).

### De arquitectura por capas

- `domain` no importa Gin, GORM, `context` ni paquetes de `repository`,
  `service` o `handler`.
- `repository` implementa el puerto de dominio; no contiene reglas de cálculo.
- `service` orquesta lecturas y validaciones de existencia; delega las fórmulas
  al dominio.
- `handler` solo traduce HTTP, DTO y errores; no calcula métricas.
- Las rutas nuevas quedan detrás del middleware de autenticación y del guard de
  membresía, sin middleware de autorización adicional.

### De negocio y contrato

- Solo lectura; no hay snapshots ni histórico persistido (`base:
  "estado_actual"`).
- Sin redondeo en la API; porcentajes en `0..1` y ratio sin cota superior.
- Sin paginación, filtros ni query params.
- Contrato mínimo del issue: no se agregan `historias_completadas` ni
  `historias_totales` al proyecto.
- Una consulta de un no miembro no revela la existencia del proyecto ni del
  Sprint (`404 PROYECTO_NO_ENCONTRADO`).

## Casos límite

| ID | Caso | Resultado |
| :--- | :--- | :--- |
| CL-01 | Sprint sin historias asignadas | `200`; SP plan/completados `0`, `porcentaje_completado null`, `historias_completadas 0`, horas `0`, desviación `null`. |
| CL-02 | Sprint con historias asignadas pero todas con `story_points: null` | SP `0/0`; `porcentaje_completado null`; `historias_completadas` cuenta las `Hecho` aunque su SP sea `null`. |
| CL-03 | Sprint `Planificado` o `Activo` | `200` con sus métricas; se puede consultar cualquier estado. |
| CL-04 | Sprint `Cerrado` | `200`; sus métricas reflejan el estado actual (`base: estado_actual`). |
| CL-05 | Sprint cerrado sin historias o sin SP | Aporta `0` a la velocidad. Si es el único cerrado, `velocidad = 0`. |
| CL-06 | Proyecto sin Sprints cerrados (aunque tenga un `Activo` con SP) | `velocidad = null`. |
| CL-07 | Proyecto sin historias | `porcentaje_historias_completadas = null`; horas `0`; desviación `null`; sin completadas, `ratio_defectos = null`. |
| CL-08 | Historia con `horas_estimadas` pero sin `esfuerzo_real` (o viceversa) | Suma al total de horas disponible que corresponda, pero queda excluida de la desviación. |
| CL-09 | Historia sin `horas_estimadas` y sin `esfuerzo_real` | No suma a totales ni a desviación. |
| CL-10 | Historia sin `story_points` | Aporta `0` a planificados y completados; sigue contando en `historias_completadas` si está `Hecho`. |
| CL-11 | Defecto `rejected` | Suma a `defectos_detectados` (Sprint y proyecto) por RN-12, pero **no** al numerador de `ratio_defectos`. |
| CL-12 | Solo defectos `rejected` y 6 historias completadas | `ratio_defectos = 0` (0/6), no `null`. |
| CL-13 | Defecto detectado en Sprint 2 y `closed` en Sprint 3 | Sprint 2 lo cuenta como detectado; Sprint 3 como resuelto. |
| CL-14 | Defecto detectado y resuelto en el mismo Sprint | Suma en `defectos_detectados` y en `defectos_resueltos` de ese Sprint. |
| CL-15 | Desviación con `ΣE == 0` o sin historias pareadas | `null`, tanto en Sprint como en proyecto. |
| CL-16 | Desviación negativa (horas reales menores que estimadas) | Se devuelve el valor negativo (ej. `-0.3`); no se acota a `0`. |
| CL-17 | Valores fraccionarios (`story_points` `0.5`, horas decimales) | Se suman y promedian como `float64`; no se redondean en la API. |
| CL-18 | Usuario no miembro (en cualquiera de los dos endpoints) | `404 PROYECTO_NO_ENCONTRADO`. |
| CL-19 | Proyecto inexistente con `id` válido | `404 PROYECTO_NO_ENCONTRADO`. |
| CL-20 | Sprint inexistente o de otro proyecto | `404 SPRINT_NO_ENCONTRADA`. |
| CL-21 | `{id}` o `{sprintId}` no numérico o `0` | `400 PARAMETRO_INVALIDO`. |
| CL-22 | Dos consultas consecutivas sin cambios en los datos | Devuelven exactamente el mismo resultado; no hay cache ni efectos laterales. |

## Criterios de aceptación

- [ ] **CA-01** Given un Sprint con historias asignadas en estados mixtos, When
  se consultan sus métricas, Then `story_points_planificados` es la suma de todas,
  `story_points_completados` la suma de las `Hecho` y `porcentaje_completado` es
  completados/planificados en `0..1`.
- [ ] **CA-02** Given un Sprint sin historias asignadas, When se consultan sus
  métricas, Then SP planificados y completados son `0`, `porcentaje_completado`
  es `null` y `historias_completadas` es `0`.
- [ ] **CA-03** Given dos Sprints cerrados con 8 y 4 SP completados, When se
  consulta el proyecto, Then `velocidad` es `6`.
- [ ] **CA-04** Given un Sprint `Activo` con SP completados, When se consulta la
  velocidad, Then ese Sprint no entra en el promedio.
- [ ] **CA-05** Given un proyecto sin Sprints cerrados, When se consulta, Then
  `velocidad` es `null`.
- [ ] **CA-06** Given historias con horas reales 8 y 4 y estimadas 5 y 5, When
  se consulta el total del proyecto, Then `desviacion_esfuerzo` es `0.2`.
- [ ] **CA-07** Given una historia sin horas reales o sin horas estimadas, When
  se calcula la desviación, Then esa historia no entra en el cálculo, aunque sí
  sume al total de horas disponible que corresponda.
- [ ] **CA-08** Given historias sin `story_points`, When se calculan los
  completados, Then cuentan como `0` y no alteran `historias_completadas`.
- [ ] **CA-09** Given 3 defectos detectados (uno `rejected`) y 6 historias
  completadas, When se consulta el proyecto, Then `ratio_defectos` es
  `0.3333333333333333` (2/6), mientras `defectos_detectados` es `3`.
- [ ] **CA-10** Given 0 historias completadas, When se consulta el proyecto,
  Then `ratio_defectos` es `null`.
- [ ] **CA-11** Given un defecto detectado en el Sprint 2 y resuelto (`closed`)
  en el Sprint 3, When se consultan las métricas de cada Sprint, Then el Sprint 2
  lo cuenta como detectado y el Sprint 3 como resuelto.
- [ ] **CA-12** Given una respuesta exitosa de cualquiera de los dos endpoints,
  When se lee, Then incluye `base: "estado_actual"` y todas las claves del
  contrato con sus tipos (números, enteros, `null` donde corresponde).
- [ ] **CA-13** Given un no miembro, When consulta cualquiera de los dos
  endpoints, Then `404 PROYECTO_NO_ENCONTRADO`.
- [ ] **CA-14** Given un Sprint de otro proyecto o inexistente, When se consultan
  sus métricas, Then `404 SPRINT_NO_ENCONTRADA`.
- [ ] **CA-15** Given cualquier consulta, When se completa, Then no se modifica
  ningún dato (ni historias, ni Sprints, ni asignaciones, ni defectos).
- [ ] **CA-16** Given un Sprint con horas estimadas y reales, When se consultan
  sus métricas, Then `horas_estimadas` y `horas_reales` son las sumas de los
  valores no nulos de las asignadas y `desviacion_esfuerzo` usa solo las
  historias con ambos valores.
- [ ] **CA-17** Given un proyecto con Sprints cerrados pero SP completados `0`,
  When se consulta, Then `velocidad` es `0` (no `null`).
- [ ] **CA-18** Given un proyecto sin historias, When se consulta, Then
  `porcentaje_historias_completadas` es `null` y `ratio_defectos` es `null`.
- [ ] **CA-19** Given un defecto `rejected`, When se consulta el proyecto, Then
  suma a `defectos_detectados` pero no al numerador de `ratio_defectos`.
- [ ] **CA-20** Given un `{id}` o `{sprintId}` no numérico o `0`, When se
  consulta, Then `400 PARAMETRO_INVALIDO`.
- [ ] **CA-21** Given una respuesta del endpoint de proyecto, Then sus claves
  son exactamente las del contrato y no incluye `historias_completadas` ni
  `historias_totales`.
- [ ] **CA-22** Given un proyecto con `historias Hecho > 0` y `total de
  historias > 0`, Then `porcentaje_historias_completadas` es
  `Hecho / total` en `0..1`, con `0` cuando no hay ninguna `Hecho`.

## Condiciones de error y mapeo de errores HTTP

| Situación | Error de dominio | Estado | Código API | Campo |
| :--- | :--- | :---: | :--- | :--- |
| `{id}` o `{sprintId}` no numérico o `0` | — (handler) | 400 | `PARAMETRO_INVALIDO` | — |
| Sin token o token inválido | `ErrTokenRequerido` / `ErrTokenInvalido` | 401 | `TOKEN_REQUERIDO` / `TOKEN_INVALIDO` | — |
| Proyecto inexistente o usuario no miembro | `ErrProyectoNoEncontrado` | 404 | `PROYECTO_NO_ENCONTRADO` | — |
| Sprint inexistente o de otro proyecto | `ErrSprintNoEncontrado` | 404 | `SPRINT_NO_ENCONTRADA` | — |
| Error de persistencia al leer | `ErrErrorDePersistencia` (o el error envuelto) | 500 | `ERROR_INTERNO` | — |

Precedencia:

1. El middleware de autenticación resuelve `401`.
2. El guard de membresía (US-F10) se ejecuta antes del handler: con `:id` válido
   y usuario no miembro responde `404 PROYECTO_NO_ENCONTRADO` (aunque
   `sprintId` sea inválido). Con `:id` inválido, el guard deja pasar y el handler
   responde `400`.
3. En el handler, `400 PARAMETRO_INVALIDO` por `{id}` o `{sprintId}` inválido.
4. En el service, primero la existencia del proyecto (`404
   PROYECTO_NO_ENCONTRADO`) y luego la del Sprint (`404 SPRINT_NO_ENCONTRADA`).

No se agregan errores de dominio nuevos.

## Arquitectura (Clean Architecture)

- **domain**: `metricas.go` con `MetricasSprint`, `MetricasProyecto` y funciones
  puras de cálculo. Sin contexto, sin Gin, sin GORM. `SprintBacklogRepository`
  gana `ListarPorProyecto(ctx, proyectoID uint) ([]SprintBacklogItem, error)`.
- **repository**: implementación GORM de `ListarPorProyecto` sobre
  `sprint_backlog_items`, filtrando por `proyecto_id`, ordenando por
  `sprint_id ASC, orden ASC, id ASC` y devolviendo slice vacío (nunca `nil`).
  Sin cambios de esquema ni de `tablas.sql`.
- **service**: `MetricasService` con `MetricasDeSprint(ctx, proyectoID,
  sprintID)` y `MetricasDeProyecto(ctx, proyectoID)`. Valida proyecto y Sprint,
  carga datos con consultas por proyecto y delega el cálculo al dominio. Envuelve
  errores con `%w` para conservar `errors.Is`.
- **handler**: `metricas_servicio.go` define el puerto `ServicioMetricas`;
  `dto_metricas.go` define las respuestas snake_case y agrega `base:
  "estado_actual"`; `metricas_handler.go` registra:
  - `GET /api/proyectos/:id/sprints/:sprintId/metricas`
  - `GET /api/proyectos/:id/metricas`

  Ambos usan `parsearID` (400) y `traducirError` (404/500). No hay cuerpo ni
  binding.
- **cmd/api**: `setupRouter` recibe el nuevo `ServicioMetricas` y llama a
  `handler.RegistrarRutasMetricas` después de `r.Use(handler.
  ExigirAccesoAProyecto(...))`, por lo que ambas rutas quedan detrás del guard
  sin middleware adicional. `main.go` construye el service inyectando
  `ProyectoRepository`, `SprintRepository`, `HistoriaRepository`,
  `SprintBacklogRepository` y `DefectoRepository`.

No se modifica `internal/repository/migraciones.go` ni `docs/tablas.sql`.

## Plan de pruebas (TDD)

Ciclo **RED → GREEN → REFACTOR**. Cada ID de escenario BDD se reutiliza como
nombre de subtest.

### Dominio (`internal/domain/metricas_test.go`)

- `TestCalcularSPPlanificados` / `TestCalcularSPCompletados`: cero historias,
  una historia, valores iguales, valores mixtos, `story_points` `nil`, `Hecho` y
  no `Hecho`.
- `TestCalcularPorcentajeCompletado`: planificados `0` → `nil`; completados `0`
  con planificados `> 0` → `0`; 8/13; 0.5/1.
- `TestContarHistoriasCompletadas`: sin historias, una, varias, con SP `nil`.
- `TestSumarHorasEstimadas` / `TestSumarHorasReales`: sin valores → `0`; valores
  parciales; decimales.
- `TestCalcularDesviacionEsfuerzo`: sin pares → `nil`; `ΣE == 0` → `nil`;
  `(8,5),(4,5)` → `0.2`; historia sin horas excluida; desviación negativa;
  valores iguales → `0`.
- `TestCalcularVelocidad`: sin cerrados → `nil`; un cerrado → su SP; dos
  cerrados 8 y 4 → `6`; activo excluido; cerrado con 0 incluido en el divisor;
  varios cerrados en `0` → `0`.
- `TestCalcularRatioDefectos`: sin completadas → `nil`; 3 detectados (1
  `rejected`) y 6 completadas → `2/6`; solo `rejected` → `0`; ratio `> 1`.
- `TestCalcularMetricasSprint` y `TestCalcularMetricasProyecto`: tablas de casos
  completos (cero, un elemento, valores iguales, valores mixtos) que verifiquen
  todas las claves del struct.

### Repositorio (`internal/repository/sprint_backlog_repository_test.go`)

- `TestSprintBacklogRepositorioListarPorProyecto`: devuelve asignaciones de dos
  Sprints en una sola llamada; excluye asignaciones de otro proyecto; devuelve
  slice vacío (no `nil`) sin asignaciones; orden estable.
- `TestMetricasProyectoConsultasFijasSinNPlus1`: con un contador de sentencias
  GORM (callback) se verifica que la cantidad de consultas es fija e
  independiente de la cantidad de Sprints (por ejemplo, mismo número con 2 y con
  10 Sprints), que `ListarPorProyecto` se invoca una sola vez y que
  `ObtenerPorSprint` no se llama una vez por Sprint para calcular la velocidad.

### Service (`internal/service/metricas_service_test.go`)

- `TestMetricasServiceSprint`: arma el `MetricasSprint` esperado; verifica que
  usa `ObtenerPorSprint` y `ObtenerPorIDs` una vez; Sprint sin historias no
  consulta historias; defectos filtrados por Sprint.
- `TestMetricasServiceProyecto`: verifica que usa `sprints.ListarPorProyecto`,
  `sprintBacklog.ListarPorProyecto`, `historias.ListarPorProyecto` y
  `defectos.ListarPorProyecto` una vez cada uno (sin llamadas por Sprint).
- Errores: `ErrProyectoNoEncontrado` y `ErrSprintNoEncontrado` se propagan con
  `errors.Is`; Sprint de otro proyecto se comporta como inexistente.
- Solo lectura: los mocks solo exponen métodos de lectura y se verifica que no se
  invoca `Crear`, `Actualizar`, `Asignar`, `Desasignar`, `Eliminar` ni similares.

### Handler (`internal/handler/metricas_handler_test.go`)

- `TestMetricasHandlerSprint` / `TestMetricasHandlerProyecto`: `200`, JSON
  snake_case con todas las claves y tipos; `null` en porcentajes/velocidad/ratio
  cuando corresponde; `base: "estado_actual"`; claves exactas (CA-21).
- `400 PARAMETRO_INVALIDO` con `{id}` o `{sprintId}` no numérico o `0`.
- `404 PROYECTO_NO_ENCONTRADO` y `404 SPRINT_NO_ENCONTRADA` traducidos desde el
  service.
- `500 ERROR_INTERNO` ante error de persistencia no controlado.
- Verificación de que el handler no escribe en el `ResponseWriter` más de una vez
  ni intenta bindear cuerpo.

### Router y guard (`cmd/api/router_test.go`)

- Las dos rutas están registradas después del middleware de autenticación y del
  guard de membresía.
- No miembro → `404 PROYECTO_NO_ENCONTRADO` en ambos endpoints.
- Miembro → la ruta llega al handler (200/404 según datos).
- Sin token → `401`.
- `:id` inválido → `400`.

### E2E por curl (manual, documentado en el PR)

1. Proyecto sin datos → `200` con ceros y `null` (proyecto y Sprint).
2. Proyecto con dos Sprints cerrados y uno activo → velocidad solo de cerrados.
3. Sprint activo con historias mixtas → SP y porcentaje correctos.
4. Sprint vacío → `0/0/null`.
5. No miembro en ambos endpoints → `404 PROYECTO_NO_ENCONTRADO`.
6. Sprint de otro proyecto → `404 SPRINT_NO_ENCONTRADA`.
7. Dos `GET` consecutivos y comparación de filas/timestamps antes y después →
   sin cambios (solo lectura).

## Resumen de escenarios BDD

Archivo: `docs/features/US-20-metricas.feature`. Se escribieron **12 escenarios
físicos (E-01 a E-12)**, dentro del rango 10-12 del DoD. E-05, E-06, E-08, E-09,
E-10 y E-12 usan `Esquema del escenario` con `Ejemplos` para cubrir varios casos
sin superar el límite de escenarios físicos.

| ID | Tipo | Escenario | CA |
| :--- | :--- | :--- | :--- |
| E-01 | Normal | Sprint con historias mixtas: SP planificados, completados, porcentaje e historias completadas | CA-01 |
| E-02 | Límite | Sprint sin historias: `0`, `0`, `null`, `0` | CA-02 |
| E-03 | Normal | Sprint con horas: totales disponibles y desviación con pares | CA-16, CA-07 |
| E-04 | Normal | Dos Sprints cerrados (8 y 4 SP) → velocidad `6` y porcentaje de historias completadas | CA-03, CA-22 |
| E-05 | Límite | Velocidad: sin cerrados → `null`; cerrado vacío/SP `0` → `0`; `Activo` excluido | CA-04, CA-05, CA-17 |
| E-06 | Normal/Límite | Desviación agregada `0.2`; historia sin horas excluida de la desviación; sin pares → `null` | CA-06, CA-07 |
| E-07 | Límite | Historias sin `story_points` cuentan `0` y no alteran `historias_completadas` | CA-08 |
| E-08 | Normal/Límite | Ratio `2/6` con un `rejected`; `rejected` suma a detectados; 0 completadas → `null` | CA-09, CA-10, CA-19 |
| E-09 | Límite | Proyecto sin historias → porcentaje `null`; con historias y ninguna `Hecho` → `0`; ratio `null` | CA-18, CA-22 |
| E-10 | Alternativo | Defecto detectado en Sprint 2 y `closed` en Sprint 3 | CA-11 |
| E-11 | Contrato/No funcional | Ambas respuestas incluyen `base: "estado_actual"` y claves exactas; consultas sin efectos laterales | CA-12, CA-15, CA-21 |
| E-12 | Error | No miembro → `404`; Sprint de otro proyecto/inexistente → `404`; parámetro inválido → `400` | CA-13, CA-14, CA-20 |

## Decisiones y supuestos

Todas las ambigüedades detectadas se resolvieron con una regla explícita; no
quedan supuestos bloqueantes.

- **D-01 — Defectos `rejected` en los conteos.** El issue solo excluye
  `rejected` del numerador del ratio. Por lo tanto, `defectos_detectados` (Sprint
  y proyecto) los incluye. Si el PO prefiere excluirlos también de los conteos,
  debe abrirse una decisión posterior: cambiaría CA-09/CA-19 y RN-12.
- **D-02 — Totales de horas independientes de la desviación.** Se siguen
  literalmente las fórmulas del issue: los totales suman valores disponibles y la
  desviación usa únicamente el subconjunto pareado.
- **D-03 — Velocidad con cerrados en cero.** El promedio divide por todos los
  Sprints `Cerrado`, incluidos los que aportan `0`, tal como exige el CA "sprint
  cerrado sin SP devuelve `0`" y la distinción con "sin sprints cerrados".
- **D-04 — Sin redondeo en la API.** Los porcentajes y ratios se emiten como
  `float64` completo. Los ejemplos con 2 decimales son solo ilustrativos.
- **D-05 — Contrato de proyecto mínimo.** No se agregan `historias_completadas`
  ni `historias_totales`; el issue no los pide y §7 solo exige el porcentaje y el
  ratio. El conteo interno se usa para calcular el ratio.
- **D-06 — `base` en el DTO.** `base: "estado_actual"` es una constante de
  presentación del handler; el dominio devuelve solo valores calculados.
- **D-07 — Sin error nuevo.** Se reutilizan `ErrProyectoNoEncontrado` y
  `ErrSprintNoEncontrado`; el código HTTP `SPRINT_NO_ENCONTRADA` ya existe.
- **D-08 — Consistencia entre consultas.** Las métricas se calculan con el
  estado disponible en cada `SELECT`; no se abre una transacción de lectura
  larga ni se garantiza una foto atómica entre tablas. Es consistente con
  `base: "estado_actual"` y no es un requisito del issue.

## Deuda documental corregida

El issue #101 incluye como tarea corregir en la spec de US-18 la referencia a la
numeración anterior. Se actualizó `docs/specs/US-18-defectos-por-sprint-spec.md`
para que diga métricas (US-20) y reporte de sprint (US-22).

## Definición de terminado (heredada del issue #101)

- [ ] `go test ./...` en verde.
- [ ] Esta spec SDD escrita antes del código.
- [ ] `docs/features/US-20-metricas.feature` con 10 a 12 escenarios en español,
  escrito antes del código.
- [ ] `SprintBacklogRepository.ListarPorProyecto` con test de una sola consulta
  para la velocidad.
- [ ] Tests de fórmulas con casos de cero, un elemento, valores iguales y valores
  mixtos.
- [ ] E2E por curl: proyecto sin datos, Sprints cerrados, Sprint activo, Sprint
  vacío, no miembro y Sprint de otro proyecto.
- [x] Cross-referencias a US-02 a US-06, US-09, US-18, US-19 y US-F10.
- [x] Corrección de numeración en la spec de US-18 (fuera de este archivo).
