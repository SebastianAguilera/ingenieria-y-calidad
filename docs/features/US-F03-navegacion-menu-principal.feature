# language: es
# =============================================================================
# US-F03 - Navegación con el menú principal
# Escenarios BDD (Gherkin)
#
# Issue de origen     : SebastianAguilera/ingenieria-y-calidad#52
# Especificacion SDD  : docs/specs/US-F03-navegacion-menu-principal-spec.md
# Metodologia         : docs/methodology.md seccion 4 (BDD)
# Rama de trabajo     : feature/US-F03-navegacion-menu-principal
# Trazabilidad        : Issue #52 -> SDD -> Criterios -> Escenarios BDD -> Tests Vitest -> Codigo
#
# Entorno: frontend Vite http://localhost:5173
# Tests: frontend/src/**/*.test.ts(x) con Vitest + React Testing Library
# -----------------------------------------------------------------------------

Feature: Navegación con el menú principal
  Como usuario del panel
  Quiero un menú lateral con las secciones del panel
  Para moverme entre proyectos y usuarios sin perder de vista en qué parte de la aplicación estoy

  # ---------------------------------------------------------------------------
  # Layout fijo y barra lateral global
  # ---------------------------------------------------------------------------

  @caso_normal
  Escenario: E-01 Barra superior y secciones globales visibles en cualquier pantalla
    Dado que estoy logueado en el panel
    Cuando miro cualquier pantalla del panel, por ejemplo "/proyectos"
    Entonces veo la barra superior con el nombre de la aplicación y mi nombre de usuario
    Y veo la barra lateral con los enlaces "Proyectos" y "Usuarios"

  @caso_normal
  Escenario: E-02 Navegación desde Proyectos hacia Usuarios
    Dado que estoy en "/proyectos"
    Cuando hago clic en "Usuarios"
    Entonces la URL cambia a "/usuarios"
    Y el enlace "Usuarios" queda marcado como activo

  @caso_alternativo
  Escenario: E-03 Navegación desde Usuarios de vuelta a Proyectos
    Dado que estoy en "/usuarios" con la sección "Usuarios" activa
    Cuando hago clic en "Proyectos"
    Entonces la URL cambia a "/proyectos"
    Y el enlace "Proyectos" queda marcado como activo

  # ---------------------------------------------------------------------------
  # Secciones de proyecto (rutas anidadas)
  # ---------------------------------------------------------------------------

  @caso_normal
  Escenario: E-04 Secciones visibles al entrar a un proyecto
    Dado que entro a "/proyectos/1"
    Cuando se muestra la barra lateral del proyecto
    Entonces aparecen las secciones "Resumen", "Backlog", "Sprints", "Equipo", "Defectos" y "Métricas"
    Y "Resumen" queda marcado como activa

  @caso_normal
  Escenario: E-05 Sección activa según la ruta actual del proyecto
    Dado que estoy en "/proyectos/1/backlog"
    Cuando miro la barra lateral
    Entonces "Backlog" queda marcado como activo
    Y las demás secciones del proyecto no están marcadas como activas

  @caso_alternativo
  Escenario: E-06 Vuelta al listado de proyectos desde una sección
    Dado que estoy en una ruta de proyecto, por ejemplo "/proyectos/1/sprints"
    Cuando hago clic en "Proyectos"
    Entonces vuelvo al listado en "/proyectos"
    Y la barra lateral vuelve a mostrar las secciones globales "Proyectos" y "Usuarios"

  @caso_alternativo
  Escenario: E-07 Cambio entre secciones del mismo proyecto
    Dado que estoy en "/proyectos/1/backlog" con "Backlog" activa
    Cuando hago clic en "Métricas"
    Entonces la URL cambia a "/proyectos/1/metricas"
    Y "Métricas" queda marcado como activo
    Y "Backlog" deja de estar marcado como activo

  # ---------------------------------------------------------------------------
  # Responsive: colapso bajo 768px
  # ---------------------------------------------------------------------------

  @caso_limite
  Escenario: E-08 Barra lateral colapsada en pantallas de 767px
    Dado que la pantalla tiene 767px de ancho
    Cuando miro la barra lateral
    Entonces la barra lateral está colapsada
    Y la abro con un ícono de menú accesible

  @caso_limite
  Escenario: E-09 Barra lateral expandida en pantallas de 768px
    Dado que la pantalla tiene 768px de ancho
    Cuando miro la barra lateral
    Entonces la barra lateral está expandida de forma permanente
    Y no necesito el ícono de menú para ver las secciones

  @caso_alternativo
  Escenario: E-10 Apertura y cierre de la barra lateral en móvil
    Dado que la pantalla tiene menos de 768px de ancho y la barra lateral está colapsada
    Cuando hago clic en el ícono de menú
    Entonces se abre la barra lateral con las secciones correspondientes
    Y el botón indica con "aria-expanded" que está expandido
    Cuando vuelvo a hacer clic en el ícono de menú
    Entonces se cierra la barra lateral

  # ---------------------------------------------------------------------------
  # Accesibilidad: teclado
  # ---------------------------------------------------------------------------

  @caso_alternativo
  Escenario: E-11 Navegación por teclado entre secciones del menú
    Dado que uso solo el teclado y el foco está en la barra lateral
    Cuando recorro los enlaces con la tecla Tab
    Entonces el foco avanza entre "Proyectos" y "Usuarios" con foco visible
    Y al presionar Enter sobre "Usuarios" la URL cambia a "/usuarios"

  @caso_alternativo
  Escenario: E-12 Apertura y cierre de la barra lateral con el teclado
    Dado que la pantalla tiene menos de 768px de ancho y la barra lateral está colapsada
    Cuando llego al botón de menú con la tecla Tab y presiono Enter
    Entonces se abre la barra lateral
    Y el botón tiene "aria-expanded" en verdadero y "aria-controls" apuntando a la barra lateral
    Cuando presiono Enter nuevamente
    Entonces se cierra la barra lateral

  # ---------------------------------------------------------------------------
  # Rutas inexistentes y errores
  # ---------------------------------------------------------------------------

  @caso_error
  Escenario: E-13 Pantalla 404 para una ruta inexistente
    Dado que ingreso a una ruta inexistente, por ejemplo "/esto-no-existe"
    Cuando la aplicación carga
    Entonces veo una pantalla 404
    Y veo un enlace que me lleva a "/proyectos"

  @caso_error
  @escenario_diferido
  Escenario: E-14 Ruta de proyecto inexistente
    # NOTA-1: validar que el id del proyecto exista requiere la API de detalle
    # (US-F06); en esta historia /proyectos/:id renderiza el placeholder.
    Dado que ingreso a "/proyectos/999/backlog" con un proyecto que no existe
    Cuando la aplicación carga
    Entonces veo la sección Backlog en modo placeholder
    Y veo un enlace que me lleva a "/proyectos"

  @caso_error
  Escenario: E-15 Sección inexistente dentro de un proyecto
    Dado que ingreso a "/proyectos/1/inventada"
    Cuando la aplicación carga
    Entonces veo una pantalla 404
    Y veo un enlace que me lleva a "/proyectos"
