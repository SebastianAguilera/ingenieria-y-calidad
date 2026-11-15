import type { EstadoHistoria, PrioridadHistoria } from '../../../api/types'

export type DatosHistoria = {
  titulo: string
  descripcion: string
  prioridad: PrioridadHistoria
  estado: EstadoHistoria
}

export type ResultadoOperacion =
  | { ok: true }
  | { ok: false; mensaje: string; campo?: string }
