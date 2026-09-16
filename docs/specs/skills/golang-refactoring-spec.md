# Especificación: SK-024 - golang-refactoring: Refactorización segura de código Go

## Objetivo
Definir la metodología de refactorización segura, a escala, de código Go existente: red de seguridad adaptada a la cobertura, transformaciones que preservan comportamiento (gopls Rename/Extract, `gofmt -r`, `gopatch`), catálogo Fowler mapeado a Go, ruptura de ciclos de importación y PRs pequeños y apilados. La regla central: **nunca mezclar cambios estructurales y de comportamiento en el mismo commit/PR**.

## Entradas
- Código Go existente a refactorizar (`**/*.go`).
- Modo: **Plan** (gate obligatorio antes de editar: mapear estructura y blast radius con gopls, inventario de refactors, ordenar, sign-off del usuario), **Execute** (human-in-the-loop: un sub-agente, un worktree, una rama, un PR por cambio atómico), **Simple-sweep** (transform mecánica única tree-wide, puede usar `ultracode`) o **Review** (verificar separación estructural/conductual y preservación de comportamiento).
- Binarios: `go`, `gopls`, `git`, `golangci-lint`, `benchstat`; auxiliares: `deadcode`, `eg`, `gopatch`, `gh`.

## Salidas esperadas
1. Plan de refactoring con blast radius mapeado (gopls references/call hierarchy/package API) y orden de ejecución aprobado.
2. Red de seguridad: tests agregados antes de tocar código con cobertura insuficiente del blast radius.
3. Cambios pequeños, mecánicos y tool-driven (100–500 líneas por PR), cada uno verificado con `go build` + `go vet` + `go test` (y `-race`/`benchstat` cuando corresponde).
4. Commits atómicos de una sola categoría (puramente estructural o puramente conductual, nunca ambos).
5. Ruptura de ciclos de importación cuando aplica.

## Reglas de negocio
1. Refactoring (Fowler) = cambiar la estructura interna para hacerla más fácil de entender/costar menos de modificar, **sin cambiar el comportamiento observable**.
2. **NUNCA mezclar cambios estructurales y de comportamiento en un commit o PR.**
3. Separar un move de código de una optimización en dos PRs secuenciales (la verificación difiere: gopls+build/test vs benchmarks).
4. Apuntar a 100–500 líneas por PR: chico para review en una sesión, grande para leerse como un cambio coherente.
5. Preferir gopls Rename/Inline sobre hand-edits LLM: son behavior-preserving por construcción (Rename rechaza shadowing, interface breakage o código malformado; Inline sustituye argumentos con side effects en temporales).
6. Cuando un cambio recurre en muchos sitios, generar una herramienta de reescritura en lugar de editar cada sitio manualmente.
7. Gatear la estrategia por la cobertura del **blast radius**, no la cobertura global; escribir el test de seguridad como mecanismo propio de verificación.
8. Flujo core: **Entender → Red de seguridad → Paso tool-driven chico → Verificar → Commit atómico de una categoría → Repetir**.
9. Los gates de aprobación (Plan mode y checkpoints) se preguntan por la herramienta de preguntas del entorno — son decisiones irreversibles.
10. No usar `ultracode`/Workflows para refactors multi-paso con revisión humana entre merges (corren agente-a-agente sin checkpoint humano).

## Restricciones
- Renames/estilos/proyecto → skills `golang-naming`, `golang-project-layout`, `golang-modernize`, `golang-code-style`, `golang-design-patterns` según el tipo de cambio.

## Casos límite
- Código sin cobertura en el blast radius: escribir tests de caracterización (recipes HIGH/MEDIUM/LOW del skill) antes de refactorizar.
- Cambio que toca hot path: verificación incluye `-race` y benchmarks con benchstat.
- Ciclo de imports entre paquetes: romper con interfaces en el consumidor o movimiento de tipos (PR secuenciales).
- Rename que rompería interface: gopls rechaza la operación — se re-planifica.

## Condiciones de error
- PR mixto estructural+conductual: se rechaza en Review mode; se divide en PRs separados.
- Transform aplicada sin red de seguridad y con build roto: se detiene y se restaura desde el worktree aislado.
- Sign-off no obtenido en Plan mode: no se toca código hasta la aprobación explícita.

## Criterios de aceptación
- [ ] Existe un plan con blast radius mapeado y sign-off del usuario.
- [ ] La red de seguridad (tests) cubre el blast radius antes de editar.
- [ ] Los cambios son tool-driven (gopls/gofmt -r/gopatch) siempre que sea posible.
- [ ] Cada PR es de una sola categoría y de 100–500 líneas.
- [ ] La verificación incluye build + vet + test (y race/benchmark en hot paths).
- [ ] No se usó ultracode para refactors multi-paso que necesitan review humano.

## Escenarios BDD

### Caso normal
Given una función de 300 líneas con un bloque extraíble
When el agente la refactoriza
Then mapea el blast radius con gopls y agrega tests si falta cobertura
And extrae el bloque con gopls Extract manteniendo el comportamiento
And verifica con build + vet + test y commitea el cambio estructural puro

### Caso alternativo
Given un rename de `UserService` a `AccountService` en decenas de archivos
When el agente lo aplica
Then usa gopls Rename (behavior-preserving por construcción, rechaza si rompe interfaces)
And separa el commit estructural de cualquier cambio de comportamiento futuro

### Caso límite
Given un hot path refactorizado que toca alocaciones
When el agente verifica el cambio
Then corre `-race` y benchmarks con `benchstat`
And exige que los deltas de rendimiento sean estadísticamente significativos

### Caso de error
Given un PR que combina un rename con un fix de bug (lógica cambiada)
When el agente lo revisa
Then detecta la mezcla estructural/conductual
And lo rechaza pidiendo dividirlo en dos PRs separados