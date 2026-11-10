import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, useLocation } from 'react-router-dom'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import type { Usuario } from '../api/types'
import { RutasPanel } from '../App'
import { CLAVE_SESION } from '../session/sesion'
import { SesionProvider } from '../session/SesionContext'
import { ThemeProvider } from '../theme/ThemeProvider'

const ana: Usuario = {
  id: 1,
  nombre: 'Ana Pérez',
  email: 'ana@example.com',
  creado_en: '2026-01-01T00:00:00Z',
}

const SECCIONES_PROYECTO = ['Resumen', 'Backlog', 'Sprints', 'Equipo', 'Defectos', 'Métricas']

const matchMediaOriginal = window.matchMedia

function simularAnchoDePantalla(ancho: number) {
  window.matchMedia = ((consulta: string) => ({
    matches: consulta === '(max-width: 767px)' && ancho < 768,
    media: consulta,
    onchange: null,
    addEventListener: () => {},
    removeEventListener: () => {},
    addListener: () => {},
    removeListener: () => {},
    dispatchEvent: () => false,
  })) as unknown as typeof window.matchMedia
}

function RutaActual() {
  const { pathname } = useLocation()
  return <p data-testid="ruta-actual">{pathname}</p>
}

function renderPanel(ruta = '/proyectos') {
  return render(
    <ThemeProvider>
      <SesionProvider>
        <MemoryRouter initialEntries={[ruta]}>
          <RutaActual />
          <RutasPanel />
        </MemoryRouter>
      </SesionProvider>
    </ThemeProvider>,
  )
}

async function tabHasta(usuario: ReturnType<typeof userEvent.setup>, limite: number, esEnfoque: () => boolean) {
  for (let intento = 0; intento < limite; intento += 1) {
    if (esEnfoque()) {
      return
    }
    await usuario.tab()
  }
}

describe('AppShell - navegación del menú principal (US-F03)', () => {
  beforeEach(() => {
    localStorage.clear()
    localStorage.setItem(CLAVE_SESION, JSON.stringify(ana))
    simularAnchoDePantalla(1440)
  })

  afterEach(() => {
    window.matchMedia = matchMediaOriginal
  })

  // ---------------------------------------------------------------------------
  // Barra superior y secciones globales
  // ---------------------------------------------------------------------------

  it('E-01: barra superior con el usuario y barra lateral con Proyectos y Usuarios', () => {
    renderPanel('/proyectos')

    expect(screen.getByText('Software Metrics')).toBeInTheDocument()
    expect(screen.getByText('Ana Pérez')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Salir' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Proyectos' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Usuarios' })).toBeInTheDocument()
    expect(
      screen.queryByRole('navigation', { name: 'Secciones del proyecto' }),
    ).not.toBeInTheDocument()
  })

  it('E-02: clic en Usuarios cambia la URL y lo marca como activo', async () => {
    const usuario = userEvent.setup()
    renderPanel('/proyectos')

    await usuario.click(screen.getByRole('link', { name: 'Usuarios' }))

    expect(screen.getByTestId('ruta-actual')).toHaveTextContent('/usuarios')
    expect(screen.getByRole('link', { name: 'Usuarios' })).toHaveAttribute('aria-current', 'page')
    expect(screen.getByRole('link', { name: 'Proyectos' })).not.toHaveAttribute('aria-current')
  })

  it('E-03: clic en Proyectos desde Usuarios vuelve al listado', async () => {
    const usuario = userEvent.setup()
    renderPanel('/usuarios')

    await usuario.click(screen.getByRole('link', { name: 'Proyectos' }))

    expect(screen.getByTestId('ruta-actual')).toHaveTextContent('/proyectos')
    expect(screen.getByRole('link', { name: 'Proyectos' })).toHaveAttribute('aria-current', 'page')
    expect(screen.getByRole('link', { name: 'Usuarios' })).not.toHaveAttribute('aria-current')
  })

  // ---------------------------------------------------------------------------
  // Secciones de proyecto (rutas anidadas)
  // ---------------------------------------------------------------------------

  it('E-04: al entrar a /proyectos/1 aparecen las secciones del proyecto', () => {
    renderPanel('/proyectos/1')

    for (const seccion of SECCIONES_PROYECTO) {
      expect(screen.getByRole('link', { name: seccion })).toBeInTheDocument()
    }
    expect(screen.getByRole('link', { name: 'Resumen' })).toHaveAttribute('aria-current', 'page')
    expect(screen.getByRole('link', { name: 'Proyectos' })).toBeInTheDocument()
  })

  it('E-05: en /proyectos/1/backlog el enlace activo es Backlog', () => {
    renderPanel('/proyectos/1/backlog')

    expect(screen.getByRole('link', { name: 'Backlog' })).toHaveAttribute('aria-current', 'page')
    expect(screen.getByRole('link', { name: 'Resumen' })).not.toHaveAttribute('aria-current')
    expect(screen.getByRole('link', { name: 'Métricas' })).not.toHaveAttribute('aria-current')
    expect(screen.getByRole('link', { name: 'Sprints' })).not.toHaveAttribute('aria-current')
  })

  it('E-06: clic en Proyectos desde una ruta de proyecto vuelve a las secciones globales', async () => {
    const usuario = userEvent.setup()
    renderPanel('/proyectos/1/sprints')

    await usuario.click(screen.getByRole('link', { name: 'Proyectos' }))

    expect(screen.getByTestId('ruta-actual')).toHaveTextContent('/proyectos')
    expect(screen.queryByRole('link', { name: 'Sprints' })).not.toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Usuarios' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Proyectos' })).toHaveAttribute('aria-current', 'page')
  })

  it('E-07: cambio entre secciones del mismo proyecto', async () => {
    const usuario = userEvent.setup()
    renderPanel('/proyectos/1/backlog')

    await usuario.click(screen.getByRole('link', { name: 'Métricas' }))

    expect(screen.getByTestId('ruta-actual')).toHaveTextContent('/proyectos/1/metricas')
    expect(screen.getByRole('link', { name: 'Métricas' })).toHaveAttribute('aria-current', 'page')
    expect(screen.getByRole('link', { name: 'Backlog' })).not.toHaveAttribute('aria-current')
  })

  // ---------------------------------------------------------------------------
  // Responsive: colapso bajo 768px
  // ---------------------------------------------------------------------------

  it('E-08: en 767px la barra lateral inicia colapsada y se abre con el botón de menú', async () => {
    const usuario = userEvent.setup()
    simularAnchoDePantalla(767)
    renderPanel('/proyectos')

    expect(
      screen.queryByRole('navigation', { name: 'Secciones del panel' }),
    ).not.toBeInTheDocument()

    const boton = screen.getByRole('button', { name: /menú/i })
    expect(boton).toHaveAttribute('aria-expanded', 'false')
    expect(boton).toHaveAttribute('aria-controls', 'barra-lateral')

    await usuario.click(boton)

    expect(boton).toHaveAttribute('aria-expanded', 'true')
    expect(screen.getByRole('navigation', { name: 'Secciones del panel' })).toBeInTheDocument()
  })

  it('E-09: en 768px la barra lateral está expandida y no se muestra el botón de menú', () => {
    simularAnchoDePantalla(768)
    renderPanel('/proyectos')

    expect(screen.getByRole('navigation', { name: 'Secciones del panel' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /menú/i })).not.toBeInTheDocument()
  })

  it('E-10: apertura y cierre de la barra lateral con el botón de menú', async () => {
    const usuario = userEvent.setup()
    simularAnchoDePantalla(767)
    renderPanel('/proyectos')

    const boton = screen.getByRole('button', { name: /menú/i })

    await usuario.click(boton)
    expect(boton).toHaveAttribute('aria-expanded', 'true')
    expect(screen.getByRole('navigation', { name: 'Secciones del panel' })).toBeInTheDocument()

    await usuario.click(boton)
    expect(boton).toHaveAttribute('aria-expanded', 'false')
    expect(
      screen.queryByRole('navigation', { name: 'Secciones del panel' }),
    ).not.toBeInTheDocument()
  })

  // ---------------------------------------------------------------------------
  // Accesibilidad: teclado
  // ---------------------------------------------------------------------------

  it('E-11: se navega hasta Usuarios con Tab y Enter cambia la URL', async () => {
    const usuario = userEvent.setup()
    renderPanel('/proyectos')

    await tabHasta(usuario, 12, () => document.activeElement?.textContent?.trim() === 'Usuarios')

    expect(document.activeElement?.textContent?.trim()).toBe('Usuarios')
    await usuario.keyboard('{Enter}')

    expect(screen.getByTestId('ruta-actual')).toHaveTextContent('/usuarios')
  })

  it('E-12: el botón de menú se abre y cierra con teclado y expone ARIA', async () => {
    const usuario = userEvent.setup()
    simularAnchoDePantalla(767)
    renderPanel('/proyectos')

    await tabHasta(
      usuario,
      12,
      () => document.activeElement?.getAttribute('aria-controls') === 'barra-lateral',
    )

    const boton = screen.getByRole('button', { name: /menú/i })
    expect(document.activeElement).toBe(boton)

    await usuario.keyboard('{Enter}')
    expect(boton).toHaveAttribute('aria-expanded', 'true')
    expect(screen.getByRole('navigation', { name: 'Secciones del panel' })).toBeInTheDocument()

    await usuario.keyboard('{Enter}')
    expect(boton).toHaveAttribute('aria-expanded', 'false')
    expect(
      screen.queryByRole('navigation', { name: 'Secciones del panel' }),
    ).not.toBeInTheDocument()
  })

  // ---------------------------------------------------------------------------
  // Rutas inexistentes y sesión
  // ---------------------------------------------------------------------------

  it('E-16: cada enlace de la navegación incluye un icono decorativo de Bootstrap Icons', () => {
    renderPanel('/proyectos/1')

    const enlaces = [
      ...screen.getAllByRole('link').filter((enlace) => enlace.closest('aside')),
    ]

    expect(enlaces.length).toBeGreaterThan(0)
    for (const enlace of enlaces) {
      expect(enlace.querySelector('i.bi[aria-hidden="true"]')).not.toBeNull()
    }
  })

  it('E-13: una ruta inexistente muestra el 404 con enlace a /proyectos', async () => {
    const usuario = userEvent.setup()
    renderPanel('/esto-no-existe')

    expect(screen.getByRole('heading', { name: '404' })).toBeInTheDocument()

    const enlace = screen.getByRole('link', { name: 'Volver al panel' })
    expect(enlace).toHaveAttribute('href', '/proyectos')

    await usuario.click(enlace)
    expect(screen.getByTestId('ruta-actual')).toHaveTextContent('/proyectos')
  })

  it('E-14 (diferida a US-F06): /proyectos/:id inexistente renderiza el placeholder', () => {
    renderPanel('/proyectos/999/backlog')

    expect(screen.getByRole('heading', { level: 1, name: 'Backlog' })).toBeInTheDocument()
    expect(screen.queryByRole('heading', { name: '404' })).not.toBeInTheDocument()
  })

  it('E-15: una sección inexistente dentro de un proyecto muestra el 404', () => {
    renderPanel('/proyectos/1/inventada')

    expect(screen.getByRole('heading', { name: '404' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Volver al panel' })).toHaveAttribute(
      'href',
      '/proyectos',
    )
  })

  it('el botón Salir limpia la sesión y lleva al login (regresión US-F02)', async () => {
    const usuario = userEvent.setup()
    renderPanel('/proyectos')

    await usuario.click(screen.getByRole('button', { name: 'Salir' }))

    expect(localStorage.getItem(CLAVE_SESION)).toBeNull()
    expect(screen.getByRole('heading', { name: 'Ingresá al panel' })).toBeInTheDocument()
  })
})
