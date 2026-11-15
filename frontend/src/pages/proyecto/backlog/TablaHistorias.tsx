import type { ReactNode } from 'react'
import type { Historia } from '../../../api/types'

export type AlineacionColumna = 'izquierda' | 'centro' | 'derecha'

export type ColumnaHistoria = {
  clave: string
  encabezado: string
  /** Ancho de la columna en clases Tailwind (ej. 'w-24'); requiere `layoutFijo`. */
  ancho?: string
  alineacion?: AlineacionColumna
  render: (historia: Historia) => ReactNode
}

type TablaHistoriasProps = {
  historias: Historia[]
  columnas: ColumnaHistoria[]
  /** Contenido de la columna de acciones; si se omite no se muestra la columna. */
  acciones?: (historia: Historia) => ReactNode
  anchoAcciones?: string
  /** Distribuye los anchos de columna de forma proporcional (`table-fixed`). */
  layoutFijo?: boolean
}

const ALINEACION: Record<AlineacionColumna, string> = {
  izquierda: 'text-left',
  centro: 'text-center',
  derecha: 'text-right',
}

const CELDA = 'px-4 py-3.5 align-middle'

/**
 * Tabla presentacional de historias, compartida por el Product Backlog y el
 * Sprint Backlog. No conoce el origen de los datos: las columnas y las
 * acciones las define quien la compone.
 */
export default function TablaHistorias({
  historias,
  columnas,
  acciones,
  anchoAcciones,
  layoutFijo = false,
}: TablaHistoriasProps) {
  return (
    <div className="overflow-x-auto rounded-tile border border-border bg-surface">
      <table
        className={`w-full min-w-[48rem] text-left text-sm ${layoutFijo ? 'table-fixed' : ''}`.trim()}
      >
        <thead className="border-b border-border text-ink-soft">
          <tr>
            {columnas.map((columna) => (
              <th
                key={columna.clave}
                scope="col"
                className={`${CELDA} font-semibold ${columna.ancho ?? ''} ${ALINEACION[columna.alineacion ?? 'izquierda']}`.trim()}
              >
                {columna.encabezado}
              </th>
            ))}
            {acciones && (
              <th
                scope="col"
                className={`${CELDA} text-right font-semibold ${anchoAcciones ?? ''}`.trim()}
              >
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
                <td
                  key={columna.clave}
                  className={`${CELDA} ${ALINEACION[columna.alineacion ?? 'izquierda']}`.trim()}
                >
                  {columna.render(historia)}
                </td>
              ))}
              {acciones && <td className={`${CELDA} text-right`}>{acciones(historia)}</td>}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
