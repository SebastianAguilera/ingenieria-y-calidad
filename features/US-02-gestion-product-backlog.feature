# language: es
# =============================================================================
# US-02 - Gestion basica del Product Backlog
# Escenarios BDD (Gherkin) del modulo de Product Backlog del
# Software Metrics & Estimation Engine.
#
# Issue de origen     : SebastianAguilera/ingenieria-y-calidad#29
# Especificacion SDD  : docs/specs/US-02-gestion-product-backlog-spec.md
# Metodologia         : docs/methodology.md seccion 4 (BDD con archivos .feature)
# Requerimiento       : docs/proyecto.md, requerimiento 2 (Product Backlog)
# Rama de trabajo     : feature/US-02-gestion-product-backlog
# Trazabilidad        : Issue #29 -> SDD -> Criterios de aceptacion -> Escenarios
#                       BDD -> Tests Go (TDD) -> Codigo
#
# -----------------------------------------------------------------------------
# CONVENCIONES DE LECTURA
# -----------------------------------------------------------------------------
# Dialecto: el archivo declara "language: es", por lo que las palabras clave
# Gherkin estan en español y son las equivalentes al ejemplo de
# docs/methodology.md seccion 4:
#
#   Contexto              = Background
#   Escenario             = Scenario
#   Esquema del Escenario = Scenario Outline
#   Ejemplos              = Examples
#   Regla                 = Rule (agrupa por subcomportamiento)
#   Dado / Cuando / Entonces / Y / Pero = Given / When / Then / And / But
#
# Notacion:
#   {id}, {historiaId} = path param segun la notacion de la spec SDD.
#   <nombre>           = placeholder de Gherkin (Esquema del Escenario).
#   "campo"            = nombre de campo del JSON, entrecomillado para que el
#                        paso sea inequivoco al traducirlo a un test Go.
#   E-nn               = identificador estable del escenario; se reutiliza
#                        como nombre del subtest t.Run() en los *_test.go.
#
# Etiquetas:
#   @caso_normal        comportamiento principal o recorrido de extremo a
#                       extremo.
#   @caso_alternativo   camino valido distinto del principal.
#   @caso_limite        borde de los valores permitidos: longitudes, correlativos
#                       de codigo, valores por defecto y transiciones idempotentes.
#   @caso_error         rechazo esperado: 400 / 404 / 409 / 422 / 500.
#   @rfc                verificacion del estandar RFC 3339 UTC en creado_en y
#                       actualizado_en.
#
# Notas de la spec (fijadas por la SDD):
#   NOTA-1  El rechazo por longitud del titulo (> 200 caracteres) usa la
#           redaccion homologada a US-01 (RB-01): "el título de la historia no
#           puede superar los 200 caracteres" (campo "titulo").
#   NOTA-2  El PATCH sin ningun campo informado responde 422 con el mensaje
#           "debés informar al menos un campo para modificar la historia" y sin
#           "campo", porque el defecto no pertenece a un campo concreto del body.
#
# Entorno de ejecucion (variables resueltas antes de correr los escenarios):
#   BASE_URL = http://localhost:$PORT   (PORT segun .env; 8080 en .env.example
#                                       y 8081 si la variable no esta definida)
#   DB       = PostgreSQL 16 de docker-compose, publicado en el host en 5433
#               (DB_PORT=5433 para la API en el host y 5432 dentro de la red de
#               contenedores, ver US-01 Restriccion 13).
#   Ejemplo curl:
#     curl -i -X POST "$BASE_URL/api/proyectos/1/historias" \
#       -H "Content-Type: application/json" -d @historia.json
#   Ejemplo httptest: los mismos pasos se ejecutan contra el *gin.Engine real en
#     internal/handler/historia_handler_test.go y cmd/api/router_test.go.
#
# -----------------------------------------------------------------------------
# NARRATIVA (US-02, issue #29)
# -----------------------------------------------------------------------------
# Como Product Builder, quiero agregar, listar y modificar historias de usuario
# en el Product Backlog, para administrar el alcance del proyecto.
#
# -----------------------------------------------------------------------------
# CRITERIOS DE ACEPTACION DE LA ISSUE #29
# -----------------------------------------------------------------------------
# [CA-I1] Cada historia contiene ID, Titulo, Descripcion, Prioridad y Estado.
#                                       -> E-01, E-02, E-03, E-04, E-05, E-06,
#                                          E-08, E-12, E-13, E-18, E-23, E-24,
#                                          E-25, E-28
# [CA-I2] Se puede ver la lista completa de historias del backlog.
#                                       -> E-18, E-19, E-22
# [CA-I3] Se puede modificar una historia existente (Titulo, Descripcion,
#         Prioridad y/o Estado).        -> E-23, E-24, E-25, E-29
#
# Criterios adicionales sugeridos por el equipo (issue #29):
# [CA-I4] El ID/codigo es unico dentro del proyecto, autogenerado con el formato
#         "US-NN" y no editable.         -> E-01, E-04, E-12, E-13, E-14
# [CA-I5] La Prioridad pertenece a {Alta, Media, Baja}, por defecto Media.
#                                       -> E-02, E-10, E-27
# [CA-I6] El Estado pertenece a {Backlog, En Progreso, Hecho}, por defecto
#         Backlog.                      -> E-02, E-11, E-24, E-27
# [CA-I7] La Descripcion es opcional.   -> E-03, E-24, E-25
# [CA-I8] Las historias pertenecen a un proyecto y el listado es por proyecto.
#                                       -> E-01, E-15, E-18, E-19, E-20, E-22,
#                                          E-30, E-31, E-32
#
# -----------------------------------------------------------------------------
# REGLAS DE NEGOCIO CUBIERTAS
# -----------------------------------------------------------------------------
# RB-01 titulo obligatorio de 1 a 200 caracteres tras TrimSpace
#                                  -> E-05, E-06, E-07, E-08, E-09, E-27, E-28
# RB-02 codigo unico por proyecto, autogenerado "US-NN" y no editable
#                                  -> E-01, E-04, E-12, E-13, E-14
# RB-03 prioridad solo "Alta", "Media" o "Baja"; por defecto "Media"
#                                  -> E-02, E-10, E-27
# RB-04 estado solo "Backlog", "En Progreso" o "Hecho"; por defecto "Backlog"
#                                  -> E-02, E-11, E-24, E-27
# RB-05 la descripcion es opcional  -> E-03, E-24, E-25
# RB-06 la historia pertenece a un proyecto existente; el listado es por proyecto
#                                  -> E-01, E-15, E-18, E-19, E-20, E-22, E-30,
#                                     E-31, E-32
# RB-07 el id y el codigo no se aceptan del cliente        -> E-04
# RB-08 en el PATCH al menos un campo debe informarse      -> E-26
# RB-09 el PATCH valida prioridad y estado con el mismo conjunto que el alta
#                                  -> E-24, E-25, E-27
# RB-10 el PATCH conserva la regla de titulo 1..200        -> E-27, E-28
#
# -----------------------------------------------------------------------------
# CASOS LIMITE CUBIERTOS
# -----------------------------------------------------------------------------
# CL-01 titulo vacio                                    -> E-06
# CL-02 titulo formado solo por espacios o tabuladores  -> E-07
# CL-03 titulo de 1 caracter                            -> E-05
# CL-04 titulo de exactamente 200 caracteres            -> E-05, E-08, E-28
# CL-05 titulo de 201 caracteres                        -> E-05, E-09
# CL-06 titulo de exactamente 150 caracteres            -> E-05
# CL-07 alta sin prioridad ni estado                    -> E-02
# CL-08 descripcion ausente, nula o vacia               -> E-03
# CL-09 prioridad con otra capitalizacion o fuera del conjunto -> E-10
# CL-10 estado con otra capitalizacion o fuera del conjunto    -> E-11
# CL-11 segunda historia del proyecto -> codigo "US-02" -> E-12
# CL-12 decima historia del proyecto  -> codigo "US-10" -> E-13
# CL-13 listado de un proyecto sin historias            -> E-19
# CL-14 PATCH sin ningun campo informado                -> E-26
# CL-15 PATCH con un solo campo informado               -> E-24
# CL-16 PATCH idempotente al estado en que ya se encuentra -> E-29
# CL-17 dos altas concurrentes en el mismo proyecto     -> E-14
# CL-18 el codigo enviado por el cliente se ignora      -> E-04
# CL-19 historia de otro proyecto: invisible para la ruta -> E-32
#
# -----------------------------------------------------------------------------
# ENDPOINTS CUBIERTOS (los 3 de la spec SDD, ninguno mas)
# -----------------------------------------------------------------------------
#   1. POST   /api/proyectos/{id}/historias                  -> E-01 a E-17
#   2. GET    /api/proyectos/{id}/historias                  -> E-18 a E-22
#   3. PATCH  /api/proyectos/{id}/historias/{historiaId}     -> E-23 a E-34
# =============================================================================

Funcionalidad: US-02 - Gestion basica del Product Backlog
  Como Product Builder
  Quiero agregar, listar y modificar historias de usuario en el Product Backlog
  Para administrar el alcance del proyecto.

  Toda historia pertenece a un proyecto y se identifica con un codigo
  autogenerado con el formato "US-NN", unico dentro de su proyecto y no
  editable por el cliente. Cada historia nace en estado "Backlog" y con
  prioridad "Media", de modo que el alta solo exige el titulo. El backlog se
  consulta siempre por proyecto y la modificacion es parcial: el cliente envia
  unicamente los campos que desea cambiar. No existe autenticacion en US-02: el
  actor es implicitamente un Product Builder que trabaja sobre el proyecto. Los
  campos Story Points y criterios de aceptacion del requerimiento 2 de
  docs/proyecto.md no forman parte de US-02: se cubren en las historias de
  estimacion y de planificacion.

  Contexto:
    Dado que la API esta levantada y responde "200 OK" en "GET /health"
    Y que la base de datos "metrics_db" fue migrada en el arranque por
      "AutoMigrate" y existen las tablas "proyectos", "historias", "usuarios" y
      "proyecto_usuarios"
    Y que la suite arranca de una base limpia, con estas sentencias ejecutadas
      por el runner antes de cada escenario:
      """
      DELETE FROM historias;
      DELETE FROM proyecto_usuarios;
      DELETE FROM usuarios;
      DELETE FROM proyectos;
      """
    Y que la ruta base de la API es "$BASE_URL/api"
    Y que se conoce el cuerpo de error estandar de la spec SDD:
      """
      {
        "error":   "<CODIGO_EN_UPPER_SNAKE_CASE>",
        "mensaje": "<mensaje en espanol>",
        "campo":   "<solo cuando el error apunta a un campo del body>"
      }
      """

  # ===========================================================================
  # ALTA DE HISTORIAS: POST /api/proyectos/{id}/historias
  # ===========================================================================
  Regla: Alta de una historia en el Product Backlog de un proyecto

    @caso_normal
    Escenario: E-01 - Alta de una historia con todos sus campos
      # Caso normal de la issue #29 (CA-I1, CA-I4, CA-I8) y reglas RB-01,
      # RB-02 y RB-06. La historia se persiste con su id y su codigo
      # autogenerados y aparece en el listado del backlog.
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y sin historias
      Y que preparo el cuerpo JSON del alta de la historia:
        """
        {
          "titulo": "Registrar horas reales de trabajo",
          "descripcion": "Permitir registrar horas por usuario, fecha y actividad.",
          "prioridad": "Alta",
          "estado": "Backlog"
        }
        """
      Cuando envío "POST" a "/api/proyectos/1/historias" con ese cuerpo
      Entonces la API responde "201 Created" con la historia persistida
      Y el campo "id" es un entero positivo generado por el sistema
      Y el campo "codigo" es "US-01"
      Y el campo "titulo" es "Registrar horas reales de trabajo"
      Y el campo "descripcion" es "Permitir registrar horas por usuario, fecha y actividad."
      Y el campo "prioridad" es "Alta"
      Y el campo "estado" es "Backlog"
      Y los campos "creado_en" y "actualizado_en" tienen formato RFC 3339 UTC
      Y la respuesta no incluye el campo "proyecto_id"
      Y cuando consulto "GET" en "/api/proyectos/1/historias"
      Entonces la API responde "200 OK" con el campo "total" igual a 1
      Y el array "historias" contiene la historia con "codigo" "US-01"

    @caso_alternativo @caso_limite
    Esquema del Escenario: E-02 - Alta con prioridad y estado por defecto en la variante <variante>
      # CA-I5, CA-I6, RB-03, RB-04 y CL-07: la prioridad por defecto es "Media"
      # y el estado por defecto es "Backlog"; los valores informados mandan.
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo"
      Y que preparo el cuerpo JSON del alta:
        """
        {
          "titulo": "Historia de valores por defecto"
          <linea_campos_opcionales>
        }
        """
      Cuando envío "POST" a "/api/proyectos/1/historias" con ese cuerpo
      Entonces la API responde "201 Created"
      Y el campo "prioridad" de la respuesta es "<prioridad_esperada>"
      Y el campo "estado" de la respuesta es "<estado_esperado>"
      Y la historia queda visible en "GET /api/proyectos/1/historias"

      Ejemplos:
        | variante                      | linea_campos_opcionales                  | prioridad_esperada | estado_esperado |
        | sin prioridad ni estado       |                                          | Media              | Backlog         |
        | solo prioridad Alta           | , "prioridad": "Alta"                    | Alta               | Backlog         |
        | solo estado En Progreso       | , "estado": "En Progreso"                | Media              | En Progreso     |
        | prioridad Baja y estado Hecho | , "prioridad": "Baja", "estado": "Hecho" | Baja               | Hecho           |

    @caso_alternativo @caso_limite
    Esquema del Escenario: E-03 - Alta con la "descripcion" en la variante <variante_descripcion>
      # CA-I7, RB-05 y CL-08: la descripcion es opcional y se normaliza a la
      # cadena vacia; la respuesta nunca devuelve "null" en ese campo.
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo"
      Y que la descripcion se envia en la variante <variante_descripcion>
      Y que preparo el cuerpo JSON del alta:
        """
        {
          "titulo": "Historia con descripcion opcional",
          "prioridad": "Media"
          <linea_descripcion>
        }
        """
      Cuando envío "POST" a "/api/proyectos/1/historias" con ese cuerpo
      Entonces la API responde "201 Created"
      Y el campo "descripcion" de la respuesta es la cadena vacia y no "null"
      Y la historia queda persistida y visible en el listado del backlog

      Ejemplos:
        | variante_descripcion | linea_descripcion      |
        | ausente              |                        |
        | nula                 | , "descripcion": null  |
        | vacia                | , "descripcion": ""    |

    @caso_alternativo
    Escenario: E-04 - El cliente no puede elegir el "id" ni el "codigo"
      # CA-I4, RB-02, RB-07 y CL-18: el id y el codigo los autogenera el
      # sistema; los valores enviados en el body se ignoran en silencio, igual
      # que cualquier campo desconocido, incluido "proyecto_id".
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y sin historias
      Y que preparo el cuerpo JSON del alta:
        """
        {
          "id": 999,
          "codigo": "US-99",
          "proyecto_id": 7,
          "titulo": "Historia con campos ajenos",
          "prioridad": "Media"
        }
        """
      Cuando envío "POST" a "/api/proyectos/1/historias" con ese cuerpo
      Entonces la API responde "201 Created"
      Y el campo "id" de la respuesta no es 999, es el generado por el sistema
      Y el campo "codigo" de la respuesta es "US-01" y no "US-99"
      Y la respuesta no incluye el campo "proyecto_id"

    @caso_limite @caso_error
    Esquema del Escenario: E-05 - Matriz de borde del campo "titulo"
      # CA-I1, RB-01 y CL-01, CL-02, CL-03, CL-04, CL-05 y CL-06: concentra en
      # un solo esquema los bordes del titulo. El titulo de 150 caracteres es la
      # cadena "ABCDE" repetida 30 veces; el de 200, "ABCDE" repetida 40 veces;
      # y el de 201, esa misma cadena seguida de "F". E-06 a E-09 cubren cada
      # caso por separado con su diagnostico completo.
      # NOTA-1: el literal del rechazo por longitud se homologa a US-01 (RB-01).
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo"
      Y que preparo el cuerpo JSON del alta con el campo "titulo" en
        <titulo_ingresado>
      Cuando envío "POST" a "/api/proyectos/1/historias" con ese cuerpo
      Entonces la API responde "<codigo_http_esperado>"
      Y el cuerpo de la respuesta es <cuerpo_esperado>
      Y cuando consulto "GET" en "/api/proyectos/1/historias"
      Entonces la API responde "200 OK" con el campo "total" igual a
        <total_esperado>

      Ejemplos:
        | titulo_ingresado                                                     | codigo_http_esperado       | cuerpo_esperado                                                                                                              | total_esperado |
        | la cadena de un caracter "P"                                         | "201 Created"              | la historia creada con "titulo" "P" y "codigo" "US-01"                                                                       | 1              |
        | la cadena "ABCDE" repetida 30 veces                                  | "201 Created"              | la historia creada con "titulo" identico al enviado                                                                          | 1              |
        | la cadena "ABCDE" repetida 40 veces                                  | "201 Created"              | la historia creada con "titulo" de exactamente 200 caracteres                                                                | 1              |
        | la cadena vacia ""                                                   | "422 Unprocessable Entity" | el error de validacion con "VALIDACION", "el título de la historia es obligatorio" y "campo" "titulo"                        | 0              |
        | la cadena de tres caracteres de espacio "   "                        | "422 Unprocessable Entity" | el error de validacion con "VALIDACION", "el título de la historia es obligatorio" y "campo" "titulo"                        | 0              |
        | la cadena "ABCDE" repetida 40 veces seguida de la letra "F"          | "422 Unprocessable Entity" | el error de validacion con "VALIDACION", "el título de la historia no puede superar los 200 caracteres" y "campo" "titulo"  | 0              |

    @caso_error
    Escenario: E-06 - Alta de una historia con el titulo vacio
      # CA-I1, RB-01 y CL-01: traduccion directa del escenario de error de la
      # issue #29. El campo reportado es "titulo".
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo"
      Y que preparo el cuerpo JSON del alta:
        """
        {
          "titulo": "",
          "descripcion": "Historia sin titulo.",
          "prioridad": "Media",
          "estado": "Backlog"
        }
        """
      Cuando envío "POST" a "/api/proyectos/1/historias" con ese cuerpo
      Entonces la API responde "422 Unprocessable Entity"
      Y el campo "error" del cuerpo es "VALIDACION"
      Y el campo "mensaje" del cuerpo es "el título de la historia es obligatorio"
      Y el campo "campo" del cuerpo es "titulo"
      Y cuando consulto "GET" en "/api/proyectos/1/historias"
      Entonces la API responde "200 OK" con el campo "total" igual a 0

    @caso_limite @caso_error
    Esquema del Escenario: E-07 - Alta con el titulo formado solo por espacios
      # CL-02: RB-01 valida sobre el resultado de TrimSpace, por lo que los
      # espacios y los tabuladores producen el mismo rechazo que el vacio.
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo"
      Y que preparo el cuerpo JSON del alta con el campo "titulo" en
        <detalle_del_titulo>
      Cuando envío "POST" a "/api/proyectos/1/historias" con ese cuerpo
      Entonces la API responde "422 Unprocessable Entity"
      Y el campo "error" del cuerpo es "VALIDACION"
      Y el campo "mensaje" del cuerpo es "el título de la historia es obligatorio"
      Y el campo "campo" del cuerpo es "titulo"
      Y cuando consulto "GET" en "/api/proyectos/1/historias"
      Entonces la API responde "200 OK" con el campo "total" igual a 0

      Ejemplos:
        | detalle_del_titulo                          |
        | la cadena de tres caracteres de espacio     |
        | la cadena de dos caracteres de tabulacion   |
        | la cadena de espacios y tabulaciones " \t " |

    @caso_limite
    Escenario: E-08 - Alta con el titulo de exactamente 200 caracteres
      # CA-I1, RB-01 y CL-04: 200 es el maximo permitido y se acepta. El titulo
      # es la cadena "ABCDE" repetida 40 veces, es decir 200 caracteres exactos.
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo"
      Y que preparo el cuerpo JSON del alta con el campo "titulo" formado por
        la cadena "ABCDE" repetida 40 veces, es decir 200 caracteres exactos
      Cuando envío "POST" a "/api/proyectos/1/historias" con ese cuerpo
      Entonces la API responde "201 Created"
      Y el campo "titulo" de la respuesta tiene exactamente 200 caracteres
      Y el campo "titulo" de la respuesta es igual al enviado

    @caso_limite @caso_error
    Escenario: E-09 - Alta con el titulo de 201 caracteres
      # RB-01 y CL-05: un caracter por encima del maximo se rechaza. El titulo
      # es la cadena "ABCDE" repetida 40 veces seguida de la letra "F".
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo"
      Y que preparo el cuerpo JSON del alta con el campo "titulo" formado por
        la cadena "ABCDE" repetida 40 veces seguida de la letra "F"
      Cuando envío "POST" a "/api/proyectos/1/historias" con ese cuerpo
      Entonces la API responde "422 Unprocessable Entity"
      Y el campo "error" del cuerpo es "VALIDACION"
      Y el campo "mensaje" del cuerpo es
        "el título de la historia no puede superar los 200 caracteres"
      Y el campo "campo" del cuerpo es "titulo"
      Y cuando consulto "GET" en "/api/proyectos/1/historias"
      Entonces la API responde "200 OK" con el campo "total" igual a 0

    @caso_error @caso_limite
    Esquema del Escenario: E-10 - Alta con la "prioridad" fuera del conjunto permitido
      # CA-I5, RB-03 y CL-09: solo se admiten "Alta", "Media" y "Baja"; la
      # comparacion del literal distingue mayusculas y la cadena vacia se
      # rechaza. El campo reportado es "prioridad".
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo"
      Y que preparo el cuerpo JSON del alta con el titulo "Historia con prioridad invalida"
        y el campo "prioridad" en <prioridad_ingresada>
      Cuando envío "POST" a "/api/proyectos/1/historias" con ese cuerpo
      Entonces la API responde "422 Unprocessable Entity"
      Y el campo "error" del cuerpo es "VALIDACION"
      Y el campo "mensaje" del cuerpo es
        'la prioridad de la historia debe ser "Alta", "Media" o "Baja"'
      Y el campo "campo" del cuerpo es "prioridad"
      Y cuando consulto "GET" en "/api/proyectos/1/historias"
      Entonces la API responde "200 OK" con el campo "total" igual a 0

      Ejemplos:
        | prioridad_ingresada                  |
        | la cadena "alta"                     |
        | la cadena "ALTA"                     |
        | la cadena "Urgente"                  |
        | la cadena vacia                      |
        | el valor null                        |

    @caso_error @caso_limite
    Esquema del Escenario: E-11 - Alta con el "estado" fuera del conjunto permitido
      # CA-I6, RB-04 y CL-10: solo se admiten "Backlog", "En Progreso" y
      # "Hecho"; "En progreso" y "en Progreso" se rechazan porque el literal
      # distingue la capitalizacion exacta. El campo reportado es "estado".
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo"
      Y que preparo el cuerpo JSON del alta con el titulo "Historia con estado invalido"
        y el campo "estado" en <estado_ingresado>
      Cuando envío "POST" a "/api/proyectos/1/historias" con ese cuerpo
      Entonces la API responde "422 Unprocessable Entity"
      Y el campo "error" del cuerpo es "VALIDACION"
      Y el campo "mensaje" del cuerpo es
        'el estado de la historia debe ser "Backlog", "En Progreso" o "Hecho"'
      Y el campo "campo" del cuerpo es "estado"
      Y cuando consulto "GET" en "/api/proyectos/1/historias"
      Entonces la API responde "200 OK" con el campo "total" igual a 0

      Ejemplos:
        | estado_ingresado            |
        | la cadena "backlog"         |
        | la cadena "En progreso"     |
        | la cadena "en Progreso"     |
        | la cadena "Done"            |
        | la cadena vacia             |
        | el valor null               |

    @caso_alternativo @caso_limite
    Escenario: E-12 - Alta de una segunda historia genera el codigo "US-02"
      # CA-I4, RB-02 y CL-11: el codigo es unico y consecutivo por proyecto; la
      # segunda historia recibe "US-02".
      Dado que el proyecto "Software Metrics & Estimation" con "id" 1 ya tiene
        una historia con "codigo" "US-01"
      Y que preparo el cuerpo JSON del alta:
        """
        { "titulo": "Segunda historia del backlog" }
        """
      Cuando envío "POST" a "/api/proyectos/1/historias" con ese cuerpo
      Entonces la API responde "201 Created"
      Y el campo "codigo" de la respuesta es "US-02"
      Y cuando consulto "GET" en "/api/proyectos/1/historias"
      Entonces la API responde "200 OK" con el campo "total" igual a 2
      Y el array "historias" contiene exactamente un elemento con "codigo" "US-01"
        y exactamente un elemento con "codigo" "US-02"

    @caso_limite
    Escenario: E-13 - Alta de la decima historia genera el codigo "US-10"
      # CA-I4, RB-02 y CL-12: el correlativo se rellena a dos digitos y el
      # formato "US-NN" se conserva al cruzar la decena.
      Dado que el proyecto "Software Metrics & Estimation" con "id" 1 ya tiene
        nueve historias con los codigos "US-01" a "US-09"
      Y que preparo el cuerpo JSON del alta:
        """
        { "titulo": "Decima historia del backlog" }
        """
      Cuando envío "POST" a "/api/proyectos/1/historias" con ese cuerpo
      Entonces la API responde "201 Created"
      Y el campo "codigo" de la respuesta es "US-10"
      Y cuando consulto "GET" en "/api/proyectos/1/historias"
      Entonces la API responde "200 OK" con el campo "total" igual a 10

    @caso_error @caso_limite
    Escenario: E-14 - Dos altas concurrentes no repiten el codigo en el mismo proyecto
      # CA-I4, RB-02 y CL-17: el indice unico (proyecto_id, codigo) es la
      # garantia final frente a la condicion de carrera. El escenario se ejecuta
      # lanzando las dos peticiones en paralelo, con dos curl en background o con
      # dos goroutines en httptest. Una peticion completa con 201 y la otra se
      # rechaza con 409 HISTORIA_DUPLICADA.
      Dado que el proyecto "Software Metrics & Estimation" con "id" 1 no tiene
        historias
      Y que el cuerpo de las dos peticiones es:
        """
        { "titulo": "Historia concurrente" }
        """
      Y que envío al mismo tiempo dos peticiones "POST" a
        "/api/proyectos/1/historias" con ese cuerpo
      Entonces una de las dos respuestas es "201 Created" y la otra es "409 Conflict"
      Y el cuerpo de la respuesta "409" tiene "error" "HISTORIA_DUPLICADA"
      Y el campo "mensaje" del cuerpo de error es
        "ya existe una historia con ese código en el proyecto"
      Y cuando consulto "GET" en "/api/proyectos/1/historias"
      Entonces la API responde "200 OK" con el campo "total" igual a 1
      Y el codigo "US-01" figura una sola vez en el array "historias"

    @caso_error
    Escenario: E-15 - Alta de una historia en un proyecto inexistente
      # CA-I8, RB-06 y CL-06: la historia pertenece a un proyecto; sin proyecto
      # no hay alta y el error es PROYECTO_NO_ENCONTRADO.
      Dado que no existe ningun proyecto con el "id" 999
      Y que preparo el cuerpo JSON del alta:
        """
        { "titulo": "Historia huerfana", "prioridad": "Media" }
        """
      Cuando envío "POST" a "/api/proyectos/999/historias" con ese cuerpo
      Entonces la API responde "404 Not Found"
      Y el campo "error" del cuerpo es "PROYECTO_NO_ENCONTRADO"
      Y el campo "mensaje" del cuerpo es "el proyecto no existe"
      Y el cuerpo de la respuesta no tiene el campo "campo"

    @caso_error @caso_limite
    Esquema del Escenario: E-16 - Alta con un identificador de proyecto no valido
      # Un fallo de strconv.ParseUint o un resultado no positivo producen
      # 400 PARAMETRO_INVALIDO antes de consultar la base de datos.
      Dado que existe al menos un proyecto dado de alta
      Y que preparo el cuerpo JSON del alta:
        """
        { "titulo": "Historia con ruta invalida" }
        """
      Cuando envío "POST" a "/api/proyectos/<id_invalido>/historias" con ese cuerpo
      Entonces la API responde "400 Bad Request"
      Y el campo "error" del cuerpo es "PARAMETRO_INVALIDO"
      Y el cuerpo de la respuesta no tiene el campo "campo"

      Ejemplos:
        | id_invalido   |
        | la cadena abc |
        | el numero 0   |
        | el numero -1  |

    @caso_error
    Esquema del Escenario: E-17 - Cuerpo del request mal formado en el alta
      # Convencion de la spec: un fallo de ShouldBindJSON produce
      # 400 JSON_INVALIDO, sin "campo", y ningun dato se persiste.
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo"
      Y que envío un cuerpo <descripcion_del_cuerpo> en
        "POST /api/proyectos/1/historias" con la cabecera
        "Content-Type: application/json"
      Entonces la API responde "400 Bad Request"
      Y el campo "error" del cuerpo es "JSON_INVALIDO"
      Y el cuerpo de la respuesta no tiene el campo "campo"
      Y cuando consulto "GET" en "/api/proyectos/1/historias"
      Entonces la API responde "200 OK" con el campo "total" igual a 0

      Ejemplos:
        | descripcion_del_cuerpo                          |
        | vacio                                           |
        | con el JSON truncado, sin la llave de cierre    |
        | con "titulo" como numero en lugar de texto      |
        | con "prioridad" como array en lugar de texto    |

  # ===========================================================================
  # LISTADO DEL BACKLOG: GET /api/proyectos/{id}/historias
  # ===========================================================================
  Regla: Consulta del listado del Product Backlog de un proyecto

    @caso_normal @rfc
    Escenario: E-18 - Listado del backlog con varias historias
      # CA-I1, CA-I2, CA-I8 y RB-06: GET lista el backlog del proyecto y
      # devuelve {"total": n, "historias": [...]}. Con @rfc se verifica la
      # serializacion RFC 3339 UTC de creado_en y actualizado_en.
      Dado que el proyecto "Software Metrics & Estimation" con "id" 1 tiene las
        historias "US-01" "Registrar horas reales" con prioridad "Alta" y estado
        "Backlog", y "US-02" "Cerrar Sprint" con prioridad "Media" y estado
        "En Progreso"
      Cuando consulto "GET" en "/api/proyectos/1/historias"
      Entonces la API responde "200 OK"
      Y el campo "total" es 2
      Y el array "historias" tiene 2 elementos
      Y cada elemento expone "id", "codigo", "titulo", "descripcion",
        "prioridad", "estado", "creado_en" y "actualizado_en"
      Y ningun elemento incluye el campo "proyecto_id"
      Y el elemento con "codigo" "US-01" tiene "prioridad" "Alta"
      Y el elemento con "codigo" "US-02" tiene "estado" "En Progreso"
      Y los campos "creado_en" y "actualizado_en" de cada elemento responden
        con el patron RFC 3339 UTC

    @caso_limite
    Escenario: E-19 - Listado de un proyecto sin historias
      # CA-I2, CA-I8 y CL-13: el array vacio se serializa explicitamente para
      # no producir null.
      Dado que el proyecto "Software Metrics & Estimation" con "id" 1 existe y
        no tiene ninguna historia en su backlog
      Cuando consulto "GET" en "/api/proyectos/1/historias"
      Entonces la API responde "200 OK"
      Y el campo "total" es 0
      Y el campo "historias" es el array vacio y no "null"
      Y el cuerpo de la respuesta es, en su forma equivalente:
        """
        { "total": 0, "historias": [] }
        """

    @caso_error
    Escenario: E-20 - Listado del backlog de un proyecto inexistente
      # CA-I8 y RB-06: el listado exige un proyecto existente.
      Dado que no existe ningun proyecto con el "id" 999
      Cuando consulto "GET" en "/api/proyectos/999/historias"
      Entonces la API responde "404 Not Found"
      Y el campo "error" del cuerpo es "PROYECTO_NO_ENCONTRADO"
      Y el campo "mensaje" del cuerpo es "el proyecto no existe"
      Y el cuerpo de la respuesta no tiene el campo "campo"

    @caso_error @caso_limite
    Esquema del Escenario: E-21 - Listado con un identificador de proyecto no valido
      # El handler parsea el path param antes de consultar la base de datos.
      Dado que existe al menos un proyecto dado de alta
      Cuando consulto "GET" en "/api/proyectos/<id_invalido>/historias"
      Entonces la API responde "400 Bad Request"
      Y el campo "error" del cuerpo es "PARAMETRO_INVALIDO"
      Y el cuerpo de la respuesta no tiene el campo "campo"

      Ejemplos:
        | id_invalido   |
        | la cadena abc |
        | el numero 0   |
        | el numero -1  |

    @caso_alternativo
    Escenario: E-22 - El listado del backlog es por proyecto
      # CA-I2 y CA-I8: dos proyectos con historias propias no mezclan sus
      # backlogs; cada listado devuelve solo las historias de su proyecto.
      Dado que el proyecto con "id" 1 tiene la historia "US-01" "Backlog del proyecto uno"
      Y que el proyecto con "id" 2 tiene las historias "US-01" "Backlog del proyecto dos"
        y "US-02" "Segunda del proyecto dos"
      Cuando consulto "GET" en "/api/proyectos/1/historias"
      Entonces la API responde "200 OK" con el campo "total" igual a 1
      Y el array "historias" contiene unicamente "Backlog del proyecto uno"
      Y cuando consulto "GET" en "/api/proyectos/2/historias"
      Entonces la API responde "200 OK" con el campo "total" igual a 2
      Y el array "historias" contiene "Backlog del proyecto dos" y
        "Segunda del proyecto dos"

  # ===========================================================================
  # MODIFICACION: PATCH /api/proyectos/{id}/historias/{historiaId}
  # ===========================================================================
  Regla: Modificacion parcial de una historia del Product Backlog

    @caso_normal
    Escenario: E-23 - Modificacion parcial de la prioridad de una historia
      # CA-I3 y RB-09: PATCH actualiza solo los campos informados, responde 200
      # con la historia actualizada y el cambio se refleja en el listado.
      Dado que el proyecto con "id" 1 tiene la historia "id" 5 con "codigo"
        "US-01", "titulo" "Registrar horas reales", "prioridad" "Media" y
        "estado" "Backlog"
      Y que preparo el cuerpo JSON de la modificacion:
        """
        { "prioridad": "Alta" }
        """
      Cuando envío "PATCH" a "/api/proyectos/1/historias/5" con ese cuerpo
      Entonces la API responde "200 OK" con la historia completa
      Y el campo "id" sigue siendo 5
      Y el campo "codigo" sigue siendo "US-01"
      Y el campo "prioridad" de la respuesta es "Alta"
      Y el campo "titulo" de la respuesta sigue siendo "Registrar horas reales"
      Y el campo "estado" de la respuesta sigue siendo "Backlog"
      Y la respuesta no incluye el campo "proyecto_id"
      Y cuando consulto "GET" en "/api/proyectos/1/historias"
      Entonces la API responde "200 OK" y la historia "US-01" figura con
        "prioridad" "Alta"

    @caso_alternativo @caso_limite
    Esquema del Escenario: E-24 - Modificacion parcial de un unico campo
      # CA-I3, CA-I6, CA-I7, RB-04, RB-05, RB-09 y CL-15: cada fila modifica un
      # unico campo y deja el resto intacto.
      Dado que el proyecto con "id" 1 tiene la historia "id" 5 con "codigo"
        "US-01", "titulo" "Informe de metricas", "descripcion" "Version inicial",
        "prioridad" "Media" y "estado" "Backlog"
      Y que preparo el cuerpo JSON de la modificacion con el campo "<campo>" en
        "<valor_nuevo>"
      Cuando envío "PATCH" a "/api/proyectos/1/historias/5" con ese cuerpo
      Entonces la API responde "200 OK"
      Y el campo "<campo>" de la respuesta es "<valor_nuevo>"
      Y cuando consulto "GET" en "/api/proyectos/1/historias"
      Entonces la API responde "200 OK" y el cambio de "<campo>" se refleja en
        el array "historias"

      Ejemplos:
        | campo       | valor_nuevo                        |
        | estado      | Hecho                              |
        | titulo      | Informe de metricas corregido      |
        | descripcion | Version ampliada con criterios     |
        | prioridad   | Baja                               |

    @caso_limite
    Escenario: E-25 - Modificacion de varios campos a la vez
      # CA-I3, CA-I7, RB-05 y RB-09: un PATCH puede combinar Titulo,
      # Descripcion, Prioridad y Estado en una sola peticion.
      Dado que el proyecto con "id" 1 tiene la historia "id" 5 con "codigo"
        "US-01", "titulo" "Registrar horas", "descripcion" "Version inicial",
        "prioridad" "Media" y "estado" "Backlog"
      Y que preparo el cuerpo JSON de la modificacion:
        """
        {
          "titulo": "Registrar horas reales de trabajo",
          "descripcion": "Permitir registrar horas por usuario, fecha y actividad.",
          "prioridad": "Alta",
          "estado": "Hecho"
        }
        """
      Cuando envío "PATCH" a "/api/proyectos/1/historias/5" con ese cuerpo
      Entonces la API responde "200 OK"
      Y el campo "titulo" de la respuesta es "Registrar horas reales de trabajo"
      Y el campo "descripcion" de la respuesta es
        "Permitir registrar horas por usuario, fecha y actividad."
      Y el campo "prioridad" de la respuesta es "Alta"
      Y el campo "estado" de la respuesta es "Hecho"
      Y el campo "id" sigue siendo 5 y el campo "codigo" sigue siendo "US-01"
      Y el campo "actualizado_en" de la respuesta tiene formato RFC 3339 UTC

    @caso_error @caso_limite
    Escenario: E-26 - Modificacion sin ningun campo informado
      # CA-I3, RB-08 y CL-14: un PATCH sin campos no tiene efecto y el service
      # lo rechaza con 422. La SDD fija el mensaje y no exige "campo", porque el
      # defecto no pertenece a un campo concreto del body.
      Dado que el proyecto con "id" 1 tiene la historia "id" 5 con "codigo"
        "US-01", "titulo" "Informe de metricas", "prioridad" "Media" y
        "estado" "Backlog"
      Y que preparo el cuerpo JSON de la modificacion:
        """
        {}
        """
      Cuando envío "PATCH" a "/api/proyectos/1/historias/5" con ese cuerpo
      Entonces la API responde "422 Unprocessable Entity"
      Y el campo "error" del cuerpo es "VALIDACION"
      Y el campo "mensaje" del cuerpo es
        "debés informar al menos un campo para modificar la historia"
      Y cuando consulto "GET" en "/api/proyectos/1/historias"
      Entonces la API responde "200 OK" y la historia conserva "prioridad" "Media"
        y "estado" "Backlog"

    @caso_error @caso_limite
    Esquema del Escenario: E-27 - Modificacion con un valor invalido en el campo
      # RB-01, RB-03, RB-04, RB-09 y RB-10: la modificacion valida con las
      # mismas reglas que el alta y no deja la historia a medio actualizar.
      Dado que el proyecto con "id" 1 tiene la historia "id" 5 con "codigo"
        "US-01", "titulo" "Informe de metricas", "descripcion" "Version inicial",
        "prioridad" "Media" y "estado" "Backlog"
      Y que preparo el cuerpo JSON de la modificacion con el campo "<campo>" en
        <valor_invalido>
      Cuando envío "PATCH" a "/api/proyectos/1/historias/5" con ese cuerpo
      Entonces la API responde "422 Unprocessable Entity"
      Y el campo "error" del cuerpo es "VALIDACION"
      Y el campo "mensaje" del cuerpo es "<mensaje_esperado>"
      Y el campo "campo" del cuerpo es "<campo>"
      Y cuando consulto "GET" en "/api/proyectos/1/historias"
      Entonces la API responde "200 OK" y la historia conserva sus valores originales

      Ejemplos:
        | campo       | valor_invalido          | mensaje_esperado                                                                       |
        | titulo      | la cadena vacia ""      | el título de la historia es obligatorio                                                |
        | titulo      | la cadena "   "         | el título de la historia es obligatorio                                                |
        | titulo      | la cadena "ABCDE" repetida 40 veces seguida de "F" | el título de la historia no puede superar los 200 caracteres         |
        | prioridad   | la cadena "Urgente"     | la prioridad de la historia debe ser "Alta", "Media" o "Baja"                          |
        | prioridad   | la cadena "alta"        | la prioridad de la historia debe ser "Alta", "Media" o "Baja"                          |
        | estado      | la cadena "Done"        | el estado de la historia debe ser "Backlog", "En Progreso" o "Hecho"                   |
        | estado      | la cadena "En progreso" | el estado de la historia debe ser "Backlog", "En Progreso" o "Hecho"                   |

    @caso_limite
    Escenario: E-28 - Modificacion con un titulo de exactamente 200 caracteres
      # RB-01 y RB-10: el maximo del titulo se respeta tambien en la
      # modificacion. El titulo enviado es "ABCDE" repetida 40 veces.
      Dado que el proyecto con "id" 1 tiene la historia "id" 5 con "codigo"
        "US-01", "titulo" "Informe de metricas" y "estado" "Backlog"
      Y que preparo el cuerpo JSON de la modificacion con el campo "titulo"
        formado por la cadena "ABCDE" repetida 40 veces, es decir 200 caracteres
      Cuando envío "PATCH" a "/api/proyectos/1/historias/5" con ese cuerpo
      Entonces la API responde "200 OK"
      Y el campo "titulo" de la respuesta tiene exactamente 200 caracteres
      Y el campo "titulo" de la respuesta es igual al enviado

    @caso_limite
    Escenario: E-29 - Modificacion idempotente del estado al mismo valor
      # CA-I3 y CL-16: enviar el estado en que ya se encuentra la historia
      # responde 200 y no altera los demas campos.
      Dado que el proyecto con "id" 1 tiene la historia "id" 5 con "codigo"
        "US-01", "titulo" "Informe de metricas" y "estado" "Backlog"
      Y que preparo el cuerpo JSON de la modificacion:
        """
        { "estado": "Backlog" }
        """
      Cuando envío "PATCH" a "/api/proyectos/1/historias/5" con ese cuerpo
      Entonces la API responde "200 OK"
      Y el campo "estado" de la respuesta es "Backlog"
      Y el campo "titulo" de la respuesta sigue siendo "Informe de metricas"
      Y el campo "codigo" de la respuesta sigue siendo "US-01"

    @caso_error
    Escenario: E-30 - Modificacion de una historia inexistente
      # El identificador de historia se resuelve dentro del proyecto de la ruta.
      Dado que el proyecto con "id" 1 no tiene ninguna historia con el "id" 999
      Y que preparo el cuerpo JSON de la modificacion:
        """
        { "prioridad": "Alta" }
        """
      Cuando envío "PATCH" a "/api/proyectos/1/historias/999" con ese cuerpo
      Entonces la API responde "404 Not Found"
      Y el campo "error" del cuerpo es "HISTORIA_NO_ENCONTRADA"
      Y el campo "mensaje" del cuerpo es "la historia no existe"
      Y el cuerpo de la respuesta no tiene el campo "campo"

    @caso_error
    Escenario: E-31 - Modificacion en un proyecto inexistente
      # CA-I8 y RB-06: sin proyecto no hay historia que modificar.
      Dado que no existe ningun proyecto con el "id" 999
      Y que preparo el cuerpo JSON de la modificacion:
        """
        { "prioridad": "Alta" }
        """
      Cuando envío "PATCH" a "/api/proyectos/999/historias/5" con ese cuerpo
      Entonces la API responde "404 Not Found"
      Y el campo "error" del cuerpo es "PROYECTO_NO_ENCONTRADO"
      Y el campo "mensaje" del cuerpo es "el proyecto no existe"
      Y el cuerpo de la respuesta no tiene el campo "campo"

    @caso_error
    Escenario: E-32 - Modificacion de una historia que pertenece a otro proyecto
      # CA-I8 y CL-19: la historia se resuelve dentro del proyecto de la ruta;
      # una historia de otro proyecto es invisible y responde 404.
      Dado que el proyecto con "id" 2 tiene la historia "id" 7 con "codigo" "US-01"
      Y que el proyecto con "id" 1 no tiene ninguna historia con el "id" 7
      Y que preparo el cuerpo JSON de la modificacion:
        """
        { "estado": "Hecho" }
        """
      Cuando envío "PATCH" a "/api/proyectos/1/historias/7" con ese cuerpo
      Entonces la API responde "404 Not Found"
      Y el campo "error" del cuerpo es "HISTORIA_NO_ENCONTRADA"
      Y el campo "mensaje" del cuerpo es "la historia no existe"
      Y cuando consulto "GET" en "/api/proyectos/2/historias"
      Entonces la API responde "200 OK" y la historia "US-01" conserva su estado
        original

    @caso_error @caso_limite
    Esquema del Escenario: E-33 - Modificacion con un identificador de historia no valido
      # El handler parsea el path param antes de consultar la base de datos.
      Dado que el proyecto con "id" 1 tiene la historia "id" 5 con "codigo" "US-01"
      Y que preparo el cuerpo JSON de la modificacion:
        """
        { "prioridad": "Alta" }
        """
      Cuando envío "PATCH" a "/api/proyectos/1/historias/<historia_id_invalido>"
        con ese cuerpo
      Entonces la API responde "400 Bad Request"
      Y el campo "error" del cuerpo es "PARAMETRO_INVALIDO"
      Y el cuerpo de la respuesta no tiene el campo "campo"

      Ejemplos:
        | historia_id_invalido |
        | la cadena abc        |
        | el numero 0          |
        | el numero -1         |

    @caso_error
    Esquema del Escenario: E-34 - Cuerpo del request mal formado en la modificacion
      # Convencion de la spec: un fallo de ShouldBindJSON produce
      # 400 JSON_INVALIDO, sin "campo", y la historia no se modifica.
      Dado que el proyecto con "id" 1 tiene la historia "id" 5 con "codigo"
        "US-01", "titulo" "Informe de metricas" y "estado" "Backlog"
      Y que envío un cuerpo <descripcion_del_cuerpo> en
        "PATCH /api/proyectos/1/historias/5" con la cabecera
        "Content-Type: application/json"
      Entonces la API responde "400 Bad Request"
      Y el campo "error" del cuerpo es "JSON_INVALIDO"
      Y el cuerpo de la respuesta no tiene el campo "campo"
      Y cuando consulto "GET" en "/api/proyectos/1/historias"
      Entonces la API responde "200 OK" y la historia conserva "estado" "Backlog"

      Ejemplos:
        | descripcion_del_cuerpo                        |
        | vacio                                         |
        | con el JSON truncado, sin la llave de cierre  |
        | con "prioridad" como numero en lugar de texto |
        | con un texto plano que no es JSON             |

# =============================================================================
# MATRIZ DE TRAZABILIDAD
# Criterio de aceptacion de la issue #29 -> escenario
#   CA-I1 E-01 E-02 E-03 E-04 E-05 E-06 E-08 E-12 E-13 E-18 E-23 E-24 E-25 E-28
#   CA-I2 E-18 E-19 E-22
#   CA-I3 E-23 E-24 E-25 E-29
#   CA-I4 E-01 E-04 E-12 E-13 E-14
#   CA-I5 E-02 E-10 E-27
#   CA-I6 E-02 E-11 E-24 E-27
#   CA-I7 E-03 E-24 E-25
#   CA-I8 E-01 E-15 E-18 E-19 E-20 E-22 E-30 E-31 E-32
#
# Regla de negocio -> escenario que la verifica
#   RB-01 titulo obligatorio de 1 a 200 caracteres       E-05 E-06 E-07 E-08 E-09
#                                                       E-27 E-28
#   RB-02 codigo unico por proyecto y no editable         E-01 E-04 E-12 E-13 E-14
#   RB-03 prioridad del conjunto cerrado, default Media   E-02 E-10 E-27
#   RB-04 estado del conjunto cerrado, default Backlog    E-02 E-11 E-24 E-27
#   RB-05 descripcion opcional                            E-03 E-24 E-25
#   RB-06 historia dentro de un proyecto existente        E-01 E-15 E-18 E-19 E-20
#                                                       E-22 E-30 E-31 E-32
#   RB-07 id y codigo no se aceptan del cliente           E-04
#   RB-08 el PATCH exige al menos un campo                E-26
#   RB-09 el PATCH valida prioridad y estado              E-24 E-25 E-27
#   RB-10 el PATCH conserva el titulo 1..200              E-27 E-28
#
# Caso limite -> escenario
#   CL-01 E-06   CL-02 E-07   CL-03 E-05   CL-04 E-05, E-08, E-28
#   CL-05 E-05, E-09   CL-06 E-05   CL-07 E-02   CL-08 E-03
#   CL-09 E-10   CL-10 E-11   CL-11 E-12   CL-12 E-13   CL-13 E-19
#   CL-14 E-26   CL-15 E-24   CL-16 E-29   CL-17 E-14   CL-18 E-04
#   CL-19 E-32
#
# Escenario de la issue #29 -> identificador de escenario en este archivo
#   Caso normal: agregar historia        E-01
#   Caso normal: listar historias        E-18
#   Caso normal: modificar historia      E-23
#   Error: titulo vacio                  E-06
#   Error: prioridad invalida            E-10
#   Error: estado invalido               E-11
#   Caso limite: ID unico por proyecto   E-12, E-13, E-14
#
# Endpoint -> escenarios
#   POST   /api/proyectos/{id}/historias                E-01 a E-17
#   GET    /api/proyectos/{id}/historias                E-18 a E-22
#   PATCH  /api/proyectos/{id}/historias/{historiaId}   E-23 a E-34
#
# -----------------------------------------------------------------------------
# SUGERENCIA DE AUTOMATIZACION (TDD)
# -----------------------------------------------------------------------------
# Los pasos de este archivo se replican como subtests t.Run() con el
# identificador E-nn como nombre, para que el reporte de go test y el reporte de
# BDD sean comparables:
#   internal/domain/historia_test.go            Prioridad.Valido, Estado.Valido,
#                                               ValidarTitulo y FormatearCodigo
#                                               (funciones puras de RB-01 a RB-04)
#   internal/service/historia_service_test.go   reglas RB-01 a RB-10 con mocks
#                                               de HistoriaRepository y
#                                               ProyectoRepository (testify/mock)
#   internal/handler/historia_handler_test.go   codigos HTTP, cuerpos de error,
#                                               parseo de path params y binding
#                                               del body con httptest
#   internal/repository/historia_repository_test.go
#                                               generacion del codigo consecutivo
#                                               y respeto del indice unico
#                                               (proyecto_id, codigo)
#   internal/repository/migraciones_test.go     AutoMigrate idempotente y creacion
#                                               de la tabla "historias"
#   cmd/api/router_test.go                      las 3 rutas de
#                                               /api/proyectos/{id}/historias
# Para la suite end to end con godog, cada Regla de este archivo se asocia al
# servicio HTTP de prueba correspondiente: "HistoriaService" con el router real
# y una base de datos de prueba.
# =============================================================================
