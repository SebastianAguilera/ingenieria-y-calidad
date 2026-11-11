import { createContext, useContext } from 'react'
import type { Proyecto } from '../../api/types'

export type EstadoProyectoDetalle =
  | { tipo: 'cargando' }
  | { tipo: 'error' }
  | { tipo: 'noEncontrado' }
  | { tipo: 'listo'; proyecto: Proyecto }

export type ValorProyectoContexto = {
  estado: EstadoProyectoDetalle
  reintentar: () => void
}

export const ProyectoContext = createContext<ValorProyectoContexto | null>(null)

export function useProyecto(): ValorProyectoContexto {
  const contexto = useContext(ProyectoContext)
  if (!contexto) {
    throw new Error('useProyecto tiene que usarse dentro de ProyectoProvider')
  }
  return contexto
}
