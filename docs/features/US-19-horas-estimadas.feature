# language: es
# =============================================================================
# US-19 - Horas estimadas por historia
# Issue: #98
# Especificación SDD: docs/specs/US-19-horas-estimadas-spec.md
# Dependencias: US-02 (historias), US-06 (esfuerzo real), US-13 (Sprint Backlog),
#               US-F10 (membresía). Futura: US-20 (desviación de esfuerzo).
# Trazabilidad: Issue #98 -> SDD -> Criterios de Aceptación -> Escenarios BDD -> Tests Go -> Código
# =============================================================================

Funcionalidad: US-19 - Horas estimadas por historia
  Como integrante del equipo de un proyecto
  Quiero registrar las horas estimadas de cada historia
  Para compararlas con las horas reales cuando se mida la desviación de esfuerzo

  Contexto:
    Dado que la API esta levantada y responde "200 OK" en "GET /health"
    Y que la base de datos fue migrada correctamente
    Y que existe un proyecto con una historia de usuario

  Regla: Las horas estimadas son opcionales y se registran en el alta

    @caso_normal
    Escenario: BDD-01 - Registrar horas estimadas en el alta
      Cuando envio "POST" a "/api/proyectos/1/historias" con {"titulo": "Historia estimada", "horas_estimadas": 5}
      Entonces la respuesta tiene el estado HTTP "201"
      Y el detalle de la historia registra "horas_estimadas" igual a 5

    @caso_normal
    Escenario: BDD-02 - Crear una historia sin horas estimadas
      Cuando envio "POST" a "/api/proyectos/1/historias" con {"titulo": "Historia sin horas"}
      Entonces la respuesta tiene el estado HTTP "201"
      Y el detalle de la historia registra "horas_estimadas" igual a null

    @caso_normal
    Escenario: BDD-03 - Crear una historia con horas estimadas null
      Cuando envio "POST" a "/api/proyectos/1/historias" con {"titulo": "Historia con null", "horas_estimadas": null}
      Entonces la respuesta tiene el estado HTTP "201"
      Y el detalle de la historia registra "horas_estimadas" igual a null

  Regla: Las horas estimadas se pueden modificar y consultar

    @caso_normal
    Escenario: BDD-04 - Modificar las horas estimadas
      Cuando envio "PATCH" a "/api/proyectos/1/historias/1" con {"horas_estimadas": 8}
      Entonces la respuesta tiene el estado HTTP "200"
      Y el detalle de la historia registra "horas_estimadas" igual a 8

    @caso_normal
    Escenario: BDD-05 - Las horas estimadas aparecen en el listado
      Dado que la historia tiene 8 horas estimadas registradas
      Cuando envio "GET" a "/api/proyectos/1/historias"
      Entonces la respuesta tiene el estado HTTP "200"
      Y la historia del listado registra "horas_estimadas" igual a 8

    @caso_normal
    Escenario: BDD-06 - Las horas estimadas aparecen en el Sprint Backlog
      Dado que la historia tiene 8 horas estimadas registradas
      Y que la historia esta asignada al sprint 1
      Cuando envio "GET" a "/api/proyectos/1/sprints/1/historias"
      Entonces la respuesta tiene el estado HTTP "200"
      Y la historia del Sprint Backlog registra "horas_estimadas" igual a 8

    @caso_normal
    Escenario: BDD-07 - Cambiar las horas estimadas no altera story points ni esfuerzo real
      Dado que la historia tiene "story_points" 5 y "esfuerzo_real" 3
      Cuando envio "PATCH" a "/api/proyectos/1/historias/1" con {"horas_estimadas": 8}
      Entonces la respuesta tiene el estado HTTP "200"
      Y la historia conserva "story_points" igual a 5 y "esfuerzo_real" igual a 3

    @caso_normal
    Escenario: BDD-08 - Un integrante no lider puede registrar horas estimadas
      Dado que soy integrante del proyecto con rol "MIEMBRO"
      Cuando envio "PATCH" a "/api/proyectos/1/historias/1" con {"horas_estimadas": 4}
      Entonces la respuesta tiene el estado HTTP "200"

  Regla: Se rechazan valores invalidos de horas estimadas

    @caso_limite
    Escenario: BDD-09 - Rechazar horas estimadas igual a cero
      Cuando envio "PATCH" a "/api/proyectos/1/historias/1" con {"horas_estimadas": 0}
      Entonces la respuesta tiene el estado HTTP "422"
      Y el cuerpo contiene el error "VALIDACION" para el campo "horas_estimadas"
      Y la historia no se modifica

    @caso_error
    Escenario: BDD-10 - Rechazar horas estimadas negativas en el alta
      Cuando envio "POST" a "/api/proyectos/1/historias" con {"titulo": "Historia invalida", "horas_estimadas": -2}
      Entonces la respuesta tiene el estado HTTP "422"
      Y el cuerpo contiene el error "VALIDACION" para el campo "horas_estimadas"

    @caso_error
    Escenario: BDD-11 - Rechazar horas estimadas no numericas
      Cuando envio "PATCH" a "/api/proyectos/1/historias/1" con {"horas_estimadas": "cinco"}
      Entonces la respuesta tiene el estado HTTP "422"
      Y el cuerpo contiene el error "VALIDACION" para el campo "horas_estimadas"

  Regla: La migracion conserva los datos existentes

    @caso_normal
    Escenario: BDD-12 - Las historias existentes quedan con horas estimadas null
      Dado una base con una historia creada antes de agregar la columna
      Cuando se ejecuta la migracion
      Entonces la historia existente conserva sus datos
      Y su "horas_estimadas" es null
