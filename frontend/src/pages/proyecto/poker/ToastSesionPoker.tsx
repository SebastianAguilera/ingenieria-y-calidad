import type { SesionPokerActiva } from '../../../api/types'

export default function ToastSesionPoker({ sesion }: { sesion: SesionPokerActiva }) {
  return (
    <div
      role="status"
      aria-live="polite"
      aria-atomic="true"
      className="flotante fixed bottom-6 right-6 z-50 max-w-sm p-4"
    >
      <div className="flex items-start gap-3">
        <span
          aria-hidden="true"
          className="flex h-9 w-9 shrink-0 items-center justify-center rounded-tile bg-[var(--accent-100)] text-[var(--accent-600)]"
        >
          <i className="bi bi-bell-fill" />
        </span>
        <p className="text-sm font-medium text-ink">
          Se abrió Planning Poker para {sesion.historia.codigo}: {sesion.historia.titulo}
        </p>
      </div>
    </div>
  )
}
