# language: es
# =============================================================================
# US-21 - Extensiones de métricas (flujo y defectos)
# Issue: #102
# Especificación SDD: docs/specs/US-21-extensiones-metricas-spec.md
# Endpoint: GET /api/proyectos/{id}/metricas (extiende US-20)
# Dependencias: US-02, US-03, US-04, US-05, US-09, US-11, US-20 y US-F10.
# Trazabilidad: Issue #102 -> SDD -> Criterios de Aceptación (CA-01..CA-15)
#               -> Escenarios BDD (E-01..E-11) -> Tests Go (TDD) -> Código
# Total: 11 escenarios físicos (E-01 a E-11). E-02, E-05 y E-11 usan
#        "Esquema del escenario" con "Ejemplos".
#
# Mapa de trazabilidad CA -> Escenario:
#   CA-01                 -> E-01
#   CA-02, CA-03          -> E-02, E-03
#   CA-04                 -> E-04
#   CA-05                 -> E-06
#   CA-06                 -> E-07
#   CA-07                 -> E-09
#   CA-08                 -> E-08
#   CA-09                 -> E-09
#   CA-10, CA-11          -> E-05
#   CA-12, CA-13          -> E-10
#   CA-14, CA-15          -> E-11
#
# Nota: la API no redondea salvo el "ceil" de sprints_restantes. Los nombres de
# campos JSON son snake_case según el contrato de la SDD. Cada escenario se
# automatiza como subtest t.Run("<ID> ...") en los *_test.go de domain, service
# y handler.
# =============================================================================

Característica: US-21 - Extensiones de métricas de flujo y defectos
  Como miembro de un proyecto
  Quiero ver indicadores de flujo y calidad además de las métricas de avance
  Para estimar cuántos Sprints faltan, cuánto trabajo está en curso y cuánto tarda una historia

  Antecedentes:
    Dado que la API esta levantada y responde "200 OK" en "GET /health"
    Y que existe un proyecto con historias, Sprints, asignaciones y defectos
    Y que el usuario es miembro del proyecto
    Y que las métricas se calculan al momento de la consulta con "base" "estado_actual"

  # ===========================================================================
  # SPRINTS RESTANTES Y TRABAJO EN CURSO
  # ===========================================================================

  Regla: Estimación de sprints restantes y trabajo en curso

    @caso_normal
    Escenario: E-01 - 40 SP pendientes y velocidad 6 dan 7 sprints restantes
      # CA-01
      Dado que existe un proyecto con 40 story points en historias que no estan "Hecho"
      Y que el proyecto tiene dos Sprints cerrados con 6 story points completados en promedio
      Cuando envio "GET" a "/api/proyectos/1/metricas"
      Entonces la respuesta tiene el estado HTTP "200"
      Y el cuerpo contiene "velocidad" igual a "6"
      Y el cuerpo contiene "sprints_restantes" igual a "7"

    @caso_limite
    Esquema del escenario: E-02 - Sprints restantes según la velocidad disponible
      # CA-02, CA-03
      Dado un proyecto con <sprints> y <pendientes> story points pendientes
      Cuando envio "GET" a "/api/proyectos/1/metricas"
      Entonces la respuesta tiene el estado HTTP "200"
      Y el cuerpo contiene "sprints_restantes" igual a <sprints_restantes>

      Ejemplos:
        | caso                | sprints                              | pendientes | sprints_restantes |
        | sin cerrados        | un Sprint Activo con historias        | 40         | null              |
        | velocidad cero      | dos Sprints Cerrados con 0 SP         | 40         | null              |
        | sin pendientes      | un Sprint Cerrado con 6 SP completados| 0          | 0                 |
        | cociente exacto     | un Sprint Cerrado con 6 SP completados| 12         | 2                 |
        | con fraccion        | un Sprint Cerrado con 3 SP completados| 7          | 3                 |

    @caso_normal
    Escenario: E-03 - Seis historias en En Progreso dan WiP 6
      # CA-03
      Dado que existe un proyecto con 6 historias en estado "En Progreso"
      Cuando envio "GET" a "/api/proyectos/1/metricas"
      Entonces la respuesta tiene el estado HTTP "200"
      Y el cuerpo contiene "wip" igual a 6

  # ===========================================================================
  # TASA DE ENTREGA Y LEAD TIME
  # ===========================================================================

  Regla: Tasa de entrega y Lead Time estimado (Ley de Little)

    @caso_normal
    Escenario: E-04 - Dos Sprints cerrados con 3 y 1 historias Hecho dan tasa de entrega 2
      # CA-04
      Dado que existe un proyecto con dos Sprints cerrados
      Y que el primer Sprint cerrado tiene 3 historias asignadas en estado "Hecho"
      Y que el segundo Sprint cerrado tiene 1 historia asignada en estado "Hecho"
      Cuando envio "GET" a "/api/proyectos/1/metricas"
      Entonces la respuesta tiene el estado HTTP "200"
      Y el cuerpo contiene "tasa_entrega" igual a "2"

    @caso_limite
    Esquema del escenario: E-05 - Tasa de entrega según los Sprints cerrados
      # CA-05, CA-10, CA-11
      Dado un proyecto con <sprints>
      Cuando envio "GET" a "/api/proyectos/1/metricas"
      Entonces la respuesta tiene el estado HTTP "200"
      Y el cuerpo contiene "tasa_entrega" igual a <tasa_entrega>

      Ejemplos:
        | caso                         | sprints                                                                 | tasa_entrega |
        | sin cerrados                 | un Sprint Activo con 5 historias Hecho                                 | null         |
        | un cerrado vacio             | un Sprint Cerrado sin historias                                        | 0            |
        | activo excluido              | un Sprint Cerrado con 2 Hecho y un Sprint Activo con 5 Hecho            | 2            |
        | cerrado vacio en el divisor  | un Sprint Cerrado con 2 Hecho y otro Sprint Cerrado sin historias       | 1            |
        | historia no Hecho no cuenta  | un Sprint Cerrado con 1 Hecho y 1 En Progreso                           | 1            |
        | historia no asignada no cuenta | un Sprint Cerrado con 1 Hecho y otra historia Hecho sin Sprint        | 1            |

    @caso_normal
    Escenario: E-06 - WiP 6 y tasa de entrega 2 dan Lead Time 3 aproximado
      # CA-05
      Dado que existe un proyecto con 6 historias en estado "En Progreso"
      Y que la tasa de entrega es 2
      Cuando envio "GET" a "/api/proyectos/1/metricas"
      Entonces la respuesta tiene el estado HTTP "200"
      Y el cuerpo contiene "lead_time_estimado" igual a "3"
      Y el cuerpo contiene "lead_time_aproximado" igual a true

    @caso_limite
    Esquema del escenario: E-07 - Lead Time nulo cuando no hay tasa de entrega
      # CA-06
      Dado un proyecto con <sprints>
      Cuando envio "GET" a "/api/proyectos/1/metricas"
      Entonces la respuesta tiene el estado HTTP "200"
      Y el cuerpo contiene "lead_time_estimado" igual a null
      Y el cuerpo contiene "lead_time_aproximado" igual a false

      Ejemplos:
        | caso            | sprints                                           |
        | sin cerrados    | un Sprint Activo con 6 historias En Progreso      |
        | tasa cero       | un Sprint Cerrado sin historias Hecho             |

  # ===========================================================================
  # DEFECTOS
  # ===========================================================================

  Regla: Defectos abiertos y desglose por severidad

    @caso_normal
    Escenario: E-08 - Defectos abiertos cuenta solo los que están en open
      # CA-08
      Dado que existe un proyecto con 1 defecto "open", 1 "in_progress", 1 "resolved" y 1 "rejected"
      Cuando envio "GET" a "/api/proyectos/1/metricas"
      Entonces la respuesta tiene el estado HTTP "200"
      Y el cuerpo contiene "defectos_abiertos" igual a 1

    @caso_limite
    Escenario: E-09 - Desglose por severidad incluye blocker en cero
      # CA-07, CA-09
      Dado que existe un proyecto con defectos
        | severidad | estado     |
        | minor     | open       |
        | major     | resolved   |
        | critical  | open       |
        | critical  | rejected   |
      Y que el proyecto no tiene defectos "blocker"
      Cuando envio "GET" a "/api/proyectos/1/metricas"
      Entonces la respuesta tiene el estado HTTP "200"
      Y el cuerpo contiene "defectos_por_severidad" con "minor" igual a 1
      Y el cuerpo contiene "defectos_por_severidad" con "major" igual a 1
      Y el cuerpo contiene "defectos_por_severidad" con "critical" igual a 2
      Y el cuerpo contiene "defectos_por_severidad" con "blocker" igual a 0

  # ===========================================================================
  # CONTRATO DE RESPUESTA Y SOLO LECTURA
  # ===========================================================================

  Regla: Contrato extensible en snake_case y consultas sin efectos laterales

    @caso_normal
    Escenario: E-10 - Contrato con claves nuevas, base estado_actual y consulta de solo lectura
      # CA-12, CA-13
      Dado que existe un proyecto con el Sprint 10
      Cuando envio "GET" a "/api/proyectos/1/metricas"
      Entonces la respuesta tiene el estado HTTP "200"
      Y el cuerpo contiene "base" igual a "estado_actual"
      Y el cuerpo contiene todas las claves de US-20 y además "sprints_restantes", "wip", "tasa_entrega", "lead_time_estimado", "lead_time_aproximado", "defectos_abiertos" y "defectos_por_severidad"
      Y el objeto "defectos_por_severidad" contiene exactamente las claves "minor", "major", "critical" y "blocker"
      Y los datos de historias, Sprints, asignaciones y defectos no cambian

  # ===========================================================================
  # ACCESO Y VALIDACIÓN DE PARÁMETROS
  # ===========================================================================

  Regla: Acceso por membresía y validación de parámetros

    @caso_error
    Esquema del escenario: E-11 - Errores de acceso y parámetros
      # CA-14, CA-15
      Dado que <contexto>
      Cuando envio "GET" a "<endpoint>"
      Entonces la respuesta tiene el estado HTTP "<http>"
      Y el cuerpo contiene el error "<codigo>"

      Ejemplos:
        | caso          | contexto                              | endpoint                    | http | codigo                 |
        | no miembro    | el usuario no es miembro del proyecto | /api/proyectos/1/metricas   | 404  | PROYECTO_NO_ENCONTRADO |
        | id cero       | el identificador de proyecto es 0     | /api/proyectos/0/metricas   | 400  | PARAMETRO_INVALIDO     |
        | id no numerico| el identificador no es numérico       | /api/proyectos/abc/metricas | 400  | PARAMETRO_INVALIDO     |
