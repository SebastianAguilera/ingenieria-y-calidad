import { describe, expect, it } from 'vitest'
import { validarAltaDefecto } from './validacionDefecto'
import type { CamposAltaDefecto } from './validacionDefecto'

const valido: CamposAltaDefecto = {
  descripcion: 'Error al guardar',
  severidad: 'major',
  historia_id: '1',
  sprint_deteccion_id: '2',
}

describe('validarAltaDefecto', () => {
  it('no devuelve errores con datos válidos', () => {
    expect(validarAltaDefecto(valido)).toEqual({})
  })

  it.each([
    ['vacía', ''],
    ['solo espacios', '   '],
  ])('rechaza una descripción %s', (_caso, descripcion) => {
    expect(validarAltaDefecto({ ...valido, descripcion })).toEqual({
      descripcion: 'La descripción es obligatoria',
    })
  })

  it('rechaza una severidad sin elegir', () => {
    expect(validarAltaDefecto({ ...valido, severidad: '' })).toEqual({
      severidad: 'Elegí una severidad',
    })
  })

  it('rechaza una historia sin elegir', () => {
    expect(validarAltaDefecto({ ...valido, historia_id: '' })).toEqual({
      historia_id: 'Elegí una historia',
    })
  })

  it('rechaza un sprint de detección sin elegir', () => {
    expect(validarAltaDefecto({ ...valido, sprint_deteccion_id: '' })).toEqual({
      sprint_deteccion_id: 'Elegí el sprint de detección',
    })
  })

  it('acumula todos los errores a la vez', () => {
    const errores = validarAltaDefecto({
      descripcion: '',
      severidad: '',
      historia_id: '',
      sprint_deteccion_id: '',
    })

    expect(Object.keys(errores).sort()).toEqual([
      'descripcion',
      'historia_id',
      'severidad',
      'sprint_deteccion_id',
    ])
  })

  it('el alta no limita el largo de la descripción (el backend tampoco)', () => {
    expect(validarAltaDefecto({ ...valido, descripcion: 'a'.repeat(1500) })).toEqual({})
  })
})
