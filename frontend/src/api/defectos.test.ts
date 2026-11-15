import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { Defecto } from './types'
import { api } from './client'
import { actualizarDefecto, crearDefecto, listarDefectos } from './defectos'

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

const defecto: Defecto = {
  id: 7,
  historia_id: 1,
  sprint_deteccion_id: 2,
  sprint_resolucion_id: null,
  reportado_por: null,
  descripcion: 'Error en validación de fechas',
  severidad: 'major',
  estado: 'open',
  creado_en: '2026-05-02T10:00:00Z',
  resuelto_en: null,
}

describe('listarDefectos', () => {
  beforeEach(() => {
    getMock.mockReset()
  })

  it('consulta GET /proyectos/:id/defectos y devuelve el arreglo', async () => {
    getMock.mockResolvedValue({ total: 1, defectos: [defecto] })

    const defectos = await listarDefectos(1)

    expect(getMock).toHaveBeenCalledWith('/proyectos/1/defectos', { signal: undefined })
    expect(defectos).toEqual([defecto])
  })

  it('pasa la señal de aborto al cliente', async () => {
    getMock.mockResolvedValue({ total: 0, defectos: [] })
    const controlador = new AbortController()

    await listarDefectos(1, controlador.signal)

    expect(getMock).toHaveBeenCalledWith('/proyectos/1/defectos', {
      signal: controlador.signal,
    })
  })

  it('tolera una respuesta sin el campo defectos', async () => {
    getMock.mockResolvedValue({ total: 0 })

    expect(await listarDefectos(1)).toEqual([])
  })

  it('tolera una respuesta con defectos nulo', async () => {
    getMock.mockResolvedValue({ total: 0, defectos: null })

    expect(await listarDefectos(1)).toEqual([])
  })
})

describe('crearDefecto', () => {
  beforeEach(() => {
    postMock.mockReset()
  })

  it('envía POST /proyectos/:id/defectos con el alta y devuelve el defecto', async () => {
    postMock.mockResolvedValue(defecto)
    const alta = {
      descripcion: 'Error en validación de fechas',
      severidad: 'major' as const,
      historia_id: 1,
      sprint_deteccion_id: 2,
    }

    const creado = await crearDefecto(1, alta)

    expect(postMock).toHaveBeenCalledWith('/proyectos/1/defectos', alta)
    expect(creado).toEqual(defecto)
  })
})

describe('actualizarDefecto', () => {
  beforeEach(() => {
    patchMock.mockReset()
  })

  it('envía PATCH /proyectos/:id/defectos/:defectoId con el cambio de estado', async () => {
    const resuelto: Defecto = { ...defecto, estado: 'resolved', sprint_resolucion_id: 3 }
    patchMock.mockResolvedValue(resuelto)

    const actualizado = await actualizarDefecto(1, 7, {
      estado: 'resolved',
      sprint_resolucion_id: 3,
    })

    expect(patchMock).toHaveBeenCalledWith('/proyectos/1/defectos/7', {
      estado: 'resolved',
      sprint_resolucion_id: 3,
    })
    expect(actualizado).toEqual(resuelto)
  })
})
