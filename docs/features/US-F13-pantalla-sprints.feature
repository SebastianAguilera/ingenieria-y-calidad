# language: es
# =============================================================================
# US-F13 - Pantalla de gestion de sprints
# Escenarios BDD (Gherkin)
#
# Especificacion SDD : docs/specs/US-F13-pantalla-sprints-spec.md
# Issue de origen    : SebastianAguilera/ingenieria-y-calidad#72
# Endpoints          : US-03
#                      GET   /api/proyectos/{id}/sprints
#                      POST  /api/proyectos/{id}/sprints
#                      PATCH /api/proyectos/{id}/sprints/{sprintId}/estado
# Trazabilidad       : Issue #72 -> SDD -> criterios -> escenarios BDD ->
#                      tests Vitest/RTL -> codigo
# Rama               : feature/US-F13-pantalla-sprints
# =============================================================================
#
# NOTA: cada escenario se automatiza como test de Vitest + React Testing
# Library. Los tests de la API usan mocks de frontend/src/api/client.ts y los
# de pantalla usan mocks de frontend/src/api/sprints.ts.
# -----------------------------------------------------------------------------
# NARRATIVA (US-F13)
# -----------------------------------------------------------------------------
# Como miembro de un proyecto, quiero ver, crear y cambiar el estado de los
# sprints desde la interfaz, para planificar y cerrar ciclos sin usar la API.
# -----------------------------------------------------------------------------
# CRITERIOS DE ACEPTACION / SDD seccion "Criterios de aceptacion"
# -----------------------------------------------------------------------------
# [CA-1]  Estado vacio con accion de crear.                        -> E-01
# [CA-2]  Listado con identificador, goal, fechas y estado.        -> E-02
# [CA-3]  Filtro por estado.                                       -> E-03, E-04
# [CA-4]  Alta valida crea un sprint Planificado.                  -> E-05
# [CA-5]  Validacion de cliente sin llamar a la API.               -> E-06..E-10
# [CA-6]  El sprint creado aparece en la lista sin recargar.       -> E-05
# [CA-7]  Activar un sprint Planificado lo pasa a Activo.          -> E-11
# [CA-8]  Cerrar el sprint Activo lo pasa a Cerrado.               -> E-12
# [CA-9]  Activar deshabilitado si ya hay un sprint Activo.        -> E-13
# [CA-10] Sprint Cerrado en modo lectura sin acciones.             -> E-14
# [CA-11] 409 SPRINT_ACTIVO_EXISTENTE como respaldo.               -> E-15
# [CA-12] 404 de proyecto sin acceso.                              -> E-16
# [CA-13] Error de red al listar/guardar y reintento.              -> E-17, E-18
# [CA-14] Modal accesible con Escape y foco.                       -> E-19
# [CA-15] Paginación de 10 en 10.                                  -> E-20
# [CA-16] Orden por fecha de inicio ASC y reset al filtrar.        -> E-21
# [CA-17] Sprint Goal truncado y expandible.                       -> E-22
# [CA-18] Estados y acciones con color.                            -> E-23
# =============================================================================

Funcionalidad: US-F13 - Pantalla de gestión de sprints

  Como miembro de un proyecto
  Quiero ver, crear y cambiar el estado de los sprints
  Para planificar y cerrar ciclos sin usar la API.

  Antecedentes:
    Dado que "ana" es miembro del proyecto "Proyecto Alfa"
    Y que "ana" está en la pantalla "/proyectos/1/sprints"

  Regla: Listado y estados de la pantalla (RB-01)

    Escenario: E-01 - Proyecto sin sprints muestra el estado vacío
      Dado que el proyecto no tiene sprints
      Cuando "ana" entra a la pantalla Sprints
      Entonces ve el mensaje "Todavía no hay sprints"
      Y ve la acción "Nuevo sprint"

    Escenario: E-02 - Proyecto con sprints muestra sus datos
      Dado que el proyecto tiene los sprints "Sprint 1" y "Sprint 2"
      Cuando "ana" entra a la pantalla Sprints
      Entonces ve una tabla con "Identificador", "Sprint Goal",
        "Fecha de inicio", "Fecha de fin", "Estado" y "Acciones"
      Y ve cada sprint con sus datos y su estado

  Regla: Filtro por estado (RB-02)

    Escenario: E-03 - El filtro Cerrado muestra solo sprints cerrados
      Dado que el listado tiene sprints "Planificado" y "Cerrado"
      Cuando "ana" elige el filtro "Cerrado"
      Entonces solo ve los sprints en estado "Cerrado"

    Escenario: E-04 - Filtro sin coincidencias muestra un aviso
      Dado que el listado no tiene sprints "Activo"
      Cuando "ana" elige el filtro "Activo"
      Entonces ve el aviso "No hay sprints con ese estado"
      Y no ve la tabla de sprints

  Regla: Alta de sprint (RB-03, RB-04)

    Escenario: E-05 - Alta válida crea un sprint Planificado
      Dado que "ana" abre el modal de alta
      Cuando completa "Sprint 3", "Entregar el MVP",
        "2026-05-01" y "2026-05-15" y confirma
      Entonces la API recibe la alta con esos datos
      Y el nuevo sprint aparece en la lista como "Planificado"
      Y el modal se cierra

    Escenario: E-06 - Identificador vacío no llama a la API
      Cuando "ana" envía el formulario con el identificador vacío
      Entonces ve "El identificador es obligatorio"
      Y la API no recibe la alta

    Escenario: E-07 - Identificador mayor a 50 caracteres no llama a la API
      Cuando "ana" envía el formulario con 51 caracteres de identificador
      Entonces ve "El identificador no puede superar los 50 caracteres"
      Y la API no recibe la alta

    Escenario: E-08 - Sprint Goal vacío no llama a la API
      Cuando "ana" envía el formulario con el Sprint Goal vacío
      Entonces ve "El Sprint Goal es obligatorio"
      Y la API no recibe la alta

    Escenario: E-09 - Sprint Goal mayor a 300 caracteres no llama a la API
      Cuando "ana" envía el formulario con 301 caracteres de Sprint Goal
      Entonces ve "El Sprint Goal no puede superar los 300 caracteres"
      Y la API no recibe la alta

    Escenario: E-10 - Fecha de fin anterior al inicio no llama a la API
      Cuando "ana" envía una fecha de fin anterior a la de inicio
      Entonces ve "La fecha de fin debe ser igual o posterior a la de inicio"
      Y la API no recibe la alta

  Regla: Cambio de estado (RB-05, RB-06, RB-07)

    Escenario: E-11 - Activar un sprint Planificado lo pasa a Activo
      Dado que el listado tiene un sprint "Planificado" y ninguno "Activo"
      Cuando "ana" elige "Activar", confirma y la API responde "Activo"
      Entonces el sprint se muestra como "Activo"

    Escenario: E-12 - Cerrar el sprint Activo lo pasa a Cerrado
      Dado que el listado tiene un sprint "Activo"
      Cuando "ana" elige "Cerrar"
      Entonces ve que al cerrarlo no se puede volver a abrir
      Y al confirmar y responder la API "Cerrado"
      Entonces el sprint se muestra como "Cerrado"

    Escenario: E-13 - Activar queda deshabilitado si ya hay un Activo
      Dado que el listado tiene un sprint "Activo"
      Cuando "ana" ve las acciones de otro sprint "Planificado"
      Entonces la acción "Activar" está deshabilitada

    Escenario: E-14 - Un sprint Cerrado no tiene acciones
      Dado que el listado tiene un sprint "Cerrado"
      Cuando "ana" ve la fila de ese sprint
      Entonces no ve las acciones "Activar" ni "Cerrar"

  Regla: Errores y accesibilidad (RB-08..RB-11)

    Escenario: E-15 - Carrera al activar muestra el error de la API
      Dado que otro usuario activó un sprint
      Cuando "ana" confirma "Activar" y la API responde
        "409 SPRINT_ACTIVO_EXISTENTE"
      Entonces ve el mensaje de la API
      Y el sprint no cambia de estado

    Escenario: E-16 - Proyecto sin acceso muestra 404
      Dado que "ana" no es miembro del proyecto
      Cuando entra a la pantalla Sprints
      Entonces ve "No tenés acceso a este proyecto o no existe"

    Escenario: E-17 - Error de red al listar permite reintentar
      Dado que la API de listado falla
      Cuando "ana" entra a la pantalla Sprints
      Entonces ve el mensaje de error
      Y puede elegir "Reintentar" para volver a cargar

    Escenario: E-18 - Error de red al crear conserva el formulario
      Dado que la API de alta falla por red
      Cuando "ana" envía un formulario válido
      Entonces ve el mensaje de error en el modal
      Y sus datos siguen cargados

    Escenario: E-19 - El modal es accesible y se cierra con Escape
      Cuando "ana" abre el modal de alta
      Entonces el foco está en el campo "Identificador"
      Y todas las etiquetas son accesibles
      Cuando presiona "Escape"
      Entonces el modal se cierra

  Regla: Presentación del listado (RB-12, RB-13, RB-14)

    Escenario: E-20 - La tabla pagina de 10 en 10
      Dado que el proyecto tiene 12 sprints
      Cuando "ana" entra a la pantalla Sprints
      Entonces ve 10 sprints y "Página 1 de 2"
      Cuando elige "Siguiente"
      Entonces ve los 2 sprints restantes y "Página 2 de 2"

    Escenario: E-21 - Orden ascendente y reinicio de página al filtrar
      Dado que la API devuelve sprints con fechas desordenadas
      Cuando "ana" entra a la pantalla Sprints
      Entonces los ve ordenados por fecha de inicio ascendente
      Y si estaba en otra página y cambia el filtro
      Entonces vuelve a la página 1

    Escenario: E-22 - Sprint Goal largo truncado y expandible
      Dado que un sprint tiene un Sprint Goal de más de 60 caracteres
      Cuando "ana" lo ve en la tabla
      Entonces el texto aparece truncado con "…"
      Cuando elige "Ver más"
      Entonces ve el Sprint Goal completo
      Y puede elegir "Ver menos" para volver a truncarlo

    Escenario: E-23 - Estados y acciones con color
      Cuando "ana" ve la tabla de sprints
      Entonces cada estado se muestra como una etiqueta con color
      Y la acción "Activar" se muestra en verde
      Y la acción "Cerrar" se muestra en rojo
