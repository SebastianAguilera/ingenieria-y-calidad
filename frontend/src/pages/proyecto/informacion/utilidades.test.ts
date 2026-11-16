import { describe, expect, it } from 'vitest'
import type { UsuarioProyecto } from '../../../api/types'
import {
  duracionEnDias,
  formatearFecha,
  iniciales,
  ordenarEquipo,
  rangoDeFechas,
} from './utilidades'

describe('formatearFecha', () => {
  it('convierte una fecha ISO a dd/mm/aaaa sin corrimiento de zona horaria', () => {
    expect(formatearFecha('2026-01-15')).toBe('15/01/2026')
    expect(formatearFecha('2026-01-15T00:00:00Z')).toBe('15/01/2026')
  })

  it('devuelve "Sin fecha" cuando no hay fecha', () => {
    expect(formatearFecha(null)).toBe('Sin fecha')
  })
})

describe('rangoDeFechas', () => {
  it('une inicio y fin con una flecha', () => {
    expect(rangoDeFechas('2026-01-15', '2026-06-30')).toBe('15/01/2026 → 30/06/2026')
  })

  it('indica que el proyecto no tiene fecha de fin', () => {
    expect(rangoDeFechas('2026-01-15', null)).toBe('Desde 15/01/2026, sin fecha de fin')
  })
})

describe('duracionEnDias', () => {
  it('cuenta los días entre inicio y fin, incluido el último día', () => {
    expect(duracionEnDias('2026-01-01', '2026-01-10')).toBe(10)
  })

  it('devuelve null cuando no hay fecha de fin', () => {
    expect(duracionEnDias('2026-01-01', null)).toBeNull()
  })

  it('devuelve null cuando el fin es anterior al inicio', () => {
    expect(duracionEnDias('2026-02-01', '2026-01-01')).toBeNull()
  })
})

describe('ordenarEquipo', () => {
  const miembros: UsuarioProyecto[] = [
    { id: 3, nombre: 'Carla', email: 'c@x.com', rol: 'MIEMBRO' },
    { id: 1, nombre: 'Ana', email: 'a@x.com', rol: 'LIDER' },
    { id: 2, nombre: 'Bruno', email: 'b@x.com', rol: 'MIEMBRO' },
    { id: 4, nombre: 'Diego', email: 'd@x.com', rol: 'LIDER' },
  ]

  it('muestra primero a los líderes y luego ordena por nombre', () => {
    expect(ordenarEquipo(miembros).map((m) => m.nombre)).toEqual([
      'Ana',
      'Diego',
      'Bruno',
      'Carla',
    ])
  })

  it('no modifica el arreglo original', () => {
    const copia = [...miembros]
    ordenarEquipo(miembros)
    expect(miembros).toEqual(copia)
  })
})

describe('iniciales', () => {
  it('toma la primera letra de las dos primeras palabras', () => {
    expect(iniciales('Ana Ruiz')).toBe('AR')
    expect(iniciales('Carla')).toBe('C')
    expect(iniciales('  juan   perez  gómez ')).toBe('JP')
  })
})
