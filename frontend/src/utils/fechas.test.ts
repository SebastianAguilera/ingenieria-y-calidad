import { describe, expect, it } from 'vitest'
import { formatearFecha } from './fechas'

describe('formatearFecha', () => {
  it('formatea una fecha YYYY-MM-DD como dd/mm/yyyy', () => {
    expect(formatearFecha('2026-01-15')).toBe('15/01/2026')
  })

  it('toma solo la parte de fecha de un timestamp ISO', () => {
    expect(formatearFecha('2026-01-01T00:00:00Z')).toBe('01/01/2026')
  })

  it('devuelve "Sin fecha" cuando la fecha es null', () => {
    expect(formatearFecha(null)).toBe('Sin fecha')
  })
})
