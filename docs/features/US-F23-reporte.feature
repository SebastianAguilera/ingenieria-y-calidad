# language: es
# =============================================================================
# US-F23 - Vista del reporte de proyecto y sprint
# Escenarios BDD (Gherkin)
#
# Especificacion SDD : docs/specs/US-F23-reporte-spec.md
# Issue de origen    : SebastianAguilera/ingenieria-y-calidad#112
# Endpoints          : US-22  GET /api/proyectos/{id}/reporte
#                             GET /api/proyectos/{id}/sprints/{sprintId}/reporte
#                      US-24  GET /api/proyectos/{id}/reporte.pdf
#                             GET /api/proyectos/{id}/sprints/{sprintId}/reporte.pdf
# Trazabilidad       : Issue #112 -> SDD -> criterios -> escenarios BDD ->
#                      tests Vitest/RTL -> codigo
# Rama               : feature/US-F23-reporte
# =============================================================================
#
# NOTA: cada escenario se automatiza con Vitest + React Testing Library. Los
# tests de pantalla usan mocks de frontend/src/api/reportes.ts y sprints.ts.
# -----------------------------------------------------------------------------
# CRITERIOS DE ACEPTACION -> ESCENARIOS
# -----------------------------------------------------------------------------
# [CA-1]  Reporte del proyecto completo al entrar.          -> E-01, E-02
# [CA-2]  Reporte de un sprint desde el selector.           -> E-03
# [CA-3]  Proyecto sin historias: tablas vacias.            -> E-04
# [CA-4]  null se muestra como "Sin datos".                  -> E-05
# [CA-5]  Descargar JSON con el contenido de la API.         -> E-06
# [CA-6]  Error de red con reintento.                        -> E-07
# [CA-7]  No miembro: mensaje de 404.                        -> E-08
# [CA-8]  Descargar PDF del alcance elegido.                 -> E-09, E-10
# [CA-9]  Falla del PDF sin perder el reporte.               -> E-11
# [CA-10] Sprints de defectos por nombre.                    -> E-12
# [CA-11] Nombre de los archivos.                            -> E-13, E-14
# [CA-12] Entrada "Reporte" en la barra lateral.             -> E-15
# [Limite] Proyecto sin sprints.                             -> E-16
# -----------------------------------------------------------------------------

Característica: Vista del reporte de proyecto y sprint
  Como líder de un proyecto
  Quiero ver el reporte consolidado del proyecto o de un sprint y descargarlo
  Para documentar el avance sin pedir los datos a mano

  Antecedentes:
    Dado que soy miembro del proyecto 1 "Gestión Ágil"
    Y que el proyecto tiene los sprints "Sprint 1" (Cerrado) y "Sprint 2" (Activo)

  @caso_normal
  Escenario: E-01 Reporte del proyecto completo al entrar
    Cuando entro a "/proyectos/1/reporte"
    Entonces se consulta GET /api/proyectos/1/reporte
    Y el selector de alcance muestra "Proyecto completo"
    Y veo el encabezado con el nombre, las fechas y el estado del proyecto

  @caso_normal
  Escenario: E-02 Secciones del reporte del proyecto
    Dado que el reporte tiene la historia "US-1" con 5 SP, 8 h estimadas y 6 h reales
    Y tiene el defecto "DEF-3" de severidad "major" en estado "open"
    Cuando entro al reporte
    Entonces veo las tarjetas de métricas del proyecto
    Y la tabla de historias muestra "US-1", "5", "8 h" y "6 h"
    Y la tabla de defectos muestra "DEF-3", "Mayor" y "Abierto"

  @caso_normal
  Escenario: E-03 Reporte de un sprint
    Cuando elijo "Sprint 2" en el selector de alcance
    Entonces se consulta GET /api/proyectos/1/sprints/2/reporte
    Y veo el identificador y el sprint goal del sprint
    Y las tablas muestran solo las historias y defectos del sprint

  @caso_limite
  Escenario: E-04 Proyecto sin historias ni defectos
    Dado que el reporte tiene las listas de historias y defectos vacías
    Cuando entro al reporte
    Entonces la tabla de historias dice "Este reporte no tiene historias"
    Y la tabla de defectos dice "Este reporte no tiene defectos"

  @caso_limite
  Escenario: E-05 Métrica sin datos
    Dado que la velocidad del reporte es null
    Cuando entro al reporte
    Entonces la tarjeta "Velocidad" dice "Sin datos"

  @caso_normal
  Escenario: E-06 Descargar JSON
    Dado que el reporte del proyecto está cargado
    Cuando presiono "Descargar JSON"
    Entonces se descarga un archivo con el mismo contenido que devolvió la API

  @caso_error
  Escenario: E-07 Error de red con reintento
    Dado que la consulta del reporte falla por un error de red
    Cuando entro al reporte
    Entonces veo "No se pudo cargar el reporte. Intentá nuevamente."
    Y cuando presiono "Reintentar" el reporte se vuelve a pedir

  @caso_error
  Escenario: E-08 No miembro del proyecto
    Dado que la API responde 404 PROYECTO_NO_ENCONTRADO
    Cuando entro al reporte
    Entonces veo "No tenés acceso a este proyecto o no existe"

  @caso_normal
  Escenario: E-09 Descargar PDF del proyecto
    Cuando presiono "Descargar PDF" con el alcance "Proyecto completo"
    Entonces se pide GET /api/proyectos/1/reporte.pdf
    Y se descarga el archivo PDF

  @caso_normal
  Escenario: E-10 Descargar PDF de un sprint
    Dado que elegí "Sprint 2" en el selector
    Cuando presiono "Descargar PDF"
    Entonces se pide GET /api/proyectos/1/sprints/2/reporte.pdf

  @caso_error
  Escenario: E-11 Falla la generación del PDF
    Dado que GET .../reporte.pdf responde con error
    Cuando presiono "Descargar PDF"
    Entonces veo "No se pudo descargar el PDF. Intentá nuevamente."
    Y el reporte sigue visible

  @caso_normal
  Escenario: E-12 Sprints de los defectos por nombre
    Dado que el defecto "DEF-3" se detectó en el sprint 1 y no tiene sprint de resolución
    Cuando entro al reporte
    Entonces su fila muestra "Sprint 1" como sprint de detección
    Y "—" como sprint de resolución

  @caso_normal
  Escenario: E-13 Nombre del archivo del proyecto
    Dado que hoy es "2026-11-16"
    Cuando descargo el reporte del proyecto en JSON
    Entonces el archivo se llama "reporte-gestion-agil-proyecto-2026-11-16.json"

  @caso_normal
  Escenario: E-14 Nombre del archivo de un sprint
    Dado que hoy es "2026-11-16"
    Cuando descargo el PDF de "Sprint 2"
    Entonces el archivo se llama "reporte-proyecto-1-sprint-sprint-2-2026-11-16.pdf"

  @caso_normal
  Escenario: E-15 Entrada en la barra lateral
    Cuando miro la barra lateral del proyecto
    Entonces existe el enlace "Reporte" a "/proyectos/1/reporte"

  @caso_limite
  Escenario: E-16 Proyecto sin sprints
    Dado que el proyecto no tiene sprints
    Cuando entro al reporte
    Entonces el selector solo ofrece "Proyecto completo"
