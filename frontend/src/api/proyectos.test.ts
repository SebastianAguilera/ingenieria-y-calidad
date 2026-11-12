import { beforeEach, describe, expect, it, vi } from 'vitest'
import { api } from './client'
import {
  actualizarProyecto,
  asociarUsuario,
  cambiarEstadoProyecto,
  crearProyecto,
  eliminarProyecto,
  listarProyectos,
  obtenerProyecto,
} from './proyectos'
import type { ProyectoListado } from './types'

vi.mock('./client', () => ({
  api: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
  },
}))

const getMock = vi.mocked(api.get)
const postMock = vi.mocked(api.post)
const putMock = vi.mocked(api.put)
const patchMock = vi.mocked(api.patch)
const deleteMock = vi.mocked(api.delete)

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
  descripcion: '',
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

describe('obtenerProyecto', () => {
  beforeEach(() => {
    getMock.mockReset()
  })

  it('consulta GET /proyectos/:id y devuelve el detalle', async () => {
    const detalle = { id: 1, nombre: 'Alfa', descripcion: 'Descripción' }
    getMock.mockResolvedValue(detalle)

    const resultado = await obtenerProyecto(1)

    expect(getMock).toHaveBeenCalledWith('/proyectos/1', { signal: undefined })
    expect(resultado).toBe(detalle)
  })
})

describe('actualizarProyecto', () => {
  beforeEach(() => {
    putMock.mockReset()
  })

  it('envía PUT /proyectos/:id con los datos y devuelve el proyecto', async () => {
    const actualizado = { id: 1, nombre: 'Nuevo', descripcion: 'Nueva' }
    putMock.mockResolvedValue(actualizado)
    const datos = {
      nombre: 'Nuevo',
      descripcion: 'Nueva',
      fecha_inicio: '2026-01-01',
      fecha_fin: null,
    }

    const resultado = await actualizarProyecto(1, datos)

    expect(putMock).toHaveBeenCalledWith('/proyectos/1', datos)
    expect(resultado).toBe(actualizado)
  })
})

describe('cambiarEstadoProyecto', () => {
  beforeEach(() => {
    patchMock.mockReset()
  })

  it('envía PATCH /proyectos/:id/estado con el estado', async () => {
    patchMock.mockResolvedValue({ id: 1, estado: 'Cerrado' })

    const resultado = await cambiarEstadoProyecto(1, 'Cerrado')

    expect(patchMock).toHaveBeenCalledWith('/proyectos/1/estado', { estado: 'Cerrado' })
    expect(resultado).toEqual({ id: 1, estado: 'Cerrado' })
  })
})

describe('eliminarProyecto', () => {
  beforeEach(() => {
    deleteMock.mockReset()
  })

  it('envía DELETE /proyectos/:id', async () => {
    deleteMock.mockResolvedValue(undefined)

    await eliminarProyecto(1)

    expect(deleteMock).toHaveBeenCalledWith('/proyectos/1')
  })
})
