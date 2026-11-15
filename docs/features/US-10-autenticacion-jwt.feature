# language: es
# =============================================================================
# US-10 - Autenticacion de usuarios con JWT
# Escenarios BDD (Gherkin)
#
# Issue de origen     : SebastianAguilera/ingenieria-y-calidad#61
# Especificacion SDD  : docs/specs/US-10-autenticacion-jwt.md
# Metodologia         : docs/methodology.md seccion 4 (BDD con archivos .feature)
# Rama de trabajo     : feature/us-10-autenticacion-jwt
# Trazabilidad        : Issue #61 -> SDD -> Criterios -> Escenarios BDD -> Tests Go -> Codigo
#
# Convenciones de lectura (dialecto, notacion, etiquetas y entorno):
# ver docs/methodology.md, seccion "BDD (Behavior-Driven Development)".
#
# NOTA: especificacion ejecutable futura. El repo no tiene runner de Godog
# (methodology.md seccion 4): cada escenario se automatiza como subtest
# t.Run("E-nn", ...) en los *_test.go indicados al final del archivo.
#
# Entorno: BASE_URL http://localhost:$PORT, PostgreSQL 16, .env con JWT_SECRET
# (obligatoria) y JWT_EXPIRY (default 24h, formato time.ParseDuration).
# Verificaciones HTTP con curl o httptest; dobles de servicio con testify.
# -----------------------------------------------------------------------------
# NARRATIVA (US-10, issue #61)
# -----------------------------------------------------------------------------
# Como usuario, quiero iniciar sesion con mi email y contrasena para que solo
# yo pueda acceder a los proyectos y datos del sistema.
#
# El login emite un JWT (HS256, firma con JWT_SECRET) y un middleware
# default-deny protege todas las rutas /api/*: las unicas publicas son
# GET /health y POST /api/auth/login. La contrasena se persiste solo como
# hash bcrypt y ningun JSON expone password_hash. No hay logout server-side
# ni blacklist: el cliente borra su token y un token valido vive hasta exp.
#
# -----------------------------------------------------------------------------
# CRITERIOS DE ACEPTACION DE LA ISSUE #61 / SDD seccion 8
# -----------------------------------------------------------------------------
# [CA-1] Login correcto -> 200 con token (HS256, sin expirar) y usuario
#        sin password_hash.                                     -> E-01, E-02
# [CA-2] Contrasena incorrecta -> 401 CREDENCIALES_INVALIDAS sin token.
#                                                                 -> E-03
# [CA-3] Email inexistente -> 401 con exactamente el mismo mensaje
#        que CA-2 (no revela existencia).                   -> E-04, E-05
# [CA-4] Ruta /api/* protegida sin Authorization -> 401
#        TOKEN_REQUERIDO (no llega al servicio).                 -> E-11
# [CA-5] Token con firma ajena o expirado -> 401 TOKEN_INVALIDO.
#                                                                 -> E-13
# [CA-6] Token valido del login + GET /api/proyectos -> 200
#        con el listado normal.                                 -> E-10
# [CA-7] En DB solo hash bcrypt; ningun JSON contiene
#        password_hash.                                  -> E-01, E-17, E-18
#
# -----------------------------------------------------------------------------
# REGLAS DE NEGOCIO CUBIERTAS (SDD seccion 4)
# -----------------------------------------------------------------------------
# R-01 contrasena solo hash bcrypt, comparacion con
#      bcrypt.CompareHashAndPassword (nunca ==)                 -> E-17
# R-02 PasswordHash con json:"-"; ningun endpoint lo devuelve  -> E-18
# R-03 email normalizado (trim + minusculas) antes de buscar   -> E-02
# R-04 email inexistente, sin hash o contrasena mala -> el MISMO
#      401 CREDENCIALES_INVALIDAS (anti user-enumeration)
#                                      -> E-03 E-04 E-05 E-06
# R-05 JWT HS256 fijo, claims sub/email/iat/exp = iat+JWT_EXPIRY
#                                                  -> E-01, E-15
# R-06 middleware default-deny: sin header/sin prefijo ->
#      TOKEN_REQUERIDO; con "Bearer " pero fallo -> TOKEN_INVALIDO
#                     -> E-10 E-11 E-12 E-13 E-14
# R-07 sin logout server-side ni blacklist; el cliente borra
#      su token (vive hasta exp)                                -> E-16
# R-08 orden de errores en login: JSON (400) -> campos (422) ->
#      credenciales (401) -> exito (200)                  -> E-07 E-08 E-09
# R-09 usuarios sin hash (p. ej. seba pre-backfill) -> 401      -> E-05
#
# -----------------------------------------------------------------------------
# CASOS LIMITE CUBIERTOS (SDD seccion 6; numeracion asignada aqui: la spec
# no los numeraba)
# -----------------------------------------------------------------------------
# CL-01 email con mayusculas/espacios (" Seba@Example.COM ")
#                                               -> E-02
# CL-02 email:"" / password:"" -> 422 con el campo, sin llegar a bcrypt
#                                               -> E-07
# CL-03 usuario sin password_hash              -> E-05
# CL-04 token con otra firma / expirado / alg:none -> 401 TOKEN_INVALIDO
#                                               -> E-13
# CL-05 header ausente o sin "Bearer " -> TOKEN_REQUERIDO;
#      "Bearer basura" -> TOKEN_INVALIDO       -> E-11 E-12 E-13
# CL-06 hash corrupto en DB -> 401 (detalle solo en logs)      -> E-06
# CL-07 JWT_SECRET vacia o JWT_EXPIRY invalido -> no arranca    -> E-19
# CL-08 dos sesiones simultaneas validas hasta su exp (sin
#      blacklist)                                                -> E-16
# CL-09 /health y login sin token responden normal, nunca 401   -> E-14
#
# -----------------------------------------------------------------------------
# ENDPOINTS CUBIERTOS
# -----------------------------------------------------------------------------
#   1. POST /api/auth/login   (publica)  -> E-01 a E-09, E-14, E-15, E-18
#   2. GET  /health           (publica)  -> E-14
#   3. GET  /api/proyectos    (protegida) -> E-10 a E-13, E-16
#   4. GET  /api/usuarios, GET /api/usuarios/{id} (protegidas) -> E-18
# =============================================================================

Funcionalidad: US-10 - Autenticacion de usuarios con JWT
  Como usuario del sistema
  Quiero iniciar sesion con mi email y contrasena para recibir un token
  que me permita acceder solo a mis proyectos y datos
  Para que ningun otro pueda consultar ni modificar la informacion del
  equipo sin una sesion valida.

  Solo la contrasena en claro del request de login se compara contra el
  hash bcrypt persistido: nunca se guarda, se loguea ni se refleja en una
  respuesta. El middleware es default-deny sobre /api/* con una lista
  publica explicita (GET /health y POST /api/auth/login); un token valido
  se emite con HS256, firma unica con JWT_SECRET y expira segun JWT_EXPIRY
  (24h por defecto). Sin logout server-side: el cliente solo borra su token.

  Contexto:
    Dado que la API esta levantada y responde "200 OK" en "GET /health"
    Y que la base de datos fue migrada con la columna "password_hash"
    Y que existe un usuario con email "ana@example.com" y contrasena
      "Secreto123" guardada solo como hash bcrypt

  # ===========================================================================
  # LOGIN: POST /api/auth/login
  # ===========================================================================

  Regla: El login emite un JWT con credenciales validas

    @caso_normal
    Escenario: E-01 - Login con credenciales validas devuelve token y usuario
      # CA-1, CA-7 y R-05: 200 con JWT HS256 aun sin expirar y con el
      # usuario sin password_hash (R-01: la comparacion es bcrypt).
      Dado que existe un usuario con email "ana@example.com" y contrasena
        "Secreto123" guardada solo como hash bcrypt
      Cuando envio "POST" a "/api/auth/login" con email "ana@example.com"
        y password "Secreto123"
      Entonces la respuesta tiene el estado HTTP "200"
      Y el cuerpo contiene un campo "token" firmado con HS256 y sin expirar
      Y el campo "usuario" tiene "id", "nombre", "email" y "creado_en"
      Y el cuerpo no contiene "password_hash"

    @caso_limite
    Escenario: E-02 - Email con mayusculas y espacios se normaliza antes de buscar
      # CA-1, R-03 y CL-01: trim + minusculas (domain.NormalizarEmail)
      # antes de ObtenerPorEmail; el resultado es el mismo que E-01.
      Dado que existe un usuario con email "ana@example.com" y contrasena
        "Secreto123"
      Cuando envio "POST" a "/api/auth/login" con email
        "  ANA@EXAMPLE.COM  " y password "Secreto123"
      Entonces la respuesta tiene el estado HTTP "200"
      Y el cuerpo contiene un campo "token"
      Y el email del usuario devuelto es "ana@example.com"

  Regla: Las credenciales invalidas responden un unico error generico

    @caso_error
    Escenario: E-03 - Contrasena incorrecta responde 401 CREDENCIALES_INVALIDAS
      # CA-2 y R-04: 401 con error CREDENCIALES_INVALIDAS, mensaje generico
      # "email o contrasena incorrectos" y sin campo "token".
      Dado que existe un usuario con email "ana@example.com" y contrasena
        "Secreto123"
      Cuando envio "POST" a "/api/auth/login" con email "ana@example.com"
        y password "OtraClave"
      Entonces la respuesta tiene el estado HTTP "401"
      Y el cuerpo contiene el error "CREDENCIALES_INVALIDAS" con mensaje
        "email o contrasena incorrectos"
      Y el cuerpo no contiene un campo "token"

    @caso_error
    Escenario: E-04 - Email inexistente responde exactamente el mismo 401
      # CA-3 y R-04 (anti user-enumeration): el mensaje debe ser identico
      # al del escenario E-03; en login no se usa ErrUsuarioNoEncontrado.
      Dado que no existe un usuario con email "nadie@example.com"
      Cuando envio "POST" a "/api/auth/login" con email "nadie@example.com"
        y password "Secreto123"
      Entonces la respuesta tiene el estado HTTP "401"
      Y el cuerpo contiene el error "CREDENCIALES_INVALIDAS"
      Y el mensaje es identico al del login con contrasena incorrecta (E-03)

    @caso_limite
    Escenario: E-05 - Usuario sin hash de contrasena responde el mismo 401
      # R-09, R-04 y CL-03: usuarios sin password_hash (p. ej. "seba"
      # antes del backfill) no llegan a bcrypt y devuelven el mismo 401.
      Dado que existe el usuario "seba@example.com" sin password_hash
        en la base
      Cuando envio "POST" a "/api/auth/login" con email "seba@example.com"
        y password "Secreto123"
      Entonces la respuesta tiene el estado HTTP "401"
      Y el cuerpo contiene el error "CREDENCIALES_INVALIDAS" con mensaje
        identico al de E-03

    @caso_error
    Escenario: E-06 - Hash corrupto en la base responde 401 sin filtrar detalle
      # R-04 y CL-06: bcrypt.CompareHashAndPassword falla; el detalle solo
      # se registra en logs y la respuesta sigue siendo el 401 generico.
      Dado que el usuario "ana@example.com" tiene un password_hash
        corrupto en la base
      Cuando envio "POST" a "/api/auth/login" con email "ana@example.com"
        y password "Secreto123"
      Entonces la respuesta tiene el estado HTTP "401"
      Y el cuerpo contiene el error "CREDENCIALES_INVALIDAS" con mensaje
        identico al de E-03
      Y el cuerpo no revela el detalle del error interno

  Regla: Validacion de entrada y orden de errores del login

    @caso_limite
    Esquema del Escenario: E-07 - Campo obligatorio <campo> vacio responde 422
      # R-08 y CL-02: el campo vacio se rechaza antes de llegar a bcrypt
      # (no se compara ningun hash) y el cuerpo indica el campo exacto.
      Dado que existe un usuario con credenciales validas
      Cuando envio "POST" a "/api/auth/login" con email "<email_ejemplo>"
        y password "<password_ejemplo>"
      Entonces la respuesta tiene el estado HTTP "422"
      Y el cuerpo contiene el error "VALIDACION" para el campo "<campo>"
      Y la contrasena no se compara contra ningun hash

      Ejemplos:
        | campo    | email_ejemplo   | password_ejemplo |
        | email    |                 | Secreto123       |
        | password | ana@example.com |                  |

    @caso_error
    Escenario: E-08 - Body que no es JSON responde 400 JSON_INVALIDO
      # R-08 y el patrón existente del repo: el parseo rechaza el body
      # antes de validar campos o credenciales.
      Dado que existe un usuario con credenciales validas
      Cuando envio "POST" a "/api/auth/login" con el cuerpo
        "email=ana@example.com&password=Secreto123"
      Entonces la respuesta tiene el estado HTTP "400"
      Y el cuerpo contiene el error "JSON_INVALIDO"

    @caso_alternativo
    Esquema del Escenario: E-09 - Precedencia de errores ante <problema>
      # R-08: el orden de validacion es JSON (400) -> campos (422) ->
      # credenciales (401) -> exito (200); manda el primer problema
      # detectado. Cada fila arma la solicitud concreta en el test.
      Dado que preparo la solicitud de login con <problema>
      Cuando envio "POST" a "/api/auth/login" con ese cuerpo
      Entonces la respuesta tiene el estado HTTP "<status_esperado>"
      Y el cuerpo contiene el error "<error_esperado>"

      Ejemplos:
        | problema                                  | status_esperado | error_esperado |
        | un body que no es JSON y campos vacios    | 400             | JSON_INVALIDO  |
        | campos vacios y una contrasena incorrecta | 422             | VALIDACION     |

  # ===========================================================================
  # PROTECCION DE RUTAS: middleware sobre /api/*
  # ===========================================================================

  Regla: Ruta protegida con token valido

    @caso_normal
    Escenario: E-10 - Token del login permite consultar el listado de proyectos
      # CA-6 y R-06: con el Bearer valido el middleware deja pasar la
      # solicitud y la ruta responde con su listado normal.
      Dado que tengo un token valido obtenido con el login de
        "ana@example.com"
      Cuando envio "GET" a "/api/proyectos" con el token en el header
        "Authorization" con prefijo Bearer
      Entonces la respuesta tiene el estado HTTP "200"
      Y el cuerpo contiene el listado normal de proyectos

  Regla: Rutas protegidas sin credenciales de acceso

    @caso_error
    Escenario: E-11 - Ruta protegida sin el header Authorization
      # CA-4, R-06 y CL-05: default-deny en /api/*; sin header la
      # solicitud se rechaza en el middleware y no llega al servicio.
      Dado que existe una ruta protegida "GET /api/proyectos"
      Cuando envio "GET" a "/api/proyectos" sin header "Authorization"
      Entonces la respuesta tiene el estado HTTP "401"
      Y el cuerpo contiene el error "TOKEN_REQUERIDO"
      Y la logica de listado de proyectos no se ejecuta

    @caso_limite
    Escenario: E-12 - Header Authorization presente pero sin el prefijo Bearer
      # R-06 y CL-05: sin "Bearer " el resultado es TOKEN_REQUERIDO,
      # identico al header ausente (E-11); el token ni siquiera se parsea.
      Dado que tengo un token valido obtenido con el login de
        "ana@example.com"
      Cuando envio "GET" a "/api/proyectos" con header "Authorization"
        usando el token sin el prefijo Bearer
      Entonces la respuesta tiene el estado HTTP "401"
      Y el cuerpo contiene el error "TOKEN_REQUERIDO"

    @caso_error
    Esquema del Escenario: E-13 - Token rechazado responde 401 TOKEN_INVALIDO con <variante>
      # CA-5, R-06 y CL-04/CL-05: con el prefijo "Bearer " presente pero
      # el token malformado, firmado con otra clave, expirado o con otro
      # alg, siempre es TOKEN_INVALIDO. "Bearer basura" entra aqui y no
      # en TOKEN_REQUERIDO. Los tokens de ejemplo se fabrican en el test.
      Dado que existe una ruta protegida "GET /api/proyectos"
      Cuando envio "GET" a "/api/proyectos" con un token <variante>
      Entonces la respuesta tiene el estado HTTP "401"
      Y el cuerpo contiene el error "TOKEN_INVALIDO"
      Y el cuerpo no revela el detalle interno de la falla

      Ejemplos:
        | variante                                |
        | firmado con otra clave                  |
        | expirado                                |
        | con alg "none"                          |
        | con alg distinto de HS256               |
        | que no es un JWT ("Bearer basura")      |

  Regla: Rutas publicas sin token

    @caso_normal
    Escenario: E-14 - /health y login responden normal sin Authorization
      # R-06 y CL-09: solo GET /health y POST /api/auth/login son
      # publicas; responden normal y nunca 401 por falta de token.
      Dado que la API esta levantada
      Cuando envio "GET" a "/health" sin header "Authorization"
      Entonces la respuesta tiene el estado HTTP "200"
      Y cuando envio "POST" a "/api/auth/login" con credenciales validas
        sin header "Authorization"
      Entonces la respuesta tiene el estado HTTP "200"

  # ===========================================================================
  # CLAIMS Y VIGENCIA DEL TOKEN
  # ===========================================================================

  Regla: El token emitido cumple el contrato de claims

    @caso_alternativo
    Escenario: E-15 - Claims del token emitido
      # R-05: HS256 fijo, sub = id del usuario en string, email, iat y
      # exp = iat + JWT_EXPIRY (default 24h); verificacion round-trip
      # con el mismo secreto que usa la aplicacion.
      Dado que hago login con credenciales validas
      Cuando descifro los claims del token emitido con el secreto
        de configuracion
      Entonces el algoritmo del token es "HS256"
      Y el claim "sub" es el id del usuario en formato texto
      Y el claim "email" es "ana@example.com"
      Y el claim "exp" es "iat" mas la duracion configurada (24h por
        defecto)

  Regla: Sin logout server-side ni blacklist

    @caso_limite
    Escenario: E-16 - Dos sesiones simultaneas validas hasta su expiracion
      # R-07 y CL-08: no hay revocacion en el backend; el cliente solo
      # borra su token y los dos tokens emitidos siguen validos hasta exp.
      Dado que hago login dos veces con "ana@example.com" y obtengo dos
        tokens
      Cuando consulto "GET /api/proyectos" primero con un token y luego
        con el otro
      Entonces ambas solicitudes responden "200"
      Y no existe ningun registro de revocacion en el backend

  # ===========================================================================
  # LA CONTRASENA SOLO VIVE COMO HASH (CA-7)
  # ===========================================================================

  Regla: La contrasena persistida es irrecuperable

    @caso_normal
    Escenario: E-17 - En la base solo hay hash bcrypt, nunca la contrasena en claro
      # CA-7 y R-01: tras el alta la columna password_hash guarda un hash
      # bcrypt (varchar(72)) y la contrasena en claro no se persiste ni
      # se escribe en los logs de la aplicacion.
      Dado que creo un usuario con contrasena "Secreto123"
      Cuando inspecciono la columna "password_hash" en la base de datos
      Entonces el valor es un hash bcrypt valido para "Secreto123"
      Y el valor no contiene la contrasena en claro
      Y la contrasena en claro no aparece en los logs de la aplicacion

    @caso_alternativo
    Escenario: E-18 - Ningun endpoint devuelve password_hash
      # CA-7 y R-02: PasswordHash tiene json:"-"; se verifican el login,
      # el listado y el detalle de usuarios (el resto de endpoints
      # hereda la misma serializacion de domain.Usuario).
      Dado que existe un usuario con hash bcrypt en la base
      Cuando consulto "POST /api/auth/login", "GET /api/usuarios" y
        "GET /api/usuarios/1"
      Entonces ninguna respuesta contiene la cadena "password_hash"
      Y ninguna respuesta contiene el hash bcrypt almacenado

  # ===========================================================================
  # CONFIGURACION OBLIGATORIA
  # ===========================================================================

  Regla: Sin secret no hay arranque

    @caso_error
    Esquema del Escenario: E-19 - La aplicacion no arranca con <configuracion>
      # SDD secciones 2 y 3, CL-07: config.Load falla y el proceso no
      # levanta. NOTA: este escenario no corre contra el Contexto de la
      # API levantada; se automatiza con config_test o go run ./cmd/api
      # en un entorno de configuracion aislado.
      Dado un entorno de configuracion con <configuracion>
      Cuando intento arrancar la aplicacion
      Entonces la aplicacion no inicia
      Y el error de configuracion se reporta al arranque

      Ejemplos:
        | configuracion                       |
        | JWT_SECRET ausente                  |
        | JWT_SECRET vacia                    |
        | JWT_EXPIRY con formato invalido     |

# -----------------------------------------------------------------------------
# MAPA DE TRAZABILIDAD (resumen)
# -----------------------------------------------------------------------------
# Criterio de aceptacion -> escenario
#   CA-1 E-01 E-02        CA-2 E-03        CA-3 E-04 E-05
#   CA-4 E-11             CA-5 E-13        CA-6 E-10
#   CA-7 E-01 E-17 E-18
#
# Regla de negocio -> escenario
#   R-01 E-17   R-02 E-18   R-03 E-02   R-04 E-03 E-04 E-05 E-06
#   R-05 E-01 E-15   R-06 E-10 E-11 E-12 E-13 E-14   R-07 E-16
#   R-08 E-07 E-08 E-09   R-09 E-05
#
# Caso limite (SDD seccion 6) -> escenario
#   CL-01 E-02   CL-02 E-07   CL-03 E-05   CL-04 E-13
#   CL-05 E-11 E-12 E-13   CL-06 E-06   CL-07 E-19   CL-08 E-16
#   CL-09 E-14
#
# Condicion de error (SDD seccion 7) -> escenario
#   ErrCredencialesInvalidas 401   E-03 E-04 E-05 E-06
#   ErrPasswordVacio / VALIDACION 422   E-07 E-09
#   ErrTokenRequerido 401          E-11 E-12
#   ErrTokenInvalido 401           E-13
#   JSON_INVALIDO 400              E-08 E-09
#   config.Load falla (sin arranque) E-19
#
# -----------------------------------------------------------------------------
# SUGERENCIA DE AUTOMATIZACION (TDD: RED -> GREEN)
# -----------------------------------------------------------------------------
# Los pasos de este archivo se replican como subtests t.Run() con el
# identificador E-nn como nombre, para que el reporte de go test y el
# reporte de BDD sean comparables (methodology.md seccion 4):
#   internal/config/config_test.go              E-19 (JWT_SECRET obligatoria,
#                                             JWT_EXPIRY default e invalido)
#   internal/domain/usuario_test.go             E-18 (PasswordHash json:"-")
#   internal/domain/errores_test.go             mapeo de los errores sentinelas
#   internal/service/auth_service_test.go       E-01 E-02 E-03 E-04 E-05
#                                             E-06 E-15 E-17 (CA-1 a CA-3,
#                                             CA-7, claims round-trip)
#   internal/handler/auth_handler_test.go       E-07 E-08 E-09 (422/400/401)
#   internal/handler/middleware_auth_test.go    E-10 E-11 E-12 E-13 (CA-4 a
#                                             CA-6, variantes de header)
#   cmd/api/router_test.go                      E-14 (lista publica explicita),
#                                             E-10 (wiring con token)
#   test de serializacion (handler o domain)    E-18 (ningun JSON con
#                                             password_hash, CA-7)
# Los mocks siguen el estilo de service/mocks_test.go y las pruebas usan
# t.Parallel() con testify. Los tokens invalidos de E-13 se fabrican en el
# test (clave ajena, expiracion vencida, alg none/RS256, cadena "basura").
