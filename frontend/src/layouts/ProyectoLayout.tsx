import { Outlet, useParams } from 'react-router-dom'
import CampanaPoker from '../pages/proyecto/poker/CampanaPoker'
import { ProveedorSesionesPoker } from '../pages/proyecto/poker/ProveedorSesionesPoker'

function parsearProyectoId(id: string | undefined): number | null {
  if (id === undefined) {
    return null
  }
  const numero = Number(id)
  return Number.isInteger(numero) && numero > 0 ? numero : null
}

export default function ProyectoLayout() {
  const { id } = useParams()
  const proyectoId = parsearProyectoId(id)

  return (
    <ProveedorSesionesPoker key={proyectoId ?? 'sin-proyecto'} proyectoId={proyectoId}>
      <div className="mb-4 flex justify-end">
        <CampanaPoker />
      </div>
      <Outlet />
    </ProveedorSesionesPoker>
  )
}
