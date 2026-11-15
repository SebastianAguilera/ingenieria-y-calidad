import { useId } from 'react'
import { ESCALA_POKER } from './escala'

type CartasVotoProps = {
  valor: number | null
  onSeleccionar: (puntos: number) => void
  deshabilitado?: boolean
}

export default function CartasVoto({
  valor,
  onSeleccionar,
  deshabilitado = false,
}: CartasVotoProps) {
  const nombreGrupo = useId()

  return (
    <fieldset disabled={deshabilitado}>
      <legend className="mb-3 text-sm font-semibold text-ink">Tu estimación</legend>
      <div className="flex flex-wrap gap-2">
        {ESCALA_POKER.map((puntos) => (
          <label key={puntos} className="relative cursor-pointer">
            <input
              type="radio"
              name={nombreGrupo}
              value={String(puntos)}
              checked={valor === puntos}
              onChange={() => onSeleccionar(puntos)}
              disabled={deshabilitado}
              aria-label={`${puntos} puntos`}
              className="peer sr-only"
            />
            <span className="flex h-16 min-w-16 items-center justify-center rounded-tile border-2 border-border bg-surface px-4 text-lg font-semibold text-ink transition hover:border-primary peer-checked:border-primary peer-checked:bg-primary peer-checked:text-[var(--on-primary)] peer-focus-visible:shadow-[var(--ring)] peer-disabled:cursor-not-allowed peer-disabled:opacity-60">
              {puntos}
            </span>
          </label>
        ))}
      </div>
    </fieldset>
  )
}
