# language: es
Característica: Notificación de sesiones activas de Planning Poker
  Como miembro de un proyecto
  Quiero ver una campana cuando se abre un Planning Poker
  Para enterarme y unirme sin depender de que alguien me avise

  Antecedentes:
    Dado que existe el proyecto 1 con la historia 10 "US-01 Registrar horas reales"
    Y inicié sesión como miembro del proyecto
    Y estoy en el layout del proyecto 1

  Escenario: Polling único cada tres segundos
    Dado que no hay sesiones activas
    Cuando el layout se monta
    Entonces la aplicación consulta las sesiones activas del proyecto 1
    Y vuelve a consultarlas cada 3 segundos
    Y no hay más de un polling por layout

  Escenario: Pausar el polling con la pestaña oculta
    Dado que el layout está montado
    Y la pestaña está oculta
    Cuando pasan 3 segundos
    Entonces la aplicación no consulta las sesiones activas
    Cuando la pestaña vuelve a estar visible
    Entonces la aplicación consulta las sesiones activas de inmediato
    Y reanuda el polling

  Escenario: Primera carga con sesiones existentes
    Dado que el proyecto 1 tiene la sesión 7 "Votando" de la historia 10
    Cuando el layout se monta
    Entonces la campana muestra la sesión 7
    Y el badge muestra 1
    Y no veo ningún toast

  Escenario: Toast y badge al aparecer una sesión nueva
    Dado que la campana está vacía
    Cuando el polling detecta la sesión 7 "Votando" de la historia 10
    Entonces veo el toast "Se abrió Planning Poker para US-01: Registrar horas reales"
    Y el badge muestra 1

  Escenario: El badge no cuenta las sesiones donde ya voté
    Dado que el proyecto 1 tiene la sesión 7 "Votando" con ya_vote verdadero
    Cuando el layout se monta
    Entonces la campana lista la sesión 7
    Y el badge no se muestra

  Escenario: No repetir el toast para una sesión ya vista
    Dado que la sesión 7 "Votando" ya generó un toast
    Cuando el polling vuelve a devolver la sesión 7
    Entonces no veo otro toast para la sesión 7

  Escenario: No notificar al creador de la sesión
    Dado que la sesión 7 fue creada por mi usuario
    Cuando el polling detecta la sesión 7 "Votando"
    Entonces la campana lista la sesión 7
    Y no veo el toast

  Escenario: Sesión revelada sin badge ni toast
    Dado que la campana está vacía
    Cuando el polling detecta la sesión 8 "Revelada" de la historia 11
    Entonces la campana lista la sesión 8 con estado "Revelada"
    Y el badge no se muestra
    Y no veo el toast

  Escenario: Varias sesiones nuevas se muestran una por una
    Dado que la campana está vacía
    Cuando el polling detecta las sesiones 7 y 8 "Votando"
    Entonces veo el toast de la sesión 7
    Y cuando el toast se cierra veo el toast de la sesión 8

  Escenario: Abrir la lista de la campana con teclado
    Dado que el proyecto 1 tiene la sesión 7 "Votando"
    Cuando enfoco la campana y la abro con el teclado
    Entonces la lista muestra "US-01", "Registrar horas reales", "Ronda 1",
      "Creada por Ana", "Votando" y "Falta tu voto"
    Y la campana tiene aria-expanded verdadero

  Escenario: Cerrar la lista con Escape
    Dado que la lista de la campana está abierta
    Cuando presiono Escape
    Entonces la lista se cierra
    Y el foco vuelve a la campana

  Escenario: Cerrar la lista con un clic fuera
    Dado que la lista de la campana está abierta
    Cuando hago clic fuera de la campana
    Entonces la lista se cierra

  Escenario: Navegar a la sesión desde la campana
    Dado que la campana lista la sesión 7 de la historia 10
    Cuando hago clic en la sesión 7
    Entonces navego a "/proyectos/1/backlog/10/poker"

  Escenario: El toast se anuncia de forma accesible
    Dado que aparece una sesión nueva "Votando"
    Cuando se muestra el toast
    Entonces el toast tiene aria-live "polite" y role "status"

  Escenario: Error de polling conserva la última lista
    Dado que la campana lista la sesión 7
    Y la siguiente consulta falla por red
    Cuando pasa el intervalo
    Entonces la campana sigue mostrando la sesión 7
    Y no veo un toast de error

  Escenario: El backlog usa el contexto compartido
    Dado que el proyecto 1 tiene la sesión 7 "Votando" de la historia 10
    Cuando abro el backlog del proyecto
    Entonces la historia 10 muestra "En votación"
    Y el backlog no realiza su propia consulta de sesiones activas
