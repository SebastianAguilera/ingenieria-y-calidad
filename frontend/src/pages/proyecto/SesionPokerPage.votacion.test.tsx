import { act, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiError } from '../../api/errors'
import { listarHistorias } from '../../api/historias'
import {
  abrirNuevaRondaSesionPoker,
  crearSesionPoker,
  listarSesionesPoker,
  obtenerDetalleSesionPoker,
  revelarSesionPoker,
  votarSesionPoker,
} from '../../api/poker'
import type {
  DetalleSesionPoker,
  Historia,
  ResultadoVotoPoker,
  SesionPokerCreada,
  SesionPokerHistorial,
} from '../../api/types'
import SesionPokerPage from './SesionPokerPage'

vi.mock('../../api/historias', () => ({
  listarHistorias: vi.fn(),
}))

vi.mock('../../api/poker', () => ({
  listarSesionesPoker: vi.fn(),
  crearSesionPoker: vi.fn(),
  obtenerDetalleSesionPoker: vi.fn(),
  votarSesionPoker: vi.fn(),
  revelarSesionPoker: vi.fn(),
  abrirNuevaRondaSesionPoker: vi.fn(),
}))

const listarHistoriasMock = vi.mocked(listarHistorias)
const listarSesionesMock = vi.mocked(listarSesionesPoker)
const crearMock = vi.mocked(crearSesionPoker)
const obtenerDetalleMock = vi.mocked(obtenerDetalleSesionPoker)
const votarMock = vi.mocked(votarSesionPoker)
const revelarMock = vi.mocked(revelarSesionPoker)
const nuevaRondaMock = vi.mocked(abrirNuevaRondaSesionPoker)

const historia: Historia = {
  id: 10,
  codigo: 'US-01',
  titulo: 'Registrar horas reales',
  descripcion: 'Permitir registrar horas por usuario.',
  prioridad: 'Alta',
  estado: 'Backlog',
  story_points: null,
  esfuerzo_real: null,
  horas_estimadas: null,
  creado_en: '2026-10-01T09:15:00Z',
  actualizado_en: '2026-10-01T09:15:00Z',
}

const sesionVotando: SesionPokerHistorial = {
  id: 7,
  estado: 'Votando',
  ronda_actual: 1,
  puntos_acordados: null,
  creado_por: { id: 2, nombre: 'Ana' },
  creado_en: '2026-11-13T10:00:00Z',
}

const sesionRevelada: SesionPokerHistorial = {
  ...sesionVotando,
  estado: 'Revelada',
}

const sesionCerrada: SesionPokerHistorial = {
  ...sesionVotando,
  id: 9,
  estado: 'Cerrada',
  ronda_actual: 2,
  puntos_acordados: 5,
  cerrado_en: '2026-11-13T11:00:00Z',
}

const sesionCreada: SesionPokerCreada = {
  id: 7,
  historia_id: 10,
  estado: 'Votando',
  ronda_actual: 1,
  creado_por: { id: 2, nombre: 'Ana' },
  creado_en: '2026-11-13T10:00:00Z',
  puntos_acordados: null,
}

const detalleVotando: DetalleSesionPoker = {
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

const detalleRevelada: DetalleSesionPoker = {
  ...detalleVotando,
  estado: 'Revelada',
  votos: [
    { usuario_id: 3, usuario_nombre: 'Ben', puntos: 13, ronda: 1 },
    { usuario_id: 2, usuario_nombre: 'Ana', puntos: 5, ronda: 1 },
  ],
}

const detalleCancelada: DetalleSesionPoker = {
  ...detalleVotando,
  estado: 'Cancelada',
  cancelado_en: '2026-11-13T10:05:00Z',
}

const resultadoVoto: ResultadoVotoPoker = {
  sesion_id: 7,
  mi_voto: 5,
  votado_en: '2026-11-13T10:01:00Z',
  votos_registrados: 1,
  total_miembros: 3,
}

function renderizar() {
  return render(
    <MemoryRouter initialEntries={['/proyectos/1/backlog/10/poker']}>
      <Routes>
        <Route path="/proyectos/:id/backlog/:historiaId/poker" element={<SesionPokerPage />} />
        <Route path="/proyectos/:id/backlog" element={<div>Backlog del proyecto</div>} />
      </Routes>
    </MemoryRouter>,
  )
}

function prepararCargaInicial() {
  listarHistoriasMock.mockResolvedValue([historia])
  listarSesionesMock.mockResolvedValue([sesionVotando])
  obtenerDetalleMock.mockResolvedValue(detalleVotando)
}

describe('SesionPokerPage - votación', () => {
  beforeEach(() => {
    listarHistoriasMock.mockReset()
    listarSesionesMock.mockReset()
    crearMock.mockReset()
    obtenerDetalleMock.mockReset()
    votarMock.mockReset()
    revelarMock.mockReset()
    nuevaRondaMock.mockReset()
    prepararCargaInicial()
  })

  it('muestra las cartas, el voto propio y el conteo sin votos ajenos', async () => {
    obtenerDetalleMock.mockResolvedValue({
      ...detalleVotando,
      votos_registrados: 2,
      votos: [{ usuario_id: 3, usuario_nombre: 'Ben', puntos: 13, ronda: 1 }],
    })

    renderizar()

    expect(await screen.findByRole('radio', { name: '5 puntos' })).toBeChecked()
    expect(screen.getByText('2 de 3 votaron')).toBeInTheDocument()
    expect(screen.getByText('Tu voto: 5')).toBeInTheDocument()
    expect(screen.getByText('2 de 3 votaron').closest('[aria-live="polite"]')).not.toBeNull()
    expect(screen.queryByText('Ben')).not.toBeInTheDocument()
    expect(screen.queryByText('Voto de Ben')).not.toBeInTheDocument()
  })

  it('envía el primer voto al elegir una carta y actualiza la selección', async () => {
    obtenerDetalleMock.mockResolvedValue({
      ...detalleVotando,
      votos_registrados: 0,
      ya_vote: false,
      mi_voto: null,
    })
    votarMock.mockResolvedValue(resultadoVoto)
    const usuario = userEvent.setup()

    renderizar()

    await usuario.click(await screen.findByRole('radio', { name: '5 puntos' }))

    expect(votarMock).toHaveBeenCalledWith(1, 7, 5)
    await waitFor(() => {
      expect(screen.getByRole('radio', { name: '5 puntos' })).toBeChecked()
    })
    expect(screen.getByText('Tu voto: 5')).toBeInTheDocument()
    expect(screen.getByText('1 de 3 votaron')).toBeInTheDocument()
  })

  it('cambia el voto sin duplicar y conserva el conteo', async () => {
    votarMock.mockResolvedValue({ ...resultadoVoto, mi_voto: 8 })
    const usuario = userEvent.setup()

    renderizar()

    await usuario.click(await screen.findByRole('radio', { name: '8 puntos' }))

    expect(votarMock).toHaveBeenCalledWith(1, 7, 8)
    await waitFor(() => {
      expect(screen.getByRole('radio', { name: '8 puntos' })).toBeChecked()
    })
    expect(screen.getByRole('radio', { name: '5 puntos' })).not.toBeChecked()
    expect(screen.getByText('Tu voto: 8')).toBeInTheDocument()
    expect(screen.getByText('1 de 3 votaron')).toBeInTheDocument()
  })

  it('con una sesión Revelada deshabilita las cartas y muestra los votos ajenos', async () => {
    listarSesionesMock.mockResolvedValue([sesionRevelada])
    obtenerDetalleMock.mockResolvedValue(detalleRevelada)

    renderizar()

    const carta = await screen.findByRole('radio', { name: '5 puntos' })

    expect(carta).toBeDisabled()
    expect(screen.getByText('Los votos fueron revelados. La sesión sigue activa.')).toBeInTheDocument()
    expect(screen.getByText('Ben: 13 puntos')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Nueva ronda' })).toBeInTheDocument()
  })

  it('con la última sesión Cerrada mantiene el resultado de US-F15 sin cartas', async () => {
    listarSesionesMock.mockResolvedValue([sesionCerrada])

    renderizar()

    expect(await screen.findByText('Puntos acordados: 5')).toBeInTheDocument()
    expect(screen.queryByRole('radio')).not.toBeInTheDocument()
    expect(obtenerDetalleMock).not.toHaveBeenCalled()
  })

  it('ante 409 SESION_POKER_NO_VOTANDO refresca el detalle y muestra el cierre', async () => {
    obtenerDetalleMock
      .mockResolvedValueOnce({
        ...detalleVotando,
        votos_registrados: 0,
        ya_vote: false,
        mi_voto: null,
      })
      .mockResolvedValueOnce(detalleRevelada)
    votarMock.mockRejectedValue(
      new ApiError({
        status: 409,
        code: 'SESION_POKER_NO_VOTANDO',
        message: 'la sesión ya no está votando',
      }),
    )
    const usuario = userEvent.setup()

    renderizar()

    await usuario.click(await screen.findByRole('radio', { name: '5 puntos' }))

    expect(await screen.findByText('La votación ya cerró')).toBeInTheDocument()
    expect(obtenerDetalleMock).toHaveBeenCalledTimes(2)
    expect(await screen.findByRole('radio', { name: '5 puntos' })).toBeDisabled()
    expect(screen.getByText('Ben: 13 puntos')).toBeInTheDocument()
  })

  it('ante un error de red al votar conserva el voto anterior y permite reintentar', async () => {
    votarMock
      .mockRejectedValueOnce(
        new ApiError({ status: 0, code: 'NETWORK_ERROR', message: 'sin conexión' }),
      )
      .mockResolvedValueOnce({ ...resultadoVoto, mi_voto: 8 })
    const usuario = userEvent.setup()

    renderizar()

    await usuario.click(await screen.findByRole('radio', { name: '8 puntos' }))

    expect(
      await screen.findByText('No se pudo registrar el voto. Intentá nuevamente.'),
    ).toBeInTheDocument()
    expect(screen.getByRole('radio', { name: '5 puntos' })).toBeChecked()
    expect(screen.getByRole('radio', { name: '8 puntos' })).not.toBeChecked()

    await usuario.click(screen.getByRole('radio', { name: '8 puntos' }))

    expect(votarMock).toHaveBeenCalledTimes(2)
    await waitFor(() => {
      expect(screen.getByRole('radio', { name: '8 puntos' })).toBeChecked()
    })
  })

  it('muestra un error de carga del detalle y permite reintentar', async () => {
    obtenerDetalleMock
      .mockRejectedValueOnce(
        new ApiError({ status: 0, code: 'NETWORK_ERROR', message: 'sin conexión' }),
      )
      .mockResolvedValueOnce(detalleVotando)
    const usuario = userEvent.setup()

    renderizar()

    expect(
      await screen.findByText('No se pudo cargar el Planning Poker. Intentá nuevamente.'),
    ).toBeInTheDocument()

    await usuario.click(screen.getByRole('button', { name: 'Reintentar' }))

    expect(await screen.findByRole('radio', { name: '5 puntos' })).toBeChecked()
    expect(obtenerDetalleMock).toHaveBeenCalledTimes(2)
  })

  it('al crear una sesión obtiene el detalle y habilita las cartas', async () => {
    listarSesionesMock.mockResolvedValue([])
    crearMock.mockResolvedValue(sesionCreada)
    obtenerDetalleMock.mockResolvedValue({
      ...detalleVotando,
      votos_registrados: 0,
      ya_vote: false,
      mi_voto: null,
    })
    const usuario = userEvent.setup()

    renderizar()

    await usuario.click(await screen.findByRole('button', { name: 'Iniciar Planning Poker' }))

    expect(crearMock).toHaveBeenCalledWith(1, 10)
    expect(obtenerDetalleMock.mock.calls[0]?.slice(0, 2)).toEqual([1, 7])
    expect(await screen.findByRole('radio', { name: '0 puntos' })).toBeEnabled()
    expect(screen.getByText('0 de 3 votaron')).toBeInTheDocument()
  })
})

describe('SesionPokerPage - polling', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    listarHistoriasMock.mockReset()
    listarSesionesMock.mockReset()
    crearMock.mockReset()
    obtenerDetalleMock.mockReset()
    votarMock.mockReset()
    revelarMock.mockReset()
    nuevaRondaMock.mockReset()
    prepararCargaInicial()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('actualiza el conteo cada 3 segundos', async () => {
    obtenerDetalleMock
      .mockResolvedValueOnce(detalleVotando)
      .mockResolvedValueOnce({ ...detalleVotando, votos_registrados: 2 })

    renderizar()

    await act(async () => {
      await vi.advanceTimersByTimeAsync(0)
    })

    expect(screen.getByText('1 de 3 votaron')).toBeInTheDocument()

    await act(async () => {
      await vi.advanceTimersByTimeAsync(3000)
    })

    expect(obtenerDetalleMock).toHaveBeenCalledTimes(2)
    expect(screen.getByText('2 de 3 votaron')).toBeInTheDocument()
  })

  it('detecta el paso a Revelada, muestra los votos y mantiene el polling', async () => {
    obtenerDetalleMock
      .mockResolvedValueOnce(detalleVotando)
      .mockResolvedValueOnce(detalleRevelada)

    renderizar()

    await act(async () => {
      await vi.advanceTimersByTimeAsync(0)
    })

    expect(screen.getByRole('radio', { name: '5 puntos' })).toBeEnabled()

    await act(async () => {
      await vi.advanceTimersByTimeAsync(3000)
    })

    expect(screen.getByRole('radio', { name: '5 puntos' })).toBeDisabled()
    expect(screen.getByText('Los votos fueron revelados. La sesión sigue activa.')).toBeInTheDocument()
    expect(screen.getByText('Ben: 13 puntos')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Nueva ronda' })).toBeInTheDocument()

    await act(async () => {
      await vi.advanceTimersByTimeAsync(3000)
    })

    expect(obtenerDetalleMock).toHaveBeenCalledTimes(3)
  })

  it('detecta el paso a Cancelada y deshabilita las cartas', async () => {
    obtenerDetalleMock
      .mockResolvedValueOnce(detalleVotando)
      .mockResolvedValueOnce(detalleCancelada)

    renderizar()

    await act(async () => {
      await vi.advanceTimersByTimeAsync(0)
    })

    await act(async () => {
      await vi.advanceTimersByTimeAsync(3000)
    })

    expect(screen.getByText('Cancelada')).toBeInTheDocument()
    expect(screen.getByRole('radio', { name: '5 puntos' })).toBeDisabled()
  })

  it('ante un error de polling conserva los datos y muestra un aviso no bloqueante', async () => {
    obtenerDetalleMock
      .mockResolvedValueOnce(detalleVotando)
      .mockRejectedValueOnce(
        new ApiError({ status: 0, code: 'NETWORK_ERROR', message: 'sin conexión' }),
      )
      .mockResolvedValueOnce({ ...detalleVotando, votos_registrados: 2 })

    renderizar()

    await act(async () => {
      await vi.advanceTimersByTimeAsync(0)
    })

    expect(screen.getByText('1 de 3 votaron')).toBeInTheDocument()

    await act(async () => {
      await vi.advanceTimersByTimeAsync(3000)
    })

    expect(
      screen.getByText('No se pudo actualizar la sesión. Reintentando…'),
    ).toBeInTheDocument()
    expect(screen.getByText('1 de 3 votaron')).toBeInTheDocument()

    await act(async () => {
      await vi.advanceTimersByTimeAsync(3000)
    })

    expect(screen.getByText('2 de 3 votaron')).toBeInTheDocument()
    expect(
      screen.queryByText('No se pudo actualizar la sesión. Reintentando…'),
    ).not.toBeInTheDocument()
  })
})
