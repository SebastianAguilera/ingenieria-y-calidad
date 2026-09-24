# Especificación: SK-020 - golang-performance: Optimización de rendimiento en Go

## Objetivo
Definir la metodología y los patrones de optimización de rendimiento en Go: si hay un bottleneck X, aplicar el patrón Y. Cubre reducción de alocaciones, eficiencia de CPU, layout de memoria, tuning del GC, pooling, caching, optimización de hot paths y review de rendimiento de código. La disciplina central: **nunca optimizar sin perfilar antes** — medir, formular hipótesis, cambiar una cosa, re-medir.

## Entradas
- Código Go con bottleneck identificado por profiling/benchmarks (`**/*.go`).
- Modo: **Review (architecture)** — hasta 3 sub-agentes (1: alocaciones y layout de memoria; 2: I/O y concurrencia; 3: complejidad algorítmica y caching) para scan de anti-patterns estructurales; **Review (hot path)** — análisis enfocado de una función/loop (secuencial, 1 sub-agente); **Optimize** — ciclo iterativo (definir métrica → baseline → diagnosticar → mejorar → comparar), un cambio a la vez.
- Binarios: `go`, `benchstat` (`go install golang.org/x/perf/cmd/benchstat@latest`), auxiliares: fieldalignment, staticcheck, fgprof, perf.

## Salidas esperadas
1. Diagnóstico del bottleneck con evidencia de perfil (pprof/fgprof/tracing), descartando bottlenecks externos primero.
2. Aplicación de UNA optimización a la vez con comentario explicativo.
3. Comparación estadística `benchstat` entre baseline y versión optimizada.
4. Commit con la salida de benchstat en el cuerpo y tipo `perf(scope): summary`.
5. Trail de auditoría: archivos `/tmp/report-*.txt` conservados y numerados.
6. Comentarios en el código explicando por qué el patrón es más rápido (con números si hay).

## Reglas de negocio
1. **Perfilar antes de optimizar** — la intuición sobre bottlenecks falla ~80% de las veces.
2. La reducción de alocaciones da el mayor ROI; el GC es rápido pero no gratis.
3. Documentar las optimizaciones: comentarios con el motivo y números de benchmark para que el futuro lector no revierta una optimización "innecesaria".
4. Descartar bottlenecks externos primero: si el 90% de la latencia es una query lenta o una API externa, reducir alocaciones no ayuda. Diagnosticar con fgprof (on/off-CPU), pprof goroutine profile y tracing distribuido.
5. Ciclo iterativo: Definir métrica → benchmark atómico → baseline → diagnosticar → mejorar (UN cambio) → comparar con benchstat → commit → repetir.
6. Cuando múltiples candidatas compiten por el mismo bottleneck, implementar cada una en un worktree aislado vía sub-agente y comparar después; medir en serie (corridas concurrentes en CPU compartida contaminan).
7. Seguir los deep-dives del skill: memory (alocaciones), cpu (hot loop), io (I/O + concurrencia), algorithms (complejidad/caching).

## Restricciones
- La metodología de medición → skill `golang-benchmark`.
- El workflow de debug → skill `golang-troubleshooting`.

## Casos límite
- Hot path single function: review secuencial sin fan-out (el fan-out solo paga a escala package/service).
- Bottleneck en I/O externo (BD, API): optimizar el componente externo (query tuning, caching, connection pools, circuit breakers) antes del código.
- Loop caliente con alocaciones en cada iteración: mover alocaciones fuera del loop / pooling.
- Estructura con mala alineación de campos que infla el size: reordenar campos (fieldalignment).

## Condiciones de error
- Cambio aplicado sin baseline medido: se revierte — no se puede probar mejora sin medición previa.
- Delta sin significancia estadística (dentro del ruido del benchmark): no se commitea como "perf win"; se reporta como inconcluyente.
- Optimización que degrada la legibilidad sin números que la justifiquen: documentar o descartar.

## Criterios de aceptación
- [ ] El bottleneck se verificó con profiling antes de optimizar.
- [ ] Se descartaron bottlenecks externos (I/O/BD) antes que los internos del proceso.
- [ ] Se aplicó una sola optimización por iteración, con comentario explicativo.
- [ ] La mejora se comparó con `benchstat` y es estadísticamente significativa.
- [ ] La salida de benchstat se incluyó en el mensaje del commit.
- [ ] Los archivos `/tmp/report-*.txt` quedaron como trail de auditoría.

## Escenarios BDD

### Caso normal
Given un endpoint con latencia alta y un heap profile que muestra 60% de tiempo en GC
When el agente optimiza
Then diagnóstica alocaciones excesivas en el hot path
And aplica una reducción de alocaciones (reuso de buffers / menos escapes)
And compara con benchstat antes de commitear

### Caso alternativo
Given un review de arquitectura en un servicio con posibles anti-patterns
When el agente audita en Review mode (architecture)
Then lanza 3 sub-agentes en paralelo (alocaciones, I/O/concurrencia, algoritmos/caching)
And consolida los hallazgos estructurados

### Caso límite
Given una función caliente con un loop que aloca en cada iteración
When el agente la optimiza
Then extrae la alocación fuera del loop o usa `sync.Pool`
And verifica con benchmarks que la mejora es real

### Caso de error
Given una optimización propuesta sin benchmark de baseline
When el agente la revisa
Then la rechaza por falta de evidencia de medición
And solicita baseline y comparación estadística antes de aceptar el cambio