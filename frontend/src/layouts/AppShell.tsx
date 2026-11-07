import { Link, Outlet, useNavigate } from 'react-router-dom'
import { useSesion } from '../session/sesion'
import ModeSwitch from '../ui/ModeSwitch'

export default function AppShell() {
  const { usuarioActivo, salir } = useSesion()
  const navegar = useNavigate()

  function manejarSalir() {
    salir()
    navegar('/login', { replace: true })
  }

  return (
    <div className="min-h-screen bg-page text-ink">
      <header className="border-b border-border bg-surface">
        <div className="mx-auto flex max-w-5xl flex-wrap items-center justify-between gap-3 px-6 py-4">
          <Link to="/proyectos" className="font-display text-lg font-semibold text-ink">
            Software Metrics
          </Link>
          <nav className="flex flex-wrap items-center gap-3 text-sm text-ink-soft">
            <Link to="/proyectos" className="hover:text-ink">
              Proyectos
            </Link>
            {usuarioActivo && <span className="hidden sm:inline">{usuarioActivo.nombre}</span>}
            <ModeSwitch />
            <button
              type="button"
              onClick={manejarSalir}
              className="rounded-pill border border-border px-4 py-2 text-sm font-medium text-ink transition hover:bg-[var(--primary-100)]"
            >
              Salir
            </button>
          </nav>
        </div>
      </header>
      <main className="mx-auto max-w-5xl px-6 py-10">
        <Outlet />
      </main>
    </div>
  )
}
