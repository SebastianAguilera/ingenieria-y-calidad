import type { ReactNode } from 'react'
import type { Historia } from '../../../api/types'

export type ColumnaHistoria = {
  clave: string
  encabezado: string
  render: (historia: Historia) => ReactNode
}

type TablaHistoriasProps = {
  historias: Historia[]
  columnas: ColumnaHistoria[]
  /** Contenido de la columna de acciones; si se omite no se muestra la columna. */
  acciones?: (historia: Historia) => ReactNode
}

/**
 * Tabla presentacional de historias, compartida por el Product Backlog y el
 * Sprint Backlog. No conoce el origen de los datos: las columnas y las
 * acciones las define quien la compone.
 */
export default function TablaHistorias({ historias, columnas, acciones }: TablaHistoriasProps) {
  return (
    <div className="overflow-x-auto rounded-tile border border-border bg-surface">
      <table className="w-full min-w-[48rem] text-left text-sm">
        <thead className="border-b border-border text-ink-soft">
          <tr>
            {columnas.map((columna) => (
              <th key={columna.clave} scope="col" className="px-6 py-4 font-semibold">
                {columna.encabezado}
              </th>
            ))}
            {acciones && (
              <th scope="col" className="px-6 py-4 font-semibold">
                Acciones
              </th>
            )}
          </tr>
        </thead>
        <tbody>
          {historias.map((historia) => (
            <tr
              key={historia.id}
              className="border-b border-border last:border-b-0 hover:bg-[var(--primary-100)]"
            >
              {columnas.map((columna) => (
                <td key={columna.clave} className="px-6 py-4">
                  {columna.render(historia)}
                </td>
              ))}
              {acciones && <td className="px-6 py-4">{acciones(historia)}</td>}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
