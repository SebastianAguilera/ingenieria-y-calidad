# language: es
# =============================================================================
# US-07 - Gestión de criterios de aceptación
# Issue: #49
# Especificación SDD: docs/specs/US-07-gestion-de-criterios-de-aceptacion-spec.md
# Trazabilidad: Issue #49 -> SDD -> Criterios de Aceptación -> Escenarios BDD -> Tests Go -> Código
# =============================================================================

Funcionalidad: US-07 - Gestión de criterios de aceptación
  Como Product Builder
  Quiero definir criterios de aceptación para cada historia de usuario
  Para formalizar los requerimientos y verificar su cumplimiento

  Contexto:
    Dado que la API esta levantada y responde "200 OK" en "GET /health"
    Y que la base de datos fue migrada correctamente
    Y que existe un proyecto con una historia de usuario

  Regla: RN-01 a RN-05 - Alta de criterios con posición automática

    @caso_normal
    Escenario: BDD-01 - Crear el primer criterio de una historia
      Cuando envio "POST" a "/api/proyectos/1/historias/1/criterios" con {"descripcion": "El usuario ve el saldo"}
      Entonces la respuesta tiene el estado HTTP "201"
      Y el criterio creado registra "position" igual a 1
      Y el criterio creado registra "is_met" igual a false

    @caso_normal
    Escenario: BDD-02 - Crear un segundo criterio asigna la posición siguiente
      Dado que la historia ya tiene un criterio con "position" igual a 1
      Cuando envio "POST" a "/api/proyectos/1/historias/1/criterios" con {"descripcion": "El usuario puede transferir"}
      Entonces la respuesta tiene el estado HTTP "201"
      Y el criterio creado registra "position" igual a 2

  Regla: RN-01 y CA-F2 - Listado de criterios por historia

    @caso_normal
    Escenario: BDD-03 - Listar los criterios de una historia en orden de posición
      Dado que la historia tiene dos criterios registrados
      Cuando envio "GET" a "/api/proyectos/1/historias/1/criterios"
      Entonces la respuesta tiene el estado HTTP "200"
      Y el listado informa "total" igual a 2
      Y los criterios aparecen ordenados por "position"

    @caso_limite
    Escenario: BDD-04 - Historia sin criterios devuelve listado vacío
      Cuando envio "GET" a "/api/proyectos/1/historias/1/criterios"
      Entonces la respuesta tiene el estado HTTP "200"
      Y el listado informa "total" igual a 0

  Regla: RN-06 - Actualización parcial de criterios

    @caso_normal
    Escenario: BDD-05 - Marcar un criterio como cumplido
      Cuando envio "PATCH" a "/api/proyectos/1/historias/1/criterios/1" con {"is_met": true}
      Entonces la respuesta tiene el estado HTTP "200"
      Y el criterio actualizado registra "is_met" igual a true

    @caso_error
    Escenario: BDD-06 - Rechazar una actualización sin campos
      Cuando envio "PATCH" a "/api/proyectos/1/historias/1/criterios/1" con {}
      Entonces la respuesta tiene el estado HTTP "422"
      Y el cuerpo contiene el error "VALIDACION"

  Regla: CA-F4 - Eliminación de criterios

    @caso_normal
    Escenario: BDD-07 - Eliminar un criterio existente
      Cuando envio "DELETE" a "/api/proyectos/1/historias/1/criterios/1"
      Entonces la respuesta tiene el estado HTTP "204"
      Y el criterio deja de aparecer en el listado

  Regla: CA-R1 - Validación de la descripción

    @caso_limite
    Escenario: BDD-08 - Rechazar descripción compuesta solo por espacios
      Cuando envio "POST" a "/api/proyectos/1/historias/1/criterios" con {"descripcion": "   "}
      Entonces la respuesta tiene el estado HTTP "422"
      Y el cuerpo contiene el error "VALIDACION" para el campo "descripcion"

    @caso_error
    Escenario: BDD-09 - Rechazar descripción de más de 500 caracteres
      Cuando envio "POST" a "/api/proyectos/1/historias/1/criterios" con una descripción de 501 caracteres
      Entonces la respuesta tiene el estado HTTP "422"
      Y el cuerpo contiene el error "VALIDACION" para el campo "descripcion"

  Regla: CL-04 - Recursos inexistentes

    @caso_error
    Escenario: BDD-10 - Operar sobre una historia inexistente
      Cuando envio "POST" a "/api/proyectos/1/historias/99/criterios" con {"descripcion": "Algo"}
      Entonces la respuesta tiene el estado HTTP "404"
      Y el cuerpo contiene el error "HISTORIA_NO_ENCONTRADA"

    @caso_error
    Escenario: BDD-11 - Actualizar un criterio inexistente
      Cuando envio "PATCH" a "/api/proyectos/1/historias/1/criterios/99" con {"is_met": true}
      Entonces la respuesta tiene el estado HTTP "404"
      Y el cuerpo contiene el error "CRITERIO_NO_ENCONTRADO"
