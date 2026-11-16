import { describe, expect, it } from 'vitest'
import {
  ESTADOS_HISTORIA,
  claseColumnaEstado,
  claseContadorEstado,
  clasePuntoEstado,
  claseEstado,
  claseTarjetaEstado,
} from './etiquetas'

const VARIABLE_POR_ESTADO = {
  Backlog: '--estado-backlog',
  'En Progreso': '--estado-progreso',
  Hecho: '--estado-hecho',
} as const

describe('colores por estado de historia', () => {
  it.each(ESTADOS_HISTORIA)('%s usa su propia variable en todas las clases', (estado) => {
    const variable = VARIABLE_POR_ESTADO[estado]

    for (const clase of [
      claseEstado(estado),
      claseColumnaEstado(estado),
      clasePuntoEstado(estado),
      claseContadorEstado(estado),
      claseTarjetaEstado(estado),
    ]) {
      expect(clase).toContain(variable)
    }
  })

  it('cada estado tiene un color distinto', () => {
    const chips = ESTADOS_HISTORIA.map((estado) => claseEstado(estado))
    expect(new Set(chips).size).toBe(ESTADOS_HISTORIA.length)
  })

  it('En Progreso es ámbar y ya no usa el rosa del acento', () => {
    for (const clase of [
      claseEstado('En Progreso'),
      claseColumnaEstado('En Progreso'),
      clasePuntoEstado('En Progreso'),
      claseContadorEstado('En Progreso'),
      claseTarjetaEstado('En Progreso'),
    ]) {
      expect(clase).not.toMatch(/accent|danger|pink/)
    }
  })

  it('la columna lleva franja superior de 4 px y la tarjeta borde izquierdo de 3 px', () => {
    expect(claseColumnaEstado('Hecho')).toContain('border-t-4')
    expect(claseTarjetaEstado('Hecho')).toContain('border-l-[3px]')
  })
})
