# language: es
# =============================================================================
# US-F21 - Dashboard de metricas y carga de horas estimadas
# Escenarios BDD (Gherkin)
#
# Especificacion SDD : docs/specs/US-F21-dashboard-metricas-spec.md
# Issue de origen    : SebastianAguilera/ingenieria-y-calidad#103
# Endpoints          : US-20 (tramo 1)
#                      GET   /api/proyectos/{id}/metricas
#                      GET   /api/proyectos/{id}/sprints/{sprintId}/metricas
#                      US-19 (tramo 2)
#                      POST  /api/proyectos/{id}/historias        (horas_estimadas)
#                      PATCH /api/proyectos/{id}/historias/{hid}  (horas_estimadas)
# Trazabilidad       : Issue #103 -> SDD -> criterios -> escenarios BDD ->
#                      tests Vitest/RTL -> codigo
# Rama               : feature/US-F21-dashboard-metricas
# =============================================================================
#
# NOTA: cada escenario se automatiza como test de Vitest + React Testing
# Library. Los tests de la API usan mocks de frontend/src/api/client.ts y los
# de pantalla usan mocks de frontend/src/api/metricas.ts, historias.ts y
# sprints.ts.
# -----------------------------------------------------------------------------
# CRITERIOS DE ACEPTACION -> ESCENARIOS
# -----------------------------------------------------------------------------
# TRAMO 1
# [CA-1]  KPI con los valores de GET /metricas.            -> E-01
# [CA-2]  null se muestra como "Sin datos" + explicacion.   -> E-02, E-03
# [CA-3]  0.75 se muestra como 75%.                         -> E-04, E-05
# [CA-4]  Barras planificados/completados por sprint.       -> E-06, E-07
# [CA-5]  Dona con cantidad y etiqueta por estado.          -> E-08
# [CA-6]  Selector de sprint con sus metricas.              -> E-09..E-11
# [CA-7]  Estado vacio sin historias.                       -> E-12
# [CA-8]  Resumen en texto de cada grafico.                 -> E-13
# [CA-9]  Aviso de estado actual visible.                   -> E-14
# [CA-10] No miembro: mensaje de acceso.                    -> E-15
#         Error de red y reintento.                         -> E-16
# TRAMO 2
# [CA-11] Horas > 0 se guardan y la tabla las muestra.      -> E-17, E-18
# [CA-12] Campo vacio: sin valor y "—" en la tabla.         -> E-19, E-20
# [CA-13] 0 o negativo: error en el campo, sin API.         -> E-21, E-22
# [CA-14] 422 del backend en el campo.                      -> E-23
# [CA-15] Edicion con el campo precargado.                  -> E-24
# -----------------------------------------------------------------------------

Característica: Dashboard de métricas y carga de horas estimadas
  Como miembro de un proyecto
  Quiero ver el avance del proyecto y de cada sprint en gráficos y cargar las horas estimadas
  Para entender el estado sin leer tablas y para que la desviación de esfuerzo tenga datos

  Antecedentes:
    Dado que inicié sesión como miembro del proyecto 1
    Y el proyecto tiene los sprints "Sprint 1" (Cerrado), "Sprint 2" (Activo) y "Sprint 3" (Planificado)

  # ---------------------------------------------------------------------------
  # TRAMO 1 - Dashboard
  # ---------------------------------------------------------------------------

  @tramo1 @CA-1
  Escenario: E-01 KPI con los valores del proyecto
    Dado que GET /api/proyectos/1/metricas responde velocidad 8, porcentaje 0.5, desviacion 0.2, ratio 0.25, 4 detectados y 3 resueltos
    Cuando entro a "Métricas"
    Entonces veo la tarjeta "Velocidad" con "8 SP"
    Y veo la tarjeta "Historias completadas" con "50%"
    Y veo la tarjeta "Desviación de esfuerzo" con "+20%"
    Y veo la tarjeta "Ratio de defectos" con "0,25"
    Y veo la tarjeta "Defectos detectados" con "4"
    Y veo la tarjeta "Defectos resueltos" con "3"

  @tramo1 @CA-2
  Escenario: E-02 Métricas nulas se muestran como "Sin datos"
    Dado que la API responde velocidad, porcentaje, desviacion y ratio en null
    Cuando entro a "Métricas"
    Entonces cada una de esas tarjetas muestra "Sin datos"
    Y la tarjeta "Velocidad" explica "Todavía no hay sprints cerrados"
    Y la tarjeta "Desviación de esfuerzo" explica "Ninguna historia tiene horas estimadas y reales"
    Y ninguna de esas tarjetas muestra "0"

  @tramo1 @CA-2 @limite
  Escenario: E-03 Un cero real no se confunde con "Sin datos"
    Dado que la API responde velocidad 0
    Cuando entro a "Métricas"
    Entonces la tarjeta "Velocidad" muestra "0 SP"

  @tramo1 @CA-3
  Escenario: E-04 Porcentaje de la API multiplicado por 100
    Dado que la API responde porcentaje_historias_completadas 0.75
    Cuando entro a "Métricas"
    Entonces la tarjeta "Historias completadas" muestra "75%"

  @tramo1 @CA-3 @limite
  Escenario: E-05 Porcentaje con decimales y desviación negativa
    Dado que la API responde porcentaje 0.6153846153846154 y desviacion -0.1
    Cuando entro a "Métricas"
    Entonces la tarjeta "Historias completadas" muestra "61,5%"
    Y la tarjeta "Desviación de esfuerzo" muestra "−10%"

  @tramo1 @CA-4
  Escenario: E-06 Barras planificados contra completados por sprint
    Dado que "Sprint 1" tiene 13 SP planificados y 8 completados
    Y "Sprint 2" tiene 10 SP planificados y 2 completados
    Cuando entro a "Métricas"
    Entonces el gráfico "Story points por sprint" tiene una fila por "Sprint 1" y "Sprint 2"
    Y cada fila tiene "SP planificados" y "SP completados"
    Y "Sprint 3" no aparece porque está planificado

  @tramo1 @CA-4 @alternativo
  Escenario: E-07 Proyecto sin sprints cerrados ni activo
    Dado que el proyecto solo tiene sprints planificados
    Cuando entro a "Métricas"
    Entonces el gráfico de barras dice "Todavía no hay sprints cerrados ni activos"

  @tramo1 @CA-5
  Escenario: E-08 Dona de historias por estado
    Dado que el backlog tiene 2 historias en "Backlog", 1 en "En Progreso" y 3 en "Hecho"
    Cuando entro a "Métricas"
    Entonces el gráfico "Historias por estado" muestra "Backlog" con 2, "En Progreso" con 1 y "Hecho" con 3

  @tramo1 @CA-6
  Escenario: E-09 Selección inicial del sprint activo
    Cuando entro a "Métricas"
    Entonces el selector de sprint tiene elegido "Sprint 2"
    Y el panel muestra las métricas de GET /api/proyectos/1/sprints/2/metricas

  @tramo1 @CA-6
  Escenario: E-10 Cambiar de sprint en el selector
    Dado que estoy en "Métricas"
    Cuando elijo "Sprint 1" en el selector
    Entonces se pide GET /api/proyectos/1/sprints/1/metricas
    Y el panel muestra "13" SP planificados, "8" completados y "61,5%"

  @tramo1 @CA-6 @limite
  Escenario: E-11 Sprint sin story points planificados
    Cuando elijo "Sprint 3" en el selector
    Entonces el porcentaje completado muestra "Sin datos"
    Y explica "El sprint no tiene story points planificados"

  @tramo1 @CA-7
  Escenario: E-12 Proyecto sin historias
    Dado que el backlog está vacío
    Cuando entro a "Métricas"
    Entonces veo "Todavía no hay historias"
    Y veo un enlace "Ir al backlog"
    Y no veo las tarjetas de KPI

  @tramo1 @CA-8
  Escenario: E-13 Resumen en texto de cada gráfico
    Cuando entro a "Métricas"
    Entonces existe la tabla "Story points por sprint" con los mismos valores del gráfico
    Y existe la tabla "Historias por estado" con los mismos valores de la dona
    Y el botón "Ver como tabla" la muestra visualmente

  @tramo1 @CA-9
  Escenario: E-14 Aviso de estado actual
    Cuando entro a "Métricas"
    Entonces veo "Calculado sobre el estado actual de las historias"

  @tramo1 @CA-10 @error
  Escenario: E-15 No miembro del proyecto
    Dado que la API responde 404 PROYECTO_NO_ENCONTRADO
    Cuando entro a "Métricas"
    Entonces veo "No tenés acceso a este proyecto o no existe"

  @tramo1 @error
  Escenario: E-16 Error de red y reintento
    Dado que la primera carga falla por error de red
    Cuando entro a "Métricas"
    Entonces veo "No se pudieron cargar las métricas. Intentá nuevamente."
    Cuando presiono "Reintentar" y la API responde bien
    Entonces veo las tarjetas de KPI

  # ---------------------------------------------------------------------------
  # TRAMO 2 - Horas estimadas
  # ---------------------------------------------------------------------------

  @tramo2 @CA-11
  Escenario: E-17 Crear una historia con horas estimadas
    Dado que abro "Nueva historia"
    Cuando completo "Título" con "Historia con horas" y "Horas estimadas" con "5"
    Y presiono "Crear historia"
    Entonces se envía horas_estimadas 5

  @tramo2 @CA-11
  Escenario: E-18 La tabla muestra las horas estimadas
    Dado una historia con horas_estimadas 5
    Cuando veo el backlog en modo tabla
    Entonces la columna "Horas estimadas" muestra "5 h"

  @tramo2 @CA-12
  Escenario: E-19 Crear sin horas estimadas
    Dado que abro "Nueva historia"
    Cuando completo solo el título y presiono "Crear historia"
    Entonces se envía horas_estimadas null

  @tramo2 @CA-12
  Escenario: E-20 La tabla muestra "—" sin horas
    Dado una historia con horas_estimadas null
    Cuando veo el backlog en modo tabla
    Entonces la columna "Horas estimadas" muestra "—"

  @tramo2 @CA-13 @error
  Esquema del escenario: E-21 Valores no positivos se rechazan en el cliente
    Dado que abro "Nueva historia"
    Cuando completo "Horas estimadas" con "<valor>" y presiono "Crear historia"
    Entonces el campo muestra "Las horas estimadas deben ser mayores a 0"
    Y el campo tiene aria-invalid
    Y no se llama a la API

    Ejemplos:
      | valor |
      | 0     |
      | -3    |

  @tramo2 @CA-13 @limite
  Escenario: E-22 Un cuarto de hora es válido
    Cuando completo "Horas estimadas" con "0.25" y guardo
    Entonces se envía horas_estimadas 0.25

  @tramo2 @CA-14 @error
  Escenario: E-23 Error 422 del backend en el campo
    Dado que el backend responde 422 VALIDACION con campo "horas_estimadas"
    Cuando guardo la historia
    Entonces el mensaje del backend aparece debajo de "Horas estimadas"

  @tramo2 @CA-15
  Escenario: E-24 Edición con el campo precargado
    Dada una historia con horas_estimadas 8
    Cuando la abro para editar
    Entonces "Horas estimadas" muestra "8"
