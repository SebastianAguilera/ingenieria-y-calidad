# Especificación: SK-023 - golang-project-layout: Estructura de proyectos Go

## Objetivo
Definir la estructura y el setup de proyectos Go: convenciones de directorios `cmd/`/`internal/`/`pkg/`, naming de módulos y paquetes, workspaces `go.work`, y archivos de configuración esenciales. Aplica al iniciar un proyecto Go nuevo, organizar un codebase existente, configurar un monorepo multi-paquetes, crear CLIs con múltiples packages main o discutir restructuración/partición de paquetes. El tamaño de la estructura se ajusta al problema: un script se queda plano; un servicio recibe capas solo cuando la complejidad real lo justifica.

## Entradas
- Tipo de proyecto a crear/organizar: CLI, librería, servicio (HTTP API/microservicio/web app), monorepo o workspace multi-módulo.
- Preferencia de arquitectura del desarrollador (se pregunta primero): clean, hexagonal, DDD, flat.
- Enfoque de dependency injection (se pregunta segundo): manual, librería (samber/do, wire, dig+fx) o ninguno.
- Convención 12-Factor App para aplicaciones (config por env vars, logs a stdout, procesos stateless).

## Salidas esperadas
1. Esqueleto de directorios acorde al tipo de proyecto (ver tabla: CLI → `cmd/{name}/`, `internal/`; Librería → `pkg/{name}/`, `internal/`; Servicio → `cmd/{service}/`, `internal/`, `api/`, `web/`; Monorepo → `go.work` + módulos separados).
2. `go.mod` con module path que cumple la convención del skill.
3. Estructura con packages en `cmd/` mínimos (parse flags, wiring, call `Run()`) y lógica de negocio en `internal/` o `pkg/`.
4. Decisiones de arquitectura y DI documentadas (preguntadas y respondidas por el desarrollador).

## Reglas de negocio
1. Preguntar PRIMERO la preferencia de arquitectura al desarrollar (evitar sobre-estructurar proyectos chicos — un CLI de 100 líneas no necesita capas ni DI).
2. Preguntar SEGUNDO el enfoque de DI (manual, librería o ninguno) — afecta wiring, lifecycle y estructura.
3. Module path en `go.mod` DEBE coincidir con la URL del repositorio, minúsculas, con guiones para multi-palabra.
4. Todos los `main` packages DEBEN residir en `cmd/` con lógica mínima; la lógica de negocio en `internal/` (no exportada) o `pkg/` (cuando es útil a consumidores externos).
5. Para aplicaciones (services, APIs, workers) seguir 12-Factor: config por env vars, logs a stdout, procesos stateless, graceful shutdown, backing services como recursos adjuntos, admin tasks como comandos one-off (`cmd/migrate/`).
6. Monorepo → `go.work` con módulos separados por paquete.
7. Los packages DEBEN ser minúsculos, singulares y coincidir con su nombre de directorio.

## Restricciones
- No reestructura código existente sin un cambio de layout (→ skill `golang-refactoring`).
- Los patrones de arquitectura detallados → skill `golang-design-patterns`; la comparación de DI → `golang-dependency-injection`.

## Casos límite
- Script de 100 líneas: estructura flat, sin layers ni DI.
- CLI con múltiples subcomandos: `cmd/{name}/` con un main y packages de comando.
- Monorepo con varios servicios y librerías compartidas: `go.work` y modules por paquete, `internal/` para código privado.
- Librería con API pública: `pkg/{name}/` + `internal/` para privados.

## Condiciones de error
- Module path que no coincide con el repo o usa Mayúsculas/underscores: corregir (ej. `github.com/org/repo`).
- Lógica de negocio dentro de `cmd/`: moverla a `internal/` (el main debe ser delgado).
- Arquitectura asumida sin preguntar al desarrollador: violación de la regla 1 — preguntar siempre.

## Criterios de aceptación
- [ ] Se preguntó la preferencia de arquitectura y de DI antes de generar estructura.
- [ ] El esqueleto corresponde al tipo de proyecto de la tabla.
- [ ] El `go.mod` sigue la convención de module path.
- [ ] Los `main` packages viven en `cmd/` con lógica mínima.
- [ ] La aplicación cumple 12-Factor cuando es un service/API/worker.
- [ ] El monorepo/workspace usa `go.work` adecuadamente.

## Escenarios BDD

### Caso normal
Given un nuevo servicio HTTP API con varios módulos de negocio
When el agente inicializa el proyecto
Then pregunta arquitectura y DI primero
And genera `cmd/service/`, `internal/`, `api/` y configura `go.mod` con el path del repositorio

### Caso alternativo
Given un script CLI de ~100 líneas
When el agente crea la estructura
Then la mantiene flat sin capas de abstracción ni DI
And evita el over-structuring

### Caso límite
Given un monorepo con múltiples servicios y una librería compartida
When el agente organiza el workspace
Then usa `go.work` con un módulo por paquete
And ubica el código compartido interno en `internal/`

### Caso de error
Given un `cmd/api/main.go` con toda la lógica de negocio inline
When el agente revisa el layout
Then detecta que el main no es delgado
And propone mover la lógica a `internal/` con un `Run()` invocable desde main