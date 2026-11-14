# language: es
# =============================================================================
# US-F15 - Entrada y sesión de Planning Poker en el backlog
# Escenarios BDD (Gherkin)
#
# Especificacion SDD : docs/specs/US-F15-entrada-planning-poker-spec.md
# Issue de origen    : SebastianAguilera/ingenieria-y-calidad#81
# Trazabilidad       : Issue #81 -> SDD -> criterios -> escenarios BDD ->
#                      tests Vitest/RTL -> codigo React
# =============================================================================
#
# NOTA: especificacion ejecutable futura (sin runner de Godog/Cucumber): cada
# escenario se automatiza como test de Vitest/RTL en
# frontend/src/api/poker.test.ts, frontend/src/pages/proyecto/SesionPokerPage.test.tsx
# y frontend/src/pages/proyecto/BacklogPage.test.tsx.
# Actores: ana es MIEMBRO del proyecto.
# Estados de sesion: Votando, Revelada, Cerrada, Cancelada.

Característica: Entrada y sesión de Planning Poker en el backlog
  Como miembro de un proyecto
  Quiero entrar al Planning Poker desde una historia del backlog
  Para iniciar una sesión o unirme a la que ya está activa

  Antecedentes:
    Dado un proyecto "Metrics" con una historia "US-01" en estado "Backlog"
    Y ana es miembro del proyecto

  # --- Casos normales ---------------------------------------------------------
  Escenario: E-F15-01 Se abre la entrada sin sesión activa
    Dado que "US-01" no tiene sesiones de Planning Poker
    Cuando ana abre /proyectos/1/backlog/10/poker
    Entonces ve el resumen de "US-01"
    Y ve el botón "Iniciar Planning Poker"
    Y no se envía ningún POST de creación

  Escenario: E-F15-02 Se inicia la sesión con clic explícito
    Dado que "US-01" no tiene sesiones activas
    Cuando ana hace clic en "Iniciar Planning Poker"
    Entonces se envía POST /api/proyectos/1/historias/10/poker-sessions
    Y la pantalla muestra una sesión "Votando" en ronda 1

  Escenario: E-F15-03 Se entra directamente a una sesión Votando
    Dado que "US-01" tiene una sesión "Votando"
    Cuando ana abre la entrada de Planning Poker
    Entonces ve el panel "Sesión en curso" con estado "Votando"
    Y no se ofrece iniciar otra sesión
    Y no se envía ningún POST

  Escenario: E-F15-04 Se entra directamente a una sesión Revelada
    Dado que "US-01" tiene una sesión "Revelada"
    Cuando ana abre la entrada de Planning Poker
    Entonces ve el panel "Sesión en curso" con estado "Revelada"
    Y se informa que los votos fueron revelados
    Y no se envía ningún POST

  Escenario: E-F15-05 Se muestra el resultado de la última sesión Cerrada
    Dado que la última sesión de "US-01" está "Cerrada" con 5 puntos acordados
    Cuando ana abre la entrada de Planning Poker
    Entonces ve "Última sesión cerrada"
    Y ve los puntos acordados 5
    Y ve el botón "Iniciar Planning Poker"

  Escenario: E-F15-06 Se muestra la cancelación y se permite iniciar otra
    Dado que la última sesión de "US-01" está "Cancelada"
    Cuando ana abre la entrada de Planning Poker
    Entonces ve que la última sesión fue cancelada
    Y ve el botón "Iniciar Planning Poker"

  Escenario: E-F15-07 El backlog muestra el distintivo de sesión activa
    Dado que "US-01" tiene una sesión "Votando"
    Cuando ana abre el backlog
    Entonces la fila de "US-01" muestra "Planning Poker en curso"

  Escenario: E-F15-08 Desde el backlog se navega a la entrada de Planning Poker
    Dado el backlog del proyecto
    Cuando ana hace clic en "Planning Poker" de "US-01"
    Entonces navega a /proyectos/1/backlog/10/poker

  # --- Casos alternativos -----------------------------------------------------
  Escenario: E-F15-09 Se resuelve una carrera de sesión activa
    Dado que "US-01" no tenía sesión activa al cargar
    Y otro miembro crea una sesión antes del clic de ana
    Cuando ana hace clic en "Iniciar Planning Poker"
    Entonces el backend responde 409 "SESION_POKER_ACTIVA"
    Y la pantalla relee el historial
    Y ana se une a la sesión activa sin crear otra

  Escenario: E-F15-10 La última sesión cerrada no tiene puntos acordados
    Dado que la última sesión de "US-01" está "Cerrada" con puntos_acordados nulo
    Cuando ana abre la entrada de Planning Poker
    Entonces ve "Sin puntos acordados"

  # --- Casos límite -----------------------------------------------------------
  Escenario: E-F15-11 La historia tiene un historial vacío
    Dado que "US-01" no tiene sesiones
    Cuando ana abre la entrada de Planning Poker
    Entonces ve el aviso de que todavía no hay sesiones
    Y ve el botón "Iniciar Planning Poker"

  Escenario: E-F15-12 Una sesión activa cambia mientras la pantalla está abierta
    Dado que "US-01" tiene una sesión "Votando"
    Cuando ana vuelve a cargar la pantalla y la sesión ya está "Revelada"
    Entonces ve el estado actualizado "Revelada"

  # --- Errores ----------------------------------------------------------------
  Escenario: E-F15-13 La historia está en estado Hecho
    Dado que "US-01" está en estado "Hecho"
    Cuando ana hace clic en "Iniciar Planning Poker"
    Entonces el backend responde 409 "HISTORIA_NO_ESTIMABLE"
    Y la pantalla muestra el mensaje de historia no estimable
    Y no se muestra una sesión activa

  Escenario: E-F15-14 La historia no existe o es de otro proyecto
    Cuando ana abre la entrada de una historia inexistente
    Entonces la pantalla muestra el mensaje de historia no encontrada
    Y ofrece volver al backlog

  Escenario: E-F15-15 El proyecto no existe o ana no es miembro
    Cuando ana abre la entrada de un proyecto ajeno
    Entonces la pantalla muestra "No tenés acceso a este proyecto o no existe"

  Escenario: E-F15-16 Falla la red al cargar la entrada
    Dado que la consulta del historial falla por red
    Cuando ana abre la entrada de Planning Poker
    Entonces ve un mensaje de error de carga
    Y puede hacer clic en "Reintentar"

  Escenario: E-F15-17 Falla la red al iniciar la sesión
    Dado que "US-01" no tiene sesión activa
    Y el POST de creación falla por red
    Cuando ana hace clic en "Iniciar Planning Poker"
    Entonces ve un mensaje de error al iniciar
    Y no se muestra una sesión activa

  Escenario: E-F15-18 La consulta de activas falla en el backlog
    Dado que GET /api/proyectos/1/poker-sessions/activas falla
    Cuando ana abre el backlog
    Entonces el backlog se muestra igual
    Y no se muestran distintivos de Planning Poker
