# language: es
# =============================================================================
# US-17 - Acuerdo, cierre y cancelación de Planning Poker
# Escenarios BDD (Gherkin)
#
# Especificacion SDD : docs/specs/US-17-acuerdo-cierre-cancelacion-planning-poker-spec.md
# Issue de origen    : SebastianAguilera/ingenieria-y-calidad#79
# Endpoints          : POST /api/proyectos/{id}/poker-sessions/{sesionId}/acuerdo
#                      POST /api/proyectos/{id}/poker-sessions/{sesionId}/cancelar
#                      GET  /api/proyectos/{id}/poker-sessions/{sesionId}
#                      GET  /api/proyectos/{id}/historias/{historiaId}/poker-sessions
# Trazabilidad       : Issue #79 -> SDD -> criterios -> escenarios BDD ->
#                      tests Go -> codigo
# =============================================================================
#
# NOTA: especificacion ejecutable futura (sin runner de Godog): cada escenario se
# automatiza como subtest en los *_test.go de domain, service, handler y
# repository. Actores: ana, ben y cris son miembros del proyecto; dana no es
# miembro.

Característica: Acuerdo, cierre y cancelación de Planning Poker
  Como miembro de un proyecto
  Quiero acordar los puntos de una sesión revelada o cancelarla
  Para dejar la historia estimada o liberarla sin estimación

  Antecedentes:
    Dado un proyecto "Metrics" con una historia "US-05" estimable
    Y ana, ben y cris son miembros del proyecto
    Y dana no es miembro del proyecto

  # --- Acuerdo ---------------------------------------------------------------
  Escenario: E-01 Acordar puntos de una sesión revelada
    Dado que la sesión está en "Revelada" con votos de ana, ben y cris
    Cuando ana acuerda 5 puntos
    Entonces recibe 200 con estado "Cerrada"
    Y puntos_acordados es 5 y cerrado_en está informado
    Y la historia queda con story_points 5

  Escenario: E-02 El acuerdo crea la estimación completa
    Dado que la sesión está en "Revelada" en la ronda 1
    Cuando ana acuerda 8 puntos
    Entonces existe story_estimate con historia_id, sesion_id y ronda_id
    Y la estimación tiene puntos 8, fuente "planning_poker" y estimado_por ana
    Y estimado_en está informado

  Escenario: E-03 Acordar 0 y 0.5 es válido
    Dado que la sesión está en "Revelada"
    Cuando se acuerda 0 puntos
    Entonces la sesión queda "Cerrada" con puntos_acordados 0
    Y cuando se acuerda 0.5 puntos en otra sesión revelada
    Entonces la sesión queda "Cerrada" con puntos_acordados 0.5

  Escenario: E-04 No se puede acordar desde Votando
    Dado que la sesión está en "Votando"
    Cuando ana intenta acordar 5 puntos
    Entonces recibe 409 con error "SESION_POKER_NO_REVELADA"
    Y la historia no cambia y la sesión sigue en "Votando"

  Escenario: E-05 No se puede acordar una sesión ya cerrada
    Dado que la sesión está en "Cerrada"
    Cuando ana intenta acordar 8 puntos
    Entonces recibe 409 con error "SESION_POKER_CERRADA"
    Y no se modifica nada

  Escenario: E-06 No se puede acordar una sesión cancelada
    Dado que la sesión está en "Cancelada"
    Cuando ana intenta acordar 8 puntos
    Entonces recibe 409 con error "SESION_POKER_CERRADA"
    Y no se modifica nada

  Escenario: E-07 Puntos ausentes, no numéricos o fuera de escala
    Dado que la sesión está en "Revelada"
    Cuando se envía un cuerpo sin puntos, con puntos de texto o con puntos 4
    Entonces recibe 422 con error "PUNTOS_POKER_INVALIDOS" en el campo "puntos"
    Y la sesión sigue en "Revelada" y la historia no cambia

  Escenario: E-08 JSON malformado en el acuerdo
    Dado que la sesión está en "Revelada"
    Cuando se envía un cuerpo JSON inválido
    Entonces recibe 400 con error "JSON_INVALIDO"
    Y no se modifica nada

  Escenario: E-09 Atomicidad del acuerdo
    Dado que la sesión está en "Revelada"
    Y la inserción en story_estimates va a fallar
    Cuando ana acuerda 5 puntos
    Entonces la operación responde con error
    Y la sesión sigue en "Revelada" sin puntos_acordados ni cerrado_en
    Y la historia conserva su story_points anterior
    Y no queda ninguna estimación parcial

  # --- Cancelación -----------------------------------------------------------
  Escenario: E-10 Cancelar una sesión en Votando
    Dado que la sesión está en "Votando" con votos registrados
    Cuando ben cancela la sesión
    Entonces recibe 200 con estado "Cancelada" y cancelado_en informado
    Y la historia conserva su story_points anterior
    Y los votos y rondas siguen guardados

  Escenario: E-11 Cancelar una sesión revelada
    Dado que la sesión está en "Revelada" en la ronda 1
    Cuando ben cancela la sesión
    Entonces recibe 200 con estado "Cancelada"
    Y no existe story_estimate para la sesión
    Y puntos_acordados y cerrado_en quedan vacíos

  Escenario: E-12 Cancelar libera la historia
    Dado que la sesión está "Cancelada"
    Cuando ana crea una nueva sesión para la misma historia
    Entonces la creación es exitosa porque el índice parcial solo cubre Votando y Revelada

  Escenario: E-13 No se puede cancelar una sesión cerrada
    Dado que la sesión está en "Cerrada"
    Cuando ana intenta cancelarla
    Entonces recibe 409 con error "SESION_POKER_CERRADA"
    Y no se modifica nada

  Escenario: E-14 No se puede cancelar una sesión ya cancelada
    Dado que la sesión está en "Cancelada"
    Cuando ana intenta cancelarla otra vez
    Entonces recibe 409 con error "SESION_POKER_CERRADA"
    Y no se modifica nada

  # --- Historial y detalle ---------------------------------------------------
  Escenario: E-15 El detalle de una sesión cerrada incluye la estimación
    Dado que la sesión está "Cerrada" con una estimación acordada
    Cuando se consulta el detalle
    Entonces story_estimate aparece como objeto completo
    Y la sesión informa puntos_acordados y cerrado_en

  Escenario: E-16 El historial incluye estimaciones y cancelaciones
    Dado que la historia tiene una sesión "Cerrada" con estimación
    Y una sesión "Cancelada" con cancelado_en
    Cuando se consulta el historial
    Entonces la sesión cerrada incluye story_estimate
    Y la sesión cancelada incluye cancelado_en y no incluye story_estimate

  # --- Concurrencia ----------------------------------------------------------
  Escenario: E-17 Acuerdo y cancelación concurrentes
    Dado que la sesión está en "Revelada"
    Cuando un miembro acuerda y otro cancela a la vez
    Entonces solo una operación gana
    Y la operación perdedora recibe 409 con error "SESION_POKER_CERRADA"
    Y no quedan escrituras parciales

  # --- Permisos y errores ----------------------------------------------------
  Escenario: E-18 Cualquier miembro puede acordar o cancelar
    Dado que ben es miembro del proyecto
    Cuando ben acuerda o cancela una sesión válida
    Entonces la operación se realiza con éxito

  Escenario: E-19 Un no miembro no puede operar
    Cuando dana intenta acordar o cancelar
    Entonces recibe 404 con error "PROYECTO_NO_ENCONTRADO"

  Escenario: E-20 La sesión no existe o es de otro proyecto
    Cuando ana intenta acordar o cancelar sobre una sesión inexistente
    Entonces recibe 404 con error "SESION_POKER_NO_ENCONTRADA"

  Escenario: E-21 Parámetros de ruta inválidos
    Cuando se envía un proyecto o sesión con id no numérico
    Entonces recibe 400 con error "PARAMETRO_INVALIDO"

  Escenario: E-22 Sin token no se puede operar
    Cuando se intenta acordar o cancelar sin token
    Entonces recibe 401
