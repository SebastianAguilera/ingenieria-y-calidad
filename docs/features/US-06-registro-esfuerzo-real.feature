# language: es
# =============================================================================
# US-06 - Registro de esfuerzo real en la historia
# Issue: #33
# Especificación SDD: docs/specs/US-06-registro-esfuerzo-real-spec.md
# Trazabilidad: Issue #33 -> SDD -> Criterios de Aceptación -> Escenarios BDD -> Tests Go -> Código
# =============================================================================

Funcionalidad: US-06 - Registro de esfuerzo real en la historia
  Como integrante del equipo
  Quiero registrar el esfuerzo real (horas trabajadas) en una historia
  Para comparar el esfuerzo estimado con el esfuerzo real

  Contexto:
    Dado que la API esta levantada y responde "200 OK" en "GET /health"
    Y que la base de datos fue migrada correctamente
    Y que existe un proyecto con una historia de usuario

  Regla: El esfuerzo real es opcional y se registra en el alta

    @caso_normal
    Escenario: BDD-01 - Registrar esfuerzo real valido en el alta
      Cuando envio "POST" a "/api/proyectos/1/historias" con {"titulo": "Historia con esfuerzo", "esfuerzo_real": 4.5}
      Entonces la respuesta tiene el estado HTTP "201"
      Y el detalle de la historia registra "esfuerzo_real" igual a 4.5

    @caso_normal
    Escenario: BDD-02 - Crear una historia sin esfuerzo real
      Cuando envio "POST" a "/api/proyectos/1/historias" con {"titulo": "Historia sin esfuerzo"}
      Entonces la respuesta tiene el estado HTTP "201"
      Y el detalle de la historia registra "esfuerzo_real" igual a null

  Regla: El esfuerzo real se puede modificar y consultar

    @caso_normal
    Escenario: BDD-03 - Registrar esfuerzo real valido en la modificacion
      Cuando envio "PATCH" a "/api/proyectos/1/historias/1" con {"esfuerzo_real": 6}
      Entonces la respuesta tiene el estado HTTP "200"
      Y el detalle de la historia registra "esfuerzo_real" igual a 6

    @caso_normal
    Escenario: BDD-07 - El esfuerzo real aparece en el listado
      Dado que la historia tiene 6 horas de esfuerzo real registradas
      Cuando envio "GET" a "/api/proyectos/1/historias"
      Entonces la respuesta tiene el estado HTTP "200"
      Y la historia del listado registra "esfuerzo_real" igual a 6

  Regla: Se rechazan valores invalidos de esfuerzo real

    @caso_limite
    Escenario: BDD-04 - Rechazar esfuerzo real igual a cero
      Cuando envio "PATCH" a "/api/proyectos/1/historias/1" con {"esfuerzo_real": 0}
      Entonces la respuesta tiene el estado HTTP "422"
      Y el cuerpo contiene el error "VALIDACION" para el campo "esfuerzo_real"

    @caso_error
    Escenario: BDD-05 - Rechazar esfuerzo real negativo
      Cuando envio "PATCH" a "/api/proyectos/1/historias/1" con {"esfuerzo_real": -2}
      Entonces la respuesta tiene el estado HTTP "422"
      Y el cuerpo contiene el error "VALIDACION" para el campo "esfuerzo_real"

    @caso_error
    Escenario: BDD-06 - Rechazar esfuerzo real no numerico
      Cuando envio "PATCH" a "/api/proyectos/1/historias/1" con {"esfuerzo_real": "6"}
      Entonces la respuesta tiene el estado HTTP "422"
      Y el cuerpo contiene el error "VALIDACION" para el campo "esfuerzo_real"
