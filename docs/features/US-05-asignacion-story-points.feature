# language: es
# =============================================================================
# US-05 - Asignación inicial de Story Points
# Issue: #32
# Especificación SDD: docs/specs/US-05-asignacion-story-points-spec.md
# Trazabilidad: Issue #32 -> SDD -> Criterios de Aceptación -> Escenarios BDD -> Tests Go -> Código
# =============================================================================

Funcionalidad: US-05 - Asignación manual de Story Points
  Como Product Builder
  Quiero asignar o editar manualmente los Story Points de una historia
  Para tener una estimación base antes de ejecutar el Planning Poker

  Contexto:
    Dado que la API esta levantada y responde "200 OK" en "GET /health"
    Y que la base de datos fue migrada correctamente
    Y que existe un proyecto con una historia de usuario

  Regla: Asignación y actualización con valores de la escala Fibonacci

    @caso_normal
    Escenario: BDD-01 - Asignar Story Points válidos a una historia
      Cuando envio "PATCH" a "/api/proyectos/1/historias/1" con {"story_points": 5}
      Entonces la respuesta tiene el estado HTTP "200"
      Y el detalle de la historia registra "story_points" igual a 5

    @caso_normal
    Escenario: BDD-02 - Actualizar Story Points ya asignados
      Dado que la historia tiene 5 Story Points
      Cuando envio "PATCH" a "/api/proyectos/1/historias/1" con {"story_points": 8}
      Entonces la respuesta tiene el estado HTTP "200"
      Y el detalle de la historia registra "story_points" igual a 8

    @caso_limite
    Escenario: BDD-03 - Asignar cero Story Points
      Cuando envio "PATCH" a "/api/proyectos/1/historias/1" con {"story_points": 0}
      Entonces la respuesta tiene el estado HTTP "200"
      Y el detalle de la historia registra "story_points" igual a 0

    @caso_limite
    Escenario: BDD-04 - Asignar medio punto
      Cuando envio "PATCH" a "/api/proyectos/1/historias/1" con {"story_points": 0.5}
      Entonces la respuesta tiene el estado HTTP "200"
      Y el detalle de la historia registra "story_points" igual a 0.5

  Regla: La estimación es opcional

    @caso_normal
    Escenario: BDD-05 - Crear una historia sin Story Points
      Cuando envio "POST" a "/api/proyectos/1/historias" con {"titulo": "Historia sin estimar"}
      Entonces la respuesta tiene el estado HTTP "201"
      Y el detalle de la historia registra "story_points" igual a null

    @caso_normal
    Escenario: BDD-06 - Crear una historia con Story Points
      Cuando envio "POST" a "/api/proyectos/1/historias" con {"titulo": "Historia estimada", "story_points": 13}
      Entonces la respuesta tiene el estado HTTP "201"
      Y el detalle de la historia registra "story_points" igual a 13

  Regla: Se rechazan valores inválidos

    @caso_error
    Escenario: BDD-07 - Asignar un valor fuera de la escala
      Cuando envio "PATCH" a "/api/proyectos/1/historias/1" con {"story_points": 4}
      Entonces la respuesta tiene el estado HTTP "422"
      Y el cuerpo contiene el error "VALIDACION" para el campo "story_points"

    @caso_error
    Escenario: BDD-08 - Asignar un valor negativo
      Cuando envio "PATCH" a "/api/proyectos/1/historias/1" con {"story_points": -1}
      Entonces la respuesta tiene el estado HTTP "422"
      Y el cuerpo contiene el error "VALIDACION" para el campo "story_points"

    @caso_error
    Escenario: BDD-09 - Asignar un valor no numérico
      Cuando envio "PATCH" a "/api/proyectos/1/historias/1" con {"story_points": "alto"}
      Entonces la respuesta tiene el estado HTTP "422"
      Y el cuerpo contiene el error "VALIDACION" para el campo "story_points"

    @caso_error
    Escenario: BDD-10 - Crear una historia con Story Points inválidos
      Cuando envio "POST" a "/api/proyectos/1/historias" con {"titulo": "Historia", "story_points": 22}
      Entonces la respuesta tiene el estado HTTP "422"
      Y el cuerpo contiene el error "VALIDACION" para el campo "story_points"
      Y no se crea la historia
