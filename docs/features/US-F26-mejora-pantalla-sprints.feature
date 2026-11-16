# language: es
# US-F26 - Mejora de pantalla de sprints y Sprint Backlog
# Trazabilidad: Issue #132 -> docs/specs/US-F26-mejora-pantalla-sprints-spec.md
# -> escenarios BDD -> tests Vitest/RTL -> código frontend

Característica: US-F26 - Mejora de pantalla de sprints y Sprint Backlog
  Como miembro de un proyecto
  Quiero ver el sprint con su información completa, sus métricas y su tabla de historias
  Para entender el avance del sprint y gestionar sus historias desde ahí

  Escenario: CA-1 La fila y "Ver sprint" abren el detalle
    Dado un sprint en el listado
    Cuando pulso cualquier parte de la fila o el botón "Ver sprint"
    Entonces se abre el detalle del sprint

  Escenario: CA-2 Filtro por estado con contador
    Dado el listado con sprints en varios estados
    Cuando elijo el chip "Activo"
    Entonces solo se muestran sprints activos y el contador coincide

  Escenario: CA-3 Sprint cerrado sin acciones de ciclo de vida
    Dado un sprint Cerrado en el listado
    Cuando se muestran sus acciones
    Entonces no aparecen Activar ni Cerrar y sí "Ver sprint"

  Escenario: CA-4 Cabecera de detalle de sprint activo
    Dado un sprint Activo
    Cuando se abre el detalle
    Entonces la cabecera muestra estado, fechas, duración, Sprint Goal y el botón "Asignar historia"

  Escenario: CA-5 KPIs del sprint
    Dado el detalle de un sprint
    Cuando carga correctamente
    Entonces se muestran los KPIs con SP planificados, SP completados, porcentaje, historias completadas y horas estimadas y reales

  Escenario: CA-6 Métricas con error no rompen la tabla
    Dado un fallo al cargar las métricas
    Cuando se abre el detalle
    Entonces la tabla de historias funciona igual y los KPIs muestran "—"

  Escenario: CA-7 Edición inline de estado y prioridad
    Dado una historia en un sprint no Cerrado
    Cuando se cambia su estado o prioridad
    Entonces se envía el cambio y la tabla refleja el nuevo valor

  Escenario: CA-8 Sprint cerrado en solo lectura
    Dado un sprint Cerrado
    Cuando se abre el detalle
    Entonces el estado y la prioridad se muestran como chips de solo lectura y no aparece "Quitar del sprint"

  Escenario: CA-9 Quitar historia del sprint
    Dado una historia en un sprint no Cerrado
    Cuando se pulsa "Quitar del sprint" y se confirma
    Entonces la historia sale del sprint y la tabla se actualiza

  Escenario: CA-10 Estado vacío en sprint activo
    Dado un sprint sin historias
    Cuando se abre el detalle
    Entonces aparece un estado vacío con el botón "Asignar historia" si el sprint está Activo
