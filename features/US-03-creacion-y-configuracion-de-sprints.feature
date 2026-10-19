# language: es
# =============================================================================
# US-03 - Creacion y configuracion de Sprints
# Escenarios BDD (Gherkin) del modulo de Sprints del
# Software Metrics & Estimation Engine.
#
# Issue de origen     : SebastianAguilera/ingenieria-y-calidad#30
# Especificacion SDD  : docs/specs/US-03-creacion-y-configuracion-de-sprints-spec.md
# Metodologia         : docs/methodology.md seccion 4 (BDD con archivos .feature)
# Requerimiento       : docs/proyecto.md, requerimiento 3 (Sprints)
# Rama de trabajo     : feature/US-03-creacion-y-configuracion-de-sprints
# Trazabilidad        : Issue #30 -> SDD -> Criterios de aceptacion -> Escenarios
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
#   {id}, {sprintId}  = path param segun la notacion de la spec SDD.
#   <nombre>          = placeholder de Gherkin (Esquema del Escenario).
#   "campo"           = nombre de campo del JSON, entrecomillado para que el
#                       paso sea inequivoco al traducirlo a un test Go.
#   E-nn              = identificador estable del escenario; se reutiliza
#                       como nombre del subtest t.Run() en los *_test.go.
#
# Etiquetas:
#   @caso_normal        comportamiento principal o recorrido de extremo a
#                       extremo.
#   @caso_alternativo   camino valido distinto del principal.
#   @caso_limite        borde de los valores permitidos: longitudes 50/300,
#                       fechas contiguas, un solo dia, etc.
#   @caso_error         rechazo esperado: 400 / 404 / 409 / 422 / 500.
#   @rfc                verificacion del estandar RFC 3339 UTC en creado_en y
#                       actualizado_en.
#
# Notas de la spec (fijadas por la SDD):
#   NOTA-1  El estado no se acepta del cliente en el alta: todo Sprint nace en
#           "Planificado" (D-03). Los campos "id", "estado", "creado_en" y
#           "actualizado_en" que lleguen en el body del alta se ignoran.
#   NOTA-2  Las fechas de Sprint son fechas de calendario "YYYY-MM-DD" (mismo
#           formato que US-01). Un Sprint de un solo dia tiene
#           "fecha_fin" == "fecha_inicio" y es valido.
#   NOTA-3  El solapamiento se define con desigualdades estrictas (D-07): dos
#           Sprints contiguos, donde la "fecha_fin" de uno es igual a la
#           "fecha_inicio" del siguiente, NO se solapan.
#   NOTA-4  El literal de estado fuera del conjunto cerrado responde 400
#           ESTADO_INVALIDO (D-06), mientras que los defectos de negocio de un
#           campo valido (vacio, largo, orden de fechas) responden 422
#           VALIDACION.
#   NOTA-5  La unicidad del identificador y el rechazo por solapamiento se
#           evaluan contra los Sprints existentes del MISMO proyecto; en otro
#           proyecto el identificador repetido esta permitido (RB-04).
#
# Entorno de ejecucion (variables resueltas antes de correr los escenarios):
#   BASE_URL = http://localhost:$PORT   (PORT segun .env; 8080 en .env.example
#                                       y 8081 si la variable no esta definida)
#   DB       = PostgreSQL 16 de docker-compose, publicado en el host en 5433
#               (DB_PORT=5433 para la API en el host y 5432 dentro de la red de
#               contenedores, ver US-01 Restriccion 13).
#   Ejemplo curl:
#     curl -i -X POST "$BASE_URL/api/proyectos/1/sprints" \
#       -H "Content-Type: application/json" -d @sprint.json
#   Ejemplo httptest: los mismos pasos se ejecutan contra el *gin.Engine real en
#     internal/handler/sprint_handler_test.go y cmd/api/router_test.go.
#
# -----------------------------------------------------------------------------
# NARRATIVA (US-03, issue #30)
# -----------------------------------------------------------------------------
# Como equipo de desarrollo, quiero crear un Sprint especificando un objetivo
# (Sprint Goal) y sus fechas, para organizar el ciclo de trabajo.
#
# -----------------------------------------------------------------------------
# CRITERIOS DE ACEPTACION DE LA ISSUE #30
# -----------------------------------------------------------------------------
# [CA-I1] Cada Sprint contiene ID, Identificador, Sprint Goal, Estado y fechas.
#                                       -> E-01, E-02, E-12, E-24, E-25, E-29,
#                                          E-30
# [CA-I2] Se puede crear un Sprint con su Sprint Goal y fechas validas, y
#         queda registrado con estado "Planificado" (estado inicial por
#         defecto).                    -> E-01, E-02, E-07, E-09, E-12, E-14,
#                                        E-17, E-24
# [CA-I3] Se puede listar el backlog de Sprints del proyecto.
#                                       -> E-25, E-26
# [CA-I4] Se puede cambiar el estado del Sprint (Planificado, Activo,
#         Cerrado) y al activarlo pasa a ser el Sprint activo del proyecto.
#                                       -> E-29, E-30
#
# Criterios adicionales sugeridos por el equipo (issue #30):
# [CA-I5] El identificador de Sprint es unico dentro del proyecto.
#                                       -> E-13, E-14
# [CA-I6] El Sprint Goal es obligatorio y no puede estar vacio.
#                                       -> E-05, E-06, E-10
# [CA-I7] Los estados forman una maquina de estados con transiciones
#         permitidas; "Cerrado" es terminal.  -> E-29, E-30, E-31, E-32, E-33,
#                                        E-34, E-35
# [CA-I8] Solo un Sprint puede estar "Activo" a la vez por proyecto.
#                                       -> E-41
# [CA-I9] Las fechas del Sprint no deben solaparse con las de otros Sprints
#         del mismo proyecto.          -> E-15, E-16, E-17
#
# -----------------------------------------------------------------------------
# REGLAS DE NEGOCIO CUBIERTAS
# -----------------------------------------------------------------------------
# RB-01 identificador obligatorio de 1 a 50 caracteres tras TrimSpace
#                                  -> E-03, E-04, E-07, E-08
# RB-02 Sprint Goal obligatorio de 1 a 300 caracteres tras TrimSpace
#                                  -> E-05, E-06, E-09, E-10
# RB-03 ambas fechas obligatorias y fecha_fin >= fecha_inicio
#                                  -> E-11, E-12, E-19, E-20, E-21
# RB-04 identificador unico dentro del proyecto
#                                  -> E-13, E-14
# RB-05 maquina de estados Planificado -> Activo -> Cerrado; Cerrado es
#       terminal y no se admite la transicion al mismo estado
#                                  -> E-29, E-30, E-31, E-32, E-33, E-34, E-35
# RB-06 solo un Sprint activo por proyecto
#                                  -> E-41
# RB-07 fechas sin solapamiento; los Sprints contiguos estan permitidos
#                                  -> E-15, E-16, E-17
# RB-08 el estado inicial por defecto es "Planificado" y no es elegible
#       por el cliente             -> E-01, E-24
# RB-09 el listado es por proyecto, ordenado por fecha_inicio ASC
#                                  -> E-25, E-26
# RB-10 el proyecto debe existir    -> E-18, E-27, E-40
# RB-11 el Sprint debe existir y pertenecer al proyecto de la ruta
#                                  -> E-38, E-39
#
# -----------------------------------------------------------------------------
# CASOS LIMITE CUBIERTOS
# -----------------------------------------------------------------------------
# CL-01 identificador de exactamente 50 caracteres          -> E-07
# CL-02 identificador de 51 caracteres                      -> E-08
# CL-03 identificador vacio o solo espacios                 -> E-03, E-04
# CL-04 Sprint Goal de exactamente 300 caracteres           -> E-09
# CL-05 Sprint Goal de 301 caracteres                       -> E-10
# CL-06 Sprint Goal vacio o solo espacios                   -> E-05, E-06
# CL-07 fecha_fin == fecha_inicio (Sprint de un solo dia)   -> E-12
# CL-08 fecha_fin < fecha_inicio                            -> E-11
# CL-09 identificador ya tomado en el mismo proyecto        -> E-13
# CL-10 mismo identificador en otro proyecto                -> E-14
# CL-11 solapamiento parcial de rangos                      -> E-15
# CL-12 un Sprint contenido completamente en el de otro     -> E-16
# CL-13 Sprints contiguos (fin de uno == inicio del otro)   -> E-17
# CL-14 Planificado -> Activo sin ningun Sprint activo      -> E-29
# CL-15 Activo -> Activo (misma estado)                     -> E-35
# CL-16 Planificado -> Cerrado (saltear Activo)             -> E-32
# CL-17 activar con otro Sprint ya activo                   -> E-41
# CL-18 listado de un proyecto sin Sprints                  -> E-26
# CL-19 Sprint de otro proyecto: invisible para la ruta     -> E-39
# CL-20 id o sprintId no numerico o 0                       -> E-22, E-28, E-37
# CL-21 fecha ausente o no parseable                        -> E-19, E-20, E-21
#
# -----------------------------------------------------------------------------
# ENDPOINTS CUBIERTOS (los 3 de la spec SDD, ninguno mas)
# -----------------------------------------------------------------------------
#   1. POST   /api/proyectos/{id}/sprints                    -> E-01 a E-24
#   2. GET    /api/proyectos/{id}/sprints                    -> E-25 a E-28
#   3. PATCH  /api/proyectos/{id}/sprints/{sprintId}/estado  -> E-29 a E-41
# =============================================================================

Funcionalidad: US-03 - Creacion y configuracion de Sprints
  Como equipo de desarrollo
  Quiero crear un Sprint especificando un objetivo (Sprint Goal) y sus fechas
  Para organizar el ciclo de trabajo.

  Todo Sprint pertenece a un proyecto y se identifica con un identificador
  visible unico dentro del proyecto ("Sprint 1", "Sprint 2", ...). Cada Sprint
  nace en estado "Planificado" y solo avanza por la maquina de estados
  Planificado -> Activo -> Cerrado, con un unico Sprint activo por proyecto y
  sin solapamiento de fechas entre Sprints del mismo proyecto. No existe
  autenticacion en US-03: el actor es implicitamente un miembro del equipo que
  planifica el proyecto. La asignacion de historias al Sprint y el cierre con
  metricas no forman parte de US-03: se cubren en US-04 y US-07.

  Contexto:
    Dado que la API esta levantada y responde "200 OK" en "GET /health"
    Y que la base de datos "metrics_db" fue migrada en el arranque por
      "AutoMigrate" y existen las tablas "proyectos", "historias", "usuarios",
      "proyecto_usuarios" y "sprints"
    Y que la suite arranca de una base limpia, con estas sentencias ejecutadas
      por el runner antes de cada escenario:
      """
      DELETE FROM sprints;
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
  # ALTA DE SPRINTS: POST /api/proyectos/{id}/sprints
  # ===========================================================================
  Regla: Alta de un Sprint dentro de un proyecto

    @caso_normal
    Escenario: E-01 - Alta de un Sprint con todos sus campos
      # Caso normal de la issue #30 (CA-I1, CA-I2, CA-I8) y reglas RB-01,
      # RB-02, RB-03 y RB-08. El Sprint se persiste con su id y su estado
      # inicial "Planificado" y aparece en el listado del proyecto.
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y sin sprints
      Y que preparo el cuerpo JSON del alta del sprint:
        """
        {
          "identificador": "Sprint 1",
          "sprint_goal": "Entregar el MVP del motor de metricas",
          "fecha_inicio": "2026-10-06",
          "fecha_fin": "2026-10-20"
        }
        """
      Cuando envio "POST" a "/api/proyectos/1/sprints" con ese cuerpo
      Entonces la respuesta tiene el estado HTTP "201"
      Y el cuerpo de la respuesta contiene:
        """
        {
          "identificador": "Sprint 1",
          "sprint_goal": "Entregar el MVP del motor de metricas",
          "estado": "Planificado",
          "fecha_inicio": "2026-10-06",
          "fecha_fin": "2026-10-20"
        }
        """
      Y la respuesta contiene un "id" numerico distinto de cero
      Y el Sprint aparece en "GET /api/proyectos/1/sprints" con "total" 1

    @caso_alternativo
    Escenario: E-02 - Alta con identificador y goal rodeados de espacios
      # Regla RB-01 y RB-02: los textos se normalizan con TrimSpace antes de
      # persistir (caja E-02, equivalente a E-07 de US-02).
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y sin sprints
      Y que preparo el cuerpo JSON del alta del sprint:
        """
        {
          "identificador": "   Sprint 1   ",
          "sprint_goal": "  Entregar el MVP  ",
          "fecha_inicio": "2026-10-06",
          "fecha_fin": "2026-10-20"
        }
        """
      Cuando envio "POST" a "/api/proyectos/1/sprints" con ese cuerpo
      Entonces la respuesta tiene el estado HTTP "201"
      Y el cuerpo de la respuesta contiene:
        """
        {
          "identificador": "Sprint 1",
          "sprint_goal": "Entregar el MVP"
        }
        """

    @caso_error
    Esquema del Escenario: E-03 - Alta con el "identificador" en la variante <variante>
      # Regla RB-01 / CL-03: el identificador es obligatorio.
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y sin sprints
      Y que preparo el cuerpo JSON del alta del sprint:
        """
        {
          "identificador": <identificador>,
          "sprint_goal": "Entregar el MVP",
          "fecha_inicio": "2026-10-06",
          "fecha_fin": "2026-10-20"
        }
        """
      Cuando envio "POST" a "/api/proyectos/1/sprints" con ese cuerpo
      Entonces la respuesta tiene el estado HTTP "422"
      Y el cuerpo de la respuesta contiene el error "VALIDACION" con campo
        "identificador" y el mensaje:
        """
        el identificador del sprint es obligatorio
        """

      Ejemplos:
        | variante        | identificador |
        | vacio           | ""            |
        | solo espacios   | "    "        |

    @caso_limite
    Escenario: E-04 - Identificador compuesto solo por tabuladores
      # Regla RB-01 / CL-03: TrimSpace tambien elimina tabuladores.
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y sin sprints
      Y que preparo el cuerpo JSON del alta del sprint:
        """
        {
          "identificador": "\t\t",
          "sprint_goal": "Entregar el MVP",
          "fecha_inicio": "2026-10-06",
          "fecha_fin": "2026-10-20"
        }
        """
      Cuando envio "POST" a "/api/proyectos/1/sprints" con ese cuerpo
      Entonces la respuesta tiene el estado HTTP "422"
      Y el cuerpo de la respuesta contiene el error "VALIDACION" con campo
        "identificador"

    @caso_error
    Esquema del Escenario: E-05 - Alta con el "sprint_goal" en la variante <variante>
      # Regla RB-02 / CL-06: el Sprint Goal es obligatorio (issue #30,
      # criterio adicional).
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y sin sprints
      Y que preparo el cuerpo JSON del alta del sprint:
        """
        {
          "identificador": "Sprint 1",
          "sprint_goal": <sprint_goal>,
          "fecha_inicio": "2026-10-06",
          "fecha_fin": "2026-10-20"
        }
        """
      Cuando envio "POST" a "/api/proyectos/1/sprints" con ese cuerpo
      Entonces la respuesta tiene el estado HTTP "422"
      Y el cuerpo de la respuesta contiene el error "VALIDACION" con campo
        "sprint_goal" y el mensaje:
        """
        el Sprint Goal del sprint es obligatorio
        """

      Ejemplos:
        | variante        | sprint_goal |
        | vacio           | ""          |
        | solo espacios   | "   "       |

    @caso_limite
    Escenario: E-06 - Identificador de exactamente 50 caracteres
      # Regla RB-01 / CL-01: el borde superior de la longitud es valido.
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y sin sprints
      Y que preparo el cuerpo JSON del alta del sprint con "identificador" de
        50 caracteres y goal "Entregar el MVP"
      Cuando envio "POST" a "/api/proyectos/1/sprints" con ese cuerpo
      Entonces la respuesta tiene el estado HTTP "201"
      Y el "identificador" persistido tiene 50 caracteres

    @caso_error
    Escenario: E-07 - Identificador de 51 caracteres
      # Regla RB-01 / CL-02: un caracter mas se rechaza.
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y sin sprints
      Y que preparo el cuerpo JSON del alta del sprint con "identificador" de
        51 caracteres y goal "Entregar el MVP"
      Cuando envio "POST" a "/api/proyectos/1/sprints" con ese cuerpo
      Entonces la respuesta tiene el estado HTTP "422"
      Y el cuerpo de la respuesta contiene el error "VALIDACION" con campo
        "identificador" y el mensaje:
        """
        el identificador del sprint no puede superar los 50 caracteres
        """

    @caso_limite
    Escenario: E-08 - Sprint Goal de exactamente 300 caracteres
      # Regla RB-02 / CL-04: el borde superior del goal es valido.
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y sin sprints
      Y que preparo el cuerpo JSON del alta del sprint con identificador
        "Sprint 1" y "sprint_goal" de 300 caracteres
      Cuando envio "POST" a "/api/proyectos/1/sprints" con ese cuerpo
      Entonces la respuesta tiene el estado HTTP "201"
      Y el "sprint_goal" persistido tiene 300 caracteres

    @caso_error
    Escenario: E-09 - Sprint Goal de 301 caracteres
      # Regla RB-02 / CL-05: un caracter mas se rechaza.
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y sin sprints
      Y que preparo el cuerpo JSON del alta del sprint con identificador
        "Sprint 1" y "sprint_goal" de 301 caracteres
      Cuando envio "POST" a "/api/proyectos/1/sprints" con ese cuerpo
      Entonces la respuesta tiene el estado HTTP "422"
      Y el cuerpo de la respuesta contiene el error "VALIDACION" con campo
        "sprint_goal" y el mensaje:
        """
        el Sprint Goal del sprint no puede superar los 300 caracteres
        """

    @caso_error
    Escenario: E-10 - Fecha de fin anterior a la de inicio
      # Regla RB-03 / CL-08: la issue #30 exige rechazar el desorden de fechas
      # con un error de validacion sobre "fecha_fin".
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y sin sprints
      Y que preparo el cuerpo JSON del alta del sprint:
        """
        {
          "identificador": "Sprint 1",
          "sprint_goal": "Entregar el MVP",
          "fecha_inicio": "2026-10-06",
          "fecha_fin": "2026-10-05"
        }
        """
      Cuando envio "POST" a "/api/proyectos/1/sprints" con ese cuerpo
      Entonces la respuesta tiene el estado HTTP "422"
      Y el cuerpo de la respuesta contiene el error "VALIDACION" con campo
        "fecha_fin" y el mensaje:
        """
        la fecha de fin del sprint no puede ser anterior a la fecha de inicio
        """

    @caso_limite
    Escenario: E-11 - Sprint de un solo dia (fecha_fin igual a fecha_inicio)
      # Regla RB-03 / CL-07: la igualdad de fechas esta permitida.
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y sin sprints
      Y que preparo el cuerpo JSON del alta del sprint:
        """
        {
          "identificador": "Sprint 1",
          "sprint_goal": "Daily de ajuste",
          "fecha_inicio": "2026-10-06",
          "fecha_fin": "2026-10-06"
        }
        """
      Cuando envio "POST" a "/api/proyectos/1/sprints" con ese cuerpo
      Entonces la respuesta tiene el estado HTTP "201"
      Y el cuerpo de la respuesta contiene:
        """
        {
          "fecha_inicio": "2026-10-06",
          "fecha_fin": "2026-10-06",
          "estado": "Planificado"
        }
        """

    @caso_error
    Escenario: E-12 - Identificador duplicado dentro del mismo proyecto
      # Regla RB-04 / CL-09 (criterio adicional de la issue #30): no puede
      # existir otro "Sprint 1" en el mismo proyecto.
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y con un sprint "Sprint 1" de "2026-09-08" a
        "2026-09-22"
      Y que preparo el cuerpo JSON del alta del sprint:
        """
        {
          "identificador": "Sprint 1",
          "sprint_goal": "Reintento del MVP",
          "fecha_inicio": "2026-10-06",
          "fecha_fin": "2026-10-20"
        }
        """
      Cuando envio "POST" a "/api/proyectos/1/sprints" con ese cuerpo
      Entonces la respuesta tiene el estado HTTP "409"
      Y el cuerpo de la respuesta contiene el error "SPRINT_DUPLICADO" con
        campo "identificador" y el mensaje:
        """
        ya existe un sprint con ese identificador en el proyecto
        """

    @caso_alternativo
    Escenario: E-13 - Mismo identificador en otro proyecto esta permitido
      # Regla RB-04 / CL-10: la unicidad es por proyecto.
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y con un sprint "Sprint 1" de "2026-09-08" a
        "2026-09-22"
      Y que existe el proyecto "Proyecto Piloto" con "id" 2 en estado "Activo"
        y sin sprints
      Y que preparo el cuerpo JSON del alta del sprint:
        """
        {
          "identificador": "Sprint 1",
          "sprint_goal": "Primer ciclo del piloto",
          "fecha_inicio": "2026-09-08",
          "fecha_fin": "2026-09-22"
        }
        """
      Cuando envio "POST" a "/api/proyectos/2/sprints" con ese cuerpo
      Entonces la respuesta tiene el estado HTTP "201"
      Y el Sprint aparece en "GET /api/proyectos/2/sprints" con "total" 1

    @caso_error
    Escenario: E-14 - Solapamiento parcial de fechas con otro Sprint
      # Regla RB-07 / CL-11 (criterio adicional de la issue #30).
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y con un sprint "Sprint 1" de "2026-10-06" a
        "2026-10-20"
      Y que preparo el cuerpo JSON del alta del sprint:
        """
        {
          "identificador": "Sprint 2",
          "sprint_goal": "Continuar el MVP",
          "fecha_inicio": "2026-10-19",
          "fecha_fin": "2026-11-02"
        }
        """
      Cuando envio "POST" a "/api/proyectos/1/sprints" con ese cuerpo
      Entonces la respuesta tiene el estado HTTP "409"
      Y el cuerpo de la respuesta contiene el error
        "SPRINT_FECHAS_SOLAPADAS" y el mensaje:
        """
        las fechas del sprint se solapan con otro sprint del proyecto
        """

    @caso_error
    Escenario: E-15 - Un Sprint completamente contenido en el rango de otro
      # Regla RB-07 / CL-12: el solapamiento no necesita compartir los
      # extremos.
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y con un sprint "Sprint 1" de "2026-10-06" a
        "2026-10-20"
      Y que preparo el cuerpo JSON del alta del sprint:
        """
        {
          "identificador": "Sprint 1.5",
          "sprint_goal": "Sprint de emergencia",
          "fecha_inicio": "2026-10-10",
          "fecha_fin": "2026-10-12"
        }
        """
      Cuando envio "POST" a "/api/proyectos/1/sprints" con ese cuerpo
      Entonces la respuesta tiene el estado HTTP "409"
      Y el cuerpo de la respuesta contiene el error
        "SPRINT_FECHAS_SOLAPADAS"

    @caso_limite
    Escenario: E-16 - Sprints contiguos sin solapamiento
      # Regla RB-07 / CL-13 y NOTA-3: la fecha de fin del primer Sprint es
      # igual a la fecha de inicio del siguiente.
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y con un sprint "Sprint 1" de "2026-10-06" a
        "2026-10-20"
      Y que preparo el cuerpo JSON del alta del sprint:
        """
        {
          "identificador": "Sprint 2",
          "sprint_goal": "Continuar el MVP",
          "fecha_inicio": "2026-10-20",
          "fecha_fin": "2026-11-03"
        }
        """
      Cuando envio "POST" a "/api/proyectos/1/sprints" con ese cuerpo
      Entonces la respuesta tiene el estado HTTP "201"
      Y el Sprint aparece en "GET /api/proyectos/1/sprints" con "total" 2

    @caso_error
    Escenario: E-17 - Alta de un Sprint en un proyecto inexistente
      # Regla RB-10: el Sprint es un recurso anidado.
      Dado que preparo el cuerpo JSON del alta del sprint:
        """
        {
          "identificador": "Sprint 1",
          "sprint_goal": "Entregar el MVP",
          "fecha_inicio": "2026-10-06",
          "fecha_fin": "2026-10-20"
        }
        """
      Cuando envio "POST" a "/api/proyectos/99/sprints" con ese cuerpo
      Entonces la respuesta tiene el estado HTTP "404"
      Y el cuerpo de la respuesta contiene el error "PROYECTO_NO_ENCONTRADO"

    @caso_error
    Esquema del Escenario: E-18 - Fecha "<campo>" en la variante <variante>
      # Regla RB-03 / CL-21: ambas fechas son obligatorias y deben parsear
      # como fecha calendario "YYYY-MM-DD".
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y sin sprints
      Y que preparo el cuerpo JSON del alta del sprint con "fecha_inicio"
        "<fecha_inicio>" y "fecha_fin" "<fecha_fin>"
      Cuando envio "POST" a "/api/proyectos/1/sprints" con ese cuerpo
      Entonces la respuesta tiene el estado HTTP "400"
      Y el cuerpo de la respuesta contiene el error "PARAMETRO_INVALIDO" con
        campo "<campo>"

      Ejemplos:
        | variante            | campo        | fecha_inicio | fecha_fin    |
        | inicio ausente      | fecha_inicio |              | 2026-10-20   |
        | inicio mal formada  | fecha_inicio | 06-10-2026   | 2026-10-20   |
        | fin ausente         | fecha_fin    | 2026-10-06   |              |
        | fin mal formada     | fecha_fin    | 2026-10-06   | 20/10/2026   |

    @caso_error
    Escenario: E-19 - Identificador de proyecto no valido
      Dado que preparo el cuerpo JSON del alta del sprint:
        """
        {
          "identificador": "Sprint 1",
          "sprint_goal": "Entregar el MVP",
          "fecha_inicio": "2026-10-06",
          "fecha_fin": "2026-10-20"
        }
        """
      Cuando envio "POST" a "/api/proyectos/abc/sprints" con ese cuerpo
      Entonces la respuesta tiene el estado HTTP "400"
      Y el cuerpo de la respuesta contiene el error "PARAMETRO_INVALIDO"

    @caso_error
    Escenario: E-20 - Cuerpo del request mal formado en el alta
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y sin sprints
      Cuando envio "POST" a "/api/proyectos/1/sprints" con el cuerpo:
        """
        {"identificador":
        """
      Entonces la respuesta tiene el estado HTTP "400"
      Y el cuerpo de la respuesta contiene el error "JSON_INVALIDO"

    @caso_limite
    Escenario: E-21 - El cliente no puede elegir el "id" ni el "estado"
      # Regla RB-08 y NOTA-1: los campos calculados que lleguen en el body del
      # alta se ignoran (mismo criterio que E-04 de US-02 para id/codigo).
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y sin sprints
      Y que preparo el cuerpo JSON del alta del sprint:
        """
        {
          "id": 777,
          "estado": "Activo",
          "identificador": "Sprint 1",
          "sprint_goal": "Entregar el MVP",
          "fecha_inicio": "2026-10-06",
          "fecha_fin": "2026-10-20"
        }
        """
      Cuando envio "POST" a "/api/proyectos/1/sprints" con ese cuerpo
      Entonces la respuesta tiene el estado HTTP "201"
      Y el cuerpo de la respuesta contiene:
        """
        {
          "estado": "Planificado"
        }
        """
      Y el "id" del Sprint es generado por la base y no es 777

    @rfc
    Escenario: E-22 - Las marcas de auditoria se emiten en RFC 3339 UTC
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y sin sprints
      Y que preparo el cuerpo JSON del alta del sprint:
        """
        {
          "identificador": "Sprint 1",
          "sprint_goal": "Entregar el MVP",
          "fecha_inicio": "2026-10-06",
          "fecha_fin": "2026-10-20"
        }
        """
      Cuando envio "POST" a "/api/proyectos/1/sprints" con ese cuerpo
      Entonces la respuesta tiene el estado HTTP "201"
      Y los campos "creado_en" y "actualizado_en" cumplen el estandar RFC 3339
        en UTC

    @caso_normal
    Escenario: E-23 - Dos altas consecutivas generan Sprint 1 y Sprint 2
      # Caso normal complementario: la numeracion visible la elige el equipo
      # (a diferencia de US-02) pero ambos identificadores conviven en el
      # proyecto.
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y sin sprints
      Cuando envio "POST" a "/api/proyectos/1/sprints" con el cuerpo:
        """
        {
          "identificador": "Sprint 1",
          "sprint_goal": "MVP",
          "fecha_inicio": "2026-09-08",
          "fecha_fin": "2026-09-22"
        }
        """
      Y envio "POST" a "/api/proyectos/1/sprints" con el cuerpo:
        """
        {
          "identificador": "Sprint 2",
          "sprint_goal": "Metricas",
          "fecha_inicio": "2026-09-23",
          "fecha_fin": "2026-10-06"
        }
        """
      Entonces ambas respuestas tienen el estado HTTP "201"
      Y el Sprint aparece en "GET /api/proyectos/1/sprints" con "total" 2

    @caso_alternativo
    Escenario: E-24 - El listado de un proyecto no muestra los Sprints de otro
      # Regla RB-09: el listado es por proyecto.
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y con un sprint "Sprint 1" de "2026-09-08" a
        "2026-09-22"
      Y que existe el proyecto "Proyecto Piloto" con "id" 2 en estado "Activo"
        y con un sprint "Sprint 1" de "2026-10-06" a "2026-10-20"
      Cuando envio "GET" a "/api/proyectos/1/sprints"
      Entonces la respuesta tiene el estado HTTP "200"
      Y el cuerpo de la respuesta contiene:
        """
        {
          "total": 1,
          "sprints": [
            {
              "identificador": "Sprint 1",
              "fecha_inicio": "2026-09-08",
              "fecha_fin": "2026-09-22"
            }
          ]
        }
        """

  # ===========================================================================
  # LISTADO DE SPRINTS: GET /api/proyectos/{id}/sprints
  # ===========================================================================
  Regla: Listado de los Sprints de un proyecto

    @caso_normal
    Escenario: E-25 - Listado ordenado por fecha de inicio
      # Regla RB-09: el orden es fecha_inicio ASC con desempate por id ASC.
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y con los sprints:
        | identificador | sprint_goal | fecha_inicio | fecha_fin |
        | Sprint 2      | Metricas    | 2026-09-23   | 2026-10-06 |
        | Sprint 1      | MVP         | 2026-09-08   | 2026-09-22 |
      Cuando envio "GET" a "/api/proyectos/1/sprints"
      Entonces la respuesta tiene el estado HTTP "200"
      Y el cuerpo de la respuesta contiene:
        """
        {
          "total": 2,
          "sprints": [
            { "identificador": "Sprint 1", "estado": "Planificado" },
            { "identificador": "Sprint 2", "estado": "Planificado" }
          ]
        }
        """

    @caso_limite
    Escenario: E-26 - Listado de un proyecto sin Sprints
      # CL-18: el array vacio se serializa explicitamente.
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y sin sprints
      Cuando envio "GET" a "/api/proyectos/1/sprints"
      Entonces la respuesta tiene el estado HTTP "200"
      Y el cuerpo de la respuesta es:
        """
        {"total": 0, "sprints": []}
        """

    @caso_error
    Escenario: E-27 - Listado del backlog de un proyecto inexistente
      Cuando envio "GET" a "/api/proyectos/99/sprints"
      Entonces la respuesta tiene el estado HTTP "404"
      Y el cuerpo de la respuesta contiene el error "PROYECTO_NO_ENCONTRADO"

    @caso_error
    Escenario: E-28 - Identificador de proyecto no valido en el listado
      Cuando envio "GET" a "/api/proyectos/abc/sprints"
      Entonces la respuesta tiene el estado HTTP "400"
      Y el cuerpo de la respuesta contiene el error "PARAMETRO_INVALIDO"

  # ===========================================================================
  # CAMBIO DE ESTADO: PATCH /api/proyectos/{id}/sprints/{sprintId}/estado
  # ===========================================================================
  Regla: La maquina de estados del Sprint es Planificado -> Activo -> Cerrado

    @caso_normal
    Escenario: E-29 - Activar un Sprint Planificado
      # Caso normal de la issue #30 (CA-I4) y reglas RB-05, RB-06: el estado
      # pasa a "Activo" y pasa a ser el Sprint activo del proyecto.
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y con un sprint "Sprint 1" en estado "Planificado"
      Cuando envio "PATCH" a "/api/proyectos/1/sprints/1/estado" con el cuerpo:
        """
        {"estado": "Activo"}
        """
      Entonces la respuesta tiene el estado HTTP "200"
      Y el cuerpo de la respuesta contiene:
        """
        {
          "identificador": "Sprint 1",
          "estado": "Activo"
        }
        """
      Y el Sprint aparece como activo en "GET /api/proyectos/1/sprints"

    @caso_normal
    Escenario: E-30 - Cerrar un Sprint Activo
      # Regla RB-05: "Cerrado" es el estado terminal.
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y con un sprint "Sprint 1" en estado "Activo"
      Cuando envio "PATCH" a "/api/proyectos/1/sprints/1/estado" con el cuerpo:
        """
        {"estado": "Cerrado"}
        """
      Entonces la respuesta tiene el estado HTTP "200"
      Y el cuerpo de la respuesta contiene:
        """
        {
          "identificador": "Sprint 1",
          "estado": "Cerrado"
        }
        """

    @caso_error
    Esquema del Escenario: E-31 - Transicion invalida desde "<origen>" hacia "<destino>"
      # Reglas RB-05 / CL-15, CL-16: no se admite retroceder, saltear
      # estados, reabrir un Sprint cerrado ni persistir el mismo estado.
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y con un sprint "Sprint 1" en estado "<origen>"
      Cuando envio "PATCH" a "/api/proyectos/1/sprints/1/estado" con el cuerpo:
        """
        {"estado": "<destino>"}
        """
      Entonces la respuesta tiene el estado HTTP "409"
      Y el cuerpo de la respuesta contiene el error
        "SPRINT_TRANSICION_INVALIDA" y el mensaje:
        """
        la transición de estado del sprint no está permitida
        """

      Ejemplos:
        | origen      | destino     |
        | Activo      | Planificado |
        | Planificado | Cerrado     |
        | Cerrado     | Activo      |
        | Cerrado     | Planificado |
        | Activo      | Activo      |

    @caso_error
    Esquema del Escenario: E-32 - Cambio de estado con el literal "<estado>" fuera del conjunto
      # NOTA-4 / D-06: el literal fuera del conjunto cerrado responde 400.
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y con un sprint "Sprint 1" en estado "Planificado"
      Cuando envio "PATCH" a "/api/proyectos/1/sprints/1/estado" con el cuerpo:
        """
        {"estado": "<estado>"}
        """
      Entonces la respuesta tiene el estado HTTP "400"
      Y el cuerpo de la respuesta contiene el error "ESTADO_INVALIDO" con
        campo "estado" y el mensaje:
        """
        el estado del sprint debe ser "Planificado", "Activo" o "Cerrado"
        """

      Ejemplos:
        | estado       |
        | Terminado    |
        | planificado  |
        | activo       |
        | Cerrada      |

    @caso_error
    Escenario: E-33 - Cambio de estado con cuerpo JSON mal formado
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y con un sprint "Sprint 1" en estado "Planificado"
      Cuando envio "PATCH" a "/api/proyectos/1/sprints/1/estado" con el cuerpo:
        """
        {"estado":
        """
      Entonces la respuesta tiene el estado HTTP "400"
      Y el cuerpo de la respuesta contiene el error "JSON_INVALIDO"

    @caso_error
    Escenario: E-34 - Cambio de estado de un Sprint inexistente
      # Regla RB-11.
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y sin sprints
      Cuando envio "PATCH" a "/api/proyectos/1/sprints/5/estado" con el cuerpo:
        """
        {"estado": "Activo"}
        """
      Entonces la respuesta tiene el estado HTTP "404"
      Y el cuerpo de la respuesta contiene el error "SPRINT_NO_ENCONTRADA"

    @caso_error
    Escenario: E-35 - Cambio de estado de un Sprint de otro proyecto
      # Regla RB-11 / CL-19: un Sprint de otro proyecto es inexistente para
      # la ruta.
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y con un sprint "Sprint 1" en estado "Planificado"
      Y que existe el proyecto "Proyecto Piloto" con "id" 2 en estado "Activo"
        y sin sprints
      Cuando envio "PATCH" a "/api/proyectos/2/sprints/1/estado" con el cuerpo:
        """
        {"estado": "Activo"}
        """
      Entonces la respuesta tiene el estado HTTP "404"
      Y el cuerpo de la respuesta contiene el error "SPRINT_NO_ENCONTRADA"

    @caso_error
    Escenario: E-36 - Cambio de estado en un proyecto inexistente
      Cuando envio "PATCH" a "/api/proyectos/99/sprints/1/estado" con el
        cuerpo:
        """
        {"estado": "Activo"}
        """
      Entonces la respuesta tiene el estado HTTP "404"
      Y el cuerpo de la respuesta contiene el error "PROYECTO_NO_ENCONTRADO"

    @caso_error
    Escenario: E-37 - Identificador de Sprint no valido
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y sin sprints
      Cuando envio "PATCH" a "/api/proyectos/1/sprints/abc/estado" con el
        cuerpo:
        """
        {"estado": "Activo"}
        """
      Entonces la respuesta tiene el estado HTTP "400"
      Y el cuerpo de la respuesta contiene el error "PARAMETRO_INVALIDO"

    @caso_error
    Escenario: E-38 - Activar un Sprint cuando ya existe otro Activo
      # Regla RB-06 / CL-17 (criterio adicional de la issue #30).
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y con los sprints:
        | identificador | sprint_goal | fecha_inicio | fecha_fin  | estado   |
        | Sprint 1      | MVP         | 2026-09-08   | 2026-09-22 | Activo   |
        | Sprint 2      | Metricas    | 2026-09-23   | 2026-10-06 | Planificado |
      Cuando envio "PATCH" a "/api/proyectos/1/sprints/2/estado" con el cuerpo:
        """
        {"estado": "Activo"}
        """
      Entonces la respuesta tiene el estado HTTP "409"
      Y el cuerpo de la respuesta contiene el error "SPRINT_ACTIVO_EXISTENTE"
        y el mensaje:
        """
        ya existe un sprint activo en el proyecto: cerrá ese sprint antes de activar otro
        """
      Y el "Sprint 2" sigue en estado "Planificado" en
        "GET /api/proyectos/1/sprints"

    @caso_alternativo
    Escenario: E-39 - Cerrar el Sprint activo y luego activar el siguiente
      # Regla RB-06: la unica forma de tener otro Sprint Activo es cerrar el
      # anterior (maquina de estados encadenada).
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y con los sprints:
        | identificador | sprint_goal | fecha_inicio | fecha_fin  | estado   |
        | Sprint 1      | MVP         | 2026-09-08   | 2026-09-22 | Activo   |
        | Sprint 2      | Metricas    | 2026-09-23   | 2026-10-06 | Planificado |
      Cuando envio "PATCH" a "/api/proyectos/1/sprints/1/estado" con el cuerpo:
        """
        {"estado": "Cerrado"}
        """
      Y envio "PATCH" a "/api/proyectos/1/sprints/2/estado" con el cuerpo:
        """
        {"estado": "Activo"}
        """
      Entonces ambas respuestas tienen el estado HTTP "200"
      Y el Sprint aparece como activo en "GET /api/proyectos/1/sprints"

    @caso_alternativo
    Escenario: E-40 - El cambio de estado refresca la marca de auditoria
      # @rfc implicito: actualizado_en avanza y creado_en se conserva.
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y con un sprint "Sprint 1" en estado "Activo"
      Cuando envio "PATCH" a "/api/proyectos/1/sprints/1/estado" con el cuerpo:
        """
        {"estado": "Cerrado"}
        """
      Entonces la respuesta tiene el estado HTTP "200"
      Y los campos "creado_en" y "actualizado_en" cumplen el estandar RFC 3339
        en UTC
      Y "actualizado_en" no es anterior a "creado_en"

    @caso_limite
    Escenario: E-41 - Cerrar un Sprint no impide listar su historial
      # CL-18 complementario: un Sprint cerrado sigue siendo visible en el
      # listado del proyecto.
      Dado que existe el proyecto "Software Metrics & Estimation" con "id" 1
        en estado "Activo" y con un sprint "Sprint 1" en estado "Cerrado"
      Cuando envio "GET" a "/api/proyectos/1/sprints"
      Entonces la respuesta tiene el estado HTTP "200"
      Y el cuerpo de la respuesta contiene:
        """
        {
          "total": 1,
          "sprints": [
            { "identificador": "Sprint 1", "estado": "Cerrado" }
          ]
        }
        """
