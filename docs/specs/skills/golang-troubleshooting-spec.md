# Especificación: SK-030 - golang-troubleshooting: Troubleshooting y depuración de programas Go

## Objetivo
Definir el proceso sistemático para encontrar y corregir la causa raíz de bugs, crashes, deadlocks, races y comportamientos inesperados en código Go. **Sin fixes sin investigación de causa raíz**: los fixes de síntoma crean bugs nuevos y desperdician tiempo; aplica especialmente bajo presión de tiempo. Cubre metodología de debugging, pitfalls comunes de Go, debugging dirigido por tests, pprof, Delve, race detection, tracing GODEBUG y debugging en producción.

## Entradas
- Reporte de bug/crash/deadlock/race/comportamiento inesperado en código Go (`**/*.go`).
- Modo: **Single-issue debug** (default: reglas de oro secuenciales — leer el error, reproducir, una hipótesis a la vez; sin sub-agentes) o **Codebase bug hunt** (auditoría explícita de un codebase grande: hasta 5 sub-agentes por categoría — nil/interface, recursos, error handling, races, context/slice/map).
- Binarios: `go`, `dlv` (`go install github.com/go-delve/delve/cmd/dlv@latest`).

## Salidas esperadas
1. Diagnosis del symptom category (build/compilación, lógica, crashes/panics, intermitente, hang, CPU alto, memoria) con el arbol de decisión del skill.
2. Reproducción confiable del bug (test que falla, comando mínimo) antes de cualquier fix.
3. Hipótesis única con razonamiento y verificación (una a la vez).
4. Causa raíz identificada y fix explicable ("nunca propose un fix que no pueda explicar").
5. Uso incremental de herramientas: `fmt.Println`/tests primero; pprof/Delve/GODEBUG solo cuando las simples no alcanzan.
6. En bug hunt: reporte consolidado por categoría de bug.

## Reglas de negocio
1. Comenzar con el Decision Tree del skill para clasificar el síntoma y saltar a la sección relevante.
2. Seguir las Golden Rules: **reproducir antes de fixear**, **una hipótesis a la vez**, **encontrar la causa raíz**.
3. No saltarse pasos de la metodología general de debugging.
4. Vigilar red flags en el propio razonamiento: si uno se está adivinando fixes sin entender la causa, parar y reunir más evidencia.
5. Escalar herramientas incrementalmente: empezar con las más simples y recurrir a pprof/Delve/GODEBUG solo cuando es necesario.
6. NUNCA proponer un fix que no se puede explicar — si no se entiende por qué pasa el bug, decirlo e investigar más.
7. En single-issue debug no lanzar sub-agentes: la investigación secuencial enfocada es más rápida para un síntoma conocido.
8. En codebase bug hunt, lanzar hasta 5 sub-agentes paralelos, uno por categoría, solo cuando se pide un sweep amplio.

## Restricciones
- Interpretación de perfiles y benchmarking → skill `golang-benchmark`.
- Aplicar patrones de optimización → skill `golang-performance`.
- Diseñar código defensivo/concurrente nuevo → skills `golang-safety`/`golang-concurrency`.

## Casos límite
- Bug intermitente ("a veces funciona"): `go test -race` y aislamiento del test sospechoso con `-count=100 -failfast`.
- Programa colgado: inspectar `debug/pprof/goroutine?debug=2`.
- Panic aleatorio: `GOTRACEBACK=all` + `-race`.
- Alto uso de CPU: pprof CPU profile de 30s e inspección con `top`/`web`/`list funcName`.
- Pérdida de memoria: heap profile por diferencia entre dos puntos en el tiempo.

## Condiciones de error
- Hito de reproducción no logrado: NO fixear — volver a juntar evidencia (input exacto, versión, entorno).
- Hipótesis descartada por la evidencia: descartarla explícitamente y formular la siguiente, sin aferrarse.
- Fix de síntoma aplicado sin causa raíz: se rechaza en review si no hay explicación del porqué.

## Criterios de aceptación
- [ ] El síntoma se clasificó con el Decision Tree del skill.
- [ ] Se reprodujo el bug de forma confiable antes de tocar código.
- [ ] Se trabajó con una hipótesis a la vez, verificada por evidencia.
- [ ] Se identificó la causa raíz y el fix es explicable.
- [ ] Las herramientas se escalaron de simples a avanzadas (pprof/Delve/GODEBUG solo cuando hicieron falta).
- [ ] En bug hunt, los hallazgos se consolidaron por categoría para el reporte.

## Escenarios BDD

### Caso normal
Given un bug "el endpoint devuelve 500 solo para ciertos inputs"
When el agente debuggea
Then reproduce con un test que falla y el input exacto
And formula una hipótesis única (validación faltante en la query)
And la verifica, corrige la causa raíz y corre el test en verde

### Caso alternativo
Given un crash aleatorio que aparece en producción
When el agente investiga
Then usa `GOTRACEBACK=all` y `go test -race` para reproducirlo
And sospecha un nil deref o un write a map nil según la evidencia
And verifica con la stack trace antes de corregir

### Caso límite
Given un programa que se cuelga ocasionalmente
When el agente diagnostica
Then inspecciona `pprof/goroutine?debug=2` en busca de goroutines bloqueadas
And determina si es deadlock, espera externa o canal sin consumidor

### Caso de error
Given un bug intermitente y una hipótesis tentadora de fix de síntoma
When el agente no logró reproducirlo aún
Then NO aplica el fix
And declara que aún no entiende la causa raíz y continúa la investigación con más evidencia