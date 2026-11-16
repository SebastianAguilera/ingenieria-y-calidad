import { describe, expect, it } from 'vitest'
import { LIMITE_DESCRIPCION_EDICION, validarAltaDefecto, validarEdicionDefecto } from './validacionDefecto'
import type { CamposAltaDefecto, CamposEdicionDefecto } from './validacionDefecto'

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

describe('validarEdicionDefecto', () => {
  const edicion: CamposEdicionDefecto = { descripcion: 'Error al guardar', severidad: 'major' }

  it('el límite de la edición es de 1000 caracteres', () => {
    expect(LIMITE_DESCRIPCION_EDICION).toBe(1000)
  })

  it('no devuelve errores con datos válidos', () => {
    expect(validarEdicionDefecto(edicion)).toEqual({})
  })

  it.each([
    ['vacía', ''],
    ['solo espacios', '   '],
  ])('rechaza una descripción %s', (_caso, descripcion) => {
    expect(validarEdicionDefecto({ ...edicion, descripcion })).toEqual({
      descripcion: 'La descripción es obligatoria',
    })
  })

  it('acepta una descripción de exactamente 1000 caracteres', () => {
    expect(validarEdicionDefecto({ ...edicion, descripcion: 'a'.repeat(1000) })).toEqual({})
  })

  it('rechaza una descripción de 1001 caracteres', () => {
    expect(validarEdicionDefecto({ ...edicion, descripcion: 'a'.repeat(1001) })).toEqual({
      descripcion: 'La descripción no puede superar los 1000 caracteres',
    })
  })

  it('cuenta caracteres y no bytes: 1000 "ñ" es válido', () => {
    expect(validarEdicionDefecto({ ...edicion, descripcion: 'ñ'.repeat(1000) })).toEqual({})
  })

  it('cuenta caracteres Unicode fuera del plano básico como uno solo', () => {
    expect(validarEdicionDefecto({ ...edicion, descripcion: '😀'.repeat(1000) })).toEqual({})
  })

  it('recorta los espacios antes de contar el largo', () => {
    expect(
      validarEdicionDefecto({ ...edicion, descripcion: `  ${'a'.repeat(1000)}  ` }),
    ).toEqual({})
  })
})
