import { createContext, useContext } from 'react'
import type { SesionPokerActiva } from '../../../api/types'

export const INTERVALO_POLLING_MS = 3000
export const DURACION_TOAST_MS = 5000

export const ETIQUETA_CAMPANA = 'Sesiones de Planning Poker activas'
export const ID_LISTA_SESIONES = 'campana-poker-lista'
export const MENSAJE_SIN_SESIONES = 'No hay sesiones de Planning Poker activas.'
export const ETIQUETA_LISTA = 'Sesiones activas de Planning Poker'

export type SesionesPokerContextValue = {
  proyectoId: number | null
  sesiones: readonly SesionPokerActiva[]
  historiasConPoker: ReadonlySet<number>
}

export const SesionesPokerContext = createContext<SesionesPokerContextValue | null>(null)

export function useSesionesPoker(): SesionesPokerContextValue {
  const valor = useContext(SesionesPokerContext)
  if (!valor) {
    throw new Error('useSesionesPoker debe usarse dentro de ProveedorSesionesPoker')
  }
  return valor
}
