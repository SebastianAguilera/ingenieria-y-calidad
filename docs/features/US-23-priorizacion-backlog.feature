# language: es
# =============================================================================
# US-23 - Valor de negocio y priorización del backlog (ROI y WSJF)
# Issue: #110
# Especificación SDD: docs/specs/US-23-priorizacion-backlog-spec.md
# Endpoint:
#   GET /api/proyectos/{id}/backlog/priorizado
# Dependencias: US-02, US-05, US-11, US-19, US-20 y US-F10.
# Trazabilidad: Issue #110 -> SDD -> Criterios de Aceptación (CA-01..CA-12)
#               -> Escenarios BDD (E-01..E-10) -> Tests Go (TDD) -> Código
# Total: 10 escenarios físicos (E-01 a E-10).
#
# Mapa de trazabilidad CA -> Escenario:
#   CA-01                 -> E-01
#   CA-02                 -> E-02
#   CA-03, CA-10          -> E-03
#   CA-04, CA-11          -> E-04
#   CA-05                 -> E-05, E-06
#   CA-06                 -> E-07
#   CA-07                 -> E-08
#   CA-08                 -> E-09
#   CA-09, CA-12          -> E-10
#
# Nota: la API no redondea; los valores con 2 decimales en el texto de los
# escenarios son solo de lectura. Los campos JSON son snake_case según el
# contrato de la SDD. Cada escenario se automatiza como subtest
# t.Run("<ID> ...") en los *_test.go de domain, service, handler, repository y
# cmd/api.
# =============================================================================

Característica: US-23 - Valor de negocio y priorización del backlog
  Como líder de un proyecto
  Quiero asignar valor de negocio a cada historia y ver el backlog ordenado por WSJF
  Para decidir qué construir primero con criterio y no solo por intuición

  Antecedentes:
    Dado que la API está levantada y responde "200 OK" en "GET /health"
    Y que existe un proyecto con historias del backlog
    Y que el usuario es miembro del proyecto
    Y que los valores de negocio son estimaciones subjetivas del equipo

  # ===========================================================================
  # CÁLCULOS
  # ===========================================================================

  Regla: Cálculo de ROI, costo de demora y WSJF

    @caso_normal
    Escenario: E-01 - ROI de una historia con valor de negocio 8 y story points 5
      # CA-01
      Dado que existe una historia con "valor_negocio" 8 y "story_points" 5
      Cuando se consulta el backlog priorizado
      Entonces la historia tiene "roi" 1.6

    @caso_normal
    Escenario: E-02 - Costo de demora y WSJF con los tres valores cargados
      # CA-02
      Dado que existe una historia con "valor_negocio" 3, "criticidad_tiempo" 1,
        "reduccion_riesgo" 1 y "story_points" 3
      Cuando se consulta el backlog priorizado
      Entonces el costo de demora es 5
      Y la historia tiene "wsjf" 1.6666666666666667 sin redondear

    @caso_limite
    Escenario: E-03 - Historia sin story points o con story points cero
      # CA-03, CA-10
      Dado que existe una historia sin "story_points"
      Y que existe otra historia con "story_points" 0
      Cuando se consulta el backlog priorizado
      Entonces ambas historias tienen "roi" null y "wsjf" null
      Y aparecen al final del listado

    @caso_alternativo
    Escenario: E-04 - Campos de valor faltantes cuentan cero en el costo de demora
      # CA-04, CA-11
      Dado que existe una historia con "valor_negocio" null, "criticidad_tiempo" 3,
        "reduccion_riesgo" 2 y "story_points" 5
      Cuando se consulta el backlog priorizado
      Entonces el costo de demora es 5
      Y la historia tiene "wsjf" 1
      Y la historia tiene "roi" null

  # ===========================================================================
  # VALIDACIÓN
  # ===========================================================================

  Regla: Los valores de negocio pertenecen a la escala 1, 2, 3, 5, 8, 13, 21

    @caso_error
    Esquema del escenario: E-05 - Alta con un valor fuera de la escala
      # CA-05
      Dado que el usuario es miembro del proyecto
      Cuando envía un alta de historia con "<campo>" igual a "<valor>"
      Entonces la respuesta es 422 "VALIDACION"
      Y el campo de la respuesta es "<campo>"
      Y no se crea ninguna historia

      Ejemplos:
        | campo             | valor |
        | valor_negocio     | 0     |
        | valor_negocio     | 4     |
        | valor_negocio     | 22    |
        | criticidad_tiempo | -1    |
        | reduccion_riesgo  | 0.5   |

    @caso_error
    Esquema del escenario: E-06 - PATCH con un valor fuera de la escala
      # CA-05
      Dado que existe una historia con valores de negocio cargados
      Cuando el usuario envía un PATCH con "<campo>" igual a "<valor>"
      Entonces la respuesta es 422 "VALIDACION"
      Y el campo de la respuesta es "<campo>"
      Y la historia conserva sus valores anteriores

      Ejemplos:
        | campo             | valor |
        | valor_negocio     | 0.5   |
        | criticidad_tiempo | 4     |
        | reduccion_riesgo  | 22    |

  # ===========================================================================
  # ORDEN Y FILTRADO
  # ===========================================================================

  Regla: El backlog priorizado ordena por WSJF descendente y excluye lo hecho

    @caso_normal
    Escenario: E-07 - Orden por WSJF descendente con desempate por ID
      # CA-06
      Dado que existe una historia con "wsjf" 2.6
      Y que existe una historia con "wsjf" 1.6
      Y que existen dos historias con "wsjf" 0.6
      Cuando se consulta el backlog priorizado
      Entonces las historias aparecen en orden descendente de "wsjf"
      Y las dos historias con "wsjf" 0.6 aparecen por ID ascendente

    @caso_alternativo
    Escenario: E-08 - Las historias en estado Hecho no aparecen
      # CA-07
      Dado que existe una historia en estado "Hecho" con valores de negocio cargados
      Y que existe una historia en estado "Backlog" con valores de negocio cargados
      Cuando se consulta el backlog priorizado
      Entonces la respuesta incluye únicamente la historia en estado "Backlog"

  # ===========================================================================
  # SEGURIDAD Y CONTRATO
  # ===========================================================================

  Regla: Solo los miembros consultan y las consultas no modifican datos

    @caso_error
    Escenario: E-09 - Un no miembro no puede consultar el backlog priorizado
      # CA-08
      Dado que existe un proyecto con historias
      Y que el usuario no es miembro del proyecto
      Cuando consulta "GET /api/proyectos/{id}/backlog/priorizado"
      Entonces la respuesta es 404 "PROYECTO_NO_ENCONTRADO"

    @no_funcional
    Escenario: E-10 - Contrato completo y consultas de solo lectura
      # CA-09, CA-12
      Dado que existe una historia con los tres valores de negocio y story points
      Cuando se consulta el backlog priorizado dos veces
      Entonces ambas respuestas tienen las claves y tipos del contrato
      Y la respuesta incluye "valor_negocio", "criticidad_tiempo",
        "reduccion_riesgo", "roi" y "wsjf"
      Y ningún dato de la historia cambia entre las dos consultas
