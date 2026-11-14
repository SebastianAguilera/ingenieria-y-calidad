import { beforeEach, describe, expect, it, vi } from 'vitest'
import { api } from './client'
import {
  crearSesionPoker,
  listarSesionesActivas,
  listarSesionesPoker,
} from './poker'
import type { SesionPokerActiva, SesionPokerHistorial } from './types'

vi.mock('./client', () => ({
  api: {
    get: vi.fn(),
    post: vi.fn(),
  },
}))

const getMock = vi.mocked(api.get)
const postMock = vi.mocked(api.post)

const sesion: SesionPokerHistorial = {
  id: 3,
  estado: 'Cerrada',
  ronda_actual: 2,
  puntos_acordados: 5,
  creado_por: { id: 2, nombre: 'Ana' },
  creado_en: '2026-11-10T09:00:00Z',
  cerrado_en: '2026-11-10T10:00:00Z',
}

const activa: SesionPokerActiva = {
  sesion_id: 7,
  estado: 'Votando',
  ronda_actual: 1,
  creado_por: { id: 2, nombre: 'Ana' },
  creado_en: '2026-11-13T10:00:00Z',
  votos_registrados: 1,
  total_miembros: 3,
  ya_vote: true,
  historia: { id: 10, codigo: 'US-01', titulo: 'Registrar horas reales' },
}

describe('listarSesionesPoker', () => {
  beforeEach(() => {
    getMock.mockReset()
  })

  it('consulta el historial de la historia y devuelve el arreglo', async () => {
    getMock.mockResolvedValue([sesion])

    const sesiones = await listarSesionesPoker(1, 10)

    expect(getMock).toHaveBeenCalledWith('/proyectos/1/historias/10/poker-sessions', {
      signal: undefined,
    })
    expect(sesiones).toEqual([sesion])
  })

  it('reenvía la señal de cancelación', async () => {
    getMock.mockResolvedValue([sesion])
    const controlador = new AbortController()

    await listarSesionesPoker(1, 10, controlador.signal)

    expect(getMock).toHaveBeenCalledWith('/proyectos/1/historias/10/poker-sessions', {
      signal: controlador.signal,
    })
  })

  it('tolera una respuesta nula como historial vacío', async () => {
    getMock.mockResolvedValue(null)

    expect(await listarSesionesPoker(1, 10)).toEqual([])
  })
})

describe('crearSesionPoker', () => {
  beforeEach(() => {
    postMock.mockReset()
  })

  it('envía POST sin cuerpo y devuelve la sesión creada', async () => {
    postMock.mockResolvedValue({
      id: 7,
      historia_id: 10,
      estado: 'Votando',
      ronda_actual: 1,
      creado_por: { id: 2, nombre: 'Ana' },
      creado_en: '2026-11-13T10:00:00Z',
      puntos_acordados: null,
    })

    const creada = await crearSesionPoker(1, 10)

    expect(postMock).toHaveBeenCalledWith('/proyectos/1/historias/10/poker-sessions')
    expect(creada.id).toBe(7)
    expect(creada.estado).toBe('Votando')
  })

  it('propaga el error del backend', async () => {
    postMock.mockRejectedValue(new Error('409'))

    await expect(crearSesionPoker(1, 10)).rejects.toThrow('409')
  })
})

describe('listarSesionesActivas', () => {
  beforeEach(() => {
    getMock.mockReset()
  })

  it('consulta las sesiones activas del proyecto', async () => {
    getMock.mockResolvedValue([activa])

    const activas = await listarSesionesActivas(1)

    expect(getMock).toHaveBeenCalledWith('/proyectos/1/poker-sessions/activas', {
      signal: undefined,
    })
    expect(activas).toEqual([activa])
  })

  it('reenvía la señal de cancelación', async () => {
    getMock.mockResolvedValue([activa])
    const controlador = new AbortController()

    await listarSesionesActivas(1, controlador.signal)

    expect(getMock).toHaveBeenCalledWith('/proyectos/1/poker-sessions/activas', {
      signal: controlador.signal,
    })
  })

  it('tolera una respuesta nula como lista vacía', async () => {
    getMock.mockResolvedValue(null)

    expect(await listarSesionesActivas(1)).toEqual([])
  })
})
