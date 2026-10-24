# language: es
# =============================================================================
# US-08 - Gestión independiente de usuarios y asociación a proyectos con roles
# Escenarios BDD (Gherkin)
#
# Issue de origen     : SebastianAguilera/ingenieria-y-calidad#39
# Especificacion SDD  : docs/specs/US-08-gestion-usuarios-spec.md
# Metodologia         : docs/methodology.md seccion 4 (BDD)
# Rama de trabajo     : feature/us-08-gestion-usuarios
# Trazabilidad        : Issue #39 -> SDD -> Criterios -> Escenarios BDD -> Tests Go -> Codigo
#
# Entorno: BASE_URL http://localhost:$PORT, PostgreSQL 16
# Ejemplo httptest en internal/handler/*_test.go
# -----------------------------------------------------------------------------

Feature: Gestión independiente de usuarios y asociación con roles en proyectos

  Contexto:
    Dado que el sistema está listo para recibir solicitudes

  @caso_normal
  Escenario: E-01 Registro exitoso de un usuario independiente
    Dado que el sistema no tiene registrado al usuario "dev@example.com"
    Cuando hago un POST a "/api/usuarios" con nombre "Ana Pérez" y email "dev@example.com"
    Entonces el código de respuesta debe ser 201 Created
    Y el usuario debe quedar guardado en el sistema

  @caso_error
  Escenario: E-02 Intento de registro con email duplicado
    Dado que el usuario "dev@example.com" ya está registrado en el sistema
    Cuando intento registrar otro usuario con el mismo email "dev@example.com"
    Entonces el código de respuesta debe ser 422 Unprocessable Entity
    Y el sistema rechaza el registro por email duplicado

  @caso_normal
  Escenario: E-03 Listado de usuarios registrados
    Dado que existen usuarios registrados en el sistema
    Cuando hago un GET a "/api/usuarios"
    Entonces el código de respuesta debe ser 200 OK
    Y la respuesta debe incluir la lista de usuarios

  @caso_normal
  Escenario: E-04 Creación de proyecto sin usuarios anidados
    Cuando hago un POST a "/api/proyectos" con nombre "Nuevo Proyecto" y sus fechas
    Entonces el código de respuesta debe ser 201 Created
    Y el proyecto se crea sin miembros asociados inicialmente

  @caso_normal
  Escenario: E-05 Asociación exitosa de un usuario existente con rol al proyecto
    Dado que existe el proyecto 1
    Y que existe el usuario 5 registrado
    Cuando hago un POST a "/api/proyectos/1/usuarios" con usuario_id 5 y rol "Constructor de Producto"
    Entonces el código de respuesta debe ser 200 OK
    Y el usuario queda vinculado al proyecto con su respectivo rol en la tabla puente
    Y la respuesta debe incluir el rol efectivo

  @caso_error
  Escenario: E-06 Intento de asociar un usuario que no existe
    Dado que el proyecto 1 existe
    Y que el usuario 999 no está registrado
    Cuando hago un POST a "/api/proyectos/1/usuarios" con usuario_id 999 y rol "Arquitecto de Producto"
    Entonces el código de respuesta debe ser 404 Not Found
    Y el sistema indica que el usuario no existe

  @caso_error
  Escenario: E-07 Asociación con rol inválido
    Dado que existe el proyecto 1
    Y que existe el usuario 5 registrado
    Cuando hago un POST a "/api/proyectos/1/usuarios" con usuario_id 5 y rol "Scrum Master"
    Entonces el código de respuesta debe ser 422 Unprocessable Entity
    Y el error debe indicar rol inválido

  @caso_error
  Escenario: E-08 Asociación con rol ausente
    Dado que existe el proyecto 1
    Y que existe el usuario 5 registrado
    Cuando hago un POST a "/api/proyectos/1/usuarios" sin informar el rol
    Entonces el código de respuesta debe ser 422 Unprocessable Entity
    Y el error debe indicar rol inválido
