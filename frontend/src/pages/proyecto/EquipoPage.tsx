import PageHeader from '../../components/PageHeader'
import { formatearFecha } from '../../utils/fechas'
import { useProyecto } from './proyectoContext'

export default function EquipoPage() {
  const { estado } = useProyecto()

  if (estado.tipo !== 'listo') {
    return null
  }

  const usuarios = estado.proyecto.usuarios ?? []

  return (
    <section>
      <PageHeader
        title="Equipo"
        description="Integrantes del proyecto, con la fecha de alta de cada uno en el sistema."
      />
      {usuarios.length === 0 ? (
        <p className="text-sm text-ink-soft">Este proyecto todavía no tiene integrantes</p>
      ) : (
        <div className="overflow-x-auto rounded-tile border border-border bg-surface">
          <table className="w-full min-w-[32rem] text-left text-sm">
            <thead className="border-b border-border text-ink-soft">
              <tr>
                <th scope="col" className="px-6 py-4 font-semibold">
                  Nombre
                </th>
                <th scope="col" className="px-6 py-4 font-semibold">
                  Email
                </th>
                <th scope="col" className="px-6 py-4 font-semibold">
                  Registrado el
                </th>
              </tr>
            </thead>
            <tbody>
              {usuarios.map((integrante) => (
                <tr key={integrante.id} className="border-b border-border last:border-b-0">
                  <td className="px-6 py-5 font-medium text-ink">{integrante.nombre}</td>
                  <td className="px-6 py-5">{integrante.email}</td>
                  <td className="px-6 py-5">{formatearFecha(integrante.creado_en ?? null)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  )
}
