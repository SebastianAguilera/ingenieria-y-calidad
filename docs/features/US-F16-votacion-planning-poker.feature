# language: es
# =============================================================================
# US-F16 - Votación con cartas en Planning Poker
# Escenarios BDD (Gherkin)
#
# Especificacion SDD : docs/specs/US-F16-votacion-planning-poker-spec.md
# Issue de origen    : SebastianAguilera/ingenieria-y-calidad#82
# Endpoints          : GET  /api/proyectos/{id}/poker-sessions/{sesionId}
#                      POST /api/proyectos/{id}/poker-sessions/{sesionId}/votos
# Trazabilidad       : Issue #82 -> SDD -> criterios -> escenarios BDD ->
#                      tests Vitest/RTL -> codigo React
# =============================================================================
#
# NOTA: especificacion ejecutable futura (sin runner Cucumber). Cada escenario
# se automatiza en:
#   - frontend/src/api/pokerVotacion.test.ts
#   - frontend/src/pages/proyecto/poker/CartasVoto.test.tsx
#   - frontend/src/pages/proyecto/SesionPokerPage.votacion.test.tsx
# Actores: ana y ben son miembros del proyecto; ana es quien usa la pantalla.
# Escala valida (US-05): 0, 0.5, 1, 2, 3, 5, 8, 13, 21.

Característica: Votación con cartas en Planning Poker
  Como miembro de una sesión de Planning Poker
  Quiero votar con cartas de la escala
  Para registrar mi estimación de forma oculta y ver el avance del equipo

  Antecedentes:
    Dado un proyecto "Metrics" con una historia "US-01"
    Y una sesión de Planning Poker en estado "Votando" con 3 miembros
    Y ana es miembro del proyecto

  # --- Casos normales ---------------------------------------------------------
  Escenario: E-F16-01 Se muestran las cartas y el conteo de votos
    Cuando ana abre la pantalla de Planning Poker
    Entonces ve las cartas 0, 0.5, 1, 2, 3, 5, 8, 13 y 21
    Y ve el texto "0 de 3 votaron"
    Y ve "Todavía no votaste"

  Escenario: E-F16-02 Ana registra su primer voto
    Dado que ana todavía no votó
    Cuando ana elige la carta 5
    Entonces se envía POST con {"puntos": 5}
    Y la carta 5 queda seleccionada
    Y ve "Tu voto: 5"
    Y ve "1 de 3 votaron"

  Escenario: E-F16-03 Ana cambia su voto mientras sigue Votando
    Dado que ana ya votó 5
    Cuando ana elige la carta 8
    Entonces se envía POST con {"puntos": 8}
    Y la carta 8 queda seleccionada
    Y la carta 5 deja de estar seleccionada
    Y ve "Tu voto: 8"
    Y el conteo sigue siendo "1 de 3 votaron"

  Escenario: E-F16-04 Solo se muestra el conteo y el voto propio
    Dado que ana votó 5 y ben votó 13
    Cuando ana consulta la pantalla
    Entonces ve "1 de 3 votaron" o el conteo informado por el backend
    Y ve su propio voto
    Y no ve el nombre de ben
    Y no ve el valor 13 como voto de ben
    Y no se renderiza la lista de votos ajenos

  Escenario: E-F16-05 Se aceptan los extremos de la escala
    Cuando ana elige la carta 0
    Entonces el voto se envía con 0 y la carta 0 queda seleccionada
    Cuando ana elige la carta 0.5
    Entonces el voto se envía con 0.5 y la carta 0.5 queda seleccionada

  Escenario: E-F16-06 El polling actualiza el conteo cada 3 segundos
    Dado que la sesión está en "Votando"
    Y el backend informa "1 de 3 votaron"
    Cuando pasan 3 segundos
    Y el backend informa "2 de 3 votaron"
    Entonces la pantalla muestra "2 de 3 votaron"
    Y no se recarga la página

  # --- Estados de la sesión ---------------------------------------------------
  Escenario: E-F16-07 Una sesión Revelada mantiene las cartas deshabilitadas
    Dado que la sesión está en "Revelada"
    Cuando ana abre la pantalla
    Entonces las cartas están deshabilitadas
    Y se informa que los votos fueron revelados
    Y no se muestran nombres ni valores ajenos

  Escenario: E-F16-08 El polling detecta el paso de Votando a Revelada
    Dado que ana está viendo una sesión "Votando"
    Cuando el polling detecta que la sesión pasó a "Revelada"
    Entonces las cartas se deshabilitan
    Y se informa que los votos fueron revelados
    Y se detiene el polling

  Escenario: E-F16-09 El polling detecta el paso a Cerrada o Cancelada
    Dado que ana está viendo una sesión "Votando"
    Cuando el polling detecta que la sesión pasó a "Cerrada" o "Cancelada"
    Entonces las cartas se deshabilitan
    Y se muestra el nuevo estado
    Y se detiene el polling

  Escenario: E-F16-10 La última sesión ya está Cerrada o Cancelada
    Dado que la última sesión terminó antes de abrir la pantalla
    Cuando ana abre la pantalla
    Entonces se mantiene el resultado de US-F15
    Y no se muestran cartas para votar

  # --- Errores ----------------------------------------------------------------
  Escenario: E-F16-11 El voto llega cuando la votación ya cerró
    Dado que ana está viendo una sesión "Votando"
    Cuando elige una carta y el backend responde 409 "SESION_POKER_NO_VOTANDO"
    Entonces se refresca el detalle de la sesión
    Y se muestra "La votación ya cerró"
    Y las cartas quedan deshabilitadas si la sesión ya no está "Votando"

  Escenario: E-F16-12 Falla la red al votar
    Dado que ana elige una carta
    Cuando el POST falla por red
    Entonces se muestra "No se pudo registrar el voto. Intentá nuevamente."
    Y conserva su voto anterior
    Y puede volver a intentar

  Escenario: E-F16-13 El backend rechaza puntos inválidos
    Dado que el backend responde 422 "PUNTOS_POKER_INVALIDOS"
    Cuando ana intenta votar
    Entonces se muestra un mensaje de valor inválido
    Y no se modifica el voto propio

  Escenario: E-F16-14 Falla la carga inicial del detalle
    Dado que el GET del detalle falla por red
    Cuando ana abre la pantalla
    Entonces se muestra "No se pudo cargar el Planning Poker. Intentá nuevamente."
    Y puede hacer clic en "Reintentar"
    Y al reintentar se vuelve a consultar el detalle

  Escenario: E-F16-15 Falla el polling de la sesión
    Dado que ana está viendo una sesión "Votando"
    Cuando el polling falla por red
    Entonces se conserva el último conteo
    Y se muestra "No se pudo actualizar la sesión. Reintentando…"
    Y el polling se reintenta automáticamente
    Y al recuperarse se limpia el aviso

  Escenario: E-F16-16 La sesión o el proyecto no existen
    Cuando ana abre una sesión inexistente o de otro proyecto
    Entonces se muestra el mensaje de sesión no encontrada
    Cuando ana no tiene acceso al proyecto
    Entonces se muestra "No tenés acceso a este proyecto o no existe"

  # --- Creación y navegación --------------------------------------------------
  Escenario: E-F16-17 Una sesión recién creada permite votar
    Dado que ana inicia una nueva sesión desde la pantalla
    Cuando la sesión se crea correctamente
    Entonces se obtiene el detalle de la nueva sesión
    Y se muestran las cartas habilitadas
    Y se muestra "0 de Y votaron"

  Escenario: E-F16-18 Al desmontar la pantalla se cancela el polling
    Dado que ana está viendo una sesión "Votando"
    Cuando navega fuera de la pantalla
    Entonces no se realizan nuevas consultas de polling
    Y se cancelan las solicitudes pendientes

  # --- Accesibilidad ----------------------------------------------------------
  Escenario: E-F16-19 Las cartas se seleccionan por teclado
    Dado que ana navega con teclado
    Cuando enfoca el grupo de cartas y presiona la tecla de selección
    Entonces la carta enfocada queda seleccionada
    Y el foco permanece visible
    Y la etiqueta accesible de la carta incluye el valor y "puntos"

  Escenario: E-F16-20 El conteo se anuncia al actualizarse
    Dado que ana está viendo el conteo de votos
    Cuando el polling actualiza "votos_registrados"
    Entonces el nuevo texto "X de Y votaron" se anuncia como estado
    Y no se mueve el foco del usuario
