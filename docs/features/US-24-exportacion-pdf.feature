# language: es
# =============================================================================
# US-24 - Exportación del reporte a PDF
# Issue: #111
# Especificación SDD: docs/specs/US-24-exportacion-pdf-spec.md
# Dependencia: US-22 (ReporteService), US-F10 (guard de membresía)
# Endpoints:
#   GET /api/proyectos/{id}/reporte.pdf
#   GET /api/proyectos/{id}/sprints/{sprintId}/reporte.pdf
# =============================================================================

Característica: US-24 - Exportación del reporte a PDF
  Como líder de un proyecto
  Quiero descargar el reporte de proyecto o de sprint en PDF
  Para compartirlo y archivarlo sin depender de la API ni del navegador

  Antecedentes:
    Dado que la API está levantada y responde "200 OK" en "GET /health"
    Y que el usuario es miembro del proyecto
    Y que la generación del PDF ocurre en el backend (Go)

  @caso_normal
  Escenario: E-01 - Descargar PDF del proyecto genera archivo con todas las secciones
    Dado que existe un proyecto con historias, métricas y defectos
    Cuando envío "GET" a "/api/proyectos/1/reporte.pdf"
    Entonces la respuesta tiene el estado HTTP "200"
    Y el Content-Type es "application/pdf"
    Y el Content-Disposition contiene "attachment; filename="
    Y el cuerpo es un PDF válido (empieza con "%PDF")
    Y el PDF contiene las secciones "Encabezado", "Metricas", "Historias" y "Defectos"

  @caso_normal
  Escenario: E-02 - Descargar PDF del sprint genera archivo con historias y defectos del sprint
    Dado que existe un proyecto con el Sprint 10 con historias y defectos asignados
    Cuando envío "GET" a "/api/proyectos/1/sprints/10/reporte.pdf"
    Entonces la respuesta tiene el estado HTTP "200"
    Y el Content-Type es "application/pdf"
    Y el PDF contiene las secciones del sprint

  @caso_limite
  Escenario: E-03 - Métrica null se muestra como "Sin datos" en el PDF
    Dado que existe un proyecto sin sprints cerrados (velocidad null)
    Cuando envío "GET" a "/api/proyectos/1/reporte.pdf"
    Entonces la respuesta tiene el estado HTTP "200"
    Y el PDF muestra "Sin datos" para la velocidad, no "0"

  @caso_limite
  Escenario: E-04 - Proyecto sin historias genera PDF con secciones vacías sin error
    Dado que existe un proyecto sin historias ni defectos
    Cuando envío "GET" a "/api/proyectos/1/reporte.pdf"
    Entonces la respuesta tiene el estado HTTP "200"
    Y el PDF se genera correctamente con las secciones vacías

  @caso_limite
  Escenario: E-05 - Nombre del archivo sigue el formato establecido
    Dado que existe un proyecto con nombre "Mi Proyecto"
    Cuando envío "GET" a "/api/proyectos/1/reporte.pdf"
    Entonces el Content-Disposition contiene "reporte-" y ".pdf"

  @caso_error
  Escenario: E-06 - Usuario no miembro recibe 404 al intentar descargar el PDF
    Dado que el usuario no es miembro del proyecto 1
    Cuando envío "GET" a "/api/proyectos/1/reporte.pdf"
    Entonces la respuesta tiene el estado HTTP "404"
    Y el cuerpo contiene el error "PROYECTO_NO_ENCONTRADO"
