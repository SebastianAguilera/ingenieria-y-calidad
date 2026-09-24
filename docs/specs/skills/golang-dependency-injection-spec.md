# Especificación: SK-009 - golang-dependency-injection: Inyección de dependencias en Go

## Objetivo
Definir cómo aplicar dependency injection (DI) en Go para construir aplicaciones testeables y débilmente acopladas: los servicios declaran qué necesitan y el caller (o contenedor) lo provee. Cubre por qué DI importa, inyección manual por constructores y comparación de librerías DI (google/wire, uber-go/dig, uber-go/fx, samber/do). Se elige el enfoque más simple que resuelva el problema sin over-engineering.

## Entradas
- Proyecto Go con grafo de dependencias existente o a diseñar.
- Modo: **Design** (nuevo proyecto/servicio: evaluar grafo y lifecycle, recomendar inyección manual o librería según la decisión table, generar el wiring) o **Refactor** (código acoplado: hasta 3 sub-agentes identifican globals/`init()`, dependencias concretas que deberían ser interfaces, y service-locator anti-patterns).
- Binarios: `go`.

## Salidas esperadas
1. Dependencias inyectadas vía constructores (manual o con librería según el caso).
2. Interfaces definidas donde se consumen ("accept interfaces, return structs").
3. Composición root centralizada en `main()` o app startup con el lifecycle de servicios.
4. Mocks en tests a través del boundary de interface.
5. En Refactor mode: plan de migración consolidado con los hallazgos de los 3 sub-agentes.

## Reglas de negocio
1. Dependencias DEBEN inyectarse vía constructores; NUNCA variables globales ni `init()` para setup de servicios.
2. Proyectos pequeños (< 10 servicios) DEBEN usar inyección manual por constructores, sin librería.
3. Interfaces DEBEN definirse donde se consumen, no donde se implementan.
4. NUNCA registries globales ni service locators a nivel paquete.
5. El contenedor DI DEBE existir solo en la composition root (`main()` o startup); NUNCA pasar el contenedor como dependencia.
6. Preferir lazy initialization — los servicios se crean recién cuando se solicitan.
7. Singletons para servicios con estado (conexiones de BD, caches); transients para stateless.
8. Mockear en el boundary de interface — DI lo hace trivial.
9. Mantener el grafo de dependencias poco profundo; cadenas profundas señalan problemas de diseño.
10. Elegir librería DI según tamaño del proyecto y equipo (decision table del skill); para scripts de 2-3 funciones, wiring manual.

## Restricciones
- Para APIs específicas de una librería DI (wire, dig, fx, samber/do) remitirse a la documentación oficial.
- Los fundamentos de interfaces se complementan con el skill `golang-structs-interfaces`.

## Casos límite
- Proyecto con 3-4 servicios y sin complejidad: inyección manual, sin librería (evitar over-engineering).
- Servicio con estado compartido (pool de conexiones): singleton gestionado por el container.
- Servicio sin estado usado en cada request: transient creado bajo demanda (lazy).

## Condiciones de error
- Código que crea sus dependencias internamente (`new()` o global init): imposible de testear aislado — refactorizar a inyección por constructor.
- Container pasado como parámetro a servicios: service locator anti-pattern — reemplazar por interfaces concretas inyectadas.
- `init()` con efectos laterales de setup: testeabilidad impredecible — migrar a constructores explícitos.

## Criterios de aceptación
- [ ] Las dependencias se inyectan por constructor.
- [ ] No hay globals ni `init()` para setup de servicios.
- [ ] Las interfaces se definen donde se consumen.
- [ ] No hay service locators (container como argumento).
- [ ] El container (si existe) vive solo en la composition root.
- [ ] Los tests mockean en el boundary de interface.
- [ ] El enfoque elegido es el más simple que resuelve el problema.

## Escenarios BDD

### Caso normal
Given un servicio UserService que necesita BD, mailer y logger
When el agente diseña el wiring
Then define un constructor `NewUserService(db, mailer, logger)`
And en los tests pasa un mock de `UserStore` y `Mailer`

### Caso alternativo
Given un proyecto pequeño con menos de 10 servicios
When el agente recomienda el enfoque DI
Then sugiere inyección manual por constructores sin librería
And evita agregar una librería DI innecesaria

### Caso límite
Given un service stateless usado en cada request pero que tarda en inicializarse
When el agente diseña el lifecycle
Then lo declara lazy (se crea al primer uso)
And evita cargarlo al startup

### Caso de error
Given funciones de servicio que se inicializan con variables globales e `init()`
When el agente ejecuta Refactor mode
Then los sub-agentes identifican los globals y el setup en `init()`
And se produce un plan de migración a inyección por constructor