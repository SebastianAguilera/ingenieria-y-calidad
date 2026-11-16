import { describe, expect, it } from 'vitest'
import { formatearDesviacion, formatearHoras, formatearNumero, formatearPorcentaje } from './formato'

describe('formatearPorcentaje', () => {
  it.each([
    [0.75, '75%'],
    [0, '0%'],
    [1, '100%'],
    [0.5, '50%'],
    [0.6153846153846154, '61,5%'],
    [0.333333, '33,3%'],
  ])('%s se muestra como %s', (valor, esperado) => {
    expect(formatearPorcentaje(valor)).toBe(esperado)
  })
})

describe('formatearDesviacion', () => {
  it.each([
    [0.2, '+20%'],
    [-0.1, '−10%'],
    [0, '0%'],
    [1.255, '+125,5%'],
  ])('%s se muestra como %s', (valor, esperado) => {
    expect(formatearDesviacion(valor)).toBe(esperado)
  })
})

describe('formatearNumero', () => {
  it.each([
    [8, '8'],
    [0, '0'],
    [6.5, '6,5'],
    [0.3333333333333333, '0,33'],
    [1.5, '1,5'],
  ])('%s se muestra como %s', (valor, esperado) => {
    expect(formatearNumero(valor)).toBe(esperado)
  })
})

describe('formatearHoras', () => {
  it('agrega la unidad', () => {
    expect(formatearHoras(5)).toBe('5 h')
    expect(formatearHoras(0.25)).toBe('0,25 h')
  })
})
