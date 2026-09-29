# language: es
# =============================================================================
# US-01 - Creacion y gestion de proyectos
# Escenarios BDD (Gherkin) del modulo de gestion de proyectos del
# Software Metrics & Estimation Engine.
#
# Issue de origen     : SebastianAguilera/ingenieria-y-calidad#28
# Especificacion SDD  : docs/specs/US-01-gestion-de-proyectos-spec.md
# Metodologia         : docs/methodology.md seccion 4 (BDD con archivos .feature)
# Requerimiento       : docs/proyecto.md, requerimiento 1
# Rama de trabajo     : feature/US-01-gestion-proyectos
# Trazabilidad        : Issue #28 -> SDD -> Criterios de aceptacion -> Escenarios
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
#   {id}, {integranteId} = path param segun la notacion de la spec SDD.
#   <nombre>, <codigo>   = placeholder de Gherkin (Esquema del Escenario).
#   "campo"              = nombre de campo del JSON, entrecomillado para que el
#                          paso sea inequivoco al traducirlo a un test Go.
#   E-nn                 = identificador estable del escenario; se reutiliza
#                          como nombre del subtest t.Run() en los *_test.go.
#
# Etiquetas:
#   @caso_normal        comportamiento principal o recorrido de extremo a
#                       extremo.
#   @caso_alternativo   camino valido distinto del principal.
#   @caso_limite        borde de los valores permitidos: longitudes, fechas,
#                       valores por defecto y transiciones idempotentes.
#   @caso_error         rechazo esperado: 400 / 404 / 409 / 422 / 500.
#   @rfc                verificacion de un estandar de formato de fecha:
#                       ISO 8601 "YYYY-MM-DD" en fechas de calendario y
#                       RFC 3339 UTC en creado_en / actualizado_en.
#   @escenario_diferido comportamiento definido en la spec que solo puede
#                       observarse cuando exista la US que cree la entidad
#                       dependiente; deja constancia de la cobertura prevista.
#
# Entorno de ejecucion (variables resueltas antes de correr los escenarios):
#   BASE_URL = http://localhost:$PORT   (PORT segun .env; 8080 en .env.example
#                                       y 8081 si la variable no esta definida)
#   DB       = PostgreSQL 16 de docker-compose, publicado en el host en 5433
#               (DB_PORT=5433 para la API en el host y 5432 dentro de la red de
#               contenedores, ver Restriccion 13 de la spec SDD).
#   Ejemplo curl:
#     curl -i -X POST "$BASE_URL/api/proyectos" \
#       -H "Content-Type: application/json" -d @proyecto.json
#   Ejemplo httptest: los mismos pasos se ejecutan contra el *gin.Engine real en
#     internal/handler/proyecto_handler_test.go y cmd/api/router_test.go.
#
# -----------------------------------------------------------------------------
# NARRATIVA (US-01, issue #28)
# -----------------------------------------------------------------------------
# Como Agile Enabler, quiero crear un proyecto con su nombre, sus fechas
# (inicio y fin estimado) y sus integrantes, y luego poder consultarlo,
# modificarlo, cambiar su estado, ajustar su composicion y darlo de baja, para
# tener la estructura base del sistema sobre la que se construyen el backlog,
# los sprints, el registro de esfuerzo, los defectos y las metricas.
#
# -----------------------------------------------------------------------------
# CRITERIOS DE ACEPTACION DE LA ISSUE #28
# -----------------------------------------------------------------------------
# [CA-I1] Permite ingresar nombre, fecha de inicio y fecha estimada de fin.
#                                                             -> E-01, E-04
# [CA-I2] Permite asociar una lista de integrantes al proyecto.
#                                                          -> E-01, E-17, E-18
# [CA-I3] Valida que la fecha de fin no sea anterior a la de inicio.
#                                        -> E-04 (acepta igual), E-11 (rechaza)
# [CA-I4] El nombre es obligatorio y no puede estar vacio ni ser solo espacios.
#                                                            -> E-08, E-09
# [CA-I5] El ID es unico y autogenerado por el sistema.        -> E-01, E-05
# [CA-I6] La fecha de inicio es obligatoria y la de fin es opcional al crear.
#                                                        -> E-02, E-10, E-32
# [CA-I7] La lista de integrantes puede estar vacia y modificarse luego.
#                                                    -> E-03, E-40, E-46
# [CA-I8] Las fechas se persisten en formato ISO 8601 "YYYY-MM-DD".
#                                                        -> E-12, E-22, E-58
# [CA-I9] El proyecto expone un estado Activo / Cerrado.  -> E-33, E-34, E-36
#
# -----------------------------------------------------------------------------
# REGLAS DE NEGOCIO CUBIERTAS
# -----------------------------------------------------------------------------
# RB-01 nombre obligatorio de 1 a 150 caracteres tras TrimSpace
#                                -> E-06, E-07, E-08, E-09, E-28, E-32, E-56
# RB-02 fecha de inicio obligatoria         -> E-10, E-19, E-32
# RB-03 fecha de fin no anterior a la de inicio
#                          -> E-04, E-11, E-19, E-27, E-29, E-32
# RB-04 todo proyecto nace Activo           -> E-01, E-05
# RB-05 estado solo "Activo" o "Cerrado"     -> E-36
# RB-06 cerrar exige fecha de fin informada -> E-33, E-37
# RB-07 nombre y email validos por integrante
#                          -> E-14, E-15, E-16, E-42, E-43, E-57
# RB-08 el mismo email no se repite en un proyecto
#                                            -> E-17, E-44, E-45
# RB-09 un email identifica a una sola persona -> E-18, E-41
# RB-10 proyecto cerrado: solo cambia su estado -> E-31, E-32, E-48
# RB-11 no se da de baja con historial asociado -> E-52
# RB-12 no se quita un integrante no asociado    -> E-47
#
# -----------------------------------------------------------------------------
# CASOS LIMITE CUBIERTOS
# -----------------------------------------------------------------------------
# CL-01 nombre solo con espacios o tabuladores            -> E-09
# CL-02 nombre de 150 caracteres                          -> E-06
# CL-03 nombre de 151 caracteres                          -> E-07
# CL-04 fecha_fin igual a fecha_inicio                    -> E-04, E-27
# CL-05 fecha_fin ausente, null o cadena vacia            -> E-02
# CL-06 alta sin el campo integrantes                     -> E-03
# CL-07 integrantes null o lista vacia                    -> E-03
# CL-08 email duplicado con otra capitalizacion en el alta -> E-17
# CL-09 email ya asociado con otra capitalizacion         -> E-44
# CL-10 email existente en otro proyecto                  -> E-18
# CL-11 email con la forma "Nombre <correo>"              -> E-14, E-42
# CL-12 listado sin proyectos                             -> E-21
# CL-13 DELETE de un proyecto ya dado de baja             -> E-50
# CL-14 estado con otra capitalizacion                    -> E-36
# CL-15 transicion al mismo estado                        -> E-35
# CL-16 integrantes sobre un proyecto dado de baja        -> E-53
# CL-17 dos peticiones concurrentes con el mismo email     -> E-45
# CL-18 falla la persistencia de un integrante en el alta -> E-54
#
# -----------------------------------------------------------------------------
# ENDPOINTS CUBIERTOS (los 8 de la spec SDD, ninguno mas)
# -----------------------------------------------------------------------------
#   1. POST   /api/proyectos                              -> E-01 a E-19, E-54
#   2. GET    /api/proyectos                              -> E-01, E-20, E-21
#   3. GET    /api/proyectos/{id}                         -> E-01, E-22, E-23
#   4. PUT    /api/proyectos/{id}                         -> E-25 a E-32, E-48
#   5. PATCH  /api/proyectos/{id}/estado                  -> E-33 a E-39
#   6. DELETE /api/proyectos/{id}                         -> E-49 a E-53
#   7. POST   /api/proyectos/{id}/integrantes             -> E-18, E-40 a E-45,
#                                                          E-48, E-53
#   8. DELETE /api/proyectos/{id}/integrantes/{integranteId}
#                                                          -> E-46 a E-48, E-53
# =============================================================================

Funcionalidad: US-01 - Creacion y gestion de proyectos
  Como Agile Enabler
  Quiero crear un proyecto con nombre, fecha de inicio, fecha de fin estimada
  e integrantes iniciales, y luego consultarlo, modificarlo, cambiar su estado,
  ajustar su composicion y darlo de baja
  Para disponer de la estructura base del sistema sobre la que se construyen
  el backlog, los sprints, el registro de esfuerzo, los defectos y las metricas
  de calidad de software.

  Todo proyecto se identifica con un id autogenerado por el sistema, nace en
  estado Activo y se da de baja logicamente mediante borrado_en, de modo que la
  baja nunca destruye el historial que usan las metricas (decisiones D-04 y D-07
  de la spec SDD). No existe autenticacion en US-01: el actor es implicitamente
  un administrador del sistema.

  Contexto:
    Dado que la API esta levantada y responde "200 OK" en "GET /health"
    Y que la base de datos "metrics_db" fue migrada en el arranque por
      "AutoMigrate" y existen las tablas "proyectos", "integrantes" y
      "proyecto_integrantes"
    Y que la suite arranca de una base limpia, con estas sentencias ejecutadas
      por el runner antes de cada escenario:
      """
      DELETE FROM proyecto_integrantes;
      DELETE FROM integrantes;
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
  # ALTA DE PROYECTOS: POST /api/proyectos
  # ===========================================================================
  Regla: Alta de un proyecto con nombre, fechas e integrantes iniciales

    Escenario: E-01 - Alta de un proyecto con nombre, fechas e integrantes
      # Caso normal de la issue #28 (CA-I1, CA-I2, CA-I5, CA-I9).
      # Cubre RB-04 y los criterios de aceptacion de alta, listado y detalle.
      @caso_normal
      Dado que preparo el cuerpo JSON del alta:
        """
        {
          "nombre": "Software Metrics & Estimation",
          "fecha_inicio": "2026-09-28",
          "fecha_fin": "2026-12-18",
          "integrantes": [
            { "nombre": "Aguilera Sebastián", "email": "sebas.aguilera@utn.edu.ar" },
            { "nombre": "Choquevillca Celeste", "email": "celeste.choque@utn.edu.ar" }
          ]
        }
        """
      Cuando envío "POST" a "/api/proyectos" con ese cuerpo
      Entonces la API responde "201 Created" con el proyecto persistido
      Y el campo "id" es un entero positivo generado por el sistema
      Y el campo "nombre" es "Software Metrics & Estimation"
      Y el campo "fecha_inicio" es "2026-09-28" y el campo "fecha_fin" es "2026-12-18"
      Y el campo "estado" es "Activo"
      Y el campo "cantidad_integrantes" es 2
      Y el array "integrantes" contiene los emails "sebas.aguilera@utn.edu.ar" y
        "celeste.choque@utn.edu.ar", cada uno con su "id" y su "nombre"
      Y los campos "creado_en" y "actualizado_en" tienen formato RFC 3339 UTC
      Y cuando consulto "GET" en "/api/proyectos"
      Entonces la API responde "200 OK" con el campo "total" igual a 1
      Y el proyecto aparece en el array "proyectos" con "cantidad_integrantes" igual a 2
      Y el listado expone "cantidad_integrantes" pero no el array "integrantes" (decision D-05)
      Y cuando consulto "GET" en "/api/proyectos/{id}" con el id devuelto en el alta
      Entonces la API responde "200 OK" y reproduce las mismas fechas en formato "YYYY-MM-DD"
      Y el array "integrantes" del detalle tiene 2 elementos y coincide con el del alta

    @caso_alternativo @caso_limite
    Esquema del Escenario: E-02 - Alta con el campo "fecha_fin" en <variante_fecha_fin>
      # CA-I6 y CL-05: la fecha de fin es opcional al crear y puede definirse
      # despues. Las tres variantes se interpretan como "sin fecha de fin".
      Dado que preparo el cuerpo JSON del alta:
        """
        {
          "nombre": "Portal de Autoatencion",
          "fecha_inicio": "2026-10-01"
          <linea_fecha_fin>
        }
        """
      Cuando envío "POST" a "/api/proyectos" con ese cuerpo
      Entonces la API responde "201 Created"
      Y el campo "fecha_fin" de la respuesta es "null"
      Y el campo "estado" es "Activo"
      Y el campo "cantidad_integrantes" es 0
      Y cuando consulto "GET" en "/api/proyectos/{id}" con el id devuelto en el alta
      Entonces la API responde "200 OK" y el campo "fecha_fin" sigue siendo "null"
      Y el proyecto aparece en el listado sin error de formato de fecha

      Ejemplos:
        | variante_fecha_fin | linea_fecha_fin          |
        | ausente            |                          |
        | nula               | , "fecha_fin": null      |
        | cadena vacia       | , "fecha_fin": ""        |

    @caso_alternativo @caso_limite
    Esquema del Escenario: E-03 - Alta con el campo "integrantes" en <variante_integrantes>
      # CA-I7, CL-06 y CL-07: la lista de integrantes puede estar vacia y la
      # respuesta nunca devuelve "null" en el array.
      Dado que preparo el cuerpo JSON del alta:
        """
        {
          "nombre": "Proyecto sin equipo inicial",
          "fecha_inicio": "2026-10-05",
          "fecha_fin": "2026-11-30"
          <linea_integrantes>
        }
        """
      Cuando envío "POST" a "/api/proyectos" con ese cuerpo
      Entonces la API responde "201 Created"
      Y el campo "cantidad_integrantes" es 0
      Y el campo "integrantes" es el array vacio y no "null"
      Y el proyecto queda disponible para asociarle integrantes despues

      Ejemplos:
        | variante_integrantes | linea_integrantes              |
        | lista vacia          | , "integrantes": []             |
        | nula                 | , "integrantes": null           |
        | ausente              |                                |

    @caso_limite
    Escenario: E-04 - Proyecto de un solo dia con fecha_fin igual a fecha_inicio
      # CA-I3 y CL-04: RB-03 compara con Before, por lo que la igualdad es
      # valida. Es el caso limite explicito de la issue #28.
      Dado que preparo el cuerpo JSON del alta:
        """
        {
          "nombre": "Proyecto de un solo dia",
          "fecha_inicio": "2026-11-02",
          "fecha_fin": "2026-11-02"
        }
        """
      Cuando envío "POST" a "/api/proyectos" con ese cuerpo
      Entonces la API responde "201 Created"
      Y el campo "fecha_inicio" es "2026-11-02" y el campo "fecha_fin" es "2026-11-02"
      Y el campo "estado" es "Activo"
      Y el proyecto queda persistido y visible en "GET /api/proyectos"

    @caso_alternativo
    Escenario: E-05 - El cliente no puede elegir el id ni el estado
      # CA-I5 y RB-04: el id lo genera el sistema y todo proyecto nace Activo,
      # sin importar el body. Los campos desconocidos se ignoran en silencio.
      Dado que preparo el cuerpo JSON del alta:
        """
        {
          "id": 999,
          "nombre": "Proyecto con campos ajenos",
          "fecha_inicio": "2026-09-28",
          "fecha_fin": "2026-12-18",
          "estado": "Cerrado",
          "borrado_en": "2026-09-28T10:00:00Z",
          "creado_en": "2020-01-01T00:00:00Z",
          "cliente_id": "no-existe",
          "integrantes": [
            { "nombre": "Aguilera Sebastián", "email": "sebas.aguilera@utn.edu.ar" }
          ]
        }
        """
      Cuando envío "POST" a "/api/proyectos" con ese cuerpo
      Entonces la API responde "201 Created"
      Y el campo "id" de la respuesta no es 999, es el generado por el sistema
      Y el campo "estado" de la respuesta es "Activo" y no "Cerrado"
      Y el campo "fecha_fin" de la respuesta es "2026-12-18"
      Y el campo "creado_en" refleja el momento del alta y no "2020-01-01T00:00:00Z"
      Y el campo "borrado_en" no se expone en la respuesta del recurso
      Y el campo desconocido "cliente_id" se ignora sin provocar error

    @caso_limite
    Escenario: E-06 - Nombre del proyecto con exactamente 150 caracteres
      # CL-02: 150 es el maximo de la columna varchar(150) y se acepta.
      # El nombre enviado es la cadena "ABCDE" repetida 30 veces, es decir
      # 150 caracteres exactos, y se construye en el test con
      # strings.Repeat("ABCDE", 30) o con una cadena de 150 caracteres.
      Dado que preparo el cuerpo JSON del alta con el campo "nombre" formado por
        la cadena "ABCDE" repetida 30 veces, es decir 150 caracteres exactos
      Y que las fechas del alta son "2026-09-28" y "2026-12-18"
      Cuando envío "POST" a "/api/proyectos" con ese cuerpo
      Entonces la API responde "201 Created"
      Y el campo "nombre" de la respuesta tiene exactamente 150 caracteres
      Y el campo "nombre" de la respuesta es igual al enviado

    @caso_limite @caso_error
    Esquema del Escenario: E-07 - Matriz de borde del campo "nombre"
      # CL-01, CL-02 y CL-03 con CA-I4 y RB-01: esta tabla concentra en un solo
      # esquema los cuatro casos de borde exigidos para el nombre: vacio, solo
      # espacios, 150 caracteres y 151 caracteres.
      # El nombre de 150 caracteres es la cadena "ABCDE" repetida 30 veces
      # (strings.Repeat("ABCDE", 30)) y el de 151 es esa misma cadena seguida de
      # la letra "F" (strings.Repeat("ABCDE", 30) + "F"). E-06, E-08 y E-09
      # cubren cada caso por separado con su diagnostico completo.
      Dado que preparo el cuerpo JSON del alta con el campo "nombre" en
        <nombre_ingresado>
      Y que las fechas del alta son "2026-09-28" y "2026-12-18"
      Cuando envío "POST" a "/api/proyectos" con ese cuerpo
      Entonces la API responde "<codigo_http_esperado>"
      Y el cuerpo de la respuesta es <cuerpo_esperado>
      Y cuando consulto "GET" en "/api/proyectos"
      Entonces la API responde "200 OK" con el campo "total" igual a
        <total_esperado>

      Ejemplos:
        | nombre_ingresado                                                       | codigo_http_esperado       | cuerpo_esperado                                                                                                             | total_esperado |
        | la cadena vacia ""                                                     | "422 Unprocessable Entity" | el error de validacion con "VALIDACION", "el nombre del proyecto es obligatorio" y "campo" "nombre"                          | 0              |
        | la cadena de tres caracteres de espacio "   "                          | "422 Unprocessable Entity" | el error de validacion con "VALIDACION", "el nombre del proyecto es obligatorio" y "campo" "nombre"                          | 0              |
        | la cadena "ABCDE" repetida 30 veces                                    | "201 Created"              | el proyecto creado con "nombre" identico al enviado y "estado" "Activo"                                                      | 1              |
        | la cadena "ABCDE" repetida 30 veces seguida de la letra "F"             | "422 Unprocessable Entity" | el error de validacion con "VALIDACION", "el nombre del proyecto no puede superar los 150 caracteres" y "campo" "nombre"  | 0              |

    @caso_error
    Escenario: E-08 - Nombre del proyecto vacio
      # CA-I4, RB-01 y caso de error explicito de la issue #28: nombre vacio.
      Dado que preparo el cuerpo JSON del alta:
        """
        {
          "nombre": "",
          "fecha_inicio": "2026-09-28",
          "fecha_fin": "2026-12-18"
        }
        """
      Cuando envío "POST" a "/api/proyectos" con ese cuerpo
      Entonces la API responde "422 Unprocessable Entity"
      Y el campo "error" del cuerpo es "VALIDACION"
      Y el campo "mensaje" del cuerpo es "el nombre del proyecto es obligatorio"
      Y el campo "campo" del cuerpo es "nombre"
      Y cuando consulto "GET" en "/api/proyectos"
      Entonces la API responde "200 OK" con el campo "total" igual a 0

    @caso_limite @caso_error
    Esquema del Escenario: E-09 - Nombre del proyecto formado solo por espacios
      # CL-01: RB-01 valida sobre el resultado de TrimSpace, por lo que los
      # espacios y los tabuladores producen el mismo rechazo que el vacio.
      Dado que preparo el cuerpo JSON del alta con el campo "nombre" en
        <detalle_del_nombre> y las fechas "2026-09-28" y "2026-12-18"
      Cuando envío "POST" a "/api/proyectos" con ese cuerpo
      Entonces la API responde "422 Unprocessable Entity"
      Y el campo "error" del cuerpo es "VALIDACION"
      Y el campo "mensaje" del cuerpo es "el nombre del proyecto es obligatorio"
      Y el campo "campo" del cuerpo es "nombre"
      Y cuando consulto "GET" en "/api/proyectos"
      Entonces la API responde "200 OK" con el campo "total" igual a 0

      Ejemplos:
        | detalle_del_nombre                                       |
        | la cadena de tres caracteres de espacio                   |
        | la cadena de dos caracteres de tabulacion                |

    @caso_error
    Esquema del Escenario: E-10 - Alta sin fecha de inicio
      # CA-I6, RB-02 y caso de error explicito de la issue #28: la fecha de
      # inicio es obligatoria. El DTO deja pasar la cadena vacia y es el service
      # el que detecta el time.Time cero, con "campo": "fecha_inicio".
      Dado que preparo el cuerpo JSON del alta con el nombre "Proyecto sin inicio"
        y la "fecha_fin" "2026-12-18"
      Y que el campo "fecha_inicio" esta en <variante_fecha_inicio>, es decir
        <valor_literal_de_fecha_inicio>
      Cuando envío "POST" a "/api/proyectos" con ese cuerpo
      Entonces la API responde "422 Unprocessable Entity"
      Y el campo "error" del cuerpo es "VALIDACION"
      Y el campo "mensaje" del cuerpo es
        "la fecha de inicio del proyecto es obligatoria"
      Y el campo "campo" del cuerpo es "fecha_inicio"
      Y cuando consulto "GET" en "/api/proyectos"
      Entonces la API responde "200 OK" con el campo "total" igual a 0

      Ejemplos:
        | variante_fecha_inicio | valor_literal_de_fecha_inicio |
        | ausente               | el campo no se envia          |
        | cadena vacia          | la cadena ""                 |

    @caso_error
    Esquema del Escenario: E-11 - Fecha de fin anterior a la fecha de inicio
      # CA-I3, RB-03 y caso de error explicito de la issue #28. El borde
      # inmediato (un dia antes) y un salto largo producen el mismo rechazo.
      Dado que preparo el cuerpo JSON del alta con el nombre "Proyecto incoherente"
        y la "fecha_inicio" "2026-09-28"
      Y que la "fecha_fin" es "<fecha_fin_invalida>"
      Cuando envío "POST" a "/api/proyectos" con ese cuerpo
      Entonces la API responde "422 Unprocessable Entity"
      Y el campo "error" del cuerpo es "VALIDACION"
      Y el campo "mensaje" del cuerpo es
        "la fecha de fin no puede ser anterior a la fecha de inicio"
      Y el campo "campo" del cuerpo es "fecha_fin"
      Y cuando consulto "GET" en "/api/proyectos"
      Entonces la API responde "200 OK" con el campo "total" igual a 0

      Ejemplos:
        | fecha_fin_invalida |
        | 2026-09-27         |
        | 2025-01-01         |

    @caso_error @rfc
    Esquema del Escenario: E-12 - Fecha con un formato distinto de ISO 8601
      # CA-I8 y @rfc: el handler valida la capa sintactica con
      # time.Parse("2006-01-02") y responde 400 PARAMETRO_INVALIDO indicando
      # el campo que no pudo interpretarse.
      Dado que preparo el cuerpo JSON del alta con el nombre "Proyecto con fecha mal escrita"
        y la "fecha_fin" "2026-12-18"
      Y que la "fecha_inicio" es "<fecha_incorrecta>"
      Cuando envío "POST" a "/api/proyectos" con ese cuerpo
      Entonces la API responde "400 Bad Request"
      Y el campo "error" del cuerpo es "PARAMETRO_INVALIDO"
      Y el campo "campo" del cuerpo es "fecha_inicio"
      Y el cuerpo de la respuesta no incluye un mensaje de regla de negocio

      Ejemplos:
        | fecha_incorrecta |
        | 28/09/2026       |
        | 2026/09/28       |
        | 28-09-2026       |
        | 2026-13-01       |
        | 20260928         |

    @caso_error
    Esquema del Escenario: E-13 - Cuerpo del request mal formado
      # Convencion de la spec: un fallo de ShouldBindJSON produce
      # 400 JSON_INVALIDO, sin "campo", y ningun dato se persiste.
      Dado que envío un cuerpo <descripcion_del_cuerpo> en "POST /api/proyectos"
        con la cabecera "Content-Type: application/json"
      Entonces la API responde "400 Bad Request"
      Y el campo "error" del cuerpo es "JSON_INVALIDO"
      Y el cuerpo de la respuesta no tiene el campo "campo"
      Y cuando consulto "GET" en "/api/proyectos"
      Entonces la API responde "200 OK" con el campo "total" igual a 0

      Ejemplos:
        | descripcion_del_cuerpo                        |
        | vacio                                         |
        | con el JSON truncado, sin la llave de cierre  |
        | con "nombre" como numero en lugar de texto    |
        | con "integrantes" como texto en lugar de array |

    @caso_error @caso_limite
    Esquema del Escenario: E-14 - Email de integrante invalido en el alta
      # RB-07 y CL-11: domain.EmailValido exige que la direccion parseada
      # coincida exactamente con la entrada normalizada, por eso rechaza la
      # forma "Nombre <correo>". En el alta del proyecto el campo reportado es
      # "integrantes[].email".
      Dado que preparo el cuerpo JSON del alta con el nombre "Proyecto con email invalido"
        y las fechas "2026-09-28" y "2026-12-18"
      Y que el array "integrantes" tiene a "Perez Castro Jazmín" con el email
        "<email_invalido>"
      Cuando envío "POST" a "/api/proyectos" con ese cuerpo
      Entonces la API responde "422 Unprocessable Entity"
      Y el campo "error" del cuerpo es "VALIDACION"
      Y el campo "mensaje" del cuerpo es
        "el email del integrante es obligatorio y debe tener un formato válido"
      Y el campo "campo" del cuerpo es "integrantes[].email"
      Y cuando consulto "GET" en "/api/proyectos"
      Entonces la API responde "200 OK" con el campo "total" igual a 0

      Ejemplos:
        | email_invalido                                                                                                                                   |
        | jazmin.perez.utn.edu.ar                                                           |
        | jazmin.perez@utn.edu.ar@extra                                                     |
        | la cadena "jazmin perez@utn.edu.ar" con un espacio embebido                        |
        | la cadena "Perez Jazmin <jazmin.perez@utn.edu.ar>"                                 |
        | la cadena vacia                                                                   |
        | adalovel@abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij |

    @caso_limite
    Escenario: E-15 - Email de integrante con exactamente 150 caracteres
      # Limite superior de EmailValido y de la columna varchar(150): se
      # acepta. El email es "adalovel@" seguido de un dominio de 142 caracteres
      # formado por 13 etiquetas de 10 caracteres.
      Dado que preparo el cuerpo JSON del alta con el nombre "Proyecto con email al limite"
        y las fechas "2026-09-28" y "2026-12-18"
      Y que el array "integrantes" tiene a "Perez Castro Jazmín" con el email
        "adolove@abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij"
      Cuando envío "POST" a "/api/proyectos" con ese cuerpo
      Entonces la API responde "201 Created"
      Y el campo "cantidad_integrantes" es 1
      Y el email persistido tiene exactamente 150 caracteres
      Y el email persistido es identico al enviado

    @caso_error @caso_limite
    Esquema del Escenario: E-16 - Nombre de integrante invalido en el alta
      # RB-07: TrimSpace(nombre) distinto de vacio y entre 1 y 100 caracteres.
      # En el alta del proyecto el campo reportado es "integrantes[].nombre".
      Dado que preparo el cuerpo JSON del alta con el nombre "Proyecto con integrante sin nombre"
        y las fechas "2026-09-28" y "2026-12-18"
      Y que el array "integrantes" tiene a <nombre_invalido> con el email
        "jazmin.perez@utn.edu.ar"
      Cuando envío "POST" a "/api/proyectos" con ese cuerpo
      Entonces la API responde "422 Unprocessable Entity"
      Y el campo "error" del cuerpo es "VALIDACION"
      Y el campo "mensaje" del cuerpo es "el nombre del integrante es obligatorio"
      Y el campo "campo" del cuerpo es "integrantes[].nombre"
      Y cuando consulto "GET" en "/api/proyectos"
      Entonces la API responde "200 OK" con el campo "total" igual a 0

      Ejemplos:
        | nombre_invalido                        |
        | la persona con la cadena de nombre ""  |
        | la persona con la cadena de nombre "  "|

        # La spec no define un mensaje propio para un nombre de mas de 100
        # caracteres en el integrante, por lo que ese caso no se cubre con un
        # codigo HTTP inventado: se documenta en el informe de cobertura.

    @caso_error @caso_limite
    Escenario: E-17 - Dos integrantes con el mismo email y distinta capitalizacion
      # CL-08 y RB-08: la identidad del integrante es el email normalizado, por
      # eso se rechaza con 409 y no con 400 (decision D-01).
      Dado que preparo el cuerpo JSON del alta con el nombre "Proyecto con integrantes repetidos"
        y las fechas "2026-09-28" y "2026-12-18"
      Y que el array "integrantes" tiene a dos personas con los emails
        "Ada@utn.edu.ar" y "ada@utn.edu.ar"
      Cuando envío "POST" a "/api/proyectos" con ese cuerpo
      Entonces la API responde "409 Conflict"
      Y el campo "error" del cuerpo es "INTEGRANTE_DUPLICADO"
      Y el campo "mensaje" del cuerpo es "el integrante ya forma parte del proyecto"
      Y el campo "campo" del cuerpo es "email"
      Y cuando consulto "GET" en "/api/proyectos"
      Entonces la API responde "200 OK" con el campo "total" igual a 0

    @caso_alternativo
    Escenario: E-18 - Un email que ya existe en otro proyecto se reutiliza
      # CA-I2, CA-I7, CL-10 y RB-09: la persona se identifica por email
      # normalizado y unico en todo el sistema. Se crea el vinculo con 201 y se
      # actualiza el nombre del registro existente.
      Dado que ya existe el proyecto con "id" 1 y el integrante
        "Aguilera Sebastián" con el email "ada@utn.edu.ar" y el "id" 10
      Y que preparo el cuerpo JSON del alta:
        """
        {
          "nombre": "Segundo proyecto del equipo",
          "fecha_inicio": "2026-09-28",
          "fecha_fin": "2026-12-18",
          "integrantes": [
            { "nombre": "Aguilera Sebastian", "email": "ADA@utn.edu.ar" }
          ]
        }
        """
      Cuando envío "POST" a "/api/proyectos" con ese cuerpo
      Entonces la API responde "201 Created"
      Y el array "integrantes" del proyecto nuevo tiene 1 elemento
      Y ese elemento conserva el "id" 10 del integrante ya existente
      Y su "nombre" paso a ser "Aguilera Sebastian"
      Y su "email" quedo normalizado en minusculas como "ada@utn.edu.ar"
      Y cuando consulto "GET" en "/api/proyectos/1"
      Entonces la API responde "200 OK" y el proyecto original conserva su integrante
      Y los dos proyectos comparten la misma persona y por lo tanto el mismo "id" de integrante

    @caso_error
    Esquema del Escenario: E-19 - Orden de evaluacion de las reglas en el alta
      # La spec fija el orden RB-01, luego RB-02, luego RB-03, luego RB-07,
      # luego RB-08 y por ultimo RB-04, para que ante varias entradas invalidas
      # el error devuelto sea determinista. Cada fila varia el defecto que debe
      # ganar; el resto de los campos invalidos acompana a la fila.
      Dado que preparo el cuerpo JSON del alta con el nombre "<nombre_ingresado>"
      Y que la "fecha_inicio" es <fecha_inicio_ingresada>
      Y que la "fecha_fin" es <fecha_fin_ingresada>
      Y que el array "integrantes" tiene a "Perez Castro Jazmín" con el email <email_ingresado>
      Cuando envío "POST" a "/api/proyectos" con ese cuerpo
      Entonces la API responde "422 Unprocessable Entity"
      Y el campo "error" del cuerpo es "VALIDACION"
      Y el campo "mensaje" del cuerpo es "<mensaje_esperado>"
      Y el campo "campo" del cuerpo es "<campo_esperado>"
      Y cuando consulto "GET" en "/api/proyectos"
      Entonces la API responde "200 OK" con el campo "total" igual a 0

      Ejemplos:
        | nombre_ingresado | fecha_inicio_ingresada | fecha_fin_ingresada | email_ingresado | mensaje_esperado                                                      | campo_esperado      |
        | la cadena ""     | "2026-09-28"           | "2026-09-27"       | invalido         | el nombre del proyecto es obligatorio                                   | nombre              |
        | "Proyecto"       | la cadena ""           | "2026-09-27"       | invalido         | la fecha de inicio del proyecto es obligatoria                          | fecha_inicio        |
        | "Proyecto"       | "2026-09-28"           | "2026-09-27"       | invalido         | la fecha de fin no puede ser anterior a la fecha de inicio             | fecha_fin           |
        | "Proyecto"       | "2026-09-28"           | "2026-12-18"       | invalido         | el email del integrante es obligatorio y debe tener un formato válido   | integrantes[].email |

  # ===========================================================================
  # CONSULTA: GET /api/proyectos y GET /api/proyectos/{id}
  # ===========================================================================
  Regla: Consulta del listado y del detalle de un proyecto

    @caso_normal
    Escenario: E-20 - Listado con varios proyectos
      # Criterio de aceptacion: GET /api/proyectos lista los proyectos no dados
      # de baja y devuelve {"total": n, "proyectos": [...]}.
      Dado que existen dos proyectos no dados de baja: el de "id" 1 con 2
        integrantes y el de "id" 2 sin integrantes
      Y que ninguno de los dos fue dado de baja logicamente
      Cuando consulto "GET" en "/api/proyectos"
      Entonces la API responde "200 OK"
      Y el campo "total" es 2
      Y el array "proyectos" tiene 2 elementos
      Y cada elemento expone "id", "nombre", "fecha_inicio", "fecha_fin",
        "estado", "cantidad_integrantes", "creado_en" y "actualizado_en"
      Y el primer elemento tiene "cantidad_integrantes" igual a 2
      Y el primer elemento tiene "fecha_fin" igual a "2026-12-18"
      Y el segundo elemento tiene "cantidad_integrantes" igual a 0
      Y el segundo elemento tiene "fecha_fin" igual a "null"
      Y ningun elemento del listado incluye el array "integrantes" (decision D-05)

    @caso_limite
    Escenario: E-21 - Listado sin ningun proyecto visible
      # CL-12: el array vacio se serializa explicitamente para no producir null.
      Dado que no existe ningun proyecto activo ni dado de baja en el sistema
      Cuando consulto "GET" en "/api/proyectos"
      Entonces la API responde "200 OK"
      Y el campo "total" es 0
      Y el campo "proyectos" es el array vacio y no "null"
      Y el cuerpo de la respuesta es, en su forma equivalente:
        """
        { "total": 0, "proyectos": [] }
        """

    @caso_normal @rfc
    Escenario: E-22 - Detalle de un proyecto con su composicion de equipo
      # Criterio de aceptacion: el detalle devuelve el array "integrantes" y una
      # "cantidad_integrantes" coherente. Con @rfc se verifica la serializacion
      # de fechas de la convencion de la spec: ISO 8601 en las fechas de
      # calendario y RFC 3339 UTC en las fechas de auditoria.
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1,
        "fecha_inicio" "2026-09-28", "fecha_fin" "2026-12-18" y 3 integrantes
      Cuando consulto "GET" en "/api/proyectos/1"
      Entonces la API responde "200 OK"
      Y el campo "cantidad_integrantes" es 3
      Y el array "integrantes" tiene exactamente 3 elementos
      Y la longitud del array "integrantes" es igual al campo "cantidad_integrantes"
      Y cada elemento de "integrantes" expone "id", "nombre" y "email"
      Y el campo "fecha_inicio" responde con el patron de una fecha ISO 8601
        "YYYY-MM-DD" y no incluye hora ni zona horaria
      Y el campo "fecha_fin" responde con el patron de una fecha ISO 8601
        "YYYY-MM-DD" y no incluye hora ni zona horaria
      Y el campo "creado_en" responde con el patron RFC 3339 UTC
      Y el campo "actualizado_en" responde con el patron RFC 3339 UTC

    @caso_error
    Escenario: E-23 - Detalle de un proyecto inexistente
      Dado que no existe ningun proyecto con el "id" 999
      Cuando consulto "GET" en "/api/proyectos/999"
      Entonces la API responde "404 Not Found"
      Y el campo "error" del cuerpo es "PROYECTO_NO_ENCONTRADO"
      Y el campo "mensaje" del cuerpo es "el proyecto no existe"
      Y el cuerpo de la respuesta no tiene el campo "campo"

    @caso_error
    Esquema del Escenario: E-24 - Identificador de proyecto no valido
      # Un fallo de strconv.ParseUint o un resultado no positivo producen
      # 400 PARAMETRO_INVALIDO antes de consultar la base de datos.
      Dado que existe al menos un proyecto dado de alta
      Cuando consulto "GET" en "/api/proyectos/<id_invalido>"
      Entonces la API responde "400 Bad Request"
      Y el campo "error" del cuerpo es "PARAMETRO_INVALIDO"
      Y el cuerpo de la respuesta no tiene el campo "campo"

      Ejemplos:
        | id_invalido   |
        | la cadena abc |
        | el numero 0   |
        | el numero -1  |

  # ===========================================================================
  # MODIFICACION: PUT /api/proyectos/{id}
  # ===========================================================================
  Regla: Modificacion de nombre y fechas de un proyecto activo

    @caso_normal
    Escenario: E-25 - Modificacion del nombre y de la fecha de fin
      # Criterio de aceptacion: PUT modifica nombre y fechas y responde 200 con
      # el recurso actualizado.
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1,
        "fecha_inicio" "2026-09-28", "fecha_fin" "2026-12-18" y 2 integrantes
      Y que preparo el cuerpo JSON de la modificacion:
        """
        {
          "nombre": "Software Metrics & Estimation Engine",
          "fecha_inicio": "2026-09-28",
          "fecha_fin": "2026-12-20"
        }
        """
      Cuando envío "PUT" a "/api/proyectos/1" con ese cuerpo
      Entonces la API responde "200 OK" con el proyecto completo
      Y el campo "id" sigue siendo 1
      Y el campo "nombre" es "Software Metrics & Estimation Engine"
      Y el campo "fecha_inicio" es "2026-09-28" y el campo "fecha_fin" es "2026-12-20"
      Y el campo "estado" sigue siendo "Activo"
      Y el campo "cantidad_integrantes" sigue siendo 2
      Y el campo "actualizado_en" es posterior al "creado_en" del proyecto
      Y cuando consulto "GET" en "/api/proyectos/1"
      Entonces la API responde "200 OK" y devuelve los mismos valores modificados

    @caso_alternativo
    Escenario: E-26 - Modificacion que borra la fecha de fin e ignora el body de integrantes
      # Caso alternativo: la fecha de fin pasa a null. Ademas, por la decision
      # D-06 el body del PUT no incluye integrantes, por lo que un campo
      # "integrantes" enviado por el cliente se ignora y la composicion del
      # equipo no cambia.
      Dado que existe el proyecto "Proyecto A" con "id" 1, "fecha_inicio"
        "2026-09-28", "fecha_fin" "2026-12-18" y 2 integrantes
      Y que preparo el cuerpo JSON de la modificacion:
        """
        {
          "nombre": "Proyecto A sin fecha de fin",
          "fecha_inicio": "2026-09-28",
          "fecha_fin": null,
          "integrantes": [
            { "nombre": "Intruso Uno", "email": "intruso1@utn.edu.ar" },
            { "nombre": "Intruso Dos", "email": "intruso2@utn.edu.ar" }
          ]
        }
        """
      Cuando envío "PUT" a "/api/proyectos/1" con ese cuerpo
      Entonces la API responde "200 OK"
      Y el campo "fecha_fin" de la respuesta es "null"
      Y el campo "cantidad_integrantes" sigue siendo 2
      Y el array "integrantes" no contiene a "intruso1@utn.edu.ar" ni a "intruso2@utn.edu.ar"
      Y la composicion del equipo solo puede cambiarse por los endpoints de integrantes (D-06)

    @caso_limite
    Escenario: E-27 - Modificacion con fecha de fin igual a la fecha de inicio
      # CL-04 aplicado a la modificacion: la igualdad de fechas sigue siendo
      # valida porque RB-03 compara con Before.
      Dado que existe el proyecto "Proyecto B" con "id" 2 y "fecha_inicio" "2026-09-28"
      Y que preparo el cuerpo JSON de la modificacion:
        """
        {
          "nombre": "Proyecto B de un solo dia",
          "fecha_inicio": "2026-09-28",
          "fecha_fin": "2026-09-28"
        }
        """
      Cuando envío "PUT" a "/api/proyectos/2" con ese cuerpo
      Entonces la API responde "200 OK"
      Y el campo "fecha_inicio" de la respuesta es "2026-09-28"
      Y el campo "fecha_fin" de la respuesta es "2026-09-28"
      Y el campo "nombre" de la respuesta es "Proyecto B de un solo dia"

    @caso_error @caso_limite
    Esquema del Escenario: E-28 - Modificacion con un nombre invalido
      # RB-01 y ErrNombreDemasiadoLargo se verifican tambien en el PUT, y el
      # proyecto debe conservar sus datos originales.
      Dado que existe el proyecto "Proyecto C" con "id" 3 y "nombre" "Proyecto C original"
      Y que preparo el cuerpo JSON de la modificacion con el nombre "<nombre_ingresado>"
        y las fechas "2026-09-28" y "2026-12-18"
      Cuando envío "PUT" a "/api/proyectos/3" con ese cuerpo
      Entonces la API responde "422 Unprocessable Entity"
      Y el campo "error" del cuerpo es "VALIDACION"
      Y el campo "mensaje" del cuerpo es "<mensaje_esperado>"
      Y el campo "campo" del cuerpo es "nombre"
      Y cuando consulto "GET" en "/api/proyectos/3"
      Entonces la API responde "200 OK" y el nombre original no cambio

      Ejemplos:
        | nombre_ingresado                                                                                                                                       | mensaje_esperado                                          |
        | la cadena vacia                                                                                                                                         | el nombre del proyecto es obligatorio                      |
        | la cadena de cuatro espacios                                                                                                                            | el nombre del proyecto es obligatorio                      |
        | la cadena "ABCDE" repetida 30 veces seguida de la letra "F", con 151 caracteres                                  | el nombre del proyecto no puede superar los 150 caracteres |

    @caso_error
    Escenario: E-29 - Modificacion con fecha de fin anterior a la de inicio
      # RB-03 en el PUT: el rechazo deja el proyecto con sus datos originales.
      Dado que existe el proyecto "Proyecto D" con "id" 4, "fecha_inicio"
        "2026-09-28" y "fecha_fin" "2026-12-18"
      Y que preparo el cuerpo JSON de la modificacion:
        """
        {
          "nombre": "Proyecto D",
          "fecha_inicio": "2026-09-28",
          "fecha_fin": "2026-09-01"
        }
        """
      Cuando envío "PUT" a "/api/proyectos/4" con ese cuerpo
      Entonces la API responde "422 Unprocessable Entity"
      Y el campo "error" del cuerpo es "VALIDACION"
      Y el campo "mensaje" del cuerpo es
        "la fecha de fin no puede ser anterior a la fecha de inicio"
      Y el campo "campo" del cuerpo es "fecha_fin"
      Y cuando consulto "GET" en "/api/proyectos/4"
      Entonces la API responde "200 OK" y conserva la "fecha_fin" "2026-12-18"

    @caso_error
    Escenario: E-30 - Modificacion de un proyecto inexistente
      Dado que no existe ningun proyecto con el "id" 999
      Y que preparo el cuerpo JSON de la modificacion:
        """
        {
          "nombre": "Proyecto inexistente",
          "fecha_inicio": "2026-09-28",
          "fecha_fin": "2026-12-18"
        }
        """
      Cuando envío "PUT" a "/api/proyectos/999" con ese cuerpo
      Entonces la API responde "404 Not Found"
      Y el campo "error" del cuerpo es "PROYECTO_NO_ENCONTRADO"
      Y el campo "mensaje" del cuerpo es "el proyecto no existe"

    @caso_error
    Escenario: E-31 - Modificacion de un proyecto cerrado
      # RB-10 y D-07: un proyecto cerrado congela sus datos y la unica
      # operacion admitida sobre el es el cambio de estado.
      Dado que existe el proyecto "Proyecto cerrado" con "id" 5, "fecha_inicio"
        "2026-09-28", "fecha_fin" "2026-12-18" y "estado" "Cerrado"
      Y que preparo el cuerpo JSON de la modificacion:
        """
        {
          "nombre": "Intento de renombrar un proyecto cerrado",
          "fecha_inicio": "2026-09-28",
          "fecha_fin": "2026-12-18"
        }
        """
      Cuando envío "PUT" a "/api/proyectos/5" con ese cuerpo
      Entonces la API responde "409 Conflict"
      Y el campo "error" del cuerpo es "PROYECTO_CERRADO"
      Y el campo "mensaje" del cuerpo es
        "el proyecto está cerrado: solo se permite cambiar su estado o reabrirlo"
      Y el cuerpo de la respuesta no tiene el campo "campo"
      Y cuando consulto "GET" en "/api/proyectos/5"
      Entonces la API responde "200 OK" y el nombre original no cambio

    @caso_error
    Esquema del Escenario: E-32 - Orden de evaluacion de las reglas en la modificacion
      # La spec fija el orden RB-01, luego RB-02, luego RB-03 y por ultimo
      # RB-10 en el PUT: la validacion de los datos ocurre antes que la
      # verificacion del estado del proyecto.
      Dado que existe el proyecto "Proyecto E" con "id" 6 y "estado" "<estado_del_proyecto>"
      Y que preparo el cuerpo JSON de la modificacion:
        """
        {
          "nombre": "<nombre_ingresado>",
          "fecha_inicio": "<fecha_inicio_ingresada>",
          "fecha_fin": "<fecha_fin_ingresada>"
        }
        """
      Cuando envío "PUT" a "/api/proyectos/6" con ese cuerpo
      Entonces la API responde "<codigo_http_esperado>"
      Y el campo "error" del cuerpo es "<error_esperado>"
      Y el campo "mensaje" del cuerpo es "<mensaje_esperado>"

      Ejemplos:
        | nombre_ingresado | fecha_inicio_ingresada | fecha_fin_ingresada | estado_del_proyecto | codigo_http_esperado | error_esperado   | mensaje_esperado                                                     |
        | ""               | 2026-09-28            | 2026-12-18         | Activo              | 422                  | VALIDACION       | el nombre del proyecto es obligatorio                                |
        | "Proyecto"       | ""                    | 2026-12-18         | Activo              | 422                  | VALIDACION       | la fecha de inicio del proyecto es obligatoria                       |
        | "Proyecto"       | 2026-09-28            | 2026-09-27         | Activo              | 422                  | VALIDACION       | la fecha de fin no puede ser anterior a la fecha de inicio          |
        | "Proyecto"       | 2026-09-28            | 2026-12-18         | Cerrado             | 409                  | PROYECTO_CERRADO | el proyecto está cerrado: solo se permite cambiar su estado o reabrirlo |

  # ===========================================================================
  # CICLO DE VIDA DEL ESTADO: PATCH /api/proyectos/{id}/estado
  # ===========================================================================
  Regla: Cambio de estado entre "Activo" y "Cerrado"

    @caso_normal
    Escenario: E-33 - Cierre de un proyecto activo que ya tiene fecha de fin
      # CA-I9 y RB-06: cerrar exige fecha de fin informada y en este caso la hay.
      Dado que existe el proyecto "Proyecto a cerrar" con "id" 1, "fecha_inicio"
        "2026-09-28", "fecha_fin" "2026-12-18", "estado" "Activo" y 2 integrantes
      Y que preparo el cuerpo JSON del cambio de estado:
        """
        { "estado": "Cerrado" }
        """
      Cuando envío "PATCH" a "/api/proyectos/1/estado" con ese cuerpo
      Entonces la API responde "200 OK" con el proyecto completo
      Y el campo "estado" de la respuesta es "Cerrado"
      Y el campo "fecha_fin" de la respuesta sigue siendo "2026-12-18"
      Y el campo "cantidad_integrantes" de la respuesta sigue siendo 2
      Y el array "integrantes" de la respuesta tiene 2 elementos con los mismos emails
      Y cuando consulto "GET" en "/api/proyectos/1"
      Entonces la API responde "200 OK" y el proyecto figura en estado "Cerrado"

    @caso_alternativo
    Escenario: E-34 - Reapertura de un proyecto cerrado
      # Decision D-03: "Cerrado" no es un estado terminal y la reapertura es la
      # unica operacion admitida sobre un proyecto cerrado.
      Dado que existe el proyecto "Proyecto reabierto" con "id" 2, "nombre"
        "Proyecto reabierto", "fecha_inicio" "2026-09-28", "fecha_fin" "2026-12-18"
        y "estado" "Cerrado"
      Y que preparo el cuerpo JSON del cambio de estado:
        """
        { "estado": "Activo" }
        """
      Cuando envío "PATCH" a "/api/proyectos/2/estado" con ese cuerpo
      Entonces la API responde "200 OK" con el proyecto completo
      Y el campo "estado" de la respuesta es "Activo"
      Y el campo "nombre" de la respuesta sigue siendo "Proyecto reabierto"
      Y las fechas de la respuesta no fueron alteradas por la reapertura
      Y cuando envío "PUT" a "/api/proyectos/2" con el cuerpo:
        """
        {
          "nombre": "Proyecto reabierto y modificado",
          "fecha_inicio": "2026-09-28",
          "fecha_fin": "2026-12-18"
        }
        """
      Entonces la API responde "200 OK"
      Y el campo "nombre" de la respuesta es "Proyecto reabierto y modificado"
      Y el proyecto vuelve a admitir modificaciones de datos

    @caso_limite
    Escenario: E-35 - Cambio de estado al mismo estado en que ya se encuentra
      # CL-15: la transicion Activo a Activo es idempotente, responde 200 sin
      # escribir en la base y no dispara el bloqueo de RB-10.
      Dado que existe el proyecto "Proyecto ya activo" con "id" 3 y "estado" "Activo"
      Y que su campo "actualizado_en" es "2026-09-28T14:03:11Z"
      Y que preparo el cuerpo JSON del cambio de estado:
        """
        { "estado": "Activo" }
        """
      Cuando envío "PATCH" a "/api/proyectos/3/estado" con ese cuerpo
      Entonces la API responde "200 OK" con el proyecto completo
      Y el campo "estado" de la respuesta es "Activo"
      Y el campo "actualizado_en" de la respuesta sigue siendo "2026-09-28T14:03:11Z"

    @caso_limite @caso_error
    Esquema del Escenario: E-36 - Literal de estado "<estado_ingresado>"
      # RB-05 y CL-14: la comparacion del literal distingue mayusculas, por lo
      # que "activo" y "CERRADO" se rechazan con 400 ESTADO_INVALIDO. Todos los
      # ejemplos parten de un proyecto Activo con "fecha_fin" informada para que
      # RB-06 no interfiera con la verificacion de RB-05.
      Dado que existe el proyecto "Proyecto de estados" con "id" 4, "fecha_inicio"
        "2026-09-28", "fecha_fin" "2026-12-18", "estado" "Activo" y 2 integrantes
      Y que preparo el cuerpo JSON del cambio de estado con el valor "<estado_ingresado>"
      Cuando envío "PATCH" a "/api/proyectos/4/estado" con ese cuerpo
      Entonces la API responde "<codigo_http_esperado>"
      Y el cuerpo de la respuesta es <cuerpo_esperado>
      Y cuando consulto "GET" en "/api/proyectos/4"
      Entonces la API responde "200 OK" y el proyecto queda en estado "<estado_final>"

      Ejemplos:
        | estado_ingresado | codigo_http_esperado | cuerpo_esperado                                                                                                                             | estado_final |
        | Activo           | 200 OK               | el proyecto completo, con "estado" "Activo" y sus 2 integrantes intactos                                                                   | Activo       |
        | Cerrado          | 200 OK               | el proyecto completo, con "estado" "Cerrado" y "fecha_fin" "2026-12-18"                                                                      | Cerrado      |
        | activo           | 400 Bad Request      | el cuerpo de error con 'error' ESTADO_INVALIDO, 'mensaje' 'el estado del proyecto debe ser "Activo" o "Cerrado"' y 'campo' estado             | Activo       |
        | CERRADO          | 400 Bad Request      | el cuerpo de error con 'error' ESTADO_INVALIDO, 'mensaje' 'el estado del proyecto debe ser "Activo" o "Cerrado"' y 'campo' estado             | Activo       |
        | ""               | 400 Bad Request      | el cuerpo de error con 'error' ESTADO_INVALIDO, 'mensaje' 'el estado del proyecto debe ser "Activo" o "Cerrado"' y 'campo' estado             | Activo       |
        | Finalizado       | 400 Bad Request      | el cuerpo de error con 'error' ESTADO_INVALIDO, 'mensaje' 'el estado del proyecto debe ser "Activo" o "Cerrado"' y 'campo' estado             | Activo       |

    @caso_error
    Escenario: E-37 - Cierre de un proyecto sin fecha de fin informada
      # RB-06 y decision D-08: sin fecha de fin no se puede cerrar, porque un
      # proyecto cerrado debe poder ubicarse en una serie temporal. El mensaje
      # conserva el tuteo del enunciado de la spec.
      Dado que existe el proyecto "Proyecto sin fecha de fin" con "id" 5,
        "fecha_inicio" "2026-09-28", "fecha_fin" "null" y "estado" "Activo"
      Y que preparo el cuerpo JSON del cambio de estado:
        """
        { "estado": "Cerrado" }
        """
      Cuando envío "PATCH" a "/api/proyectos/5/estado" con ese cuerpo
      Entonces la API responde "422 Unprocessable Entity"
      Y el campo "error" del cuerpo es "VALIDACION"
      Y el campo "mensaje" del cuerpo es
        "no se puede cerrar un proyecto sin fecha de fin: informá la fecha de fin antes de cerrarlo"
      Y el campo "campo" del cuerpo es "fecha_fin"
      Y cuando consulto "GET" en "/api/proyectos/5"
      Entonces la API responde "200 OK" y el proyecto sigue en estado "Activo"
      Y cuando envío "PUT" a "/api/proyectos/5" con el cuerpo:
        """
        {
          "nombre": "Proyecto sin fecha de fin",
          "fecha_inicio": "2026-09-28",
          "fecha_fin": "2026-12-18"
        }
        """
      Entonces la API responde "200 OK"
      Y el proyecto ya puede cerrarse informando el estado "Cerrado"

    @caso_error
    Escenario: E-38 - Cambio de estado de un proyecto inexistente
      Dado que no existe ningun proyecto con el "id" 999
      Y que preparo el cuerpo JSON del cambio de estado:
        """
        { "estado": "Cerrado" }
        """
      Cuando envío "PATCH" a "/api/proyectos/999/estado" con ese cuerpo
      Entonces la API responde "404 Not Found"
      Y el campo "error" del cuerpo es "PROYECTO_NO_ENCONTRADO"
      Y el campo "mensaje" del cuerpo es "el proyecto no existe"

    @caso_error
    Esquema del Escenario: E-39 - Cuerpo invalido en el cambio de estado
      # El handler valida la capa sintactica del body del PATCH: un JSON
      # malformado o un "estado" de tipo incorrecto producen
      # 400 JSON_INVALIDO antes de evaluar RB-05.
      Dado que existe el proyecto "Proyecto de estados" con "id" 6 en estado "Activo"
      Y que envío un cuerpo <descripcion_del_cuerpo> en
        "PATCH /api/proyectos/6/estado" con la cabecera "Content-Type: application/json"
      Entonces la API responde "400 Bad Request"
      Y el campo "error" del cuerpo es "JSON_INVALIDO"
      Y el cuerpo de la respuesta no tiene el campo "campo"
      Y cuando consulto "GET" en "/api/proyectos/6"
      Entonces la API responde "200 OK" y el proyecto sigue en estado "Activo"

      Ejemplos:
        | descripcion_del_cuerpo                        |
        | vacio                                         |
        | con el JSON truncado, sin la llave de cierre  |
        | con "estado" como numero en lugar de texto    |
        | con un texto plano que no es JSON             |

  # ===========================================================================
  # COMPOSICION DEL EQUIPO
  #   POST   /api/proyectos/{id}/integrantes
  #   DELETE /api/proyectos/{id}/integrantes/{integranteId}
  # ===========================================================================
  Regla: Alta y baja de integrantes asociados a un proyecto

    @caso_normal
    Escenario: E-40 - Alta de un integrante en un proyecto existente
      # Criterio de aceptacion: POST /api/proyectos/{id}/integrantes agrega un
      # integrante y responde 201. El recurso creado es la asociacion entre el
      # proyecto y la persona, por eso incluye "proyecto_id".
      Dado que existe el proyecto "Proyecto del equipo" con "id" 1,
        "fecha_inicio" "2026-09-28", "fecha_fin" "2026-12-18" y 1 integrante
      Y que preparo el cuerpo JSON del integrante:
        """
        {
          "nombre": "Perez Castro Jazmín",
          "email": "jazmin.perez@utn.edu.ar"
        }
        """
      Cuando envío "POST" a "/api/proyectos/1/integrantes" con ese cuerpo
      Entonces la API responde "201 Created" con la asociacion creada
      Y el campo "id" de la respuesta es un entero positivo
      Y el campo "nombre" de la respuesta es "Perez Castro Jazmín"
      Y el campo "email" de la respuesta es "jazmin.perez@utn.edu.ar"
      Y el campo "proyecto_id" de la respuesta es 1
      Y cuando consulto "GET" en "/api/proyectos/1"
      Entonces la API responde "200 OK" y el campo "cantidad_integrantes" es 2
      Y el array "integrantes" tiene 2 elementos e incluye a "jazmin.perez@utn.edu.ar"

    @caso_alternativo @caso_limite
    Esquema del Escenario: E-41 - Alta de un integrante con email normalizado
      # RB-09 y NormalizarEmail: el email se canoniza a minusculas y sin
      # espacios antes de persistirse, y un email ya existente reutiliza el
      # registro en lugar de crear uno nuevo.
      Dado que existe el proyecto "Proyecto del equipo" con "id" 1 en estado "Activo"
      Y que preparo el cuerpo JSON del integrante con el email <email_ingresado>
      Cuando envío "POST" a "/api/proyectos/1/integrantes" con ese cuerpo
      Entonces la API responde "201 Created"
      Y el campo "email" de la respuesta es "<email_esperado>"
      Y el campo "proyecto_id" de la respuesta es 1
      Y cuando consulto "GET" en "/api/proyectos/1"
      Entonces la API responde "200 OK" y el campo "cantidad_integrantes" es 1

      Ejemplos:
        | email_ingresado                          | email_esperado                  |
        | la cadena ada.lovelace@utn.edu.ar        | ada.lovelace@utn.edu.ar        |
        | la cadena con espacios y mayusculas "  ADA.LOVELACE@UTN.edu.ar  " | ada.lovelace@utn.edu.ar |
        | la cadena con espacios y mayusculas "  ADOLOVE@ABCDEFGHIJ.ABCDEFGHIJ.ABCDEFGHIJ.ABCDEFGHIJ.ABCDEFGHIJ.ABCDEFGHIJ.ABCDEFGHIJ.ABCDEFGHIJ.ABCDEFGHIJ.ABCDEFGHIJ.ABCDEFGHIJ  " |adolove@abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij |

    @caso_error @caso_limite
    Esquema del Escenario: E-42 - Alta de un integrante con email invalido
      # RB-07 y CL-11. En este endpoint el campo reportado es "email", y no
      # "integrantes[].email" como en el alta del proyecto.
      Dado que existe el proyecto "Proyecto del equipo" con "id" 1 en estado "Activo"
      Y que preparo el cuerpo JSON del integrante con el nombre "Perez Castro Jazmín"
        y el email <email_invalido>
      Cuando envío "POST" a "/api/proyectos/1/integrantes" con ese cuerpo
      Entonces la API responde "422 Unprocessable Entity"
      Y el campo "error" del cuerpo es "VALIDACION"
      Y el campo "mensaje" del cuerpo es
        "el email del integrante es obligatorio y debe tener un formato válido"
      Y el campo "campo" del cuerpo es "email"
      Y cuando consulto "GET" en "/api/proyectos/1"
      Entonces la API responde "200 OK" y el campo "cantidad_integrantes" es 0

      Ejemplos:
        | email_invalido                                                                        |
        | la cadena jazmin.perez.utn.edu.ar                                                     |
        | la cadena "jazmin perez@utn.edu.ar" con un espacio embebido                            |
        | la cadena "Perez Jazmin <jazmin.perez@utn.edu.ar>"                                     |
        | la cadena vacia                                                                        |
        | adalovel@abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij.abcdefghij |

    @caso_error
    Esquema del Escenario: E-43 - Alta de un integrante con nombre invalido
      # RB-07: el nombre del integrante es obligatorio y se valida sobre el
      # resultado de TrimSpace. En este endpoint el campo es "nombre".
      Dado que existe el proyecto "Proyecto del equipo" con "id" 1 en estado "Activo"
      Y que preparo el cuerpo JSON del integrante con <nombre_invalido> y el email
        "jazmin.perez@utn.edu.ar"
      Cuando envío "POST" a "/api/proyectos/1/integrantes" con ese cuerpo
      Entonces la API responde "422 Unprocessable Entity"
      Y el campo "error" del cuerpo es "VALIDACION"
      Y el campo "mensaje" del cuerpo es "el nombre del integrante es obligatorio"
      Y el campo "campo" del cuerpo es "nombre"
      Y cuando consulto "GET" en "/api/proyectos/1"
      Entonces la API responde "200 OK" y el campo "cantidad_integrantes" es 0

      Ejemplos:
        | nombre_invalido                        |
        | la persona con la cadena de nombre ""  |
        | la persona con la cadena de nombre "  "|

    @caso_error @caso_limite
    Escenario: E-44 - Alta de un integrante que ya forma parte del proyecto
      # CL-09 y RB-08: el email ya asociado se detecta por su forma
      # normalizada, por eso la capitalizacion distinta no evita el rechazo.
      Dado que existe el proyecto "Proyecto del equipo" con "id" 1 en estado "Activo"
      Y que el email "ada@utn.edu.ar" ya esta asociado a ese proyecto
      Y que preparo el cuerpo JSON del integrante:
        """
        {
          "nombre": "Aguilera Sebastian",
          "email": "ADA@utn.edu.ar"
        }
        """
      Cuando envío "POST" a "/api/proyectos/1/integrantes" con ese cuerpo
      Entonces la API responde "409 Conflict"
      Y el campo "error" del cuerpo es "INTEGRANTE_DUPLICADO"
      Y el campo "mensaje" del cuerpo es "el integrante ya forma parte del proyecto"
      Y el campo "campo" del cuerpo es "email"
      Y cuando consulto "GET" en "/api/proyectos/1"
      Entonces la API responde "200 OK" y el campo "cantidad_integrantes" es 1

    @caso_limite
    Escenario: E-45 - Dos peticiones concurrentes con el mismo email al mismo proyecto
      # CL-17: la clave primaria compuesta de la tabla puente
      # "proyecto_integrantes" actua como garantia final frente a la condicion
      # de carrera. El escenario se ejecuta lanzando las dos peticiones en
      # paralelo, por ejemplo con dos comandos curl en background o con dos
      # goroutines en httptest.
      Dado que existe el proyecto "Proyecto del equipo" con "id" 1 en estado "Activo"
        y sin integrantes
      Y que el cuerpo de las dos peticiones es:
        """
        {
          "nombre": "Aguilera Sebastián",
          "email": "sebas.aguilera@utn.edu.ar"
        }
        """
      Y que envío al mismo tiempo dos peticiones "POST" a "/api/proyectos/1/integrantes"
        con ese cuerpo
      Entonces una de las dos respuestas es "201 Created" y la otra es "409 Conflict"
      Y el cuerpo de la respuesta "409" tiene "error" "INTEGRANTE_DUPLICADO"
      Y el campo "mensaje" del cuerpo de error es "el integrante ya forma parte del proyecto"
      Y cuando consulto "GET" en "/api/proyectos/1"
      Entonces la API responde "200 OK" y el campo "cantidad_integrantes" es 1
      Y el email "sebas.aguilera@utn.edu.ar" figura una sola vez en el array "integrantes"

    @caso_normal
    Escenario: E-46 - Baja de un integrante del proyecto
      # Criterio de aceptacion: DELETE /api/proyectos/{id}/integrantes/
      # {integranteId} quita un integrante del proyecto y responde 204.
      Dado que existe el proyecto "Proyecto del equipo" con "id" 1 en estado "Activo"
        y 2 integrantes
      Y que el email "jazmin.perez@utn.edu.ar" figura con el "id" 2 en el detalle
      Cuando envío "DELETE" a "/api/proyectos/1/integrantes/2" sin cuerpo
      Entonces la API responde "204 No Content" con el cuerpo de respuesta vacio
      Y cuando consulto "GET" en "/api/proyectos/1"
      Entonces la API responde "200 OK" y el campo "cantidad_integrantes" es 1
      Y el array "integrantes" ya no incluye a "jazmin.perez@utn.edu.ar"
      Y el proyecto sigue existiendo y en estado "Activo"

    @caso_error
    Esquema del Escenario: E-47 - Baja de un integrante que no forma parte del proyecto
      # RB-12: EstaAsociado debe devolver true; el repository filtra por
      # borrado_en IS NULL, por lo que un integrante de otro proyecto y un id
      # inexistente se comportan igual.
      Dado que existen el proyecto con "id" 1 y el proyecto con "id" 2, ambos
        en estado "Activo"
      Y que el email "ada@utn.edu.ar" esta asociado al proyecto con "id" 2
        y tiene el "id" 30
      Y que el proyecto con "id" 1 no tiene integrantes
      Y que uso el identificador <integrante_id> porque es
        <detalle_del_integrante_id>
      Cuando envío "DELETE" a "/api/proyectos/1/integrantes/<integrante_id>" sin cuerpo
      Entonces la API responde "404 Not Found"
      Y el campo "error" del cuerpo es "INTEGRANTE_NO_ENCONTRADO"
      Y el campo "mensaje" del cuerpo es "el integrante no forma parte del proyecto"
      Y el cuerpo de la respuesta no tiene el campo "campo"
      Y cuando consulto "GET" en "/api/proyectos/2"
      Entonces la API responde "200 OK" y el proyecto conserva a "ada@utn.edu.ar"

      Ejemplos:
        | detalle_del_integrante_id | integrante_id |
        | el integrante del otro proyecto | 30          |
        | un identificador inexistente     | 999999      |

    @caso_error
    Esquema del Escenario: E-48 - Operaciones bloqueadas sobre un proyecto cerrado
      # RB-10 y D-07: un proyecto cerrado bloquea el PUT y las dos operaciones
      # de composicion del equipo, pero no el cambio de estado, que es la
      # excepcion de la decision D-03.
      Dado que existe el proyecto "Proyecto congelado" con "id" 1, "fecha_inicio"
        "2026-09-28", "fecha_fin" "2026-12-18", "estado" "Cerrado" y 1 integrante
      Y que el email "ada@utn.edu.ar" figura con el "id" 10 en el detalle
      Y que envío la peticion <descripcion_de_la_operacion> sin cuerpo
      Entonces la API responde "409 Conflict"
      Y el campo "error" del cuerpo es "PROYECTO_CERRADO"
      Y el campo "mensaje" del cuerpo es
        "el proyecto está cerrado: solo se permite cambiar su estado o reabrirlo"
      Y el cuerpo de la respuesta no tiene el campo "campo"
      Y cuando consulto "GET" en "/api/proyectos/1"
      Entonces la API responde "200 OK" y el proyecto sigue en estado "Cerrado"
      Y el campo "cantidad_integrantes" sigue siendo 1
      Y cuando envío "PATCH" a "/api/proyectos/1/estado" con el cuerpo:
        """
        { "estado": "Activo" }
        """
      Entonces la API responde "200 OK" y el proyecto vuelve a estado "Activo"

      Ejemplos:
        | descripcion_de_la_operacion                                                                                     |
        | "PUT" a "/api/proyectos/1" con el cuerpo "nombre": "Intento", "fecha_inicio": "2026-09-28", "fecha_fin": "2026-12-18" |
        | "POST" a "/api/proyectos/1/integrantes" con el cuerpo "nombre": "Intruso", "email": "intruso@utn.edu.ar"          |
        | "DELETE" a "/api/proyectos/1/integrantes/10"                                                                     |

  # ===========================================================================
  # BAJA LOGICA: DELETE /api/proyectos/{id}
  # ===========================================================================
  Regla: Baja logica de un proyecto

    @caso_normal
    Escenario: E-49 - Baja logica de un proyecto
      # Criterio de aceptacion: DELETE realiza la baja logica, el proyecto deja
      # de aparecer en el listado y en el detalle, y responde 204. Decision
      # D-04: la fila permanece con "borrado_en" informado para no destruir el
      # historial que usan las metricas.
      Dado que existen dos proyectos: el de "id" 1 y el de "id" 2
      Y que el proyecto de "id" 1 tiene 2 integrantes
      Y que ninguno fue dado de baja logicamente
      Cuando envío "DELETE" a "/api/proyectos/1" sin cuerpo
      Entonces la API responde "204 No Content" con el cuerpo de respuesta vacio
      Y cuando consulto "GET" en "/api/proyectos/1"
      Entonces la API responde "404 Not Found"
      Y el campo "error" del cuerpo es "PROYECTO_NO_ENCONTRADO"
      Y el campo "mensaje" del cuerpo es "el proyecto no existe"
      Y cuando consulto "GET" en "/api/proyectos"
      Entonces la API responde "200 OK" con el campo "total" igual a 1
      Y el array "proyectos" ya no incluye el proyecto de "id" 1
      Y el proyecto de "id" 2 sigue visible en el listado
      Y cuando consulto la tabla "proyectos" con la sentencia
        "SELECT id, borrado_en FROM proyectos WHERE id = 1;"
      Entonces la fila del proyecto de "id" 1 sigue existiendo
      Y la columna "borrado_en" de esa fila tiene un valor no nulo

    @caso_error
    Escenario: E-50 - Baja de un proyecto que ya fue dado de baja
      # CL-13: la baja logica es invisible para el cliente, igual que un
      # identificador inexistente.
      Dado que existe el proyecto con "id" 1 y que fue dado de baja logicamente
      Y que su detalle responde "404 Not Found"
      Cuando envío "DELETE" a "/api/proyectos/1" sin cuerpo
      Entonces la API responde "404 Not Found"
      Y el campo "error" del cuerpo es "PROYECTO_NO_ENCONTRADO"
      Y el campo "mensaje" del cuerpo es "el proyecto no existe"

    @caso_error
    Escenario: E-51 - Baja de un proyecto inexistente
      Dado que no existe ningun proyecto con el "id" 999
      Cuando envío "DELETE" a "/api/proyectos/999" sin cuerpo
      Entonces la API responde "404 Not Found"
      Y el campo "error" del cuerpo es "PROYECTO_NO_ENCONTRADO"
      Y el campo "mensaje" del cuerpo es "el proyecto no existe"

    @caso_error @escenario_diferido
    Escenario: E-52 - Baja de un proyecto con historial asociado
      # RB-11: ProyectoRepository.TieneHistorial se consulta antes de la baja.
      # Este escenario queda diferido porque en US-01 todavia no existe ninguna
      # US que registre Sprints, historias ni worklogs, de modo que
      # TieneHistorial devuelve false y la baja responde 204 (E-49). Se
      # ejecuta cuando exista la primera US que cree esas entidades; hasta
      # entonces la precondicion se prepara insertando la fila de dependencia
      # por SQL.
      Dado que existe el proyecto con "id" 1 en estado "Activo"
        y con "fecha_fin" "2026-12-18"
      Y que ese proyecto ya tiene historial asociado en las tablas de Sprints,
        historias o worklogs de las historias de usuario posteriores
      Cuando envío "DELETE" a "/api/proyectos/1" sin cuerpo
      Entonces la API responde "409 Conflict"
      Y el campo "error" del cuerpo es "PROYECTO_CON_HISTORIAL"
      Y el campo "mensaje" del cuerpo es
        "el proyecto tiene historial asociado y no puede eliminarse"
      Y el cuerpo de la respuesta no tiene el campo "campo"
      Y cuando consulto "GET" en "/api/proyectos/1"
      Entonces la API responde "200 OK" y el proyecto sigue existiendo

    @caso_error
    Esquema del Escenario: E-53 - Operaciones de composicion sobre un proyecto dado de baja
      # CL-16: el filtrado por borrado_en IS NULL se aplica tambien a las
      # operaciones de composicion del equipo, por lo que respondem
      # PROYECTO_NO_ENCONTRADO y no INTEGRANTE_NO_ENCONTRADO.
      Dado que existe el proyecto con "id" 1 y que fue dado de baja logicamente
      Y que el email "ada@utn.edu.ar" estaba asociado a ese proyecto con el "id" 10
      Y que envío <descripcion_de_la_operacion> sin cuerpo
      Entonces la API responde "404 Not Found"
      Y el campo "error" del cuerpo es "PROYECTO_NO_ENCONTRADO"
      Y el campo "mensaje" del cuerpo es "el proyecto no existe"
      Y el cuerpo de la respuesta no tiene el campo "campo"

      Ejemplos:
        | descripcion_de_la_operacion                                                                              |
        | "POST" a "/api/proyectos/1/integrantes" con el cuerpo "nombre": "Intruso", "email": "intruso@utn.edu.ar"   |
        | "DELETE" a "/api/proyectos/1/integrantes/10"                                                             |

  # ===========================================================================
  # ROBUSTEZ, FORMATO Y FALLAS DE INFRAESTRUCTURA
  # ===========================================================================
  Regla: Comportamiento ante limites inferiores, formato de fechas e infraestructura

    @caso_error
    Escenario: E-54 - Falla la persistencia de un integrante durante el alta
      # CL-18 y atomicidad: CrearConIntegrantes inserta el proyecto y todos sus
      # vinculos en una sola transaccion, de modo que ante un fallo no queda
      # persistencia parcial. Para provocar el fallo se revoca el permiso de
      # INSERT sobre la tabla "integrantes" para el usuario de la aplicacion y
      # se restaura al final del escenario.
      Dado que existe el proyecto con "id" 1 en estado "Activo"
        como victima de la violacion del indice unico "idx_integrantes_email"
        para forzar un fallo de persistencia en un integrante del alta
      Y que preparo el cuerpo JSON del alta:
        """
        {
          "nombre": "Proyecto con fallo de persistencia",
          "fecha_inicio": "2026-09-28",
          "fecha_fin": "2026-12-18",
          "integrantes": [
            { "nombre": "Aguilera Sebastián", "email": "sebas.aguilera@utn.edu.ar" },
            { "nombre": "Choquevillca Celeste", "email": "celeste.choque@utn.edu.ar" }
          ]
        }
        """
      Y que la sentencia "REVOKE INSERT ON integrantes FROM <usuario_app>;"
        dejo a la aplicacion sin permiso de insercion
      Cuando envío "POST" a "/api/proyectos" con ese cuerpo
      Entonces la API responde "500 Internal Server Error"
      Y el campo "error" del cuerpo es "ERROR_INTERNO"
      Y el campo "mensaje" del cuerpo es
        "ocurrió un error inesperado al procesar la solicitud"
      Y el cuerpo de la respuesta no filtra detalles de infraestructura
      Y cuando consulto "GET" en "/api/proyectos"
      Entonces la API responde "200 OK" y el proyecto del body no fue persistido
      Y cuando consulto la tabla "integrantes" con la sentencia
        "SELECT email FROM integrantes WHERE email = 'sebas.aguilera@utn.edu.ar';"
      Entonces la sentencia no devuelve filas
      Y el permiso se restaura con "GRANT INSERT ON integrantes TO <usuario_app>;"

    @caso_error
    Escenario: E-55 - Base de datos inaccesible durante una request
      # Fallas de infraestructura: un error del driver no reconocido por el
      # switch del handler produce 500 ERROR_INTERNO con el mensaje generico,
      # y el error real queda en el log del servidor.
      Dado que la API esta levantada y la base de datos fue detenida, por
        ejemplo con "docker compose stop postgres"
      Cuando consulto "GET" en "/api/proyectos"
      Entonces la API responde "500 Internal Server Error"
      Y el campo "error" del cuerpo es "ERROR_INTERNO"
      Y el campo "mensaje" del cuerpo es
        "ocurrió un error inesperado al procesar la solicitud"
      Y el cuerpo de la respuesta no tiene el campo "campo"
      Y la API sigue respondiendo "200 OK" en "GET /health"
        porque gin.Recovery no cierra el proceso
      Y la base de datos vuelve a estar disponible con "docker compose start postgres"

    @caso_limite
    Escenario: E-56 - Nombre del proyecto con un solo caracter
      # Limite inferior de RB-01: 1 caracter es el minimo valido.
      Dado que preparo el cuerpo JSON del alta:
        """
        {
          "nombre": "P",
          "fecha_inicio": "2026-09-28",
          "fecha_fin": "2026-12-18"
        }
        """
      Cuando envío "POST" a "/api/proyectos" con ese cuerpo
      Entonces la API responde "201 Created"
      Y el campo "nombre" de la respuesta es "P"
      Y el campo "nombre" de la respuesta tiene 1 caracter

    @caso_limite
    Escenario: E-57 - Nombre de integrante con exactamente 100 caracteres
      # Limite superior de la longitud del nombre de un integrante, segun el
      # criterio de RB-07 y la columna varchar(100). El nombre enviado es
      # "ABCDE" repetido 20 veces, es decir 100 caracteres exactos.
      Dado que existe el proyecto con "id" 1 en estado "Activo"
      Y que preparo el cuerpo JSON del integrante con el nombre formado por
        "ABCDE" repetido 20 veces y el email "ada@utn.edu.ar"
      Cuando envío "POST" a "/api/proyectos/1/integrantes" con ese cuerpo
      Entonces la API responde "201 Created"
      Y el campo "nombre" de la respuesta tiene 100 caracteres
      Y el campo "proyecto_id" de la respuesta es 1

    @caso_limite @rfc
    Escenario: E-58 - Serializacion de fechas en el alta y en la modificacion
      # CA-I8 y @rfc: las fechas de calendario se devuelven siempre en ISO 8601
      # "YYYY-MM-DD" sin hora ni zona, la fecha de fin ausente se serializa como
      # "null" y las fechas de auditoria en RFC 3339 UTC.
      Dado que preparo el cuerpo JSON del alta:
        """
        {
          "nombre": "Proyecto de formatos",
          "fecha_inicio": "2026-09-28",
          "fecha_fin": "2026-12-18"
        }
        """
      Cuando envío "POST" a "/api/proyectos" con ese cuerpo
      Entonces la API responde "201 Created"
      Y la respuesta serializada contiene "fecha_inicio":"2026-09-28"
        y "fecha_fin":"2026-12-18"
      Y la respuesta serializada contiene "creado_en" y "actualizado_en"
        con valores en formato RFC 3339 UTC
      Y cuando envío "PUT" al proyecto con el cuerpo:
        """
        {
          "nombre": "Proyecto de formatos",
          "fecha_inicio": "2026-09-28",
          "fecha_fin": null
        }
        """
      Entonces la API responde "200 OK"
      Y la respuesta serializada contiene "fecha_fin":null
      Y la respuesta serializada no contiene "fecha_fin":""
      Y la respuesta serializada no contiene "2026-09-28T00:00:00"
      Y el campo "fecha_inicio" de la respuesta sigue siendo "2026-09-28"
        y no fue alterado por el paso a null de la fecha de fin

# =============================================================================
# MATRIZ DE TRAZABILIDAD
# Regla de negocio -> escenario que la verifica
#   RB-01 nombre obligatorio de 1 a 150 caracteres      E-06 E-07 E-08 E-09
#                                                      E-28 E-32 E-56
#   RB-02 fecha de inicio obligatoria                   E-10 E-19 E-32
#   RB-03 fecha de fin no anterior a la de inicio       E-04 E-11 E-19 E-27
#                                                      E-29 E-32
#   RB-04 todo proyecto nace Activo                     E-01 E-05
#   RB-05 estado solo Activo o Cerrado                  E-36
#   RB-06 cerrar exige fecha de fin informada           E-33 E-37
#   RB-07 nombre y email validos por integrante         E-14 E-15 E-16 E-42
#                                                      E-43 E-57
#   RB-08 el mismo email no se repite en un proyecto    E-17 E-44 E-45
#   RB-09 un email identifica a una sola persona        E-18 E-41
#   RB-10 proyecto cerrado: solo cambia su estado       E-31 E-32 E-48
#   RB-11 no se da de baja con historial asociado       E-52
#   RB-12 no se quita un integrante no asociado          E-47
#
# Caso limite -> escenario
#   CL-01 E-09   CL-02 E-06   CL-03 E-07   CL-04 E-04, E-27   CL-05 E-02
#   CL-06 E-03   CL-07 E-03   CL-08 E-17   CL-09 E-44         CL-10 E-18
#   CL-11 E-14, E-42   CL-12 E-21   CL-13 E-50   CL-14 E-36
#   CL-15 E-35   CL-16 E-53   CL-17 E-45   CL-18 E-54
#
# Criterio de aceptacion de la issue #28 -> escenario
#   CA-I1 E-01 E-04    CA-I2 E-01 E-17 E-18   CA-I3 E-04 E-11
#   CA-I4 E-08 E-09    CA-I5 E-01 E-05        CA-I6 E-02 E-10 E-32
#   CA-I7 E-03 E-40 E-46   CA-I8 E-12 E-22 E-58   CA-I9 E-33 E-34 E-36
#
# Escenario de la issue -> identificador de escenario en este archivo
#   Caso normal de la issue            E-01
#   Caso limite de fechas (mismo dia) E-04
#   Error fecha de fin anterior       E-11
#   Error nombre vacio                E-08
#   Error sin fecha de inicio         E-10
#
# -----------------------------------------------------------------------------
# SUGERENCIA DE AUTOMATIZACION (TDD)
# -----------------------------------------------------------------------------
# Los pasos de este archivo se replican como subtests t.Run() con el
# identificador E-nn como nombre, para que el reporte de go test y el reporte
# de BDD sean comparables:
#   internal/service/proyecto_service_test.go   reglas RB-01 a RB-12 con mocks
#   internal/handler/proyecto_handler_test.go  codigos HTTP, cuerpos de error y
#                                             parseo de path params con httptest
#   internal/domain/proyecto_test.go           Valido, NormalizarEmail, EmailValido
#   internal/repository/migraciones_test.go    AutoMigrate idempotente e indices
#   cmd/api/router_test.go                     las 8 rutas de /api y /health
# Para la suite end to end con godog, cada Regla de este archivo se asocia al
# servicio HTTP de prueba correspondiente: "ProyectoService" con el router real
# y una base de datos de prueba.
# =============================================================================
