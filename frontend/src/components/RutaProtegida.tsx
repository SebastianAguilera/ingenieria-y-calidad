import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { useSesion } from '../session/sesion'

export default function RutaProtegida() {
  const { usuarioActivo } = useSesion()
  const ubicacion = useLocation()

  if (!usuarioActivo) {
    return <Navigate to="/login" replace state={{ desde: ubicacion.pathname }} />
  }

  return <Outlet />
}
