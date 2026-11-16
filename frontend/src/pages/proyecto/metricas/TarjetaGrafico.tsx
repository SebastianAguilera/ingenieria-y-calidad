import { useId, useState } from 'react'
import type { ReactNode } from 'react'

type TarjetaGraficoProps = {
  titulo: string
  descripcion?: string
  /** Tabla alternativa con los mismos valores; si falta, no hay botón "Ver como tabla". */
  tabla?: { columnas: string[]; filas: string[][] }
  children: ReactNode
  className?: string
}

/**
 * Tarjeta de gráfico con su tabla accesible. La tabla siempre está en el DOM
 * (oculta con `sr-only`) para los lectores de pantalla, y el botón "Ver como
 * tabla" la muestra visualmente (US-F21, CA-8).
 */
export default function TarjetaGrafico({
  titulo,
  descripcion,
  tabla,
  children,
  className = '',
}: TarjetaGraficoProps) {
  const idTitulo = useId()
  const idTabla = useId()
  const [tablaVisible, setTablaVisible] = useState(false)

  return (
    <section
      aria-labelledby={idTitulo}
      className={`flex flex-col rounded-card border border-border bg-surface p-5 shadow-card ${className}`.trim()}
    >
      <header className="mb-4 flex flex-wrap items-start justify-between gap-2">
        <div>
          <h2 id={idTitulo} className="text-lg font-semibold text-ink">
            {titulo}
          </h2>
          {descripcion && <p className="text-xs text-ink-soft">{descripcion}</p>}
        </div>
        {tabla && (
          <button
            type="button"
            aria-expanded={tablaVisible}
            aria-controls={idTabla}
            onClick={() => setTablaVisible((visible) => !visible)}
            className="inline-flex items-center gap-1.5 rounded-pill px-3 py-1 text-xs font-semibold text-primary transition hover:bg-primary-100 focus-visible:shadow-[var(--ring)]"
          >
            <i className={`bi ${tablaVisible ? 'bi-bar-chart' : 'bi-table'}`} aria-hidden="true" />
            {tablaVisible ? 'Ocultar tabla' : 'Ver como tabla'}
          </button>
        )}
      </header>

      {children}

      {tabla && (
        <table
          id={idTabla}
          className={tablaVisible ? 'mt-4 w-full text-left text-sm' : 'sr-only'}
        >
          <caption className="sr-only">{titulo}</caption>
          <thead className="border-b border-border text-xs text-ink-soft">
            <tr>
              {tabla.columnas.map((columna) => (
                <th key={columna} scope="col" className="px-3 py-2 font-semibold">
                  {columna}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {tabla.filas.map(([encabezado, ...valores]) => (
              <tr key={encabezado} className="border-b border-border last:border-b-0">
                <th scope="row" className="px-3 py-2 font-medium text-ink">
                  {encabezado}
                </th>
                {valores.map((valor, indice) => (
                  <td key={indice} className="px-3 py-2 text-ink">
                    {valor}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </section>
  )
}
