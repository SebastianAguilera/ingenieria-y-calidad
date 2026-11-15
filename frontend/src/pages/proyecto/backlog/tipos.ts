import type { EstadoHistoria, PrioridadHistoria } from '../../../api/types'

export type DatosHistoria = {
  titulo: string
  descripcion: string
  prioridad: PrioridadHistoria
  estado: EstadoHistoria
  valor_negocio: number | null
  criticidad_tiempo: number | null
  reduccion_riesgo: number | null
}

export type ResultadoOperacion =
  | { ok: true }
  | { ok: false; mensaje: string; campo?: string }
