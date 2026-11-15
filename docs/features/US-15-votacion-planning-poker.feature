# language: es
# =============================================================================
# US-15 - Votación oculta en Planning Poker
# Escenarios BDD (Gherkin)
#
# Especificacion SDD : docs/specs/US-15-votacion-planning-poker-spec.md
# Issue de origen    : SebastianAguilera/ingenieria-y-calidad#75
# Endpoint           : POST /api/proyectos/{id}/poker-sessions/{sesionId}/votos
# Trazabilidad       : Issue #75 -> SDD -> criterios -> escenarios BDD ->
#                      tests Go -> codigo
# =============================================================================
#
# NOTA: especificacion ejecutable futura (sin runner de Godog): cada escenario se
# automatiza como subtest en los *_test.go de service, handler, repository y
# cmd/api. Actores: ana y ben son miembros del proyecto; cris no es miembro.
# Escala valida (US-05): 0, 0.5, 1, 2, 3, 5, 8, 13, 21.

Característica: Votación oculta en Planning Poker
  Como miembro de un proyecto
  Quiero votar mi estimación de forma oculta
  Para que nadie condicione su voto antes del revelado

  Antecedentes:
    Dado un proyecto "Metrics" con una sesión de Planning Poker en estado "Votando"
    Y ana y ben son miembros del proyecto
    Y cris no es miembro del proyecto

  # --- Casos normales ---------------------------------------------------------
  Escenario: E-01 Un miembro registra su voto
    Cuando ana vota 5
    Entonces recibe 200 con mi_voto 5 y votos_registrados 1
    Y el voto queda guardado para la ronda actual

  Escenario: E-02 Un miembro cambia su voto mientras la sesión sigue Votando
    Dado que ana ya votó 5
    Cuando ana vota 8
    Entonces recibe 200 con mi_voto 8 y votos_registrados 1
    Y no se crea un segundo voto de ana en la ronda

  Escenario: E-03 Los votos de varios miembros se cuentan sin revelarse
    Dado que ana votó 5
    Cuando ben vota 13
    Entonces recibe 200 con mi_voto 13 y votos_registrados 2
    Y la respuesta no incluye el voto de ana

  Escenario: E-04 El detalle solo muestra el conteo y el voto propio
    Dado que ana votó 5 y ben votó 13
    Cuando ben pide el detalle de la sesión
    Entonces ve votos_registrados 2 y mi_voto 13
    Y la lista de votos está vacía
    Y no aparece ningún valor ni nombre de ana

  # --- Casos alternativos -----------------------------------------------------
  Escenario: E-05 Se aceptan los extremos de la escala
    Cuando ana vota 0
    Entonces recibe 200 con mi_voto 0
    Cuando ana vota 0.5
    Entonces recibe 200 con mi_voto 0.5

  # --- Errores ----------------------------------------------------------------
  Escenario: E-06 No se puede votar si la sesión no está Votando
    Dado que la sesión está en estado "Revelada", "Cerrada" o "Cancelada"
    Cuando ana vota 5
    Entonces recibe 409 con error "SESION_POKER_NO_VOTANDO"
    Y no se inserta ni se modifica ningún voto

  Escenario: E-07 Puntos fuera de la escala
    Cuando ana vota 4
    Entonces recibe 422 con error "PUNTOS_POKER_INVALIDOS"
    Y no se guarda ningún voto

  Escenario: E-08 Puntos negativos
    Cuando ana vota -1
    Entonces recibe 422 con error "PUNTOS_POKER_INVALIDOS"

  Escenario: E-09 Puntos no numéricos o ausentes
    Cuando ana envía el cuerpo {"puntos": "cinco"} o {}
    Entonces recibe 422 con error "PUNTOS_POKER_INVALIDOS"

  Escenario: E-10 Cuerpo que no es JSON
    Cuando ana envía un cuerpo malformado
    Entonces recibe 400 con error "JSON_INVALIDO"

  Escenario: E-11 Un no miembro no puede votar
    Cuando cris vota 5
    Entonces recibe 404 con error "PROYECTO_NO_ENCONTRADO"
    Y no se guarda ningún voto

  Escenario: E-12 La sesión no existe o es de otro proyecto
    Cuando ana vota en una sesión inexistente o de otro proyecto
    Entonces recibe 404 con error "SESION_POKER_NO_ENCONTRADA"

  Escenario: E-13 Sin token no se puede votar
    Cuando se envía el voto sin token
    Entonces recibe 401

  # --- Concurrencia -----------------------------------------------------------
  Escenario: E-14 El revelado gana la carrera contra un voto
    Dado que ana y el sistema actúan al mismo tiempo sobre la sesión "Votando"
    Cuando el revelado cambia el estado a "Revelada" antes de que el voto escriba
    Entonces el voto tardío recibe 409 con error "SESION_POKER_NO_VOTANDO"
    Y el voto no se inserta ni se actualiza

  Escenario: E-15 Dos votos simultáneos del mismo usuario no se duplican
    Cuando ana envía dos votos al mismo tiempo en la misma ronda
    Entonces queda una única fila de voto de ana en la ronda
