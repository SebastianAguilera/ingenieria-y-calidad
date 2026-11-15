# language: es
# =============================================================================
# US-16 - Revelado, dispersión y nuevas rondas
# Escenarios BDD (Gherkin)
#
# Especificacion SDD : docs/specs/US-16-revelado-dispersion-planning-poker-spec.md
# Issue de origen    : SebastianAguilera/ingenieria-y-calidad#78
# Endpoints          : POST /api/proyectos/{id}/poker-sessions/{sesionId}/revelar
#                      POST /api/proyectos/{id}/poker-sessions/{sesionId}/rondas
#                      GET  /api/proyectos/{id}/poker-sessions/{sesionId}
# Trazabilidad       : Issue #78 -> SDD -> criterios -> escenarios BDD ->
#                      tests Go -> codigo
# =============================================================================
#
# NOTA: especificacion ejecutable futura (sin runner de Godog): cada escenario se
# automatiza como subtest en los *_test.go de domain, service, handler,
# repository y cmd/api. Actores: ana, ben y cris son miembros del proyecto; dana
# no es miembro.

Característica: Revelado, dispersión y nuevas rondas de Planning Poker
  Como miembro de un proyecto
  Quiero revelar los votos, ver la dispersión y abrir nuevas rondas
  Para discutir diferencias y volver a estimar

  Antecedentes:
    Dado un proyecto "Metrics" con una sesión de Planning Poker
    Y ana, ben y cris son miembros del proyecto
    Y dana no es miembro del proyecto

  # --- Revelado ---------------------------------------------------------------
  Escenario: E-01 Revelar una sesión con votos
    Dado que la sesión está en "Votando" y ana votó 1, ben votó 3 y cris votó 8
    Cuando ana revela la sesión
    Entonces recibe 200 con estado "Revelada"
    Y ve los votos de ana, ben y cris con nombre y puntos
    Y la ronda actual queda marcada como revelada

  Escenario: E-02 No se puede revelar sin votos
    Dado que la sesión está en "Votando" y no tiene votos
    Cuando ana intenta revelar la sesión
    Entonces recibe 422 con error "SESION_POKER_SIN_VOTOS"
    Y la sesión sigue en "Votando"

  Escenario: E-03 No se puede revelar fuera de Votando
    Dado que la sesión está en estado "Revelada", "Cerrada" o "Cancelada"
    Cuando ana intenta revelar la sesión
    Entonces recibe 409 con error "SESION_POKER_NO_VOTANDO"
    Y no se modifica nada

  # --- Dispersión ------------------------------------------------------------
  Escenario: E-04 Estadísticas de votos distintos 1, 3 y 8
    Dado que ana votó 1, ben votó 3 y cris votó 8
    Cuando se revela la sesión
    Entonces el promedio es 4
    Y la mediana es 3
    Y el mínimo es 1 y el máximo es 8
    Y el rango es 7
    Y la desviación estándar poblacional es aproximadamente 2.9439
    Y hay_dispersion es true
    Y total_votos es 3
    Y votos_por_valor es {"1":1,"3":1,"8":1}

  Escenario: E-05 Votos iguales no tienen dispersión
    Dado que ana, ben y cris votaron 3
    Cuando se revela la sesión
    Entonces hay_dispersion es false
    Y el rango es 0 y la desviación estándar es 0

  Escenario: E-06 Votos distintos 1, 2 y 5 tienen dispersión
    Dado que ana votó 1, ben votó 2 y cris votó 5
    Cuando se revela la sesión
    Entonces hay_dispersion es true

  Escenario: E-07 Un único voto no tiene dispersión
    Dado que solo ana votó 5
    Cuando se revela la sesión
    Entonces hay_dispersion es false
    Y total_votos es 1

  # --- Nuevas rondas ---------------------------------------------------------
  Escenario: E-08 Abrir una nueva ronda desde Revelada
    Dado que la sesión está en "Revelada" con ronda_actual 1
    Cuando ben abre una nueva ronda
    Entonces recibe 200 con ronda_actual 2 y estado "Votando"
    Y la ronda 1 sigue visible en el historial con sus votos y estadísticas

  Escenario: E-09 No se puede abrir una nueva ronda fuera de Revelada
    Dado que la sesión está en "Votando", "Cerrada" o "Cancelada"
    Cuando ana intenta abrir una nueva ronda
    Entonces recibe 409 con error "SESION_POKER_NO_REVELADA"
    Y no se modifica nada

  # --- No filtración e historial --------------------------------------------
  Escenario: E-10 En Votando no se filtran votos ni estadísticas
    Dado que la sesión está en "Votando" con votos registrados
    Cuando ben consulta el detalle
    Entonces ve votos_registrados y su propio voto
    Y la lista de votos está vacía
    Y no aparece la dispersión de la ronda

  Escenario: E-11 El historial conserva rondas anteriores
    Dado que la sesión tuvo una ronda 1 revelada y ahora está en la ronda 2
    Cuando se consulta el detalle
    Entonces historial_rondas incluye la ronda 1 con estado "Revelada"
    Y la ronda 1 conserva sus votos y estadísticas
    Y la ronda 2 aparece en "Votando" sin votos ni estadísticas

  # --- Permisos y errores ----------------------------------------------------
  Escenario: E-12 Cualquier miembro puede revelar o abrir una nueva ronda
    Dado que ben es miembro del proyecto
    Cuando ben revela la sesión o abre una nueva ronda
    Entonces la operación se realiza con éxito

  Escenario: E-13 Un no miembro no puede operar
    Cuando dana intenta revelar o abrir una nueva ronda
    Entonces recibe 404 con error "PROYECTO_NO_ENCONTRADO"

  Escenario: E-14 La sesión no existe o es de otro proyecto
    Cuando ana intenta revelar o abrir una nueva ronda sobre una sesión inexistente
    Entonces recibe 404 con error "SESION_POKER_NO_ENCONTRADA"

  Escenario: E-15 Sin token no se puede operar
    Cuando se intenta revelar o abrir una nueva ronda sin token
    Entonces recibe 401
