import { useId } from 'react'
import type { ChangeEvent } from 'react'

const CLASES_CONTROL =
  'rounded-tile border bg-[var(--surface-2)] px-3.5 py-2.5 text-sm text-ink placeholder:text-ink-soft focus:border-primary focus:outline-none'

function claseBorde(error: string | undefined): string {
  return error ? 'border-[var(--danger)]' : 'border-border'
}

export type OpcionSelect = { valor: string; texto: string }

type CampoSelectProps = {
  label: string
  value: string
  onChange: (valor: string) => void
  opciones: readonly OpcionSelect[]
  placeholder?: string
  error?: string
  ayuda?: string
  autoFocus?: boolean
  disabled?: boolean
}

export function CampoSelect({
  label,
  value,
  onChange,
  opciones,
  placeholder,
  error,
  ayuda,
  autoFocus,
  disabled,
}: CampoSelectProps) {
  const id = useId()
  const idError = `${id}-error`
  const idAyuda = `${id}-ayuda`
  const descritoPor = [error ? idError : null, ayuda ? idAyuda : null].filter(Boolean).join(' ')

  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={id} className="text-sm font-medium text-ink">
        {label}
      </label>
      <select
        id={id}
        value={value}
        onChange={(evento: ChangeEvent<HTMLSelectElement>) => onChange(evento.target.value)}
        aria-invalid={error ? true : undefined}
        aria-describedby={descritoPor !== '' ? descritoPor : undefined}
        autoFocus={autoFocus}
        disabled={disabled}
        className={`${CLASES_CONTROL} ${claseBorde(error)}`}
      >
        {placeholder !== undefined && <option value="">{placeholder}</option>}
        {opciones.map((opcion) => (
          <option key={opcion.valor} value={opcion.valor}>
            {opcion.texto}
          </option>
        ))}
      </select>
      {ayuda && (
        <p id={idAyuda} className="text-xs text-ink-soft">
          {ayuda}
        </p>
      )}
      {error && (
        <p id={idError} className="text-xs text-[var(--danger)]">
          {error}
        </p>
      )}
    </div>
  )
}

type CampoAreaTextoProps = {
  label: string
  value: string
  onChange: (valor: string) => void
  error?: string
  autoFocus?: boolean
}

export function CampoAreaTexto({ label, value, onChange, error, autoFocus }: CampoAreaTextoProps) {
  const id = useId()
  const idError = `${id}-error`

  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={id} className="text-sm font-medium text-ink">
        {label}
      </label>
      <textarea
        id={id}
        value={value}
        onChange={(evento: ChangeEvent<HTMLTextAreaElement>) => onChange(evento.target.value)}
        aria-invalid={error ? true : undefined}
        aria-describedby={error ? idError : undefined}
        autoFocus={autoFocus}
        className={`min-h-24 ${CLASES_CONTROL} ${claseBorde(error)}`}
      />
      {error && (
        <p id={idError} className="text-xs text-[var(--danger)]">
          {error}
        </p>
      )}
    </div>
  )
}
