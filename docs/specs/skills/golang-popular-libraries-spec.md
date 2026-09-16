# Especificación: SK-022 - golang-popular-libraries: Selección de librerías Go

## Objetivo
Recomendar librerías y frameworks Go probados en producción por categoría (web, base de datos, testing, logging, messaging), cubriendo también paquetes nuevos/experimentales del stdlib, tradeoffs de standard-library-first y señales de madurez (mantenimiento, licencia, cantidad de importadores). Se aplica cuando el usuario pide sugerencias de librerías, quiere comparar alternativas, elegir una librería para una tarea o cuando se agrega una dependencia nueva al proyecto.

## Entradas
- Requerimiento funcional o tecnológico que necesita una librería Go.
- Catálogos de referencia del skill: stdlib nuevo/experimental, librerías por categoría, herramientas de desarrollo.
- Tarea del operador: sugerencia, comparación o decisión de adopción de una dependencia.

## Salidas esperadas
1. Recomendación de la librería más simple y production-ready para el caso, o la recomendación de NO usar librería (stdlib suficiente).
2. Comparación de alternativas con criterios: madurez, mantenimiento, licencia, simplicidad, dependencias, performance.
3. Señales de madurez del candidato: estado de mantenimiento, licencia, `imported-by` count en pkg.go.dev.
4. Advertencias de anti-patterns (librerías que envuelven stdlib sin valor agregado, abandonadas, con footprint grande).

## Reglas de negocio
1. Priorizar: **production-readiness** (maduras y mantenidas), **simplicidad** (filosofía Go), **performance** y **stdlib first** (preferir stdlib cuando cubre el caso; solo externas si aportan valor claro).
2. Evaluar requerimientos primero (caso de uso, necesidades de performance, restricciones).
3. Chequear siempre la stdlib antes de recomendar una externa.
4. Priorizar madurez: DEBE verificarse mantenimiento, licencia y adopción de la comunidad antes de recomendar. Usar el conteo de `imported-by` como señal de popularidad y presión de backward-compat.
5. Considerar la complejidad — las soluciones simples suelen ser mejores en Go.
6. Pensar en dependencias — más dependencias = más superficie de ataque y carga de mantenimiento.
7. Anti-patterns a evitar: sobre-ingeniería de problemas simples, librerías que envuelven stdlib sin valor, abandonadas (preguntar al desarrollador antes de recomendar), footprints grandes para necesidades simples, ignorar alternativas stdlib.
8. El catálogo no es exhaustivo: para verificar un candidato usar `godig` (pakete del skill `golang-pkg-go-dev`); una vez agregado al build, `gopls` para navegar su código resuelto.

## Restricciones
- No documenta la API de una librería una vez elegida (→ skill específico de la librería).
- No cubre mecánicas de go.mod ni auditorías de versiones (→ `golang-dependency-management`).

## Casos límite
- Caso cubierto por stdlib (`slices`, `maps`, `slog`, `http`): recomendar no agregar dependencia.
- Librería abandonada pero funcional: preguntar antes de recomendarla; no sugerir en silencio.
- Dos candidatas con madurez similar: comparar por `imported-by`, licencia y mantenimiento con godig.

## Condiciones de error
- Candidato sin mantenimiento activo (último release hace años) y sin alternativa clara: advertir el riesgo y pedir decisión explícita del equipo.
- Licencia incompatible con el proyecto: descartar el candidato.
- Librería que duplica una función estándar sin valor agregado: no recomendar.

## Criterios de aceptación
- [ ] La recomendación evalúa primero si el stdlib resuelve el caso.
- [ ] Todo candidato se verifica por mantenimiento, licencia y adopción (imported-by).
- [ ] Las alternativas se comparan con criterios explícitos.
- [ ] Se advierte sobre anti-patterns y librerías en riesgo.
- [ ] Las verificaciones de candidatos usan `godig` antes que Context7.

## Escenarios BDD

### Caso normal
Given "necesito logging estructurado en la API"
When el agente recomienda
Then evalúa `log/slog` del stdlib (Go 1.21+)
And recomienda no agregar dependencia externa salvo que se requiera fan-out complejo

### Caso alternativo
Given "necesito manejar concurrencia con pools de workers"
When el agente evalúa candidatos
Then verifica madurez, licencia e imported-by con godig
And sugiere la opción más simple probada en producción (o patrón stdlib con errgroup si alcanza)

### Caso límite
Given una librería abandonada que resuelve exactamente el caso
When el agente debe decidir
Then advierte el riesgo de no mantenimiento
And pregunta explícitamente al desarrollador antes de recomendarla (o propone fork/alternativa)

### Caso de error
Given una librería que envuelve funcionalidad del stdlib sin valor agregado
When el agente revisa la suglicación
Then identifica el anti-pattern de over-engineering
And recomienda el stdlib en su lugar