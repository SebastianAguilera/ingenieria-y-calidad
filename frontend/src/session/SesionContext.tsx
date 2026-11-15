import { useCallback, useEffect, useMemo, useState } from 'react'
import type { ReactNode } from 'react'
import type { Usuario } from '../api/types'
import {
  EVENTO_SESION_EXPIRADA,
  SesionContext,
  guardarSesion,
  leerSesionGuardada,
  limpiarSesion,
} from './sesion'
import type { SesionGuardada } from './sesion'

export function SesionProvider({ children }: { children: ReactNode }) {
  const [sesion, setSesion] = useState<SesionGuardada | null>(() => leerSesionGuardada())

  useEffect(() => {
    const expirar = () => setSesion(null)
    window.addEventListener(EVENTO_SESION_EXPIRADA, expirar)
    return () => window.removeEventListener(EVENTO_SESION_EXPIRADA, expirar)
  }, [])

  const ingresar = useCallback((token: string, usuario: Usuario) => {
    const nueva: SesionGuardada = { token, usuario }
    setSesion(nueva)
    guardarSesion(nueva)
  }, [])

  const salir = useCallback(() => {
    setSesion(null)
    limpiarSesion()
  }, [])

  const valor = useMemo(
    () => ({
      usuarioActivo: sesion?.usuario ?? null,
      token: sesion?.token ?? null,
      ingresar,
      salir,
    }),
    [sesion, ingresar, salir],
  )

  return <SesionContext.Provider value={valor}>{children}</SesionContext.Provider>
}
