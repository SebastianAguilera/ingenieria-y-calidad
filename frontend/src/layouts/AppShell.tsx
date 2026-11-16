import { useState, useSyncExternalStore } from 'react'
import { Link, NavLink, Outlet, useLocation, useMatch, useNavigate } from 'react-router-dom'
import { useSesion } from '../session/sesion'
import ModeSwitch from '../ui/ModeSwitch'
import BarraLateralProyecto from './BarraLateralProyecto'
import ChipUsuario from './ChipUsuario'
import type { ContextoAppShell } from './contextoAppShell'
import { claseEnlaceDeMenu } from './enlaces'
import MigasDePan from './MigasDePan'
import { useContextoProyecto } from './useContextoProyecto'

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
  const [slotAcciones, setSlotAcciones] = useState<HTMLElement | null>(null)

  if (rutaPrevia !== ubicacion.pathname) {
    setRutaPrevia(ubicacion.pathname)
    setMenuAbierto(false)
  }

  function manejarSalir() {
    salir()
    navegar('/login', { replace: true })
  }

  const coincidenciaProyecto = useMatch({ path: '/proyectos/:id', end: false })
  const idProyecto = Number(coincidenciaProyecto?.params.id)
  const proyectoId = Number.isInteger(idProyecto) && idProyecto > 0 ? idProyecto : null
  const { proyecto, sprintActivo } = useContextoProyecto(proyectoId, ubicacion.pathname)

  const colapsada = esMovil ? !menuAbierto : !menuEscritorioVisible
  const riel = !esMovil && colapsada

  return (
    <div className="flex min-h-screen text-ink">
      <div className="relative flex min-h-screen flex-1">
        <aside
          id="barra-lateral"
          hidden={esMovil && colapsada}
          className={`glass shrink-0 rounded-2xl py-6 transition-[width] ${
            riel ? 'w-16 px-2' : 'w-56 px-4'
          } ${esMovil ? 'absolute inset-y-3 left-3 z-20 bg-[var(--page-end)]' : 'sticky top-4 m-4 mr-0 h-[calc(100vh-2rem)]'}`}
        >
          <div
            className={`flex items-center gap-2 ${riel ? 'justify-center' : 'justify-between pl-3.5'}`}
          >
            <Link
              to="/proyectos"
              className="texto-acento font-display text-2xl font-extrabold"
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
                className="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-full border border-glass-border bg-glass text-ink-soft transition-all duration-150 hover:bg-acento hover:text-white"
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
          <header className="glass sticky top-4 z-10 mx-6 mt-4 rounded-2xl">
            <div className="flex flex-wrap items-center gap-3 px-4 py-3">
              {esMovil && (
                <button
                  type="button"
                  onClick={() => setMenuAbierto((abierto) => !abierto)}
                  aria-expanded={menuAbierto}
                  aria-controls="barra-lateral"
                  aria-label="Menú de secciones"
                  className="inline-flex h-10 w-10 items-center justify-center rounded-full border border-glass-border bg-glass text-ink transition-all duration-150 hover:bg-glass-2"
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
              <MigasDePan
                pathname={ubicacion.pathname}
                proyectoId={proyectoId}
                nombreProyecto={proyecto?.nombre ?? null}
                sprintActivo={sprintActivo}
                compacto={esMovil}
              />
              <div className="ml-auto flex flex-wrap items-center gap-3 text-sm text-ink-soft">
                <div ref={setSlotAcciones} className="flex items-center" />
                <ModeSwitch />
                {usuarioActivo && (
                  <ChipUsuario
                    usuario={usuarioActivo}
                    rol={proyecto?.rol_actual}
                    onSalir={manejarSalir}
                  />
                )}
                <button
                  type="button"
                  onClick={manejarSalir}
                  className="inline-flex items-center gap-2 rounded-full border border-glass-border bg-glass px-4 py-2 text-sm font-medium text-ink transition-all duration-150 hover:bg-glass-2"
                >
                  <i className="bi bi-box-arrow-right" aria-hidden="true" />
                  Salir
                </button>
              </div>
            </div>
          </header>
          <main className="min-w-0 flex-1 px-6 py-8">
            <div className="mx-auto max-w-6xl">
              <Outlet context={{ slotAcciones } satisfies ContextoAppShell} />
            </div>
          </main>
        </div>
      </div>
    </div>
  )
}
