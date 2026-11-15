import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { useSesion } from '../session/sesion'

export default function RutaProtegida() {
  const { token } = useSesion()
  const ubicacion = useLocation()

  if (!token) {
    return <Navigate to="/login" replace state={{ desde: ubicacion.pathname }} />
  }

  return <Outlet />
}
