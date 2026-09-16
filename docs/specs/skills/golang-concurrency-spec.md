# Especificación: SK-004 - golang-concurrency: Diseño de concurrencia en Go

## Objetivo
Brindar reglas de diseño de concurrencia en Go: ciclo de vida de goroutines y prevención de leaks, channels y `select`, ownership y dirección de channels, primitivas de sincronización (`Mutex`/`RWMutex`/`sync.Map`/`sync.Once`/atómicos), `errgroup`, `singleflight`, worker pools y pipelines fan-out/fan-in. La correctitud y la ausencia de leaks priman sobre el rendimiento.

## Entradas
- Código Go concurrente a escribir, revisar o auditar (`**/*.go`).
- Modo de operación: Write (implementar), Review (revisar diff de PR) o Audit (auditar codebase, hasta 5 sub-agentes paralelos).
- Binarios: `go`.

## Salidas esperadas
1. Código concurrente correcto y libre de leaks (Write mode).
2. Reporte de revisión sobre el diff concurrente (Review mode): leaks, propagación de context, violaciones de ownership, estado compartido sin protección.
3. Reporte consolidado de auditoría con hallazgos por categoría (Audit mode).
4. Tests que detectan leaks de goroutines con `go.uber.org/goleak`.

## Reglas de negocio
1. Toda goroutine DEBE tener una salida clara (context, done channel o WaitGroup); sin mecanismo de cierre, leak y acumulación hasta el crash.
2. Compartir memoria comunicando: los channels transfieren ownership explícitamente.
3. Enviar **copias, no punteros** por channels (los punteros crean memoria compartida invisible).
4. Solo el **sender** cierra un channel (cerrar desde el receptor panic si el sender escribe después).
5. Especificar la dirección del channel (`chan<-`, `<-chan`) — el compilador previene el mal uso en build time.
6. Preferir channels sin buffer por defecto; buffers más grandes enmascaran backpressure.
7. Incluir siempre `ctx.Done()` en `select` para evitar leaks ante cancelación.
8. Evitar `time.After` repetido en loops calientes (`time.NewTimer` + `Reset`).
9. Rastrear leaks en tests con `go.uber.org/goleak`.
10. Elegir la primitiva según la tabla del skill: channels (transferencia de datos), Mutex/RWMutex (estado compartido), atómicos (contadores/flags), sync.Map (mapa read-heavy), errgroup (esperar + primer error + límite de concurrencia).

## Restricciones
- No cubre bugs defensivos no relacionados con concurrencia (nil, slices aliasing, overflow → skill `golang-safety`).
- No depura programas específicos colgados o con race condition después del hecho (→ skill `golang-troubleshooting`).

## Casos límite
- Goroutine sin exit path ante cancelación temprana del caller: la regla 7 obliga `ctx.Done()` en el select.
- Upgrade de RLock a Lock en `RWMutex`: deadlock — nunca se permite.
- Mapa accedido concurrentemente: crash duro — usar `sync.Map` solo read-heavy o `Mutex`+map si dominan las escrituras.
- Canal sin buffer con productor más rápido que consumidor: backpressure natural, se deja así salvo justificación medida.

## Condiciones de error
- Panic por cerrar un channel dos veces o escribir tras cierre: se corrige garantizando single-sender y single-closer.
- Leak de goroutine detectado en CI con goleak: el test falla hasta eliminar el leak.
- Deadlock por upgrade RLock→Lock: se detecta en review/test y se rediseña la sección crítica.

## Criterios de aceptación
- [ ] Toda goroutine tiene un mecanismo de salida y no se filtran en tests (goleak).
- [ ] Los channels declaran dirección y solo el sender los cierra.
- [ ] Los select incluyen `ctx.Done()` cuando aplica cancelación.
- [ ] El estado compartido se protege con la primitiva correcta según el caso de uso.
- [ ] En Review mode, el informe cubre leaks, context, ownership y estado compartido.
- [ ] En Audit mode, los hallazgos se consolidan por categoría.

## Escenarios BDD

### Caso normal
Given un worker pool que procesa tareas de un channel
When el agente lo implementa con el skill
Then las goroutines tienen salida limpia vía context o done channel
And se detectan leaks con goleak en los tests

### Caso alternativo
Given una operación que debe ejecutarse una sola vez por arranque
When se compara `sync.Once` vs inicialización manual
Then se usa `sync.Once` o `singleflight` para garantizar ejecución exactamente una vez bajo concurrencia

### Caso límite
Given un mapa compartido con muchos lectores y pocos escritores
When el agente elige la estructura
Then usa `sync.Map` para el caso read-heavy
And documenta que con escrituras dominantes conviene `RWMutex` + map

### Caso de error
Given un canal que el receptor cierra y el sender continúa escribiendo
When el agente revisa el código
Then detecta la violación del ownership (solo el sender cierra)
And corrige para que el cierre ocurra desde el sender