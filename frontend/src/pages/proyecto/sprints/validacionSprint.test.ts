import { describe, expect, it } from 'vitest'
import {
  LIMITE_IDENTIFICADOR_SPRINT,
  LIMITE_SPRINT_GOAL,
  validarSprint,
} from './validacionSprint'

const valido = {
  identificador: 'Sprint 1',
  sprint_goal: 'Entregar el MVP',
  fecha_inicio: '2026-05-01',
  fecha_fin: '2026-05-15',
}

describe('validarSprint', () => {
  it('acepta un sprint válido', () => {
    expect(validarSprint(valido)).toEqual({})
  })

  it('exige el identificador', () => {
    expect(validarSprint({ ...valido, identificador: '   ' })).toEqual({
      identificador: 'El identificador es obligatorio',
    })
  })

  it('acepta un identificador de exactamente 50 caracteres (límite)', () => {
    const identificador = 'a'.repeat(LIMITE_IDENTIFICADOR_SPRINT)

    expect(validarSprint({ ...valido, identificador })).toEqual({})
  })

  it('rechaza un identificador de 51 caracteres', () => {
    expect(
      validarSprint({ ...valido, identificador: 'a'.repeat(LIMITE_IDENTIFICADOR_SPRINT + 1) }),
    ).toEqual({ identificador: 'El identificador no puede superar los 50 caracteres' })
  })

  it('exige el Sprint Goal', () => {
    expect(validarSprint({ ...valido, sprint_goal: '' })).toEqual({
      sprint_goal: 'El Sprint Goal es obligatorio',
    })
  })

  it('acepta un Sprint Goal de exactamente 300 caracteres (límite)', () => {
    const sprint_goal = 'a'.repeat(LIMITE_SPRINT_GOAL)

    expect(validarSprint({ ...valido, sprint_goal })).toEqual({})
  })

  it('rechaza un Sprint Goal de 301 caracteres', () => {
    expect(validarSprint({ ...valido, sprint_goal: 'a'.repeat(LIMITE_SPRINT_GOAL + 1) })).toEqual({
      sprint_goal: 'El Sprint Goal no puede superar los 300 caracteres',
    })
  })

  it('exige ambas fechas', () => {
    expect(validarSprint({ ...valido, fecha_inicio: '', fecha_fin: '' })).toEqual({
      fecha_inicio: 'La fecha de inicio es obligatoria',
      fecha_fin: 'La fecha de fin es obligatoria',
    })
  })

  it('rechaza una fecha de fin anterior a la de inicio', () => {
    expect(validarSprint({ ...valido, fecha_inicio: '2026-05-15', fecha_fin: '2026-05-01' })).toEqual(
      { fecha_fin: 'La fecha de fin debe ser igual o posterior a la de inicio' },
    )
  })

  it('acepta fechas iguales (límite)', () => {
    expect(validarSprint({ ...valido, fecha_inicio: '2026-05-01', fecha_fin: '2026-05-01' })).toEqual(
      {},
    )
  })
})
