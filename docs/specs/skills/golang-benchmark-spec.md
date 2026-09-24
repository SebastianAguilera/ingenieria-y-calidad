# Especificación: SK-002 - golang-benchmark: Benchmarking y medición de rendimiento en Go

## Objetivo
Proveer la metodología de medición de rendimiento en Go: escribir y ejecutar benchmarks, perfilar código caliente con pprof, interpretar perfiles CPU/memoria/trace, comparar resultados con rigor estadístico (benchstat) y detectar regresiones en CI. El skill NO provee patrones de optimización (eso corresponde a `golang-performance`): mide primero, optimiza después.

## Entradas
- Código Go candidato a medición (`**/*.go`).
- Solicitud de benchmark, perfilado o comparación de rendimiento (tarea del operador).
- Binarios requeridos: `go`, `benchstat` (`go install golang.org/x/perf/cmd/benchstat@latest`).
- Herramientas permitidas: Read, Edit, Write, Glob, Grep, Bash (go, golangci-lint, git, benchstat, benchdiff, cob, gobenchdata, curl), WebFetch, WebSearch, agentes.

## Salidas esperadas
1. Archivos `*_bench_test.go` nombrados según el archivo fuente (`parser.go` → `parser_bench_test.go`), con `Benchmark*` en el mismo orden que las funciones medidas.
2. Benchmarks con `b.Loop()` (Go 1.24+) o `b.N` legacy, con `b.ReportAllocs()`/`-benchmem` cuando corresponda.
3. Comparaciones con `benchstat` de corridas antes/después (idealmente 6+ iteraciones) con conclusión estadísticamente válida.
4. Perfiles pprof interpretados (CPU, memoria, goroutine, trace) aplicando la metodología del skill.
5. Reporte de detección/inexistencia de regresiones de rendimiento.

## Reglas de negocio
1. No se extraen conclusiones de una sola corrida de benchmark (rigor estadístico y condiciones controladas antes de decidir).
2. Los benchmarks viven en archivos `_bench_test.go` separados de los tests de correctitud.
3. Para Go 1.24+ se prefiere `b.Loop()` sobre `b.N` (evita errores de dead-code elimination).
4. Comparaciones que cruzan un cambio de toolchain (ej. Go 1.26→1.27) se deben re-correr con el mismo toolchain antes de confiar en el delta (el allocator de 1.27 cambia las líneas base).
5. Benchmark con fixtures de tamaño adecuado para medición y setup excluido del timing.
6. Las corridas de benchmark concurrentes en CPU compartida contaminan los resultados (medir en serie).

## Restricciones
- Alcance: medición y metodología; la optimización es del skill `golang-performance`; el pprof en servicios corriendo es del skill `golang-troubleshooting`.
- Requiere toolchain Go y `benchstat` instalados.

## Casos límite
- Benchmark cuya función es eliminada por el compilador: usar sink (`_ = sink`) y `b.Loop()`.
- Benchmark a través de una actualización de toolchain de Go: re-medir la línea base "before" en el mismo toolchain.
- Benchmarks sobre código con variabilidad externa (I/O de red/BD): descartar mediciones contaminadas, medir en condiciones controladas.
- Sub-80 bytes de asignación entre Go 1.26 y 1.27: atribuir el delta al allocator, no al código.

## Condiciones de error
- Resultados sin significancia estadística (delta dentro del ruido): se reporta "sin cambio concluyente".
- Falta de `report-*.txt` o corridas insuficientes: se solicitan más corridas antes de concluir.
- Herramientas ausentes (`benchstat` no instalado): se provee la instrucción de instalación y se pausa la comparación.

## Criterios de aceptación
- [ ] Los benchmarks se escriben en archivos `_bench_test.go` con el convenio de nombrado del skill.
- [ ] Se usa `b.Loop()` cuando el go.mod soporta Go 1.24+.
- [ ] Se miden asignaciones cuando el objetivo incluye memoria.
- [ ] Toda conclusión se respalda con comparación `benchstat` y corridas suficientes.
- [ ] Los perfiles pprof se interpretan con la metodología del skill.
- [ ] Se identifica cuándo un delta corresponde al toolchain y no al código.

## Escenarios BDD

### Caso normal
Given un archivo `parser.go` con funciones `Parse` y `Encode`
When el agente crea benchmarks para ellas
Then genera `parser_bench_test.go` con `BenchmarkParse` y `BenchmarkEncode` en el mismo orden que el fuente

### Caso alternativo
Given un benchmark sobre una versión con una optimización aplicada
When se comparan las corridas "before" y "after" con `benchstat`
Then se reporta si la mejora es estadísticamente significativa, con el delta y su intervalo

### Caso límite
Given un benchmark que verifica el paso entre Go 1.26 y Go 1.27
When se comparan resultados que cruzan el cambio de toolchain
Then el agente re-corre la línea base "before" en el toolchain de "after"
And atribuye el delta al allocator si corresponde, no al código

### Caso de error
Given una medición de un solo run con alta varianza
When el agente debe concluir sobre una mejora
Then no emite conclusión de mejora
And solicita más iteraciones o condiciones controladas antes de comparar