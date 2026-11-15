import { describe, expect, it } from 'vitest'
import type { DispersionPoker } from '../../../api/types'
import { ESCALA_POKER, esValorDeEscala, sugerirAcuerdo } from './escala'

function dispersion(parcial: Partial<DispersionPoker>): DispersionPoker {
  return {
    promedio: 0,
    mediana: 0,
    minimo: 0,
    maximo: 0,
    rango: 0,
    desviacion_estandar: 0,
    hay_dispersion: false,
    total_votos: 0,
    votos_por_valor: {},
    ...parcial,
  }
}

describe('ESCALA_POKER', () => {
  it('contiene los nueve valores de la escala de US-05', () => {
    expect([...ESCALA_POKER]).toEqual([0, 0.5, 1, 2, 3, 5, 8, 13, 21])
  })
})

describe('esValorDeEscala', () => {
  it('acepta valores de la escala y rechaza el resto', () => {
    expect(esValorDeEscala(0)).toBe(true)
    expect(esValorDeEscala(0.5)).toBe(true)
    expect(esValorDeEscala(21)).toBe(true)
    expect(esValorDeEscala(4)).toBe(false)
    expect(esValorDeEscala(-1)).toBe(false)
  })
})

describe('sugerirAcuerdo', () => {
  it('usa la mediana cuando pertenece a la escala', () => {
    const resultado = sugerirAcuerdo(
      dispersion({ mediana: 3, total_votos: 3, votos_por_valor: { '1': 1, '3': 1, '8': 1 } }),
    )

    expect(resultado).toBe(3)
  })

  it('usa el valor más votado cuando la mediana no pertenece a la escala', () => {
    const resultado = sugerirAcuerdo(
      dispersion({ mediana: 4, total_votos: 4, votos_por_valor: { '3': 2, '8': 1, '13': 1 } }),
    )

    expect(resultado).toBe(3)
  })

  it('deja vacío cuando hay empate en la mayor cantidad de votos', () => {
    const resultado = sugerirAcuerdo(
      dispersion({ mediana: 4, total_votos: 3, votos_por_valor: { '1': 1, '3': 1, '8': 1 } }),
    )

    expect(resultado).toBeNull()
  })

  it('deja vacío cuando no hay dispersión', () => {
    expect(sugerirAcuerdo(undefined)).toBeNull()
    expect(sugerirAcuerdo(null)).toBeNull()
  })

  it('deja vacío cuando no hay votos', () => {
    const resultado = sugerirAcuerdo(dispersion({ mediana: 0, total_votos: 0, votos_por_valor: {} }))

    expect(resultado).toBeNull()
  })

  it('prefija 0 cuando la mediana es 0 y hay votos', () => {
    const resultado = sugerirAcuerdo(
      dispersion({ mediana: 0, total_votos: 3, votos_por_valor: { '0': 3 } }),
    )

    expect(resultado).toBe(0)
  })
})
