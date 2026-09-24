# Especificación: SK-013 - golang-error-handling: Manejo idiomático de errores en Go

## Objetivo
Definir la creación, wrapping e inspección de errores en Go de forma idiomática y robusta: creación de errores, wrapping con `%w`, `errors.Is`/`errors.As`/`errors.AsType`, `errors.Join`, tipos de error custom, sentinel errors, la regla de manejo único, panic/recover, logging estructurado con `slog` y `samber/oops` para errores de producción. Cada error es un evento que se maneja o se propaga con contexto: fallas silenciosas y logs duplicados son igualmente inaceptables.

## Entradas
- Código Go que crea, propaga, inspecciona o loguea errores (`**/*.go`).
- Modo: **Coding** (escribir código de errores en orden secuencial; opcional: sub-agente de fondo grep de violaciones adyacentes), **Review** (diff de PR: errores tragados, wrapping sin contexto, log-and-return, panic misuse) o **Audit** (hasta 5 sub-agentes por categoría: creación, wrapping, regla de manejo único, panic/recover, logging estructurado).
- Binarios: `go`.

## Salidas esperadas
1. Errores creados con `errors.New`/`fmt.Errorf`; mensajes en minúscula, sin puntuación final.
2. Errores wrapped con contexto (`fmt.Errorf("contexto: %w", err)`) en cada capa.
3. Inspección con `errors.Is` (sentinel) y `errors.As`/`errors.AsType` (tipos), nunca con `==` ni type assertion pelada.
4. Errores independientes combinados con `errors.Join` (Go 1.20+).
5. Cada error logueado O retornado, nunca ambos (single handling rule).
6. Sentinels para condiciones esperadas; custom types para transportar datos.
7. Sin `panic` para condiciones esperadas.
8. Logging estructurado con `slog` (Go 1.21+), niveles de severidad, middleware de logging HTTP.

## Reglas de negocio
1. Los errores retornados DEBEN chequearse siempre — NUNCA descartarse con `_`.
2. Los errores DEBEN wrappearse con contexto: `fmt.Errorf("contexto: %w", err)`.
3. Los strings de error DEBEN ir en minúscula y sin puntuación final.
4. `%w` internamente, `%v` en las fronteras del sistema (controlar la cadena expuesta).
5. DEBE usarse `errors.Is` para sentinels y `errors.As`/`errors.AsType` para inspección tipada (Go 1.26+: `errors.AsType[T](err)`).
6. Se DEBE usar `errors.Join` (Go 1.20+) para combinar errores independientes.
7. Los errores DEBEN loguearse O retornarse, NUNCA ambos (regla de manejo único; previene logs duplicados).
8. Sentinel errors para condiciones esperadas; custom types para transportar datos.
9. NUNCA `panic` para condiciones de error esperadas — reservado para estados irrecoverables.
10. Se DEBE usar `slog` para logging estructurado, no `fmt.Println` ni `log.Printf`.
11. NUNCA exponer errores técnicos al usuario — traducir a mensajes amigables y loguear los detalles técnicos por separado.
12. Mantener baja cardenalidad en los mensajes de log estables; adjuntar IDs/paths/lineas como atributos estructurados.

## Restricciones
- Errores sin explotar defensivos (nil interface trap, comparaciones nil) → skill `golang-safety`.
- Alertas/APM completos → skill `golang-observability`.

## Casos límite
- Error envuelto varias capas con `%w` y comparado con `==`: falla — usar `errors.Is`.
- Error con datos ricos (user/tenant, stack): custom type o `samber/oops` en producción.
- Dos errores independientes en un flujo (validaciones múltiples): `errors.Join`.
- Error en goroutine con panic: recovery en el boundary de la goroutine, no en el proceso.

## Condiciones de error
- Log + return del mismo error: entrada duplicada en el agregador — aplicar single handling rule.
- Error descartado con `_`: failure silenciosa — reemplazar por manejo explícito o propagación.
- Mensaje con mayúscula inicial o puntuación final: estilo inconsistente y rompe agregación — corregir.
- Violación PII en mensajes de error: no loguear datos personales; adjuntar IDs de correlación, no el contenido.

## Criterios de aceptación
- [ ] Ningún error se descarta con `_`.
- [ ] Todo error retornado lleva contexto de wrapping con `%w`.
- [ ] La inspección usa `errors.Is`/`errors.As`/`errors.AsType`.
- [ ] No hay pares log-and-return del mismo error.
- [ ] No hay `panic` para condiciones esperadas.
- [ ] El logging usa `slog` estructurado con niveles y sin PII.

## Escenarios BDD

### Caso normal
Given una capa de servicio que llama al repository
When ocurre un error de BD
Then el servicio lo wrappea con `fmt.Errorf("obtener usuario: %w", err)`
And el handler traduce el error a un mensaje amigable y un status HTTP apropiado

### Caso alternativo
Given múltiples validaciones independientes que pueden fallar simultáneamente
When el agente las combina
Then usa `errors.Join` para reportar los tres errores juntos
And el caller puede inspeccionarlos con `errors.As`/`errors.Is` individualmente

### Caso límite
Given un error esperado (usuario no encontrado) modelado como sentinel
When el caller lo compara
Then usa `errors.Is(err, ErrUserNotFound)`
And distingue el no-encontrado de otros errores reales

### Caso de error
Given un handler que loguea el error técnico Y lo retorna con detalle técnico al cliente
When el agente audita
Then detecta el log-and-return y la exposición de detalle técnico
And corrige logueando el detalle y devolviendo un mensaje genérico al cliente