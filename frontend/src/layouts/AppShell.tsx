import { useState, useSyncExternalStore } from 'react'
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
  const [menuEscritorioVisible, setMenuEscritorioVisible] = useState(true)
  const [rutaPrevia, setRutaPrevia] = useState(ubicacion.pathname)

  if (rutaPrevia !== ubicacion.pathname) {
    setRutaPrevia(ubicacion.pathname)
    setMenuAbierto(false)
  }

  function manejarSalir() {
    salir()
    navegar('/login', { replace: true })
  }

  const colapsada = esMovil ? !menuAbierto : !menuEscritorioVisible
  const riel = !esMovil && colapsada

  return (
    <div className="flex min-h-screen bg-page text-ink">
      <div className="relative flex min-h-screen flex-1">
        <aside
          id="barra-lateral"
          hidden={esMovil && colapsada}
          className={`shrink-0 border-r border-border bg-surface py-6 transition-[width] ${
            riel ? 'w-16 px-2' : 'w-56 px-4'
          } ${esMovil ? 'absolute inset-y-0 left-0 z-20 shadow-card' : 'sticky top-0 h-screen'}`}
        >
          <div
            className={`flex items-center gap-2 ${riel ? 'justify-center' : 'justify-between'}`}
          >
            <Link
              to="/proyectos"
              className="font-display text-xl font-semibold text-ink"
              title="Software Metrics"
            >
              <span className={riel ? 'sr-only' : ''}>Software Metrics</span>
            </Link>
            {!esMovil && (
              <button
                type="button"
                onClick={() => setMenuEscritorioVisible((visible) => !visible)}
                aria-expanded={menuEscritorioVisible}
                aria-controls="barra-lateral"
                aria-label={
                  menuEscritorioVisible ? 'Contraer barra lateral' : 'Expandir barra lateral'
                }
                className="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-full border border-border bg-bg text-ink-soft shadow-card transition hover:border-primary hover:bg-primary hover:text-[var(--on-primary)]"
              >
                <i
                  className={`bi ${menuEscritorioVisible ? 'bi-chevron-double-left' : 'bi-chevron-double-right'} text-sm`}
                  aria-hidden="true"
                />
              </button>
            )}
          </div>
          <nav aria-label="Secciones del panel" className="mt-10 flex flex-col gap-1">
            <NavLink
              to="/proyectos"
              end
              className={(props) => claseEnlaceDeMenu(props, riel)}
              title={riel ? 'Proyectos' : undefined}
            >
              <i className="bi bi-kanban" aria-hidden="true" />
              <span className={riel ? 'sr-only' : ''}>Proyectos</span>
            </NavLink>
            <NavLink
              to="/usuarios"
              className={(props) => claseEnlaceDeMenu(props, riel)}
              title={riel ? 'Usuarios' : undefined}
            >
              <i className="bi bi-people" aria-hidden="true" />
              <span className={riel ? 'sr-only' : ''}>Usuarios</span>
            </NavLink>
          </nav>
          <BarraLateralProyecto compacto={riel} />
        </aside>
        <div className="flex min-w-0 flex-1 flex-col">
          <header className="border-b border-border bg-surface">
            <div className="flex flex-wrap items-center justify-end gap-3 px-6 py-4">
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
              <div className="ml-auto flex flex-wrap items-center gap-3 text-sm text-ink-soft">
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
          <main className="min-w-0 flex-1 px-6 py-10">
            <div className="mx-auto max-w-5xl">
              <Outlet />
            </div>
          </main>
        </div>
      </div>
    </div>
  )
}
