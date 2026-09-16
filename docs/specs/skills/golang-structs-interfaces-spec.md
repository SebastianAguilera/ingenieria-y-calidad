# Especificación: SK-028 - golang-structs-interfaces: Diseño de structs e interfaces en Go

## Objetivo
Definir los patrones de diseño de tipos en Go: composición y embedding de structs/interfaces, type assertions y type switches, segregación de interfaces, inyección de dependencias vía interfaces, struct field tags, y receivers por puntero vs valor. Aplica al diseñar tipos, definir/implementar interfaces, embeber structs/interfaces, escribir assertions/switches, agregar tags para serialización JSON/YAML/DB, o elegir receivers. Favorece interfaces chicas y composables y tipos de retorno concretos: diseño para testeabilidad y claridad, no por la abstracción en sí.

## Entradas
- Código Go a diseñar/revisar u opinar (`**/*.go`).
- Tarea del operador: diseño de tipos, elección de receivers, definición de interfaces, tags, "accept interfaces, return structs", compile-time interface checks.
- Binarios: `go`.

## Salidas esperadas
1. Interfaces de 1-3 métodos, compuestas de interfaces chicas cuando se necesita un contrato mayor.
2. Interfaces definidas donde se consumen ("interfaces belong to consumers"), no donde se implementan.
3. Type assertions con comma-ok; type switches agrupando casos relacionados.
4. Structs con embedding/composición idiomática y tags de serialización correctos.
5. Receivers por puntero (mutación / tipos grandes) o por valor (inmutabilidad / tipos chicos) justificados.
6. Compile-time interface checks: `var _ Interface = (*Type)(nil)`.

## Reglas de negocio
1. "Cuanto más grande la interface, más débil la abstracción" — las interfaces DEBEN tener 1-3 métodos; compose interfaces chicas en contratos mayores.
2. Interfaces DEBEN definirse donde se consumen, no donde se implementan (el consumidor controla el contrato; evita importar un paquete solo por su interface).
3. Usar `v, ok := x.(T)` para assertions; `switch v := x.(type)` para switches por tipo.
4. Para inspeccionar errores envueltos usar `errors.As`, no assertions peladas (no recorre la wrap chain).
5. Preferir composición sobre herencia: embedding para reuso de comportamiento, no para modelar jerarquías.
6. Evitar `any`/`interface{}` cuando un tipo concreto sirve.
7. "Accept interfaces, return structs": las funciones dependen de interfaces; los constructores devuelven tipos concretos.
8. Struct field tags para JSON/YAML/DB coding; elegir nombres de campo estables para la serialización.
9. Receivers por puntero: para mutar el receptor o tipos grandes (evitar copia); por valor: para inmutabilidad y tipos chicos. Consistencia: todos los métodos del tipo usan el mismo receiver salvo justificación.
10. Compile-time checks: `var _ Interface = (*Type)(nil)` para validar que un tipo implementa la interface.

## Restricciones
- Naming de interfaces (`-er`) → skill `golang-naming`.
- Patrones de diseño/arquitectura → skill `golang-design-patterns`.

## Casos límite
- Interface con un solo método que necesita contrate mayor: compose (`io.ReadWriter`).
- Struct embebido que compite en métodos con el outer: el embedding expone el método según prioridad — documentar el comportamiento.
- Campo `any` para datos flexibles (JSON desconocido): aceptable solo con valor `json.RawMessage` tipado o validación; nunca `any` por comodidad.

## Condiciones de error
- Assertion de tipo pelada (`err.(*MyError)`) que no detecta un error envuelto con `%w`: usar `errors.As`.
- Interface definida en el paquete implementador en lugar del consumidor: ciclo de imports o acoplamiento — moverla al consumidor.
- Receiver mixto (puntero en unos métodos, valor en otros) sin justificación: inconsistencia y copias inesperadas — unificar.

## Criterios de aceptación
- [ ] Las interfaces tienen 1-3 métodos (o se componen de chicas).
- [ ] Las interfaces se definen en el punto de consumo.
- [ ] Las assertions usan comma-ok; errores envueltos con `errors.As`.
- [ ] Los receivers son consistentes y justificados (mutación/tamaño vs inmutabilidad).
- [ ] Los tags de serialización están correctamente definidos.
- [ ] Los tipos implementan interfaces con checks compile-time.

## Escenarios BDD

### Caso normal
Given un repositorio con métodos `Save`, `FindByID` y `Delete`
When el agente define el contrato para el service
Then define la interface `UserStore` (chica, solo lo que el service consume) en el paquete del service
And el repository la implementa validada por `var _ domain.UserStore = (*repo)(nil)`

### Caso alternativo
Given un struct `Config` que necesita serialización a JSON y validación
When el agente diseña los tags
Then agrega `json:"campo"` y tags de validación apropiados
And mantiene nombres de campo estables en la API pública

### Caso límite
Given un método que muta el estado del receptor (ej. `AddUser`)
When el agente elige el receiver
Then usa receiver por puntero (`func (s *Service) AddUser(...)`)
And mantiene la consistencia en todos los métodos del tipo

### Caso de error
Given una assertion `err.(*MyError)` sobre un error envuelto con `%w`
When el agente revisa el código
Then detecta que la assertion pelada no recorre la wrap chain
And la reemplaza por `errors.As(err, &myErr)`