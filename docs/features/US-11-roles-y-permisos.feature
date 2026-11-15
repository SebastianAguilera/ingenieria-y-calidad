# language: es
# =============================================================================
# US-11 - Roles y permisos dentro del proyecto
# Escenarios BDD (Gherkin)
#
# Especificacion SDD : docs/specs/US-11-roles-y-permisos-spec.md
# Issue de origen    : SebastianAguilera/ingenieria-y-calidad#69
# Trazabilidad       : Issue #69 -> SDD -> criterios -> escenarios BDD ->
#                      tests Go -> codigo
# =============================================================================
#
# NOTA: especificacion ejecutable futura (sin runner de Godog): cada escenario se
# automatiza como subtest en los *_test.go de domain, service, handler y repository.
# Actores: ana = LIDER, ben = MIEMBRO, cris = no miembro.

Característica: Roles y permisos dentro del proyecto
  Como líder de un proyecto
  Quiero que existan roles con permisos distintos
  Para que los miembros trabajen el contenido sin eliminar el proyecto ni gestionar el equipo

  Antecedentes:
    Dado un proyecto "Metrics" donde ana es LIDER y ben es MIEMBRO

  # --- Casos normales ---------------------------------------------------------
  Escenario: E-01 El creador de un proyecto queda como líder
    Cuando ana crea un proyecto
    Entonces ana queda vinculada al proyecto con rol "LIDER"

  Escenario: E-02 El usuario agregado entra como miembro por defecto
    Cuando ana agrega a cris sin indicar rol
    Entonces cris queda vinculada con rol "MIEMBRO"

  Escenario: E-03 El líder elimina el proyecto
    Cuando ana pide DELETE /api/proyectos/{id}
    Entonces recibe 204

  Escenario: E-04 El líder gestiona el equipo
    Cuando ana quita a ben del proyecto
    Entonces recibe 204 y ben deja de pertenecer al proyecto

  Escenario: E-05 El miembro trabaja el contenido
    Cuando ben crea una historia en el proyecto
    Entonces la operación se permite

  Escenario: E-06 El detalle informa el rol del usuario que pide
    Cuando ben pide GET /api/proyectos/{id}
    Entonces la respuesta incluye "rol_actual" igual a "MIEMBRO"

  # --- Casos alternativos -----------------------------------------------------
  Escenario: E-07 El miembro se quita a sí mismo
    Cuando ben pide DELETE /api/proyectos/{id}/usuarios/{id de ben}
    Entonces recibe 204 y ben deja de pertenecer al proyecto

  Escenario: E-08 Un líder se quita si hay otro líder
    Dado que ben también es LIDER
    Cuando ana se quita a sí misma
    Entonces recibe 204

  # --- Errores ----------------------------------------------------------------
  Escenario: E-09 El miembro no puede eliminar el proyecto
    Cuando ben pide DELETE /api/proyectos/{id}
    Entonces recibe 403 ACCESO_DENEGADO y el proyecto no cambia

  Escenario: E-10 El miembro no puede gestionar el equipo
    Cuando ben intenta agregar o quitar a otro integrante
    Entonces recibe 403 ACCESO_DENEGADO y el equipo no cambia

  Escenario: E-11 El no miembro recibe 404, no 403
    Cuando cris pide DELETE /api/proyectos/{id}
    Entonces recibe 404 PROYECTO_NO_ENCONTRADO

  Escenario: E-12 No se puede quitar al último líder
    Cuando ana, única líder, se quita a sí misma
    Entonces recibe 409 ULTIMO_LIDER

  Escenario: E-13 No se puede degradar al último líder
    Cuando se cambia el rol de ana, única líder, a "MIEMBRO"
    Entonces se rechaza con ULTIMO_LIDER y el rol no cambia

  Escenario: E-14 Un rol inválido se rechaza
    Cuando ana agrega a cris con rol "Scrum Master"
    Entonces recibe 422 VALIDACION con campo "rol"

  # --- Migración de datos -----------------------------------------------------
  Escenario: E-15 La migración deja un líder en cada proyecto
    Dado proyectos existentes con los roles originales de US-08 y sin líder
    Cuando se ejecuta la migración
    Entonces los roles viejos pasan a "MIEMBRO"
    Y el integrante de menor id de cada proyecto pasa a "LIDER"
    Y ningún proyecto con equipo queda sin líder

  Escenario: E-16 La migración es idempotente
    Dado una base ya migrada
    Cuando se ejecuta la migración otra vez
    Entonces los roles no cambian
