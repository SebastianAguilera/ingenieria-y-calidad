# language: es
# =============================================================================
# US-18 - Defectos por Sprint y Edición de Defectos
# Issue: #92
# Especificación SDD: docs/specs/US-18-defectos-por-sprint-spec.md
# Trazabilidad: Issue #92 -> SDD -> Criterios de Aceptación -> Escenarios BDD -> Tests Go -> Código
# Total: 22 escenarios (E-01 a E-22).
# Guardas de regresión (pasan desde el primer RED): E-11, E-12, E-14, E-15, E-16,
# E-18, E-19, E-20. Ver docs/specs/US-18-defectos-por-sprint-spec.md.
# =============================================================================

Funcionalidad: US-18 - Defectos por Sprint y Edición de Defectos
  Como miembro de un proyecto
  Quiero ver los defectos de cada Sprint y corregir su descripción y severidad
  Para hacer seguimiento de la calidad por Sprint

  Contexto:
    Dado que la API esta levantada y responde "200 OK" en "GET /health"
    Y que existe un proyecto con un Sprint de detección y un Sprint de resolución
    Y que el usuario es miembro del proyecto

  # ===========================================================================
  # LISTADO: GET /api/proyectos/{id}/sprints/{sprintId}/defectos?tipo=
  # ===========================================================================

  Regla: Listado por Sprint de detección o de resolución

    @caso_normal
    Escenario: E-01 - Listado por detección devuelve solo los defectos detectados en el Sprint
      Dado que existe un defecto detectado en el Sprint 1
      Y que existe un defecto detectado en el Sprint 2
      Cuando envio "GET" a "/api/proyectos/1/sprints/1/defectos?tipo=deteccion"
      Entonces la respuesta tiene el estado HTTP "200"
      Y el listado contiene 1 defecto
      Y todos los defectos del listado tienen "sprint_deteccion_id" igual a "1"

    @caso_normal
    Escenario: E-02 - Listado por resolución devuelve solo los defectos resueltos en el Sprint
      Dado que existe un defecto detectado en el Sprint 1 y resuelto en el Sprint 2
      Y que existe un defecto detectado en el Sprint 1 y resuelto en el Sprint 1
      Cuando envio "GET" a "/api/proyectos/1/sprints/2/defectos?tipo=resolucion"
      Entonces la respuesta tiene el estado HTTP "200"
      Y el listado contiene 1 defecto
      Y todos los defectos del listado tienen "sprint_resolucion_id" igual a "2"

    @caso_limite
    Escenario: E-03 - Sprint sin defectos para el tipo pedido devuelve lista vacía
      Cuando envio "GET" a "/api/proyectos/1/sprints/1/defectos?tipo=resolucion"
      Entonces la respuesta tiene el estado HTTP "200"
      Y el cuerpo contiene "total" igual a "0"
      Y el cuerpo contiene "defectos" como arreglo vacío y no como "null"

    @caso_error
    Escenario: E-04 - Falta el parámetro tipo
      Cuando envio "GET" a "/api/proyectos/1/sprints/1/defectos"
      Entonces la respuesta tiene el estado HTTP "422"
      Y el cuerpo contiene el error "VALIDACION" para el campo "tipo"

    @caso_error
    Escenario: E-05 - Valor de tipo inválido
      Cuando envio "GET" a "/api/proyectos/1/sprints/1/defectos?tipo=Deteccion"
      Entonces la respuesta tiene el estado HTTP "422"
      Y el cuerpo contiene el error "VALIDACION" para el campo "tipo"

    @caso_error
    Escenario: E-06 - Sprint inexistente
      Cuando envio "GET" a "/api/proyectos/1/sprints/999/defectos?tipo=deteccion"
      Entonces la respuesta tiene el estado HTTP "404"
      Y el cuerpo contiene el error "SPRINT_NO_ENCONTRADA"

    @caso_error
    Escenario: E-07 - Sprint de otro proyecto se trata como inexistente
      Dado que existe un Sprint 5 en otro proyecto
      Cuando envio "GET" a "/api/proyectos/1/sprints/5/defectos?tipo=deteccion"
      Entonces la respuesta tiene el estado HTTP "404"
      Y el cuerpo contiene el error "SPRINT_NO_ENCONTRADA"

    @caso_error
    Escenario: E-08 - Un no miembro no puede listar los defectos del Sprint
      Dado que el usuario no es miembro del proyecto
      Cuando envio "GET" a "/api/proyectos/1/sprints/1/defectos?tipo=deteccion"
      Entonces la respuesta tiene el estado HTTP "404"
      Y el cuerpo contiene el error "PROYECTO_NO_ENCONTRADO"

  # ===========================================================================
  # EDICIÓN: PATCH /api/proyectos/{id}/defectos/{defectoId}
  # ===========================================================================

  Regla: Edición de descripción y severidad

    @caso_normal
    Escenario: E-09 - Editar la descripción con un valor válido
      Dado que existe un defecto con descripción "Error viejo"
      Cuando envio "PATCH" a "/api/proyectos/1/defectos/1" con {"descripcion": "Error corregido"}
      Entonces la respuesta tiene el estado HTTP "200"
      Y la respuesta contiene "descripcion" igual a "Error corregido"

    @caso_normal
    Escenario: E-10 - Editar la severidad con un valor de la lista
      Dado que existe un defecto con severidad "minor"
      Cuando envio "PATCH" a "/api/proyectos/1/defectos/1" con {"severidad": "critical"}
      Entonces la respuesta tiene el estado HTTP "200"
      Y la respuesta contiene "severidad" igual a "critical"

    @caso_limite
    Escenario: E-11 - Descripción de exactamente 1000 caracteres
      Cuando envio "PATCH" a "/api/proyectos/1/defectos/1" con una descripción de 1000 caracteres
      Entonces la respuesta tiene el estado HTTP "200"

    @caso_limite
    Escenario: E-12 - Descripción de 1000 caracteres multibyte se acepta
      Cuando envio "PATCH" a "/api/proyectos/1/defectos/1" con una descripción de 1000 caracteres "ñ"
      Entonces la respuesta tiene el estado HTTP "200"

    @caso_error
    Escenario: E-13 - Descripción de 1001 caracteres
      Cuando envio "PATCH" a "/api/proyectos/1/defectos/1" con una descripción de 1001 caracteres
      Entonces la respuesta tiene el estado HTTP "422"
      Y el cuerpo contiene el error "VALIDACION" para el campo "descripcion"

    @caso_error
    Escenario: E-14 - Descripción formada solo por espacios
      Cuando envio "PATCH" a "/api/proyectos/1/defectos/1" con {"descripcion": "   "}
      Entonces la respuesta tiene el estado HTTP "422"
      Y el cuerpo contiene el error "VALIDACION" para el campo "descripcion"

    @caso_error
    Escenario: E-15 - Severidad fuera de la lista
      Cuando envio "PATCH" a "/api/proyectos/1/defectos/1" con {"severidad": "Major"}
      Entonces la respuesta tiene el estado HTTP "422"
      Y el cuerpo contiene el error "VALIDACION" para el campo "severidad"

    @caso_limite
    Escenario: E-16 - Editar severidad no modifica estado ni Sprint de resolución
      Dado que existe un defecto "resolved" con sprint de resolución 2
      Cuando envio "PATCH" a "/api/proyectos/1/defectos/1" con {"severidad": "blocker"}
      Entonces la respuesta tiene el estado HTTP "200"
      Y el defecto conserva "estado" igual a "resolved"
      Y el defecto conserva "sprint_resolucion_id" igual a "2"

    @caso_error
    Escenario: E-17 - Enviar sprint_deteccion_id en el PATCH se rechaza
      Dado que existe un defecto detectado en el Sprint 1
      Cuando envio "PATCH" a "/api/proyectos/1/defectos/1" con {"sprint_deteccion_id": 2, "severidad": "major"}
      Entonces la respuesta tiene el estado HTTP "422"
      Y el cuerpo contiene el error "VALIDACION" para el campo "sprint_deteccion_id"
      Y el mensaje contiene "sprint_deteccion_id no es modificable"
      Y el defecto conserva "sprint_deteccion_id" igual a "1"
      Y el defecto conserva "severidad" sin cambios

    @caso_error
    Escenario: E-21 - Enviar sprint_deteccion_id con el mismo valor actual también se rechaza
      Dado que existe un defecto detectado en el Sprint 1
      Cuando envio "PATCH" a "/api/proyectos/1/defectos/1" con {"sprint_deteccion_id": 1}
      Entonces la respuesta tiene el estado HTTP "422"
      Y el cuerpo contiene el error "VALIDACION" para el campo "sprint_deteccion_id"

    @caso_error
    Escenario: E-22 - Enviar sprint_deteccion_id null también se rechaza
      Dado que existe un defecto detectado en el Sprint 1
      Cuando envio "PATCH" a "/api/proyectos/1/defectos/1" con {"sprint_deteccion_id": null}
      Entonces la respuesta tiene el estado HTTP "422"
      Y el cuerpo contiene el error "VALIDACION" para el campo "sprint_deteccion_id"

    @caso_error
    Escenario: E-18 - Defecto inexistente
      Cuando envio "PATCH" a "/api/proyectos/1/defectos/999" con {"severidad": "major"}
      Entonces la respuesta tiene el estado HTTP "404"
      Y el cuerpo contiene el error "DEFECTO_NO_ENCONTRADO"

    @caso_error
    Escenario: E-19 - Defecto de otro proyecto se trata como inexistente
      Dado que existe un defecto en otro proyecto con id 7
      Cuando envio "PATCH" a "/api/proyectos/1/defectos/7" con {"descripcion": "Intento"}
      Entonces la respuesta tiene el estado HTTP "404"
      Y el cuerpo contiene el error "DEFECTO_NO_ENCONTRADO"

    @caso_error
    Escenario: E-20 - Un no miembro no puede editar un defecto
      Dado que el usuario no es miembro del proyecto
      Cuando envio "PATCH" a "/api/proyectos/1/defectos/1" con {"severidad": "major"}
      Entonces la respuesta tiene el estado HTTP "404"
      Y el cuerpo contiene el error "PROYECTO_NO_ENCONTRADO"
