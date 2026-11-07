import { useCallback, useMemo, useState } from 'react'
import type { ReactNode } from 'react'
import type { Usuario } from '../api/types'
import { CLAVE_SESION, SesionContext, leerSesionGuardada } from './sesion'

export function SesionProvider({ children }: { children: ReactNode }) {
  const [usuarioActivo, setUsuarioActivo] = useState<Usuario | null>(() => leerSesionGuardada())

  const ingresar = useCallback((usuario: Usuario) => {
    setUsuarioActivo(usuario)
    try {
      localStorage.setItem(CLAVE_SESION, JSON.stringify(usuario))
    } catch {
      // El almacenamiento puede no estar disponible; la sesión sigue en memoria.
    }
  }, [])

  const salir = useCallback(() => {
    setUsuarioActivo(null)
    try {
      localStorage.removeItem(CLAVE_SESION)
    } catch {
      // Sin almacenamiento disponible no hay nada que limpiar.
    }
  }, [])

  const valor = useMemo(
    () => ({ usuarioActivo, ingresar, salir }),
    [usuarioActivo, ingresar, salir],
  )

  return <SesionContext.Provider value={valor}>{children}</SesionContext.Provider>
}
