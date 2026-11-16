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
    <div className="tabla-contenedor">
      <table
        className={`tabla min-w-[48rem] ${layoutFijo ? 'table-fixed' : ''}`.trim()}
      >
        <thead>
          <tr>
            {columnas.map((columna) => (
              <th
                key={columna.clave}
                scope="col"
                className={`${columna.ancho ?? ''} ${ALINEACION[columna.alineacion ?? 'izquierda']} ${layoutFijo ? '!whitespace-normal' : ''}`.trim()}
              >
                {columna.encabezado}
              </th>
            ))}
            {acciones && (
              <th
                scope="col"
                className={`col-acciones ${anchoAcciones ?? ''}`.trim()}
              >
                Acciones
              </th>
            )}
          </tr>
        </thead>
        <tbody>
          {historias.map((historia) => (
            <tr key={historia.id}>
              {columnas.map((columna) => (
                <td
                  key={columna.clave}
                  className={ALINEACION[columna.alineacion ?? 'izquierda']}
                >
                  {columna.render(historia)}
                </td>
              ))}
              {acciones && <td className="col-acciones">{acciones(historia)}</td>}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
