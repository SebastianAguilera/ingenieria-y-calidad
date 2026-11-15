# language: es
# =============================================================================
# US-13 - Consultar las historias asignadas a un sprint
# Issue: #71
# Especificación SDD: docs/specs/US-13-consulta-sprint-backlog-spec.md
# Trazabilidad: Issue #71 -> SDD -> Criterios de Aceptación -> Escenarios BDD -> Tests Go -> Código
# =============================================================================

Funcionalidad: US-13 - Consultar las historias asignadas a un sprint
  Como miembro de un proyecto
  Quiero consultar las historias asignadas a un Sprint
  Para ver el trabajo planificado y el estado de cada historia

  Contexto:
    Dado que la API esta levantada y responde "200 OK" en "GET /health"
    Y que la base de datos fue migrada correctamente

  # ===========================================================================
  # CONSULTA DEL SPRINT BACKLOG: GET /api/proyectos/{id}/sprints/{sprintId}/historias
  # ===========================================================================

  Regla: Consulta del Sprint Backlog por cualquier miembro

    @caso_normal
    Escenario: E-01 - Consultar un Sprint con historias asignadas
      Dado que existe un proyecto con un Sprint y dos historias asignadas
      Cuando envio "GET" a "/api/proyectos/1/sprints/1/historias"
      Entonces la respuesta tiene el estado HTTP "200"
      Y el cuerpo contiene el Sprint solicitado
      Y "total" es 2
      Y cada item incluye "orden", "fecha_asignacion" e "historia"

    @caso_limite
    Escenario: E-02 - Sprint sin historias asignadas
      Dado que existe un Sprint sin historias asignadas
      Cuando envio "GET" a "/api/proyectos/1/sprints/1/historias"
      Entonces la respuesta tiene el estado HTTP "200"
      Y "total" es 0
      Y "items" es un arreglo vacio, no null

    @caso_normal
    Escenario: E-03 - Las historias respetan el orden del Sprint Backlog
      Dado que existe un Sprint con historias en orden 2 y 1
      Cuando envio "GET" a "/api/proyectos/1/sprints/1/historias"
      Entonces los items se devuelven por "orden" ascendente

    @caso_limite
    Escenario: E-04 - Items con el mismo orden se desempatan por id
      Dado que existe un Sprint con dos items con el mismo "orden"
      Cuando envio "GET" a "/api/proyectos/1/sprints/1/historias"
      Entonces los items se desempatan por "id" ascendente

    @caso_normal
    Escenario: E-05 - Consultar un Sprint cerrado
      Dado que existe un Sprint en estado "Cerrado" con historias
      Cuando envio "GET" a "/api/proyectos/1/sprints/1/historias"
      Entonces la respuesta tiene el estado HTTP "200"
      Y las historias mantienen su estado y asociacion

    @caso_error
    Escenario: E-06 - Sprint inexistente
      Dado que no existe el Sprint solicitado
      Cuando envio "GET" a "/api/proyectos/1/sprints/99/historias"
      Entonces la respuesta tiene el estado HTTP "404"
      Y el cuerpo contiene el error "SPRINT_NO_ENCONTRADA"

    @caso_error
    Escenario: E-07 - Sprint de otro proyecto
      Dado que existe un Sprint que pertenece a otro proyecto
      Cuando envio "GET" a "/api/proyectos/1/sprints/2/historias"
      Entonces la respuesta tiene el estado HTTP "404"
      Y el cuerpo contiene el error "SPRINT_NO_ENCONTRADA"

    @caso_error
    Escenario: E-08 - Proyecto inexistente
      Dado que no existe el proyecto solicitado
      Cuando envio "GET" a "/api/proyectos/99/sprints/1/historias"
      Entonces la respuesta tiene el estado HTTP "404"
      Y el cuerpo contiene el error "PROYECTO_NO_ENCONTRADO"

    @caso_error
    Escenario: E-09 - Usuario que no es miembro del proyecto
      Dado que el usuario autenticado no es miembro del proyecto
      Cuando envio "GET" a "/api/proyectos/1/sprints/1/historias"
      Entonces la respuesta tiene el estado HTTP "404"
      Y el cuerpo contiene el error "PROYECTO_NO_ENCONTRADO"

    @caso_error
    Escenario: E-10 - Identificador de proyecto invalido
      Cuando envio "GET" a "/api/proyectos/abc/sprints/1/historias"
      Entonces la respuesta tiene el estado HTTP "400"
      Y el cuerpo contiene el error "PARAMETRO_INVALIDO"

    @caso_error
    Escenario: E-11 - Identificador de sprint invalido
      Cuando envio "GET" a "/api/proyectos/1/sprints/abc/historias"
      Entonces la respuesta tiene el estado HTTP "400"
      Y el cuerpo contiene el error "PARAMETRO_INVALIDO"

    @caso_alternativo
    Escenario: E-12 - El listado refleja asignaciones y desasignaciones de US-04
      Dado que se asigna una historia a un Sprint
      Cuando consulto su Sprint Backlog
      Entonces la historia aparece en el listado
      Cuando la desasigno y vuelvo a consultar
      Entonces la historia ya no aparece en el listado
