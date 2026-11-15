# language: es
# =============================================================================
# US-F20 - Pantalla de defectos del proyecto
# Escenarios BDD (Gherkin)
#
# Especificacion SDD : docs/specs/US-F20-pantalla-defectos-spec.md
# Issue de origen    : SebastianAguilera/ingenieria-y-calidad#93
# Endpoints          : US-09
#                      GET   /api/proyectos/{id}/defectos
#                      POST  /api/proyectos/{id}/defectos
#                      PATCH /api/proyectos/{id}/defectos/{defectoId}
#                      US-18 (tramo 2)
#                      GET   /api/proyectos/{id}/sprints/{sprintId}/defectos?tipo=
#                      PATCH /api/proyectos/{id}/defectos/{defectoId}
#                            (descripcion y severidad)
# Trazabilidad       : Issue #93 -> SDD -> criterios -> escenarios BDD ->
#                      tests Vitest/RTL -> codigo
# Rama               : feature/US-F20-pantalla-defectos
# =============================================================================
#
# NOTA: cada escenario se automatiza como test de Vitest + React Testing
# Library. Los tests de la API usan mocks de frontend/src/api/client.ts y los
# de pantalla usan mocks de frontend/src/api/defectos.ts, historias.ts y
# sprints.ts.
# -----------------------------------------------------------------------------
# NARRATIVA (US-F20)
# -----------------------------------------------------------------------------
# Como miembro de un proyecto, quiero registrar, listar y cambiar el estado de
# los defectos desde la interfaz, para seguir la calidad del proyecto sin usar
# la API.
# -----------------------------------------------------------------------------
# CRITERIOS DE ACEPTACION / SDD seccion "Criterios de aceptacion"
# -----------------------------------------------------------------------------
# TRAMO 1
# [CA-1]  Listado con DEF-{id}, descripcion, US-xx, severidad, estado y
#         sprint de deteccion.                                    -> E-01..E-03
# [CA-2]  Historia ausente se muestra como "—".                   -> E-04
# [CA-3]  Estado vacio con accion de crear.                       -> E-05
# [CA-4]  Alta valida aparece en la lista sin recargar.           -> E-06
# [CA-5]  Validacion de cliente sin llamar a la API.              -> E-07..E-10
# [CA-6]  Resolver/cerrar sin sprint de resolucion: 422 en campo. -> E-11
# [CA-7]  Cambio de estado valido actualiza la lista.             -> E-12..E-15
# [CA-8]  Error de red al listar/guardar/cambiar y reintento.     -> E-16..E-19
# [CA-9]  404 de proyecto sin acceso.                             -> E-20
# [CA-10] Modales accesibles con foco y Escape.                   -> E-21, E-22
# TRAMO 2 (requiere US-18 mergeada)
# [CA-11] Filtro por sprint de deteccion.                         -> E-23
# [CA-12] Filtro por sprint de resolucion.                        -> E-24..E-26, E-37
# [CA-13] Edicion valida de descripcion o severidad.              -> E-27..E-29, E-35
# [CA-14] Descripcion de mas de 1000 caracteres.                  -> E-30..E-34
# [CA-11] (tambien)                                               -> E-36
# =============================================================================

Funcionalidad: US-F20 - Pantalla de defectos del proyecto

  Como miembro de un proyecto
  Quiero registrar, listar y cambiar el estado de los defectos
  Para seguir la calidad del proyecto sin usar la API.

  Antecedentes:
    Dado que "ana" es miembro del proyecto "Proyecto Alfa"
    Y que "ana" está en la pantalla "/proyectos/1/defectos"

  # ===========================================================================
  # TRAMO 1
  # ===========================================================================

  Regla: Listado y estados de la pantalla (RB-01, RB-02, RB-03)

    Escenario: E-01 - Proyecto con defectos muestra sus datos
      Dado que el proyecto tiene los defectos 7 y 9
      Cuando "ana" entra a la pantalla Defectos
      Entonces ve una tabla con "Código", "Descripción", "Historia",
        "Severidad", "Estado", "Sprint de detección", "Sprint de resolución"
        y "Acciones"
      Y ve los códigos "DEF-7" y "DEF-9"
      Y ve la historia como "US-01" y el sprint como "Sprint 1"

    Escenario: E-02 - La descripción larga se trunca a 80 caracteres
      Dado que un defecto tiene una descripción de 81 caracteres
      Cuando "ana" entra a la pantalla Defectos
      Entonces ve los primeros 80 caracteres seguidos de "…"
      Y el texto completo está disponible como título de la celda

    Escenario: E-03 - Una descripción de exactamente 80 caracteres no se trunca
      Dado que un defecto tiene una descripción de 80 caracteres
      Cuando "ana" entra a la pantalla Defectos
      Entonces ve la descripción completa sin "…"

    Escenario: E-04 - Historia fuera de la lista se muestra como guion
      Dado que un defecto referencia la historia 99
      Y que la historia 99 no figura en el listado de historias
      Cuando "ana" entra a la pantalla Defectos
      Entonces la columna Historia de ese defecto muestra "—"
      Y la pantalla no falla

    Escenario: E-05 - Proyecto sin defectos muestra el estado vacío
      Dado que el proyecto no tiene defectos
      Y que la API responde el listado sin el campo "defectos"
      Cuando "ana" entra a la pantalla Defectos
      Entonces ve el mensaje "Todavía no hay defectos"
      Y ve la acción "Nuevo defecto"

  Regla: Alta de defecto (RB-04, RB-05, RB-06)

    Escenario: E-06 - Un defecto válido aparece en la lista sin recargar
      Dado que el proyecto tiene la historia "US-01" y el sprint "Sprint 1"
      Cuando "ana" abre "Nuevo defecto"
      Y completa la descripción "Error al guardar"
      Y elige la severidad "Mayor", la historia "US-01" y el sprint "Sprint 1"
      Y guarda el formulario
      Entonces se llama a la API de alta con "descripcion", "severidad",
        "historia_id" y "sprint_deteccion_id"
      Y la llamada no incluye "reportado_por"
      Y el modal se cierra
      Y el defecto nuevo aparece en la tabla con estado "Abierto"

    Escenario: E-07 - Descripción vacía no llama a la API
      Cuando "ana" abre "Nuevo defecto"
      Y deja la descripción vacía o con solo espacios
      Y guarda el formulario
      Entonces ve el error "La descripción es obligatoria" junto al campo
      Y no se llama a la API

    Esquema del escenario: E-08 - Campos obligatorios sin elegir no llaman a la API
      Cuando "ana" abre "Nuevo defecto"
      Y completa una descripción válida
      Y no elige "<campo>"
      Y guarda el formulario
      Entonces ve el error "<mensaje>" junto al campo
      Y no se llama a la API

      Ejemplos:
        | campo               | mensaje                              |
        | severidad           | Elegí una severidad                  |
        | historia            | Elegí una historia                   |
        | sprint de detección | Elegí el sprint de detección         |

    Escenario: E-09 - Un 422 del backend en el alta se muestra junto al campo
      Dado que la API responde "422 VALIDACION" con el campo "descripcion"
      Cuando "ana" guarda un alta con datos aceptados por el cliente
      Entonces ve el mensaje del backend junto al campo "descripcion"
      Y el modal sigue abierto
      Y la tabla no cambia

    Escenario: E-10 - Un proyecto sin historias o sin sprints avisa qué falta
      Dado que el proyecto no tiene historias
      Cuando "ana" abre "Nuevo defecto"
      Entonces ve un aviso que indica que hay que crear una historia primero
      Y el botón de guardar está deshabilitado

  Regla: Cambio de estado y sprint de resolución (RB-07, RB-08, RB-09, RB-10)

    Escenario: E-11 - Resolver sin sprint de resolución muestra el 422 del backend
      Dado que existe un defecto "Abierto"
      Y que la API responde "422 VALIDACION" con el campo "sprint_resolucion_id"
      Cuando "ana" abre "Cambiar estado" de ese defecto
      Y elige el estado "Resuelto" sin elegir un sprint de resolución
      Y confirma
      Entonces se envía el PATCH con "estado" "resolved" y sin sprint de resolución
      Y ve el mensaje del backend junto al campo "sprint_resolucion_id"
      Y el defecto sigue "Abierto" en la tabla

    Escenario: E-12 - Resolver con sprint de resolución actualiza la lista
      Dado que existe un defecto "Abierto"
      Cuando "ana" cambia su estado a "Resuelto" eligiendo el sprint "Sprint 2"
      Y confirma
      Entonces se envía el PATCH con "estado" "resolved" y "sprint_resolucion_id" 2
      Y el modal se cierra
      Y la tabla muestra el estado "Resuelto" y el sprint de resolución "Sprint 2"

    Escenario: E-13 - Cerrar con sprint de resolución actualiza la lista
      Dado que existe un defecto "En progreso"
      Cuando "ana" cambia su estado a "Cerrado" eligiendo el sprint "Sprint 2"
      Y confirma
      Entonces la tabla muestra el estado "Cerrado"

    Escenario: E-14 - Reabrir un defecto no envía sprint de resolución
      Dado que existe un defecto "Resuelto" con sprint de resolución "Sprint 2"
      Cuando "ana" cambia su estado a "Abierto"
      Y confirma
      Entonces se envía el PATCH solo con "estado" "open"
      Y la tabla muestra el estado "Abierto" y el sprint de resolución "—"

    Escenario: E-15 - Un estado rechazado por el backend se muestra en el modal
      Dado que la API responde "400 ESTADO_INVALIDO" con el campo "estado"
      Cuando "ana" confirma un cambio de estado
      Entonces ve el mensaje del backend en el modal
      Y la tabla no cambia

  Regla: Errores de red y reintento (RB-12)

    Escenario: E-16 - Error de red al listar permite reintentar
      Dado que la API de defectos falla por red
      Cuando "ana" entra a la pantalla Defectos
      Entonces ve el mensaje "No se pudieron cargar los defectos. Intentá nuevamente."
      Y ve la acción "Reintentar"
      Cuando la API vuelve a responder y "ana" presiona "Reintentar"
      Entonces ve el listado de defectos

    Escenario: E-17 - Error al cargar historias o sprints muestra el error con reintento
      Dado que la API de historias falla por red
      Cuando "ana" entra a la pantalla Defectos
      Entonces ve el mensaje de error con la acción "Reintentar"
      Y no ve la tabla de defectos

    Escenario: E-18 - Error de red al guardar el alta mantiene el modal
      Dado que la API de alta falla por red
      Cuando "ana" guarda un alta válida
      Entonces ve el mensaje "No se pudo crear el defecto. Intentá nuevamente." en el modal
      Y el modal sigue abierto con los datos cargados
      Y puede volver a guardar

    Escenario: E-19 - Error de red al cambiar el estado mantiene el modal
      Dado que la API de actualización falla por red
      Cuando "ana" confirma un cambio de estado
      Entonces ve el mensaje "No se pudo cambiar el estado del defecto. Intentá nuevamente." en el modal
      Y la tabla no cambia

  Regla: Acceso al proyecto (RB-11)

    Escenario: E-20 - Un no miembro ve el mensaje de acceso
      Dado que la API responde "404 PROYECTO_NO_ENCONTRADO"
      Cuando "ana" entra a la pantalla Defectos
      Entonces ve "No tenés acceso a este proyecto o no existe"
      Y no ve la tabla de defectos

  Regla: Accesibilidad de los modales (RB-13)

    Escenario: E-21 - El modal de alta tiene foco, etiquetas y cierra con Escape
      Cuando "ana" abre "Nuevo defecto"
      Entonces el diálogo tiene "aria-modal" y un título accesible
      Y el primer campo recibe el foco
      Y cada campo tiene su etiqueta asociada
      Cuando "ana" presiona "Escape"
      Entonces el modal se cierra

    Escenario: E-22 - Un campo con error queda marcado como inválido
      Cuando "ana" guarda el alta con la descripción vacía
      Entonces el campo "Descripción" tiene "aria-invalid"
      Y el mensaje de error está asociado con "aria-describedby"

  # ===========================================================================
  # TRAMO 2 (requiere US-18 mergeada)
  # ===========================================================================

  Regla: Filtro por sprint (RB-14)

    Escenario: E-23 - Filtrar por sprint de detección
      Dado que el proyecto tiene defectos detectados en distintos sprints
      Cuando "ana" elige el sprint "Sprint 1" y el tipo "Detección"
      Entonces se consulta el listado del sprint 1 con tipo "deteccion"
      Y la tabla muestra solo los defectos detectados en "Sprint 1"

    Escenario: E-24 - Filtrar por sprint de resolución
      Dado que el proyecto tiene defectos resueltos en distintos sprints
      Cuando "ana" elige el sprint "Sprint 2" y el tipo "Resolución"
      Entonces se consulta el listado del sprint 2 con tipo "resolucion"
      Y la tabla muestra solo los defectos resueltos en "Sprint 2"

    Escenario: E-25 - Un filtro sin resultados muestra un aviso
      Dado que el sprint elegido no tiene defectos del tipo elegido
      Cuando "ana" aplica el filtro
      Entonces ve el aviso "No hay defectos con ese filtro"
      Y no ve el estado vacío de alta

    Escenario: E-26 - Quitar el filtro vuelve al listado completo
      Dado que "ana" tiene un filtro aplicado
      Cuando elige "Todos los sprints"
      Entonces se consulta el listado completo del proyecto
      Y la tabla muestra todos los defectos

  Regla: El filtro activo se respeta tras un alta o un cambio de estado (RB-14)

    Escenario: E-36 - Un defecto nuevo de otro sprint no aparece con el filtro por detección
      Dado que "ana" filtra por el sprint "Sprint 1" y el tipo "Detección"
      Cuando registra un defecto detectado en "Sprint 2"
      Entonces la tabla sigue mostrando solo los defectos detectados en "Sprint 1"

    Escenario: E-37 - Reabrir un defecto lo quita del filtro por resolución
      Dado que "ana" filtra por el sprint "Sprint 2" y el tipo "Resolución"
      Cuando reabre el único defecto resuelto en "Sprint 2"
      Entonces ve el aviso "No hay defectos con ese filtro"

  Regla: Edición de descripción y severidad (RB-15, RB-16)

    Escenario: E-27 - Editar la descripción actualiza la lista
      Dado que existe un defecto con descripción "Error al guardar"
      Cuando "ana" abre "Editar" de ese defecto
      Y cambia la descripción a "Error al guardar el sprint"
      Y guarda
      Entonces se envía el PATCH solo con "descripcion"
      Y la tabla muestra la nueva descripción

    Escenario: E-28 - Editar la severidad actualiza la lista
      Dado que existe un defecto con severidad "Menor"
      Cuando "ana" cambia la severidad a "Crítica" y guarda
      Entonces se envía el PATCH solo con "severidad" "critical"
      Y la tabla muestra la severidad "Crítica"

    Escenario: E-29 - La edición nunca envía el sprint de detección ni el estado
      Cuando "ana" guarda una edición válida
      Entonces el cuerpo del PATCH no contiene "sprint_deteccion_id"
      Y no contiene "estado" ni "sprint_resolucion_id"

  Regla: Límite de 1000 caracteres (RB-15, RB-16)

    Escenario: E-30 - Más de 1000 caracteres no llama a la API
      Cuando "ana" edita la descripción con 1001 caracteres
      Y guarda
      Entonces ve el error "La descripción no puede superar los 1000 caracteres" junto al campo
      Y no se llama a la API

    Escenario: E-31 - Exactamente 1000 caracteres es válido
      Cuando "ana" edita la descripción con 1000 caracteres
      Y guarda
      Entonces se llama a la API

    Escenario: E-32 - El límite cuenta caracteres y no bytes
      Cuando "ana" edita la descripción con 1000 veces "ñ"
      Y guarda
      Entonces se llama a la API

    Escenario: E-33 - Descripción vacía o solo con espacios en la edición
      Cuando "ana" edita la descripción dejándola vacía o con solo espacios
      Y guarda
      Entonces ve el error "La descripción es obligatoria" junto al campo
      Y no se llama a la API

    Escenario: E-34 - Un 422 del backend en la edición se muestra junto al campo
      Dado que la API responde "422 VALIDACION" con el campo "descripcion"
      Cuando "ana" guarda una edición aceptada por el cliente
      Entonces ve el mensaje del backend junto al campo "descripcion"
      Y la tabla no cambia

    Escenario: E-35 - Guardar sin cambios cierra el modal sin llamar a la API
      Dado que "ana" abre "Editar" de un defecto
      Cuando guarda sin modificar la descripción ni la severidad
      Entonces el modal se cierra
      Y no se llama a la API
