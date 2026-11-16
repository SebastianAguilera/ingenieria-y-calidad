# language: es
# =============================================================================
# US-20 - Métricas de proyecto y sprint (núcleo §7)
# Issue: #101
# Especificación SDD: docs/specs/US-20-metricas-spec.md
# Endpoints:
#   GET /api/proyectos/{id}/sprints/{sprintId}/metricas
#   GET /api/proyectos/{id}/metricas
# Dependencias: US-02, US-03, US-04, US-05, US-06, US-09, US-11, US-18,
#               US-19 y US-F10.
# Trazabilidad: Issue #101 -> SDD -> Criterios de Aceptación (CA-01..CA-22)
#               -> Escenarios BDD (E-01..E-12) -> Tests Go (TDD) -> Código
# Total: 12 escenarios físicos (E-01 a E-12). E-05, E-06, E-08, E-09, E-10 y
#        E-12 usan "Esquema del escenario" con "Ejemplos" para cubrir varios
#        casos.
#
# Mapa de trazabilidad CA -> Escenario:
#   CA-01                 -> E-01
#   CA-02                 -> E-02
#   CA-03                 -> E-04
#   CA-04, CA-05, CA-17   -> E-05
#   CA-06, CA-07          -> E-03, E-06
#   CA-08                 -> E-07
#   CA-09, CA-10, CA-19   -> E-08
#   CA-11                 -> E-10
#   CA-12, CA-21          -> E-11
#   CA-13, CA-14, CA-20   -> E-12
#   CA-15                 -> E-11
#   CA-16                 -> E-03
#   CA-18                 -> E-09
#   CA-22                 -> E-04, E-09
#
# Nota: la API no redondea; los valores con 2 decimales en el texto de los
# escenarios son solo de lectura. Los nombres de campos JSON son snake_case
# según el contrato de la SDD. Cada escenario se automatiza como subtest
# t.Run("<ID> ...") en los *_test.go de domain, service, handler, repository y
# cmd/api.
# =============================================================================

Característica: US-20 - Métricas de proyecto y sprint (núcleo §7)
  Como miembro de un proyecto
  Quiero consultar cuánto se planificó, cuánto se completó y qué tan bien se estimó
  Para conocer el avance y la calidad del Sprint y del proyecto sin modificar datos

  Antecedentes:
    Dado que la API esta levantada y responde "200 OK" en "GET /health"
    Y que existe un proyecto con historias, Sprints, asignaciones y defectos
    Y que el usuario es miembro del proyecto
    Y que las métricas se calculan al momento de la consulta con "base" "estado_actual"

  # ===========================================================================
  # MÉTRICAS DE UN SPRINT
  # ===========================================================================

  Regla: Métricas calculadas para un Sprint del proyecto

    @caso_normal
    Escenario: E-01 - Sprint con historias en estados mixtos calcula SP planificados, completados, porcentaje e historias completadas
      # CA-01
      Dado que existe un proyecto con el Sprint 10
      Y que el Sprint 10 tiene asignadas 3 historias:
        | historia | story_points | estado      |
        | A        | 5            | Hecho       |
        | B        | 3            | En Progreso |
        | C        | null         | Backlog     |
      Cuando envio "GET" a "/api/proyectos/1/sprints/10/metricas"
      Entonces la respuesta tiene el estado HTTP "200"
      Y el cuerpo contiene "story_points_planificados" igual a "8"
      Y el cuerpo contiene "story_points_completados" igual a "5"
      Y el cuerpo contiene "porcentaje_completado" igual a "0.625"
      Y el cuerpo contiene "historias_completadas" igual a "1"

    @caso_limite
    Escenario: E-02 - Sprint sin historias asignadas devuelve ceros y porcentaje nulo
      # CA-02
      Dado que existe un proyecto con el Sprint 10
      Y que el Sprint 10 no tiene historias asignadas
      Cuando envio "GET" a "/api/proyectos/1/sprints/10/metricas"
      Entonces la respuesta tiene el estado HTTP "200"
      Y el cuerpo contiene "story_points_planificados" igual a "0"
      Y el cuerpo contiene "story_points_completados" igual a "0"
      Y el cuerpo contiene "porcentaje_completado" igual a null
      Y el cuerpo contiene "historias_completadas" igual a "0"
      Y el cuerpo contiene "horas_estimadas" igual a "0"
      Y el cuerpo contiene "horas_reales" igual a "0"
      Y el cuerpo contiene "desviacion_esfuerzo" igual a null

    @caso_normal
    Escenario: E-03 - Sprint con horas estimadas y reales suma totales disponibles y desviación con pares
      # CA-16, CA-07
      Dado que existe un proyecto con el Sprint 10
      Y que el Sprint 10 tiene asignadas 3 historias:
        | historia | horas_estimadas | esfuerzo_real |
        | A        | 5               | 8             |
        | B        | 5               | 4             |
        | C        | 3               | null          |
      Cuando envio "GET" a "/api/proyectos/1/sprints/10/metricas"
      Entonces la respuesta tiene el estado HTTP "200"
      Y el cuerpo contiene "base" igual a "estado_actual"
      Y el cuerpo contiene "horas_estimadas" igual a "13"
      Y el cuerpo contiene "horas_reales" igual a "12"
      Y el cuerpo contiene "desviacion_esfuerzo" igual a "0.2"

    @caso_limite
    Escenario: E-07 - Historias sin story_points aportan 0 y no alteran historias completadas
      # CA-08
      Dado que existe un proyecto con el Sprint 10
      Y que el Sprint 10 tiene asignadas 2 historias en estado "Hecho":
        | historia | story_points |
        | A        | null         |
        | B        | 5            |
      Cuando envio "GET" a "/api/proyectos/1/sprints/10/metricas"
      Entonces la respuesta tiene el estado HTTP "200"
      Y el cuerpo contiene "story_points_planificados" igual a "5"
      Y el cuerpo contiene "story_points_completados" igual a "5"
      Y el cuerpo contiene "historias_completadas" igual a "2"

    @caso_normal
    Esquema del escenario: E-10 - Un defecto detectado en un Sprint y cerrado en otro se cuenta en cada Sprint según su rol
      # CA-11
      Dado que existe un defecto detectado en el Sprint 2 y cerrado en el Sprint 3
      Cuando envio "GET" a "/api/proyectos/1/sprints/<sprint>/metricas"
      Entonces la respuesta tiene el estado HTTP "200"
      Y el cuerpo contiene "defectos_detectados" igual a <detectados>
      Y el cuerpo contiene "defectos_resueltos" igual a <resueltos>

      Ejemplos:
        | caso                 | sprint | detectados | resueltos |
        | Sprint de detección  | 2      | 1          | 0         |
        | Sprint de resolución | 3      | 0          | 1         |

  # ===========================================================================
  # MÉTRICAS DEL PROYECTO
  # ===========================================================================

  Regla: Métricas agregadas del proyecto completo

    @caso_normal
    Escenario: E-04 - Dos Sprints cerrados con 8 y 4 SP completados dan velocidad 6
      # CA-03, CA-22
      Dado que existe un proyecto con un Sprint cerrado con 8 story points completados
      Y que existe un Sprint cerrado con 4 story points completados
      Y que el proyecto tiene 4 historias, 2 en estado "Hecho"
      Cuando envio "GET" a "/api/proyectos/1/metricas"
      Entonces la respuesta tiene el estado HTTP "200"
      Y el cuerpo contiene "velocidad" igual a "6"
      Y el cuerpo contiene "porcentaje_historias_completadas" igual a "0.5"

    @caso_limite
    Esquema del escenario: E-05 - Velocidad del proyecto según los Sprints cerrados
      # CA-04, CA-05, CA-17
      Dado un proyecto con <sprints>
      Cuando envio "GET" a "/api/proyectos/1/metricas"
      Entonces la respuesta tiene el estado HTTP "200"
      Y el cuerpo contiene "velocidad" igual a <velocidad>

      Ejemplos:
        | caso                        | sprints                                                        | velocidad |
        | sin cerrados                | un Sprint Activo con 20 story points completados               | null      |
        | un cerrado sin historias    | un Sprint Cerrado sin historias                                | 0         |
        | cerrados con SP 0           | dos Sprints Cerrados con 0 story points completados            | 0         |
        | Activo excluido             | un Sprint Cerrado con 8 y un Sprint Activo con 20 completados  | 8         |

    @caso_normal @caso_limite
    Esquema del escenario: E-06 - Desviación de esfuerzo agregada del proyecto con pares de horas
      # CA-06, CA-07
      # Los totales "horas_estimadas" y "horas_reales" suman todos los valores
      # disponibles; la "desviacion_esfuerzo" usa solo el subconjunto pareado.
      Dado un proyecto con historias de horas: <historias>
      Cuando envio "GET" a "/api/proyectos/1/metricas"
      Entonces la respuesta tiene el estado HTTP "200"
      Y el cuerpo contiene "horas_estimadas" igual a <horas_estimadas>
      Y el cuerpo contiene "horas_reales" igual a <horas_reales>
      Y el cuerpo contiene "desviacion_esfuerzo" igual a <desviacion_esfuerzo>

      Ejemplos:
        | caso                                | historias                                       | horas_estimadas | horas_reales | desviacion_esfuerzo |
        | dos pares                           | A (est 5, real 8) y B (est 5, real 4)           | 10              | 12           | 0.2                 |
        | una sin horas reales queda excluida | A (est 5, real 8) y B (est 5, real null)        | 10              | 8            | 0.6                 |
        | sin pares                           | A (est 5, real null) y B (est null, real 4)     | 5               | 4            | null                |

    @caso_normal @caso_limite
    Esquema del escenario: E-08 - Ratio de defectos del proyecto excluye rejected del numerador
      # CA-09, CA-10, CA-19
      Dado un proyecto con <historias_completadas> historias completadas y <defectos>
      Cuando envio "GET" a "/api/proyectos/1/metricas"
      Entonces la respuesta tiene el estado HTTP "200"
      Y el cuerpo contiene "defectos_detectados" igual a <defectos_detectados>
      Y el cuerpo contiene "ratio_defectos" igual a <ratio_defectos>

      Ejemplos:
        | caso                      | historias_completadas | defectos                             | defectos_detectados | ratio_defectos |
        | un rejected entre tres    | 6                     | 3 detectados, 1 de ellos "rejected"  | 3                   | 0.3333333333333333 |
        | sin historias completadas | 0                     | 3 detectados, 1 de ellos "rejected"  | 3                   | null           |
        | solo rejected             | 6                     | 1 detectado "rejected"               | 1                   | 0              |

    @caso_limite
    Esquema del escenario: E-09 - Proyecto sin historias o sin historias completadas devuelve porcentajes nulos o cero
      # CA-18, CA-22
      Dado un proyecto con <historias> y sin Sprints cerrados
      Cuando envio "GET" a "/api/proyectos/1/metricas"
      Entonces la respuesta tiene el estado HTTP "200"
      Y el cuerpo contiene "porcentaje_historias_completadas" igual a <porcentaje>
      Y el cuerpo contiene "ratio_defectos" igual a <ratio>
      Y el cuerpo contiene "velocidad" igual a null
      Y el cuerpo contiene "horas_estimadas" igual a "0"
      Y el cuerpo contiene "horas_reales" igual a "0"
      Y el cuerpo contiene "desviacion_esfuerzo" igual a null

      Ejemplos:
        | caso                | historias                                              | porcentaje | ratio |
        | sin historias       | sin historias                                          | null       | null  |
        | sin historias Hecho | 4 historias sin horas, 0 en estado "Hecho"              | 0          | null  |

  # ===========================================================================
  # CONTRATO DE RESPUESTA Y SOLO LECTURA
  # ===========================================================================

  Regla: Contrato en snake_case, base estado_actual y consultas sin efectos laterales

    @caso_normal
    Escenario: E-11 - Contrato en snake_case, base estado_actual y consulta de solo lectura
      # CA-12, CA-21, CA-15
      Dado que existe un proyecto con el Sprint 10
      Cuando consulto las métricas del Sprint 10 y del proyecto
      Entonces la respuesta del Sprint tiene el estado HTTP "200"
      Y el cuerpo del Sprint contiene "base" igual a "estado_actual"
      Y el cuerpo del Sprint contiene exactamente las claves "sprint_id", "base", "story_points_planificados", "story_points_completados", "porcentaje_completado", "historias_completadas", "horas_estimadas", "horas_reales", "desviacion_esfuerzo", "defectos_detectados" y "defectos_resueltos"
      Y la respuesta del proyecto tiene el estado HTTP "200"
      Y el cuerpo del proyecto contiene "base" igual a "estado_actual"
      Y el cuerpo del proyecto contiene exactamente las claves "proyecto_id", "base", "velocidad", "porcentaje_historias_completadas", "horas_estimadas", "horas_reales", "desviacion_esfuerzo", "defectos_detectados", "defectos_resueltos" y "ratio_defectos"
      Y el cuerpo del proyecto no contiene "historias_completadas" ni "historias_totales"
      Y los datos de historias, Sprints, asignaciones y defectos no cambian

  # ===========================================================================
  # ACCESO Y VALIDACIÓN DE PARÁMETROS
  # ===========================================================================

  Regla: Acceso por membresía, existencia del Sprint y validación de parámetros

    @caso_error
    Esquema del escenario: E-12 - Errores de acceso, Sprint y parámetros
      # CA-13, CA-14, CA-20
      Dado que <contexto>
      Cuando envio "GET" a "<endpoint>"
      Entonces la respuesta tiene el estado HTTP "<http>"
      Y el cuerpo contiene el error "<codigo>"

      Ejemplos:
        | caso                    | contexto                              | endpoint                               | http | codigo                 |
        | no miembro en Sprint    | el usuario no es miembro del proyecto | /api/proyectos/1/sprints/10/metricas   | 404  | PROYECTO_NO_ENCONTRADO |
        | no miembro en proyecto  | el usuario no es miembro del proyecto | /api/proyectos/1/metricas              | 404  | PROYECTO_NO_ENCONTRADO |
        | Sprint de otro proyecto | el Sprint 10 pertenece a otro proyecto| /api/proyectos/1/sprints/10/metricas   | 404  | SPRINT_NO_ENCONTRADA   |
        | Sprint inexistente      | el Sprint 999 no existe en el proyecto| /api/proyectos/1/sprints/999/metricas  | 404  | SPRINT_NO_ENCONTRADA   |
        | id 0                    | el identificador de proyecto es 0     | /api/proyectos/0/metricas              | 400  | PARAMETRO_INVALIDO     |
        | sprintId no numérico    | el identificador de Sprint no es numérico | /api/proyectos/1/sprints/abc/metricas | 400  | PARAMETRO_INVALIDO     |
