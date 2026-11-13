# language: es
# =============================================================================
# US-14 - Sesiones de Planning Poker por historia
# Escenarios BDD (Gherkin)
#
# Especificacion SDD : docs/specs/US-14-sesiones-planning-poker-spec.md
# Issue de origen    : SebastianAguilera/ingenieria-y-calidad#74
# Trazabilidad       : Issue #74 -> SDD -> criterios -> escenarios BDD ->
#                      tests Go -> codigo
# =============================================================================
#
# NOTA: especificacion ejecutable futura (sin runner de Godog): cada escenario se
# automatiza como subtest en los *_test.go de domain, service, handler y repository.
# Actores: ana y ben son MIEMBRO del proyecto; cris no es miembro.
# Estados de sesion: Votando, Revelada, Cerrada, Cancelada.

Característica: Sesiones de Planning Poker por historia
  Como miembro de un proyecto
  Quiero iniciar y consultar sesiones de Planning Poker por historia
  Para estimar en equipo sin perder el contexto del backlog

  Antecedentes:
    Dado un proyecto "Metrics" con una historia "US-01" en estado "Backlog"
    Y ana y ben son miembros del proyecto
    Y cris no es miembro del proyecto

  # --- Casos normales ---------------------------------------------------------
  Escenario: E-01 Se crea una sesión sobre una historia sin sesión activa
    Cuando ana crea una sesión de Planning Poker para "US-01"
    Entonces recibe 201 con estado "Votando", ronda_actual 1 y creado_por ana
    Y no se persiste ningún voto

  Escenario: E-02 El historial lista las sesiones más recientes primero
    Dado que "US-01" tiene una sesión "Cerrada" y luego una "Votando"
    Cuando ana pide GET /api/proyectos/{id}/historias/{historiaId}/poker-sessions
    Entonces recibe 200 con las dos sesiones ordenadas por creación descendente

  Escenario: E-03 El detalle informa el total de miembros y el voto propio
    Dado una sesión "Votando" con ana y ben como miembros
    Cuando ana pide el detalle de la sesión
    Entonces recibe 200 con total_miembros 2 y ya_vote según su voto

  Escenario: E-04 Las sesiones activas listan Votando y Revelada
    Dado una sesión "Votando" y otra "Revelada" en el proyecto
    Cuando ana pide GET /api/proyectos/{id}/poker-sessions/activas
    Entonces recibe 200 con ambas sesiones, sus conteos y el resumen de historia

  Escenario: E-05 El detalle de una sesión Revelada expone los votos
    Dado una sesión "Revelada" con votos de ana (3) y ben (5)
    Cuando ana pide el detalle de la sesión
    Entonces recibe 200 con los votos de ana y ben, con nombre y puntos

  # --- Casos alternativos -----------------------------------------------------
  Escenario: E-06 Se puede crear otra sesión después de una Cerrada
    Dado que la última sesión de "US-01" está "Cerrada"
    Cuando ana crea una sesión de Planning Poker para "US-01"
    Entonces recibe 201 con estado "Votando"

  Escenario: E-07 Se puede crear otra sesión después de una Cancelada
    Dado que la última sesión de "US-01" está "Cancelada"
    Cuando ana crea una sesión de Planning Poker para "US-01"
    Entonces recibe 201 con estado "Votando"

  Escenario: E-08 Sin sesiones activas la lista de activas está vacía
    Dado un proyecto sin sesiones "Votando" ni "Revelada"
    Cuando ana pide GET /api/proyectos/{id}/poker-sessions/activas
    Entonces recibe 200 con la lista vacía []

  Escenario: E-09 Las sesiones Cerrada y Cancelada no aparecen en activas
    Dado una sesión "Cerrada" y otra "Cancelada"
    Cuando ana pide la lista de activas
    Entonces recibe 200 con la lista vacía []

  Escenario: E-10 Una sesión Revelada no expone votos en activas
    Dado una sesión "Revelada" con votos
    Cuando ana pide la lista de activas
    Entonces cada ítem incluye solo conteos y total_miembros
    Y la respuesta no incluye el campo votos

  # --- Casos límite -----------------------------------------------------------
  Escenario: E-11 En estado Votando no se exponen votos ajenos
    Dado una sesión "Votando" donde ben votó
    Cuando ana pide el detalle de la sesión
    Entonces la lista de votos está vacía
    Y solo se informan votos_registrados, ya_vote y mi_voto de ana

  Escenario: E-12 El historial incluye Cerrada y Cancelada
    Dado que "US-01" tiene sesiones "Votando", "Cerrada" y "Cancelada"
    Cuando ana pide el historial de "US-01"
    Entonces recibe las tres sesiones

  # --- Errores ----------------------------------------------------------------
  Escenario: E-13 Ya existe una sesión activa para la historia
    Dado una sesión "Votando" o "Revelada" para "US-01"
    Cuando ana intenta crear otra sesión
    Entonces recibe 409 con error "SESION_POKER_ACTIVA"
    Y no se crea ninguna sesión

  Escenario: E-14 La historia está en estado Hecho
    Dado una historia "US-01" en estado "Hecho"
    Cuando ana intenta crear una sesión
    Entonces recibe 409 con error "HISTORIA_NO_ESTIMABLE"
    Y no se crea ninguna sesión

  Escenario: E-15 Un no miembro no puede crear ni consultar
    Dado que cris no es miembro del proyecto
    Cuando cris intenta crear o consultar una sesión
    Entonces recibe 404 con error "PROYECTO_NO_ENCONTRADO"

  Escenario: E-16 La historia no existe o es de otro proyecto
    Cuando ana crea una sesión para una historia de otro proyecto
    Entonces recibe 404 con error "HISTORIA_NO_ENCONTRADA"

  Escenario: E-17 La sesión no existe o es de otro proyecto
    Cuando ana pide el detalle de una sesión de otro proyecto
    Entonces recibe 404 con error "SESION_POKER_NO_ENCONTRADA"
