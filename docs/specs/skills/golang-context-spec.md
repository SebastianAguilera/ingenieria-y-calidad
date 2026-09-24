# Especificación: SK-005 - golang-context: Uso idiomático de context.Context en Go

## Objetivo
Garantizar el uso correcto e idiomático de `context.Context` en Go: propagación a través de los límites de API, cancelación, timeouts y deadlines, valores por request, y `context.WithoutCancel` para trabajo en background que debe sobrevivir al request. Aplica al diseñar propagación de context, depurar contextos filtrados o no expirados, y elegir entre `Background`/`TODO`/`WithoutCancel`.

## Entradas
- Código Go que interactúa con `context.Context` (funciones, handlers HTTP, clientes, operaciones de BD).
- Tarea del operador: diseño de propagación de context, revisión de código, depuración de leaks de context.
- Binarios: `go`.

## Salidas esperadas
1. Código Go que propaga el mismo context a través de toda la cadena de llamadas.
2. Contextos creados con el constructor correcto según la situación (`Background`, `TODO`, `WithCancel`, `WithTimeout`, `WithDeadline`).
3. `cancel()` llamado en todos los caminos de control (o ownership transferido explícitamente).
4. Keys de valores de context declarados como tipos no exportados.
5. Uso de `WithoutCancel` para trabajo en background que debe outvivir al request (audit logs, cleanup).

## Reglas de negocio
1. Propagar el mismo context por todo el ciclo del request: handler HTTP → service → BD → APIs externas.
2. `ctx` como PRIMER parámetro de función, nombrado `ctx context.Context`.
3. Pasar context por parámetros, NUNCA almacenarlo en un struct (el struct outvive al request).
4. Pasar `context.TODO()` en lugar de un `nil` context (nil panics en el primer `Done()` o `Value()`).
5. Llamar `cancel()` en todos los caminos de control de `WithCancel/WithTimeout/WithDeadline`, salvo que se transfiera el ownership explícitamente (cancel sin llamar filtra el timer del hijo).
6. Crear `context.Background()` solo en puntos de entrada top-level (main, init, tests).
7. `context.TODO()` como placeholder cuando falta un context, marcando la deuda.
8. Keys de valores de context: tipos no exportados (un string plano permite colisiones entre paquetes).
9. Solo metadata scoped al request en valores de context; nunca parámetros de función.
10. `context.WithoutCancel` (Go 1.21+) para trabajo en background que debe outvivir al request padre.

## Restricciones
- No aplica a código que acepta `ctx` solo "de adorno" sin usarlo en la cadena de llamadas.
- El contexto de cancelación de goroutines se complementa con el skill `golang-concurrency`.

## Casos límite
- Función que rompe la cadena con `context.Background()` interno: se detecta como bug de propagación (el trabajo sigue tras la cancelación del cliente).
- Operación en background (audit log) que debe completarse post-request: sin `WithoutCancel`, el handler al retornar cancela el trabajo.
- Llamada a API externa sin timeout: una upstream lenta cuelga la goroutine indefinidamente — aplicar `WithTimeout`.

## Condiciones de error
- `nil` context pasado como argumento: panic lejano al llamador — reemplazar por `TODO()` o `Background()`.
- `cancel()` no invocado: timer filtrado y child pegado al parent — corregir invocando cancel en todos los caminos.
- Colisión de keys de context con string plano: refactorizar a tipo no exportado.

## Criterios de aceptación
- [ ] Todo el código propaga el context del llamador, sin `Background()` intermedio en la cadena.
- [ ] `ctx` es el primer parámetro con el nombre `ctx context.Context`.
- [ ] No hay contextos almacenados en structs.
- [ ] Todos los `With*` invocan `cancel()` en todos los caminos o transfieren ownership.
- [ ] `Background()` solo aparece en entry points.
- [ ] Los valores de context usan keys de tipo no exportado.
- [ ] El trabajo en background post-request usa `WithoutCancel`.

## Escenarios BDD

### Caso normal
Given un handler HTTP que llama a un service que llama a la BD
When el agente implementa la propagación
Then el mismo `r.Context()` viaja handler → service → `QueryContext`
And cancelar el request cancela automáticamente la consulta

### Caso alternativo
Given una función que necesita un context pero el caller aún no lo provee
When el agente escribe la firma
Then usa `context.TODO()` como placeholder explícito
And documenta la deuda para reemplazarlo luego

### Caso límite
Given la escritura de un audit log que debe completarse aunque el request termine
When el agente diseña el flujo
Then usa `context.WithoutCancel` para que el trabajo background outviva al parent

### Caso de error
Given un service que crea `context.Background()` para una consulta de BD interna
When el agente revisa el código
Then detecta el quiebre de propagación (la BD sigue trabajando tras cancelarse el cliente)
And lo corrige propagando el context del llamador