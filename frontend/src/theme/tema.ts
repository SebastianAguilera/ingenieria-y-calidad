import { createContext, useContext } from 'react'

export type Modo = 'light' | 'dark' | 'system'
export type ModoResuelto = 'light' | 'dark'

export const CLAVE_TEMA = 'app.tema.v1'

export type TemaContextValue = {
  modo: Modo
  esOscuro: boolean
  setModo: (modo: Modo) => void
  alternarModo: () => void
}

export const TemaContext = createContext<TemaContextValue | null>(null)

function esModo(valor: unknown): valor is Modo {
  return valor === 'light' || valor === 'dark' || valor === 'system'
}

export function leerModoGuardado(): Modo {
  try {
    const crudo = localStorage.getItem(CLAVE_TEMA)
    if (!crudo) {
      return 'system'
    }
    const parseado: unknown = JSON.parse(crudo)
    return esModo(parseado) ? parseado : 'system'
  } catch {
    return 'system'
  }
}

export function resolverModo(modo: Modo): ModoResuelto {
  if (modo === 'system') {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
  }
  return modo
}

export function useTema(): TemaContextValue {
  const contexto = useContext(TemaContext)
  if (!contexto) {
    throw new Error('useTema debe usarse dentro de ThemeProvider')
  }
  return contexto
}
