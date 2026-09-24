# Especificación: SK-029 - golang-testing: Tests Go de producción

## Objetivo
Definir las prácticas de testing de producción en Go: table-driven tests con subtests nombrados, testify como helper, tests paralelos, fuzzing, fixtures, detección de goroutine leaks con goleak, snapshot testing, cobertura de código, tests de integración con build tags y naming idiomático. Se escriben tests como especificaciones ejecutables: restringen comportamiento, no persiguen metas de cobertura.

## Entradas
- Código Go a testear o tests existentes (`**/*.go`).
- Modo: **Write** (generar tests: `gotests` para scaffold table-driven + edge cases + error paths), **Review** (diff de PR de tests: cobertura de comportamiento nuevo, calidad de assertions, estructura table-driven, ausencia de flakiness), **Audit** (hasta 3 sub-agentes: calidad unit y gaps de cobertura; aislamiento de integración y build tags; leaks de goroutines y races) o **Debug** (test fallando/flaky: reproducir, aislar la assertion, trazar root cause).
- Binarios: `go`, `gotests` (`go install github.com/cweill/gotests/gotests@latest`).

## Salidas esperadas
1. Tests table-driven con el campo `name` pasado a `t.Run`, un test file por archivo fuente (`foo.go` → `foo_test.go`).
2. Tests de integración separados con `//go:build integration`.
3. Subtests independientes entre sí y ejecutables en cualquier orden.
4. `t.Parallel()` en tests independientes cuando sea seguro.
5. Asserts sobre comportamiento observable y contracts públicos, no detalles de implementación.
6. `goleak.VerifyTestMain` en packages con goroutines.
7. Fuzzing para parsers de input; fixtures; cobertura en CI con race detection.

## Reglas de negocio
1. Los table-driven tests DEBEN usar subtests nombrados — cada caso necesita un campo `name` pasado a `t.Run`.
2. Los tests de integración DEBEN usar build tags (`//go:build integration`) para separarse de los unit.
3. Los tests NO DEBEN depender del orden de ejecución — cada test DEBE poder correrse de forma independiente.
4. Tests independientes DEBERÍAN usar `t.Parallel()` cuando es posible.
5. Los tests DEBEN assert comportamiento observable y contracts públicos, no detalles de implementación (un test acoplado a internals convierte cada refactor en reescritura del test).
6. Los packages con goroutines DEBERÍAN usar `goleak.VerifyTestMain` en `TestMain`.
7. Usar testify como helper, no como reemplazo del stdlib.
8. Mockear interfaces, no tipos concretos.
9. Mantener tests unit rápidos (< 1ms); tests de integración con build tags.
10. Correr tests con race detection en CI.
11. Incluir examples como documentación ejecutable.
12. Los test files DEBEN nombrarse según el archivo fuente que testean, no según la función/método.
13. Los test functions DEBERÍAN aparecer en el mismo orden que las funciones/métodos en el archivo fuente.
14. Convenciones de archivo: `package foo` (white-box, acceso a no exportados) vs `package foo_test` (black-box, API pública).

## Restricciones
- APIs específicas de testify → skill `golang-stretchr-testify`.
- Metodología de medición → skill `golang-benchmark`.

## Casos límite
- Test que depende de otro (estado compartido no reseteado): flaky e ilegal — independencia obligatoria.
- Función con muchos edge cases: tabla con casos de error y límite además de los normales.
- Test lento de red/BD: build tag de integración, fuera del `go test ./...` default.
- Parser de input de usuario: fuzzing (correr en CI, hallazgos tratados como críticos).

## Condiciones de error
- Goroutine leak detectado por goleak: el test falla — eliminar el leak antes de continuar.
- Race detectado por `-race` en CI: bloquea el merge.
- Subtest sin `name`: falla la convención y los reports son inidentificables — corregir.
- Test acoplado a internals en un refactor: se reescribe el test al contract público, no al revés.

## Criterios de aceptación
- [ ] Todos los table-driven tests usan subtests nombrados.
- [ ] Los tests de integración usan build tags y no corren en el default.
- [ ] Los tests son independientes del orden y corren con `-race` sin fallos.
- [ ] Los packages con goroutines verifican leaks con goleak.
- [ ] Los asserts cubren contracts públicos, no implementación.
- [ ] El naming de archivos y funciones sigue el convenio (`foo_test.go`, `TestFoo`).

## Escenarios BDD

### Caso normal
Given una función `CalculaVelocidad(sprints []Sprint) float64`
When el agente escribe los tests
Then crea `velocidad_test.go` con una tabla de casos (normal, sin datos, con datos parciales)
And cada caso pasa por `t.Run(tt.name, ...)` con inputs y expected

### Caso alternativo
Given un service que conversa con la BD y una API externa
When el agente prueba el happy path sin infraestructura
Then mockea las interfaces de repository y API
And verifica el comportamiento observable del service, no sus internals

### Caso límite
Given un parser de input de usuario (fechas, rangos, texto)
When el agente refuerza los tests
Then agrega casos límite (rango inválido, cadena vacía, caracteres especiales) a la tabla
And agrega fuzzing para el parser si cruza un trust boundary

### Caso de error
Given un package con goroutines que se filtran en un test
When el agente integra goleak
Then agrega `goleak.VerifyTestMain` en `TestMain`
And el test falla hasta eliminar el leak detectado