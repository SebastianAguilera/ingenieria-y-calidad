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
# automatiza como subtest en los *_test.go de domain, service, handler,
# repository y cmd/api. Actores: ana, ben y cris son miembros del proyecto; dana
# no es miembro. La historia "HU-5" tiene story_points 3 antes de cada escenario.

Característica: Acuerdo, cierre y cancelación de Planning Poker
  Como miembro de un proyecto
  Quiero registrar la estimación acordada, cerrar la sesión o cancelarla
  Para dejar la historia estimada o liberarla sin cambios

  Antecedentes:
    Dado un proyecto "Metrics" con la historia "HU-5" con 3 story points
    Y una sesión de Planning Poker para "HU-5"
    Y ana, ben y cris son miembros del proyecto
    Y dana no es miembro del proyecto

  # --- Acuerdo y cierre (caso normal) ------------------------------------------
  Escenario: E-01 Acordar un punto válido cierra la sesión y estima la historia
    Dado que la sesión está en "Revelada" en la ronda 1 con votos 3, 5 y 5
    Cuando ana acuerda 5 puntos
    Entonces recibe 200 con estado "Cerrada" y puntos_acordados 5
    Y cerrado_en queda informado
    Y la historia "HU-5" tiene 5 story points
    Y existe un story_estimate con puntos 5, fuente "planning_poker", la sesión, la ronda 1 y estimado_por ana

  Escenario: E-02 Acordar medio punto
    Dado que la sesión está en "Revelada"
    Cuando ben acuerda 0.5 puntos
    Entonces recibe 200 con puntos_acordados 0.5
    Y la historia "HU-5" tiene 0.5 story points

  Escenario: E-03 Acordar un valor distinto de todos los votos
    Dado que la sesión está en "Revelada" con votos 3 y 8
    Cuando ana acuerda 5 puntos
    Entonces recibe 200 con estado "Cerrada" y puntos_acordados 5

  # --- Acuerdo (errores y límites) ---------------------------------------------
  Escenario: E-04 No se puede acordar mientras se vota
    Dado que la sesión está en "Votando"
    Cuando ana intenta acordar 5 puntos
    Entonces recibe 409 con error "SESION_POKER_NO_REVELADA"
    Y la sesión sigue en "Votando"
    Y la historia "HU-5" sigue con 3 story points

  Esquema del escenario: E-05 No se puede acordar una sesión finalizada
    Dado que la sesión está en "<estado>"
    Cuando ana intenta acordar 5 puntos
    Entonces recibe 409 con error "SESION_POKER_CERRADA"
    Y no se modifica nada

    Ejemplos:
      | estado    |
      | Cerrada   |
      | Cancelada |

  Esquema del escenario: E-06 Puntos fuera de la escala de US-05
    Dado que la sesión está en "Revelada"
    Cuando ana intenta acordar <puntos>
    Entonces recibe 422 con error "PUNTOS_POKER_INVALIDOS"
    Y la sesión sigue en "Revelada"
    Y la historia "HU-5" sigue con 3 story points

    Ejemplos:
      | puntos |
      | 4      |
      | -1     |
      | 34     |
      | "cinco"|
      | null   |

  Escenario: E-07 Cuerpo JSON malformado
    Cuando ana envía un cuerpo JSON malformado al acuerdo
    Entonces recibe 400 con error "JSON_INVALIDO"

  # --- Atomicidad --------------------------------------------------------------
  Escenario: E-08 Una falla al crear el story_estimate revierte todo
    Dado que la sesión está en "Revelada"
    Y la creación del story_estimate va a fallar
    Cuando ana acuerda 5 puntos
    Entonces recibe 500 con error "ERROR_INTERNO"
    Y la sesión sigue en "Revelada" sin puntos_acordados
    Y la historia "HU-5" sigue con 3 story points
    Y no existe ningún story_estimate para la sesión

  Escenario: E-09 Una falla al actualizar la historia revierte todo
    Dado que la sesión está en "Revelada"
    Y la actualización de la historia va a fallar
    Cuando ana acuerda 5 puntos
    Entonces la sesión sigue en "Revelada"
    Y no existe ningún story_estimate para la sesión

  Escenario: E-10 Dos acuerdos simultáneos producen un único cierre
    Dado que la sesión está en "Revelada"
    Cuando ana acuerda 5 puntos y ben acuerda 8 puntos al mismo tiempo
    Entonces solo uno de los dos recibe 200
    Y el otro recibe 409 con error "SESION_POKER_CERRADA"
    Y existe un único story_estimate para la sesión

  # --- Cancelación -------------------------------------------------------------
  Esquema del escenario: E-11 Cancelar una sesión activa
    Dado que la sesión está en "<estado>" con votos registrados
    Cuando cris cancela la sesión
    Entonces recibe 200 con estado "Cancelada"
    Y cancelado_en queda informado
    Y puntos_acordados es null y no hay story_estimate
    Y la historia "HU-5" sigue con 3 story points

    Ejemplos:
      | estado   |
      | Votando  |
      | Revelada |

  Esquema del escenario: E-12 No se puede cancelar una sesión finalizada
    Dado que la sesión está en "<estado>"
    Cuando ana intenta cancelar la sesión
    Entonces recibe 409 con error "SESION_POKER_CERRADA"
    Y no se modifica nada

    Ejemplos:
      | estado    |
      | Cerrada   |
      | Cancelada |

  Escenario: E-13 La cancelación libera la historia
    Dado que la sesión fue cancelada
    Cuando ana crea una nueva sesión para "HU-5"
    Entonces recibe 201 con estado "Votando"

  Escenario: E-14 La cancelación conserva el historial
    Dado que la sesión tuvo la ronda 1 revelada con votos y fue cancelada en la ronda 2
    Cuando ana consulta el detalle y el historial de "HU-5"
    Entonces la sesión aparece como "Cancelada"
    Y historial_rondas conserva la ronda 1 con sus votos y estadísticas

  # --- Consultas ---------------------------------------------------------------
  Escenario: E-15 El historial muestra el resultado final de cada sesión
    Dado que "HU-5" tiene una sesión cancelada y otra cerrada con 5 puntos
    Cuando ana consulta el historial de sesiones de "HU-5"
    Entonces la sesión cerrada informa puntos_acordados 5 y su story_estimate
    Y la sesión cancelada informa cancelado_en y no tiene story_estimate

  # --- Permisos y errores ------------------------------------------------------
  Escenario: E-16 Cualquier miembro puede acordar o cancelar
    Dado que ben es miembro y no creó la sesión
    Cuando ben acuerda o cancela la sesión
    Entonces la operación se realiza con éxito

  Escenario: E-17 Un no miembro no puede acordar ni cancelar
    Cuando dana intenta acordar o cancelar la sesión
    Entonces recibe 404 con error "PROYECTO_NO_ENCONTRADO"

  Escenario: E-18 La sesión no existe o es de otro proyecto
    Cuando ana intenta acordar o cancelar una sesión inexistente
    Entonces recibe 404 con error "SESION_POKER_NO_ENCONTRADA"

  Escenario: E-19 Parámetros inválidos
    Cuando ana acuerda o cancela con un sesionId "abc"
    Entonces recibe 400 con error "PARAMETRO_INVALIDO"

  Escenario: E-20 Sin token no se puede operar
    Cuando se intenta acordar o cancelar sin token
    Entonces recibe 401
