import { createPortal } from 'react-dom'
import { Outlet, useOutletContext, useParams } from 'react-router-dom'
import CampanaPoker from '../pages/proyecto/poker/CampanaPoker'
import { ProveedorSesionesPoker } from '../pages/proyecto/poker/ProveedorSesionesPoker'
import type { ContextoAppShell } from './contextoAppShell'

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
  const contexto = useOutletContext<ContextoAppShell | undefined>()
  const slotAcciones = contexto?.slotAcciones ?? null

  return (
    <ProveedorSesionesPoker key={proyectoId ?? 'sin-proyecto'} proyectoId={proyectoId}>
      {slotAcciones && createPortal(<CampanaPoker />, slotAcciones)}
      <Outlet />
    </ProveedorSesionesPoker>
  )
}
