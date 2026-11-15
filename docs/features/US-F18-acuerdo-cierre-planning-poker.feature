# language: es
Característica: Acuerdo, cierre y cancelación en Planning Poker
  Como miembro de una sesión de Planning Poker
  Quiero registrar el acuerdo, cerrar la sesión o cancelarla
  Para dejar la historia estimada o liberarla y ver el resultado en el backlog

  Antecedentes:
    Dado que existe el proyecto 1 con la historia 10
    Y existe la sesión de Planning Poker 7
    Y inicié sesión como miembro del proyecto

  Escenario: Registrar el acuerdo desde una sesión revelada
    Dado que la sesión 7 está "Revelada"
    Y los votos son 1, 3 y 8 con mediana 3
    Cuando presiono "Registrar acuerdo"
    Entonces veo el modal "Registrar acuerdo"
    Y la opción "3 puntos" aparece marcada
    Cuando confirmo el acuerdo
    Entonces la aplicación llama a acordar la sesión 7 con 3 puntos
    Y la sesión queda en estado "Cerrada"
    Y veo "Puntos acordados: 3"
    Y no veo "Registrar acuerdo" ni las cartas

  Escenario: Prefill con el valor más votado cuando la mediana no es de la escala
    Dado que la sesión 7 está "Revelada"
    Y los votos son 1, 3, 5 y 8 con mediana 4
    Y el valor más votado es 3
    Cuando presiono "Registrar acuerdo"
    Entonces la opción "3 puntos" aparece marcada

  Escenario: Prefill vacío cuando no hay consenso claro
    Dado que la sesión 7 está "Revelada"
    Y los votos 1, 3 y 8 empatan en cantidad
    Cuando presiono "Registrar acuerdo"
    Entonces ninguna opción de la escala aparece marcada

  Escenario: Confirmar sin elegir un valor
    Dado que la sesión 7 está "Revelada"
    Y el prefill del acuerdo está vacío
    Cuando presiono "Confirmar acuerdo"
    Entonces veo "Elegí una estimación de la escala."
    Y la aplicación no llama a acordar la sesión

  Escenario: No acordar desde una sesión en votación
    Dado que la sesión 7 está "Votando"
    Cuando abro la pantalla de Planning Poker
    Entonces no veo el botón "Registrar acuerdo"

  Escenario: Valor rechazado por el backend
    Dado que la sesión 7 está "Revelada"
    Y el backend rechaza el valor con 422 "PUNTOS_POKER_INVALIDOS"
    Cuando confirmo el acuerdo
    Entonces veo "El valor elegido no es válido. Elegí otro de la escala."
    Y la sesión continúa en estado "Revelada"

  Escenario: Cancelar una sesión revelada
    Dado que la sesión 7 está "Revelada"
    Cuando presiono "Cancelar sesión"
    Entonces veo el diálogo de confirmación
    Cuando confirmo la cancelación
    Entonces la aplicación llama a cancelar la sesión 7
    Y la sesión queda en estado "Cancelada"
    Y veo "La última sesión fue cancelada."
    Y no veo puntos acordados
    Y puedo iniciar otra sesión de Planning Poker

  Escenario: Cancelar una sesión en votación
    Dado que la sesión 7 está "Votando"
    Cuando abro la pantalla de Planning Poker
    Entonces veo el botón "Cancelar sesión"

  Escenario: Descartar el diálogo de cancelación
    Dado que la sesión 7 está "Votando"
    Cuando presiono "Cancelar sesión"
    Y presiono Escape
    Entonces el diálogo se cierra
    Y la aplicación no llama a cancelar la sesión

  Escenario: Cerrar el modal de acuerdo con Escape
    Dado que la sesión 7 está "Revelada"
    Cuando presiono "Registrar acuerdo"
    Y presiono Escape
    Entonces el modal se cierra
    Y la aplicación no llama a acordar la sesión

  Escenario: Sesión ya cerrada o cancelada
    Dado que la última sesión del proyecto está "Cerrada" o "Cancelada"
    Cuando abro la pantalla de Planning Poker
    Entonces no veo "Registrar acuerdo" ni "Cancelar sesión"
    Y la pantalla está en modo lectura

  Escenario: Otro miembro cierra la sesión mientras el modal está abierto
    Dado que la sesión 7 está "Revelada"
    Y tengo el modal de acuerdo abierto
    Cuando otro miembro cierra la sesión
    Y confirmo el acuerdo
    Entonces la API responde 409 con "SESION_POKER_CERRADA"
    Y la aplicación refresca el detalle
    Y veo "La sesión ya fue cerrada o cancelada."

  Escenario: Error de red al acordar
    Dado que la sesión 7 está "Revelada"
    Y la red falla al acordar
    Cuando confirmo el acuerdo
    Entonces veo "No se pudo registrar el acuerdo. Intentá nuevamente."
    Cuando se restablece la red y confirmo nuevamente
    Entonces la sesión queda en estado "Cerrada"

  Escenario: Error de red al cancelar
    Dado que la sesión 7 está "Votando"
    Y la red falla al cancelar
    Cuando confirmo la cancelación
    Entonces veo "No se pudo cancelar la sesión. Intentá nuevamente."
    Y el botón "Cancelar sesión" sigue disponible

  Escenario: Ver la estimación en el backlog
    Dado que la historia 10 quedó con story_points 5 tras cerrar la sesión
    Cuando vuelvo al backlog del proyecto
    Entonces la columna "Estimación" muestra "5"

  Escenario: Historia sin estimar en el backlog
    Dado que la historia 10 no está estimada
    Cuando veo el backlog del proyecto
    Entonces la columna "Estimación" muestra "–"
