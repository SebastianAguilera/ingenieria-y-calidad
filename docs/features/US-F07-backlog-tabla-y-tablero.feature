# language: es
# =============================================================================
# US-F07 - Backlog con vista de tabla y tablero
# Escenarios BDD de la pantalla Backlog del detalle de un proyecto.
#
# Especificacion SDD : docs/specs/US-F07-backlog-tabla-y-tablero-spec.md
# Endpoints          : US-02 (GET/POST/PATCH /api/proyectos/{id}/historias)
# Trazabilidad       : Spec -> criterios de aceptacion -> escenarios -> tests
# =============================================================================

Funcionalidad: US-F07 - Backlog con vista de tabla y tablero
  Como miembro del equipo
  Quiero ver y gestionar las historias del Product Backlog
  Para planificar el alcance del proyecto.

  Antecedentes:
    Dado que estoy dentro de un proyecto existente en "/proyectos/1/backlog"

  Regla: Listado y estados de la pantalla

    Escenario: E-01 - Carga inicial del backlog
      Cuando se abre la pantalla Backlog
      Entonces se muestra "Cargando historias…"
      Y no se muestra ninguna tabla ni tablero

    Escenario: E-02 - Listado con historias
      Dado que el proyecto tiene las historias "US-01" y "US-02"
      Cuando la carga termina
      Entonces la vista Tabla muestra las columnas "Código", "Título", "Estado", "Prioridad" y "Acciones"
      Y se muestran los códigos "US-01" y "US-02"

    Escenario: E-03 - Proyecto sin historias
      Dado que el proyecto no tiene historias
      Cuando la carga termina
      Entonces se muestra "Todavía no hay historias"
      Y se ofrece la acción "Nueva historia"

    Escenario: E-04 - Error al listar
      Dado que la API de historias falla
      Cuando la carga termina
      Entonces se muestra un error con la acción "Reintentar"
      Y al presionar "Reintentar" se vuelve a consultar el listado

  Regla: Vista Tabla

    Escenario: E-05 - Cambio inline del estado
      Dado que la historia "US-01" está en estado "Backlog"
      Cuando cambio su estado a "Hecho"
      Entonces se envía un PATCH con el nuevo estado
      Y la fila refleja el estado "Hecho"

    Escenario: E-06 - Cambio inline de la prioridad
      Dado que la historia "US-01" tiene prioridad "Media"
      Cuando cambio su prioridad a "Alta"
      Entonces se envía un PATCH con la nueva prioridad
      Y la fila refleja la prioridad "Alta"

    Escenario: E-07 - Error al actualizar
      Dado que el PATCH de una historia falla
      Cuando cambio su estado
      Entonces la fila vuelve al estado anterior
      Y se muestra un mensaje de error

  Regla: Alta y edicion

    Escenario: E-08 - Alta de una historia
      Cuando abro "Nueva historia"
      Y completo "Título", "Descripción", "Prioridad" y "Estado"
      Y presiono "Crear historia"
      Entonces se envía un POST con los datos ingresados
      Y la historia aparece en la vista activa

    Escenario: E-09 - Validacion del titulo
      Cuando abro "Nueva historia"
      Y dejo el título vacío o escribo más de 200 caracteres
      Entonces veo el error de validación
      Y no se envía ninguna petición al backend

    Escenario: E-10 - Edicion de una historia
      Dado que existe la historia "US-01"
      Cuando abro "Editar US-01"
      Y modifico su título
      Y presiono "Guardar cambios"
      Entonces se envía un PATCH con los campos modificados
      Y la historia actualizada se refleja en la tabla

  Regla: Vista Tablero

    Escenario: E-11 - Agrupacion por estado
      Cuando cambio a la vista "Tablero"
      Entonces veo las columnas "Backlog", "En Progreso" y "Hecho"
      Y cada historia aparece en la columna de su estado

    Escenario: E-12 - Movimiento por teclado
      Dado que la historia "US-01" está en "Backlog"
      Cuando uso el selector "Mover US-01 a" y elijo "En Progreso"
      Entonces se envía un PATCH con el nuevo estado
      Y la tarjeta se mueve a la columna "En Progreso"

    Escenario: E-13 - Movimiento por arrastre
      Dado que la historia "US-01" está en "Backlog"
      Cuando la arrastro y la suelto en "Hecho"
      Entonces se envía un PATCH con el nuevo estado
      Y la tarjeta se mueve a la columna "Hecho"

    Escenario: E-14 - Soltar en la misma columna
      Dado que la historia "US-01" está en "Backlog"
      Cuando la suelto en la misma columna "Backlog"
      Entonces no se envía ninguna petición
