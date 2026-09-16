# Especificación: SK-007 - golang-data-structures: Estructuras de datos en Go

## Objetivo
Guía para elegir y usar correctamente las estructuras de datos de Go y su standard library: slices (internals, crecimiento de capacidad, preasignación, paquete `slices`), maps (internals, buckets hash, paquete `maps`), arrays, `container/list/heap/ring`, `strings.Builder` vs `bytes.Buffer`, colecciones genéricas y punteros (`unsafe.Pointer`, `weak.Pointer`). Se elige la estructura correcta razonando sobre layout de memoria, costo de asignación y patrones de acceso.

## Entradas
- Código Go que usa o selecciona estructuras de datos (`**/*.go`).
- Tarea del operador: elegir/optimizar una estructura, implementar contenedores genéricos, usar paquetes `container/`, punteros unsafe/weak, o dudas sobre internals de slice/map.
- Binarios: `go`; herramientas auxiliares: `godig`, `gopls`, LSP/MCP.

## Salidas esperadas
1. Código Go que usa la estructura de datos apropiada con su API correcta.
2. Slices y maps preasignados con `make(T, 0, n)` / `make(map[K]V, n)` cuando el tamaño se conoce o estima.
3. Uso de `slices`/`maps` (Go 1.21+) cuando corresponde.
4. Selección documentada entre `container/list/heap/ring` según el patrón de acceso.
5. Uso correcto de `strings.Builder` (concat) vs `bytes.Buffer` (I/O bidireccional).
6. Genéricos con el constraint más ajustado; `unsafe.Pointer` solo con los 6 patrones válidos; `weak.Pointer[T]` para caches (Go 1.24+).

## Reglas de negocio
1. Preasignar slices y maps cuando el tamaño se conoce o estima (evita copias de crecimiento y rehashing).
2. Preferir arrays solo para tamaños fijos conocidos en compile time (digests, IPv4, dimensiones de matriz).
3. NUNCA depender del timing de crecimiento de capacidad del slice (el algoritmo cambió entre versiones y puede volver a cambiar).
4. `container/heap` para priority queues, `container/list` solo con inserciones frecuentes al medio, `container/ring` para buffers circulares de tamaño fijo.
5. `strings.Builder` para construir strings; `bytes.Buffer` para I/O bidireccional (implementa `io.Reader` y `io.Writer`).
6. Estructuras genéricas DEBEN usar el constraint más ajustado (`comparable` para keys, interfaces propias para orden).
7. `unsafe.Pointer` DEBE seguir solo los 6 patrones de conversión del Go spec; NUNCA guardarlo en una variable `uintptr` entre statements.
8. `weak.Pointer[T]` (Go 1.24+) para caches y maps de canonicalización que permitan al GC reclamar entradas.
9. Un slice es un header de 3 palabras (puntero, len, cap); múltiples slices pueden compartir backing array (cuidado con el aliasing → skill `golang-safety`).

## Restricciones
- Correctitud defensiva (nil maps, append aliasing, copias defensivas) → skill `golang-safety`.
- Channels y sync primitives → skill `golang-concurrency`.
- Optimización una vez identificado el bottleneck con profiling → skill `golang-performance`.

## Casos límite
- Slice sin preasignar en loop con miles de items: crecimiento O(n) repetido — usar `slices.Grow` o `make` con capacidad estimada.
- Map con key no comparable: el compilador rechaza — rediseñar el tipo de key.
- Cache que impide al GC recuperar memoria: migrar a `weak.Pointer[T]` (Go 1.24+).
- String construida con `+=` en loop: N veces alocaciones y copias O(n²) — usar `strings.Builder`.

## Condiciones de error
- Escritura en map nil: panic — inicializar siempre (`make` o literal vacío).
- `unsafe.Pointer` guardado en `uintptr` entre statements: undefined behavior — corregir siguiendo los 6 patrones válidos.
- Uso de `container/list` en hot path con muchos nodos chicos: alta presión de GC — evaluar slice/slab.

## Criterios de aceptación
- [ ] La estructura elegida es la apropiada para el patrón de acceso (justificado en la revisión).
- [ ] Slices y maps se preasignan cuando el tamaño se conoce o estima.
- [ ] No se depende del timing de crecimiento de capacidad.
- [ ] `strings.Builder`/`bytes.Buffer` se usan según el caso (concat vs I/O bidireccional).
- [ ] Los genéricos usan el constraint más ajustado.
- [ ] `unsafe.Pointer` respeta los 6 patrones válidos; los caches consideran `weak.Pointer[T]`.

## Escenarios BDD

### Caso normal
Given un loop que construye un string concatenando miles de fragmentos
When el agente aplica el skill
Then reemplaza la concatenación por `strings.Builder`
And documenta el motivo del cambio

### Caso alternativo
Given un buffer que debe soportar lectura y escritura simultánea (I/O)
When el agente elige el tipo
Then usa `bytes.Buffer` porque implementa `io.Reader` e `io.Writer`
And evita `strings.Builder` que solo sirve para escritura

### Caso límite
Given un map que se poblará con 10.000 entradas de tamaño conocido
When el agente lo inicializa
Then usa `make(map[K]V, n)` para evitar rehashing durante la carga

### Caso de error
Given un slice compartido con `append` que reutiliza el backing array
When el agente detecta escritura cruzada entre dos slices
Then corrige con copia defensiva o `slices.Clone`
And documenta el riesgo de aliasing