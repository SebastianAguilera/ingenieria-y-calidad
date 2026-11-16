import { useEffect, useId } from 'react'
import type { ReactNode } from 'react'

type MarcoModalProps = {
  titulo: string
  subtitulo?: string
  onCerrar: () => void
  children: ReactNode
}

export default function MarcoModal({ titulo, subtitulo, onCerrar, children }: MarcoModalProps) {
  const idTitulo = useId()

  useEffect(() => {
    function alPresionarTecla(evento: KeyboardEvent) {
      if (evento.key === 'Escape') {
        onCerrar()
      }
    }

    document.addEventListener('keydown', alPresionarTecla)
    return () => document.removeEventListener('keydown', alPresionarTecla)
  }, [onCerrar])

  return (
    <div
      data-testid="fondo-modal"
      className="modal-fondo"
      onMouseDown={(evento) => {
        if (evento.target === evento.currentTarget) {
          onCerrar()
        }
      }}
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby={idTitulo}
        className="w-full max-w-lg modal-panel p-6"
      >
        <h2 id={idTitulo} className="text-2xl font-semibold text-ink">
          {titulo}
        </h2>
        {subtitulo && <p className="mt-1 text-sm text-ink-soft">{subtitulo}</p>}
        {children}
      </div>
    </div>
  )
}
