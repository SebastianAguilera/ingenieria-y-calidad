# language: es
# =============================================================================
# US-F14 - Detalle de sprint y Sprint Backlog
# Escenarios BDD (Gherkin)
#
# Especificacion SDD : docs/specs/US-F14-detalle-sprint-backlog-spec.md
# Issue de origen    : SebastianAguilera/ingenieria-y-calidad#73
# Endpoints          : US-13 GET    /api/proyectos/{id}/sprints/{sprintId}/historias
#                      US-04 POST   /api/proyectos/{id}/sprints/{sprintId}/historias
#                      US-04 DELETE /api/proyectos/{id}/sprints/{sprintId}/historias/{historiaId}
#                      US-02 GET    /api/proyectos/{id}/historias (candidatas)
# Trazabilidad       : Issue #73 -> SDD -> criterios -> escenarios BDD ->
#                      tests Vitest/RTL -> codigo
# Rama               : feature/US-F14-detalle-sprint-backlog
# =============================================================================
#
# NOTA: cada escenario se automatiza como test de Vitest + React Testing
# Library. Los tests de la API usan mocks de frontend/src/api/client.ts y los
# de pantalla usan mocks de frontend/src/api/sprints.ts y historias.ts.
# El GET de US-13 todavia no existe en el backend: se prueba contra su contrato.
# -----------------------------------------------------------------------------
# NARRATIVA (US-F14)
# -----------------------------------------------------------------------------
# Como miembro de un proyecto, quiero abrir un sprint y ver las historias
# asignadas, para hacer seguimiento del trabajo y asignar o quitar historias.
# -----------------------------------------------------------------------------
# CRITERIOS DE ACEPTACION / SDD seccion "Criterios de aceptacion"
# -----------------------------------------------------------------------------
# [CA-1]  Cabecera del sprint y sus historias.                     -> E-01, E-02
# [CA-2]  Sprint vacio con accion de asignar si esta Activo.       -> E-03, E-04
# [CA-3]  Asignar una historia en Backlog.                         -> E-05, E-06
# [CA-4]  Historia ya asignada o no pendiente: error, sin duplicar -> E-07, E-08
# [CA-5]  Sprint Planificado: aviso y sin asignar.                 -> E-09
# [CA-6]  Sprint Cerrado en modo lectura.                          -> E-10
# [CA-7]  Quitar con confirmacion.                                 -> E-11, E-12
# [CA-8]  Sprint cerrado: sin accion de quitar.                    -> E-10
# [CA-9]  Asignacion fuera de la ventana temporal.                 -> E-13
# [CA-10] Sprint inexistente o de otro proyecto: 404.              -> E-14
# [CA-11] Error de red y reintento.                                -> E-15, E-16
# [CA-12] Refactor de tabla sin romper US-F07.                     -> E-17
# [CA-13] Modales accesibles.                                      -> E-18
# [CA-14] Sin edicion de goal/fechas ni cierre.                    -> E-19
# =============================================================================

Funcionalidad: US-F14 - Detalle de sprint y Sprint Backlog

  Como miembro de un proyecto
  Quiero abrir un sprint y ver las historias asignadas
  Para hacer seguimiento del trabajo y asignar o quitar historias

  Antecedentes:
    Dado que soy miembro del proyecto 1
    Y existe el sprint "Sprint 1" con goal "Entregar el MVP" del 2026-10-06 al 2026-10-20

  # --- Casos normales ---------------------------------------------------------
  Escenario: E-01 Ver la cabecera del sprint
    Dado que el sprint está "Activo"
    Cuando entro a "/proyectos/1/sprints/1"
    Entonces veo el identificador "Sprint 1", el goal "Entregar el MVP",
      las fechas "2026-10-06" y "2026-10-20" y el estado "Activo"

  Escenario: E-02 Ver las historias asignadas al sprint
    Dado que el sprint tiene asignadas "US-01 Crear proyecto" (3 story points) y "US-02 Ver proyectos"
    Cuando entro a "/proyectos/1/sprints/1"
    Entonces veo una tabla con Código, Título, Estado y Story Points
    Y las historias aparecen en el orden devuelto por la API

  Escenario: E-03 Sprint activo sin historias
    Dado que el sprint está "Activo" y no tiene historias
    Cuando entro a "/proyectos/1/sprints/1"
    Entonces veo el estado vacío
    Y veo la acción "Asignar historia"

  Escenario: E-04 Sprint sin historias que no está activo
    Dado que el sprint está "Cerrado" y no tiene historias
    Cuando entro a "/proyectos/1/sprints/1"
    Entonces veo el estado vacío
    Y no veo la acción "Asignar historia"

  Escenario: E-05 Asignar una historia en Backlog
    Dado que el sprint está "Activo"
    Y el proyecto tiene la historia "US-05 Exportar informe" en estado "Backlog"
    Cuando abro "Asignar historia", elijo "US-05" y confirmo
    Entonces se envía POST con historia_id de "US-05"
    Y "US-05" aparece en la tabla sin recargar la página

  Escenario: E-06 La historia asignada deja de estar disponible
    Dado que asigné "US-05" al sprint
    Cuando vuelvo a abrir "Asignar historia"
    Entonces "US-05" ya no figura entre las historias disponibles

  Escenario: E-11 Quitar una historia con confirmación
    Dado que el sprint está "Activo" y tiene asignada "US-01"
    Cuando elijo "Quitar del sprint" en "US-01" y confirmo
    Entonces se envía DELETE de "US-01"
    Y "US-01" desaparece de la tabla

  # --- Casos alternativos -----------------------------------------------------
  Escenario: E-12 Cancelar la confirmación de quitar
    Dado que el sprint tiene asignada "US-01"
    Cuando elijo "Quitar del sprint" y cancelo
    Entonces no se envía ninguna petición
    Y "US-01" sigue en la tabla

  Escenario: E-09 Sprint planificado no permite asignar
    Dado que el sprint está "Planificado"
    Cuando entro a "/proyectos/1/sprints/1"
    Entonces veo el aviso "Activá el sprint para poder asignar historias"
    Y no veo la acción "Asignar historia"

  Escenario: E-10 Sprint cerrado en modo lectura
    Dado que el sprint está "Cerrado" y tiene asignada "US-01"
    Cuando entro a "/proyectos/1/sprints/1"
    Entonces veo "US-01" en la tabla
    Y no veo "Asignar historia" ni "Quitar del sprint"

  Escenario: E-20 No hay historias disponibles para asignar
    Dado que el sprint está "Activo" y todas las historias en "Backlog" ya están en el sprint
    Cuando abro "Asignar historia"
    Entonces veo "No hay historias disponibles para asignar"

  # --- Errores ----------------------------------------------------------------
  Escenario: E-07 Historia ya asignada a otro sprint
    Dado que el sprint está "Activo"
    Y el backend responde 409 HISTORIA_YA_ASIGNADA al asignar "US-05"
    Cuando asigno "US-05"
    Entonces veo el mensaje del backend
    Y la tabla no cambia

  Escenario: E-08 Historia que no está pendiente
    Dado que el backend responde 400 HISTORIA_NO_PENDIENTE al asignar "US-06"
    Cuando asigno "US-06"
    Entonces veo el mensaje del backend
    Y la historia no se agrega

  Escenario: E-13 Asignación fuera de la ventana temporal
    Dado que el backend responde 400 FUERA_DE_VENTANA_TEMPORAL al asignar "US-05"
    Cuando asigno "US-05"
    Entonces veo el mensaje del backend
    Y la historia no se agrega a la tabla

  Escenario: E-14 Sprint inexistente o de otro proyecto
    Dado que el backend responde 404 SPRINT_NO_ENCONTRADA
    Cuando entro a "/proyectos/1/sprints/99"
    Entonces veo el estado "sprint no encontrado"
    Y un enlace para volver a la lista de sprints

  Escenario: E-15 Error de red al listar
    Dado que la red falla al consultar el sprint
    Cuando entro a "/proyectos/1/sprints/1"
    Entonces veo un mensaje de error
    Y al pulsar "Reintentar" se vuelve a consultar

  Escenario: E-16 Error de red al asignar o quitar
    Dado que la red falla al enviar la operación
    Cuando asigno o quito una historia
    Entonces veo un mensaje de error
    Y puedo reintentar sin que la tabla cambie

  # --- Refactor y accesibilidad -----------------------------------------------
  Escenario: E-17 El refactor no rompe el Product Backlog
    Dado que extraje TablaHistorias como componente compartido
    Cuando se ejecutan los tests de US-F07
    Entonces siguen en verde

  Escenario: E-18 Modales accesibles
    Dado que abro el modal "Asignar historia"
    Entonces el foco queda dentro del modal y tiene etiquetas accesibles
    Y al pulsar Escape se cierra

  Escenario: E-19 La pantalla no edita ni cierra el sprint
    Cuando entro al detalle de un sprint
    Entonces no hay acciones para editar goal o fechas, ni para crear o cerrar el sprint

  Escenario: E-21 Navegar desde la lista de sprints
    Dado que estoy en "/proyectos/1/sprints"
    Cuando pulso el identificador "Sprint 1"
    Entonces voy a "/proyectos/1/sprints/1"

# -----------------------------------------------------------------------------
# AUTOMATIZACION (futuro): frontend/src/api/sprints.test.ts,
# pages/proyecto/sprints/SprintDetallePage.test.tsx,
# TablaSprintBacklog.test.tsx, ModalAsignarHistoria.test.tsx y los tests de
# US-F07 (TableroBacklog.test.tsx, ModalHistoria.test.tsx, BacklogPage.test.tsx).
# -----------------------------------------------------------------------------
