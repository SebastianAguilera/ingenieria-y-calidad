import { useEffect, useRef, useState } from 'react'
import type { Historia } from '../../../api/types'
import Alert from '../../../ui/Alert'
import Button from '../../../ui/Button'

type ModalAsignarHistoriaProps = {
  historias: Historia[]
  onAsignar: (historiaId: number) => void
  onCerrar: () => void
  error?: string | null
  enviando?: boolean
  cargando?: boolean
}

export default function ModalAsignarHistoria({
  historias,
  onAsignar,
  onCerrar,
  error = null,
  enviando = false,
  cargando = false,
}: ModalAsignarHistoriaProps) {
  const [seleccion, setSeleccion] = useState<number | null>(null)
  const dialogoRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    dialogoRef.current?.focus()
  }, [])

  useEffect(() => {
    function alPresionarTecla(evento: KeyboardEvent) {
      if (evento.key === 'Escape' && !enviando) {
        onCerrar()
      }
    }

    document.addEventListener('keydown', alPresionarTecla)
    return () => document.removeEventListener('keydown', alPresionarTecla)
  }, [enviando, onCerrar])

  const sinHistorias = !cargando && historias.length === 0

  return (
    <div
      data-testid="fondo-modal"
      className="fixed inset-0 z-50 flex items-center justify-center bg-[color-mix(in_oklab,var(--ink)_45%,transparent)] p-4"
      onMouseDown={(evento) => {
        if (evento.target === evento.currentTarget && !enviando) {
          onCerrar()
        }
      }}
    >
      <div
        ref={dialogoRef}
        tabIndex={-1}
        role="dialog"
        aria-modal="true"
        aria-labelledby="modal-asignar-titulo"
        className="w-full max-w-lg rounded-card border border-border bg-surface p-6 shadow-card focus:outline-none"
      >
        <h2 id="modal-asignar-titulo" className="text-2xl font-semibold text-ink">
          Asignar historia
        </h2>
        <p className="mt-2 text-sm text-ink-soft">
          Elegí una historia del Product Backlog para sumarla al sprint.
        </p>

        <div className="mt-5 flex flex-col gap-4">
          {error && <Alert tono="error">{error}</Alert>}

          {cargando && (
            <p role="status" className="text-sm text-ink-soft">
              Cargando historias…
            </p>
          )}

          {sinHistorias && (
            <p className="text-sm text-ink-soft">No hay historias disponibles para asignar</p>
          )}

          {!cargando && historias.length > 0 && (
            <fieldset className="flex max-h-72 flex-col gap-2 overflow-y-auto">
              <legend className="mb-2 text-sm font-semibold text-ink">
                Historias disponibles
              </legend>
              {historias.map((historia) => (
                <label
                  key={historia.id}
                  className="flex cursor-pointer items-center gap-3 rounded-tile border border-border px-4 py-3 text-sm text-ink hover:bg-[var(--primary-100)]"
                >
                  <input
                    type="radio"
                    name="historia-a-asignar"
                    value={historia.id}
                    checked={seleccion === historia.id}
                    disabled={enviando}
                    onChange={() => setSeleccion(historia.id)}
                  />
                  <span>
                    <span className="font-mono text-xs font-semibold text-ink-soft">
                      {historia.codigo}
                    </span>{' '}
                    {historia.titulo}
                  </span>
                </label>
              ))}
            </fieldset>
          )}
        </div>

        <div className="mt-6 flex flex-wrap justify-end gap-3">
          <Button variante="ghost" onClick={onCerrar} disabled={enviando}>
            Cancelar
          </Button>
          <Button
            onClick={() => seleccion !== null && onAsignar(seleccion)}
            disabled={seleccion === null || enviando || sinHistorias || cargando}
          >
            Asignar
          </Button>
        </div>
      </div>
    </div>
  )
}
