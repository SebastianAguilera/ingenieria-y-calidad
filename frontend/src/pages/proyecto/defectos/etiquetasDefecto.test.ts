import { describe, expect, it } from 'vitest'
import { COLOR_SEVERIDAD } from '../metricas/estilosGrafico'
import { SEVERIDADES_DEFECTO, claseSeveridadDefecto, colorSeveridad } from './etiquetasDefecto'

// Paleta de los gráficos de métricas (Story points por sprint e Historias por
// estado): --chart-1, --chart-2, --chart-3 y su variante intensa --accent-600.
const PALETA = ['var(--chart-3)', 'var(--chart-1)', 'var(--chart-2)', 'var(--accent-600)']

describe('colores de severidad', () => {
  it('van de Menor a Bloqueante, del más suave al más intenso, con la paleta de los gráficos', () => {
    expect(SEVERIDADES_DEFECTO.map((severidad) => colorSeveridad(severidad))).toEqual(PALETA)
  })

  it('cada severidad tiene un color distinto', () => {
    expect(new Set(Object.values(COLOR_SEVERIDAD)).size).toBe(SEVERIDADES_DEFECTO.length)
  })

  it.each(SEVERIDADES_DEFECTO)('el chip de %s usa el mismo color que el gráfico', (severidad) => {
    expect(claseSeveridadDefecto(severidad)).toContain(colorSeveridad(severidad))
  })

  it('no define colores propios fuera de la paleta de los gráficos', () => {
    for (const severidad of SEVERIDADES_DEFECTO) {
      expect(claseSeveridadDefecto(severidad)).not.toContain('--severidad-')
    }
  })
})
