# Especificación: SK-025 - golang-safety: Programación defensiva en Go

## Objetivo
Prevenir errores de programación en Go — bugs, panics y corrupción silenciosa de datos en código normal (no adversarial): nil panics, interfaces con nil tipado, aliasing del backing array en `append`, truncamiento silencioso int64→int32, comparación de floats con `==`, `defer` dentro de loops, copias defensivas de slices/maps y zero values útiles. "Security maneja atacantes; safety nos manejamos a nosotros mismos."

## Entradas
- Código Go a escribir o revisar (`**/*.go`) con foco en correctitud defensiva.
- Tarea del operador: corregir panics (nil map y deep), revisar código para nil-safety, overflow de conversiones numéricas o lifecycle de recursos, o diseñar tipos con zero value seguro.
- Binarios: `go`.

## Salidas esperadas
1. Código con safe type assertions (comma-ok; `reflect.TypeAssert[T]` en Go 1.25+ para reflection).
2. Manejo correcto de interfaces con nil tipado (devolver `nil` explícito, no puntero tipado nil).
3. Maps inicializados antes de escribir; slices con copias defensivas al exportar.
4. `defer` fuera de loops (loop body extraído a función).
5. Conversiones numéricas con chequeo de rango (o uso de `math/bits`/`strconv` apropiados).
6. Comparaciones float con epsilon o `math/big`.
7. Zero values diseñados como útiles (lazy init con `sync.Once`).

## Reglas de negocio
1. Preferir generics sobre `any` cuando el type set se conoce — el compilador atrapa mismatches en vez de panics en runtime.
2. SIEMPRE usar safe type assertions: comma-ok (`v, ok := x.(T)`); en reflexión (Go 1.25+) preferir `reflect.TypeAssert[T](value)` sobre `value.Interface().(T)`.
3. Un puntero tipado nil dentro de una interface NO es `== nil` — el type descriptor lo hace no-nil; devolver `nil` explícito.
4. Escribir en un map nil panic — inicializar siempre antes de usar.
5. `append` puede reusar el backing array — dos slices comparten memoria si la capacidad alcanza, corrompiéndose silenciosamente.
6. Devolver **copias defensivas** desde funciones exportadas — si no, los callers mutan tus internals.
7. `defer` corre al salir de la función, no por iteración del loop — extraer el cuerpo del loop a una función.
8. Las conversiones de enteros truncán silenciosamente — `int64` a `int32` wrappea sin error.
9. La aritmética float no es exacta — usar comparación con epsilon o `math/big`.
10. Diseñar zero values útiles — campos de map nil panic en el primer write; usar lazy init.
11. Usar `sync.Once` para lazy init — garantiza exactly-once incluso bajo concurrencia.

## Restricciones
- Diseño de acceso concurrente con goroutines/channels/sync → skill `golang-concurrency`.
- Vulnerabilidades explotables (injection, crypto débil, secrets) → skill `golang-security`.
- Debug de un programa ya fallando → skill `golang-troubleshooting`.

## Casos límite
- Función que retorna `(*MyHandler, error)` y devuelve puntero nil: la interface resultante no es nil — devolver nil explícito o documentar contrato.
- Slice de retorno compartido con el buffer interno: el caller modifica el internal — copia defensiva.
- Conversión de un ID de BD int64 a int32: truncamiento silencioso con montos/seeds grandes — validar rango.
- Comparación de precios en float: `==` inestable — usar epsilon o decimal exacto.

## Condiciones de error
- Panic por mapa nil en el primer write: inicializar `make`/literal; revisar el zero value del struct.
- Tipo de assertion sin comma-ok que panic en runtime: reemplazar por `v, ok := x.(T)`.
- Corrupción silenciosa por backing array compartido: clonar con `slices.Clone` o copia explícita.

## Criterios de aceptación
- [ ] No hay type assertions sin la forma comma-ok (salvo en código con invariantes probadas y comentadas).
- [ ] Las interfaces no devuelven punteros tipados nil.
- [ ] Los maps se inicializan antes de escribir.
- [ ] Los retornos exportados entregan copias defensivas cuando el internal podría mutarse.
- [ ] No hay `defer` de cierre de recursos dentro de loops.
- [ ] Las conversiones numéricas potencialmente truncantes están validadas.
- [ ] Los floats no se comparan con `==`.

## Escenarios BDD

### Caso normal
Given una función exportada que devuelve un slice del estado interno
When el agente aplica programación defensiva
Then retorna una copia (`append([]T(nil), s...)` o `slices.Clone`)
And evita que el caller mute los internals

### Caso alternativo
Given un struct con un campo map que se inicializa lazy
When el agente lo diseña
Then provee un zero value útil con inicialización lazy vía `sync.Once`
And garantiza exactly-once bajo uso concurrente

### Caso límite
Given la conversión de un `int64` de un ID de BD a `int32`
When hay riesgo de truncamiento silencioso
Then el agente agrega validación de rango antes de convertir
And documenta el caso (o usa un tipo 64-bit de punta a punta)

### Caso de error
Given un mapa nil que panics al primer write en un flujo poco ejercitado
When el agente revisa el código
Then detecta el map no inicializado
And lo inicializa explícitamente, corrigiendo también el zero value del struct si aplica