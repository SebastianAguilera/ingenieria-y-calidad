export type ElementoLeyenda = { etiqueta: string; color: string; valor?: string }

/** Leyenda con cuadrados de color y texto: el color nunca es el único canal. */
export default function Leyenda({
  nombre,
  elementos,
  className = '',
}: {
  nombre: string
  elementos: ElementoLeyenda[]
  className?: string
}) {
  return (
    <ul aria-label={nombre} className={`flex flex-wrap gap-x-4 gap-y-2 text-xs ${className}`.trim()}>
      {elementos.map(({ etiqueta, color, valor }) => (
        <li key={etiqueta} className="inline-flex items-center gap-2 text-ink">
          <span
            aria-hidden="true"
            className="h-3 w-3 shrink-0 rounded-[3px]"
            style={{ background: color }}
          />
          <span>{etiqueta}</span>
          {valor !== undefined && <span className="font-semibold text-ink-soft">{valor}</span>}
        </li>
      ))}
    </ul>
  )
}
