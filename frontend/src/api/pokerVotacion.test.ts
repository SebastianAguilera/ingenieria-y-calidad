import { beforeEach, describe, expect, it, vi } from 'vitest'
import { api } from './client'
import { ApiError } from './errors'
import { obtenerDetalleSesionPoker, votarSesionPoker } from './poker'
import type { DetalleSesionPoker, ResultadoVotoPoker } from './types'

vi.mock('./client', () => ({
  api: {
    get: vi.fn(),
    post: vi.fn(),
  },
}))

const getMock = vi.mocked(api.get)
const postMock = vi.mocked(api.post)

const detalle: DetalleSesionPoker = {
  id: 7,
  historia_id: 10,
  estado: 'Votando',
  ronda_actual: 1,
  creado_por: { id: 2, nombre: 'Ana' },
  creado_en: '2026-11-13T10:00:00Z',
  cerrado_en: null,
  cancelado_en: null,
  puntos_acordados: null,
  total_miembros: 3,
  votos_registrados: 1,
  ya_vote: true,
  mi_voto: 5,
  votos: [],
  historial_rondas: [],
}

const resultadoVoto: ResultadoVotoPoker = {
  sesion_id: 7,
  mi_voto: 8,
  votado_en: '2026-11-13T10:01:00Z',
  votos_registrados: 1,
  total_miembros: 3,
}

describe('obtenerDetalleSesionPoker', () => {
  beforeEach(() => {
    getMock.mockReset()
  })

  it('consulta el detalle de la sesión y devuelve voto propio y conteos', async () => {
    getMock.mockResolvedValue(detalle)

    const respuesta = await obtenerDetalleSesionPoker(1, 7)

    expect(getMock).toHaveBeenCalledWith('/proyectos/1/poker-sessions/7', {
      signal: undefined,
    })
    expect(respuesta).toEqual(detalle)
  })

  it('reenvía la señal de cancelación', async () => {
    getMock.mockResolvedValue(detalle)
    const controlador = new AbortController()

    await obtenerDetalleSesionPoker(1, 7, controlador.signal)

    expect(getMock).toHaveBeenCalledWith('/proyectos/1/poker-sessions/7', {
      signal: controlador.signal,
    })
  })

  it('propaga el error del backend', async () => {
    const error = new ApiError({
      status: 404,
      code: 'SESION_POKER_NO_ENCONTRADA',
      message: 'la sesión no existe',
    })
    getMock.mockRejectedValue(error)

    await expect(obtenerDetalleSesionPoker(1, 7)).rejects.toBe(error)
  })
})

describe('votarSesionPoker', () => {
  beforeEach(() => {
    postMock.mockReset()
  })

  it('envía los puntos al endpoint de votos y devuelve el voto registrado', async () => {
    postMock.mockResolvedValue(resultadoVoto)

    const respuesta = await votarSesionPoker(1, 7, 8)

    expect(postMock).toHaveBeenCalledWith('/proyectos/1/poker-sessions/7/votos', {
      puntos: 8,
    })
    expect(respuesta).toEqual(resultadoVoto)
  })

  it('acepta los extremos de la escala', async () => {
    postMock.mockResolvedValue({ ...resultadoVoto, mi_voto: 0 })

    await votarSesionPoker(1, 7, 0)

    expect(postMock).toHaveBeenCalledWith('/proyectos/1/poker-sessions/7/votos', {
      puntos: 0,
    })
  })

  it('propaga SESION_POKER_NO_VOTANDO', async () => {
    const error = new ApiError({
      status: 409,
      code: 'SESION_POKER_NO_VOTANDO',
      message: 'la sesión ya no está votando',
    })
    postMock.mockRejectedValue(error)

    await expect(votarSesionPoker(1, 7, 5)).rejects.toBe(error)
  })
})
