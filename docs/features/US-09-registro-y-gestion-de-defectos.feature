# language: es
# =============================================================================
# US-09 - Registro y Gestión de Defectos (Bugs)
# Issue: #45
# Especificación SDD: docs/specs/US-09-registro-y-gestion-de-defectos-spec.md
# Trazabilidad: Issue #45 -> SDD -> Criterios de Aceptación -> Escenarios BDD -> Tests Go -> Código
# =============================================================================

Funcionalidad: US-09 - Registro y Gestión de Defectos
  Como integrante del equipo
  Quiero registrar y actualizar defectos asociados a una historia y a un Sprint
  Para controlar la calidad del software y alimentar las métricas de defectos

  Contexto:
    Dado que la API esta levantada y responde "200 OK" en "GET /health"
    Y que la base de datos fue migrada correctamente
    Y que existe un proyecto con historias y Sprints

  # ===========================================================================
  # REGISTRO: POST /api/proyectos/{id}/defectos
  # ===========================================================================

  Regla: Registro de defectos con estado inicial abierto

    @caso_normal
    Escenario: BDD-01 - Registro exitoso de un defecto
      Dado que existe una historia de usuario y un Sprint de detección válidos
      Cuando envio "POST" a "/api/proyectos/1/defectos" con {"descripcion": "Error en validación", "severidad": "major", "historia_id": 1, "sprint_deteccion_id": 1}
      Entonces la respuesta tiene el estado HTTP "201"
      Y el defecto registra "estado" igual a "open"
      Y el defecto registra "severidad" igual a "major"

    @caso_error
    Escenario: BDD-02 - Severidad inválida
      Dado que existe una historia de usuario y un Sprint de detección válidos
      Cuando envio "POST" a "/api/proyectos/1/defectos" con {"descripcion": "Error", "severidad": "urgente", "historia_id": 1, "sprint_deteccion_id": 1}
      Entonces la respuesta tiene el estado HTTP "422"
      Y el cuerpo contiene el error "VALIDACION" para el campo "severidad"
      Y no se registra el defecto

    @caso_error
    Escenario: BDD-03 - Descripción vacía
      Cuando envio "POST" a "/api/proyectos/1/defectos" con {"descripcion": "   ", "severidad": "minor", "historia_id": 1, "sprint_deteccion_id": 1}
      Entonces la respuesta tiene el estado HTTP "422"
      Y el cuerpo contiene el error "VALIDACION" para el campo "descripcion"

    @caso_error
    Escenario: BDD-04 - Historia inexistente
      Cuando envio "POST" a "/api/proyectos/1/defectos" con {"descripcion": "Error", "severidad": "major", "historia_id": 99, "sprint_deteccion_id": 1}
      Entonces la respuesta tiene el estado HTTP "404"
      Y el cuerpo contiene el error "HISTORIA_NO_ENCONTRADA"

    @caso_error
    Escenario: BDD-05 - Sprint de detección inexistente
      Cuando envio "POST" a "/api/proyectos/1/defectos" con {"descripcion": "Error", "severidad": "major", "historia_id": 1, "sprint_deteccion_id": 99}
      Entonces la respuesta tiene el estado HTTP "404"
      Y el cuerpo contiene el error "SPRINT_NO_ENCONTRADA"

    @caso_error
    Escenario: BDD-06 - Usuario reportante inexistente
      Cuando envio "POST" a "/api/proyectos/1/defectos" con {"descripcion": "Error", "severidad": "major", "historia_id": 1, "sprint_deteccion_id": 1, "reportado_por": 99}
      Entonces la respuesta tiene el estado HTTP "404"
      Y el cuerpo contiene el error "USUARIO_NO_ENCONTRADO"

  # ===========================================================================
  # ACTUALIZACIÓN: PATCH /api/proyectos/{id}/defectos/{defectoId}
  # ===========================================================================

  Regla: Actualización de estado y resolución del defecto

    @caso_normal
    Escenario: BDD-07 - Actualizar a en progreso
      Dado que existe un defecto en estado "open"
      Cuando envio "PATCH" a "/api/proyectos/1/defectos/1" con {"estado": "in_progress"}
      Entonces la respuesta tiene el estado HTTP "200"
      Y el defecto registra "estado" igual a "in_progress"

    @caso_normal
    Escenario: BDD-08 - Resolver un defecto indicando Sprint de resolución
      Dado que existe un defecto en estado "open"
      Y que existe un Sprint de resolución válido
      Cuando envio "PATCH" a "/api/proyectos/1/defectos/1" con {"estado": "resolved", "sprint_resolucion_id": 2}
      Entonces la respuesta tiene el estado HTTP "200"
      Y el defecto registra "estado" igual a "resolved"
      Y el defecto registra el Sprint de resolución
      Y se registra la fecha de resolución

    @caso_normal
    Escenario: BDD-09 - Cerrar un defecto indicando Sprint de resolución
      Dado que existe un defecto en estado "resolved"
      Cuando envio "PATCH" a "/api/proyectos/1/defectos/1" con {"estado": "closed", "sprint_resolucion_id": 2}
      Entonces la respuesta tiene el estado HTTP "200"
      Y el defecto registra "estado" igual a "closed"

    @caso_error
    Escenario: BDD-10 - Resolver sin Sprint de resolución
      Dado que existe un defecto en estado "open"
      Cuando envio "PATCH" a "/api/proyectos/1/defectos/1" con {"estado": "resolved"}
      Entonces la respuesta tiene el estado HTTP "422"
      Y el cuerpo contiene el error "VALIDACION" para el campo "sprint_resolucion_id"

    @caso_error
    Escenario: BDD-11 - Informar Sprint de resolución con estado abierto
      Dado que existe un defecto en estado "open"
      Cuando envio "PATCH" a "/api/proyectos/1/defectos/1" con {"estado": "open", "sprint_resolucion_id": 2}
      Entonces la respuesta tiene el estado HTTP "422"
      Y el cuerpo contiene el error "VALIDACION" para el campo "sprint_resolucion_id"

    @caso_error
    Escenario: BDD-12 - Estado inválido
      Cuando envio "PATCH" a "/api/proyectos/1/defectos/1" con {"estado": "fixed"}
      Entonces la respuesta tiene el estado HTTP "400"
      Y el cuerpo contiene el error "ESTADO_INVALIDO" para el campo "estado"

    @caso_error
    Escenario: BDD-13 - Actualización sin campos
      Cuando envio "PATCH" a "/api/proyectos/1/defectos/1" con {}
      Entonces la respuesta tiene el estado HTTP "422"
      Y el cuerpo contiene el error "VALIDACION"

    @caso_error
    Escenario: BDD-14 - Defecto inexistente
      Cuando envio "PATCH" a "/api/proyectos/1/defectos/99" con {"estado": "in_progress"}
      Entonces la respuesta tiene el estado HTTP "404"
      Y el cuerpo contiene el error "DEFECTO_NO_ENCONTRADO"

    @caso_limite
    Escenario: BDD-15 - Reabrir un defecto resuelto
      Dado que existe un defecto resuelto con Sprint de resolución
      Cuando envio "PATCH" a "/api/proyectos/1/defectos/1" con {"estado": "open"}
      Entonces la respuesta tiene el estado HTTP "200"
      Y el defecto ya no registra Sprint de resolución
      Y el defecto ya no registra fecha de resolución

  # ===========================================================================
  # LISTADO: GET /api/proyectos/{id}/defectos
  # ===========================================================================

  Regla: Consulta de defectos del proyecto

    @caso_normal
    Escenario: BDD-16 - Listar defectos del proyecto
      Dado que existen defectos registrados en el proyecto
      Cuando envio "GET" a "/api/proyectos/1/defectos"
      Entonces la respuesta tiene el estado HTTP "200"
      Y el cuerpo contiene la lista de defectos con su total

    @caso_limite
    Escenario: BDD-17 - Listar proyecto sin defectos
      Dado que el proyecto no tiene defectos registrados
      Cuando envio "GET" a "/api/proyectos/1/defectos"
      Entonces la respuesta tiene el estado HTTP "200"
      Y el cuerpo contiene {"total": 0, "defectos": []}
