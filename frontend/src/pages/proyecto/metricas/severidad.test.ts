import { describe, expect, it } from 'vitest'
import { desgloseSeveridad, resumenSeveridad } from './severidad'

describe('desgloseSeveridad', () => {
  it('devuelve siempre las cuatro severidades en orden fijo, sin reordenar por valor', () => {
    const desglose = desgloseSeveridad({ minor: 0, major: 5, critical: 2, blocker: 9 })

    expect(desglose.map(({ etiqueta }) => etiqueta)).toEqual([
      'Menor',
      'Mayor',
      'Crítica',
      'Bloqueante',
    ])
    expect(desglose.map(({ valor }) => valor)).toEqual([0, 5, 2, 9])
  })

  it('conserva blocker con valor 0', () => {
    const desglose = desgloseSeveridad({ minor: 1, major: 1, critical: 1, blocker: 0 })

    expect(desglose[3]).toMatchObject({ clave: 'blocker', etiqueta: 'Bloqueante', valor: 0 })
  })

  it('toma como 0 una clave faltante', () => {
    const incompleto = { minor: 2, major: 1, critical: 0 } as Parameters<typeof desgloseSeveridad>[0]

    expect(desgloseSeveridad(incompleto).map(({ valor }) => valor)).toEqual([2, 1, 0, 0])
  })
})

describe('resumenSeveridad', () => {
  it('lista las cuatro severidades con sus valores', () => {
    const resumen = resumenSeveridad(desgloseSeveridad({ minor: 1, major: 1, critical: 1, blocker: 0 }))

    expect(resumen).toBe('Menor: 1, Mayor: 1, Crítica: 1, Bloqueante: 0')
  })

  it('indica que no hay defectos cuando todo vale 0', () => {
    const resumen = resumenSeveridad(desgloseSeveridad({ minor: 0, major: 0, critical: 0, blocker: 0 }))

    expect(resumen).toBe('Sin defectos registrados.')
  })
})
