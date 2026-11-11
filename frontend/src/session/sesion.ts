import { createContext, useContext } from 'react'
import type { Usuario } from '../api/types'

export const CLAVE_SESION = 'app.sesion.v1'

export const EVENTO_SESION_EXPIRADA = 'sesion:expirada'

export type SesionGuardada = {
  token: string
  usuario: Usuario
}

export type SesionContextValue = {
  usuarioActivo: Usuario | null
  token: string | null
  ingresar: (token: string, usuario: Usuario) => void
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

export function esSesionGuardada(valor: unknown): valor is SesionGuardada {
  if (typeof valor !== 'object' || valor === null) {
    return false
  }

  const candidato = valor as Record<string, unknown>
  return (
    typeof candidato.token === 'string' &&
    candidato.token.trim() !== '' &&
    esUsuario(candidato.usuario)
  )
}

export function leerSesionGuardada(): SesionGuardada | null {
  try {
    const crudo = localStorage.getItem(CLAVE_SESION)
    if (!crudo) {
      return null
    }

    const parseado: unknown = JSON.parse(crudo)
    return esSesionGuardada(parseado) ? parseado : null
  } catch {
    return null
  }
}

export function leerToken(): string | null {
  return leerSesionGuardada()?.token ?? null
}

export function guardarSesion(sesion: SesionGuardada): void {
  try {
    localStorage.setItem(CLAVE_SESION, JSON.stringify(sesion))
  } catch {
    // El almacenamiento puede no estar disponible; la sesión sigue en memoria.
  }
}

export function limpiarSesion(): void {
  try {
    localStorage.removeItem(CLAVE_SESION)
  } catch {
    // Sin almacenamiento disponible no hay nada que limpiar.
  }
  window.dispatchEvent(new Event(EVENTO_SESION_EXPIRADA))
}

export function useSesion(): SesionContextValue {
  const contexto = useContext(SesionContext)
  if (!contexto) {
    throw new Error('useSesion debe usarse dentro de SesionProvider')
  }
  return contexto
}
