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
        className={`rounded-xl border bg-glass-2 px-3.5 py-2.5 text-sm text-ink backdrop-blur-md transition-all duration-150 placeholder:text-ink-soft focus:border-violet focus:outline-none focus-visible:shadow-[var(--ring)] ${
          error ? 'border-[var(--danger)]' : 'border-input-border'
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
