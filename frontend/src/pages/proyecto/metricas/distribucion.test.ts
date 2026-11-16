import { describe, expect, it } from 'vitest'
import type { EstadoSprint, Historia, Sprint } from '../../../api/types'
import { contarPorEstado, sprintInicial, sprintsDelGrafico } from './distribucion'

function historia(id: number, estado: Historia['estado']): Historia {
  return {
    id,
    codigo: `US-${id}`,
    titulo: `Historia ${id}`,
    descripcion: '',
    prioridad: 'Media',
    estado,
    story_points: null,
    esfuerzo_real: null,
    horas_estimadas: null,
    valor_negocio: null,
    criticidad_tiempo: null,
    reduccion_riesgo: null,
    creado_en: '2026-04-01T00:00:00Z',
    actualizado_en: '2026-04-01T00:00:00Z',
  }
}

function sprint(id: number, estado: EstadoSprint): Sprint {
  return {
    id,
    identificador: `Sprint ${id}`,
    sprint_goal: '',
    estado,
    fecha_inicio: '2026-05-01',
    fecha_fin: '2026-05-15',
    creado_en: '2026-04-01T00:00:00Z',
    actualizado_en: '2026-04-01T00:00:00Z',
  }
}

describe('contarPorEstado', () => {
  it('cuenta las historias de cada estado en orden Backlog, En Progreso, Hecho', () => {
    const historias = [
      historia(1, 'Hecho'),
      historia(2, 'Backlog'),
      historia(3, 'Hecho'),
      historia(4, 'En Progreso'),
      historia(5, 'Backlog'),
      historia(6, 'Hecho'),
    ]

    expect(contarPorEstado(historias)).toEqual([
      { estado: 'Backlog', cantidad: 2 },
      { estado: 'En Progreso', cantidad: 1 },
      { estado: 'Hecho', cantidad: 3 },
    ])
  })

  it('devuelve ceros sin historias', () => {
    expect(contarPorEstado([])).toEqual([
      { estado: 'Backlog', cantidad: 0 },
      { estado: 'En Progreso', cantidad: 0 },
      { estado: 'Hecho', cantidad: 0 },
    ])
  })
})

describe('sprintsDelGrafico', () => {
  it('incluye cerrados y el activo, en el orden recibido, sin planificados', () => {
    const sprints = [sprint(1, 'Cerrado'), sprint(2, 'Activo'), sprint(3, 'Planificado')]

    expect(sprintsDelGrafico(sprints).map((s) => s.id)).toEqual([1, 2])
  })
})

describe('sprintInicial', () => {
  it('elige el sprint activo', () => {
    expect(sprintInicial([sprint(1, 'Cerrado'), sprint(2, 'Activo')])).toBe(2)
  })

  it('sin activo elige el último cerrado', () => {
    expect(
      sprintInicial([sprint(1, 'Cerrado'), sprint(2, 'Cerrado'), sprint(3, 'Planificado')]),
    ).toBe(2)
  })

  it('sin activo ni cerrados elige el primero', () => {
    expect(sprintInicial([sprint(4, 'Planificado'), sprint(5, 'Planificado')])).toBe(4)
  })

  it('sin sprints devuelve null', () => {
    expect(sprintInicial([])).toBeNull()
  })
})
