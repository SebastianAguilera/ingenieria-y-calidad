import PageHeader from '../../components/PageHeader'
import { formatearFecha } from '../../utils/fechas'
import { useProyecto } from './proyectoContext'

export default function ResumenPage() {
  const { estado } = useProyecto()

  if (estado.tipo !== 'listo') {
    return null
  }

  const proyecto = estado.proyecto
  const filas: Array<[string, string]> = [
    ['Nombre', proyecto.nombre],
    ['Estado', proyecto.estado],
    ['Fecha de inicio', formatearFecha(proyecto.fecha_inicio)],
    ['Fecha de fin', formatearFecha(proyecto.fecha_fin)],
    ['Cantidad de integrantes', String(proyecto.cantidad_usuarios)],
  ]

  return (
    <section>
      <PageHeader title="Resumen" description="Datos principales del proyecto." />
      <div className="overflow-x-auto rounded-tile border border-border bg-surface">
        <table className="w-full min-w-[32rem] text-left text-sm">
          <thead className="border-b border-border text-ink-soft">
            <tr>
              <th scope="col" className="px-6 py-4 font-semibold">
                Campo
              </th>
              <th scope="col" className="px-6 py-4 font-semibold">
                Valor
              </th>
            </tr>
          </thead>
          <tbody>
            {filas.map(([campo, valor]) => (
              <tr key={campo} className="border-b border-border last:border-b-0">
                <th scope="row" className="px-6 py-5 font-semibold text-ink-soft">
                  {campo}
                </th>
                <td className="px-6 py-5 font-medium text-ink">{valor}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  )
}
