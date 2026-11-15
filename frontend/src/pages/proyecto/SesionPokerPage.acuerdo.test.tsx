import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiError } from '../../api/errors'
import { listarHistorias } from '../../api/historias'
import {
  acordarSesionPoker,
  cancelarSesionPoker,
  listarSesionesPoker,
  obtenerDetalleSesionPoker,
} from '../../api/poker'
import type {
  DetalleSesionPoker,
  DispersionPoker,
  Historia,
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
  acordarSesionPoker: vi.fn(),
  cancelarSesionPoker: vi.fn(),
}))

const listarHistoriasMock = vi.mocked(listarHistorias)
const listarSesionesMock = vi.mocked(listarSesionesPoker)
const obtenerDetalleMock = vi.mocked(obtenerDetalleSesionPoker)
const acordarMock = vi.mocked(acordarSesionPoker)
const cancelarMock = vi.mocked(cancelarSesionPoker)

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

const sesionRevelada: SesionPokerHistorial = { ...sesionVotando, estado: 'Revelada' }

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

const detalleSinMedianaValida: DetalleSesionPoker = {
  ...detalleRevelada,
  dispersion: { ...dispersion, mediana: 4, votos_por_valor: { '3': 2, '8': 1 } },
}

const detalleEmpate: DetalleSesionPoker = {
  ...detalleRevelada,
  dispersion: { ...dispersion, mediana: 4, votos_por_valor: { '1': 1, '3': 1, '8': 1 } },
}

const detalleCerrada: DetalleSesionPoker = {
  ...detalleRevelada,
  estado: 'Cerrada',
  puntos_acordados: 3,
  cerrado_en: '2026-11-13T11:00:00Z',
}

const detalleCancelada: DetalleSesionPoker = {
  ...detalleRevelada,
  estado: 'Cancelada',
  cancelado_en: '2026-11-13T11:30:00Z',
  puntos_acordados: null,
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

function prepararCargaInicial(detalle: DetalleSesionPoker = detalleRevelada) {
  listarHistoriasMock.mockResolvedValue([historia])
  listarSesionesMock.mockResolvedValue([
    detalle.estado === 'Votando' ? sesionVotando : sesionRevelada,
  ])
  obtenerDetalleMock.mockResolvedValue(detalle)
}

describe('SesionPokerPage - acuerdo y cierre', () => {
  beforeEach(() => {
    listarHistoriasMock.mockReset()
    listarSesionesMock.mockReset()
    obtenerDetalleMock.mockReset()
    acordarMock.mockReset()
    cancelarMock.mockReset()
    prepararCargaInicial()
  })

  it('muestra Registrar acuerdo solo en Revelada', async () => {
    renderizar()

    expect(await screen.findByRole('button', { name: 'Registrar acuerdo' })).toBeInTheDocument()
  })

  it('no muestra Registrar acuerdo en Votando', async () => {
    prepararCargaInicial(detalleVotando)

    renderizar()

    expect(await screen.findByRole('button', { name: 'Revelar votos' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Registrar acuerdo' })).not.toBeInTheDocument()
  })

  it('prefija la mediana válida y cierra la sesión con el acuerdo', async () => {
    acordarMock.mockResolvedValue(detalleCerrada)
    const usuario = userEvent.setup()

    renderizar()

    await usuario.click(await screen.findByRole('button', { name: 'Registrar acuerdo' }))

    const dialogo = screen.getByRole('dialog')
    expect(within(dialogo).getByRole('radio', { name: '3 puntos' })).toBeChecked()

    await usuario.click(screen.getByRole('button', { name: 'Confirmar acuerdo' }))

    expect(acordarMock).toHaveBeenCalledWith(1, 7, 3)
    expect(await screen.findByText('Puntos acordados: 3')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Registrar acuerdo' })).not.toBeInTheDocument()
    expect(screen.queryByRole('radio')).not.toBeInTheDocument()
  })

  it('prefija el valor más votado cuando la mediana no es de la escala', async () => {
    prepararCargaInicial(detalleSinMedianaValida)
    const usuario = userEvent.setup()

    renderizar()

    await usuario.click(await screen.findByRole('button', { name: 'Registrar acuerdo' }))

    const dialogo = screen.getByRole('dialog')
    expect(within(dialogo).getByRole('radio', { name: '3 puntos' })).toBeChecked()
  })

  it('deja el acuerdo vacío ante un empate y valida antes de enviar', async () => {
    prepararCargaInicial(detalleEmpate)
    const usuario = userEvent.setup()

    renderizar()

    await usuario.click(await screen.findByRole('button', { name: 'Registrar acuerdo' }))

    const dialogo = screen.getByRole('dialog')
    for (const radio of within(dialogo).getAllByRole('radio')) {
      expect(radio).not.toBeChecked()
    }

    await usuario.click(screen.getByRole('button', { name: 'Confirmar acuerdo' }))

    expect(acordarMock).not.toHaveBeenCalled()
    expect(screen.getByText('Elegí una estimación de la escala.')).toBeInTheDocument()
  })

  it('muestra 422 PUNTOS_POKER_INVALIDOS y mantiene la sesión Revelada', async () => {
    acordarMock.mockRejectedValue(
      new ApiError({
        status: 422,
        code: 'PUNTOS_POKER_INVALIDOS',
        message: 'los puntos deben pertenecer a la escala',
        campo: 'puntos',
      }),
    )
    const usuario = userEvent.setup()

    renderizar()

    await usuario.click(await screen.findByRole('button', { name: 'Registrar acuerdo' }))
    await usuario.click(screen.getByRole('button', { name: 'Confirmar acuerdo' }))

    expect(
      await screen.findByText('El valor elegido no es válido. Elegí otro de la escala.'),
    ).toBeInTheDocument()
    expect(screen.getByText('Revelada')).toBeInTheDocument()
  })

  it('ante un error de red al acordar permite reintentar', async () => {
    acordarMock
      .mockRejectedValueOnce(
        new ApiError({ status: 0, code: 'NETWORK_ERROR', message: 'sin conexión' }),
      )
      .mockResolvedValueOnce(detalleCerrada)
    const usuario = userEvent.setup()

    renderizar()

    await usuario.click(await screen.findByRole('button', { name: 'Registrar acuerdo' }))
    await usuario.click(screen.getByRole('button', { name: 'Confirmar acuerdo' }))

    expect(
      await screen.findByText('No se pudo registrar el acuerdo. Intentá nuevamente.'),
    ).toBeInTheDocument()

    await usuario.click(screen.getByRole('button', { name: 'Confirmar acuerdo' }))

    expect(await screen.findByText('Puntos acordados: 3')).toBeInTheDocument()
    expect(acordarMock).toHaveBeenCalledTimes(2)
  })

  it('cierra el modal con Escape sin enviar el acuerdo', async () => {
    const usuario = userEvent.setup()

    renderizar()

    await usuario.click(await screen.findByRole('button', { name: 'Registrar acuerdo' }))
    await usuario.keyboard('{Escape}')

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    expect(acordarMock).not.toHaveBeenCalled()
  })

  it('ante 409 SESION_POKER_CERRADA al acordar refresca y avisa', async () => {
    obtenerDetalleMock
      .mockReset()
      .mockResolvedValueOnce(detalleRevelada)
      .mockResolvedValueOnce(detalleCerrada)
    acordarMock.mockRejectedValue(
      new ApiError({
        status: 409,
        code: 'SESION_POKER_CERRADA',
        message: 'la sesión ya está cerrada',
      }),
    )
    const usuario = userEvent.setup()

    renderizar()

    await usuario.click(await screen.findByRole('button', { name: 'Registrar acuerdo' }))
    await usuario.click(screen.getByRole('button', { name: 'Confirmar acuerdo' }))

    expect(await screen.findByText('La sesión ya fue cerrada o cancelada.')).toBeInTheDocument()
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })
})

describe('SesionPokerPage - cancelación', () => {
  beforeEach(() => {
    listarHistoriasMock.mockReset()
    listarSesionesMock.mockReset()
    obtenerDetalleMock.mockReset()
    acordarMock.mockReset()
    cancelarMock.mockReset()
    prepararCargaInicial()
  })

  it('muestra Cancelar sesión en Votando', async () => {
    prepararCargaInicial(detalleVotando)

    renderizar()

    expect(await screen.findByRole('button', { name: 'Cancelar sesión' })).toBeInTheDocument()
  })

  it('cancela la sesión Revelada tras confirmar y permite iniciar otra', async () => {
    cancelarMock.mockResolvedValue(detalleCancelada)
    const usuario = userEvent.setup()

    renderizar()

    await usuario.click(await screen.findByRole('button', { name: 'Cancelar sesión' }))
    await usuario.click(screen.getByRole('button', { name: 'Sí, cancelar' }))

    expect(cancelarMock).toHaveBeenCalledWith(1, 7)
    expect(await screen.findByText('La última sesión fue cancelada.')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Iniciar Planning Poker' })).toBeInTheDocument()
    expect(screen.queryByText(/Puntos acordados/)).not.toBeInTheDocument()
  })

  it('descarta el diálogo de cancelación con Escape sin enviar', async () => {
    const usuario = userEvent.setup()

    renderizar()

    await usuario.click(await screen.findByRole('button', { name: 'Cancelar sesión' }))
    await usuario.keyboard('{Escape}')

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    expect(cancelarMock).not.toHaveBeenCalled()
  })

  it('ante un error de red al cancelar avisa y permite reintentar', async () => {
    cancelarMock
      .mockRejectedValueOnce(
        new ApiError({ status: 0, code: 'NETWORK_ERROR', message: 'sin conexión' }),
      )
      .mockResolvedValueOnce(detalleCancelada)
    const usuario = userEvent.setup()

    renderizar()

    await usuario.click(await screen.findByRole('button', { name: 'Cancelar sesión' }))
    await usuario.click(screen.getByRole('button', { name: 'Sí, cancelar' }))

    expect(
      await screen.findByText('No se pudo cancelar la sesión. Intentá nuevamente.'),
    ).toBeInTheDocument()

    await usuario.click(screen.getByRole('button', { name: 'Cancelar sesión' }))
    await usuario.click(screen.getByRole('button', { name: 'Sí, cancelar' }))

    expect(await screen.findByText('La última sesión fue cancelada.')).toBeInTheDocument()
    expect(cancelarMock).toHaveBeenCalledTimes(2)
  })

  it('con una sesión ya Cerrada no ofrece acordar ni cancelar', async () => {
    listarSesionesMock.mockResolvedValue([
      { ...sesionRevelada, estado: 'Cerrada', puntos_acordados: 5 },
    ])

    renderizar()

    expect(await screen.findByText('Puntos acordados: 5')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Registrar acuerdo' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Cancelar sesión' })).not.toBeInTheDocument()
  })
})
