# Especificación: SK-017 - golang-modernize: Modernización de código Go

## Objetivo
Modernizar código Go para usar features recientes del lenguaje, mejoras de la standard library y patrones idiomáticos dentro de un alcance controlado (aproximadamente los últimos 3 años de releases). Se priorizan los fixes de seguridad y correctitud primero, luego legibilidad, luego mejoras graduales. Se usa al revisar código con patrones viejos, ante deprecation warnings, pedidos de modernización, upgrades de versión de Go (ej. a 1.27) o refresh de tooling/CI.

## Entradas
- Proyecto Go con `go.mod`/`go.work` (directiva `go`) y código a modernizar (`**/*.go`).
- Modo: **Inline** (el desarrollador está codificando: sugerir solo modernizaciones del archivo/feature actual; registrar el resto como nota) o **Full-scan** (invocación explícita o CI: hasta 5 sub-agentes — paquetes deprecated, language features, stdlib upgrades, testing patterns, tooling/infra — y consolidación).
- Archivo `.modernize` en la raíz del proyecto (sugerencias previamente ignoradas).
- Binarios: `go`, `golangci-lint`.

## Salidas esperadas
1. Código Go modernizado: reemplazos de APIs deprecated, features de lenguaje (range-over-int, min/max, `any`, iterators), stdlib upgrades (slices, maps, cmp, slog), testing (t.Context, b.Loop, synctest), tooling (golangci-lint v2, govulncheck, PGO, CI).
2. Reporte de revisión de `go.mod` con sugerencia de upgrade de versión Go si está atrasado.
3. Archivo `.modernize` actualizado con las sugerencias ignoradas por el equipo.
4. En Full-scan mode: aplicación de la modernización en un worktree aislado para revisión antes del merge.

## Reglas de negocio
1. NUNCA realizar refactoring grande si el desarrollador está trabajando en otra tarea — pero intentar convencerlo de que mejoraría la calidad.
2. En Inline mode, solo sugerir mejoras relacionadas con el código actual; mencionar las demás oportunidades como nota con su ganancia de calidad.
3. Chequear el `go` directive del `go.mod`/`go.work` y comparar con la última versión de Go en la tabla de changelogs del skill.
4. Leer `.modernize` (sugerencias ignoradas) y NO re-sugerir nada listado allí.
5. Correr `golangci-lint` con el linter `modernize` si está disponible y `go test ./...`; en Go 1.27+ el vet check `stdversion` marca APIs más nuevas que el directive del módulo (subir el directive o revertir la sugerencia, no ignorar el hit).
6. Antes de sugerir una actualización de dependencia, correr `go mod tidy` y la suite de tests; pedir al desarrollador que revise changelog y release notes.
7. Si el desarrollador ignora explícitamente una sugerencia, escribir una línea en `.modernize` (formato: `fecha categoria descripcion`).
8. Para renames de identificadores o reemplazos de APIs deprecated, usar `golang-gopls` (rename seguro con diagnostics post-edit).
9. En Full-scan y cambios grandes, aplicar en un worktree aislado — el árbol principal queda a salvo hasta la revisión.

## Restricciones
- No cubre refactorings estructurales, extracción de funciones ni movimientos entre paquetes (→ skill `golang-refactoring`).

## Casos límite
- Proyecto con `go.mod` más viejo que la fila más antigua de la tabla del skill: mejoras con cobertura más acotada; sugerir upgrade de Go primero.
- Sugerencia que choca con una dependencia legacy (`math/rand` vs `math/rand/v2`): registrar la decisión en `.modernize`.
- Modernización que toca decenas de archivos: worktree aislado + PR por categoría.

## Condiciones de error
- `golangci-lint` modernize linter no disponible: el escaneo se apoya en grep/vet/stdversion sin bloquearse.
- Suggestion ignorada por el desarrollador: se escribe en `.modernize` y no se re-sugiere en la sesión.
- Ventana de Go 1.26→1.27 tocando allocator: no atribuir deltas de benchmark al código (ver skill `golang-benchmark`).

## Criterios de aceptación
- [ ] Se verificó el `go` directive y se comparó con la última versión de Go.
- [ ] En Inline mode solo se tocaron archivos de la tarea actual.
- [ ] `.modernize` se respeta (nada listado se vuelve a sugerir) y se actualiza con cada rechazo.
- [ ] Se corrieron `golangci-lint` (modernize) y `go test ./...` antes de finalizar.
- [ ] Los renames de APIs se hicieron con gopls (seguros).
- [ ] Las modernizaciones masivas se aplicaron en worktree aislado.

## Escenarios BDD

### Caso normal
Given un proyecto con `go 1.22` y código que usa `interface{}` y `ioutil`
When el agente ejecuta Full-scan mode
Then sugiere subir el directive y reemplazar por `any` y las APIs de `os`/`io`
And verifica con `go mod tidy` + tests + vet stdversion

### Caso alternativo
Given el desarrollador trabajando en una feature con código legacy adyacente
When el agente nota oportunidades de modernización fuera del archivo actual
Then solo menciona las oportunidades como nota con su ganancia
And no refactoriza los archivos fuera de la tarea

### Caso límite
Given una modernización que renombra un identificador en decenas de call sites
When el agente la aplica
Then usa el rename seguro de `golang-gopls`
And corre diagnostics post-edit para confirmar que no rompió interfaces

### Caso de error
Given el equipo que rechaza la migración a `math/rand/v2` por compatibilidad con un módulo legacy
When el agente registra el rechazo
Then agrega la línea correspondiente en `.modernize`
And no vuelve a sugerir la migración en sesiones futuras