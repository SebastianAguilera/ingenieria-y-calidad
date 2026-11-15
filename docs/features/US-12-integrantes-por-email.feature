# language: es
# =============================================================================
# US-12 - Integrantes por email con rol
# Escenarios BDD (Gherkin)
#
# Especificacion SDD : docs/specs/US-12-integrantes-por-email-spec.md
# Issue de origen    : SebastianAguilera/ingenieria-y-calidad#68
# Trazabilidad       : Issue #68 -> SDD -> criterios -> escenarios BDD ->
#                      tests Go -> codigo
# =============================================================================
#
# NOTA: especificacion ejecutable futura (sin runner de Godog): cada escenario se
# automatiza como subtest en los *_test.go de repository, service, handler y router.
# Actores: ana = LIDER, ben = MIEMBRO, cris = no miembro, ada = usuario registrado
# que todavia no pertenece al proyecto.

Característica: Integrantes por email con rol
  Como líder de un proyecto
  Quiero agregar integrantes por email, cambiarles el rol y quitarlos
  Para armar mi equipo sin que nadie entre a un proyecto sin ser invitado

  Antecedentes:
    Dado un proyecto "Metrics" donde ana es LIDER y ben es MIEMBRO
    Y ada es una persona registrada que no pertenece al proyecto
    Y cris es una persona que no pertenece al proyecto

  # --- Casos normales ---------------------------------------------------------
  Escenario: E-01 Un miembro lista el equipo con sus roles
    Cuando ben pide GET /api/proyectos/{id}/usuarios
    Entonces recibe 200 con id, nombre, email y rol de cada integrante
    Y ana aparece con rol "LIDER"
    Y ben aparece con rol "MIEMBRO"

  Escenario: E-02 El líder agrega a una persona registrada por email
    Cuando ana envía POST /api/proyectos/{id}/usuarios con el email de ada y rol "MIEMBRO"
    Entonces recibe 200 con ada vinculada y su rol "MIEMBRO"

  Escenario: E-03 El alta sin rol deja a la persona como MIEMBRO
    Cuando ana agrega a ada sin indicar rol
    Entonces ada queda vinculada con rol "MIEMBRO"

  Escenario: E-04 El líder cambia el rol de un integrante
    Cuando ana envía PATCH /api/proyectos/{id}/usuarios/{id de ben} con rol "LIDER"
    Entonces recibe 200 y el rol de ben queda en "LIDER"

  Escenario: E-05 El líder quita a un integrante
    Cuando ana envía DELETE /api/proyectos/{id}/usuarios/{id de ben}
    Entonces recibe 204 y ben deja de pertenecer al proyecto

  # --- Casos alternativos -----------------------------------------------------
  Escenario: E-06 El email con mayúsculas y espacios se normaliza
    Cuando ana agrega a "  ADA@UTN.EDU.AR "
    Entonces se encuentra a ada y queda vinculada al proyecto

  Escenario: E-07 Un miembro se quita a sí mismo
    Cuando ben envía DELETE /api/proyectos/{id}/usuarios/{su id}
    Entonces recibe 204 y deja de pertenecer al proyecto

  Escenario: E-08 Un líder se degrada si queda otro líder
    Dado que ben también es LIDER
    Cuando ana cambia su propio rol a "MIEMBRO"
    Entonces recibe 200 y su rol se actualiza

  # --- Límite -----------------------------------------------------------------
  Escenario: E-09 Un proyecto sin integrantes devuelve un array vacío
    Dado un proyecto recién creado sin equipo
    Cuando su líder pide GET /api/proyectos/{id}/usuarios
    Entonces recibe 200 con []

  # --- Errores ----------------------------------------------------------------
  Escenario: E-10 Un no miembro no puede listar el equipo
    Cuando cris pide GET /api/proyectos/{id}/usuarios
    Entonces recibe 404 PROYECTO_NO_ENCONTRADO

  Escenario: E-11 Un miembro no líder no puede gestionar el equipo
    Cuando ben intenta POST, PATCH o DELETE sobre el equipo
    Entonces recibe 403 ACCESO_DENEGADO y nada cambia

  Escenario: E-12 Un email inexistente no crea usuario
    Cuando ana agrega el email "nadie@utn.edu.ar"
    Entonces recibe 404 con "No existe un usuario registrado con ese email"
    Y no se crea ningún usuario

  Escenario: E-13 Un email que ya es integrante se rechaza
    Cuando ana agrega el email de ben
    Entonces recibe 409 USUARIO_DUPLICADO con campo "email"
    Y el mensaje es "Ese usuario ya es integrante"

  Escenario: E-14 Un email vacío o inválido se rechaza
    Cuando ana agrega el email "Nombre <ada@utn.edu.ar>"
    Entonces recibe 422 VALIDACION con campo "email"

  Escenario: E-15 Un rol inválido se rechaza
    Cuando ana agrega a ada con rol "Scrum Master"
    Entonces recibe 422 VALIDACION con campo "rol"

  Escenario: E-16 El último líder no puede quitarse ni degradarse
    Cuando ana, única líder, se quita a sí misma o se cambia a "MIEMBRO"
    Entonces recibe 409 ULTIMO_LIDER y el equipo no cambia
