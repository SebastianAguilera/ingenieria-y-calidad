# language: es
# =============================================================================
# US-F06 - Ver el detalle de un proyecto en formato tabla
# Escenarios BDD (Gherkin)
#
# Issue de origen     : SebastianAguilera/ingenieria-y-calidad#55
# Especificacion SDD  : docs/specs/US-F06-detalle-proyecto-spec.md
# Metodologia         : docs/methodology.md seccion 4 (BDD)
# Rama de trabajo     : feature/US-F06-detalle-proyecto
# Trazabilidad        : Issue #55 -> SDD -> Criterios -> Escenarios BDD -> Tests Vitest -> Codigo
#
# Entorno: frontend Vite http://localhost:5173
# Tests: frontend/src/**/*.test.tsx con Vitest + React Testing Library
#
# NOTA-1: el texto del enlace "Volver a proyectos" es propuesto en la spec;
#         si se ajusta en la implementacion, se actualizan spec, feature y
#         tests en el mismo commit.
# -----------------------------------------------------------------------------

Feature: Ver el detalle de un proyecto en formato tabla
  Como usuario
  Quiero abrir un proyecto y ver sus datos en una tabla, junto con un menú para cambiar entre sus secciones
  Para tener toda la información del proyecto en un solo lugar y con un formato fácil de leer

  # ---------------------------------------------------------------------------
  # Carga del proyecto, encabezado y sección Resumen
  # ---------------------------------------------------------------------------

  @caso_normal
  Escenario: E-01 Resumen con la tabla completa del proyecto
    # CA-01 · RB-02, RB-04, RB-10, RB-11, RB-12, RB-14, RB-17
    Dado que abro "/proyectos/1" para el proyecto activo "Sistema de Métricas" con 3 integrantes
    Cuando la página termina de cargar
    Entonces veo en el encabezado del detalle el nombre "Sistema de Métricas"
    Y no veo ninguna etiqueta "Cerrado" en el encabezado
    Y en Resumen veo una tabla de dos columnas con los encabezados "Campo" y "Valor"
    Y veo las filas "Nombre", "Estado", "Fecha de inicio", "Fecha de fin" y "Cantidad de integrantes" en ese orden
    Y la fila "Estado" muestra "Activo"
    Y la fila "Cantidad de integrantes" muestra "3"
    Y no veo ninguna fila "Descripción"

  @caso_limite
  Escenario: E-02 Fecha de fin sin definir en Resumen
    # CA-02 · RB-13
    Dado que abro el detalle de un proyecto que no tiene fecha de fin
    Cuando la sección Resumen termina de cargar
    Entonces veo la fila "Fecha de fin"
    Y su valor es "Sin fecha"
    Y el resto de la tabla de Resumen se muestra normalmente

  @caso_alternativo
  Escenario: E-03 Etiqueta "Cerrado" en el encabezado de un proyecto cerrado
    # CA-03 · RB-11
    Dado que abro "/proyectos/2" para el proyecto "Migración Legacy" con estado "Cerrado"
    Cuando la página termina de cargar
    Entonces veo en el encabezado del detalle el nombre "Migración Legacy"
    Y veo la etiqueta "Cerrado" al lado del nombre
    Y la etiqueta sigue visible al pasar a otra sección del mismo proyecto

  # ---------------------------------------------------------------------------
  # Navegación entre secciones del mismo proyecto
  # ---------------------------------------------------------------------------

  @caso_normal
  Escenario: E-04 Cambio de sección sin recargar la página ni volver a cargar el proyecto
    # CA-04 · RB-01, RB-10, RB-19
    Dado que estoy en "/proyectos/1" con el proyecto ya cargado y la sección Resumen activa
    Cuando hago clic en "Equipo"
    Entonces la URL cambia a "/proyectos/1/equipo" sin recargar la página
    Y el contenido cambia de Resumen a Equipo
    Y el encabezado del detalle sigue mostrando el mismo nombre
    Y no veo "Cargando proyecto…" porque el proyecto no se vuelve a pedir
    Cuando hago clic en "Métricas"
    Entonces la URL cambia a "/proyectos/1/metricas" sin recargar la página
    Y el contenido se actualiza a la sección Métricas sin volver a cargar el proyecto

  # ---------------------------------------------------------------------------
  # Sección Equipo
  # ---------------------------------------------------------------------------

  @caso_normal
  Escenario: E-05 Equipo con la tabla de integrantes
    # CA-05 · RB-15, RB-17
    Dado que abro la sección Equipo de un proyecto con los integrantes "Ana Ruiz" y "Bruno Paz"
    Cuando la página termina de cargar
    Entonces veo una tabla con las columnas "Nombre", "Email" y "Registrado el"
    Y veo una fila por integrante con su nombre y su email
    Y la columna "Registrado el" muestra la fecha de alta de cada integrante en el sistema
    Y no veo ninguna columna "Rol"

  @caso_limite
  Escenario: E-06 Equipo sin integrantes
    # CA-06 · RB-14, RB-16
    Dado que abro la sección Equipo de un proyecto sin integrantes
    Cuando la página termina de cargar
    Entonces veo el mensaje "Este proyecto todavía no tiene integrantes"
    Y no veo la tabla de integrantes ni sus encabezados
    Y en Resumen la fila "Cantidad de integrantes" muestra "0"

  # ---------------------------------------------------------------------------
  # Proyecto inexistente y errores de carga
  # ---------------------------------------------------------------------------

  @caso_error
  Escenario: E-07 Proyecto inexistente con respuesta 404
    # CA-07 · RB-09, RB-10
    Dado que abro "/proyectos/999" para un proyecto que no existe y el servidor responde 404
    Cuando la página termina de cargar
    Entonces veo el mensaje "Proyecto no encontrado"
    Y veo el enlace "Volver a proyectos" que me lleva a "/proyectos"
    Y no veo el botón "Reintentar"
    Y no veo ninguna tabla de Resumen debajo del mensaje

  @caso_error
  Escenario: E-08 Error genérico de carga con botón "Reintentar"
    # CA-08 · RB-07, RB-10
    Dado que abro el detalle de un proyecto y la API falla por un motivo distinto de 404, por ejemplo porque se cayó la red o el servidor respondió con un error 500
    Cuando la página termina de cargar
    Entonces veo el mensaje de error "No se pudo cargar el proyecto. Intentá nuevamente."
    Y veo el botón "Reintentar"
    Y no veo ninguna tabla ni contenido de sección debajo del mensaje

  # ---------------------------------------------------------------------------
  # Estados de carga y reintento
  # ---------------------------------------------------------------------------

  @caso_normal
  Escenario: E-09 Estado de carga mientras se pide el proyecto
    # CA-01 · RB-05, RB-10
    Dado que abro el detalle de un proyecto y el servidor todavía no respondió
    Cuando miro la pantalla mientras espera la respuesta
    Entonces veo "Cargando proyecto…"
    Y no veo ninguna tabla de Resumen ni contenido de sección
    Cuando llega la respuesta del servidor
    Entonces desaparece "Cargando proyecto…" y aparece el detalle del proyecto

  @caso_alternativo
  Escenario: E-10 Reintento exitoso después de un error
    # CA-08 · RB-06, RB-08
    Dado que estoy en el detalle con el mensaje de error y el botón "Reintentar" visibles
    Cuando hago clic en "Reintentar" y la API ya responde bien
    Entonces vuelvo a ver "Cargando proyecto…" mientras se pide el proyecto
    Y al terminar la carga veo el encabezado y la tabla de Resumen
    Y el mensaje de error desaparece

  # ---------------------------------------------------------------------------
  # Id inválido en la URL
  # ---------------------------------------------------------------------------

  @caso_error
  Escenario: E-11 Id inválido en la URL del detalle
    # CA-07 · RB-03
    Dado que ingreso a la URL del detalle con un id que no es un entero positivo, por ejemplo "/proyectos/abc", "/proyectos/1.5" o "/proyectos/0"
    Cuando la página carga
    Entonces veo "Proyecto no encontrado" con el enlace "Volver a proyectos"
    Y el mensaje aparece sin pasar por "Cargando proyecto…"
    Y no se hace ninguna petición al servidor para buscar el proyecto
    Y no veo el botón "Reintentar"

  # ---------------------------------------------------------------------------
  # Casos límite: pantallas chicas y formato de fechas
  # ---------------------------------------------------------------------------

  @caso_limite
  Escenario: E-12 Tablas con desplazamiento horizontal en pantallas chicas
    # RB-18, RB-20 · verificación manual; en Vitest se comprueban las clases del contenedor
    Dado que abro el detalle en una pantalla chica, por ejemplo de 375px de ancho
    Cuando miro la tabla de Resumen o la de Equipo
    Entonces la tabla se desplaza hacia los costados dentro de su tarjeta
    Y el resto del layout de la página no se rompe
    Y un nombre de proyecto muy largo se ajusta en el encabezado sin romper el layout
    Y con muchos integrantes la tabla de Equipo crece hacia abajo manteniendo el mismo desplazamiento

  @caso_limite @rfc
  Escenario: E-13 Fechas en formato dd/mm/yyyy en Resumen y Equipo
    # CA-01, CA-05 · RB-13, RB-15
    Dado que el servidor devuelve el proyecto con fecha de inicio "2026-03-02" y fecha de fin "2026-06-30" en ISO 8601
    Y sus integrantes tienen fecha de alta "2026-01-01T00:00:00Z" y "2025-11-20T14:30:00Z"
    Cuando abro Resumen y luego la sección Equipo
    Entonces la fila "Fecha de inicio" muestra "02/03/2026"
    Y la fila "Fecha de fin" muestra "30/06/2026"
    Y la columna "Registrado el" muestra "01/01/2026" y "20/11/2025"

  # ---------------------------------------------------------------------------
  # Deep links directos a una sección
  # ---------------------------------------------------------------------------

  @caso_alternativo
  Escenario: E-14 Deep link directo a una sección con proyecto válido
    # CA-04, CA-05 · RB-01, RB-19
    Dado que ingreso directamente a "/proyectos/7/equipo" sin pasar por el listado ni por Resumen
    Cuando la página termina de cargar
    Entonces veo "Cargando proyecto…" y luego el encabezado con el nombre del proyecto
    Y la sección Equipo aparece ya con su tabla de integrantes
    Y el enlace "Equipo" del menú queda marcado como activo

  @caso_error
  Escenario: E-15 Deep link a una sección de un proyecto inexistente
    # CA-07 · RB-09, RB-10
    Dado que ingreso directamente a "/proyectos/999/equipo" de un proyecto que no existe
    Cuando la página termina de cargar
    Entonces veo "Proyecto no encontrado" con el enlace "Volver a proyectos"
    Y no veo la tabla de Equipo ni ningún contenido de sección debajo del mensaje
