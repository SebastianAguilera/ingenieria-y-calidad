# Especificación: SK-026 - golang-security: Seguridad en código Go

## Objetivo
Aplicar mejores prácticas de seguridad y prevención de vulnerabilidades en Go: injection (SQL, comando, XSS), criptografía, path traversal, SSRF y headers HTTP, cookies, gestión de secrets, memoria, PII en logs, threat modeling STRIDE/DREAD, más `gosec` (SAST), race detection y fuzz testing. Se aplica al escribir, revisar o auditar código Go, o al tocar crypto, I/O de archivos/red, secrets, input de usuario o autenticación. Principio de defensa en profundidad: proteger en múltiples capas, validar todos los inputs, usar defaults seguros.

## Entradas
- Código Go a escribir/revisar/auditar (`**/*.go`) que cruza trust boundaries o maneja datos sensibles.
- Modo: **Review** (PR: desde los archivos cambiados, trazar call sites y flujos de datos hacia código adyacente), **Audit** (codebase completo: hasta 5 sub-agentes por dominio — injection, criptografía/secrets, web security/headers, auth/authz, concurrency safety + dependency vulnerabilities; hallazgos puntuados con DREAD) o **Coding** (escribir código nuevo o fixear una vulnerabilidad reportada).
- Binarios: `go`, `govulncheck` (`go install golang.org/x/vuln/cmd/govulncheck@latest`).

## Salidas esperadas
1. Código Go sin las vulnerabilidades cubiertas (queries parametrizadas, crypto seguro, headers/cookies seguros, secrets gestionados).
2. Análisis de trust boundaries: identificación de dónde entra data no confiable y qué controles aplican aguas arriba.
3. Reporte de auditoría con hallazgos clasificados por severidad (DREAD) y dominio, con foco de investigación en el flujo de datos completo (no el snippet aislado).
4. Threat model STRIDE para funcionalidades críticas.
5. Findings con severity ajustada: la protección aguas arriba NO elimina el hallazgo (defensa en profundidad) pero modera su severidad; si se descarta, comentario inline documentado.

## Reglas de negocio
1. Antes de escribir o revisar, hacer tres preguntas: **¿Cuáles son los trust boundaries?** (dónde entra data no confiable), **¿Qué puede controlar un atacante?** (qué inputs fluyen a operaciones sensibles) y **¿Cuál es el blast radius?** (peor resultado si falla la defensa).
2. Niveles de severidad DREAD: Critical 8-10 (fix inmediato), High 6-7.9 (sprint actual), Medium 4-5.9 (siguiente sprint), Low 1-3.9 (oportunista).
3. Investigar ANTES de reportar: trazar el origen del dato (¿user input, constante, interno?), chequear validación aguas arriba, examinar el trust boundary, leer el código circundante (middleware/interceptors que ya defienden).
4. La protección upstream ajusta la severidad pero no descarta el hallazgo (defensa en profundidad). Al downgradear/descartar: comentario inline (`// security: ...`) documentando la decisión.
5. Aplicar STRIDE para threat modeling de los componentes; DREAD para priorizar.
6. En Audit mode, cada fix se aplica en su propio worktree (un fix = un worktree = un PR enfocado y revertible).
7. Los hallazgos se consolidan deduplicando y jerarquizando por severidad.

## Restricciones
- Bugs defensivos no explotables (nil panics, aliasing) → skill `golang-safety`.
- Escaneo de dependencias con govulncheck como herramienta → skill `golang-dependency-management`.
- Cableado de scanners en CI → skill `golang-continuous-integration`.

## Casos límite
- SQL concatenación protegida aguas arriba por un parser estricto: hallazgo medium (no critical), pero se reporta igual con las defensas existentes documentadas.
- Fuzz testing sobre parsers de input: correr en CI y tratar hallazgos como Critical.
- pprof expuesto en producción: vector de información/DoS — proteger con auth y aislamiento de red.
- PII en mensajes de error/logs: no loguear; adjuntar IDs de correlación.

## Condiciones de error
- Hallazgo Critical (RCE, breach de datos, robo de credenciales): fix inmediato, bloquea release.
- Vulnerability reportada por `govulncheck` en el árbol: no ignorar — actualizar o migrar.
- Header de seguridad faltante (CSP, HSTS) en respuestas HTTP: agregar en middleware.

## Criterios de aceptación
- [ ] Se identificaron los trust boundaries del código escrito/revisado.
- [ ] No hay queries con concatenación de input, comandos shell con input sin sanitizar, ni crypto débil (PRNG para tokens).
- [ ] Los secrets no se hardcodean ni loguean.
- [ ] Los hallazgos de auditoría reportan severidad DREAD con ajuste por defensas aguas arriba.
- [ ] Los hallazgos descartados/rebajados tienen comentario inline documentado.
- [ ] En Audit mode, cada fix vive en su propio worktree/PR.

## Escenarios BDD

### Caso normal
Given un endpoint que recibe un término de búsqueda
When el agente implementa la consulta a BD
Then usa placeholder parametrizado ($1)
And escapa/valida el input antes de responder al cliente

### Caso alternativo
Given un parser de archivos subidos por usuarios
When el agente lo revisa
Then identifica el trust boundary (upload = data no confiable)
And verifica validación aguas arriba y propone fuzzing para el parser

### Caso límite
Given un SQL que concatena input pero protegido por un parser estricto aguas arriba
When el agente audita
Then reporta el hallazgo con severidad ajustada (medium, no critical)
And documenta cuáles defensas existen y qué pasa si se remueven

### Caso de error
Given una respuesta HTTP sin header de seguridad (CSP/HSTS) en una app web
When el agente revisa la configuración
Then marca la deficiencia de web security
And propone agregar headers seguros en el middleware del servidor