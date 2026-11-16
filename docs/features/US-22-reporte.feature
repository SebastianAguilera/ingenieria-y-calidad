# language: es
# =============================================================================
# US-22 - Reporte de proyecto y sprint
# Issue: #109
# Especificación SDD: docs/specs/US-22-reporte-spec.md
# Endpoints:
#   GET /api/proyectos/{id}/reporte
#   GET /api/proyectos/{id}/sprints/{sprintId}/reporte
# =============================================================================

Característica: US-22 - Reporte de proyecto y sprint
  Como líder de un proyecto
  Quiero obtener un reporte consolidado del proyecto o de un sprint
  Para documentar el avance, el esfuerzo y la calidad en una sola consulta

  Antecedentes:
    Dado que la API esta levantada y responde "200 OK" en "GET /health"
    Y que existe un proyecto con historias, Sprints, asignaciones y defectos
    Y que el usuario es miembro del proyecto
    Y que el reporte incluye base "estado_actual"

  @caso_normal
  Escenario: E-01 - Consultar reporte de proyecto completo incluye historias, métricas y defectos
    Dado que existe un proyecto con historias y defectos
    Cuando envio "GET" a "/api/proyectos/1/reporte"
    Entonces la respuesta tiene el estado HTTP "200"
    Y el cuerpo contiene "base" igual a "estado_actual"
    Y el cuerpo contiene la lista de historias del proyecto
    Y el cuerpo contiene las métricas del proyecto
    Y el cuerpo contiene la lista de defectos del proyecto

  @caso_normal
  Escenario: E-02 - Consultar reporte de sprint incluye solo historias asignadas y defectos del sprint
    Dado que existe un proyecto con el Sprint 10 y asignaciones y defectos
    Cuando envio "GET" a "/api/proyectos/1/sprints/10/reporte"
    Entonces la respuesta tiene el estado HTTP "200"
    Y el cuerpo contiene "base" igual a "estado_actual"
    Y el cuerpo contiene solo las historias asignadas al Sprint 10
    Y el cuerpo contiene las métricas del Sprint 10
    Y el cuerpo contiene solo los defectos detectados o resueltos en el Sprint 10

  @caso_limite
  Escenario: E-03 - Consultar reporte de proyecto sin historias responde 200 con listas vacías y no null
    Dado que existe un proyecto sin historias ni defectos
    Cuando envio "GET" a "/api/proyectos/1/reporte"
    Entonces la respuesta tiene el estado HTTP "200"
    Y el cuerpo contiene listas vacías para historias y defectos

  @caso_limite
  Escenario: E-04 - Historia sin horas estimadas muestra horas_estimadas como null
    Dado que existe un proyecto con una historia sin horas estimadas
    Cuando envio "GET" a "/api/proyectos/1/reporte"
    Entonces la respuesta tiene el estado HTTP "200"
    Y la historia muestra "horas_estimadas" igual a null

  @caso_error
  Escenario: E-05 - Usuario no miembro consulta reporte de proyecto recibe 404 proyecto no encontrado
    Dado que el usuario no es miembro del proyecto 1
    Cuando envio "GET" a "/api/proyectos/1/reporte"
    Entonces la respuesta tiene el estado HTTP "404"
    Y el cuerpo contiene el error "PROYECTO_NO_ENCONTRADO"

  @caso_error
  Escenario: E-06 - Consultar reporte de sprint inexistente o de otro proyecto recibe 404 sprint no encontrada
    Dado que el Sprint 999 no existe en el proyecto 1
    Cuando envio "GET" a "/api/proyectos/1/sprints/999/reporte"
    Entonces la respuesta tiene el estado HTTP "404"
    Y el cuerpo contiene el error "SPRINT_NO_ENCONTRADA"
