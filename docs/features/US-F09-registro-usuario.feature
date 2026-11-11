# language: es
# =============================================================================
# US-F09 - Registro de usuario con contraseña desde el login
# Escenarios BDD (Gherkin)
#
# Issue de origen     : SebastianAguilera/ingenieria-y-calidad#65
# Especificacion SDD  : docs/specs/US-F09-registro-usuario-spec.md
# Metodologia         : docs/methodology.md seccion 4 (BDD)
# Rama de trabajo     : feature/us-f09-registro-usuario
# Trazabilidad        : Issue #65 -> SDD -> Criterios -> Escenarios BDD -> Tests Go/Vitest -> Codigo
#
# Entorno: frontend Vite http://localhost:5173 proxying /api -> http://localhost:8080
# Tests  : internal/service, internal/handler y cmd/api (Go) + frontend/src/**/*.test.ts(x)
# -----------------------------------------------------------------------------

Feature: Registro de usuario con contraseña desde el login

  Contexto:
    Dado que el backend expone POST /api/auth/registro como ruta pública
    Y que en /login hay un link "Crear cuenta" visible

  # ---------------------------------------------------------------------------
  # Punto de entrada y alta
  # ---------------------------------------------------------------------------

  @caso_normal
  Escenario: E-01 El link de registro siempre está visible
    Dado que estoy en "/login"
    Cuando miro el formulario de ingreso
    Entonces veo el link "Crear cuenta"
    Y no se realiza ninguna petición a la API solo por mostrarlo

  @caso_normal
  Escenario: E-02 Registro completo y auto-login
    Dado que estoy en "/login" y abro "Crear cuenta"
    Cuando ingreso el nombre "Ana Pérez", el email "ana@example.com" y la contraseña "Secreto123"
    Y envío el alta
    Entonces la API responde 200 con un token y el usuario
    Y se guardan el token y el usuario en la sesión
    Y soy redirigido a "/proyectos"

  @caso_limite
  Escenario: E-03 Campos vacíos
    Dado que estoy en la vista de registro
    Cuando intento registrar dejando un campo sin completar
    Entonces veo el mensaje de validación en el campo correspondiente
    Y no se realiza ninguna petición a la API

  # ---------------------------------------------------------------------------
  # Errores del backend
  # ---------------------------------------------------------------------------

  @caso_error
  Escenario: E-04 Email ya registrado
    Dado que existe un usuario con email "ana@example.com"
    Cuando registro otro usuario con ese mismo email
    Y la API responde 422 USUARIO_DUPLICADO con el campo "email"
    Entonces veo el mensaje "Ese email ya está registrado. Volvé a ingresar."
    Y no se guarda ningún token
    Y sigo en la vista de registro

  @caso_error
  Escenario: E-05 Email con formato inválido
    Dado que estoy en la vista de registro
    Cuando envío un email sin formato válido con el resto de los campos completos
    Y la API responde 422 VALIDACION con el campo "email"
    Entonces veo el mensaje devuelto por el backend
    Y el campo de email queda resaltado
    Y no se guarda ningún token

  @caso_error
  Escenario: E-08 Servidor sin respuesta
    Dado que el backend no responde
    Cuando envío el alta con nombre, email y contraseña
    Entonces veo el mensaje "No se pudo conectar con el servidor."
    Y puedo reintentar sin perder lo escrito

  # ---------------------------------------------------------------------------
  # Persistencia y salida
  # ---------------------------------------------------------------------------

  @caso_normal
  Escenario: E-06 El usuario creado puede iniciar sesión
    Dado que registré al usuario "ana@example.com" con la contraseña "Secreto123"
    Cuando vuelvo al login y ingreso con esas credenciales
    Entonces la API responde 200 con token y usuario
    Y soy redirigido a "/proyectos"

  @caso_normal
  Escenario: E-07 Volver al login sin registrar
    Dado que estoy en la vista de registro
    Cuando presiono "Volver"
    Entonces vuelvo al formulario de ingreso
    Y no se creó ningún usuario

  # ---------------------------------------------------------------------------
  # Contrato del endpoint y límites
  # ---------------------------------------------------------------------------

  @caso_normal
  Escenario: E-09 El endpoint es público
    Dado que no tengo token de sesión
    Cuando envío un registro válido sin el header Authorization
    Entonces la API responde 200 (nunca 401 TOKEN_REQUERIDO)

  @caso_limite
  Escenario: E-10 Nombre demasiado largo
    Dado que estoy en la vista de registro
    Cuando envío un nombre de más de 100 caracteres
    Y la API responde 422 VALIDACION con el campo "nombre"
    Entonces veo el mensaje del backend con el campo "nombre" resaltado
    Y no se creó ningún usuario

  @caso_limite
  Escenario: E-11 Email con espacios y mayúsculas
    Dado que estoy en la vista de registro
    Cuando envío el email "  ANA@Example.com  "
    Y el backend lo normaliza y no había usuario con ese email
    Entonces el usuario se crea con email "ana@example.com"
    Y la sesión inicia en "/proyectos"

  @caso_limite
  Escenario: E-12 La contraseña no se guarda en claro
    Dado que registré un usuario con contraseña
    Cuando inspecciono la respuesta y el almacenamiento del navegador
    Entonces la respuesta incluye token y usuario sin password_hash
    Y la contraseña no aparece en el almacenamiento del navegador

  @caso_limite
  Escenario: E-13 Doble envío del formulario
    Dado que estoy en la vista de registro con todos los campos completos
    Cuando presiono "Crear cuenta" dos veces seguidas
    Entonces se envía una sola petición
    Y el botón queda deshabilitado mientras la respuesta llega
