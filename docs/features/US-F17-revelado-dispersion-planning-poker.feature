# language: es
Característica: Revelado y dispersión en Planning Poker
  Como miembro de una sesión de Planning Poker
  Quiero revelar los votos y ver la dispersión
  Para discutir las diferencias del equipo

  Antecedentes:
    Dado que existe el proyecto 1 con la historia 9
    Y existe la sesión de Planning Poker 7 en estado "Votando"
    Y inicié sesión como miembro del proyecto

  Escenario: Revelar una ronda con votos
    Dado que la sesión 7 tiene los votos Ana 1, Ben 3 y Cris 8
    Cuando presiono "Revelar votos"
    Entonces la aplicación llama a revelar la sesión 7
    Y los votos de Ana, Ben y Cris se muestran con sus puntos
    Y la sesión queda en estado "Revelada"
    Y se muestra el botón "Nueva ronda"

  Escenario: Revelar sin votos
    Dado que la sesión 7 no tiene votos
    Cuando presiono "Revelar votos"
    Entonces la API responde 422 con "SESION_POKER_SIN_VOTOS"
    Y veo el mensaje "No se puede revelar sin votos."
    Y la sesión continúa en estado "Votando"

  Escenario: Ver el panel de dispersión con diferencias
    Dado que la sesión 7 está "Revelada"
    Y los votos son 1, 3 y 8
    Y el backend devuelve promedio 4, mediana 3, mínimo 1, máximo 8, rango 7, desviación estándar 2.9439202887759492, total 3 y hay dispersión
    Cuando se muestra el detalle
    Entonces veo "Promedio: 4", "Mediana: 3", "Mínimo: 1" y "Máximo: 8"
    Y veo "Rango: 7" y "Desviación estándar: 2,94"
    Y veo "Total de votos: 3"
    Y veo la distribución "1 punto: 1 voto", "3 puntos: 1 voto" y "8 puntos: 1 voto"
    Y el badge dice "Hay diferencias"
    Y los votos de 1 y 8 llevan las marcas "Mínimo" y "Máximo"

  Escenario: Ver un consenso
    Dado que la sesión 7 está "Revelada"
    Y los votos son 3, 3 y 3
    Y el backend devuelve que no hay dispersión
    Cuando se muestra el detalle
    Entonces el badge dice "Consenso"
    Y los votos 3, 3 y 3 llevan la marca "Mínimo y máximo"

  Escenario: Abrir una nueva ronda después de revelar
    Dado que la sesión 7 está "Revelada" en la ronda 1
    Y mi voto guardado es 5
    Cuando presiono "Nueva ronda"
    Entonces la aplicación llama a abrir la ronda de la sesión 7
    Y la sesión vuelve a estado "Votando" en la ronda 2
    Y mi voto queda vacío
    Y ya no veo el botón "Nueva ronda"
    Y vuelvo a ver el botón "Revelar votos"

  Escenario: Consultar el historial de la ronda anterior
    Dado que la sesión 7 está "Votando" en la ronda 2
    Y la ronda 1 está revelada con los votos Ana 1, Ben 3 y Cris 8
    Cuando se muestra el detalle
    Entonces veo la región "Historial de rondas"
    Y veo la "Ronda 1" con "Revelada el 13/11/2026"
    Y veo los votos y las estadísticas de la ronda 1
    Y no veo la ronda 2 dentro del historial

  Escenario: No duplicar la ronda actual en el historial
    Dado que la sesión 7 está "Revelada" en la ronda 1
    Y no existen rondas anteriores
    Cuando se muestra el detalle
    Entonces no se muestra la región "Historial de rondas"

  Escenario: Sesión cerrada sin acciones de votación
    Dado que la última sesión del proyecto está "Cerrada"
    Y el proyecto ya tiene un resultado de Planning Poker
    Cuando abro la pantalla de Planning Poker
    Entonces veo el resultado de la sesión
    Y no veo "Revelar votos", "Nueva ronda" ni las cartas

  Escenario: Sesión cancelada sin acciones de votación
    Dado que la última sesión del proyecto está "Cancelada"
    Cuando abro la pantalla de Planning Poker
    Entonces no veo "Revelar votos", "Nueva ronda" ni las cartas

  Escenario: Error de red al revelar
    Dado que la sesión 7 está "Votando"
    Y la red falla al revelar
    Cuando presiono "Revelar votos"
    Entonces veo "No se pudieron revelar los votos. Intentá nuevamente."
    Y el botón "Revelar votos" sigue disponible
    Cuando se restablece la red y presiono nuevamente
    Entonces los votos se revelan correctamente

  Escenario: Error de red al abrir una nueva ronda
    Dado que la sesión 7 está "Revelada"
    Y la red falla al abrir la ronda
    Cuando presiono "Nueva ronda"
    Entonces veo "No se pudo abrir la nueva ronda. Intentá nuevamente."
    Y el botón "Nueva ronda" sigue disponible
    Cuando se restablece la red y presiono nuevamente
    Entonces la sesión vuelve a "Votando"

  Escenario: Otro miembro revela durante el polling
    Dado que la sesión 7 está "Votando"
    Y el polling consulta el detalle cada 3 segundos
    Cuando otro miembro revela los votos
    Entonces la siguiente consulta devuelve la sesión "Revelada"
    Y veo los votos y el panel de dispersión sin recargar la página
    Y ya no veo el botón "Revelar votos"

  Escenario: Otro miembro abre una nueva ronda durante el polling
    Dado que la sesión 7 está "Revelada" en la ronda 1
    Y tenía seleccionados 5 puntos
    Cuando otro miembro abre la ronda 2
    Entonces la siguiente consulta devuelve la sesión "Votando" en la ronda 2
    Y mi selección local queda vacía
    Y veo las cartas habilitadas y el botón "Revelar votos"

  Escenario: La sesión se cierra durante el polling
    Dado que la sesión 7 está "Revelada"
    Cuando otro miembro cierra la sesión
    Entonces la siguiente consulta devuelve la sesión "Cerrada"
    Y se detiene el polling
    Y no veo "Revelar votos" ni "Nueva ronda"

  Escenario: Falla transitoria durante el polling
    Dado que la sesión 7 está "Votando"
    Y una consulta de polling falla por red
    Cuando se muestra el error
    Entonces veo "No se pudo actualizar la sesión. Reintentando…"
    Y conservo los últimos votos y el conteo visibles
    Y el siguiente polling exitoso limpia el aviso

  Escenario: Intento de revelar una sesión que ya no vota
    Dado que la sesión 7 cambió a "Revelada"
    Y la pantalla todavía muestra "Votando"
    Cuando presiono "Revelar votos"
    Entonces la API responde 409 con "SESION_POKER_NO_VOTANDO"
    Y la aplicación refresca el detalle
    Y veo "La votación ya cerró"

  Escenario: Intento de abrir una ronda que ya no está revelada
    Dado que la sesión 7 volvió a "Votando"
    Y la pantalla todavía muestra "Revelada"
    Cuando presiono "Nueva ronda"
    Entonces la API responde 409 con "SESION_POKER_NO_REVELADA"
    Y la aplicación refresca el detalle
    Y veo "La sesión ya no está revelada."
