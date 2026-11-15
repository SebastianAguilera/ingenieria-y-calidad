import { useTema } from '../theme/tema'

function IconoSol() {
  return (
    <svg aria-hidden="true" viewBox="0 0 24 24" className="h-5 w-5" fill="none" stroke="currentColor" strokeWidth="1.8">
      <circle cx="12" cy="12" r="4" />
      <path d="M12 2v2M12 20v2M2 12h2M20 12h2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M19.1 4.9l-1.4 1.4M6.3 17.7l-1.4 1.4" />
    </svg>
  )
}

function IconoLuna() {
  return (
    <svg aria-hidden="true" viewBox="0 0 24 24" className="h-5 w-5" fill="none" stroke="currentColor" strokeWidth="1.8">
      <path d="M21 12.8A8.5 8.5 0 1 1 11.2 3a6.5 6.5 0 0 0 9.8 9.8Z" />
    </svg>
  )
}

export default function ModeSwitch() {
  const { esOscuro, alternarModo } = useTema()

  return (
    <button
      type="button"
      onClick={alternarModo}
      aria-label={esOscuro ? 'Cambiar a modo claro' : 'Cambiar a modo oscuro'}
      className="inline-flex h-10 w-10 items-center justify-center rounded-pill border border-border text-ink transition hover:bg-[var(--primary-100)]"
    >
      {esOscuro ? <IconoSol /> : <IconoLuna />}
    </button>
  )
}
