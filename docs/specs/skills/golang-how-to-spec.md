# Especificación: SK-015 - golang-how-to: Orquestador de skills de Go

## Objetivo
Actuar como orquestador de la colección de skills Go del proyecto: para cualquier tarea de coding, review, debug o setup en Go, identifica el/los skill(s) primario(s) y secundarios aplicables y los carga juntos, porque una tarea rara vez pertenece a un solo skill. Además desambigua clusters solapados (performance vs benchmark vs troubleshooting, DI, safety vs security) y configura el archivo de agent-config del proyecto para forzar la activación de skills.

## Entradas
- Cualquier tarea Go: escribir código, revisar, debuggear, configurar herramientas o CI.
- Tabla de mapeo intent → primary skill → secondary skills del propio skill.
- Modo: **Orchestrate** (cargar primary + secundarios juntos desde el inicio), **Disambiguate** (mostrar boundary table entre skills que se solapan) o **Configure** (escribir directiva always-load + bloque `## Required Go skills` en el agent-config).
- Binarios: `go`, `gopls`, `git`.

## Salidas esperadas
1. Determinación del skill primario y los secundarios aplicables para la tarea (cargados en el mismo paso, sin esperar).
2. Tabla de límites entre skills solapados cuando el operador pide desambiguación.
3. Archivo de configuración del agente (AGENTS.md/CLAUDE.md/GEMINI.md/Cursor rules/Copilot instructions) actualizado con la directiva always-load del orquestador y el bloque de skills requeridos (Configure mode).
4. Respuestas a preguntas de configuración a través de la herramienta de preguntas del entorno (una a la vez).

## Reglas de negocio
1. Para cada tarea, cargar el skill primario Y todos los secundarios aplicables en el mismo momento (no esperar).
2. La asignación sigue la tabla del skill (ej: escribir goroutines → primario `golang-concurrency` + secundario `golang-context` si hay cancelación).
3. En Configure mode, preguntar al usuario mediante la herramienta de preguntas del entorno — nunca en texto plano; una pregunta a la vez, esperando la respuesta.
4. Al configurar, escribir la directiva always-load para `golang-how-to` y opcionalmente un bloque `## Required Go skills` en el agent-config del proyecto.
5. Los identificadores de skill son formas cortas de `samber/cc-skills-golang@<nombre>`.

## Restricciones
- No reemplaza el contenido de los skills individuales: los selecciona y los carga.
- Los skills referenciados pero ausentes en esta instalación (golang-grpc, golang-samber-lo, etc.) se ignoran sin romper la orquestación.

## Casos límite
- Tarea que cruza dos skills con solapamiento aparente (pánico en código concurrente): primario `golang-troubleshooting`, secundarios `golang-safety` y `golang-concurrency`.
- Tarea de tests con testify ya en el código: primario `golang-testing` + secundario `golang-stretchr-testify`.
- Tarea de optimización sin medición previa: primario `golang-benchmark` (medir) y se recuerda que la optimización es `golang-performance`.

## Condiciones de error
- Si la environment no tiene herramienta de preguntas, se formulan las preguntas en prosa con las mismas opciones.
- Configuración en un archivo de agent-config inexistente: el skill lo crea o indica el archivo válido según la convención del proyecto.

## Criterios de aceptación
- [ ] Toda tarea Go inicia cargando el skill primario y sus secundarios juntos.
- [ ] Las desambiguaciones entre skills solapados devuelven una tabla de límites.
- [ ] Configure mode escribe la directiva always-load y el bloque de skills requeridos en el agent-config.
- [ ] Las preguntas de configuración usan la herramienta de preguntas del entorno (o prosa con las mismas opciones si no existe).

## Escenarios BDD

### Caso normal
Given el operador pide "escribí un endpoint con goroutines y cancelación"
When el agente orquesta
Then carga `golang-concurrency` como primario y `golang-context` como secundario
And también carga `golang-error-handling` si el código produce errores

### Caso alternativo
Given el operador pregunta la diferencia entre `golang-safety` y `golang-security`
When el agente desambigua
Then muestra la boundary table indicando que safety cubre bugs no adversariales y security cubre atacantes

### Caso límite
Given una tarea de optimización de un hot path ya perfilado con pprof
When el agente orquesta
Then carga `golang-performance` como primario
And referencia `golang-benchmark` para la verificación de la mejora

### Caso de error
Given una tarea de debug de un panic en código con goroutines
When el agente elige los skills
Then no asume root cause con el skill de optimización
And carga `golang-troubleshooting` + `golang-safety` (+ `golang-concurrency` para el contexto)
And prioriza la investigación de causa raíz antes que cualquier fix