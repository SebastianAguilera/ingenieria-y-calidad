import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiError } from '../../api/errors'
import { listarHistorias } from '../../api/historias'
import {
  abrirNuevaRondaSesionPoker,
  listarSesionesPoker,
  obtenerDetalleSesionPoker,
  revelarSesionPoker,
} from '../../api/poker'
import type {
  DetalleSesionPoker,
  DispersionPoker,
  Historia,
  HistorialRondaPoker,
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
const obtenerDetalleMock = vi.mocked(obtenerDetalleSesionPoker)
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

const sesionCancelada: SesionPokerHistorial = {
  ...sesionVotando,
  id: 11,
  estado: 'Cancelada',
  cerrado_en: '2026-11-13T11:30:00Z',
}

const dispersion: DispersionPoker = {
  promedio: 4,
  mediana: 3,
  minimo: 1,
  maximo: 8,
  rango: 7,
  desviacion_estandar: 2.9439202887759492,
  hay_dispersion: true,
  total_votos: 3,
  votos_por_valor: { '1': 1, '3': 1, '8': 1 },
}

const votos = [
  { usuario_id: 2, usuario_nombre: 'Ana', puntos: 1, ronda: 1 },
  { usuario_id: 3, usuario_nombre: 'Ben', puntos: 3, ronda: 1 },
  { usuario_id: 4, usuario_nombre: 'Cris', puntos: 8, ronda: 1 },
]

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
  votos_registrados: 3,
  mi_voto: 1,
  votos,
  dispersion,
}

const rondaAnterior: HistorialRondaPoker = {
  numero: 1,
  estado: 'Revelada',
  revelada_en: '2026-11-13T10:05:00Z',
  votos,
  dispersion,
}

const detalleNuevaRonda: DetalleSesionPoker = {
  ...detalleVotando,
  estado: 'Votando',
  ronda_actual: 2,
  votos_registrados: 0,
  ya_vote: false,
  mi_voto: null,
  votos: [],
  historial_rondas: [rondaAnterior],
}

const detalleCerrada: DetalleSesionPoker = {
  ...detalleVotando,
  estado: 'Cerrada',
  cerrado_en: '2026-11-13T11:00:00Z',
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

function prepararCargaInicial(detalle: DetalleSesionPoker = detalleVotando) {
  listarHistoriasMock.mockResolvedValue([historia])
  listarSesionesMock.mockResolvedValue([
    detalle.estado === 'Votando' ? sesionVotando : sesionRevelada,
  ])
  obtenerDetalleMock.mockResolvedValue(detalle)
}

describe('SesionPokerPage - revelado', () => {
  beforeEach(() => {
    listarHistoriasMock.mockReset()
    listarSesionesMock.mockReset()
    obtenerDetalleMock.mockReset()
    revelarMock.mockReset()
    nuevaRondaMock.mockReset()
    prepararCargaInicial()
  })

  it('revela los votos y muestra nombres, panel, badge y nueva ronda', async () => {
    revelarMock.mockResolvedValue(detalleRevelada)
    const usuario = userEvent.setup()

    renderizar()

    await usuario.click(await screen.findByRole('button', { name: 'Revelar votos' }))

    expect(revelarMock).toHaveBeenCalledWith(1, 7)
    expect(await screen.findByText('Ana: 1 punto')).toBeInTheDocument()
    expect(screen.getByText('Ben: 3 puntos')).toBeInTheDocument()
    expect(screen.getByText('Cris: 8 puntos')).toBeInTheDocument()
    expect(screen.getByText('Mínimo')).toBeInTheDocument()
    expect(screen.getByText('Máximo')).toBeInTheDocument()
    expect(screen.getByRole('region', { name: 'Dispersión de votos' })).toBeInTheDocument()
    expect(screen.getByText('Hay diferencias')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Nueva ronda' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Revelar votos' })).not.toBeInTheDocument()
  })

  it('muestra el error de revelar sin votos y mantiene la sesión votando', async () => {
    revelarMock.mockRejectedValue(
      new ApiError({
        status: 422,
        code: 'SESION_POKER_SIN_VOTOS',
        message: 'la sesión no tiene votos',
      }),
    )
    const usuario = userEvent.setup()

    renderizar()

    await usuario.click(await screen.findByRole('button', { name: 'Revelar votos' }))

    expect(await screen.findByText('No se puede revelar sin votos.')).toBeInTheDocument()
    expect(screen.getByText('Votando')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Revelar votos' })).toBeEnabled()
  })

  it('ante un error de red al revelar permite reintentar', async () => {
    revelarMock
      .mockRejectedValueOnce(
        new ApiError({ status: 0, code: 'NETWORK_ERROR', message: 'sin conexión' }),
      )
      .mockResolvedValueOnce(detalleRevelada)
    const usuario = userEvent.setup()

    renderizar()

    await usuario.click(await screen.findByRole('button', { name: 'Revelar votos' }))

    expect(
      await screen.findByText('No se pudieron revelar los votos. Intentá nuevamente.'),
    ).toBeInTheDocument()

    await usuario.click(screen.getByRole('button', { name: 'Revelar votos' }))

    expect(await screen.findByText('Ana: 1 punto')).toBeInTheDocument()
    expect(revelarMock).toHaveBeenCalledTimes(2)
  })

  it('ante 409 refresca el detalle y muestra que la votación cerró', async () => {
    obtenerDetalleMock
      .mockResolvedValueOnce(detalleVotando)
      .mockResolvedValueOnce(detalleRevelada)
    revelarMock.mockRejectedValue(
      new ApiError({
        status: 409,
        code: 'SESION_POKER_NO_VOTANDO',
        message: 'la sesión ya no está votando',
      }),
    )
    const usuario = userEvent.setup()

    renderizar()

    await usuario.click(await screen.findByRole('button', { name: 'Revelar votos' }))

    expect(await screen.findByText('La votación ya cerró')).toBeInTheDocument()
    expect(await screen.findByText('Ana: 1 punto')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Nueva ronda' })).toBeInTheDocument()
  })

  it('abre una nueva ronda, limpia el voto y conserva el historial', async () => {
    prepararCargaInicial(detalleRevelada)
    nuevaRondaMock.mockResolvedValue(detalleNuevaRonda)
    const usuario = userEvent.setup()

    renderizar()

    await usuario.click(await screen.findByRole('button', { name: 'Nueva ronda' }))

    expect(nuevaRondaMock).toHaveBeenCalledWith(1, 7)
    expect(await screen.findByText('Ronda 2')).toBeInTheDocument()
    expect(screen.getByText('Votando')).toBeInTheDocument()
    expect(screen.getByRole('radio', { name: '5 puntos' })).not.toBeChecked()
    expect(screen.getByRole('radio', { name: '0 puntos' })).toBeEnabled()
    expect(screen.getByRole('button', { name: 'Revelar votos' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Nueva ronda' })).not.toBeInTheDocument()
    const historial = screen.getByRole('region', { name: 'Historial de rondas' })
    expect(historial).toHaveTextContent('Ronda 1')
    expect(historial).not.toHaveTextContent('Ronda 2')
  })

  it('ante un error de red al abrir ronda permite reintentar', async () => {
    prepararCargaInicial(detalleRevelada)
    nuevaRondaMock
      .mockRejectedValueOnce(
        new ApiError({ status: 0, code: 'NETWORK_ERROR', message: 'sin conexión' }),
      )
      .mockResolvedValueOnce(detalleNuevaRonda)
    const usuario = userEvent.setup()

    renderizar()

    await usuario.click(await screen.findByRole('button', { name: 'Nueva ronda' }))

    expect(
      await screen.findByText('No se pudo abrir la nueva ronda. Intentá nuevamente.'),
    ).toBeInTheDocument()

    await usuario.click(screen.getByRole('button', { name: 'Nueva ronda' }))

    expect(await screen.findByText('Ronda 2')).toBeInTheDocument()
    expect(nuevaRondaMock).toHaveBeenCalledTimes(2)
  })

  it('ante 409 de nueva ronda refresca y muestra que ya no está revelada', async () => {
    prepararCargaInicial(detalleRevelada)
    obtenerDetalleMock
      .mockResolvedValueOnce(detalleRevelada)
      .mockResolvedValueOnce(detalleVotando)
    nuevaRondaMock.mockRejectedValue(
      new ApiError({
        status: 409,
        code: 'SESION_POKER_NO_REVELADA',
        message: 'la sesión no está revelada',
      }),
    )
    const usuario = userEvent.setup()

    renderizar()

    await usuario.click(await screen.findByRole('button', { name: 'Nueva ronda' }))

    expect(await screen.findByText('La sesión ya no está revelada.')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Revelar votos' })).toBeInTheDocument()
  })

  it('muestra rondas anteriores en una sesión revelada', async () => {
    prepararCargaInicial({
      ...detalleRevelada,
      ronda_actual: 2,
      historial_rondas: [rondaAnterior],
    })

    renderizar()

    expect(
      await screen.findByRole('region', { name: 'Historial de rondas' }),
    ).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Ronda 1' })).toBeInTheDocument()
    expect(screen.queryByRole('heading', { name: 'Ronda 2' })).not.toBeInTheDocument()
  })

  it('con la última sesión Cerrada no muestra acciones de revelado ni ronda', async () => {
    listarSesionesMock.mockResolvedValue([sesionCerrada])

    renderizar()

    expect(await screen.findByText('Puntos acordados: 5')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Revelar votos' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Nueva ronda' })).not.toBeInTheDocument()
    expect(screen.queryByRole('radio')).not.toBeInTheDocument()
  })

  it('con la última sesión Cancelada no muestra acciones de revelado ni ronda', async () => {
    listarSesionesMock.mockResolvedValue([sesionCancelada])

    renderizar()

    expect(await screen.findByText('La última sesión fue cancelada.')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Revelar votos' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Nueva ronda' })).not.toBeInTheDocument()
  })
})

describe('SesionPokerPage - polling F17', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    listarHistoriasMock.mockReset()
    listarSesionesMock.mockReset()
    obtenerDetalleMock.mockReset()
    revelarMock.mockReset()
    nuevaRondaMock.mockReset()
    prepararCargaInicial(detalleRevelada)
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('detecta una nueva ronda abierta por otro miembro durante el polling', async () => {
    obtenerDetalleMock
      .mockResolvedValueOnce(detalleRevelada)
      .mockResolvedValueOnce(detalleNuevaRonda)

    renderizar()

    await act(async () => {
      await vi.advanceTimersByTimeAsync(0)
    })

    expect(screen.getByRole('radio', { name: '5 puntos' })).toBeDisabled()

    await act(async () => {
      await vi.advanceTimersByTimeAsync(3000)
    })

    expect(screen.getByText('Ronda 2')).toBeInTheDocument()
    expect(screen.getByText('Votando')).toBeInTheDocument()
    expect(screen.getByRole('radio', { name: '5 puntos' })).not.toBeChecked()
    expect(screen.getByRole('radio', { name: '0 puntos' })).toBeEnabled()
  })

  it('se detiene al detectar una sesión Cerrada', async () => {
    obtenerDetalleMock
      .mockResolvedValueOnce(detalleRevelada)
      .mockResolvedValueOnce(detalleCerrada)

    renderizar()

    await act(async () => {
      await vi.advanceTimersByTimeAsync(0)
    })

    await act(async () => {
      await vi.advanceTimersByTimeAsync(3000)
    })

    expect(screen.getByText('La sesión fue cerrada.')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Revelar votos' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Nueva ronda' })).not.toBeInTheDocument()

    await act(async () => {
      await vi.advanceTimersByTimeAsync(3000)
    })

    expect(obtenerDetalleMock).toHaveBeenCalledTimes(2)
  })
})
