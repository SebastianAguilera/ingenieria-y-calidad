# language: es
# =============================================================================
# US-F22 - Extensiones de métricas en el dashboard
# Escenarios BDD (Gherkin)
#
# Especificacion SDD : docs/specs/US-F22-extensiones-metricas-spec.md
# Issue de origen    : SebastianAguilera/ingenieria-y-calidad#104
# Endpoint           : US-20 / US-21
#                      GET /api/proyectos/{id}/metricas
# Dependencias       : US-21 (campos nuevos del endpoint), US-F21 tramo 1
#                      (pantalla de metricas, tarjeta KPI, grafico)
# Trazabilidad       : Issue #104 -> SDD -> criterios -> escenarios BDD ->
#                      tests Vitest/RTL -> codigo
# Rama               : feature/US-F22-extensiones-metricas
# =============================================================================
#
# NOTA: cada escenario se automatiza como test de Vitest + React Testing
# Library sobre MetricasPage, mockeando el cliente de metricas de
# frontend/src/api. Esta historia no agrega llamadas a la API.
# -----------------------------------------------------------------------------
# NARRATIVA (US-F22)
# -----------------------------------------------------------------------------
# Como miembro de un proyecto, quiero ver los indicadores de flujo (sprints
# restantes, WiP, tasa de entrega y Lead Time) y el desglose de defectos por
# severidad en el dashboard, para estimar cuanto falta y donde esta el trabajo.
# -----------------------------------------------------------------------------
# CRITERIOS DE ACEPTACION / SDD seccion "Criterios de aceptacion"
# -----------------------------------------------------------------------------
# [CA-01] Sprints restantes "N sprints".                          -> E-01, E-02
# [CA-02] Sprints restantes null: "Sin datos".                    -> E-03
# [CA-03] Lead Time con "aprox." y nota de la Ley de Little.      -> E-04
# [CA-04] Lead Time o tasa null: "Sin datos", sin "aprox.".       -> E-05, E-06
# [CA-05] WiP y defectos abiertos coinciden con la API.           -> E-07..E-09
# [CA-06] blocker visible con valor 0.                            -> E-10
# [CA-07] Orden fijo de severidades.                              -> E-11
# [CA-08] Resumen en texto del grafico.                           -> E-12, E-13
# [CA-09] Campos ausentes (backend viejo): se ocultan.            -> E-14, E-15
# [CA-10] Integracion con la pantalla de F21 y sus estados.       -> E-16..E-18
# [CA-11] Formato de numeros.                                     -> E-19
# [CA-12] Legibilidad sin depender solo del color.                -> E-20
# =============================================================================

Funcionalidad: US-F22 - Extensiones de métricas en el dashboard

  Como miembro de un proyecto
  Quiero ver los indicadores de flujo y el desglose de defectos por severidad
  Para estimar cuánto falta y dónde está el trabajo.

  Antecedentes:
    Dado que "ana" es miembro del proyecto "Proyecto Alfa"
    Y que "ana" está en la pantalla "/proyectos/1/metricas"

  Regla: Indicadores de flujo (RB-01, RB-02, RB-03, RB-04, RB-05)

    Escenario: E-01 - Sprints restantes se muestra en plural
      Dado que la API devuelve "sprints_restantes" igual a 7
      Cuando "ana" entra a la pantalla Métricas
      Entonces ve la tarjeta "Sprints restantes" con el valor "7 sprints"

    Escenario: E-02 - Sprints restantes se muestra en singular y en cero
      Dado que la API devuelve "sprints_restantes" igual a 1
      Cuando "ana" entra a la pantalla Métricas
      Entonces ve la tarjeta "Sprints restantes" con el valor "1 sprint"
      Cuando la API devuelve "sprints_restantes" igual a 0
      Entonces ve la tarjeta "Sprints restantes" con el valor "0 sprints"

    Escenario: E-03 - Sprints restantes sin velocidad muestra "Sin datos"
      Dado que la API devuelve "sprints_restantes" igual a null
      Cuando "ana" entra a la pantalla Métricas
      Entonces la tarjeta "Sprints restantes" muestra "Sin datos"
      Y ve la explicación "Hace falta la velocidad de al menos un sprint cerrado."
      Y la tarjeta no muestra "0"

    Escenario: E-04 - Lead Time estimado muestra la etiqueta "aprox." y la nota
      Dado que la API devuelve "lead_time_estimado" igual a 3
      Y devuelve "lead_time_aproximado" igual a true
      Cuando "ana" entra a la pantalla Métricas
      Entonces ve la tarjeta "Lead Time estimado" con el valor "3 sprints"
      Y ve la etiqueta "aprox."
      Y ve una nota que explica que es una estimación según la Ley de Little
        y no una medición

    Escenario: E-05 - Lead Time null muestra "Sin datos" sin "aprox."
      Dado que la API devuelve "lead_time_estimado" igual a null
      Y devuelve "lead_time_aproximado" igual a false
      Cuando "ana" entra a la pantalla Métricas
      Entonces la tarjeta "Lead Time estimado" muestra "Sin datos"
      Y ve la explicación "Hace falta la tasa de entrega de al menos un sprint cerrado."
      Y no ve la etiqueta "aprox."
      Y la tarjeta no muestra "0"

    Escenario: E-06 - Tasa de entrega null muestra "Sin datos" y un cero real se muestra
      Dado que la API devuelve "tasa_entrega" igual a null
      Cuando "ana" entra a la pantalla Métricas
      Entonces la tarjeta "Tasa de entrega" muestra "Sin datos"
      Cuando la API devuelve "tasa_entrega" igual a 0
      Entonces la tarjeta "Tasa de entrega" muestra "0"
      Y no muestra "Sin datos"

    Escenario: E-07 - WiP coincide con la API
      Dado que la API devuelve "wip" igual a 6
      Cuando "ana" entra a la pantalla Métricas
      Entonces ve la tarjeta "WiP" con el valor "6"

    Escenario: E-08 - WiP en cero se muestra como cero
      Dado que la API devuelve "wip" igual a 0
      Cuando "ana" entra a la pantalla Métricas
      Entonces ve la tarjeta "WiP" con el valor "0"

  Regla: Defectos abiertos (RB-01, RB-05)

    Escenario: E-09 - Defectos abiertos coincide con la API
      Dado que la API devuelve "defectos_abiertos" igual a 4
      Cuando "ana" entra a la pantalla Métricas
      Entonces ve la tarjeta "Defectos abiertos" con el valor "4"
      Y ve una nota de que cuenta los defectos en estado "Abierto"

  Regla: Gráfico de defectos por severidad (RB-06, RB-07, RB-08)

    Escenario: E-10 - Blocker aparece con valor 0
      Dado que la API devuelve "defectos_por_severidad" con
        minor 1, major 1, critical 1 y blocker 0
      Cuando "ana" entra a la pantalla Métricas
      Entonces ve el gráfico "Defectos por severidad"
      Y ve la barra "Bloqueante" con el valor 0

    Escenario: E-11 - Las severidades van en orden fijo
      Dado que la API devuelve "defectos_por_severidad" con
        minor 0, major 5, critical 2 y blocker 9
      Cuando "ana" entra a la pantalla Métricas
      Entonces las barras aparecen en el orden "Menor", "Mayor", "Crítica",
        "Bloqueante"

    Escenario: E-12 - El gráfico tiene un resumen en texto con los mismos valores
      Dado que la API devuelve "defectos_por_severidad" con
        minor 1, major 1, critical 1 y blocker 0
      Cuando "ana" entra a la pantalla Métricas
      Entonces existe un resumen en texto accesible
      Y el resumen dice "Menor: 1, Mayor: 1, Crítica: 1, Bloqueante: 0"

    Escenario: E-13 - Sin defectos el resumen lo indica
      Dado que la API devuelve "defectos_por_severidad" con las cuatro
        severidades en 0
      Cuando "ana" entra a la pantalla Métricas
      Entonces ve las cuatro barras con valor 0
      Y el resumen dice "Sin defectos registrados."

  Regla: Compatibilidad con un backend sin US-21 (RB-09)

    Escenario: E-14 - Una respuesta sin los campos de US-21 no muestra las tarjetas nuevas
      Dado que la API devuelve solo los campos de US-20
      Cuando "ana" entra a la pantalla Métricas
      Entonces no ve "Sprints restantes", "WiP", "Tasa de entrega",
        "Lead Time estimado", "Defectos abiertos" ni "Defectos por severidad"
      Y no ve "Sin datos" por esos campos
      Y no ve ningún error
      Y sigue viendo las tarjetas de US-F21

    Escenario: E-15 - Una clave faltante de severidad se toma como cero
      Dado que la API devuelve "defectos_por_severidad" con
        minor 2, major 1 y critical 0, sin la clave "blocker"
      Cuando "ana" entra a la pantalla Métricas
      Entonces ve la barra "Bloqueante" con el valor 0

  Regla: Integración con la pantalla de US-F21 (RB-10, RB-11, RB-12)

    Escenario: E-16 - Las tarjetas nuevas aparecen debajo de las de F21
      Dado que la API devuelve todos los campos de US-20 y de US-21
      Cuando "ana" entra a la pantalla Métricas
      Entonces ve primero los KPI y gráficos de US-F21
      Y debajo ve las tarjetas de flujo y el gráfico de severidad
      Y las tarjetas de US-F21 conservan su contenido

    Escenario: E-17 - El aviso de estado actual no se duplica
      Dado que la API devuelve todos los campos de US-20 y de US-21
      Cuando "ana" entra a la pantalla Métricas
      Entonces ve una sola vez "Calculado sobre el estado actual de las historias"

    Escenario: E-18 - Los estados de carga, error y 404 son los de la pantalla
      Dado que la API responde 404 PROYECTO_NO_ENCONTRADO
      Cuando "ana" entra a la pantalla Métricas
      Entonces ve "No tenés acceso a este proyecto o no existe"
      Y no ve ninguna tarjeta de US-F22

  Regla: Formato de números (RB-13)

    Escenario: E-19 - Los números se muestran con formato local
      Dado que la API devuelve "tasa_entrega" igual a 2.5
      Cuando "ana" entra a la pantalla Métricas
      Entonces la tarjeta "Tasa de entrega" muestra "2,5"
      Cuando la API devuelve "tasa_entrega" igual a 2
      Entonces la tarjeta "Tasa de entrega" muestra "2" y no "2,00"

  Regla: Accesibilidad (RB-07, RB-08)

    Escenario: E-20 - El gráfico no depende solo del color
      Dado que la API devuelve "defectos_por_severidad" con valores distintos
      Cuando "ana" entra a la pantalla Métricas
      Entonces cada barra muestra su etiqueta de severidad y su valor en texto
      Y el gráfico es legible en tema claro y en tema oscuro
