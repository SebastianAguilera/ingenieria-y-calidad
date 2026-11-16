# US-23 — Valor de negocio y priorización del backlog (ROI y WSJF)

## Objetivo

Permitir que el equipo cargue tres valores opcionales por historia
(`valor_negocio`, `criticidad_tiempo` y `reduccion_riesgo`) y consultar el
backlog no completado ordenado por **WSJF**, mostrando también el **ROI** de cada
historia. Los cálculos son de solo lectura, no se persisten y no reemplazan la
prioridad manual (`Alta`, `Media`, `Baja`) de US-02.

**Trazabilidad:** Issue [#110](https://github.com/SebastianAguilera/ingenieria-y-calidad/issues/110)
→ Esta especificación (SDD) → Criterios de aceptación (CA-01 a CA-12) →
Escenarios BDD (`docs/features/US-23-priorizacion-backlog.feature`) → Tests Go
(TDD) → Código.

| Eslabón | Artefacto |
| :--- | :--- |
| Issue | [SebastianAguilera/ingenieria-y-calidad#110](https://github.com/SebastianAguilera/ingenieria-y-calidad/issues/110) |
| SDD | `docs/specs/US-23-priorizacion-backlog-spec.md` (este archivo) |
| Criterios de aceptación | CA-01 a CA-12 de este documento |
| BDD | `docs/features/US-23-priorizacion-backlog.feature` (E-01 a E-10) |
| Tests | `internal/domain/priorizacion_test.go`, `internal/service/historia_priorizacion_test.go`, `internal/handler/historia_handler_test.go`, `internal/handler/historia_priorizacion_test.go`, `internal/repository/historia_priorizacion_test.go`, `cmd/api/router_test.go` |
| Código | `internal/domain/priorizacion.go`, `internal/domain/historia.go`, `internal/domain/errores.go`, `internal/repository/models/historia_model.go`, `internal/repository/models/mappers.go`, `internal/repository/historia_repository.go`, `internal/service/historia_service.go`, `internal/handler/dto_historia.go`, `internal/handler/historia_handler.go`, `internal/handler/errores.go` |

## Dependencias y relación con otras historias

| Historia | Relación |
| :--- | :--- |
| US-02 | Aporta `Historia`, su alta, PATCH y listado. La prioridad manual `Alta`/`Media`/`Baja` sigue vigente y es independiente de WSJF/ROI. |
| US-05 | Aporta `story_points`, divisor de ROI y WSJF, con escala `0, 0.5, 1, 2, 3, 5, 8, 13, 21`. |
| US-19 | Aporta `horas_estimadas`; no interviene en las fórmulas de esta historia. |
| US-20 | Aporta métricas de avance. ROI y WSJF **no** son métricas de US-20 ni entran en el dashboard. |
| US-11 / US-F10 | Guard de membresía: cualquier miembro consulta y edita; un no miembro recibe `404 PROYECTO_NO_ENCONTRADO`. |
| US-F24 (futura) | Consumirá los campos y el endpoint priorizado en la interfaz. **No se implementa aquí.** |

**Dependencias cumplidas:** US-02, US-05, US-19, US-20 y US-F10 ya están
implementadas. **Bloquea a:** US-F24.

## Alcance

**Incluye**

- Tres campos opcionales en `Historia`, con escala `1, 2, 3, 5, 8, 13, 21`:
  - `valor_negocio`
  - `criticidad_tiempo`
  - `reduccion_riesgo`
- Persistencia con `AutoMigrate`, columnas nullable y sin valor por defecto.
- Alta (`POST`) y PATCH de historia aceptan los tres campos.
- Validación de la escala en alta y PATCH con `422 VALIDACION` y `campo`.
- Funciones puras de dominio para ROI, costo de demora y WSJF.
- `GET /api/proyectos/{id}/backlog/priorizado` de solo lectura.
- Devolución de los tres valores en alta, PATCH, listado y Sprint Backlog, que
  reutilizan la representación de historia.
- Tests TDD de fórmulas, validación, service, handler, repositorio, migración y
  guard de membresía.

**No incluye**

- Reordenar el backlog persistido ni guardar el orden calculado.
- Pantallas de frontend (US-F24).
- Monte Carlo, historial de cambios de valor, snapshots ni caché.
- Reemplazar o modificar la prioridad manual de US-02.
- Cambios en `tablas.sql`.
- Redondeo de resultados en la API.

## Entradas y precondiciones

### Precondiciones

- Esquema de historias de US-02, US-05 y US-19 existente (`AutoMigrate`).
- Request autenticado con JWT válido (US-10).
- El usuario es miembro del proyecto (US-F10); cualquier rol alcanza.

### Endpoint

| Endpoint | Método | Cuerpo | Query |
| :--- | :---: | :---: | :---: |
| `/api/proyectos/{id}/backlog/priorizado` | `GET` | No | No |

| Parámetro | Ubicación | Obligatorio | Tipo / formato | Límites |
| :--- | :--- | :---: | :--- | :--- |
| `id` | path | Sí | entero positivo (`uint`) | `> 0`; no numérico o `0` → `400 PARAMETRO_INVALIDO` |

```http
GET /api/proyectos/1/backlog/priorizado
Authorization: Bearer <jwt>
```

Sin cuerpo, sin filtros y sin paginación. Los query params desconocidos se
ignoran.

## Salidas esperadas y contrato JSON

`GET /api/proyectos/{id}/backlog/priorizado` responde `200 OK` con:

| Clave | Tipo JSON | Nulo | Descripción |
| :--- | :--- | :---: | :--- |
| `total` | entero | No | Cantidad de historias devueltas. |
| `historias` | array | No | Historias no `Hecho` ordenadas por WSJF. |
| `historias[].id` | entero | No | Identificador. |
| `historias[].codigo` | string | No | Código `US-NN`. |
| `historias[].titulo` | string | No | Título. |
| `historias[].descripcion` | string | No | Descripción. |
| `historias[].prioridad` | string | No | Prioridad manual de US-02. |
| `historias[].estado` | string | No | Estado de la historia. |
| `historias[].story_points` | número | Sí | Tamaño relativo. |
| `historias[].esfuerzo_real` | número | Sí | Horas reales. |
| `historias[].horas_estimadas` | número | Sí | Horas estimadas. |
| `historias[].valor_negocio` | número | Sí | Valor para el negocio o cliente. |
| `historias[].criticidad_tiempo` | número | Sí | Necesidad de entrega en el tiempo. |
| `historias[].reduccion_riesgo` | número | Sí | Reducción de riesgos u oportunidad. |
| `historias[].roi` | número | Sí | `valor_negocio / story_points`. |
| `historias[].wsjf` | número | Sí | `costo_demora / story_points`. |
| `historias[].creado_en` | string | No | Timestamp ISO-8601. |
| `historias[].actualizado_en` | string | No | Timestamp ISO-8601. |

```json
{
  "total": 2,
  "historias": [
    {
      "id": 5,
      "codigo": "US-05",
      "titulo": "Historia priorizada",
      "descripcion": "Con valores cargados",
      "prioridad": "Alta",
      "estado": "Backlog",
      "story_points": 5,
      "esfuerzo_real": null,
      "horas_estimadas": null,
      "valor_negocio": 8,
      "criticidad_tiempo": 3,
      "reduccion_riesgo": 2,
      "roi": 1.6,
      "wsjf": 2.6,
      "creado_en": "2026-11-15T10:00:00Z",
      "actualizado_en": "2026-11-15T10:00:00Z"
    }
  ]
}
```

Sin datos, la respuesta es válida y no un error:

```json
{
  "total": 0,
  "historias": []
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

`campo` se omite cuando está vacío.

## Fórmulas y modelo

### Notación

- `VN(h)`: `h.valor_negocio`; `null` si no está cargado.
- `CT(h)`: `h.criticidad_tiempo`; `null` si no está cargado.
- `RR(h)`: `h.reduccion_riesgo`; `null` si no está cargado.
- `SP(h)`: `h.story_points`; `null` si no está cargado.
- `vacio(x)`: `true` si `x == null`.

### Cálculos

```text
ROI(h)      = (SP(h) == null ∨ SP(h) == 0 ∨ VN(h) == null)
                ? null
                : VN(h) / SP(h)

costo(h)    = (vacio(VN(h)) ? 0 : VN(h))
            + (vacio(CT(h)) ? 0 : CT(h))
            + (vacio(RR(h)) ? 0 : RR(h))

WSJF(h)     = (SP(h) == null ∨ SP(h) == 0)
                ? null
                : costo(h) / SP(h)
```

`costo(h)` siempre es un número: una historia sin valores de negocio tiene costo
`0`. El ROI, en cambio, requiere `valor_negocio` cargado y un `story_points`
distinto de `null` y de `0`.

### Orden

1. WSJF descendente.
2. Las historias con `WSJF == null` van al final.
3. Desempate por `ID` ascendente, tanto entre WSJF iguales como entre nulos.

### Historias consideradas

Todas las historias del proyecto con `estado != "Hecho"`. No importa si están o
no asignadas a un Sprint.

### Estructuras de dominio de referencia

```go
type HistoriaPriorizada struct {
    Historia Historia
    ROI      *float64
    WSJF     *float64
}

func ValorPriorizacionValido(valor float64) bool
func CalcularROI(valorNegocio, storyPoints *float64) *float64
func CalcularCostoDemora(valorNegocio, criticidadTiempo, reduccionRiesgo *float64) float64
func CalcularWSJF(costoDemora float64, storyPoints *float64) *float64
func PriorizarBacklog(historias []Historia) []HistoriaPriorizada
```

Los nombres son indicativos; el requisito es que sean funciones puras en
`internal/domain`, sin Gin, GORM ni `context`, y con tests de tabla.

## Reglas de negocio

| ID | Regla |
| :--- | :--- |
| RN-01 | Los tres campos son opcionales y nullable. Ausentes o `null` significan "sin cargar". |
| RN-02 | Si se informan, deben pertenecer a la escala `1, 2, 3, 5, 8, 13, 21`. Cualquier otro valor, incluido `0`, `0.5`, decimales, negativos o mayores a `21`, produce `422 VALIDACION`. |
| RN-03 | La escala de los valores de negocio es independiente de la escala de `story_points`: `0` y `0.5` no son válidos como valor de negocio. |
| RN-04 | En alta, un valor inválido impide la creación: no se inserta ninguna fila. |
| RN-05 | En PATCH, un valor inválido impide la modificación: no se actualiza ningún campo de la historia. |
| RN-06 | Un PATCH con `null` o con el campo ausente significa "no informado". No borra un valor ya cargado. |
| RN-07 | `ROI = valor_negocio / story_points`. Es `null` si falta `valor_negocio`, si no hay `story_points` o si `story_points == 0`. |
| RN-08 | `costo_demora = valor_negocio + criticidad_tiempo + reduccion_riesgo`; cada campo ausente cuenta como `0`. |
| RN-09 | `WSJF = costo_demora / story_points`. Es `null` si no hay `story_points` o si `story_points == 0`. |
| RN-10 | El endpoint devuelve únicamente historias con estado distinto de `Hecho`. |
| RN-11 | El orden es WSJF descendente, nulos al final y desempate por ID ascendente. |
| RN-12 | Los cálculos no se persisten ni se cachean: se recalculan en cada consulta. |
| RN-13 | La API no redondea. Los valores `float64` se emiten completos. |
| RN-14 | El endpoint es de solo lectura: no crea, actualiza ni elimina filas. |
| RN-15 | Cualquier miembro del proyecto puede consultar; un no miembro recibe `404 PROYECTO_NO_ENCONTRADO`, igual que un proyecto inexistente. |
| RN-16 | La prioridad manual de US-02 sigue existiendo y no se recalcula a partir de WSJF. |
| RN-17 | Los tres valores se devuelven en alta, PATCH, listado y Sprint Backlog porque reutilizan la representación de historia. |
| RN-18 | No se agregan parámetros de consulta, filtros ni paginación. |

## Restricciones

### Técnicas

- Solo se usan `SELECT` para el endpoint priorizado.
- No se crean tablas ni índices. Las tres columnas se agregan con `AutoMigrate`.
- No se modifica `tablas.sql`.
- Se reutilizan `parsearID`, `traducirError`, `ExigirAccesoAProyecto` y el
  repositorio `ListarPorProyecto`; no se agrega un método nuevo de persistencia.
- No se agregan dependencias externas.

### De arquitectura por capas

- `domain` no importa Gin, GORM, `context` ni capas superiores.
- `repository` solo persiste y recupera las columnas nuevas; no calcula ROI ni
  WSJF.
- `service` valida y delega las fórmulas y el orden al dominio.
- `handler` traduce HTTP, DTO y errores; no calcula métricas.

### De negocio y contrato

- Solo lectura para los cálculos; los únicos cambios de datos son los valores
  que el usuario carga por alta o PATCH.
- Sin redondeo en la API.
- Sin reordenamiento automático del backlog persistido.
- Sin cambios en la prioridad manual de US-02.

## Casos límite

| ID | Caso | Resultado |
| :--- | :--- | :--- |
| CL-01 | Historia sin `story_points` | `roi` y `wsjf` son `null`; aparece al final. |
| CL-02 | Historia con `story_points = 0` | `roi` y `wsjf` son `null`; aparece al final. |
| CL-03 | Historia sin `valor_negocio` pero con SP | `roi` es `null`; `costo_demora` y `wsjf` se calculan con `0` por ese campo. |
| CL-04 | Historia sin `criticidad_tiempo` y/o `reduccion_riesgo` | Esos campos cuentan `0` en el costo; el WSJF se calcula igual. |
| CL-05 | Historia sin ninguno de los tres valores y con SP | `roi: null`, `costo_demora: 0`, `wsjf: 0`. |
| CL-06 | `story_points = 0.5` | Se permite dividir por `0.5`; ROI y WSJF pueden ser mayores que los valores cargados. |
| CL-07 | Dos historias con el mismo WSJF | Desempate por ID ascendente. |
| CL-08 | Varias historias sin WSJF | Quedan después de las que tienen WSJF, ordenadas por ID ascendente. |
| CL-09 | Historia en `Hecho` | No aparece en la respuesta. |
| CL-10 | Proyecto sin historias o con todas en `Hecho` | `200` con `{ "total": 0, "historias": [] }`. |
| CL-11 | Valor `0`, `0.5`, `4`, `22` o negativo | `422 VALIDACION` con el campo correspondiente. |
| CL-12 | PATCH solo con un valor de negocio | Es una actualización válida y se persiste. |
| CL-13 | PATCH con `null` en un valor ya cargado | Se ignora; el valor persistido no cambia. |
| CL-14 | No miembro | `404 PROYECTO_NO_ENCONTRADO`. |
| CL-15 | `{id}` no numérico o `0` | `400 PARAMETRO_INVALIDO`. |
| CL-16 | Dos consultas consecutivas | Mismo resultado; no hay efectos laterales. |

## Criterios de aceptación

- [ ] **CA-01** Given una historia con `valor_negocio` `8` y `story_points` `5`,
  When se consulta, Then `roi` es `1.6`.
- [ ] **CA-02** Given valor `3`, criticidad `1`, riesgo `1` y story points `3`,
  When se consulta, Then el costo de demora es `5` y `wsjf` es `5/3`
  (`1.6666666666666667` sin redondear).
- [ ] **CA-03** Given una historia sin `story_points`, When se consulta, Then
  `roi` y `wsjf` son `null` y la historia queda al final.
- [ ] **CA-04** Given una historia sin uno de los tres campos de valor, When se
  calcula el costo de demora, Then ese campo cuenta como `0` y el WSJF se
  calcula igual.
- [ ] **CA-05** Given un valor fuera de la escala Fibonacci (`0`, `0.5`, `4`,
  `22`) o negativo, When se envía en alta o PATCH, Then `422 VALIDACION` con el
  campo correspondiente y no se modifica nada.
- [ ] **CA-06** Given varias historias no `Hecho`, When se consulta el backlog
  priorizado, Then aparecen ordenadas por WSJF descendente, con desempate por ID
  ascendente y las de WSJF `null` al final.
- [ ] **CA-07** Given historias en `Hecho`, When se consulta el backlog
  priorizado, Then no aparecen.
- [ ] **CA-08** Given un no miembro, When consulta, Then `404
  PROYECTO_NO_ENCONTRADO`.
- [ ] **CA-09** Given cualquier consulta, When se completa, Then no se modifica
  ningún dato.
- [ ] **CA-10** Given una historia con `story_points = 0`, Then `roi` y `wsjf`
  son `null` y no se produce división por cero.
- [ ] **CA-11** Given una historia con `valor_negocio` vacío y story points
  cargados, Then `roi` es `null`, el costo usa `0` y el WSJF se calcula.
- [ ] **CA-12** Given una respuesta exitosa, Then las claves y tipos coinciden
  con el contrato y los tres valores aparecen también en alta, PATCH, listado y
  Sprint Backlog.

## Condiciones de error y mapeo HTTP

| Situación | Error de dominio | Estado | Código API | Campo |
| :--- | :--- | :---: | :--- | :--- |
| `{id}` no numérico o `0` | — (handler) | 400 | `PARAMETRO_INVALIDO` | — |
| Sin token o token inválido | `ErrTokenRequerido` / `ErrTokenInvalido` | 401 | `TOKEN_REQUERIDO` / `TOKEN_INVALIDO` | — |
| Proyecto inexistente o usuario no miembro | `ErrProyectoNoEncontrado` | 404 | `PROYECTO_NO_ENCONTRADO` | — |
| `valor_negocio` inválido | `ErrValorNegocioInvalido` | 422 | `VALIDACION` | `valor_negocio` |
| `criticidad_tiempo` inválida | `ErrCriticidadTiempoInvalida` | 422 | `VALIDACION` | `criticidad_tiempo` |
| `reduccion_riesgo` inválida | `ErrReduccionRiesgoInvalida` | 422 | `VALIDACION` | `reduccion_riesgo` |
| Error de persistencia al leer | `ErrErrorDePersistencia` o error envuelto | 500 | `ERROR_INTERNO` | — |

Precedencia:

1. El middleware de autenticación resuelve `401`.
2. El guard de membresía se ejecuta antes del handler: con `:id` válido y no
   miembro responde `404`; con `:id` inválido deja pasar y el handler responde
   `400`.
3. En alta y PATCH, la validación de los valores ocurre antes de tocar el
   repositorio.

## Arquitectura (Clean Architecture)

- **domain**: `priorizacion.go` con las fórmulas puras, `HistoriaPriorizada` y
  `PriorizarBacklog`; `historia.go` gana los tres campos en `Historia`,
  `NuevaHistoria` y `ActualizacionHistoria`; `errores.go` gana los tres
  centinelas.
- **repository**: `historia_model.go` gana las tres columnas nullable;
  `mappers.go` las mapea; `historia_repository.go` las escribe en `Actualizar`.
  Se reutiliza `ListarPorProyecto`.
- **service**: `HistoriaService.Priorizar` valida la existencia del proyecto,
  lista las historias y delega filtrado, cálculo y orden a
  `domain.PriorizarBacklog`.
- **handler**: la interfaz `ServicioHistorias` gana `Priorizar`; se registra
  `GET /api/proyectos/:id/backlog/priorizado` en `RegistrarRutasHistorias`; el
  DTO agrega los campos y una respuesta priorizada con `roi` y `wsjf`.
- **cmd/api**: sin cambio de firma de `setupRouter`; la ruta queda detrás de
  autenticación y del guard de membresía por estar bajo `/api/proyectos/:id`.

## Plan de pruebas (TDD)

Ciclo **RED → GREEN → REFACTOR**. Cada ID de escenario BDD se reutiliza como
nombre de subtest.

### Dominio (`internal/domain/priorizacion_test.go`)

- `TestValorPriorizacionValido`: valores válidos, `0`, `0.5`, `4`, `22`,
  negativos.
- `TestCalcularROI`: `8/5`, `3/3`, sin SP, SP `0`, sin valor, SP `0.5`.
- `TestCalcularCostoDemora`: todos cargados, uno o dos vacíos, todos vacíos.
- `TestCalcularWSJF`: `5/3`, sin SP, SP `0`, costo `0`.
- `TestPriorizarBacklog`: orden descendente, empate por ID, nulos al final,
  exclusión de `Hecho`, lista vacía no `nil`.

### Service (`internal/service/historia_priorizacion_test.go`)

- `Priorizar` verifica el proyecto y usa `ListarPorProyecto`.
- `ErrProyectoNoEncontrado` se propaga con `errors.Is`.
- Lista vacía devuelve slice vacío.
- Solo se invocan métodos de lectura.
- Alta y PATCH rechazan valores inválidos sin tocar el repositorio.

### Handler (`internal/handler/historia_priorizacion_test.go`)

- `200` con contrato snake_case, `roi`/`wsjf` y nulos.
- Orden y exclusión de `Hecho`.
- `400 PARAMETRO_INVALIDO`.
- `404 PROYECTO_NO_ENCONTRADO` y `500 ERROR_INTERNO`.
- Alta y PATCH con valores válidos e inválidos, campo correcto.

### Repositorio y migración

- `TestHistoriaPersisteValoresDeNegocio`: crear, leer, actualizar y listar con
  valores y `nil`.
- `TestMigrarEsquemaAgregaValoresDeNegocioNullableSobreFilasExistentes`: quitar
  las columnas, migrar y verificar que son nullable, sin default y que las filas
  previas quedan en `NULL`.

### Router y guard (`cmd/api/router_test.go`)

- No miembro → `404` en `/api/proyectos/1/backlog/priorizado`.
- Sin token → `401`.
- Miembro → llega al handler.

### E2E por curl (manual, documentado en el PR)

1. Alta con los tres valores.
2. PATCH de los tres valores.
3. Backlog priorizado con orden, ROI y WSJF.
4. Historia sin SP y con SP `0` → `null` al final.
5. Valores inválidos en alta y PATCH → `422` con campo.
6. Historia `Hecho` excluida.
7. No miembro → `404`.
8. Dos GET consecutivos sin cambios en la base.

## Resumen de escenarios BDD

Archivo: `docs/features/US-23-priorizacion-backlog.feature`.

| ID | Tipo | Escenario | CA |
| :--- | :--- | :--- | :--- |
| E-01 | Normal | ROI de una historia con valor `8` y SP `5` | CA-01 |
| E-02 | Normal | Costo de demora y WSJF con valores cargados | CA-02 |
| E-03 | Límite | Historia sin story points: ROI/WSJF `null` al final | CA-03, CA-10 |
| E-04 | Alternativo | Campo de valor faltante cuenta `0`; ROI `null` sin valor de negocio | CA-04, CA-11 |
| E-05 | Error | Valores inválidos en alta: `422` y sin escritura | CA-05 |
| E-06 | Error | Valores inválidos en PATCH: `422` y sin modificación | CA-05 |
| E-07 | Normal | Orden por WSJF descendente con desempate por ID y nulos al final | CA-06 |
| E-08 | Alternativo | Historias en `Hecho` excluidas | CA-07 |
| E-09 | Error | No miembro recibe `404 PROYECTO_NO_ENCONTRADO` | CA-08 |
| E-10 | No funcional | Consultas de solo lectura y contrato completo | CA-09, CA-12 |

## Decisiones

- **D-01 — Nulos en el orden.** Las historias sin WSJF se incluyen y quedan al
  final; no se excluyen del endpoint.
- **D-02 — ROI sin valor de negocio.** Si `valor_negocio` es `null`, el ROI es
  `null`; no se aplica `0` como en el costo de demora.
- **D-03 — Story points cero.** `story_points = 0` produce ROI y WSJF `null`
  para evitar la división por cero.
- **D-04 — PATCH con `null`.** Conserva la semántica actual de US-02/US-19: es
  "no informado" y no borra valores.
- **D-05 — Sin redondeo.** La API emite `float64` completo.
- **D-06 — `tablas.sql`.** No se modifica; el esquema se mantiene con
  `AutoMigrate`, igual que en US-19 y US-20.
- **D-07 — Representación compartida.** Los tres valores se agregan a
  `respuestaHistoria`, por lo que también aparecen en Sprint Backlog sin trabajo
  adicional. `roi` y `wsjf` solo se emiten en el endpoint priorizado.

## Definición de terminado

- [ ] `go test ./...` en verde.
- [ ] Spec SDD y feature BDD escritos antes del código.
- [ ] Migración verificada en base limpia y en base existente.
- [ ] Tests de ROI y WSJF con casos sin story points, con campos vacíos y con
  valores mixtos.
- [ ] E2E por curl: alta con valores, PATCH, backlog priorizado y valores
  inválidos.
- [ ] Cross-referencias a US-02, US-05, US-19, US-20, US-11, US-F10 y US-F24.
