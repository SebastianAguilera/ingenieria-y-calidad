import type { EstadoHistoria, PrioridadHistoria } from '../../../api/types'

export type DatosHistoria = {
  titulo: string
  descripcion: string
  prioridad: PrioridadHistoria
  estado: EstadoHistoria
  valor_negocio: number | null
  criticidad_tiempo: number | null
  reduccion_riesgo: number | null
  /**
   * Horas estimadas (US-19). En el alta, `null` si el campo quedó vacío. En la
   * edición se omite cuando está vacío: el `PATCH` no borra el valor guardado.
   */
  horas_estimadas?: number | null
}

export type ResultadoOperacion =
  | { ok: true }
  | { ok: false; mensaje: string; campo?: string }
