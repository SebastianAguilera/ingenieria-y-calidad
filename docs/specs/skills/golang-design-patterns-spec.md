# Especificación: SK-011 - golang-design-patterns: Patrones de diseño idiomáticos en Go

## Objetivo
Aplicar patrones y modismos de diseño en Go de forma idiomática: functional options, APIs de constructores, evitación de `init()` y estado global, enums, decisiones panic vs error, gestión de recursos y lifecycle, graceful shutdown, timeouts y retries, streaming e iterators, y estilos de arquitectura (clean, hexagonal, DDD, flat). Los patrones se aplican solo cuando resuelven un problema real, nunca para demostrar sofisticación.

## Entradas
- Código Go nuevo o existente a diseñar/revisar (`**/*.go`).
- Modo: **Design** (crear APIs/paquetes/estructura: preguntar preferencia de arquitectura, favorecer el patrón más chico que satisfaga el requerimiento) o **Review** (auditar: abuso de `init()`, recursos sin límite, timeouts faltantes, estado global implícito).
- Binarios: `go`.

## Salidas esperadas
1. Constructores con functional options donde la API deba evolucionar.
2. Opciones que retornan error cuando la validación puede fallar.
3. Código sin `init()` con efectos de setup (constructores explícitos).
4. Enums con sentinel Unknown en 0.
5. Errores manejados primero con early return; happy path plano.
6. Panic reservado para bugs, no para errores esperados.
7. `defer Close()` inmediato, `runtime.AddCleanup` sobre `SetFinalizer`.
8. Toda llamada externa con timeout; recursos limitados; retries con check de context entre intentos.
9. `strings.Builder` para concatenación en loops; `[]byte` para mutación/I/O; iterators Go 1.23+ para lazy eval; streaming para transferencias grandes.
10. `//go:embed` para assets estáticos; regexp compilado a nivel paquete; compile-time interface checks.

## Reglas de negocio
1. Constructores DEBEN usar functional options (una función por opción, sin breaking changes).
2. Las options DEBEN retornar error si la validación puede fallar (config mala atrapada en construcción, no en runtime).
3. Evitar `init()`: se ejecuta implícitamente, no puede retornar error y vuelve los tests impredecibles.
4. Enums DEBEN empezar en 1 o con sentinel Unknown en 0 (el zero value pasa silenciosamente como primer miembro).
5. Errores DEBEN manejarse primero con early return.
6. Panic es para bugs, no para errores esperados.
7. `defer Close()` inmediatamente después de abrir.
8. `runtime.AddCleanup` sobre `runtime.SetFinalizer` (los finalizers son impredecibles y pueden resucitar objetos).
9. Toda llamada externa DEBE tener timeout; limitar pools, colas y buffers.
10. Retry DEBE chequear cancelación del context entre intentos.
11. Usar `strings.Builder` para concatenar en loops.
12. Iterators (Go 1.23+) para lazy evaluation; streaming para transferencias grandes.
13. `crypto/rand` para keys/tokens (`math/rand` es predecible).
14. Regexp DEBE compilarse una vez a nivel paquete.
15. Compile-time interface checks: `var _ Interface = (*Type)(nil)`.

## Restricciones
- No cubre wiring de contenedores DI ni comparación de librerías DI (→ `golang-dependency-injection`).
- No cubre wrapping de errores ni mecánicas de logging (→ `golang-error-handling`).

## Casos límite
- Constructor con 10 parámetros opcionales: functional options evitan breaking changes al agregar opciones.
- Regexp compilado dentro de una función hot: compilación O(n) con alocaciones en cada llamada — mover a package level.
- Transferencia de millones de filas a memoria: OOM — aplicar streaming para mantener la memoria constante.
- Recurso sin límite (pool/cola/buffer): crece hasta el crash — limitar siempre.

## Condiciones de error
- Opción inválida detectada tarde (runtime): la validación debió ocurrir en el constructor con error.
- Panic por condición esperada (ej. input inválido): viola la regla 6 — reemplazar por error retornado.
- Mal uso de `SetFinalizer` resucitando objetos: migrar a `runtime.AddCleanup`.

## Criterios de aceptación
- [ ] Los constructores usan functional options y validan en construcción.
- [ ] No hay `init()` con setup de servicios.
- [ ] Los enums empiezan en 1 o usan sentinel Unknown en 0.
- [ ] Los errores esperados se manejan con early return; el happy path queda plano.
- [ ] No hay llamadas externas sin timeout ni recursos sin límite.
- [ ] Los assets estáticos usan `//go:embed` y los regexp se compilan a nivel paquete.

## Escenarios BDD

### Caso normal
Given un server con 6 opciones de configuración opcionales
When el agente diseña el constructor
Then usa functional options (`WithReadTimeout`, `WithMaxConns`, ...)
And valida opciones inválidas retornando error en construcción

### Caso alternativo
Given un enum de estado con valores 0, 1, 2
When el agente lo define con iota
Then usa `StatusUnknown` en 0
And `StatusReady`/`StatusClosing` a partir de 1 para no colisionar con el zero value

### Caso límite
Given un endpoint que devuelve millones de registros
When el agente implementa la transferencia
Then usa streaming (o iterators) para mantener la memoria constante
And evita cargar todo el resultado en un slice

### Caso de error
Given un panic usado para validar input de usuario en un handler
When el agente revisa el código
Then detecta que el panic es para bugs, no errores esperados
And lo reemplaza por retorno de error con status HTTP adecuado