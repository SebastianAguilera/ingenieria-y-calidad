# language: es
# =============================================================================
# US-F11 - Editar y eliminar proyectos desde la interfaz
# Especificación SDD : docs/specs/US-F11-editar-eliminar-proyectos-spec.md
# Issue de origen    : #67
# Trazabilidad       : Issue #67 -> SDD -> CA -> BDD -> tests Go/Vitest -> código
# =============================================================================
# NOTA: especificación ejecutable futura (sin runner Godog). Cada escenario se
# automatiza como subtest en los *_test.go / *.test.tsx correspondientes.

Funcionalidad: US-F11 - Editar y eliminar proyectos desde la interfaz
  Como miembro de un proyecto
  Quiero editar sus datos, ver su información general y eliminarlo
  Para mantener la información al día y quitar proyectos que ya no sirven

  Antecedentes:
    Dado que el usuario "ana" está autenticada y es miembro del proyecto "Proyecto Alfa" (id 1, estado "Activo")

  Regla: Descripción del proyecto (RB-02, CA-1)

    Escenario: E-01 - Alta y lectura incluyen descripción
      Cuando "ana" crea un proyecto con nombre "Proyecto Beta", descripción "  Plataforma interna  " y fecha de inicio "2026-11-01"
      Entonces la respuesta es "201" con "descripcion" igual a "Plataforma interna"
      Y al pedir "GET /api/proyectos/{id}" la respuesta incluye "descripcion" "Plataforma interna"
      Y el listado "GET /api/proyectos" incluye "descripcion" para cada proyecto

    Escenario: E-02 - Descripción vacía y límite
      Dado que la descripción es solo espacios
      Cuando "ana" actualiza el proyecto 1
      Entonces "descripcion" se guarda como cadena vacía
      Y una descripción de 2001 caracteres responde "422 VALIDACION" en el campo "descripcion"

  Regla: Eliminar desde el listado (RB-06, CA-2)

    Escenario: E-03 - Eliminar proyecto con confirmación desde el listado
      Dado que "ana" ve "Proyecto Alfa" en el listado
      Cuando abre la acción de eliminar de la fila y confirma en el modal
      Entonces se envía "DELETE /api/proyectos/1"
      Y al recibir "204" la fila desaparece sin recargar la página
      Y la acción de eliminar no dispara la navegación al detalle de la fila

    Escenario: E-04 - Eliminar el último proyecto muestra estado vacío
      Dado que el listado tiene un solo proyecto
      Cuando "ana" lo elimina y recibe "204"
      Entonces se muestra el estado vacío de proyectos

    Escenario: E-05 - No se puede eliminar un proyecto con historial
      Dado que "DELETE /api/proyectos/1" responde "409 PROYECTO_CON_HISTORIAL"
      Cuando "ana" confirma la eliminación desde el listado
      Entonces se muestra el mensaje de error y la fila se conserva

  Regla: Información general y edición (RB-03, RB-04, RB-05, CA-3, CA-4, CA-5)

    Escenario: E-06 - La ruta index muestra Información general
      Cuando "ana" abre "/proyectos/1"
      Entonces ve la sección "Información general" con nombre, descripción, estado, fecha de inicio, fecha de fin y cantidad de usuarios
      Y el menú lateral no muestra "Resumen"

    Escenario: E-07 - Editar solo datos del proyecto
      Dado que el formulario está precargado con los datos del proyecto
      Cuando "ana" modifica nombre, descripción o fechas y guarda
      Entonces se envía "PUT /api/proyectos/1" y la UI muestra los datos actualizados sin recargar

    Escenario: E-08 - Cambiar solo el estado
      Cuando "ana" cambia el estado a "Cerrado" y guarda
      Entonces se envía solo "PATCH /api/proyectos/1/estado"

    Escenario: E-09 - Cambiar datos y estado
      Cuando "ana" modifica datos y el estado en el mismo guardado
      Entonces se envía "PUT /api/proyectos/1" y luego "PATCH /api/proyectos/1/estado"

    Escenario: E-10 - Proyecto cerrado requiere reapertura explícita
      Dado que el proyecto 1 está "Cerrado"
      Cuando "ana" abre Información general
      Entonces ve "Reabrí el proyecto para editar sus datos" y el botón "Reabrir proyecto"
      Y al reabrir se envía "PATCH" a "Activo" y no se envía "PUT" automáticamente

    Escenario: E-11 - Fallo parcial al cambiar estado tras guardar datos
      Dado que "PUT /api/proyectos/1" responde "200" y "PATCH /api/proyectos/1/estado" falla
      Cuando "ana" guarda datos y estado
      Entonces los datos del "PUT" se conservan, el formulario permanece abierto y puede reintentar solo el estado

  Regla: Eliminar desde información general (CA-7)

    Escenario: E-12 - Eliminar desde el detalle navega al listado
      Cuando "ana" confirma eliminar el proyecto desde Información general y recibe "204"
      Entonces navega a "/proyectos"

  Regla: Validaciones y errores en UI (CA-6)

    Escenario: E-13 - Errores de validación y acceso
      Dado que el nombre está vacío o la fecha de fin es anterior a la de inicio
      Entonces la UI muestra el error de campo sin enviar la solicitud inválida
      Y si el backend responde "422 VALIDACION" se muestra el error en el campo correspondiente
      Y si responde "404 PROYECTO_NO_ENCONTRADO" se muestra "No tenés acceso a este proyecto o no existe"
      Y ante error de red los datos del formulario se conservan y se puede reintentar

    Escenario: E-14 - Modales accesibles
      Cuando se abre un modal de confirmación o edición
      Entonces tiene role "dialog", aria-modal "true" y aria-labelledby
      Y recibe foco inicial, se cierra con "Escape" y devuelve el foco al disparador
