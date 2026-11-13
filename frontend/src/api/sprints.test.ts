import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { Historia, Sprint, SprintBacklog } from './types'
import { api } from './client'
import {
  asignarHistoria,
  cambiarEstadoSprint,
  crearSprint,
  listarSprints,
  obtenerSprintBacklog,
  quitarHistoria,
} from './sprints'

vi.mock('./client', () => ({
  api: {
    get: vi.fn(),
    post: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
  },
}))

const getMock = vi.mocked(api.get)
const postMock = vi.mocked(api.post)
const patchMock = vi.mocked(api.patch)
const deleteMock = vi.mocked(api.delete)

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

const historia: Historia = {
  id: 100,
  codigo: 'US-01',
  titulo: 'Crear proyecto',
  descripcion: 'Como usuario...',
  prioridad: 'Alta',
  estado: 'En Progreso',
  story_points: 3,
  esfuerzo_real: null,
  creado_en: '2026-04-01T00:00:00Z',
  actualizado_en: '2026-04-02T00:00:00Z',
}

const backlog: SprintBacklog = {
  sprint: { ...sprint, estado: 'Activo' },
  total: 1,
  items: [{ id: 10, orden: 1, fecha_asignacion: '2026-05-01T10:00:00Z', historia }],
}

describe('obtenerSprintBacklog', () => {
  beforeEach(() => {
    getMock.mockReset()
  })

  it('consulta GET /proyectos/:id/sprints/:sprintId/historias y devuelve el sprint con sus items', async () => {
    getMock.mockResolvedValue(backlog)

    const resultado = await obtenerSprintBacklog(1, 5)

    expect(getMock).toHaveBeenCalledWith('/proyectos/1/sprints/5/historias', {
      signal: undefined,
    })
    expect(resultado).toEqual(backlog)
  })

  it('reenvía la señal de cancelación', async () => {
    getMock.mockResolvedValue(backlog)
    const controlador = new AbortController()

    await obtenerSprintBacklog(1, 5, controlador.signal)

    expect(getMock).toHaveBeenCalledWith('/proyectos/1/sprints/5/historias', {
      signal: controlador.signal,
    })
  })

  it('tolera una respuesta con items nulos o ausentes', async () => {
    getMock.mockResolvedValue({ sprint: backlog.sprint, total: 0, items: null })

    const resultado = await obtenerSprintBacklog(1, 5)

    expect(resultado.items).toEqual([])
    expect(resultado.sprint).toEqual(backlog.sprint)
  })
})

describe('asignarHistoria', () => {
  beforeEach(() => {
    postMock.mockReset()
  })

  it('envía POST /proyectos/:id/sprints/:sprintId/historias con historia_id', async () => {
    postMock.mockResolvedValue(undefined)

    await asignarHistoria(1, 5, 100)

    expect(postMock).toHaveBeenCalledWith('/proyectos/1/sprints/5/historias', {
      historia_id: 100,
    })
  })

  it('propaga el error del backend', async () => {
    postMock.mockRejectedValue(new Error('409'))

    await expect(asignarHistoria(1, 5, 100)).rejects.toThrow('409')
  })
})

describe('quitarHistoria', () => {
  beforeEach(() => {
    deleteMock.mockReset()
  })

  it('envía DELETE /proyectos/:id/sprints/:sprintId/historias/:historiaId', async () => {
    deleteMock.mockResolvedValue(undefined)

    await quitarHistoria(1, 5, 100)

    expect(deleteMock).toHaveBeenCalledWith('/proyectos/1/sprints/5/historias/100')
  })

  it('propaga el error del backend', async () => {
    deleteMock.mockRejectedValue(new Error('409'))

    await expect(quitarHistoria(1, 5, 100)).rejects.toThrow('409')
  })
})
