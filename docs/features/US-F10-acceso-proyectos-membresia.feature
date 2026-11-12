# language: es
# =============================================================================
# US-F10 - Acceso a proyectos por membresia
# Escenarios BDD (Gherkin)
#
# Especificacion SDD : docs/specs/US-F10-acceso-proyectos-membresia-spec.md
# Issue de origen    : SebastianAguilera/ingenieria-y-calidad#66
# Endpoints          : US-01 (GET/POST /api/proyectos), US-08 (usuarios de un
#                      proyecto), US-10 (JWT y middleware) y todas las rutas
#                      anidadas bajo /api/proyectos/{id}/...
# Trazabilidad       : Issue #66 -> SDD -> criterios -> escenarios BDD ->
#                      tests Go / Vitest -> codigo
# Rama               : develop
# =============================================================================
#
# NOTA: especificacion ejecutable futura. El repo no tiene runner de Godog
# (methodology.md seccion 4): cada escenario se automatiza como subtest
# t.Run("E-nn", ...) en los *_test.go / *.test.tsx indicados al final.
#
# Entorno: BASE_URL http://localhost:$PORT, PostgreSQL 16, JWT_SECRET.
# Actores: A = usuario "ana" (creador), B = usuario "ben" (no miembro).
# Verificaciones HTTP con curl o httptest; frontend con Vitest + RTL.
# -----------------------------------------------------------------------------
# NARRATIVA (US-F10)
# -----------------------------------------------------------------------------
# Como usuario, quiero ver y gestionar solo los proyectos de los que soy
# miembro, para que el trabajo ajeno no quede expuesto en mi sesion.
#
# El middleware de autenticacion expone el sub del JWT como identidad y un
# guard registrado justo despues exige membresia en toda ruta con :id de
# proyecto: quien no es miembro recibe el mismo 404 que un proyecto
# inexistente. El alta de proyecto vincula al creador en la misma
# transaccion y el listado se filtra por membresia.
#
# -----------------------------------------------------------------------------
# CRITERIOS DE ACEPTACION / SDD seccion "Criterios de aceptacion"
# -----------------------------------------------------------------------------
# [CA-1] sub del JWT disponible como identidad; sin sub -> 401 TOKEN_INVALIDO.
#                                                                 -> E-01
# [CA-2] Listado filtrado por membresia (A ve los suyos, no los de B).
#                                                                 -> E-02
# [CA-3] Sin proyectos -> 200 con total 0 (estado vacio).
#                                                                 -> E-03
# [CA-4] Crear proyecto vincula al creador (cantidad_usuarios 1) y el
#        proyecto aparece en su listado.                           -> E-04
# [CA-5] Proyecto ajeno en el detalle -> 404 PROYECTO_NO_ENCONTRADO.
#                                                                 -> E-05, E-06
# [CA-6] Proyecto ajeno en rutas anidadas (historias, sprints,
#        POST/DELETE usuarios) -> 404 sin llegar al service; una
#        vez agregado, B pasa a operar normalmente.                -> E-07..E-10
# [CA-7] El guard no interviene en rutas sin :id ni en id malformado.
#                                                                 -> E-11, E-12
# [CA-8] Cualquier miembro puede editar, cerrar y eliminar (opcion A).
#                                                                 -> E-13
# [Frontend] Mensaje de 404 y estado vacio en espanol.             -> E-14, E-15
# =============================================================================

Funcionalidad: US-F10 - Acceso a proyectos por membresia
  Como usuario autenticado
  Quiero ver y operar solo los proyectos de los que soy miembro
  Para que el trabajo de otros equipos no quede expuesto en mi sesión.

  Antecedentes:
    Dado que el usuario "ana" está autenticado con su token
    Y que el usuario "ben" está autenticado con su token
    Y que "ana" creó el proyecto "Proyecto Alfa"

  Regla: Identidad del token (RB-01)

    Escenario: E-01 - El handler recibe el sub como identidad
      Dado que el token de "ana" tiene el claim "sub" con su id numérico
      Cuando "ana" pide "GET /api/proyectos"
      Entonces el servicio de listado recibe el usuario id de "ana"
      Y si el token no trae "sub" o su "sub" no es numérico
      Entonces la respuesta es "401 TOKEN_INVALIDO"

  Regla: Listado filtrado por membresía (RB-02)

    Escenario: E-02 - Cada usuario solo ve sus proyectos
      Dado que "ana" es miembro de "Proyecto Alfa" y "Proyecto Delta"
      Y que "ben" no es miembro de ninguno
      Cuando "ana" pide "GET /api/proyectos"
      Entonces recibe "Proyecto Alfa" y "Proyecto Delta"
      Y cuando "ben" pide "GET /api/proyectos"
      Entonces recibe una lista sin "Proyecto Alfa"

    Escenario: E-03 - Usuario sin proyectos
      Dado que "ben" no es miembro de ningún proyecto
      Cuando "ben" pide "GET /api/proyectos"
      Entonces la respuesta es "200" con "total" en "0"
      Y "proyectos" es una lista vacía y no "null"

  Regla: El alta vincula al creador (RB-03)

    Escenario: E-04 - Crear proyecto deja al creador como miembro
      Dado que "ben" no es miembro de ningún proyecto
      Cuando "ben" crea el proyecto "Proyecto Beta" con fecha de inicio "2026-11-01"
      Entonces la respuesta es "201" con "cantidad_usuarios" en "1"
      Y "Proyecto Beta" aparece en el listado de "ben"
      Y si el alta del vínculo falla, no queda ningún proyecto a medias

  Regla: Proyecto ajeno responde 404 (RB-04)

    Escenario: E-05 - Detalle de un proyecto ajeno
      Dado que "ben" no es miembro de "Proyecto Alfa"
      Cuando "ben" pide "GET /api/proyectos/1"
      Entonces la respuesta es "404 PROYECTO_NO_ENCONTRADO"
      Y el mensaje es "el proyecto no existe"

    Escenario: E-06 - Editar, cerrar y borrar un proyecto ajeno
      Dado que "ben" no es miembro de "Proyecto Alfa"
      Cuando "ben" pide "PUT /api/proyectos/1" con nombre "Hack"
      Entonces la respuesta es "404 PROYECTO_NO_ENCONTRADO"
      Y cuando pide "PATCH /api/proyectos/1/estado" con estado "Cerrado"
      Entonces la respuesta es "404 PROYECTO_NO_ENCONTRADO"
      Y cuando pide "DELETE /api/proyectos/1"
      Entonces la respuesta es "404 PROYECTO_NO_ENCONTRADO"

    Escenario: E-07 - Ruta anidada de un proyecto ajeno
      Dado que "ben" no es miembro de "Proyecto Alfa"
      Cuando "ben" pide "GET /api/proyectos/1/historias"
      Entonces la respuesta es "404 PROYECTO_NO_ENCONTRADO"
      Y el servicio de historias no fue invocado
      Y cuando "ben" pide "GET /api/proyectos/1/sprints"
      Entonces la respuesta es "404 PROYECTO_NO_ENCONTRADO"

    Escenario: E-08 - No puede auto-insertarse en un proyecto ajeno
      Dado que "ben" no es miembro de "Proyecto Alfa"
      Cuando "ben" pide "POST /api/proyectos/1/usuarios" con su propio id
      Entonces la respuesta es "404 PROYECTO_NO_ENCONTRADO"
      Y "ben" sigue sin ser miembro del proyecto

    Escenario: E-09 - No puede quitarse a otros de un proyecto ajeno
      Dado que "ana" es miembro de "Proyecto Alfa"
      Cuando "ben" pide "DELETE /api/proyectos/1/usuarios/1"
      Entonces la respuesta es "404 PROYECTO_NO_ENCONTRADO"
      Y "ana" sigue siendo miembro

    Escenario: E-10 - Agregado por A, B pasa a ver el proyecto
      Dado que "ben" no es miembro de "Proyecto Alfa"
      Y que "ben" recibe "404" en "GET /api/proyectos/1"
      Cuando "ana" agrega a "ben" con "POST /api/proyectos/1/usuarios"
      Entonces la respuesta de "ana" es "201"
      Y "ben" ve "Proyecto Alfa" en "GET /api/proyectos"
      Y "ben" recibe "200" en "GET /api/proyectos/1"
      Y "ben" recibe "200" en "GET /api/proyectos/1/historias"

  Regla: El guard no interviene donde no corresponde (RB-05, RB-06)

    Escenario: E-11 - Rutas sin :id de proyecto
      Dado que "ben" está autenticado
      Cuando "ben" pide "GET /health"
      Entonces la respuesta es "200" con "status" en "ok"
      Y cuando pide "GET /api/usuarios"
      Entonces recibe "200"
      Y cuando pide "POST /api/proyectos" con nombre "Proyecto Gamma"
      Entonces la respuesta es "201" y el creador queda vinculado

    Escenario: E-12 - Id malformado no pasa por el guard
      Dado que "ana" está autenticado
      Cuando "ana" pide "GET /api/proyectos/abc"
      Entonces la respuesta es "400 PARAMETRO_INVALIDO"

  # NOTA US-11: eliminar el proyecto y gestionar el equipo pasaron a exigir LIDER
  # (403 ACCESO_DENEGADO para MIEMBRO); editar y cerrar siguen abiertos a todo miembro.
  Regla: Cualquier miembro opera el proyecto (RB-07)

    Escenario: E-13 - El miembro edita, cierra y borra su proyecto
      Dado que "ben" es miembro de "Proyecto Beta"
      Cuando "ben" pide "PUT /api/proyectos/2" con nombre "Proyecto Beta v2"
      Entonces la respuesta es "200"
      Y cuando pide "PATCH /api/proyectos/2/estado" con estado "Cerrado"
      Entonces la respuesta es "200"
      Y cuando pide "DELETE /api/proyectos/2"
      Entonces la respuesta es "200"

  Regla: Estados vacíos y errores en la interfaz

    Escenario: E-14 - Listado vacío explica qué falta
      Dado que "ben" no es miembro de ningún proyecto
      Cuando "ben" abre "/proyectos"
      Entonces se muestra "Todavía no participás en ningún proyecto"
      Y se ofrece la acción "Crear proyecto"

    Escenario: E-15 - Acceso denegado se muestra como proyecto inexistente
      Dado que "ben" no es miembro del proyecto 1
      Cuando "ben" abre "/proyectos/1/backlog"
      Entonces se muestra "No tenés acceso a este proyecto o no existe"
      Y no se muestra el backlog ni la tablita de historias

    Escenario: E-16 - Proyecto ajeno no aparece tras un refresco
      Dado que "ben" no es miembro del proyecto 1
      Cuando "ben" vuelve a pedir "GET /api/proyectos"
      Entonces el proyecto 1 no está en la lista
