import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiError } from '../../api/errors'
import { listarHistorias } from '../../api/historias'
import { crearSesionPoker, listarSesionesPoker } from '../../api/poker'
import type { Historia, SesionPokerHistorial } from '../../api/types'
import SesionPokerPage from './SesionPokerPage'

vi.mock('../../api/historias', () => ({
  listarHistorias: vi.fn(),
}))

vi.mock('../../api/poker', () => ({
  listarSesionesPoker: vi.fn(),
  crearSesionPoker: vi.fn(),
}))

const listarHistoriasMock = vi.mocked(listarHistorias)
const listarSesionesMock = vi.mocked(listarSesionesPoker)
const crearMock = vi.mocked(crearSesionPoker)

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
  id: 8,
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

describe('SesionPokerPage', () => {
  beforeEach(() => {
    listarHistoriasMock.mockReset()
    listarSesionesMock.mockReset()
    crearMock.mockReset()
    listarHistoriasMock.mockResolvedValue([historia])
    listarSesionesMock.mockResolvedValue([])
  })

  it('muestra el indicador de carga mientras resuelve historia e historial', () => {
    listarHistoriasMock.mockReturnValue(new Promise(() => {}))
    listarSesionesMock.mockReturnValue(new Promise(() => {}))

    renderizar()

    expect(screen.getByText('Cargando Planning Poker…')).toBeInTheDocument()
  })

  it('sin sesiones muestra el resumen y el botón para iniciar, sin crear nada', async () => {
    renderizar()

    expect(
      await screen.findByRole('button', { name: 'Iniciar Planning Poker' }),
    ).toBeInTheDocument()
    expect(screen.getByText('US-01 · Registrar horas reales')).toBeInTheDocument()
    expect(
      screen.getByText('Todavía no hay sesiones de Planning Poker para esta historia.'),
    ).toBeInTheDocument()
    expect(crearMock).not.toHaveBeenCalled()
  })

  it('inicia la sesión solo al hacer clic y muestra el estado Votando', async () => {
    crearMock.mockResolvedValue({
      id: 7,
      historia_id: 10,
      estado: 'Votando',
      ronda_actual: 1,
      creado_por: { id: 2, nombre: 'Ana' },
      creado_en: '2026-11-13T10:00:00Z',
      puntos_acordados: null,
    })
    const usuario = userEvent.setup()

    renderizar()

    await usuario.click(await screen.findByRole('button', { name: 'Iniciar Planning Poker' }))

    expect(crearMock).toHaveBeenCalledWith(1, 10)
    expect(await screen.findByRole('heading', { name: 'Sesión en curso' })).toBeInTheDocument()
    expect(screen.getByText('Votando')).toBeInTheDocument()
    expect(screen.getByText('Ronda 1')).toBeInTheDocument()
  })

  it('con una sesión Votando entra directamente y no ofrece iniciar otra', async () => {
    listarSesionesMock.mockResolvedValue([sesionVotando])

    renderizar()

    expect(await screen.findByRole('heading', { name: 'Sesión en curso' })).toBeInTheDocument()
    expect(screen.getByText('Votando')).toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: 'Iniciar Planning Poker' }),
    ).not.toBeInTheDocument()
    expect(crearMock).not.toHaveBeenCalled()
  })

  it('con una sesión Revelada informa que los votos fueron revelados', async () => {
    listarSesionesMock.mockResolvedValue([sesionRevelada])

    renderizar()

    expect(await screen.findByRole('heading', { name: 'Sesión en curso' })).toBeInTheDocument()
    expect(screen.getByText('Revelada')).toBeInTheDocument()
    expect(
      screen.getByText('Los votos fueron revelados. La sesión sigue activa.'),
    ).toBeInTheDocument()
    expect(crearMock).not.toHaveBeenCalled()
  })

  it('con la última sesión Cerrada muestra los puntos acordados', async () => {
    listarSesionesMock.mockResolvedValue([sesionCerrada])

    renderizar()

    expect(
      await screen.findByRole('heading', { name: 'Última sesión cerrada' }),
    ).toBeInTheDocument()
    expect(screen.getByText('Puntos acordados: 5')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Iniciar Planning Poker' })).toBeInTheDocument()
  })

  it('con la última sesión Cerrada sin puntos informa que no hubo acuerdo', async () => {
    listarSesionesMock.mockResolvedValue([{ ...sesionCerrada, puntos_acordados: null }])

    renderizar()

    expect(await screen.findByText('Sin puntos acordados')).toBeInTheDocument()
  })

  it('con la última sesión Cancelada lo informa y permite iniciar otra', async () => {
    listarSesionesMock.mockResolvedValue([sesionCancelada])

    renderizar()

    expect(await screen.findByText('La última sesión fue cancelada.')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Iniciar Planning Poker' })).toBeInTheDocument()
  })

  it('muestra HISTORIA_NO_ESTIMABLE al intentar iniciar una historia Hecho', async () => {
    listarHistoriasMock.mockResolvedValue([{ ...historia, estado: 'Hecho' }])
    crearMock.mockRejectedValue(
      new ApiError({
        status: 409,
        code: 'HISTORIA_NO_ESTIMABLE',
        message: 'la historia ya está hecha',
      }),
    )
    const usuario = userEvent.setup()

    renderizar()

    await usuario.click(await screen.findByRole('button', { name: 'Iniciar Planning Poker' }))

    expect(
      await screen.findByText('La historia está en estado Hecho y no se puede estimar.'),
    ).toBeInTheDocument()
    expect(screen.queryByRole('heading', { name: 'Sesión en curso' })).not.toBeInTheDocument()
  })

  it('ante SESION_POKER_ACTIVA relee el historial y se une a la sesión', async () => {
    listarSesionesMock
      .mockResolvedValueOnce([])
      .mockResolvedValueOnce([sesionVotando])
    crearMock.mockRejectedValue(
      new ApiError({
        status: 409,
        code: 'SESION_POKER_ACTIVA',
        message: 'ya hay una sesión activa',
      }),
    )
    const usuario = userEvent.setup()

    renderizar()

    await usuario.click(await screen.findByRole('button', { name: 'Iniciar Planning Poker' }))

    expect(await screen.findByRole('heading', { name: 'Sesión en curso' })).toBeInTheDocument()
    expect(listarSesionesMock).toHaveBeenCalledTimes(2)
  })

  it('muestra historia no encontrada ante 404 HISTORIA_NO_ENCONTRADA', async () => {
    listarSesionesMock.mockRejectedValue(
      new ApiError({
        status: 404,
        code: 'HISTORIA_NO_ENCONTRADA',
        message: 'la historia no existe',
      }),
    )

    renderizar()

    expect(
      await screen.findByText('La historia no existe o no pertenece a este proyecto.'),
    ).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Volver al backlog' })).toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: 'Iniciar Planning Poker' }),
    ).not.toBeInTheDocument()
  })

  it('muestra falta de acceso ante 404 PROYECTO_NO_ENCONTRADO', async () => {
    listarSesionesMock.mockRejectedValue(
      new ApiError({
        status: 404,
        code: 'PROYECTO_NO_ENCONTRADO',
        message: 'el proyecto no existe',
      }),
    )

    renderizar()

    expect(
      await screen.findByText('No tenés acceso a este proyecto o no existe'),
    ).toBeInTheDocument()
  })

  it('muestra error de red al cargar y permite reintentar', async () => {
    listarSesionesMock
      .mockRejectedValueOnce(
        new ApiError({ status: 0, code: 'NETWORK_ERROR', message: 'sin conexión' }),
      )
      .mockResolvedValueOnce([])
    const usuario = userEvent.setup()

    renderizar()

    expect(
      await screen.findByText('No se pudo cargar el Planning Poker. Intentá nuevamente.'),
    ).toBeInTheDocument()
    await usuario.click(screen.getByRole('button', { name: 'Reintentar' }))

    expect(
      await screen.findByRole('button', { name: 'Iniciar Planning Poker' }),
    ).toBeInTheDocument()
    expect(listarSesionesMock).toHaveBeenCalledTimes(2)
  })

  it('muestra error de red al iniciar y no deja una sesión activa', async () => {
    crearMock.mockRejectedValue(
      new ApiError({ status: 0, code: 'NETWORK_ERROR', message: 'sin conexión' }),
    )
    const usuario = userEvent.setup()

    renderizar()

    await usuario.click(await screen.findByRole('button', { name: 'Iniciar Planning Poker' }))

    expect(
      await screen.findByText('No se pudo iniciar la sesión. Intentá nuevamente.'),
    ).toBeInTheDocument()
    expect(screen.queryByRole('heading', { name: 'Sesión en curso' })).not.toBeInTheDocument()
  })
})
