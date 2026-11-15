import { beforeEach, describe, expect, it, vi } from 'vitest'
import { api } from './client'
import { ApiError } from './errors'
import { abrirNuevaRondaSesionPoker, revelarSesionPoker } from './poker'
import type { DetalleSesionPoker } from './types'

vi.mock('./client', () => ({
  api: {
    get: vi.fn(),
    post: vi.fn(),
  },
}))

const postMock = vi.mocked(api.post)

const detalle: DetalleSesionPoker = {
  id: 7,
  historia_id: 10,
  estado: 'Revelada',
  ronda_actual: 1,
  creado_por: { id: 2, nombre: 'Ana' },
  creado_en: '2026-11-13T10:00:00Z',
  cerrado_en: null,
  cancelado_en: null,
  puntos_acordados: null,
  total_miembros: 3,
  votos_registrados: 3,
  ya_vote: true,
  mi_voto: 1,
  votos: [
    { usuario_id: 2, usuario_nombre: 'Ana', puntos: 1, ronda: 1 },
    { usuario_id: 3, usuario_nombre: 'Ben', puntos: 3, ronda: 1 },
    { usuario_id: 4, usuario_nombre: 'Cris', puntos: 8, ronda: 1 },
  ],
  dispersion: {
    promedio: 4,
    mediana: 3,
    minimo: 1,
    maximo: 8,
    rango: 7,
    desviacion_estandar: 2.9439202887759492,
    hay_dispersion: true,
    total_votos: 3,
    votos_por_valor: { '1': 1, '3': 1, '8': 1 },
  },
  historial_rondas: [],
}

describe('revelarSesionPoker', () => {
  beforeEach(() => {
    postMock.mockReset()
  })

  it('revela la sesión sin enviar cuerpo y devuelve el detalle', async () => {
    postMock.mockResolvedValue(detalle)

    const respuesta = await revelarSesionPoker(1, 7)

    expect(postMock).toHaveBeenCalledWith('/proyectos/1/poker-sessions/7/revelar')
    expect(respuesta).toEqual(detalle)
  })

  it('propaga SESION_POKER_SIN_VOTOS', async () => {
    const error = new ApiError({
      status: 422,
      code: 'SESION_POKER_SIN_VOTOS',
      message: 'la sesión no tiene votos',
    })
    postMock.mockRejectedValue(error)

    await expect(revelarSesionPoker(1, 7)).rejects.toBe(error)
  })
})

describe('abrirNuevaRondaSesionPoker', () => {
  beforeEach(() => {
    postMock.mockReset()
  })

  it('abre una nueva ronda sin enviar cuerpo y devuelve el detalle actualizado', async () => {
    const detalleNuevaRonda: DetalleSesionPoker = {
      ...detalle,
      estado: 'Votando',
      ronda_actual: 2,
      votos_registrados: 0,
      ya_vote: false,
      mi_voto: null,
      votos: [],
      historial_rondas: [
        {
          numero: 1,
          estado: 'Revelada',
          revelada_en: '2026-11-13T10:05:00Z',
          votos: detalle.votos,
          dispersion: detalle.dispersion,
        },
      ],
    }
    postMock.mockResolvedValue(detalleNuevaRonda)

    const respuesta = await abrirNuevaRondaSesionPoker(1, 7)

    expect(postMock).toHaveBeenCalledWith('/proyectos/1/poker-sessions/7/rondas')
    expect(respuesta).toEqual(detalleNuevaRonda)
  })

  it('propaga SESION_POKER_NO_REVELADA', async () => {
    const error = new ApiError({
      status: 409,
      code: 'SESION_POKER_NO_REVELADA',
      message: 'la sesión no está revelada',
    })
    postMock.mockRejectedValue(error)

    await expect(abrirNuevaRondaSesionPoker(1, 7)).rejects.toBe(error)
  })
})
