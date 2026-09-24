# Especificación: SK-016 - golang-lint: Linting con golangci-lint

## Objetivo
Definir las mejores prácticas de linting y la configuración de `golangci-lint` para proyectos Go: ejecución, configuración de `.golangci.yml`, supresión de warnings con directivas `//nolint`, interpretación de salida y selección de linters. El linting es parte central del workflow de desarrollo, no un paso de limpieza al final.

## Entradas
- Proyecto Go con `**/*.go` y opcionalmente `.golangci.yml`.
- Modo: **Setup** (configurar `.golangci.yml`, elegir linters, habilitar CI), **Coding** (sub-agente de fondo con `golangci-lint run --fix` sobre archivos modificados mientras el main continúa) o **Interpret/fix** (leer salida, suprimir warnings, corregir issue sobre código existente).
- Binarios: `go`, `golangci-lint` (`go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`).

## Salidas esperadas
1. `.golangci.yml` como source of truth de linters habilitados y su configuración (referencia recomendada con 33+ linters del skill).
2. Ejecuciones de `golangci-lint run ./...` limpias (o con hallazgos clasificados).
3. Correcciones automáticas con `golangci-lint run --fix` cuando es seguro.
4. Directivas `//nolint` específicas con justificación (nunca genéricas sin motivo).
5. Interpretación de la salida: errores, warnings y supresiones clasificadas.

## Reglas de negocio
1. Todo proyecto Go DEBE tener `.golangci.yml` — es el source of truth de linters y su configuración.
2. Ejecutar `golangci-lint run ./...` frecuentemente y siempre en CI.
3. Usar `//nolint` con moderación — primero corregir la causa raíz; las directivas deben ser específicas (nombre de linter) y justificadas.
4. Para adoptar linting en un codebase legacy, paralelizar la limpieza por categorías de linters (auto-fix, seguridad, error handling, estilo/formato, calidad) con sub-agentes.
5. En Coding mode, el sub-agente de fondo corre `--fix` sobre los archivos modificados mientras el agente principal continúa la feature; los resultados se comunican al terminar.
6. `golangci-lint fmt` (v2+) para formateo; `--enable-only` para correr un linter individual.

## Restricciones
- No cablea el paso de lint en un pipeline de GitHub Actions (→ skill `golang-continuous-integration`).

## Casos límite
- Codebase legacy sin linting con cientos de issues: limpieza paralela por categorías, sin bloquear el desarrollo nuevo.
- Warning falso positivo de un linter sobre patrón intencional: `//nolint:<linter>` específico con justificación en la línea.
- Un linter nuevo en una configuración existente: agregarlo sin romper la build de CI (correr primero localmente con `--enable-only`).

## Condiciones de error
- `golangci-lint` no instalado: el flujo se pausa con la instrucción de instalación.
- Issue de seguridad (gosec) en el código: bloqueante, se corrige antes del merge; no se suprime con `//nolint` salvo justificación de riesgo aceptado documentado.

## Criterios de aceptación
- [ ] El proyecto tiene `.golangci.yml` versionado como source of truth.
- [ ] `golangci-lint run ./...` pasa en CI.
- [ ] No hay directivas `//nolint` sin linter específico ni justificación.
- [ ] Los fixes automáticos se aplican solo sobre archivos del alcance solicitado.
- [ ] La salida de lint se interpreta y clasifica (correctness, style, complexity, performance, security).

## Escenarios BDD

### Caso normal
Given un proyecto Go sin `.golangci.yml`
When el agente ejecuta Setup mode
Then crea `.golangci.yml` siguiendo la configuración recomendada del skill
And define los linters habilitados y su configuración

### Caso alternativo
Given el agente implementando una feature mientras el código cambia
When arranca el sub-agente de fondo
Then corre `golangci-lint run --fix` solo sobre los archivos modificados
And reporta los resultados al terminar la feature

### Caso límite
Given un linter que marca una línea con un patrón intencional del equipo
When el warning debe suprimirse
Then se usa `//nolint:<linter>` específico en esa línea
And se agrega la justificación del porqué se ignora

### Caso de error
Given `gosec` reporta una vulnerabilidad de inyección en una query SQL
When el agente interpreta la salida
Then la clasifica como issue de seguridad crítico
And no la suprime con `//nolint`
And corrige la query con placeholders parametrizados