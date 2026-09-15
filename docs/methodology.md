# Guía de Git, GitHub y Metodologías de Trabajo

## Flujo de Trabajo General

```
main (producción)
  └── develop (integración)
       ├── feature/user-story-1
       ├── feature/user-story-2
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

1. El **Scrum Master** crea un Sprint en **GitHub Projects** con nombre tipo `Sprint 2 - 01/09 al 15/09`.
2. Se definen las **User Stories** del sprint y se asignan.
3. Cada User Story se crea como **Issue** en GitHub con:

```markdown
Título: [US-001] Como usuario quiero registrar mis datos

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

### Paso 3: Desarrollar con TDD/BDD/SDD

Ver sección 4 más abajo para la metodología.

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
2. Título claro: `[US-001] Registro de usuario`.
3. Descripción con resumen del cambio y link a la Issue.
4. Asignar **1 reviewer** mínimo.
5. Vincular la Issue para que se cierre automáticamente al mergeear.

### Paso 6: Code Review

El reviewer:
- Revisa el código, tests y convenciones.
- Comenta sugerencias inline.
- Aprueba o pide cambios.
- **Regla**: no mergeear tu propio PR.

### Paso 7: Merge

```bash
# Después de aprobar el PR, el autor o SM mergeea
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
```

### BDD (Behavior-Driven Development)

Usar **Godog** para BDD con archivos `.feature`:

```gherkin
# features/usuario.feature
Feature: Gestión de usuarios
  Como administrador
  Quiero registrar usuarios
  Para gestionar el sistema

  Scenario: Registrar usuario válido
    Given un usuario con nombre "Juan" y email "juan@mail.com"
    When registro el usuario
    Then el usuario se guarda exitosamente

  Scenario: Registrar usuario con email duplicado
    Given un usuario existente con email "juan@mail.com"
    When intento registrar otro usuario con email "juan@mail.com"
    Then recibo un error de email duplicado
```

### SDD (Specification-Driven Development)

Documentar especificaciones antes de implementar:

```markdown
# Especificación: US-001 - Registro de Usuario

## Entrada
- Nombre: string, requerido, max 100 chars
- Email: string, requerido, formato válido, único
- Password: string, requerido, min 8 chars

## Salida
- 201 Created con JSON del usuario (sin password)
- 400 Bad Request con mensaje de error
- 409 Conflict si el email ya existe

## Reglas de Negocio
1. El email debe ser único en el sistema
2. El password debe tener al menos 8 caracteres
3. El nombre no puede estar vacío
4. La fecha de creación se genera automáticamente
```

---

## 5. GitHub Actions (CI/CD)

### Workflow para Pull Requests

Crear archivo `.github/workflows/ci.yml`:

```yaml
name: CI

on:
  pull_request:
    branches: [develop, main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.27'

      - name: Download dependencies
        run: go mod download

      - name: Run vet
        run: go vet ./...

      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out ./...

      - name: Build
        run: go build ./cmd/api
```

### Regla del repositorio

Configurar en **Settings → Branches → Branch protection rules** para `develop`:

- ✅ Require pull request before merging
- ✅ Require approvals (1 mínimo)
- ✅ Require status checks to pass (el job `test` de CI)
- ✅ Require branches to be up to date
- ❌ Allow force pushes: deshabilitado
- ❌ Allow deletions: deshabilitado

---

## 6. GitHub Projects - Configuración

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

## 7. Comandos de Uso Frecuente

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

## 8. Reglas del Equipo

1. **Nunca hacer push directo a `develop` o `main`**. Siempre PR.
2. **Commits pequeños y descriptivos**. Un commit = un cambio lógico.
3. **Tests antes de crear PR**. CI debe pasar.
4. **1 reviewer mínimo** antes de mergeear.
5. **Asignarse Issues** antes de empezar a trabajar.
6. **Mover la Issue en el Board** según el estado del trabajo.
7. **Cerrar la Issue** desde el PR con `Closes #XX` en la descripción.
8. **No trabajar en develop directamente**. Siempre crear feature branch.
9. **Pull de develop antes de crear feature branch** para tener últimos cambios.
10. **Eliminar ramas locales y remotas** después de mergeear.

---

## 9. Resumen: Flujo Completo de una User Story

```
1.  git checkout develop && git pull origin develop
2.  git checkout -b feature/US-001-registro-usuario
3.  Escribir especificación (SDD)
4.  Escribir test (TDD - RED)
5.  Implementar función (TDD - GREEN)
6.  Refactorizar (TDD - REFACTOR)
7.  git add -A && git commit -m "feat: ..."
8.  git push origin feature/US-001-registro-usuario
9.  Abrir PR en GitHub → develop
10. Esperar review y que pase CI
11. Mergeear (squash o merge commit)
12. git checkout develop && git pull origin develop
13. git branch -d feature/US-001-registro-usuario
14. git push origin --delete feature/US-001-registro-usuario
15. Mover Issue a "Done" en GitHub Projects
```
