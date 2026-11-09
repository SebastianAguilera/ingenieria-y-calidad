import { beforeEach, describe, expect, it, vi } from 'vitest'
import { api } from './client'
import { listarProyectos } from './proyectos'
import type { ProyectoListado } from './types'

vi.mock('./client', () => ({
  api: {
    get: vi.fn(),
  },
}))

const getMock = vi.mocked(api.get)

const proyecto: ProyectoListado = {
  id: 1,
  nombre: 'Proyecto Alfa',
  fecha_inicio: '2026-01-01',
  fecha_fin: null,
  estado: 'Activo',
  cantidad_usuarios: 2,
  creado_en: '2026-01-01T00:00:00Z',
  actualizado_en: '2026-01-01T00:00:00Z',
}

describe('listarProyectos', () => {
  beforeEach(() => {
    getMock.mockReset()
  })

  it('consulta GET /proyectos y devuelve el arreglo de proyectos', async () => {
    getMock.mockResolvedValue({ total: 1, proyectos: [proyecto] })

    const proyectos = await listarProyectos()

    expect(getMock).toHaveBeenCalledWith('/proyectos', { signal: undefined })
    expect(proyectos).toEqual([proyecto])
  })

  it('devuelve un arreglo vacío cuando no hay proyectos', async () => {
    getMock.mockResolvedValue({ total: 0, proyectos: [] })

    expect(await listarProyectos()).toEqual([])
  })

  it('tolera una respuesta sin el campo proyectos', async () => {
    getMock.mockResolvedValue({ total: 0 })

    expect(await listarProyectos()).toEqual([])
  })
})
