import { beforeEach, describe, expect, it, vi } from 'vitest'
import { api } from './client'
import { ApiError } from './errors'
import { asociarUsuario, crearProyecto, listarProyectos, obtenerProyecto } from './proyectos'
import type { Proyecto, ProyectoListado } from './types'

vi.mock('./client', () => ({
  api: {
    get: vi.fn(),
    post: vi.fn(),
  },
}))

const getMock = vi.mocked(api.get)
const postMock = vi.mocked(api.post)

describe('crearProyecto', () => {
  beforeEach(() => {
    postMock.mockReset()
  })

  it('envía POST /proyectos con el alta y devuelve el proyecto creado', async () => {
    const creado = { id: 7, nombre: 'Alfa' }
    postMock.mockResolvedValue(creado)
    const alta = { nombre: 'Alfa', fecha_inicio: '2026-01-01', fecha_fin: null }

    const resultado = await crearProyecto(alta)

    expect(postMock).toHaveBeenCalledWith('/proyectos', alta)
    expect(resultado).toBe(creado)
  })
})

describe('asociarUsuario', () => {
  beforeEach(() => {
    postMock.mockReset()
  })

  it('envía POST /proyectos/:id/usuarios con usuario_id y rol', async () => {
    postMock.mockResolvedValue({})

    await asociarUsuario(7, 3, 'Facilitador Ágil')

    expect(postMock).toHaveBeenCalledWith('/proyectos/7/usuarios', {
      usuario_id: 3,
      rol: 'Facilitador Ágil',
    })
  })
})

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

const detalle: Proyecto = {
  id: 7,
  nombre: 'Proyecto Alfa',
  fecha_inicio: '2026-01-01',
  fecha_fin: null,
  estado: 'Activo',
  cantidad_usuarios: 2,
  usuarios: [
    { id: 3, nombre: 'Ana Pérez', email: 'ana@ejemplo.com', creado_en: '2026-01-01T00:00:00Z' },
  ],
  creado_en: '2026-01-01T00:00:00Z',
  actualizado_en: '2026-01-01T00:00:00Z',
}

describe('obtenerProyecto', () => {
  beforeEach(() => {
    getMock.mockReset()
  })

  it('consulta GET /proyectos/:id y devuelve el detalle del proyecto', async () => {
    getMock.mockResolvedValue(detalle)

    const resultado = await obtenerProyecto(7)

    expect(getMock).toHaveBeenCalledWith('/proyectos/7', { signal: undefined })
    expect(resultado).toBe(detalle)
  })

  it('propaga el signal al cliente tipado', async () => {
    const controlador = new AbortController()
    getMock.mockResolvedValue(detalle)

    await obtenerProyecto(7, controlador.signal)

    expect(getMock).toHaveBeenCalledWith('/proyectos/7', { signal: controlador.signal })
  })

  it('propaga el ApiError de la API sin transformarlo', async () => {
    const error = new ApiError({
      status: 404,
      code: 'PROYECTO_NO_ENCONTRADO',
      message: 'el proyecto no existe',
    })
    getMock.mockRejectedValue(error)

    await expect(obtenerProyecto(999)).rejects.toBe(error)
  })
})
