# US-21 — Extensiones de métricas (flujo y defectos)

## Objetivo

Agregar, en solo lectura, métricas opcionales de flujo y de calidad al endpoint
de métricas de proyecto de US-20, para estimar cuántos Sprints faltan, cuánto
trabajo está en curso y cuánto tarda una historia en completarse.

Las métricas se calculan al momento de la consulta a partir del estado actual de
historias, Sprints, asignaciones al Sprint Backlog y defectos. No se persisten
resultados, no se crean tablas y no se modifica ningún dato.

**Trazabilidad:** Issue #102 → Esta especificación → Criterios de aceptación
(CA-01 a CA-15) → Escenarios BDD
(`docs/features/US-21-extensiones-metricas.feature`) → Tests Go (TDD) → Código.

| Eslabón | Artefacto |
| :--- | :--- |
| Issue | [SebastianAguilera/ingenieria-y-calidad#102](https://github.com/SebastianAguilera/ingenieria-y-calidad/issues/102) |
| SDD | `docs/specs/US-21-extensiones-metricas-spec.md` (este archivo) |
| Criterios de aceptación | CA-01 a CA-15 de este documento |
| BDD | `docs/features/US-21-extensiones-metricas.feature` (E-01 a E-11) |
| Tests | `internal/domain/metricas_extensiones_test.go`, `internal/service/metricas_service_test.go`, `internal/handler/metricas_handler_test.go` |
| Código | `internal/domain/metricas.go`, `internal/handler/dto_metricas.go` |

## Dependencias y relación con otras historias

| Historia | Relación |
| :--- | :--- |
| US-20 | Base bloqueante: aporta `GET /api/proyectos/{id}/metricas`, la velocidad, las historias y los defectos ya calculados. |
| US-02 | Aporta `Historia` y su `Estado` (`Backlog`, `En Progreso`, `Hecho`). |
| US-03 | Aporta `Sprint` y sus estados `Planificado`, `Activo` y `Cerrado`. |
| US-04 | Aporta las asignaciones historia↔Sprint necesarias para la tasa de entrega. |
| US-05 | Aporta `story_points` (`float64`). |
| US-09 | Aporta `Defecto`, `EstadoDefecto` y `SeveridadDefecto`. |
| US-11 / US-F10 | Guards compartidos; cualquier miembro consulta y un no miembro recibe `404 PROYECTO_NO_ENCONTRADO`. |

**Requiere:** US-20 (mergeada).

**No bloquea** a ninguna historia de backend. US-F21 no muestra estas métricas;
las extensiones de US-21 se muestran en US-F22.

## Alcance

**Incluye**

- Campos nuevos en `GET /api/proyectos/{id}/metricas` (sin endpoints nuevos).
- `sprints_restantes`, `wip`, `tasa_entrega`, `lead_time_estimado`,
  `lead_time_aproximado`, `defectos_abiertos` y `defectos_por_severidad`.
- Funciones puras de cálculo en `internal/domain`, con tests de tabla.
- Extensión del DTO snake_case del handler de proyecto.

**No incluye**

- Endpoints nuevos ni cambios en `GET .../sprints/{sprintId}/metricas`.
- Coste en días-persona, desglose por Sprint de estas métricas, diagramas de
  flujo acumulado, scatterplot, histograma ni Monte Carlo.
- Persistencia, snapshots, cacheo, tablas nuevas o migraciones.
- Frontend (US-F22).
- Filtros, paginación o parámetros de consulta.

## Entradas y precondiciones

### Precondiciones

- Esquema existente de US-01 a US-06, US-09 y US-20.
- Request autenticado con JWT válido (US-10).
- El usuario es miembro del proyecto (US-F10); cualquier rol alcanza.

### Endpoint y parámetros

| Endpoint | Método | Cuerpo | Query |
| :--- | :---: | :---: | :---: |
| `/api/proyectos/{id}/metricas` | `GET` | No | No |

| Parámetro | Ubicación | Obligatorio | Tipo / formato | Límites |
| :--- | :--- | :---: | :--- | :--- |
| `id` | path | Sí | entero positivo (`uint`) | `> 0`; no numérico o `0` → `400 PARAMETRO_INVALIDO` |

```http
GET /api/proyectos/1/metricas
Authorization: Bearer <jwt>
```

Sin cuerpo y sin parámetros de consulta. Los query params desconocidos se
ignoran.

## Salidas esperadas y contrato JSON

El endpoint responde `200 OK` con el objeto de US-20 **más** las claves nuevas,
en `snake_case`. Los campos de US-20 no cambian de nombre, tipo ni semántica.
Los porcentajes y ratios siguen en `0..1` y no se redondean en la API.

### Claves nuevas de `GET .../metricas` → `200`

| Clave | Tipo JSON | Nulo | Descripción |
| :--- | :--- | :---: | :--- |
| `sprints_restantes` | número | Sí | `ceil(SP pendientes / velocidad)`; `null` si `velocidad` es `null` o `0`. |
| `wip` | entero | No | Cantidad de historias del proyecto en estado `En Progreso`. |
| `tasa_entrega` | número | Sí | Promedio de historias `Hecho` asignadas a cada Sprint `Cerrado`; `null` sin Sprints cerrados. |
| `lead_time_estimado` | número | Sí | `wip / tasa_entrega`; `null` si `tasa_entrega` es `null` o `0`. |
| `lead_time_aproximado` | booleano | No | `true` cuando hay `lead_time_estimado` (Ley de Little); `false` cuando es `null`. |
| `defectos_abiertos` | entero | No | Defectos del proyecto en estado `open`. |
| `defectos_por_severidad` | objeto | No | Conteo de defectos por `minor`, `major`, `critical` y `blocker`; `0` donde no haya. |

`defectos_por_severidad` siempre incluye las cuatro claves:

| Clave | Tipo JSON | Nulo | Descripción |
| :--- | :--- | :---: | :--- |
| `minor` | entero | No | Defectos de severidad `minor`. |
| `major` | entero | No | Defectos de severidad `major`. |
| `critical` | entero | No | Defectos de severidad `critical`. |
| `blocker` | entero | No | Defectos de severidad `blocker`. |

### Ejemplo con datos

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
  "ratio_defectos": 0.3333333333333333,
  "sprints_restantes": 7,
  "wip": 6,
  "tasa_entrega": 2,
  "lead_time_estimado": 3,
  "lead_time_aproximado": true,
  "defectos_abiertos": 1,
  "defectos_por_severidad": {
    "minor": 1,
    "major": 1,
    "critical": 1,
    "blocker": 0
  }
}
```

### Caso sin datos (no es error)

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
  "ratio_defectos": null,
  "sprints_restantes": null,
  "wip": 0,
  "tasa_entrega": null,
  "lead_time_estimado": null,
  "lead_time_aproximado": false,
  "defectos_abiertos": 0,
  "defectos_por_severidad": {
    "minor": 0,
    "major": 0,
    "critical": 0,
    "blocker": 0
  }
}
```

### Cuerpo estándar de error

Sin cambios respecto de US-20:

```json
{
  "error": "CODIGO",
  "mensaje": "Descripción en español",
  "campo": "campo opcional"
}
```

## Fórmulas y modelo de datos

### Notación

- `Historias(P)`: todas las historias del proyecto.
- `SP(h)`: `h.story_points`; si es `null`, vale `0`.
- `Hecho(h)`: `h.estado == "Hecho"`.
- `EnProgreso(h)`: `h.estado == "En Progreso"`.
- `Pendientes(P)`: historias con `¬Hecho(h)`, tengan o no Sprint asignado.
- `Cerrados(P)`: Sprints del proyecto con `estado == "Cerrado"`.
- `H(s)`: historias asignadas al Sprint `s` mediante `sprint_backlog_items`.
- `Hecho(s)`: `{ h ∈ H(s) : Hecho(h) }`.
- `Defectos(P)`: todos los defectos del proyecto.

### Fórmulas

```text
SP_pend(P)        = Σ_{h ∈ Pendientes(P)} SP(h)
sprints_rest(P)   = vel(P) == null ∨ vel(P) == 0
                      ? null
                      : ceil( SP_pend(P) / vel(P) )
wip(P)            = |{ h ∈ Historias(P) : EnProgreso(h) }|
tasa(P)           = Cerrados(P) == ∅
                      ? null
                      : (Σ_{s ∈ Cerrados(P)} |Hecho(s)|) / |Cerrados(P)|
lead_time(P)      = tasa(P) == null ∨ tasa(P) == 0
                      ? null
                      : wip(P) / tasa(P)
lead_aprox(P)     = lead_time(P) != null
def_abiertos(P)   = |{ d ∈ Defectos(P) : d.estado == "open" }|
sev_minor(P)      = |{ d ∈ Defectos(P) : d.severidad == "minor" }|
sev_major(P)      = |{ d ∈ Defectos(P) : d.severidad == "major" }|
sev_critical(P)   = |{ d ∈ Defectos(P) : d.severidad == "critical" }|
sev_blocker(P)    = |{ d ∈ Defectos(P) : d.severidad == "blocker" }|
```

`vel(P)` es la velocidad de US-20: promedio de SP completados de los Sprints
`Cerrado`, o `null` sin cerrados.

### Estructuras de dominio (referencia)

```go
type DefectosPorSeveridad struct {
    Minor    int
    Major    int
    Critical int
    Blocker  int
}

type MetricasProyecto struct {
    // Campos de US-20 (sin cambios)
    Velocidad                      *float64
    PorcentajeHistoriasCompletadas *float64
    HorasEstimadas                 float64
    HorasReales                    float64
    DesviacionEsfuerzo             *float64
    DefectosDetectados             int
    DefectosResueltos              int
    RatioDefectos                  *float64
    // Extensiones de US-21
    SprintsRestantes     *float64
    WIP                  int
    TasaEntrega          *float64
    LeadTimeEstimado     *float64
    LeadTimeAproximado   bool
    DefectosAbiertos     int
    DefectosPorSeveridad DefectosPorSeveridad
}
```

Funciones puras nuevas (o equivalentes):

```go
func CalcularSPPendientes(historias []Historia) float64
func CalcularSprintsRestantes(historias []Historia, velocidad *float64) *float64
func ContarHistoriasEnProgreso(historias []Historia) int
func CalcularTasaEntrega(sprints []Sprint, asignaciones []SprintBacklogItem,
    historias []Historia) *float64
func CalcularLeadTimeEstimado(wip int, tasaEntrega *float64) *float64
func ContarDefectosAbiertos(defectos []Defecto) int
func CalcularDefectosPorSeveridad(defectos []Defecto) DefectosPorSeveridad
```

`CalcularMetricasProyecto` incorpora las nuevas claves reutilizando
`CalcularVelocidad`.

### Campos existentes utilizados (verificados en el código)

| Struct | Campos |
| :--- | :--- |
| `domain.Historia` | `ID`, `Estado`, `StoryPoints *float64` |
| `domain.Sprint` | `ID`, `Estado` |
| `domain.SprintBacklogItem` | `SprintID`, `HistoriaID` |
| `domain.Defecto` | `Estado`, `Severidad` |

No se requieren campos nuevos ni cambios de esquema.

### Consultas

No se agregan consultas: se reutilizan las de US-20
(`sprints.ListarPorProyecto`, `sprintBacklog.ListarPorProyecto`,
`historias.ListarPorProyecto`, `defectos.ListarPorProyecto`). El número de
consultas sigue siendo fijo (sin N+1).

## Reglas de negocio

| ID | Regla |
| :--- | :--- |
| RN-01 | Los campos nuevos son de solo lectura: no crean, actualizan ni eliminan filas; no ejecutan transacciones de escritura. |
| RN-02 | Cualquier miembro del proyecto puede consultar; un no miembro recibe `404 PROYECTO_NO_ENCONTRADO`. |
| RN-03 | `SP_pend` suma `story_points` de todas las historias no `Hecho`, tengan o no Sprint asignado. Una historia sin `story_points` aporta `0`. |
| RN-04 | `sprints_restantes` usa `ceil` (redondeo hacia arriba) y reutiliza la velocidad de US-20. Si la velocidad es `null` o `0`, es `null`. |
| RN-05 | Si no hay SP pendientes y la velocidad es mayor a `0`, `sprints_restantes` es `0`, no `null`. |
| RN-06 | `wip` cuenta historias en estado `En Progreso`. Incluye las que no tienen Sprint asignado. |
| RN-07 | `tasa_entrega` promedia las historias `Hecho` **asignadas** a cada Sprint `Cerrado`. Las historias `Hecho` no asignadas a ningún Sprint no cuentan. |
| RN-08 | El divisor de `tasa_entrega` es la cantidad de Sprints `Cerrado`, incluidos los que aportan `0`. Los Sprints `Planificado` y `Activo` no entran. |
| RN-09 | Sin Sprints cerrados, `tasa_entrega` es `null`. Con cerrados y ningún `Hecho`, es `0` (no `null`). |
| RN-10 | `lead_time_estimado` = `wip / tasa_entrega`. Si la tasa es `null` o `0`, es `null`. Es una aproximación (Ley de Little), no una medición. |
| RN-11 | `lead_time_aproximado` es `true` cuando `lead_time_estimado` no es `null`; `false` cuando es `null`. |
| RN-12 | `defectos_abiertos` cuenta **solo** los defectos en estado `open`. Los estados `in_progress`, `resolved`, `closed` y `rejected` no cuentan. |
| RN-13 | `defectos_por_severidad` cuenta todos los defectos del proyecto, en cualquier estado (incluye `rejected`), agrupados por severidad; coherencia con `defectos_detectados` de US-20. |
| RN-14 | `defectos_por_severidad` siempre emite las cuatro claves; las severidades sin defectos valen `0`. |
| RN-15 | Los campos de US-20 no cambian; las claves nuevas se agregan sin eliminarlas ni renombrarlas. `base` sigue siendo `"estado_actual"` y no se redondea en la API. |

## Restricciones

### Técnicas

- Solo sentencias `SELECT`; sin escrituras, bloqueos ni transacciones de escritura.
- No se crean tablas, columnas, índices ni migraciones.
- No se agregan endpoints, errores centinela, dependencias externas ni parámetros.
- Se reutilizan `parsearID`, `traducirError` y el guard `ExigirAccesoAProyecto`.
- El número de consultas es fijo y no depende de la cantidad de Sprints,
  historias o defectos.
- Las respuestas no se cachean; cada request recalcula.

### De arquitectura por capas

- `domain` no importa Gin, GORM ni `context`; las fórmulas son puras.
- `service` no cambia: ya carga todos los datos necesarios.
- `handler` solo traduce HTTP y DTO; no calcula métricas.

### De negocio y contrato

- Solo lectura, sin snapshots ni histórico (`base: "estado_actual"`).
- Sin redondeo en la API (el `ceil` de `sprints_restantes` es parte de la regla,
  no un redondeo de presentación).
- `null` significa "sin datos" y nunca se emite como `0` para las métricas que
  pueden no tener base.

## Casos límite

| ID | Caso | Resultado |
| :--- | :--- | :--- |
| CL-01 | Proyecto sin historias | `wip 0`, `sprints_restantes null` (velocidad `null`), `tasa_entrega null`, `lead_time null`, `aproximado false`. |
| CL-02 | Velocidad `0` | `sprints_restantes null` aunque haya SP pendientes. |
| CL-03 | SP pendientes `0` y velocidad `> 0` | `sprints_restantes 0`. |
| CL-04 | SP pendientes fraccionarios (`0.5`) | Se dividen como `float64` y se aplica `ceil` sobre el cociente. |
| CL-05 | Sin Sprints cerrados | `tasa_entrega null` y `lead_time null`, aunque haya historias `Hecho`. |
| CL-06 | Sprint cerrado sin historias `Hecho` | Aporta `0` al promedio; si es el único cerrado, `tasa_entrega 0`. |
| CL-07 | Historia `Hecho` sin Sprint asignado | No suma a `tasa_entrega`. |
| CL-08 | `wip 0` con tasa `> 0` | `lead_time_estimado 0`, `lead_time_aproximado true`. |
| CL-09 | Tasa `0` | `lead_time null`, `aproximado false`. |
| CL-10 | Defecto `in_progress` | No suma a `defectos_abiertos`; sí suma a su severidad. |
| CL-11 | Defecto `rejected` | No suma a `defectos_abiertos`; sí suma a su severidad. |
| CL-12 | Ningún defecto de una severidad | Esa clave vale `0`. |
| CL-13 | Sin defectos | `defectos_abiertos 0` y las cuatro severidades en `0`. |
| CL-14 | No miembro | `404 PROYECTO_NO_ENCONTRADO`. |
| CL-15 | `{id}` no numérico o `0` | `400 PARAMETRO_INVALIDO`. |

## Criterios de aceptación

- [ ] **CA-01** Given 40 SP pendientes y velocidad 6, When se consulta, Then
  `sprints_restantes` es `7` (redondeo hacia arriba).
- [ ] **CA-02** Given velocidad `null` o `0`, When se consulta, Then
  `sprints_restantes` es `null`.
- [ ] **CA-03** Given 6 historias en `En Progreso`, When se consulta, Then `wip`
  es `6`.
- [ ] **CA-04** Given dos Sprints cerrados con 3 y 1 historias `Hecho` asignadas
  respectivamente, When se consulta, Then `tasa_entrega` es `2`.
- [ ] **CA-05** Given WiP 6 y tasa de entrega 2, When se consulta, Then
  `lead_time_estimado` es `3` y `lead_time_aproximado` es `true`.
- [ ] **CA-06** Given tasa de entrega `null` o `0`, When se consulta, Then
  `lead_time_estimado` es `null` y `lead_time_aproximado` es `false`.
- [ ] **CA-07** Given defectos `minor` y `critical` y ninguno `blocker`, When se
  consulta, Then el desglose muestra `blocker: 0`.
- [ ] **CA-08** Given defectos en `open` y en `resolved`, When se consulta,
  Then `defectos_abiertos` cuenta solo los `open`.
- [ ] **CA-09** Given un proyecto sin datos, When se consulta, Then las métricas
  sin base son `null`, los conteos `0`, `lead_time_aproximado` `false` y las
  severidades `0`.
- [ ] **CA-10** Given un Sprint `Activo` con historias `Hecho`, When se calcula
  `tasa_entrega`, Then ese Sprint no entra en el promedio.
- [ ] **CA-11** Given un Sprint cerrado sin historias, When se calcula la tasa,
  Then aporta `0` al promedio y cuenta en el divisor.
- [ ] **CA-12** Given cualquier consulta, When se completa, Then no se modifica
  ningún dato.
- [ ] **CA-13** Given una respuesta exitosa, Then conserva todas las claves de
  US-20 e incluye exactamente las claves nuevas del contrato, con sus tipos.
- [ ] **CA-14** Given un no miembro, When consulta, Then
  `404 PROYECTO_NO_ENCONTRADO`.
- [ ] **CA-15** Given un `{id}` no numérico o `0`, When consulta, Then
  `400 PARAMETRO_INVALIDO`.

## Condiciones de error y mapeo de errores HTTP

| Situación | Error de dominio | Estado | Código API | Campo |
| :--- | :--- | :---: | :--- | :--- |
| `{id}` no numérico o `0` | — (handler) | 400 | `PARAMETRO_INVALIDO` | — |
| Sin token o token inválido | `ErrTokenRequerido` / `ErrTokenInvalido` | 401 | `TOKEN_REQUERIDO` / `TOKEN_INVALIDO` | — |
| Proyecto inexistente o usuario no miembro | `ErrProyectoNoEncontrado` | 404 | `PROYECTO_NO_ENCONTRADO` | — |
| Error de persistencia al leer | `ErrErrorDePersistencia` | 500 | `ERROR_INTERNO` | — |

No se agregan errores de dominio nuevos.

## Arquitectura (Clean Architecture)

- **domain**: `metricas.go` suma `DefectosPorSeveridad`, los campos nuevos de
  `MetricasProyecto` y las funciones puras de cálculo. Sin contexto, Gin ni GORM.
- **service**: sin cambios. `CalcularPorProyecto` ya carga Sprints,
  asignaciones, historias y defectos, y delega en el dominio.
- **handler**: `dto_metricas.go` agrega las claves nuevas a
  `respuestaMetricasProyecto` y las mapea desde el dominio; el objeto
  `defectos_por_severidad` se emite siempre con las cuatro severidades.
- **cmd/api**: sin cambios; no hay rutas nuevas.

## Plan de pruebas (TDD)

Ciclo **RED → GREEN → REFACTOR**. Cada ID de escenario BDD se reutiliza como
nombre de subtest.

### Dominio (`internal/domain/metricas_extensiones_test.go`)

- `TestCalcularSPPendientes`: sin historias, todas `Hecho`, estados mixtos,
  `story_points` `nil`.
- `TestCalcularSprintsRestantes`: 40/6 → 7; velocidad `nil` → `null`; velocidad
  `0` → `null`; pendientes `0` → `0`; cociente exacto; `ceil` con fracciones.
- `TestContarHistoriasEnProgreso`: sin historias, una, varias mixtas.
- `TestCalcularTasaEntrega`: sin cerrados → `null`; 3 y 1 → 2; cerrado sin
  historias → `0`; activo excluido; historia no `Hecho` no cuenta; historia
  `Hecho` no asignada no cuenta.
- `TestCalcularLeadTimeEstimado`: 6/2 → 3; tasa `nil` → `null`; tasa `0` →
  `null`; WiP `0` → `0`.
- `TestContarDefectosAbiertos`: solo `open`; excluye `in_progress`, `resolved`,
  `closed` y `rejected`.
- `TestCalcularDefectosPorSeveridad`: conteo por severidad, claves en `0`,
  severidades inválidas ignoradas, incluye todos los estados.
- `TestCalcularMetricasProyectoExtensiones`: caso completo y caso sin datos que
  verifiquen todas las claves nuevas, incluido `LeadTimeAproximado`.

### Service (`internal/service/metricas_service_test.go`)

- `TestMetricasServiceCalcularPorProyectoExtensiones`: con mocks de US-20,
  verifica que las claves nuevas llegan calculadas desde el dominio y que no se
  agregan consultas.

### Handler (`internal/handler/metricas_handler_test.go`)

- `TestMetricasHandlerProyectoExtensiones`: `200`, claves nuevas con sus valores
  y tipos, `defectos_por_severidad` anidado con las cuatro claves.
- `TestMetricasHandlerProyectoExtensionesSinDatos`: `null`, `0` y `false` donde
  corresponde.
- `TestMetricasHandlerProyecto`: actualiza la lista exacta de claves para
  incluir las nuevas (CA-13).

### E2E por curl (manual, documentado en el PR)

1. Proyecto sin datos → `200` con `wip 0`, `sprints_restantes null`,
   `tasa_entrega null`, `lead_time null`, `aproximado false`, severidades `0`.
2. Proyecto con Sprints cerrados y SP pendientes → `sprints_restantes` y
   `tasa_entrega` coherentes.
3. Proyecto con historias `En Progreso` → `wip` correcto y `lead_time` cuando la
   tasa sea mayor a `0`.
4. Defectos en distintos estados y severidades → `defectos_abiertos` solo `open`
   y desglose por severidad correcto.
5. Dos `GET` consecutivos → mismos valores, sin cambios en los datos.

## Resumen de escenarios BDD

Archivo: `docs/features/US-21-extensiones-metricas.feature`. Se escriben **11
escenarios físicos (E-01 a E-11)** con casos normales, límite, alternativos y de
error.

| ID | Tipo | Escenario | CA |
| :--- | :--- | :--- | :--- |
| E-01 | Normal | 40 SP pendientes y velocidad 6 → `sprints_restantes 7` | CA-01 |
| E-02 | Límite | Velocidad `null`/`0` → `sprints_restantes null`; pendientes `0` → `0` | CA-02, CA-03 |
| E-03 | Normal | 6 historias `En Progreso` → `wip 6` | CA-03 |
| E-04 | Normal | Dos cerrados con 3 y 1 `Hecho` → `tasa_entrega 2` | CA-04 |
| E-05 | Límite | Sin cerrados → `null`; cerrado vacío cuenta `0`; activo excluido | CA-05, CA-10, CA-11 |
| E-06 | Normal | WiP 6 y tasa 2 → `lead_time 3` y `aproximado true` | CA-05 |
| E-07 | Límite | Tasa `null`/`0` → `lead_time null` y `aproximado false` | CA-06 |
| E-08 | Normal | Defectos `open` y `resolved` → `defectos_abiertos` solo `open` | CA-08 |
| E-09 | Límite | Desglose por severidad con `blocker: 0` | CA-07, CA-09 |
| E-10 | Contrato/No funcional | Claves nuevas exactas, `base` y consulta sin efectos laterales | CA-12, CA-13 |
| E-11 | Error | No miembro → `404`; `{id}` inválido → `400` | CA-14, CA-15 |

## Decisiones y supuestos

- **D-01 — `lead_time_aproximado` según exista valor.** Se emite `true` cuando
  `lead_time_estimado` no es `null` y `false` cuando es `null`. Así el booleano
  indica si la aproximación está disponible, coherente con "`null` significa sin
  datos".
- **D-02 — Defectos abiertos solo `open`.** El criterio de aceptación del issue
  es explícito ("los abiertos cuentan solo los `open`"), por lo que
  `in_progress` no cuenta.
- **D-03 — Desglose por severidad sobre todos los defectos.** El issue no
  restringe el estado del desglose; se cuentan todos los defectos del proyecto
  (incluidos `rejected`), igual que `defectos_detectados` de US-20. El estado se
  refleja en `defectos_abiertos`.
- **D-04 — `sprints_restantes` usa la velocidad de US-20.** No se recalcula: se
  reutiliza el valor ya devuelto por el endpoint, garantizando consistencia.
- **D-05 — Sin redondeo de presentación.** El único redondeo es el `ceil` de
  `sprints_restantes`, exigido por la regla; el resto se emite como `float64`
  completo.
- **D-06 — Contrato aditivo.** No se renombra ni elimina ninguna clave de US-20;
  los tests existentes se actualizan solo para incluir las claves nuevas.

## Definición de terminado (heredada del issue #102)

- [ ] `go test ./...` en verde.
- [ ] Esta spec SDD escrita antes del código.
- [ ] `docs/features/US-21-extensiones-metricas.feature` escrito antes del
  código.
- [ ] Tests de cada fórmula con casos sin datos, cero y valores mixtos.
- [ ] E2E por curl del endpoint de métricas con los campos nuevos.
- [ ] Cross-referencias a US-20 y a las historias de datos utilizadas.
