# Especificación: SK-001 - go-feature: Creación de features en Go con Clean Architecture

## Objetivo
Guía paso a paso al agente de IA para crear una feature completa en el proyecto **Software Metrics & Estimation Engine**, respetando la arquitectura por capas (`domain → repository → service → handler → tests`). Garantiza que cada feature nueva siga la misma estructura, sea testeable mediante mocks y mantenga la separación de responsabilidades definida en `AGENTS.md`.

## Entradas
- Pedido de creación de una feature, expresado en lenguaje natural (ej: "Creá la feature de Planning Poker: modelar Votacion con votos individuales, revelado y detección de dispersión").
- Arquitectura base del proyecto: `cmd/api/`, `internal/domain/`, `internal/repository/`, `internal/service/`, `internal/handler/`.
- Convenciones del repo: manejo explícito de errores (`if err != nil`, `fmt.Errorf("contexto: %w", err)`), TDD (tests antes de implementar).
- Requerimientos de dominio relevante (historias del Product Backlog).

## Salidas esperadas
1. Código de la capa **domain**: structs en `internal/domain/` e interfaces que el repository debe implementar (solo stdlib, sin dependencias externas).
2. Código de la capa **repository**: implementación de la interface de domain en `internal/repository/` (PostgreSQL/GORM), solo acceso a datos, sin reglas de negocio.
3. Código de la capa **service**: lógica de negocio en `internal/service/` que depende solo de interfaces de domain; errores con `fmt.Errorf("contexto: %w", err)`.
4. Código de la capa **handler**: controlador HTTP en `internal/handler/` que valida entradas antes de llamar al service y responde con los códigos HTTP correctos.
5. **Tests** escritos antes de la implementación (TDD): table-driven tests y mocks para dependencias externas.
6. Lista de archivos a crear/modificar.

## Reglas de negocio
1. El flujo de capas es obligatorio: domain → repository → service → handler → tests, en ese orden.
2. `internal/domain/` no puede tener dependencias externas (solo stdlib).
3. `internal/repository/` implementa las interfaces definidas en domain y NO contiene reglas de negocio.
4. `internal/service/` depende solo de interfaces de domain (nunca de implementaciones concretas) y se testea con mocks de repository.
5. `internal/handler/` valida las entradas antes de delegar en el service y responde con códigos HTTP correctos.
6. Los tests se escriben ANTES que la implementación (ciclo TDD RED → GREEN → REFACTOR).
7. Todo error se maneja explícitamente (`if err != nil`), nunca se descarta.

## Restricciones
- Solo aplica a features del proyecto Go. No modifica `docs/` ni la configuración del harness.
- Los structs de domain no pueden importar GORM, drivers de BD ni ninguna librería externa.
- El repository es la única capa permitida para interactuar con la persistencia.

## Casos límite
- Feature que requiere varias entidades relacionadas (ej: Sprint con Historias): se modelan todas las structs de domain con sus interfaces antes de pasar a repository.
- Feature sin capa HTTP (solo regla de negocio): se omite el handler pero se mantienen domain, service y tests.
- Repository con GORM: los tests usan mocks de la interface de domain, no la base de datos real.

## Condiciones de error
- Diseño de feature que viola la separación de capas: se corrige antes de generar código.
- Dependencia externa requerida en domain: se rechaza y se busca la alternativa (interface propia).
- Código sin tests asociados: se solicita completar los tests antes de entregar la feature.

## Criterios de aceptación
- [ ] Los structs de domain se definen sin dependencias externas.
- [ ] Las interfaces de domain se definen en `internal/domain/`.
- [ ] El repository implementa las interfaces de domain sin reglas de negocio.
- [ ] El service depende solo de interfaces y maneja errores con `fmt.Errorf("contexto: %w", err)`.
- [ ] El handler valida entradas y responde códigos HTTP correctos.
- [ ] Existen tests escritos antes de la implementación, con table-driven tests y mocks.
- [ ] Se entrega la lista de archivos a crear/modificar.

## Escenarios BDD

### Caso normal
Given un pedido de feature "modelar Votacion" para la capa de Planning Poker
When el agente ejecuta el flujo del skill
Then genera domain, repository, service, handler y tests en el orden correcto
And lista los archivos creados/modificados

### Caso alternativo
Given una feature que es solo regla de negocio sin exposición HTTP
When el agente ejecuta el flujo del skill
Then genera los modelos de domain, el service y sus tests
And omite la capa handler

### Caso límite
Given una feature con múltiples entidades relacionadas (Sprint e Historias)
When el agente modela la feature
Then define todas las structs e interfaces de domain primero
And recién después implementa repository, service y handler

### Caso de error
Given un diseño propuesto que incluye lógica de negocio dentro del repository
When el agente revisa la arquitectura generada
Then detecta la violación de capas
And corrige moviendo la lógica al service antes de entregar el código