import { beforeEach, describe, expect, it, vi } from 'vitest'
import { api } from './client'
import { ApiError } from './errors'
import { acordarSesionPoker, cancelarSesionPoker } from './poker'
import type { DetalleSesionPoker } from './types'

vi.mock('./client', () => ({
  api: {
    get: vi.fn(),
    post: vi.fn(),
  },
}))

const postMock = vi.mocked(api.post)

const detalleCerrada: DetalleSesionPoker = {
  id: 7,
  historia_id: 10,
  estado: 'Cerrada',
  ronda_actual: 1,
  creado_por: { id: 2, nombre: 'Ana' },
  creado_en: '2026-11-13T10:00:00Z',
  cerrado_en: '2026-11-13T11:00:00Z',
  cancelado_en: null,
  puntos_acordados: 5,
  story_estimate: {
    id: 12,
    historia_id: 10,
    sesion_id: 7,
    ronda_id: 1,
    puntos: 5,
    fuente: 'planning_poker',
    estimado_por: 2,
    estimado_en: '2026-11-13T11:00:00Z',
  },
  total_miembros: 3,
  votos_registrados: 3,
  ya_vote: true,
  mi_voto: 3,
  votos: [],
  historial_rondas: [],
}

const detalleCancelada: DetalleSesionPoker = {
  ...detalleCerrada,
  estado: 'Cancelada',
  cerrado_en: null,
  cancelado_en: '2026-11-13T11:30:00Z',
  puntos_acordados: null,
  story_estimate: null,
}

describe('acordarSesionPoker', () => {
  beforeEach(() => {
    postMock.mockReset()
  })

  it('envía los puntos acordados y devuelve el detalle cerrado', async () => {
    postMock.mockResolvedValue(detalleCerrada)

    const respuesta = await acordarSesionPoker(1, 7, 5)

    expect(postMock).toHaveBeenCalledWith('/proyectos/1/poker-sessions/7/acuerdo', { puntos: 5 })
    expect(respuesta).toEqual(detalleCerrada)
  })

  it('propaga PUNTOS_POKER_INVALIDOS', async () => {
    const error = new ApiError({
      status: 422,
      code: 'PUNTOS_POKER_INVALIDOS',
      message: 'los puntos deben pertenecer a la escala',
      campo: 'puntos',
    })
    postMock.mockRejectedValue(error)

    await expect(acordarSesionPoker(1, 7, 4)).rejects.toBe(error)
  })
})

describe('cancelarSesionPoker', () => {
  beforeEach(() => {
    postMock.mockReset()
  })

  it('cancela la sesión sin enviar cuerpo y devuelve el detalle cancelado', async () => {
    postMock.mockResolvedValue(detalleCancelada)

    const respuesta = await cancelarSesionPoker(1, 7)

    expect(postMock).toHaveBeenCalledWith('/proyectos/1/poker-sessions/7/cancelar')
    expect(respuesta).toEqual(detalleCancelada)
  })

  it('propaga SESION_POKER_CERRADA', async () => {
    const error = new ApiError({
      status: 409,
      code: 'SESION_POKER_CERRADA',
      message: 'la sesión ya está cerrada',
    })
    postMock.mockRejectedValue(error)

    await expect(cancelarSesionPoker(1, 7)).rejects.toBe(error)
  })
})
