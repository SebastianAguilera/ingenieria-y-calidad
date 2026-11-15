import { describe, expect, it } from 'vitest'
import { LIMITE_NOMBRE_PROYECTO, validarProyecto } from './validacionProyecto'

describe('validarProyecto', () => {
  it('acepta un proyecto válido sin fecha de fin', () => {
    expect(validarProyecto({ nombre: 'Alfa', fecha_inicio: '2026-01-01', fecha_fin: '' })).toEqual(
      {},
    )
  })

  it('rechaza el nombre vacío o solo con espacios', () => {
    for (const nombre of ['', '   ']) {
      expect(validarProyecto({ nombre, fecha_inicio: '2026-01-01', fecha_fin: '' })).toEqual({
        nombre: 'El nombre es obligatorio',
      })
    }
  })

  it('acepta un nombre de exactamente 150 caracteres (límite)', () => {
    const nombre = 'a'.repeat(LIMITE_NOMBRE_PROYECTO)

    expect(validarProyecto({ nombre, fecha_inicio: '2026-01-01', fecha_fin: '' })).toEqual({})
  })

  it('rechaza un nombre de 151 caracteres', () => {
    const nombre = 'a'.repeat(LIMITE_NOMBRE_PROYECTO + 1)

    expect(validarProyecto({ nombre, fecha_inicio: '2026-01-01', fecha_fin: '' })).toEqual({
      nombre: 'El nombre no puede superar los 150 caracteres',
    })
  })

  it('mide la longitud del nombre después de recortar espacios', () => {
    const nombre = `  ${'a'.repeat(LIMITE_NOMBRE_PROYECTO)}  `

    expect(validarProyecto({ nombre, fecha_inicio: '2026-01-01', fecha_fin: '' })).toEqual({})
  })

  it('exige la fecha de inicio', () => {
    expect(validarProyecto({ nombre: 'Alfa', fecha_inicio: '', fecha_fin: '' })).toEqual({
      fecha_inicio: 'La fecha de inicio es obligatoria',
    })
  })

  it('rechaza una fecha de fin anterior a la de inicio', () => {
    expect(
      validarProyecto({ nombre: 'Alfa', fecha_inicio: '2026-02-01', fecha_fin: '2026-01-31' }),
    ).toEqual({ fecha_fin: 'La fecha de fin debe ser igual o posterior a la de inicio' })
  })

  it('acepta fechas de inicio y fin iguales (límite)', () => {
    expect(
      validarProyecto({ nombre: 'Alfa', fecha_inicio: '2026-02-01', fecha_fin: '2026-02-01' }),
    ).toEqual({})
  })

  it('acepta una fecha de fin posterior a la de inicio', () => {
    expect(
      validarProyecto({ nombre: 'Alfa', fecha_inicio: '2026-02-01', fecha_fin: '2026-03-01' }),
    ).toEqual({})
  })

  it('informa todos los errores a la vez', () => {
    expect(validarProyecto({ nombre: '', fecha_inicio: '', fecha_fin: '' })).toEqual({
      nombre: 'El nombre es obligatorio',
      fecha_inicio: 'La fecha de inicio es obligatoria',
    })
  })
})
