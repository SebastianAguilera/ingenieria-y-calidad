import { useEffect, useId, useRef } from 'react'
import type { ReactNode } from 'react'
import Button from './Button'

type ModalConfirmacionProps = {
  titulo: string
  children: ReactNode
  textoConfirmar: string
  textoCancelar?: string
  procesando?: boolean
  onConfirmar: () => void
  onCancelar: () => void
}

export default function ModalConfirmacion({
  titulo,
  children,
  textoConfirmar,
  textoCancelar = 'Cancelar',
  procesando = false,
  onConfirmar,
  onCancelar,
}: ModalConfirmacionProps) {
  const tituloId = useId()
  const botonConfirmarRef = useRef<HTMLButtonElement>(null)

  useEffect(() => {
    const elementoPrevio = document.activeElement as HTMLElement | null
    botonConfirmarRef.current?.focus()

    function alPresionarTecla(evento: KeyboardEvent) {
      if (evento.key === 'Escape' && !procesando) {
        onCancelar()
      }
    }

    document.addEventListener('keydown', alPresionarTecla)
    return () => {
      document.removeEventListener('keydown', alPresionarTecla)
      elementoPrevio?.focus()
    }
  }, [onCancelar, procesando])

  return (
    <div
      data-testid="fondo-modal"
      className="fixed inset-0 z-50 flex items-center justify-center bg-[color-mix(in_oklab,var(--ink)_45%,transparent)] p-4"
      onMouseDown={(evento) => {
        if (evento.target === evento.currentTarget && !procesando) {
          onCancelar()
        }
      }}
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby={tituloId}
        className="w-full max-w-lg rounded-card border border-border bg-surface p-6 shadow-card"
      >
        <h2 id={tituloId} className="text-2xl font-semibold text-ink">
          {titulo}
        </h2>
        <div className="mt-4 flex flex-col gap-4 text-sm text-ink-soft">{children}</div>
        <div className="mt-6 flex flex-wrap justify-end gap-3">
          <Button variante="ghost" onClick={onCancelar} disabled={procesando}>
            {textoCancelar}
          </Button>
          <Button ref={botonConfirmarRef} onClick={onConfirmar} disabled={procesando}>
            {procesando ? 'Eliminando…' : textoConfirmar}
          </Button>
        </div>
      </div>
    </div>
  )
}
