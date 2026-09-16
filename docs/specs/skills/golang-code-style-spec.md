# Especificación: SK-003 - golang-code-style: Estilo de código Go

## Objetivo
Definir y aplicar las reglas de estilo de código Go que requieren juicio humano (claridad, estructura, legibilidad), complementando lo que los linters automatizan (formato). Sirve para escribir y revisar código Go con un estándar de claridad consistente ("Clear is better than clever").

## Entradas
- Código Go a escribir o revisar (`**/*.go`).
- Tarea del operador: pedido de estilo, review de claridad o establecimiento de estándares del proyecto.
- Binarios: `go`, `golangci-lint` (opcional).
- Herramientas permitidas: Read, Edit, Write, Glob, Grep, Bash (go, golangci-lint, git), agentes.

## Salidas esperadas
1. Código Go que cumple las reglas de estilo (o recomendaciones de cambio en reviews).
2. Código con quiebre de líneas en límites semánticos (no columnas arbitrarias).
3. Declaraciones de variables idiomáticas (`:=` para valores no cero, `var` para cero).
4. Slices y maps inicializados explícitamente (nunca nil).
5. Composite literals con nombres de campo.
6. Comentarios (no requeridos, pero cuando se ignora una regla se documenta el motivo).

## Reglas de negocio
1. Líneas de más de ~120 caracteres DEBEN quiebrarse en **límites semánticos**, no por conteo de columnas.
2. Llamadas a funciones con 4+ argumentos DEBEN usar un argumento por línea.
3. Signaturas demasiado largas: la corrección real suele ser menos parámetros (options struct), no mejor wrapping.
4. `:=` para valores no cero; `var` para inicialización en cero (la forma señala intención).
5. Slices y maps DEBEN inicializarse explícitamente (`[]T{}`, `make(...)`); nunca nil. Los nil maps panic en escritura y los nil slices serializan a `null` en JSON.
6. No preasignar especulativamente (`make([]T, 0, 1000)` desperdicia memoria si el caso común es 10 items).
7. Composite literals DEBEN usar nombres de campo (los literales posicionales se rompen al agregar/reordenar campos).
8. Al ignorar una regla, se DEBE agregar un comentario explicando el motivo.

## Restricciones
- No cubre convenciones de nomenclatura (skill `golang-naming`), configuración de linters (skill `golang-lint`) ni doc comments (skill `golang-documentation`).
- Solo aplica a archivos Go del proyecto; no modifica infraestructura ni configuración.

## Casos límite
- Línea cercana al límite de 120 caracteres sin sobrepasar: no se fuerza el quiebre si el límite semántico no lo justifica.
- Función estructurada naturalmente en una sola línea con 4 argumentos cortos: aún así se aplica un argumento por línea (regla MUST).
- Reutilización de código existente con estilo legacy: el review marca la deuda pero no bloquea el cambio si es ajeno al diff.

## Condiciones de error
- Conflicto entre una regla de estilo y una restricción de formato del linter: prevalece la configuración del linter para mecanismos, el skill para claridad.
- Regla ignorada sin comentario: se solicita el comentario justificativo en el review.

## Criterios de aceptación
- [ ] Todo el código nuevo respeta el límite de longitud y los quiebres semánticos.
- [ ] Las llamadas con 4+ argumentos usan un argumento por línea.
- [ ] Las declaraciones distinguen `:=` (no cero) de `var` (cero).
- [ ] Slices y maps están inicializados explícitamente.
- [ ] Los composite literals usan nombres de campo.
- [ ] Cada omisión de una regla está documentada con un comentario.

## Escenarios BDD

### Caso normal
Given código Go nuevo con una llamada de 5 argumentos
When el agente lo escribe o revisa
Then los argumentos se colocan uno por línea con el paréntesis de cierre separado

### Caso alternativo
Given una función con una lista larga de parámetros
When el agente aplica el skill
Then propone reducir parámetros con un options struct en lugar de solo mejorar el wrapping

### Caso límite
Given una línea de 121 caracteres que rompe en el último token
When el agente decide el quiebre
Then lo realiza en el límite semántico más cercano, no en una columna arbitraria

### Caso de error
Given un map asignado como nil en una estructura que luego se escribe
When el agente revisa el código
Then detecta el riesgo de panic y el serializado a JSON inconsistente
And corrige inicializando el map explícitamente