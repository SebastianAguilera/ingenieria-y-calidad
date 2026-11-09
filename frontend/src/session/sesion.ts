import { createContext, useContext } from 'react'
import type { Usuario } from '../api/types'

export const CLAVE_SESION = 'app.sesion.v1'

export type SesionContextValue = {
  usuarioActivo: Usuario | null
  ingresar: (usuario: Usuario) => void
  salir: () => void
}

export const SesionContext = createContext<SesionContextValue | null>(null)

function esUsuario(valor: unknown): valor is Usuario {
  if (typeof valor !== 'object' || valor === null) {
    return false
  }

  const candidato = valor as Record<string, unknown>
  return (
    typeof candidato.id === 'number' &&
    typeof candidato.nombre === 'string' &&
    typeof candidato.email === 'string' &&
    typeof candidato.creado_en === 'string'
  )
}

export function leerSesionGuardada(): Usuario | null {
  try {
    const crudo = localStorage.getItem(CLAVE_SESION)
    if (!crudo) {
      return null
    }

    const parseado: unknown = JSON.parse(crudo)
    return esUsuario(parseado) ? parseado : null
  } catch {
    return null
  }
}

export function useSesion(): SesionContextValue {
  const contexto = useContext(SesionContext)
  if (!contexto) {
    throw new Error('useSesion debe usarse dentro de SesionProvider')
  }
  return contexto
}
