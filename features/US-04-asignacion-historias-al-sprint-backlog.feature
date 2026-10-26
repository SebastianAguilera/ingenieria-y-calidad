# language: es
# =============================================================================
# US-04 - Asignación de historias al Sprint Backlog
# Issue: #31
# Especificación SDD: docs/specs/US-04-asignacion-historias-al-sprint-backlog-spec.md
# Trazabilidad: Issue #31 -> SDD -> Criterios de Aceptación -> Escenarios BDD -> Tests Go -> Código
# =============================================================================

Funcionalidad: US-04 - Asignación de historias al Sprint Backlog
  Como Product Builder
  Quiero asignar historias del Product Backlog a un Sprint activo
  Para planificar el trabajo del Sprint

  Contexto:
    Dado que la API esta levantada y responde "200 OK" en "GET /health"
    Y que la base de datos fue migrada correctamente
    Y que existen proyectos, historias en estado "Backlog" y Sprints activos

  # ===========================================================================
  # ASIGNAR HISTORIA AL SPRINT BACKLOG: POST /api/proyectos/{id}/sprints/{sprintId}/historias
  # ===========================================================================

  Regla: Asignación de historias a un Sprint activo

    @caso_normal
    Escenario: BDD-01 - Asignar historia pendiente a sprint activo
      Dado que existe un proyecto con una historia en estado "Backlog"
      Y que existe un Sprint en estado "Activo" dentro de su ventana temporal
      Cuando envio "POST" a "/api/proyectos/1/sprints/1/historias" con {"historia_id": 1}
      Entonces la respuesta tiene el estado HTTP "204"
      Y la historia ya no esta disponible para otros Sprints

    @caso_limite
    Escenario: BDD-02 - Asignar primera historia al Sprint Backlog vacio
      Dado que existe un Sprint "Activo" sin historias asignadas
      Y que existe una historia en estado "Backlog"
      Cuando envio "POST" a "/api/proyectos/1/sprints/1/historias" con {"historia_id": 1}
      Entonces la respuesta tiene el estado HTTP "204"
      Y la historia se asigna en la primera posición del Sprint Backlog

    @caso_error
    Escenario: BDD-03 - Historia ya asignada a otro Sprint
      Dado que existe una historia ya asignada a otro Sprint
      Cuando envio "POST" a "/api/proyectos/1/sprints/1/historias" con {"historia_id": 1}
      Entonces la respuesta tiene el estado HTTP "409"
      Y el cuerpo contiene el error "HISTORIA_YA_ASIGNADA"

    @caso_error
    Escenario: BDD-04 - Historia con estado distinto a Backlog
      Dado que existe una historia en estado "En Progreso"
      Y que existe un Sprint "Activo" dentro de su ventana temporal
      Cuando envio "POST" a "/api/proyectos/1/sprints/1/historias" con {"historia_id": 1}
      Entonces la respuesta tiene el estado HTTP "400"
      Y el cuerpo contiene el error "HISTORIA_NO_PENDIENTE"

    @caso_error
    Escenario: BDD-05 - Asignar a Sprint no activo
      Dado que existe una historia en estado "Backlog"
      Y que existe un Sprint en estado "Planificado"
      Cuando envio "POST" a "/api/proyectos/1/sprints/1/historias" con {"historia_id": 1}
      Entonces la respuesta tiene el estado HTTP "400"
      Y el cuerpo contiene el error "SPRINT_NO_ACTIVO"

    @caso_error
    Escenario: BDD-05b - Asignar a Sprint cerrado
      Dado que existe una historia en estado "Backlog"
      Y que existe un Sprint en estado "Cerrado"
      Cuando envio "POST" a "/api/proyectos/1/sprints/1/historias" con {"historia_id": 1}
      Entonces la respuesta tiene el estado HTTP "400"
      Y el cuerpo contiene el error "SPRINT_NO_ACTIVO"

    @caso_error
    Escenario: BDD-09a - Asignar fuera de ventana temporal (antes del inicio)
      Dado que existe una historia en estado "Backlog"
      Y que existe un Sprint "Activo" con fecha de inicio futura
      Cuando envio "POST" a "/api/proyectos/1/sprints/1/historias" con {"historia_id": 1}
      Entonces la respuesta tiene el estado HTTP "400"
      Y el cuerpo contiene el error "FUERA_DE_VENTANA_TEMPORAL"

    @caso_error
    Escenario: BDD-09b - Asignar fuera de ventana temporal (después del fin)
      Dado que existe una historia en estado "Backlog"
      Y que existe un Sprint "Activo" con fecha de fin pasada
      Cuando envio "POST" a "/api/proyectos/1/sprints/1/historias" con {"historia_id": 1}
      Entonces la respuesta tiene el estado HTTP "400"
      Y el cuerpo contiene el error "FUERA_DE_VENTANA_TEMPORAL"

    @caso_normal
    Escenario: BDD-08 - Se registra fecha de asignacion
      Dado que existe una historia en estado "Backlog"
      Y que existe un Sprint "Activo" dentro de su ventana temporal
      Cuando envio "POST" a "/api/proyectos/1/sprints/1/historias" con {"historia_id": 1}
      Entonces la respuesta tiene el estado HTTP "204"
      Y se registra la fecha y hora de asignación

  # ===========================================================================
  # DESASIGNAR HISTORIA DEL SPRINT BACKLOG: DELETE /api/proyectos/{id}/sprints/{sprintId}/historias/{historiaId}
  # ===========================================================================

  Regla: Desasignación de historias

    @caso_normal
    Escenario: BDD-06 - Desasignar historia vuelve al Product Backlog
      Dado que existe una historia asignada a un Sprint "Activo"
      Cuando envio "DELETE" a "/api/proyectos/1/sprints/1/historias/1"
      Entonces la respuesta tiene el estado HTTP "204"
      Y la historia vuelve al Product Backlog en estado "Backlog"

    @caso_error
    Escenario: BDD-07 - No se puede desasignar de un Sprint cerrado
      Dado que existe una historia asignada a un Sprint "Cerrado"
      Cuando envio "DELETE" a "/api/proyectos/1/sprints/1/historias/1"
      Entonces la respuesta tiene el estado HTTP "409"
      Y el cuerpo contiene el error "SPRINT_CERRADO"
