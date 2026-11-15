# language: es
# =============================================================================
# US-F02 - Ingreso al panel con un usuario existente
# Escenarios BDD (Gherkin)
#
# Issue de origen     : SebastianAguilera/ingenieria-y-calidad#51
# Especificacion SDD  : docs/specs/US-F02-ingreso-panel-spec.md
# Metodologia         : docs/methodology.md seccion 4 (BDD)
# Rama de trabajo     : feature/US-F02-ingreso-con-usuario-existente
# Trazabilidad        : Issue #51 -> SDD -> Criterios -> Escenarios BDD -> Tests Vitest -> Codigo
#
# HISTORICO - REEMPLAZADO POR US-F08:
#   Estos escenarios describen el login simulado (busqueda por nombre/email e
#   registro inline). La pantalla de login vigente esta definida en
#   docs/features/US-F08-login-real-jwt.feature (issue #62). Se conservan los
#   escenarios E-01 a E-03 y E-05 de proteccion/persistencia que US-F08 reutiliza;
#   E-04, E-07 a E-10 y E-12 pasan a US-F09 (issue #65, registro con contraseña).
#
# Entorno: frontend Vite http://localhost:5173 proxying /api -> http://localhost:8080
# Tests: frontend/src/**/*.test.ts(x) con Vitest + React Testing Library
# -----------------------------------------------------------------------------

Feature: Ingreso al panel con un usuario existente

  Contexto:
    Dado que el backend expone GET /api/usuarios y POST /api/usuarios

  # ---------------------------------------------------------------------------
  # Protección de rutas
  # ---------------------------------------------------------------------------

  @caso_normal
  Escenario: E-01 Acceso al panel sin sesión iniciada
    Dado que no hay usuario activo en la sesión
    Cuando intento abrir "/proyectos"
    Entonces soy redirigido a "/login"

  @caso_normal
  Escenario: E-02 Acceso al panel con sesión iniciada
    Dado que hay un usuario activo en la sesión
    Cuando abro "/proyectos"
    Entonces veo la vista del panel de proyectos

  # ---------------------------------------------------------------------------
  # Ingreso
  # ---------------------------------------------------------------------------

  @caso_normal
  Escenario: E-03 Ingreso con un email existente
    Dado que existe un usuario con email "ana@example.com"
    Cuando ingreso "ana@example.com" en "/login"
    Entonces se guarda el usuario activo
    Y soy redirigido a "/proyectos"

  @caso_normal
  Escenario: E-04 Ingreso con un nombre que identifica a un único usuario
    Dado que existe un único usuario llamado "Ana Pérez"
    Cuando ingreso "Ana Pérez" en "/login"
    Entonces se guarda el usuario activo
    Y soy redirigido a "/proyectos"

  @caso_normal
  Escenario: E-05 Ingreso normalizando mayúsculas y espacios
    Dado que existe un usuario con email "ana@example.com"
    Cuando ingreso "  ANA@EXAMPLE.COM  " en "/login"
    Entonces se encuentra al mismo usuario
    Y soy redirigido a "/proyectos"

  @caso_limite
  Escenario: E-06 Intento de ingreso con el campo vacío
    Dado que estoy en "/login"
    Cuando intento ingresar sin escribir nada
    Entonces veo el mensaje "Ingresá tu email"
    Y no se realiza ninguna búsqueda en la API

  # ---------------------------------------------------------------------------
  # Usuario inexistente y registro
  # ---------------------------------------------------------------------------

  @caso_alternativo
  Escenario: E-07 Ingreso con un email que no existe
    Dado que no existe un usuario con email "nuevo@example.com"
    Cuando ingreso "nuevo@example.com" en "/login"
    Entonces veo la tarjeta "Usuario no encontrado" con un formulario de registro

  @caso_normal
  Escenario: E-08 Registro exitoso de un usuario nuevo
    Dado que veo la tarjeta de registro con el email "nuevo@example.com"
    Cuando completo el nombre "Nuevo Usuario" y confirmo el registro
    Entonces la API crea el usuario con POST /api/usuarios
    Y se guarda el usuario creado como usuario activo
    Y soy redirigido a "/proyectos"

  @caso_error
  Escenario: E-09 Registro con un email ya existente
    Dado que el email "ana@example.com" ya está registrado
    Cuando intento registrar un usuario nuevo con ese email
    Entonces veo el mensaje "Ese email ya está registrado"
    Y no se inicia sesión

  @caso_error
  Escenario: E-10 Registro con datos inválidos
    Dado que veo la tarjeta de registro
    Cuando envío un nombre vacío o un email inválido
    Entonces veo el mensaje de validación correspondiente
    Y no se inicia sesión

  @caso_error
  Escenario: E-11 Error de red durante el ingreso o el registro
    Dado que el backend no responde
    Cuando intento ingresar o registrar un usuario
    Entonces veo un mensaje de error de conexión
    Y la pantalla sigue disponible para reintentar

  # ---------------------------------------------------------------------------
  # Ambigüedad
  # ---------------------------------------------------------------------------

  @caso_alternativo
  Escenario: E-12 Nombre compartido por más de un usuario
    Dado que existe más de un usuario llamado "Ana Pérez"
    Cuando ingreso "Ana Pérez" en "/login"
    Entonces veo el mensaje "Hay más de un usuario con ese nombre. Ingresá tu email."
    Y no se inicia sesión

  # ---------------------------------------------------------------------------
  # Persistencia y salida
  # ---------------------------------------------------------------------------

  @caso_normal
  Escenario: E-13 Persistencia de la sesión al recargar
    Dado que ingresé correctamente
    Cuando recargo la página
    Entonces sigo con la sesión iniciada en "/proyectos"

  @caso_normal
  Escenario: E-14 Salida del panel
    Dado que estoy dentro del panel con sesión iniciada
    Cuando presiono "Salir"
    Entonces se limpia el usuario activo
    Y vuelvo a "/login"

  @caso_limite
  Escenario: E-15 Sesión corrupta en el almacenamiento
    Dado que localStorage contiene datos de sesión inválidos
    Cuando abro el panel
    Entonces la sesión se ignora
    Y soy redirigido a "/login"
