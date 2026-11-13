import { useEffect } from 'react'
import Button from '../../../ui/Button'

type ModalConfirmacionProps = {
  titulo: string
  mensaje: string
  textoConfirmar: string
  onConfirmar: () => void
  onCerrar: () => void
  deshabilitado?: boolean
}

export default function ModalConfirmacion({
  titulo,
  mensaje,
  textoConfirmar,
  onConfirmar,
  onCerrar,
  deshabilitado = false,
}: ModalConfirmacionProps) {
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
      className="fixed inset-0 z-50 flex items-center justify-center bg-[color-mix(in_oklab,var(--ink)_45%,transparent)] p-4"
      onMouseDown={(evento) => {
        if (evento.target === evento.currentTarget) {
          onCerrar()
        }
      }}
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="modal-confirmacion-titulo"
        className="w-full max-w-md rounded-card border border-border bg-surface p-6 shadow-card"
      >
        <h2 id="modal-confirmacion-titulo" className="text-2xl font-semibold text-ink">
          {titulo}
        </h2>
        <p className="mt-2 text-sm text-ink-soft">{mensaje}</p>

        <div className="mt-6 flex flex-wrap justify-end gap-3">
          <Button variante="ghost" onClick={onCerrar} disabled={deshabilitado}>
            Cancelar
          </Button>
          <Button onClick={onConfirmar} disabled={deshabilitado}>
            {textoConfirmar}
          </Button>
        </div>
      </div>
    </div>
  )
}
