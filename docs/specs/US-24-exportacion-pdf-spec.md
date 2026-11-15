# US-24 — Exportación del reporte a PDF

## Objetivo

Permitir que los líderes de proyecto descarguen el reporte generado por US-22 en formato PDF, para compartirlo y archivarlo sin depender de la API ni del navegador.

**Trazabilidad:** Issue #111 → Esta especificación → Criterios de aceptación (CA-01 a CA-07) → Escenarios BDD (`docs/features/US-24-exportacion-pdf.feature`) → Tests Go (TDD) → Código.

| Eslabón | Artefacto |
| :--- | :--- |
| Issue | #111 |
| SDD | `docs/specs/US-24-exportacion-pdf-spec.md` (este archivo) |
| BDD | `docs/features/US-24-exportacion-pdf.feature` (E-01 a E-06) |
| Tests | `internal/service/pdf_service_test.go`, `internal/handler/pdf_handler_test.go` |
| Código | `internal/service/pdf_service.go`, `internal/handler/pdf_handler.go`, `internal/handler/pdf_servicio.go`, `cmd/api/router.go` |

## Decisión de generación: Backend (Go)

| Opción | Ventajas | Costos |
| :--- | :--- | :--- |
| **Backend (Go) — elegida** | Reproducible con `curl`, testeable con `go test`, independiente del navegador y del frontend, `Content-Type: application/pdf` nativo | Librería de Go en el binario; sin CGO; sin servidor externo |
| Frontend (impresión/librería) | Sin cambios en la API | No testeable con `go test`, depende del navegador y del estado de la pantalla; requiere US-F23 implementado |

**Librería elegida:** `github.com/go-pdf/fpdf` v0.9+. Pura Go, sin CGO, con fuentes integradas (Helvetica) que cubren cp1252 (suficiente para español). No requiere archivos externos en producción.

## Dependencias

| Historia | Relación |
| :--- | :--- |
| US-22 | Fuente del reporte (`ReporteService`). No se recalcula ninguna métrica. |
| US-F10 | Guard de membresía: no miembro recibe `404 PROYECTO_NO_ENCONTRADO`. Aplica en el router global (ya implementado). |

## Alcance

**Incluye**
- `GET /api/proyectos/{id}/reporte.pdf`: descarga el PDF del proyecto.
- `GET /api/proyectos/{id}/sprints/{sprintId}/reporte.pdf`: descarga el PDF del sprint.
- `Content-Type: application/pdf` y `Content-Disposition: attachment; filename=…`.
- Nombre del archivo: `reporte-{slug-proyecto}-proyecto-{YYYY-MM-DD}.pdf` y `reporte-{slug-proyecto}-sprint-{slug-id}-{YYYY-MM-DD}.pdf`.
- Secciones del PDF: Encabezado, Métricas, Historias (tabla) y Defectos (tabla).
- Métricas `null` → texto `"Sin datos"` (nunca `0`).
- Listas vacías → sección presente con mensaje `"Sin datos"`.
- Guard de membresía ya aplicado globalmente.
- Tests unitarios (generación pura) e integración (service con repositorios mockeados).

**No incluye**
- Gráficos, firma digital, envío por correo, plantillas personalizables.

## Endpoints

| Endpoint | Método | Respuesta |
| :--- | :---: | :--- |
| `/api/proyectos/{id}/reporte.pdf` | `GET` | `200 application/pdf` ó `404` |
| `/api/proyectos/{id}/sprints/{sprintId}/reporte.pdf` | `GET` | `200 application/pdf` ó `404` |

## Criterios de aceptación

- **CA-01**: Given un proyecto con datos, When se descarga el PDF del proyecto, Then el archivo tiene encabezado, métricas, historias y defectos.
- **CA-02**: Given un sprint, When se descarga su PDF, Then incluye solo sus historias y defectos.
- **CA-03**: Given una métrica `null`, When aparece en el PDF, Then se muestra como `"Sin datos"`, no como `0`.
- **CA-04**: Given un proyecto sin historias, When se descarga, Then el PDF se genera con las secciones vacías y sin error.
- **CA-05**: Given la descarga, When se nombra el archivo, Then sigue el formato `reporte-{slug}-proyecto-{fecha}.pdf` o `reporte-{slug}-sprint-{id}-{fecha}.pdf`.
- **CA-06**: Given un no miembro, When intenta descargar, Then `404 PROYECTO_NO_ENCONTRADO`.
- **CA-07**: Given cualquier descarga, When se completa, Then no se modifica ningún dato.
