# Especificación: SK-006 - golang-continuous-integration: CI/CD con GitHub Actions para Go

## Objetivo
Configurar y mejorar pipelines de CI/CD de nivel productivo para proyectos Go con GitHub Actions: workflows de test, lint, SAST, cobertura y escaneo de vulnerabilidades, archivos de Dependabot/Renovate, releases con GoReleaser, builds Docker multi-plataforma, ajustes de seguridad del repositorio y code review asistido por IA. El pipeline se trata como un quality gate: cada decisión se pondera contra velocidad de build, fiabilidad de la señal y postura de seguridad.

## Entradas
- Proyecto Go con `go.mod` y, cuando existan, workflows previos en `.github/workflows/*.yml`.
- Modo de operación: **Setup** (CI desde cero: orden test → lint → security → release) o **Improve** (auditar pipeline existente y proponer adiciones sin duplicar).
- Binarios y herramientas: `go`, `goreleaser`, `gh`, `skills` (npm), `golangci-lint`.
- Herramientas permitidas: Read, Edit, Write, Glob, Grep, Bash (go, golangci-lint, git, goreleaser, gh), WebFetch, agentes.

## Salidas esperadas
1. `.github/workflows/test.yml` con matrix de versiones de Go acorde al `go` directive del `go.mod`, `-race` y cobertura.
2. Workflows de lint (`golangci-lint`), SAST (`gosec`/CodeQL/Bearer) y escaneo de vulnerabilidades (`govulncheck`) según el Quick Reference.
3. Configuración de Dependabot o Renovate para actualizaciones automáticas.
4. Pipeline de release con GoReleaser (o GoReleaser Action).
5. Build y push de imágenes Docker multi-plataforma (si aplica).
6. Integración de AI review (Claude Code / Copilot) en los PRs.
7. En Improve mode: diagnóstico de gaps contra el Quick Reference y adiciones puntuales sin duplicar pasos existentes.

## Reglas de negocio
1. En Setup mode se generan los workflows en este orden: test → lint → security → release.
2. La matrix de versiones de Go se adapta al `go` directive del `go.mod`.
3. Se prefiere la última versión estable major de cada GitHub Action.
4. Test job usa `-race`; cobertura se reporta con `codecov/codecov-action`.
5. Todo pipeline productivo DEBE incluir lint (`golangci-lint`), vet y escaneo de vulnerabilidades (`govulncheck`).
6. Cada subsection es un artefacto generado para un reviewer específico; las tool names del asset que corre en CI runner son literales.
7. Para control de costos de AI review agents, se pueden remover jobs o restringir el trigger a ramas específicas.
8. Improve mode lee primero los workflows existentes e identifica gaps; no duplica pasos.

## Restricciones
- El skill cablea herramientas en el pipeline; NO interpreta hallazgos de seguridad (→ `golang-security`) ni elige/audita versiones de dependencias (→ `golang-dependency-management`).
- Las acciones de ejemplo pueden estar desactualizadas: verificar la versión major vigente de cada acción.

## Casos límite
- Proyecto con `go 1.23` en `go.mod`: matrix ["1.23"…"1.27","stable"].
- Repositorio con CI existente: no se duplican steps; se agregan los gaps detectados.
- AI review agents costosos: restringir trigger a ramas específicas o eliminar jobs innecesarios.

## Condiciones de error
- Version de una action desactualizada o eliminada: el workflow falla en checkout — usar la versión major vigente documentada por cada acción.
- Job de seguridad que falla por hallazgo: el pipeline se bloquea; el hallazgo se deriva al equipo para tratamiento (el skill no lo interpreta).
- Secrets sin configurar en el repo (tokens de release/coverage): el job correspondiente falla hasta declararlos en el repositorio.

## Criterios de aceptación
- [ ] Existe workflow de test con `-race` y matrix adaptada al `go.mod`.
- [ ] Existen jobs de lint (golangci-lint), vet y `govulncheck`.
- [ ] SAST configurado (gosec/CodeQL/Bearer) al menos para ramas protegidas.
- [ ] Dependabot o Renovate habilitado.
- [ ] Release con GoReleaser y Docker build/push si aplica.
- [ ] En Improve mode no se duplican pasos existentes.

## Escenarios BDD

### Caso normal
Given un proyecto Go nuevo con `go 1.26` en go.mod
When el agente ejecuta Setup mode
Then genera test.yml con matrix ["1.26","1.27","stable"] y `-race`
And genera lint (golangci-lint), SAST y release (GoReleaser) en el orden definido

### Caso alternativo
Given un repo que ya tiene un workflow de test pero sin lint ni escaneo de vulnerabilidades
When el agente ejecuta Improve mode
Then lee el workflow existente y propone solo los jobs faltantes (lint + govulncheck)
And no duplica el job de test existente

### Caso límite
Given un despliegue multi-plataforma que requiere imágenes Docker
When el agente diseña el pipeline
Then agrega `docker/build-push-action` con build multi-plataforma
And lo ubica después de los stages de calidad (test/lint/security)

### Caso de error
Given una branch rule que exige status checks para mergear a develop
When el agente configura CI sin job de lint
Then el pipeline queda incompleto respecto del quality gate
And el agente agrega el job de lint antes de declarar el CI como productivo