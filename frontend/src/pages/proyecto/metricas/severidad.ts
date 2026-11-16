import type { DefectosPorSeveridad, SeveridadDefecto } from '../../../api/types'
import { etiquetaSeveridad, SEVERIDADES_DEFECTO } from '../defectos/etiquetasDefecto'

export type FilaSeveridad = {
  clave: SeveridadDefecto
  etiqueta: string
  valor: number
}

/**
 * Desglose de defectos por severidad (US-F22, RB-06): siempre las cuatro
 * severidades, en el orden fijo de `SEVERIDADES_DEFECTO` y sin reordenar por
 * valor. Una clave faltante se toma como `0`.
 */
export function desgloseSeveridad(defectos: DefectosPorSeveridad): FilaSeveridad[] {
  return SEVERIDADES_DEFECTO.map((clave) => ({
    clave,
    etiqueta: etiquetaSeveridad(clave),
    valor: defectos[clave] ?? 0,
  }))
}

/** Resumen en texto del gráfico, para lectores de pantalla (RB-08). */
export function resumenSeveridad(desglose: FilaSeveridad[]): string {
  if (desglose.every(({ valor }) => valor === 0)) {
    return 'Sin defectos registrados.'
  }
  return desglose.map(({ etiqueta, valor }) => `${etiqueta}: ${valor}`).join(', ')
}
