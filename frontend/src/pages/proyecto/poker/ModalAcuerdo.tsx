import { useEffect, useId, useRef, useState } from 'react'
import Alert from '../../../ui/Alert'
import Button from '../../../ui/Button'
import { ESCALA_POKER } from './escala'

type ModalAcuerdoProps = {
  valorInicial: number | null
  procesando?: boolean
  error?: string | null
  onConfirmar: (puntos: number) => void
  onCerrar: () => void
}

export default function ModalAcuerdo({
  valorInicial,
  procesando = false,
  error = null,
  onConfirmar,
  onCerrar,
}: ModalAcuerdoProps) {
  const tituloId = useId()
  const grupoRef = useRef<HTMLFieldSetElement>(null)
  const [seleccion, setSeleccion] = useState<number | null>(valorInicial)
  const [errorLocal, setErrorLocal] = useState<string | null>(null)

  useEffect(() => {
    const elementoPrevio = document.activeElement as HTMLElement | null
    const grupo = grupoRef.current
    const aEnfocar =
      grupo?.querySelector<HTMLInputElement>('input:checked') ??
      grupo?.querySelector<HTMLInputElement>('input')
    aEnfocar?.focus()

    function alPresionarTecla(evento: KeyboardEvent) {
      if (evento.key === 'Escape' && !procesando) {
        onCerrar()
      }
    }

    document.addEventListener('keydown', alPresionarTecla)
    return () => {
      document.removeEventListener('keydown', alPresionarTecla)
      elementoPrevio?.focus()
    }
  }, [onCerrar, procesando])

  function confirmar() {
    if (seleccion === null) {
      setErrorLocal('Elegí una estimación de la escala.')
      return
    }
    setErrorLocal(null)
    onConfirmar(seleccion)
  }

  return (
    <div
      data-testid="fondo-modal"
      className="fixed inset-0 z-50 flex items-center justify-center overflow-y-auto bg-[color-mix(in_oklab,var(--ink)_45%,transparent)] p-4"
      onMouseDown={(evento) => {
        if (evento.target === evento.currentTarget && !procesando) {
          onCerrar()
        }
      }}
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby={tituloId}
        className="w-full max-w-xl rounded-card border border-border bg-surface p-6 shadow-card"
      >
        <header className="flex items-start gap-4">
          <span
            aria-hidden="true"
            className="flex h-11 w-11 shrink-0 items-center justify-center rounded-tile bg-[var(--primary-100)] text-xl text-primary"
          >
            <i className="bi bi-hand-thumbs-up" />
          </span>
          <div>
            <h2 id={tituloId} className="text-2xl font-semibold text-ink">
              Registrar acuerdo
            </h2>
            <p className="mt-1 text-sm text-ink-soft">
              Elegí los puntos acordados. Al confirmar, la sesión se cierra y la historia queda
              estimada.
            </p>
          </div>
        </header>

        {error && (
          <div className="mt-4">
            <Alert tono="error">{error}</Alert>
          </div>
        )}

        {errorLocal && (
          <div className="mt-4">
            <Alert tono="error">{errorLocal}</Alert>
          </div>
        )}

        <fieldset ref={grupoRef} className="mt-6" disabled={procesando}>
          <legend className="mb-3 text-sm font-semibold text-ink">Puntos acordados</legend>
          <div className="grid grid-cols-5 gap-2 sm:grid-cols-9">
            {ESCALA_POKER.map((puntos) => (
              <label key={puntos} className="relative cursor-pointer">
                <input
                  type="radio"
                  name="puntos-acuerdo"
                  value={String(puntos)}
                  checked={seleccion === puntos}
                  onChange={() => {
                    setSeleccion(puntos)
                    setErrorLocal(null)
                  }}
                  disabled={procesando}
                  aria-label={`${puntos} puntos`}
                  className="peer sr-only"
                />
                <span className="flex aspect-[3/4] w-full items-center justify-center rounded-tile border-2 border-border bg-surface text-xl font-semibold text-ink shadow-card transition hover:-translate-y-0.5 hover:border-primary peer-checked:border-primary peer-checked:bg-primary peer-checked:text-[var(--on-primary)] peer-focus-visible:shadow-[var(--ring)] peer-disabled:cursor-not-allowed peer-disabled:opacity-60">
                  {puntos}
                </span>
                {valorInicial !== null && puntos === valorInicial && (
                  <span
                    aria-hidden="true"
                    className="absolute -top-2 left-1/2 -translate-x-1/2 whitespace-nowrap rounded-pill bg-[var(--accent-500)] px-2 py-0.5 text-[10px] font-semibold text-[var(--on-accent)]"
                  >
                    Sugerido
                  </span>
                )}
              </label>
            ))}
          </div>
        </fieldset>

        <div className="mt-6 flex flex-wrap items-center justify-between gap-3 border-t border-border pt-5">
          <p className="text-sm text-ink-soft" aria-live="polite">
            {seleccion === null
              ? 'Ninguna estimación elegida'
              : `Seleccionado: ${seleccion} ${seleccion === 1 ? 'punto' : 'puntos'}`}
          </p>
          <div className="flex flex-wrap gap-3">
            <Button variante="ghost" onClick={onCerrar} disabled={procesando}>
              Cancelar
            </Button>
            <Button onClick={confirmar} disabled={procesando} aria-busy={procesando}>
              {procesando ? 'Registrando…' : 'Confirmar acuerdo'}
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
