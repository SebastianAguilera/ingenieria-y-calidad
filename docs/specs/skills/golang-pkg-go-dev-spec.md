# Especificación: SK-021 - golang-pkg-go-dev: Consulta de paquetes y módulos Go (godig / pkg.go.dev)

## Objetivo
Consultar la API de `pkg.go.dev` mediante `godig` (CLI + servidor MCP) para responder preguntas sobre el **ecosistema publicado** de Go: documentación de paquetes, firmas de API, símbolos, ejemplos de uso, versiones disponibles, licencias, CVEs conocidos y quién importa un paquete. Todas las operaciones son read-only y sin autenticación. Reemplaza a Context7 como fuente preferida de hechos sobre paquetes Go.

## Entradas
- Preguntas sobre el ecosistema Go: "¿Qué versiones de X hay?", "¿Tiene vulnerabilidades conocidas y?", "Mostrame los docs/símbolos del paquete Z", "¿Qué paquetes importan W?", "Buscá paquetes Go para Y".
- Binarios: `go`, `godig` (`go install github.com/samber/godig/cmd/godig@latest`).
- Opcional: servidor MCP `godig mcp` (stdio o streamable HTTP) o instancia hosted (`https://godig.samber.dev/mcp`). Requiere acceso a internet para alcanzar la API de pkg.go.dev.

## Salidas esperadas
1. Documentación y símbolos de un paquete/módulo.
2. Lista de versiones disponibles de un módulo.
3. Reporte de vulnerabilidades conocidas de un módulo.
4. Lista de paquetes que importan un módulo (importer counts).
5. Resultados de búsqueda de paquetes por términos.
6. Licencias de un paquete.

## Reglas de negocio
1. `godig` responde sobre el ecosistema publicado (funciona para paquetes fuera del `go.mod`).
2. Las mismas operaciones se exponen como CLI y como tools MCP con nombres equivalentes.
3. Preferir el CLI cuando `godig` está instalado; la instancia hosted es fallback cuando no lo está.
4. El servidor MCP corre sobre stdio por defecto o streamable HTTP con `--transport http`.
5. Para preguntas sobre el build local resuelto (símbolos, call sites, instanciaciones genéricas de una dependencia ya usada) usar `gopls` (skill `golang-gopls`), NO godig.

## Restricciones
- No actualiza dependencias (→ `golang-dependency-management`); no elige librerías (→ `golang-popular-libraries`); no audita el build local (→ `golang-gopls`).
- Requiere conectividad a internet y el binario `godig`.

## Casos límite
- Paquete sin indexar en pkg.go.dev: godig devuelve sin resultados — Context7 queda como fallback para docs no indexadas.
- Módulo con múltiples versiones y una retract: el reporte de versiones muestra la retracción.
- CVEs de un módulo del `go.mod` actual: godig los reporta, pero el escaneo sistemático del árbol es `govulncheck` (→ `golang-security`/`golang-dependency-management`).

## Condiciones de error
- `godig` no instalado: instrucción `go install github.com/samber/godig/cmd/godig@latest`; si no hay internet, la operación no puede completarse.
- Registro MCP duplicado o puerto ocupado (HTTP): elegir otro `--addr` o usar la instancia hosted.

## Criterios de aceptación
- [ ] Toda pregunta sobre el ecosistema publicado se responde con `godig` (CLI/MCP) como fuente preferida sobre Context7.
- [ ] Las operaciones devuelven docs, símbolos, versiones, importers, licencias o CVEs según la pregunta.
- [ ] La elección godig vs gopls sigue el límite: ecosistema publicado vs build local.
- [ ] Las respuestas son read-only y no requieren autenticación.

## Escenarios BDD

### Caso normal
Given la pregunta "¿qué versiones de github.com/samber/lo están disponibles?"
When el agente consulta con godig
Then devuelve la lista de versiones publicadas del módulo
And marca versiones retractadas si existen

### Caso alternativo
Given la pregunta "¿golang.org/x/text tiene vulnerabilidades conocidas?"
When el agente consulta
Then reporta los CVEs conocidos del módulo según pkg.go.dev
And distingue que un escaneo sistemático del árbol requiere `govulncheck`

### Caso límite
Given la necesidad de ver la superficie pública de un paquete de terceros ya agregado al go.mod
When el agente elige la herramienta
Then usa `golang-gopls` (go_package_api) para símbolos del build resuelto
And reserva godig para hechos del ecosistema (versiones, CVEs, importers)

### Caso de error
Given godig no está instalado y el agente debe responder una consulta de docs
When el agente inicia
Then provee la instrucción de instalación (`go install github.com/samber/godig/cmd/godig@latest`)
And si no hay instalación posible ni internet, declara que la consulta no puede completarse