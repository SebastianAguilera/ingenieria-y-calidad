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
    <select
      aria-label={label}
      value={value}
      onChange={(evento) => onChange(evento.target.value as T)}
      disabled={disabled}
      className={`cursor-pointer rounded-pill border px-3 py-1 text-xs font-semibold transition focus-visible:shadow-[var(--ring)] disabled:cursor-not-allowed disabled:opacity-60 ${className}`.trim()}
    >
      {opciones.map((opcion) => (
        <option key={opcion} value={opcion}>
          {opcion}
        </option>
      ))}
    </select>
  )
}
