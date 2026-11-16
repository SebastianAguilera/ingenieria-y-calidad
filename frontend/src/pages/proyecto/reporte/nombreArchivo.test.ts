import { describe, expect, it } from 'vitest'
import { fechaLocal, nombreArchivoProyecto, nombreArchivoSprint, slug } from './nombreArchivo'

// US-F23, D-03: mismo esquema de nombres que pdf_service.go (US-24).

describe('slug', () => {
  it.each([
    ['Gestión Ágil', 'gestion-agil'],
    ['Gestión & Co', 'gestion-co'],
    ['  Sprint 2  ', 'sprint-2'],
    ['ÑANDÚ', 'nandu'],
    ['---', 'reporte'],
    ['', 'reporte'],
  ])('%j → %s', (texto, esperado) => {
    expect(slug(texto)).toBe(esperado)
  })

  it('recorta a 50 caracteres sin dejar un guion al final', () => {
    const resultado = slug(`${'a'.repeat(49)} b`)
    expect(resultado).toBe('a'.repeat(49))
  })
})

describe('fechaLocal', () => {
  it('devuelve el día local en formato AAAA-MM-DD', () => {
    expect(fechaLocal(new Date(2026, 10, 6, 23, 30))).toBe('2026-11-06')
  })
})

describe('nombres de archivo', () => {
  it('reporte del proyecto: reporte-{nombre}-proyecto-{fecha}.{ext}', () => {
    expect(nombreArchivoProyecto('Gestión Ágil', '2026-11-16', 'json')).toBe(
      'reporte-gestion-agil-proyecto-2026-11-16.json',
    )
  })

  it('reporte de un sprint: reporte-proyecto-{id}-sprint-{sprint}-{fecha}.{ext}', () => {
    expect(nombreArchivoSprint(1, 'Sprint 2', '2026-11-16', 'pdf')).toBe(
      'reporte-proyecto-1-sprint-sprint-2-2026-11-16.pdf',
    )
  })
})
