# language: es
# =============================================================================
# US-F08 - Login real con email y contraseña (JWT)
# Escenarios BDD (Gherkin)
#
# Issue de origen     : SebastianAguilera/ingenieria-y-calidad#62
# Especificacion SDD  : docs/specs/US-F08-login-real-jwt-spec.md
# Metodologia         : docs/methodology.md seccion 4 (BDD)
# Rama de trabajo     : feature/us-f08-login-jwt
# Trazabilidad        : Issue #62 -> SDD -> Criterios -> Escenarios BDD -> Tests Vitest -> Codigo
#
# Reemplaza a         : US-F02 (docs/features/US-F02-ingreso-panel.feature), que queda
#                       como registro historico del login simulado. Los escenarios de
#                       busqueda por nombre y de registro inline (E-04, E-07 a E-10 y
#                       E-12 de F02) se retiran de la pantalla de login y pasan a US-F17.
#
# Entorno: frontend Vite http://localhost:5173 proxying /api -> http://localhost:8080
# Tests: frontend/src/**/*.test.ts(x) con Vitest + React Testing Library
# -----------------------------------------------------------------------------

Feature: Login real con email y contraseña (JWT)

  Contexto:
    Dado que el backend expone POST /api/auth/login como ruta pública
    Y que el resto de las rutas /api/* exigen Authorization Bearer con un JWT vigente

  # ---------------------------------------------------------------------------
  # Protección de rutas
  # ---------------------------------------------------------------------------

  @caso_normal
  Escenario: E-01 Acceso al panel sin token
    Dado que no hay token en la sesión
    Cuando intento abrir "/proyectos"
    Entonces soy redirigido a "/login"

  @caso_normal
  Escenario: E-02 Acceso al panel con un token vigente
    Dado que la sesión contiene un token vigente
    Cuando abro "/proyectos"
    Entonces veo la vista del panel de proyectos

  # ---------------------------------------------------------------------------
  # Ingreso
  # ---------------------------------------------------------------------------

  @caso_normal
  Escenario: E-03 Ingreso con credenciales válidas
    Dado que estoy en "/login"
    Cuando ingreso el email "ana@example.com" y la contraseña "Secreto123"
    Entonces la API responde 200 con un token y el usuario
    Y se guardan el token y el usuario en la sesión
    Y soy redirigido a "/proyectos"

  @caso_error
  Escenario: E-04 Contraseña incorrecta
    Dado que el email "ana@example.com" tiene contraseña "Secreto123"
    Cuando ingreso ese email con la contraseña "otra-clave"
    Entonces veo el mensaje "Email o contraseña incorrectos"
    Y no se guarda ningún token
    Y sigo en "/login"

  @caso_error
  Escenario: E-05 Email inexistente
    Dado que no existe un usuario con email "nadie@example.com"
    Cuando ingreso ese email con cualquier contraseña
    Entonces veo el mensaje "Email o contraseña incorrectos"
    Y el mensaje es idéntico al de una contraseña incorrecta
    Y no se guarda ningún token

  @caso_limite
  Escenario: E-06 Campos vacíos
    Dado que estoy en "/login"
    Cuando intento ingresar sin escribir email o sin escribir contraseña
    Entonces veo el mensaje de validación en el campo correspondiente
    Y no se realiza ninguna petición a la API

  @caso_alternativo
  Escenario: E-07 Validación del backend
    Dado que estoy en "/login"
    Cuando envío un email con formato inválido y una contraseña no vacía
    Y la API responde 422 VALIDACION con el campo "email"
    Entonces veo el mensaje devuelto por el backend
    Y el campo queda resaltado
    Y no se guarda ningún token

  @caso_error
  Escenario: E-08 Servidor sin respuesta
    Dado que el backend no responde
    Cuando intento ingresar con email y contraseña
    Entonces veo el mensaje "No se pudo conectar con el servidor."
    Y puedo reintentar sin perder lo escrito en los campos

  # ---------------------------------------------------------------------------
  # Persistencia y salida
  # ---------------------------------------------------------------------------

  @caso_normal
  Escenario: E-09 Persistencia de la sesión al recargar
    Dado que inicié sesión correctamente
    Cuando recargo la página
    Entonces el token se recupera de localStorage
    Y sigo con la sesión iniciada en "/proyectos"

  @caso_normal
  Escenario: E-10 Salida del panel
    Dado que estoy dentro del panel con sesión iniciada
    Cuando presiono "Salir"
    Entonces se borran el token y el usuario
    Y vuelvo a "/login"

  @caso_error
  Escenario: E-11 Token expirado en cualquier request
    Dado que la sesión contiene un token ya vencido
    Cuando hago cualquier request al panel
    Entonces la API responde 401 TOKEN_INVALIDO
    Y se limpia la sesión
    Y termino en "/login"

  @caso_error
  Escenario: E-12 Token ausente en un request del panel
    Dado que la sesión no tiene token pero el almacenamiento conserva datos de usuario
    Cuando hago cualquier request al panel
    Entonces la API responde 401 TOKEN_REQUERIDO
    Y se limpia la sesión
    Y termino en "/login"

  # ---------------------------------------------------------------------------
  # Límites del almacenamiento y de la entrada
  # ---------------------------------------------------------------------------

  @caso_limite
  Escenario: E-13 Sesión con el formato anterior o corrupta
    Dado que localStorage contiene datos sin token o con JSON inválido
    Cuando abro el panel
    Entonces la sesión se ignora
    Y soy redirigido a "/login"

  @caso_limite
  Escenario: E-14 Email con espacios y mayúsculas
    Dado que existe un usuario con email "ana@example.com"
    Cuando ingreso "  ANA@EXAMPLE.COM  " con su contraseña
    Entonces la petición se envía con el email recortado
    Y el backend lo normaliza y responde 200
    Y soy redirigido a "/proyectos"

  @caso_limite
  Escenario: E-15 Doble envío del formulario
    Dado que estoy en "/login" con email y contraseña completos
    Cuando presiono "Ingresar" dos veces seguidas
    Entonces se envía una sola petición
    Y el botón queda deshabilitado mientras la respuesta llega

  @caso_limite
  Escenario: E-16 La contraseña no se persiste
    Dado que inicié sesión correctamente con una contraseña
    Cuando inspecciono el almacenamiento de la sesión
    Entonces se guardan el token y el usuario
    Y no se guarda la contraseña en ningún lugar del almacenamiento del navegador
