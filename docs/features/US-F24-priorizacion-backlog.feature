# language: es
# =============================================================================
# US-F24 - Valor de negocio y backlog priorizado
# Escenarios BDD (Gherkin)
#
# Especificación SDD : docs/specs/US-F24-priorizacion-backlog-spec.md
# Issue de origen    : SebastianAguilera/ingenieria-y-calidad#113
# Backend consumido  : US-23
#                      GET /api/proyectos/{id}/backlog/priorizado
#                      Respuesta: { "total": n, "historias": [...] }
# Dependencias       : US-23 (bloqueante), US-F07, US-F10, US-11
# Trazabilidad       : Issue #113 -> SDD -> criterios -> escenarios BDD ->
#                      tests Vitest/RTL -> código
# Rama               : feature/US-F24
# =============================================================================
#
# NOTA: cada escenario se automatiza como test de Vitest + React Testing
# Library. Los tests de API usan mocks de frontend/src/api/client.ts y los de
# pantalla usan mocks de frontend/src/api/historias.ts.
# -----------------------------------------------------------------------------
# NARRATIVA (US-F24)
# -----------------------------------------------------------------------------
# Como líder de un proyecto
# Quiero cargar el valor de negocio de cada historia y ver el backlog ordenado
# por WSJF
# Para decidir qué construir primero con criterio.
# -----------------------------------------------------------------------------
# CRITERIOS DE ACEPTACIÓN
# [CA-01] Campos de priorización con ayuda visible.           -> E-01
# [CA-02] Escala Fibonacci cerrada en los selectores.         -> E-02
# [CA-03] Guardar valores y verlos al editar.                 -> E-03, E-04
# [CA-04] Vista priorizada en el orden del backend.           -> E-05
# [CA-05] Historia sin story points muestra "—" en ROI/WSJF.  -> E-06
# [CA-06] Historias en Hecho no aparecen.                     -> E-07
# [CA-07] 422 en el campo correspondiente.                    -> E-08
# [CA-08] Error de red y reintento.                           -> E-09
# [CA-09] Estado de carga.                                    -> E-10
# [CA-10] Estado vacío sin historias pendientes.              -> E-10
# =============================================================================

Característica: US-F24 - Valor de negocio y backlog priorizado
  Como líder de un proyecto
  Quiero cargar el valor de negocio de cada historia y ver el backlog ordenado
  por WSJF
  Para decidir qué construir primero con criterio.

  Antecedentes:
    Dado que el usuario está autenticado
    Y que es miembro del proyecto
    Y que US-23 expone "GET /api/proyectos/{id}/backlog/priorizado"

  # ===========================================================================
  # MODAL DE HISTORIA
  # ===========================================================================

  Regla: Carga de los valores de priorización

    @caso_normal
    Escenario: E-01 - Campos de priorización con ayuda visible
      # CA-01
      Dado que el usuario abre el modal de una historia
      Cuando se muestra el formulario
      Entonces ve los selectores "Valor de negocio", "Criticidad en el tiempo"
        y "Reducción de riesgo"
      Y cada selector tiene una ayuda asociada con "aria-describedby"

    @limite
    Escenario: E-02 - El selector solo ofrece la escala Fibonacci
      # CA-02
      Dado que el usuario abre un selector de priorización
      Cuando despliega sus opciones
      Entonces las únicas opciones son "Sin asignar", 1, 2, 3, 5, 8, 13 y 21
      Y no existe un valor "0", "0.5" ni texto libre

    @caso_normal
    Escenario: E-03 - Guardar los tres valores y verlos al editar
      # CA-03
      Dado que el usuario completa valor de negocio 8, criticidad 3 y riesgo 2
      Cuando guarda la historia
      Entonces la interfaz envía "valor_negocio": 8, "criticidad_tiempo": 3
        y "reduccion_riesgo": 2
      Y al volver a abrir el modal los tres selectores muestran esos valores

    @caso_normal
    Escenario: E-04 - Valores existentes cargados en el modal
      # CA-03
      Dado que una historia ya tiene valor de negocio 5, criticidad 1 y riesgo 1
      Cuando el usuario abre el modal de edición
      Entonces los selectores muestran 5, 1 y 1
      Y la opción "Sin asignar" está deshabilitada en esos campos

  # ===========================================================================
  # VISTA PRIORIZADA
  # ===========================================================================

  Regla: Consulta del backlog priorizado

    @caso_normal
    Escenario: E-05 - Vista priorizada en el orden de WSJF del backend
      # CA-04
      Dado que el backend devuelve las historias ordenadas por WSJF descendente
      Cuando el usuario elige la vista "Priorizado"
      Entonces se consulta "GET /api/proyectos/{id}/backlog/priorizado"
      Y la tabla muestra las filas en el mismo orden recibido
      Y cada fila muestra Código, Título, Estado, Prioridad, SP, Valor,
        Criticidad, Riesgo, ROI y WSJF

    @limite
    Escenario: E-06 - Historia sin story points muestra "—" en ROI y WSJF
      # CA-05
      Dado que una historia tiene "story_points": null, "roi": null y "wsjf": null
      Cuando aparece en la vista priorizada
      Entonces sus columnas ROI y WSJF muestran "—"
      Y el resto de sus datos se muestran normalmente

    @alternativo
    Escenario: E-07 - Historia en Hecho no aparece
      # CA-06
      Dado que una historia tiene estado "Hecho"
      Cuando el backend responde el backlog priorizado
      Entonces esa historia no forma parte de la respuesta
      Y la vista no la muestra

  # ===========================================================================
  # ERRORES Y ESTADOS
  # ===========================================================================

  Regla: Manejo de errores y estados de pantalla

    @error
    Escenario: E-08 - 422 en un campo de priorización
      # CA-07
      Dado que el backend rechaza "criticidad_tiempo" con 422 "VALIDACION"
      Cuando el usuario guarda el modal
      Entonces el mensaje de error aparece junto al selector "Criticidad en el tiempo"
      Y no se cierra el modal

    @error
    Escenario: E-09 - Error de red al cargar la vista y reintento
      # CA-08
      Dado que la consulta del backlog priorizado falla por red
      Cuando el usuario abre la vista "Priorizado"
      Entonces ve un mensaje de error y el botón "Reintentar"
      Cuando presiona "Reintentar"
      Entonces se vuelve a llamar al endpoint

    @no_funcional
    Escenario: E-10 - Estados de carga y vacío
      # CA-09, CA-10
      Dado que la consulta del backlog priorizado está en curso
      Entonces la vista muestra "Cargando backlog priorizado…"
      Cuando el backend devuelve "total": 0 e "historias": []
      Entonces la vista muestra "No hay historias pendientes de priorizar"
      Y no muestra un error
