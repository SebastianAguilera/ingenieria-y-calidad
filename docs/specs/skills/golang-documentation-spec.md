# Especificación: SK-012 - golang-documentation: Documentación de proyectos Go

## Objetivo
Definir cómo documentar proyectos Go de forma completa y consistente: doc comments godoc, README, CONTRIBUTING, CHANGELOG, Go Playground, funciones Example, API docs y `llms.txt`. La documentación es un deliverable de primera clase — precisa, orientada a ejemplos y escrita para quien nunca vio el codebase. Sirve tanto para escribir documentación faltante como para auditar la existente.

## Entradas
- Proyecto Go a documentar (con o sin `main`, `cmd/`, `docs/`).
- Modo: **Write** (generar/completar documentación: doc comments, README, CONTRIBUTING, CHANGELOG, llms.txt; secuencial o con sub-agentes paralelos por paquete/archivo) o **Review** (auditar: hasta 5 sub-agentes, uno por capa: doc comments, README, CONTRIBUTING, CHANGELOG, extras por tipo de proyecto).
- Binarios: `go`.

## Salidas esperadas
1. Doc comments godoc que explican el "por qué"/"cuándo"/"restricciones", no solo el "qué".
2. Funciones `Example*` ejecutables como documentación (y tests).
3. README, CONTRIBUTING y CHANGELOG acordes al tipo de proyecto (librería vs aplicación/CLI).
4. `llms.txt` para orientar agentes IA (si aplica).
5. En Review mode: informe de gaps por capa de documentación.

## Reglas de negocio
1. **Concisión**: escribir la versión más corta que transmita la idea; nunca soltar hechos, warnings ni profundidad solicitada.
2. **Intención sobre paráfrasis**: el código muestra qué pasa; los docs explican por qué existe, cuándo usarlo y qué restricciones aplica.
3. **Sin contexto inventado**: omitir racionalizaciones sin sustento, claims de marketing (`seamlessly`, `robust`, `enterprise-grade`) y promesas futuras.
4. **Preservar significado al editar**: mantener la modalidad (`must`/`should`/`may` son obligaciones distintas); preservar condiciones, warnings y acciones requeridas.
5. **Anti-patterns a eliminar**: comentarios que solo parafrasean la signatura, restatement de firma, vocabulario de marketing, claims futuros sin base (`future extensibility`, `easy to scale`), transiciones huecas, template padding.
6. Detectar el tipo de proyecto primero: **Librería** (godoc comments + `Example*` + playground + pkg.go.dev) vs **Aplicación/CLI** (instalación, CLI help, configuración); ambos comparten función comments, README, CONTRIBUTING, CHANGELOG.
7. Para proyectos complejos, usar `docs/` con design docs.

## Restricciones
- Naming conventions de doc comments → skill `golang-naming`.
- Funciones Example → skill `golang-testing` (son tests ejecutables).
- Ubicación de archivos de documentación → skill `golang-project-layout`.

## Casos límite
- Doc comment que empieza con el nombre (requisito godoc) pero agrega auténtica información después del nombre: permitido.
- Comentario que solo repite el nombre y la firma: anti-pattern, se elimina.
- Documentación de librería vs aplicación: difieren en prioridades (godoc vs instalación/config).

## Condiciones de error
- Edición que cambia la modalidad (`must` por `may`): altera obligaciones — revertir y preservar el significado.
- README con instructions desactualizadas respecto al CLI real: auditar contra el código y corregir.

## Criterios de aceptación
- [ ] Los doc comments explican intención/restricciones, no solo parafrasean la signatura.
- [ ] No hay claims de marketing ni promesas futuras sin sustento.
- [ ] El tipo de proyecto fue detectado y la documentación priorizada en consecuencia.
- [ ] README/CONTRIBUTING/CHANGELOG existen y están actualizados.
- [ ] Hay funciones `Example*` para APIs públicas clave.
- [ ] En Review mode, el informe cubre cada capa de documentación.

## Escenarios BDD

### Caso normal
Given una función `Parse(input string) (*Config, error)` pública
When el agente escribe su doc comment
Then explica el propósito, cuándo usarla y qué restricciones aplican
And no parafrasea la firma (evita restatement)

### Caso alternativo
Given un proyecto CLI con `cmd/` y binario
When el agente decide qué documentación priorizar
Then enfoca en instalación, help text del CLI y configuración
And mantiene doc comments para el código interno relevante

### Caso límite
Given una lógica de negocio con una obligación fuerte (`must`)
When el agente edita la documentación existente
Then preserva la modalidad exacta del texto original
And no suaviza la obligación al "mejorar" la redacción

### Caso de error
Given un README que incluye `future extensibility` y `easy to scale` sin sustento
When el agente lo audita en Review mode
Then marca esos claims como contexto inventado / promesas futuras
And propone eliminarlos