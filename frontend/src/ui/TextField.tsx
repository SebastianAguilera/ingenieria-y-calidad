import { useId } from 'react'
import type { InputHTMLAttributes } from 'react'

type TextFieldProps = InputHTMLAttributes<HTMLInputElement> & {
  label: string
  error?: string
}

export default function TextField({
  label,
  error,
  id,
  className = '',
  ...props
}: TextFieldProps) {
  const idGenerado = useId()
  const inputId = id ?? idGenerado
  const errorId = `${inputId}-error`

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
        aria-describedby={error ? errorId : undefined}
        {...props}
      />
      {error && (
        <p id={errorId} className="text-xs text-[var(--danger)]">
          {error}
        </p>
      )}
    </div>
  )
}
