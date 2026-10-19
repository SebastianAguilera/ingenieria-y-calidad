# Guía de Git, GitHub y Metodologías de Trabajo

> Reglas obligatorias del proyecto: `AGENTS.md`. Este documento describe el flujo operativo y resume lo que el agente debe recordar al encarar una tarea.

## Flujo de Trabajo General

```
master (rama por defecto / producción)
  └── develop (integración)
       ├── feature/<tarea-1>
       ├── feature/<tarea-2>
       └── ...
```

---

## 1. Configuración Inicial (una sola vez por miembro)

```bash
# Configurar nombre y email
git config --global user.name "Tu Nombre"
git config --global user.email "tu@email.com"

# Clonar el repositorio
git clone https://github.com/org/repo.git
cd repo

# Configurar remote si no existe
git remote add origin https://github.com/org/repo.git
```

---

## 2. Día a Día con Scrum

### 2.1 Al inicio del Sprint

1. El **Agile Enabler** (o quien coordine el Sprint) crea un Sprint en **GitHub Projects** con nombre tipo `Sprint 2 - 01/09 al 15/09`.
2. Se definen las **User Stories** del sprint y se asignan.
3. Cada User Story se crea como **Issue** en GitHub con:

```markdown
Título: [US-XXX] Como usuario quiero registrar mis datos

## Criterios de Aceptación
- [ ] El formulario valida campos obligatorios
- [ ] Se muestra mensaje de éxito al guardar
- [ ] Se redirige al dashboard

## Definición de Hecho
- [ ] Código escrito y funcionando
- [ ] Tests pasando
- [ ] Code review aprobado
- [ ] Merge a develop
```

4. La Issue se vincula al **Project** y se asigna al miembro responsable.

### 2.2 Standup Diario

- Cada miembro reporta en el canal del equipo:
  - Qué hizo ayer
  - Qué va a hacer hoy
  - Bloqueos
- No se usa git para esto, solo comunicación.

### 2.3 Revisión y Retrospectiva

- Al final del sprint, revisar PRs mergeados y Issues cerradas en GitHub Projects.
- Usar la vista de **Board** para ver el flujo de trabajo.
- Dejar acta breve de la retrospectiva: qué funcionó, qué no y qué acciones se toman.

---

## 3. Trabajar en una User Story (paso a paso)

### Paso 1: Actualizar develop

```bash
git checkout develop
git pull origin develop
```

### Paso 2: Crear la rama de la feature

```bash
# Formato: feature/US-XXX-descripcion-corta
git checkout -b feature/US-001-registro-usuario
```

Mover la Issue a **In Progress** en el tablero.

### Paso 3: Especificar, probar y recién después codificar

Las reglas obligatorias están en `AGENTS.md`: SDD antes de implementar, BDD en formato Given/When/Then y TDD con ciclo **RED → GREEN → REFACTOR**.

Con opencode, trabajar siempre en **modo Plan** y luego en **modo Build**:

1. **Tarea asignada**: `@github-helper` lista las issues del Sprint y su estado.
2. **Especificación (SDD)**: `@spec-writer` crea `docs/specs/US-XX-descripcion-spec.md` con objetivo, entradas, salidas, reglas de negocio, restricciones, casos límite, condiciones de error y criterios de aceptación.
3. **Escenarios (BDD)**: `@bdd-scenarios` escribe `features/US-XX-descripcion.feature` con casos normales, alternativos, límite y error. Cada escenario lleva un ID `E-nn` que se reutiliza como nombre de subtest en Go.
4. **Pruebas (TDD - RED)**: `@test-writer` escribe los `*_test.go` que fallan, anotando cada caso con el `E-nn` que lo origina.
5. **Implementación (GREEN)**: recién ahora se escribe el código de dominio, service, repositorio y handler.
6. **Refactor**: mejorar sin romper tests.
7. **Controles previos al PR**: `@inspector` (calidad y capas), `@migration-checker` (si hay cambios de esquema), `@traceability-checker` (Issue → Spec → BDD → Tests → Código), `@coverage-reporter` (cobertura) y `@doc-writer` (documentación y manual).

Toda salida generada con IA debe revisarse y entenderse antes de commitear. El equipo es responsable del código aunque lo haya generado un agente. No incluir secretos ni `.env` en prompts ni en commits.

Ver sección 4 para el detalle de cada metodología.

### Paso 4: Commits frecuentes y descriptivos

```bash
# Formato del mensaje:
# <tipo>: <descripción corta>
#
# Tipos: feat, fix, test, docs, refactor, chore

git add -A
git commit -m "feat: agregar formulario de registro"
git commit -m "test: agregar tests de validación de formulario"
git commit -m "fix: corregir bug en envío de email"
```

### Paso 5: Empujar y crear Pull Request

```bash
git push origin feature/US-001-registro-usuario
```

Luego en GitHub:
1. Crear PR de `feature/US-001-registro-usuario` → `develop`.
2. Título claro: `[US-XXX] Registro de usuario`.
3. Descripción con resumen del cambio, link a la Issue y `Closes #XX`.
4. Asignar **1 reviewer** mínimo.
5. Mover la Issue a **In Review**.

### Paso 6: Code Review

El reviewer:
- Revisa el código, tests y convenciones.
- Comenta sugerencias inline.
- Aprueba o pide cambios.
- **Regla**: no mergeear tu propio PR.
- El PR solo se mergea con el CI en verde.

### Paso 7: Merge

```bash
# Después de aprobar el PR, el autor o el Agile Enabler mergea
# Usar "Squash and Commit" si hay muchos commits menores
# O "Merge Commit" para preservar historial completo
```

### Paso 8: Limpiar rama local

```bash
git checkout develop
git pull origin develop
git branch -d feature/US-001-registro-usuario
git push origin --delete feature/US-001-registro-usuario
```

Mover la Issue a **Done**.

---

## 4. Metodologías: TDD, BDD, SDD

### TDD (Test-Driven Development)

Ciclo: **Red → Green → Refactor**

```bash
# 1. RED: Escribir test que falle
# Crear archivo: internal/service/usuario_test.go
# Escribir test para la función que aún no existe

# 2. GREEN: Escribir la mínima implementación que pase el test
# Crear archivo: internal/service/usuario.go

# 3. REFACTOR: Mejorar código sin romper tests
```

La evidencia del ciclo queda en el historial: un commit `test:` con la prueba en rojo y luego un commit `feat:` con la implementación que la pone en verde.

**Estructura de test Go:**

```go
func TestCrearUsuario(t *testing.T) {
    tests := []struct {
        nombre    string
        input     Usuario
        wantErr   bool
    }{
        {
            nombre:  "usuario válido",
            input:   Usuario{Nombre: "Juan", Email: "juan@mail.com"},
            wantErr: false,
        },
        {
            nombre:  "email vacío",
            input:   Usuario{Nombre: "Juan", Email: ""},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.nombre, func(t *testing.T) {
            // ...
        })
    }
}
```

**Ejecutar tests:**

```bash
# Todos los tests
go test ./...

# Test específico
go test ./internal/service/ -run TestCrearUsuario

# Con verbose
go test -v ./...

# Con cobertura
go test -race -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

### BDD (Behavior-Driven Development)

Los escenarios viven en `features/US-XX-descripcion.feature` con `# language: es`. No se usa Godog: cada escenario se automatiza como subtest de Go y su ID `E-nn` se reutiliza en `t.Run()`.

```gherkin
# language: es
# features/US-XX-descripcion.feature

Funcionalidad: Gestión de usuarios

  Regla: Registro de usuarios

    @caso_normal
    Escenario: E-01 - Registrar usuario válido
      Dado un usuario con nombre "Juan" y email "juan@mail.com"
      Cuando registro el usuario
      Entonces el usuario se guarda exitosamente

    @caso_error
    Escenario: E-02 - Registrar usuario con email duplicado
      Dado un usuario existente con email "juan@mail.com"
      Cuando intento registrar otro usuario con email "juan@mail.com"
      Entonces recibo un error de email duplicado
```

Cada funcionalidad debe cubrir, según corresponda: caso normal, alternativo, límite y error.

### SDD (Specification-Driven Development)

Cada funcionalidad se especifica antes de implementar en `docs/specs/US-XX-descripcion-spec.md`, versionada junto con el repo. La spec debe incluir como mínimo:

1. Objetivo y alcance.
2. Entradas: precondiciones, formato y cuerpos de request.
3. Salidas esperadas: endpoints, códigos HTTP y representaciones JSON.
4. Reglas de negocio.
5. Restricciones y decisiones de diseño.
6. Casos límite.
7. Condiciones de error y mapeo a respuestas HTTP.
8. Criterios de aceptación.
9. Modelo de datos y plan de pruebas.

Tomar como referencia `docs/specs/US-02-gestion-product-backlog-spec.md`.

---

## 5. Entorno local (Docker y base de datos)

```bash
# 1. Configuración (una sola vez)
cp .env.example .env
cp docker/.env.example docker/.env
docker network create mired   # solo si la red no existe

# 2. Base de datos
cd docker
docker compose up -d

# 3. API (desde la raíz del repo)
go run cmd/api/main.go

# 4. Consola de PostgreSQL
docker exec -it postgresql-servidor psql -U postgres -d metrics_db
```

- La app local lee el `.env` de la raíz y se conecta a `localhost:5433`, el puerto que publica el compose.
- `APP_ENV` selecciona la base: `development` → `DB_NAME_DEV`, `test` → `DB_NAME_TEST`, `production` → `DB_NAME_PROD`. El nombre debe coincidir con `POSTGRES_DB` del compose.
- El esquema se crea al arrancar mediante `AutoMigrate`; no hay migraciones manuales.
- No commitear `.env` ni credenciales.

---

## 6. GitHub Actions (CI/CD)

El workflow real es `.github/workflows/go.yml` y se dispara en `push` y `pull_request` contra `master` y `develop`.

Pasos del job `build`:

1. `gofmt -l .`: falla si hay archivos sin formatear.
2. `go vet ./...`.
3. `go build ./...`.
4. `go test -race -covermode=atomic -coverprofile=coverage.out ./...` con un servicio PostgreSQL 15.4.
5. Verifica que los tests de integración de `internal/repository` no se salten (falla si aparece `--- SKIP`).
6. Publica `coverage.out` como artefacto.

Comandos equivalentes en local:

```bash
gofmt -l .
go vet ./...
go build ./...
go test -race -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

Los tests de integración necesitan `DB_HOST`; sin base de datos se saltan y el CI lo considera un error.

### Regla del repositorio

Configurar en **Settings → Branches → Branch protection rules** para `develop` y `master`:

- ✅ Require pull request before merging
- ✅ Require approvals (1 mínimo)
- ✅ Require status checks to pass (job `build` de `go.yml`)
- ✅ Require branches to be up to date
- ❌ Allow force pushes: deshabilitado
- ❌ Allow deletions: deshabilitado

---

## 7. GitHub Projects - Configuración

### Columnas del Board

| Columna        | Descripción                          |
|----------------|--------------------------------------|
| Backlog        | Issues sin priorizar aún             |
| Sprint Backlog | Issues seleccionadas para el sprint  |
| En Progreso    | Rama activa, trabajo en curso        |
| En Review      | PR abierto esperando review          |
| Listo para QA  | Mergeado a develop, testeando        |
| Done           | Verificado y funcionando en develop  |

### Labels sugeridas

| Label         | Color   | Uso                        |
|---------------|---------|----------------------------|
| `US`          | verde   | User Story                 |
| `bug`         | rojo    | Error reportado            |
| `feature`     | azul    | Nueva funcionalidad        |
| `docs`        | gris    | Documentación              |
| `urgent`      | naranja | Prioridad alta             |
| `blocked`     | negro   | No se puede avanzar        |

### Etiquetas de ramas (Issues)

- Asignar **sprint** como Milestone en la Issue.
- Usar **Assignee** para indicar quién trabaja en ello.

---

## 8. Comandos de Uso Frecuente

```bash
# Ver estado actual
git status
git log --oneline -10

# Sincronizar con remoto
git fetch origin
git pull origin develop

# Ver ramas existentes
git branch -a

# Cambiar a otra rama
git checkout feature/US-002-otro-cambio

# Ver diferencia de tu rama contra develop
git diff develop

# Ver historial de una rama específica
git log --oneline develop..feature/US-001-registro-usuario

# Cancelar último commit (sin perder cambios)
git reset --soft HEAD~1

# Stash: guardar cambios temporalmente
git stash
git stash pop

# Ver archivos modificados en un PR
git diff --name-only develop...feature/US-001-registro-usuario
```

---

## 9. Reglas del Equipo

1. **Nunca hacer push directo a `develop` o `master`**. Siempre PR.
2. **Commits pequeños y descriptivos**. Un commit = un cambio lógico.
3. **Tests antes de crear PR**. CI debe pasar.
4. **1 reviewer mínimo** antes de mergeear.
5. **Asignarse Issues** antes de empezar a trabajar.
6. **Mover la Issue en el Board** según el estado del trabajo.
7. **Cerrar la Issue** desde el PR con `Closes #XX` en la descripción.
8. **No trabajar en develop directamente**. Siempre crear feature branch.
9. **Pull de develop antes de crear feature branch** para tener últimos cambios.
10. **Eliminar ramas locales y remotas** después de mergeear.
11. **No commitear `.env` ni secretos**.
12. **Revisar y entender toda salida de IA** antes de commitear.
13. **No mergear si los tests de integración se saltaron** o el CI está en rojo.

---

## 10. Resumen: Flujo Completo de una User Story

```
1.  git checkout develop && git pull origin develop
2.  git checkout -b feature/US-XXX-descripcion-corta
3.  Mover la Issue a In Progress
4.  @spec-writer: docs/specs/US-XX-descripcion-spec.md
5.  @bdd-scenarios: features/US-XX-descripcion.feature (E-nn)
6.  @test-writer: tests en rojo (TDD - RED)
7.  Implementar la funcionalidad (TDD - GREEN)
8.  Refactorizar (TDD - REFACTOR)
9.  @inspector, @migration-checker, @traceability-checker, @coverage-reporter, @doc-writer
10. gofmt, go vet, go build y go test; probar los endpoints
11. git add -A && git commit -m "feat: ..."
12. git push origin feature/US-XXX-descripcion-corta
13. Abrir PR en GitHub → develop con Closes #XX y mover la Issue a In Review
14. Esperar review y que el job build de go.yml pase en verde
15. Mergeear (squash o merge commit)
16. git checkout develop && git pull origin develop
17. git branch -d feature/US-XXX-descripcion-corta
18. git push origin --delete feature/US-XXX-descripcion-corta
19. Mover la Issue a Done
```
