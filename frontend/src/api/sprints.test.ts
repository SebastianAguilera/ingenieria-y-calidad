import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { Sprint } from './types'
import { api } from './client'
import { cambiarEstadoSprint, crearSprint, listarSprints } from './sprints'

vi.mock('./client', () => ({
  api: {
    get: vi.fn(),
    post: vi.fn(),
    patch: vi.fn(),
  },
}))

const getMock = vi.mocked(api.get)
const postMock = vi.mocked(api.post)
const patchMock = vi.mocked(api.patch)

const sprint: Sprint = {
  id: 5,
  identificador: 'Sprint 1',
  sprint_goal: 'Entregar el MVP',
  estado: 'Planificado',
  fecha_inicio: '2026-05-01',
  fecha_fin: '2026-05-15',
  creado_en: '2026-04-01T00:00:00Z',
  actualizado_en: '2026-04-01T00:00:00Z',
}

describe('listarSprints', () => {
  beforeEach(() => {
    getMock.mockReset()
  })

  it('consulta GET /proyectos/:id/sprints y devuelve el arreglo', async () => {
    getMock.mockResolvedValue({ total: 1, sprints: [sprint] })

    const sprints = await listarSprints(1)

    expect(getMock).toHaveBeenCalledWith('/proyectos/1/sprints', { signal: undefined })
    expect(sprints).toEqual([sprint])
  })

  it('tolera una respuesta sin el campo sprints', async () => {
    getMock.mockResolvedValue({ total: 0 })

    expect(await listarSprints(1)).toEqual([])
  })
})

describe('crearSprint', () => {
  beforeEach(() => {
    postMock.mockReset()
  })

  it('envía POST /proyectos/:id/sprints con el alta', async () => {
    postMock.mockResolvedValue(sprint)
    const alta = {
      identificador: 'Sprint 1',
      sprint_goal: 'Entregar el MVP',
      fecha_inicio: '2026-05-01',
      fecha_fin: '2026-05-15',
    }

    expect(await crearSprint(1, alta)).toBe(sprint)
    expect(postMock).toHaveBeenCalledWith('/proyectos/1/sprints', alta)
  })
})

describe('cambiarEstadoSprint', () => {
  beforeEach(() => {
    patchMock.mockReset()
  })

  it('envía PATCH /proyectos/:id/sprints/:sprintId/estado con el estado', async () => {
    patchMock.mockResolvedValue({ ...sprint, estado: 'Activo' })

    const actualizado = await cambiarEstadoSprint(1, 5, 'Activo')

    expect(patchMock).toHaveBeenCalledWith('/proyectos/1/sprints/5/estado', {
      estado: 'Activo',
    })
    expect(actualizado.estado).toBe('Activo')
  })
})
