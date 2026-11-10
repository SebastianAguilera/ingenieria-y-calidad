import { useEffect, useState, useSyncExternalStore } from 'react'
import { Link, NavLink, Outlet, useLocation, useNavigate } from 'react-router-dom'
import { useSesion } from '../session/sesion'
import ModeSwitch from '../ui/ModeSwitch'
import BarraLateralProyecto from './BarraLateralProyecto'
import { claseEnlaceDeMenu } from './enlaces'

const CONSULTA_MOVIL = '(max-width: 767px)'

function suscribirACambioDePantalla(onChange: () => void) {
  const consulta = window.matchMedia(CONSULTA_MOVIL)
  consulta.addEventListener('change', onChange)
  return () => consulta.removeEventListener('change', onChange)
}

function useEsPantallaMovil(): boolean {
  return useSyncExternalStore(
    suscribirACambioDePantalla,
    () => window.matchMedia(CONSULTA_MOVIL).matches,
    () => false,
  )
}

export default function AppShell() {
  const { usuarioActivo, salir } = useSesion()
  const navegar = useNavigate()
  const ubicacion = useLocation()
  const esMovil = useEsPantallaMovil()
  const [menuAbierto, setMenuAbierto] = useState(false)
  const [rutaPrevia, setRutaPrevia] = useState(ubicacion.pathname)

  if (rutaPrevia !== ubicacion.pathname) {
    setRutaPrevia(ubicacion.pathname)
    setMenuAbierto(false)
  }

  useEffect(() => {
    if (!menuAbierto) {
      return undefined
    }
    function manejarTecla(evento: KeyboardEvent) {
      if (evento.key === 'Escape') {
        setMenuAbierto(false)
      }
    }
    window.addEventListener('keydown', manejarTecla)
    return () => window.removeEventListener('keydown', manejarTecla)
  }, [menuAbierto])

  function manejarSalir() {
    salir()
    navegar('/login', { replace: true })
  }

  const colapsada = esMovil && !menuAbierto

  return (
    <div className="flex min-h-screen flex-col bg-page text-ink">
      <header className="border-b border-border bg-surface">
        <div className="flex flex-wrap items-center justify-between gap-3 px-6 py-4">
          <div className="flex items-center gap-3">
            {esMovil && (
              <button
                type="button"
                onClick={() => setMenuAbierto((abierto) => !abierto)}
                aria-expanded={menuAbierto}
                aria-controls="barra-lateral"
                aria-label="Menú de secciones"
                className="inline-flex h-10 w-10 items-center justify-center rounded-pill border border-border text-ink transition hover:bg-primary-100"
              >
                <svg
                  aria-hidden="true"
                  viewBox="0 0 24 24"
                  className="h-5 w-5"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="1.8"
                >
                  <path d="M4 7h16M4 12h16M4 17h16" />
                </svg>
              </button>
            )}
            <Link to="/proyectos" className="font-display text-lg font-semibold text-ink">
              Software Metrics
            </Link>
          </div>
          <div className="flex flex-wrap items-center gap-3 text-sm text-ink-soft">
            {usuarioActivo && <span>{usuarioActivo.nombre}</span>}
            <ModeSwitch />
            <button
              type="button"
              onClick={manejarSalir}
              className="inline-flex items-center gap-2 rounded-pill border border-border px-4 py-2 text-sm font-medium text-ink transition hover:bg-primary-100"
            >
              <i className="bi bi-box-arrow-right" aria-hidden="true" />
              Salir
            </button>
          </div>
        </div>
      </header>
      <div className="relative flex flex-1">
        {esMovil && !colapsada && (
          <div
            aria-hidden="true"
            onClick={() => setMenuAbierto(false)}
            className="absolute inset-0 z-10 bg-overlay"
          />
        )}
        <aside
          id="barra-lateral"
          hidden={colapsada}
          className={`w-56 shrink-0 border-r border-border bg-surface px-4 py-6 ${
            esMovil ? 'absolute inset-y-0 left-0 z-20 shadow-card' : ''
          }`}
        >
          <nav aria-label="Secciones del panel" className="flex flex-col gap-1">
            <NavLink to="/proyectos" end className={claseEnlaceDeMenu}>
              <i className="bi bi-kanban" aria-hidden="true" />
              Proyectos
            </NavLink>
            <NavLink to="/usuarios" className={claseEnlaceDeMenu}>
              <i className="bi bi-people" aria-hidden="true" />
              Usuarios
            </NavLink>
          </nav>
          <BarraLateralProyecto />
        </aside>
        <main className="min-w-0 flex-1 px-6 py-10">
          <div className="mx-auto max-w-5xl">
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  )
}
