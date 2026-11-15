import { beforeEach, describe, expect, it, vi } from 'vitest'
import { api } from './client'
import { actualizarHistoria, crearHistoria, listarHistorias } from './historias'
import type { Historia } from './types'

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

const historia: Historia = {
  id: 10,
  codigo: 'US-01',
  titulo: 'Registrar horas reales',
  descripcion: 'Permitir registrar horas por usuario.',
  prioridad: 'Alta',
  estado: 'Backlog',
  story_points: 5,
  esfuerzo_real: null,
  creado_en: '2026-10-01T09:15:00Z',
  actualizado_en: '2026-10-01T09:15:00Z',
}

describe('listarHistorias', () => {
  beforeEach(() => {
    getMock.mockReset()
  })

  it('consulta GET /proyectos/:id/historias y devuelve el arreglo', async () => {
    getMock.mockResolvedValue({ total: 1, historias: [historia] })

    const historias = await listarHistorias(1)

    expect(getMock).toHaveBeenCalledWith('/proyectos/1/historias', { signal: undefined })
    expect(historias).toEqual([historia])
  })

  it('devuelve un arreglo vacío cuando el proyecto no tiene historias', async () => {
    getMock.mockResolvedValue({ total: 0, historias: [] })

    expect(await listarHistorias(2)).toEqual([])
  })

  it('tolera una respuesta sin el campo historias', async () => {
    getMock.mockResolvedValue({ total: 0 })

    expect(await listarHistorias(2)).toEqual([])
  })
})

describe('crearHistoria', () => {
  beforeEach(() => {
    postMock.mockReset()
  })

  it('envía POST /proyectos/:id/historias con el alta', async () => {
    postMock.mockResolvedValue(historia)
    const alta = {
      titulo: 'Registrar horas reales',
      descripcion: 'Permitir registrar horas por usuario.',
      prioridad: 'Alta' as const,
      estado: 'Backlog' as const,
    }

    const creada = await crearHistoria(1, alta)

    expect(postMock).toHaveBeenCalledWith('/proyectos/1/historias', alta)
    expect(creada).toBe(historia)
  })
})

describe('actualizarHistoria', () => {
  beforeEach(() => {
    patchMock.mockReset()
  })

  it('envía PATCH /proyectos/:id/historias/:historiaId con los cambios', async () => {
    patchMock.mockResolvedValue({ ...historia, estado: 'Hecho' })

    const actualizada = await actualizarHistoria(1, 10, { estado: 'Hecho' })

    expect(patchMock).toHaveBeenCalledWith('/proyectos/1/historias/10', { estado: 'Hecho' })
    expect(actualizada.estado).toBe('Hecho')
  })
})
