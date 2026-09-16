# Especificación: SK-010 - golang-dependency-management: Gestión de dependencias en Go

## Objetivo
Definir la gestión correcta de dependencias en Go: `go.mod` y `go.sum`, flujos de `go get` (instalación y upgrade), Minimal Version Selection, resolución de conflictos con `replace`/`exclude`/`retract`, escaneo con `govulncheck`, auditoría de dependencias desactualizadas y tamaño de binarios, vendoring, directivas `tool` y workspaces `go.work`. Cada nueva dependencia se trata como un compromiso de mantenimiento a largo plazo: antes de sumarla se evalúa si la standard library ya resuelve el problema.

## Entradas
- Proyecto Go con `go.mod`/`go.sum` y código que referencia paquetes.
- Tarea del operador: agregar, remover o actualizar dependencias; resolver conflictos de versiones; auditar el árbol de módulos.
- Binarios: `go`, `govulncheck` (`go install golang.org/x/vuln/cmd/govulncheck@latest`).

## Salidas esperadas
1. `go.mod`/`go.sum` actualizados de forma consistente (`go mod tidy` aplicado).
2. Dependencias agregadas solo tras confirmación explícita del operador (regla del skill para agentes IA).
3. Escaneo de vulnerabilidades con `govulncheck` antes de cada release.
4. `go.sum` commiteado (checksums criptográficos que detectan tampering del supply chain).
5. Vendoring cuando se requieren builds herméticos.
6. Workspaces `go.work` para módulos locales múltiples.

## Reglas de negocio
1. **Los agentes IA DEBEN pedir confirmación al usuario antes de `go get` para agregar una dependencia nueva** (pueden sugerir paquetes no mantenidos, de baja calidad o innecesarios). `go get -u` para actualizar una existente es seguro.
2. Evaluar antes de proponer una dependencia: ¿la stdlib ya cubre el caso? ¿la licencia es compatible? ¿hay alternativas conocidas? ¿qué hace y por qué se necesita?
3. Preferir paquetes de la lista curada del skill `golang-popular-libraries`; si no hay opción curada, favorecer paquetes del Go team (`golang.org/x/...`) u organizaciones establecidas.
4. `go.sum` DEBE committearse — permite a `go mod verify` detectar manipulación del proxy.
5. `govulncheck ./...` (o `go tool govulncheck ./...`) antes de cada release — detecta CVEs conocidos en el árbol de dependencias.
6. Considerar estado de mantenimiento, licencia y alternativas stdlib antes de sumar una dependencia (cada una aumenta superficie de ataque, carga de mantenimiento y tamaño de binario).
7. `go mod tidy` antes de cada commit que cambie dependencias.
8. Usar `go mod vendor` para builds herméticos (sin red, reproducibilidad), CI y Docker; commitear el directorio `vendor/`.

## Restricciones
- No corrige vulnerabilidades explotables en código (→ skill `golang-security`).
- No cablea Dependabot/Renovate en CI (→ skill `golang-continuous-integration`).

## Casos límite
- Conflicto de versiones entre dos dependencias transitivas: MVS resuelve automáticamente; `replace`/`exclude`/`retract` solo con justificación documentada.
- Módulo con dependencia marcada `retract` por el upstream: no usarla; buscar alternativa o versión corregida.
- Entorno de build sin acceso al proxy de módulos: activar vendoring y commitear `vendor/`.
- Dependencia que requiere Go más nuevo que el `go` directive: decidir entre actualizar el directive o descartar la dependencia.

## Condiciones de error
- `go.sum` no verifica (`go mod verify` falla): posible tampering — investigar antes de cualquier build de producción.
- `govulncheck` reporta CVE conocido: bloquear el release hasta actualizar o mitigar.
- `go mod tidy` elimina una dependencia usada en código no compilado (ej. build tags): verificar antes de hacer commit.

## Criterios de aceptación
- [ ] Ninguna dependencia nueva se agregó sin confirmación del operador.
- [ ] `go.sum` está commiteado y verificado.
- [ ] `govulncheck` se ejecutó y no quedan vulnerabilidades conocidas sin tratar.
- [ ] `go mod tidy` se aplicó antes de commitear cambios de dependencias.
- [ ] Las decisiones de `replace`/`retract` están documentadas.
- [ ] El vendoring se usa cuando el entorno lo requiere.

## Escenarios BDD

### Caso normal
Given un módulo que necesita soporte de logging estructurado
When el operador pide agregar una dependencia
Then el agente evalúa stdlib, licencia y alternativas
And solicita confirmación antes de ejecutar `go get`

### Caso alternativo
Given una dependencia existente que requiere actualización de versión
When el operador solicita el upgrade
Then el agente ejecuta `go get -u <paquete>`
And corre `go mod tidy` y la suite de tests para verificar compatibilidad

### Caso límite
Given un entorno de CI/Docker sin acceso al proxy de módulos
When el agente configura el build
Then habilita vendoring con `go mod vendor`
And commitea el directorio `vendor/` para builds herméticos

### Caso de error
Given un `govulncheck` que reporta un CVE en una dependencia del árbol
When el agente se prepara para un release
Then detiene el release
And propone actualizar la versión o migrar a una alternativa segura