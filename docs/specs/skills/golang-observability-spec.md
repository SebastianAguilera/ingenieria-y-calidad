# Especificación: SK-019 - golang-observability: Observabilidad de servicios Go

## Objetivo
Instrumentar servicios Go con las cinco señales de observabilidad: **logs** estructurados (slog), **métricas** (Prometheus), **traces** distribuidos (OpenTelemetry), **perfiles** (pprof/Pyroscope) y **RUM** server-side. Una feature no está terminada hasta que es observable. Aplica al instrumentar servicios, configurar métricas/alertas, agregar tracing OTel, correlacionar logs con traces, migrar loggers legacy a slog o implementar tracking compatible con GDPR/CCPA.

## Entradas
- Servicio Go a instrumentar (`**/*.go`).
- Modo: **Coding/instrumentation** (declarar métricas, abrir/cerrar spans, logging estructurado, pprof toggles), **Review** (diff de PR: señales esperadas exportadas) o **Audit** (hasta 5 sub-agentes, uno por señal).
- Binarios: `go`.

## Salidas esperadas
1. Logs estructurados JSON con `log/slog` (no strings libres), niveles correctos, contexto correlacionado (`slog.InfoContext`).
2. Métricas Prometheus: Histogram para latencia (no Summary), métricas de error rate por endpoint.
3. Spans de OpenTelemetry en operaciones significativas (service methods, DB queries, API calls, message queue).
4. Propagación de context con trace_id/span_id/deadlines.
5. Profiling on/off por variables de entorno (sin redeploy).
6. Correlación de señales: trace_id en logs, exemplars en métricas.
7. En Audit mode: reporte de cobertura por señal con hallazgos consolidados.

## Reglas de negocio
1. USAR logging estructurado con `log/slog` — los servicios de producción DEBEN emitir logs estructurados (JSON), no strings libres.
2. Elegir el nivel correcto: Debug (dev), Info (normal), Warn (estado degradado), Error (fallas que requieren atención).
3. Loguear con context: `slog.InfoContext(ctx, ...)` para correlacionar con traces.
4. Preferir **Histogram** sobre Summary para latencia; TODO endpoint HTTP DEBE tener métricas de latencia y error rate.
5. Mantener baja cardinalidad de labels en Prometheus — NUNCA valores unbounded (user IDs, URLs completas) como valores de label.
6. Percentiles P50/P90/P99/P99.9 con Histograms + `histogram_quantile()`.
7. Configurar el TracerProvider de OTel temprano y agregar spans en todas las operaciones significativas.
8. Propagar context en todas partes — es el vehículo de trace_id, span_id y deadlines.
9. Habilitar profiling por variables de entorno (pprof/continuous profiling on/off sin redeploy).
10. Correlacionar señales: trace_id en logs, exemplars para unir métricas con traces.
11. En Go 1.26+: preferir `slog.NewMultiHandler` del stdlib para fan-out antes de librerías third-party de handlers.
12. Una feature no está terminada hasta que es observable.

## Restricciones
- La investigación profunda de rendimiento (profiling temporal) → skills `golang-benchmark`/`golang-performance`.
- El uso de observabilidad para diagnosticar issues de producción → skill `golang-troubleshooting`.

## Casos límite
- Endpoint sin métricas de latencia ni error rate: violación — agregar Histogram + counter.
- Log con user ID en el mensaje (alta cardenalidad): degrada la agrupación — adjuntar como atributo estructurado, no en el template del mensaje.
- Fan-out a múltiples sinks de logs: `slog.NewMultiHandler` estándar antes que composición third-party.

## Condiciones de error
- `slog` sin variante de context en handlers con trace activo: logs sin correlación — usar `slog.InfoContext`.
- Métrica con label de URL completa o user ID: cardinality explosion en Prometheus — rediseñar labels a valores acotados.
- pprof expuesto sin protección de red/auth: vector de ataque — aplicar las medidas del skill `golang-security`.

## Criterios de aceptación
- [ ] Los logs son estructurados con `slog` y niveles correctos.
- [ ] Cada endpoint HTTP exporta métricas de latencia y error rate.
- [ ] Los spans OTel cubren operaciones significativas.
- [ ] El context se propaga con trace_id/spans deadlines.
- [ ] Los labels de métricas mantienen baja cardinalidad.
- [ ] El profiling se activa/desactiva por entorno sin redeploy.
- [ ] En Review mode el diff exporta las señales esperadas.

## Escenarios BDD

### Caso normal
Given un nuevo endpoint HTTP de creación de usuario
When el agente lo instrumenta
Then agrega un Histogram de latencia y un counter de errores para el endpoint
And registra un span OTel y usa `slog.InfoContext` en el handler

### Caso alternativo
Given un servicio que migra de `log.Printf` a logging estructurado
When el agente realiza la migración
Then reemplaza los llamados por `slog` con niveles y atributos estructurados
And conserva los mensajes de log estables para agregación (baja cardenalidad)

### Caso límite
Given un fan-out de logs a stdout y a un audit handler
When el agente evita dependencias innecesarias
Then usa `slog.NewMultiHandler` del stdlib (Go 1.26+)
And agrega librerías third-party solo si el stdlib no alcanza

### Caso de error
Given una métrica con label de URL completa (alta cardenalidad)
When el agente audita la instrumentación
Then detecta el riesgo de cardinality explosion en Prometheus
And corrige usando labels acotados (ej. ruta normalizada o endpoint name)