type SelectPillProps<T extends string> = {
  label: string
  value: T
  opciones: readonly T[]
  onChange: (valor: T) => void
  disabled?: boolean
  className?: string
}

export default function SelectPill<T extends string>({
  label,
  value,
  opciones,
  onChange,
  disabled = false,
  className = '',
}: SelectPillProps<T>) {
  return (
    <span className="relative inline-flex items-center">
      <select
        aria-label={label}
        value={value}
        onChange={(evento) => onChange(evento.target.value as T)}
        disabled={disabled}
        className={`appearance-none cursor-pointer rounded-pill border py-1.5 pl-3.5 pr-9 text-xs font-semibold transition focus-visible:shadow-[var(--ring)] disabled:cursor-not-allowed disabled:opacity-60 ${className}`.trim()}
      >
        {opciones.map((opcion) => (
          <option key={opcion} value={opcion}>
            {opcion}
          </option>
        ))}
      </select>
      <i
        className="bi bi-chevron-down pointer-events-none absolute right-3 text-xs opacity-75"
        aria-hidden="true"
      />
    </span>
  )
}
