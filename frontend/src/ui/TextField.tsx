import { useId } from 'react'
import type { InputHTMLAttributes } from 'react'

type TextFieldProps = InputHTMLAttributes<HTMLInputElement> & {
  label: string
  error?: string
  /** Texto de ayuda permanente debajo del campo, enlazado con `aria-describedby`. */
  ayuda?: string
}

export default function TextField({
  label,
  error,
  ayuda,
  id,
  className = '',
  ...props
}: TextFieldProps) {
  const idGenerado = useId()
  const inputId = id ?? idGenerado
  const errorId = `${inputId}-error`
  const ayudaId = `${inputId}-ayuda`
  const descripciones = [error ? errorId : null, ayuda ? ayudaId : null].filter(Boolean).join(' ')

  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={inputId} className="text-sm font-medium text-ink">
        {label}
      </label>
      <input
        id={inputId}
        className={`rounded-tile border bg-[var(--surface-2)] px-3.5 py-2.5 text-sm text-ink placeholder:text-ink-soft focus:border-primary focus:outline-none ${
          error ? 'border-[var(--danger)]' : 'border-border'
        } ${className}`.trim()}
        aria-invalid={error ? true : undefined}
        aria-describedby={descripciones || undefined}
        {...props}
      />
      {error && (
        <p id={errorId} className="text-xs text-[var(--danger)]">
          {error}
        </p>
      )}
      {ayuda && (
        <p id={ayudaId} className="text-xs text-ink-soft">
          {ayuda}
        </p>
      )}
    </div>
  )
}
