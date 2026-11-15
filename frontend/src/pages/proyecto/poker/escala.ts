import type { DispersionPoker } from '../../../api/types'

export const ESCALA_POKER = [0, 0.5, 1, 2, 3, 5, 8, 13, 21] as const

export function esValorDeEscala(valor: number): boolean {
  return (ESCALA_POKER as readonly number[]).includes(valor)
}

export function sugerirAcuerdo(dispersion: DispersionPoker | null | undefined): number | null {
  if (!dispersion || dispersion.total_votos === 0) {
    return null
  }

  if (esValorDeEscala(dispersion.mediana)) {
    return dispersion.mediana
  }

  const entradas = Object.entries(dispersion.votos_por_valor).map(([valor, cantidad]) => ({
    valor: Number(valor),
    cantidad,
  }))
  if (entradas.length === 0) {
    return null
  }

  const maximo = Math.max(...entradas.map((entrada) => entrada.cantidad))
  const ganadores = entradas.filter((entrada) => entrada.cantidad === maximo)
  return ganadores.length === 1 ? ganadores[0].valor : null
}
