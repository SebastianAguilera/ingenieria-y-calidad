# Especificación: SK-014 - golang-gopls: Inteligencia semántica de código Go con gopls

## Objetivo
Usar `gopls`, el language server oficial de Go, para obtener inteligencia semántica del código: go-to-definition, find references, call/implementation hierarchy, search de símbolos en el workspace, descubrimiento de API de paquetes, diagnostics, rename seguro, refactors (extract/inline/fill/rewrite), formato y generación de tests. El skill responde preguntas sobre el build local resuelto del proyecto; se accede vía el servidor MCP de gopls, la herramienta nativa `LSP` o el CLI `gopls`.

## Entradas
- Proyecto Go con build local resuelto (workspace + dependencias según `go.sum`, incluyendo `replace`).
- Tarea del operador: navegar código, buscar referencias, renombrar, diagnosticar, refactorizar, formatear.
- Binarios: `go`, `gopls` (v0.20+). Opcional: registro del MCP server y plugin `gopls-lsp@claude-plugins-official` para la herramienta `LSP`.

## Salidas esperadas
1. Respuestas semánticas: definiciones, referencias, jerarquías de llamadas/implementaciones, símbolos del workspace, superficie pública de paquetes.
2. Diagnostics de compilación/análisis por archivo modificado (automáticos vía `LSP`, on-demand vía `go_diagnostics`/CLI).
3. Renombres seguros que rechazan cambios que romperían satisfacción de interfaces o shadowing.
4. Refactors guiados: extract/inline y familia `refactor.rewrite.*`.
5. Formato canónico (equivalente a `gofmt`) y organización de imports.
6. Verify de alcanzabilidad de vulnerabilidades (`go_vulncheck`) en el workspace (baseline y tras cambios de `go.mod`).

## Reglas de negocio
1. Preferencia de acceso: **MCP → LSP nativo → CLI** (MCP piensa en nombres/rutas; LSP agrega diagnostics automáticos; CLI es fallback documentado).
2. `gopls` solo responde sobre el build local resuelto — para hechos publicados del ecosistema (versiones, licencias, CVEs de algo no añadido) usar `godig` (skill `golang-pkg-go-dev`).
3. Session start: llamar `go_workspace` una vez; si es workspace Go, seguir con `go_vulncheck` baseline.
4. Workflow de lectura: `go_workspace` → `go_search` → `go_file_context` → `go_package_api`.
5. Workflow de edición: leer primero → `go_symbol_references` antes de modificar definiciones (blast radius) → realizar todas las ediciones planeadas → `go_diagnostics` obligatorio en cada archivo cambiado → corregir → re-ejecutar diagnostics → si cambió `go.mod`, `go_vulncheck` del workspace → `go test` solo sobre los paquetes cambiados.
6. Los quick-fixes sugeridos se revisan antes de aplicarse; los diagnostics hint/info no relacionados con la tarea se ignoran.
7. Un mensaje de diagnostic puede parafrasear el código en lugar de citarlo textualmente.

## Restricciones
- No responde sobre el ecosistema publicado (paquetes fuera del `go.mod`, versiones, docentes, importer counts) → skill `golang-pkg-go-dev`.
- No hace auditoría de vulnerabilidades de todo el árbol a nivel release → skill `golang-security`.

## Casos límite
- Rename que rompería una interface: gopls lo rechaza en lugar de producir un diff roto.
- Regla de inline con argumentos con side effects: `Inline` los sustituye en temporales `var` en lugar de duplicarlos.
- Código con shadowing: gopls señala/previene el rename seguro.
- Workspace sin módulo (GOPATH mode): `go_workspace` lo detecta y ajusta el flujo.

## Condiciones de error
- gopls no instalado o versión < v0.20: la herramienta falla al iniciar — instalar `golang.org/x/tools/gopls@latest`.
- `go_vulncheck` con hallazgos tras cambio de go.mod: diagnosticar antes de continuar el flujo de edición.
- Diagnostics de compilación tras las ediciones: bloqueante — ningún edit adicional sin corregir los errores reportados.

## Criterios de aceptación
- [ ] Se usa gopls (MCP/LSP/CLI) para navegación y refactors semánticos en lugar de grep/texto.
- [ ] El workflow de session start detecta el workspace y corre el vulncheck baseline.
- [ ] Antes de renombrar/modificar definiciones se inspeccionan las referencias (blast radius).
- [ ] Todo archivo modificado pasa por diagnostics antes de continuar.
- [ ] Los renames se ejecutan a través de gopls (seguros) en lugar de ediciones manuales masivas.
- [ ] Después de cambios de `go.mod` se ejecuta `go_vulncheck` sobre el workspace.

## Escenarios BDD

### Caso normal
Given un agente que necesita renombrar `ParseToken` a `DecodeToken`
When ejecuta el rename con gopls
Then actualiza todos los call sites del build resuelto
And rechaza el cambio si rompe satisfacción de una interface

### Caso alternativo
Given una duda sobre qué paquete importa un símbolo y su alcance
When el agente necesita contexto del archivo
Then usa `go_file_context` y `go_search` para localizar el símbolo
And evita leer archivos innecesarios

### Caso límite
Given una función que se quiere inlinear con argumentos que tienen efectos secundarios
When gopls aplica el refactor
Then introduce variables temporales para preservar la semántica
And mantiene el comportamiento idéntico

### Caso de error
Given un archivo modificado que quedó con error de compilación
When el agente continúa el flujo de edición
Then `go_diagnostics` reporta el error y bloquea el avance
And el agente corrige o revierte antes de seguir