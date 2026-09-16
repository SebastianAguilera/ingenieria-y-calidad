# Especificación: SK-018 - golang-naming: Convenciones de nomenclatura en Go

## Objetivo
Aplicar las convenciones de nomenclatura de Go de forma consistente: paquetes, constructores, structs, interfaces, constantes, enums, errores, booleanos, receivers, getters/setters, functional options, acrónimos, funciones de test y nombres de subtest. Aplica al escribir código nuevo, revisar/refactorizar, elegir entre alternativas de nombres, debatir nombres de paquete o responder dudas sobre mejores prácticas de nombres.

## Entradas
- Código Go a nombrar o revisar (`**/*.go`).
- Tarea del operador: decisiones de naming (package, type, function, variable, error, test).
- Binarios: `go`.

## Salidas esperadas
1. Identificadores Go con `MixedCaps`/`mixedCaps` (nunca underscores, fuera de excepciones).
2. Paquetes en minúscula de una palabra; archivos en minúscula con underscores permitidos.
3. Nombres exportados en UpperCamelCase; no exportados en lowerCamelCase.
4. Interfaces nombradas por método + `-er` (`Reader`); errores con prefijo `Err` (vars) y sufijo `Error` (types); constructores `New`/`NewTypeName`.
5. Booleanos con prefijo `is`/`has`/`can`; opciones con `With` + campo; enums con prefijo del tipo y unknown en 0.
6. Strings de error en minúscula sin puntuación; acrónimos en caps consistentes (`URL`, `xmlParser`).

## Reglas de negocio
1. Todos los identificadores DEBEN usar `MixedCaps` (exported) / `mixedCaps` (unexported); NUNCA underscores — excepciones: subtest names (`TestFoo_InvalidInput`), código generado e interop OS/cgo.
2. Paquete: minúscula, una palabra, sufijo `_test` OK solo para files de test.
3. Interface: nombre de método + `-er` (`Reader`, `Closer`); evita sufijos redundantes.
4. Struct: sustantivo en MixedCaps (`Request`, `FileHeader`).
5. Constantes: MixedCaps, NUNCA ALL_CAPS.
6. Receiver: abreviatura de 1-2 letras (`s *Server`, `b *Buffer`).
7. Errores: variable `Err` prefix (`ErrNotFound`), tipo `Error` suffix (`PathError`).
8. Constructores: `New` (tipo único) o `NewTypeName` (multi-tipo).
9. Opciones funcionales: `With` + nombre de campo (`WithPort()`).
10. Enum (iota): prefijo del nombre de tipo + valor; zero value = Unknown que NO colisiona con un estado real.
11. Funciones de test: `Test` + nombre de la función probada.
12. Acrónimos: todo caps o todo lower, consistentes (`URL`, `HTTPServer`, `xmlParser`).
13. Error strings: minúscula (incluidos acrónimos), sin puntuación final.
14. Alias de import: corto, solo en colisión.
15. Formato de funciones: sufijo `f` (`Errorf`, `Wrapf`, `Logf`).
16. Al ignorar una regla, agregar un comentario al código.

## Restricciones
- Style de código (line length, control flow) → skill `golang-code-style`.
- Doc comments → skill `golang-documentation`.
- Linter configuration → skill `golang-lint`.

## Casos límite
- Constante usada como magic number: nombre descriptivo en MixedCaps en lugar de ALL_CAPS.
- Interface con un solo método no nombrable con `-er` (`ReadWriteSeeker` no aplica): usar convenciones canónicas del stdlib.
- Enum con `iota` empezando en 0 sin Unknown: el zero value de la struct pasa como estado válido — mitigar con sentinel 0.
- Boolean field en struct externo: prefijo `is`, `has`, `can` en campos y métodos (`isReady`, `IsConnected()`).

## Condiciones de error
- `MAX_PACKET_SIZE` o `max_packet_size`: viola MixedCaps y rompe convenciones de exportación/tooling — renombrar.
- `getUser()` con prefijo `get`: anti-pattern en Go — usar nombres directos o descriptivos según semántica.
- Error string con mayúscula inicial: inconsistente con el estilo Go y rompe agrupación de logs — corregir.

## Criterios de aceptación
- [ ] Todos los identificadores usan MixedCaps (sin underscores fuera de excepciones).
- [ ] Los nombres de paquete son minúsculas de una palabra.
- [ ] Las interfaces siguen el patrón método + `-er`.
- [ ] Los errores usan `Err` prefix / `Error` suffix según corresponda.
- [ ] Los constructores y opciones siguen `New`/`NewTypeName`/`With`+campo.
- [ ] Los enums arrancan con unknown en 0 o en 1.
- [ ] Las funciones de test se nombran `Test` + función probada.

## Escenarios BDD

### Caso normal
Given un tipo `User` con un constructor de múltiples opciones
When el agente lo nombra
Then usa `NewUser` como constructor
And opciones `WithEmail()`, `WithRole()` (functional options)

### Caso alternativo
Given una interface con un único método `Read(p []byte) (n int, err error)`
When el agente la nombra
Then la llama `Reader` (método + `-er`)
And compone interfaces grandes desde `Reader`/`Writer` en lugar de una interface monolítica

### Caso límite
Given un enum de estados que empieza en 0 con iota
When el agente lo define
Then agrega `StatusUnknown` en 0
And los estados reales comienzan en 1 para que el zero value no sea un estado válido silencioso

### Caso de error
Given una constante global escrita como `MAX_PACKET_SIZE` (C-style)
When el agente revisa el código
Then marca la violación de MixedCaps
And la renombra a `MaxPacketSize` preservando la semántica