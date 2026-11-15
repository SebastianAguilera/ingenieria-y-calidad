import { act, fireEvent, render, screen, within } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { listarSesionesActivas } from '../../../api/poker'
import type { SesionPokerActiva, Usuario } from '../../../api/types'
import { SesionContext } from '../../../session/sesion'
import CampanaPoker from './CampanaPoker'
import { ProveedorSesionesPoker } from './ProveedorSesionesPoker'
import {
  ETIQUETA_CAMPANA,
  MENSAJE_SIN_SESIONES,
  useSesionesPoker,
} from './sesionesPoker'

vi.mock('../../../api/poker', () => ({
  listarSesionesActivas: vi.fn(),
}))

const listarMock = vi.mocked(listarSesionesActivas)

const usuario: Usuario = {
  id: 1,
  nombre: 'Yo',
  email: 'yo@test.com',
  creado_en: '2026-01-01T00:00:00Z',
}

function sesionActiva(overrides: Partial<SesionPokerActiva> = {}): SesionPokerActiva {
  return {
    sesion_id: 7,
    estado: 'Votando',
    ronda_actual: 1,
    creado_por: { id: 2, nombre: 'Ana' },
    creado_en: '2026-11-13T10:00:00Z',
    votos_registrados: 0,
    total_miembros: 3,
    ya_vote: false,
    historia: { id: 10, codigo: 'US-01', titulo: 'Registrar horas reales' },
    ...overrides,
  }
}

function ConsumidorHistorias() {
  const { historiasConPoker } = useSesionesPoker()
  const ids = [...historiasConPoker].sort((a, b) => a - b)
  return <p>Historias con poker: {ids.join(',')}</p>
}

function renderizar({ usuarioId = 1 }: { usuarioId?: number } = {}) {
  return render(
    <SesionContext.Provider
      value={{
        usuarioActivo: { ...usuario, id: usuarioId },
        token: 'token',
        ingresar: vi.fn(),
        salir: vi.fn(),
      }}
    >
      <MemoryRouter initialEntries={['/proyectos/1/backlog']}>
        <ProveedorSesionesPoker proyectoId={1}>
          <CampanaPoker />
          <ConsumidorHistorias />
          <Routes>
            <Route
              path="/proyectos/:id/backlog/:historiaId/poker"
              element={<div>Pantalla de Planning Poker</div>}
            />
            <Route path="*" element={<div>Otra pantalla</div>} />
          </Routes>
        </ProveedorSesionesPoker>
      </MemoryRouter>
    </SesionContext.Provider>,
  )
}

async function avanzar(ms: number) {
  await act(async () => {
    await vi.advanceTimersByTimeAsync(ms)
  })
}

function definirOculto(oculto: boolean) {
  Object.defineProperty(document, 'hidden', {
    configurable: true,
    get: () => oculto,
  })
}

async function cambiarVisibilidad(oculto: boolean) {
  definirOculto(oculto)
  await act(async () => {
    document.dispatchEvent(new Event('visibilitychange'))
    await vi.advanceTimersByTimeAsync(0)
  })
}

function campana() {
  return screen.getByRole('button', { name: ETIQUETA_CAMPANA })
}

describe('ProveedorSesionesPoker y CampanaPoker', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    listarMock.mockReset()
    definirOculto(false)
  })

  afterEach(() => {
    vi.useRealTimers()
    definirOculto(false)
  })

  it('hace un único polling cada 3 segundos', async () => {
    listarMock.mockResolvedValue([])

    renderizar()
    await avanzar(0)

    expect(listarMock).toHaveBeenCalledTimes(1)
    expect(listarMock.mock.calls[0][0]).toBe(1)

    await avanzar(3000)
    expect(listarMock).toHaveBeenCalledTimes(2)

    await avanzar(3000)
    expect(listarMock).toHaveBeenCalledTimes(3)
  })

  it('pausa el polling con la pestaña oculta y lo reanuda al volver', async () => {
    listarMock.mockResolvedValue([])

    renderizar()
    await avanzar(0)
    expect(listarMock).toHaveBeenCalledTimes(1)

    await cambiarVisibilidad(true)
    await avanzar(3000)
    expect(listarMock).toHaveBeenCalledTimes(1)

    await cambiarVisibilidad(false)
    expect(listarMock).toHaveBeenCalledTimes(2)

    await avanzar(3000)
    expect(listarMock).toHaveBeenCalledTimes(3)
  })

  it('no consulta al montar con la pestaña oculta', async () => {
    listarMock.mockResolvedValue([])
    definirOculto(true)

    renderizar()
    await avanzar(0)
    await avanzar(3000)

    expect(listarMock).not.toHaveBeenCalled()

    await cambiarVisibilidad(false)
    expect(listarMock).toHaveBeenCalledTimes(1)
  })

  it('no muestra toast en la primera carga y sí lista la sesión con badge', async () => {
    listarMock.mockResolvedValue([sesionActiva()])

    renderizar()
    await avanzar(0)

    expect(screen.queryByRole('status')).not.toBeInTheDocument()
    expect(screen.getByLabelText('1 sesión sin votar')).toBeInTheDocument()
    expect(screen.getByText('Historias con poker: 10')).toBeInTheDocument()

    fireEvent.click(campana())

    expect(screen.getByText('US-01')).toBeInTheDocument()
    expect(screen.getByText('Registrar horas reales')).toBeInTheDocument()
    expect(screen.getByText(/Ronda 1/)).toBeInTheDocument()
    expect(screen.getByText(/Creada por Ana/)).toBeInTheDocument()
    expect(screen.getByText('Votando')).toBeInTheDocument()
    expect(screen.getByText('Falta tu voto')).toBeInTheDocument()
  })

  it('muestra toast y badge cuando aparece una sesión nueva en votación', async () => {
    listarMock.mockResolvedValueOnce([]).mockResolvedValue([sesionActiva()])

    renderizar()
    await avanzar(0)
    expect(screen.queryByRole('status')).not.toBeInTheDocument()

    await avanzar(3000)

    const toast = screen.getByRole('status')
    expect(toast).toHaveTextContent(
      'Se abrió Planning Poker para US-01: Registrar horas reales',
    )
    expect(toast).toHaveAttribute('aria-live', 'polite')
    expect(screen.getByLabelText('1 sesión sin votar')).toBeInTheDocument()
  })

  it('no repite el toast para una sesión ya vista', async () => {
    listarMock.mockResolvedValueOnce([]).mockResolvedValue([sesionActiva()])

    renderizar()
    await avanzar(0)
    await avanzar(3000)
    expect(screen.getByRole('status')).toHaveTextContent('Se abrió Planning Poker para US-01')

    await avanzar(3000)
    expect(listarMock).toHaveBeenCalledTimes(3)
    expect(screen.getByRole('status')).toBeInTheDocument()

    await avanzar(5000)

    expect(screen.queryByRole('status')).not.toBeInTheDocument()
  })

  it('no notifica al creador de la sesión', async () => {
    listarMock.mockResolvedValueOnce([]).mockResolvedValue([
      sesionActiva({ creado_por: { id: 1, nombre: 'Yo' } }),
    ])

    renderizar({ usuarioId: 1 })
    await avanzar(0)
    await avanzar(3000)

    expect(screen.queryByRole('status')).not.toBeInTheDocument()
    expect(screen.getByLabelText('1 sesión sin votar')).toBeInTheDocument()

    fireEvent.click(campana())
    expect(screen.getByText('US-01')).toBeInTheDocument()
  })

  it('lista las sesiones Revelada sin badge ni toast', async () => {
    listarMock.mockResolvedValueOnce([]).mockResolvedValue([
      sesionActiva({
        sesion_id: 8,
        estado: 'Revelada',
        historia: { id: 11, codigo: 'US-02', titulo: 'Cerrar el sprint' },
      }),
    ])

    renderizar()
    await avanzar(0)
    await avanzar(3000)

    expect(screen.queryByRole('status')).not.toBeInTheDocument()
    expect(screen.queryByLabelText(/sin votar/)).not.toBeInTheDocument()

    fireEvent.click(campana())

    expect(screen.getByText('US-02')).toBeInTheDocument()
    expect(screen.getByText('Revelada')).toBeInTheDocument()
  })

  it('muestra las sesiones nuevas una por una', async () => {
    const primera = sesionActiva()
    const segunda = sesionActiva({
      sesion_id: 8,
      historia: { id: 11, codigo: 'US-02', titulo: 'Cerrar el sprint' },
      creado_por: { id: 3, nombre: 'Ben' },
    })
    listarMock.mockResolvedValueOnce([]).mockResolvedValue([primera, segunda])

    renderizar()
    await avanzar(0)
    await avanzar(3000)

    expect(screen.getByRole('status')).toHaveTextContent(
      'Se abrió Planning Poker para US-01: Registrar horas reales',
    )

    await avanzar(5000)

    expect(screen.getByRole('status')).toHaveTextContent(
      'Se abrió Planning Poker para US-02: Cerrar el sprint',
    )

    await avanzar(5000)

    expect(screen.queryByRole('status')).not.toBeInTheDocument()
  })

  it('conserva la última lista cuando falla un polling', async () => {
    listarMock
      .mockResolvedValueOnce([])
      .mockResolvedValueOnce([sesionActiva()])
      .mockRejectedValue(new Error('sin conexión'))

    renderizar()
    await avanzar(0)
    await avanzar(3000)
    expect(screen.getByLabelText('1 sesión sin votar')).toBeInTheDocument()

    await avanzar(3000)

    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
    expect(screen.getByLabelText('1 sesión sin votar')).toBeInTheDocument()

    fireEvent.click(campana())
    expect(screen.getByText('US-01')).toBeInTheDocument()
  })

  it('abre y cierra la lista con teclado y devuelve el foco a la campana', async () => {
    listarMock.mockResolvedValue([sesionActiva()])

    renderizar()
    await avanzar(0)

    const boton = campana()
    expect(boton).toHaveAttribute('aria-expanded', 'false')
    expect(boton).toHaveAttribute('aria-controls')

    fireEvent.click(boton)
    expect(boton).toHaveAttribute('aria-expanded', 'true')

    const lista = screen.getByRole('region', {
      name: 'Sesiones activas de Planning Poker',
    })
    expect(within(lista).getByRole('button', { name: /US-01/ })).toBeInTheDocument()

    fireEvent.keyDown(document, { key: 'Escape' })

    expect(boton).toHaveAttribute('aria-expanded', 'false')
    expect(
      screen.queryByRole('region', { name: 'Sesiones activas de Planning Poker' }),
    ).not.toBeInTheDocument()
    expect(boton).toHaveFocus()
  })

  it('cierra la lista al hacer clic fuera de la campana', async () => {
    listarMock.mockResolvedValue([sesionActiva()])

    renderizar()
    await avanzar(0)

    fireEvent.click(campana())
    expect(
      screen.getByRole('region', { name: 'Sesiones activas de Planning Poker' }),
    ).toBeInTheDocument()

    fireEvent.mouseDown(document.body)

    expect(
      screen.queryByRole('region', { name: 'Sesiones activas de Planning Poker' }),
    ).not.toBeInTheDocument()
  })

  it('muestra el mensaje vacío cuando no hay sesiones', async () => {
    listarMock.mockResolvedValue([])

    renderizar()
    await avanzar(0)

    expect(screen.queryByLabelText(/sin votar/)).not.toBeInTheDocument()
    fireEvent.click(campana())

    expect(screen.getByText(MENSAJE_SIN_SESIONES)).toBeInTheDocument()
  })

  it('navega a la pantalla de Planning Poker de la historia al hacer clic', async () => {
    listarMock.mockResolvedValue([sesionActiva()])

    renderizar()
    await avanzar(0)

    fireEvent.click(campana())
    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: /US-01/ }))
    })

    expect(screen.getByText('Pantalla de Planning Poker')).toBeInTheDocument()
  })

  it('no cuenta en el badge las sesiones donde ya voté', async () => {
    listarMock.mockResolvedValue([
      sesionActiva({ sesion_id: 7, ya_vote: true }),
      sesionActiva({ sesion_id: 8, historia: { id: 11, codigo: 'US-02', titulo: 'Otra' } }),
    ])

    renderizar()
    await avanzar(0)

    expect(screen.getByLabelText('1 sesión sin votar')).toBeInTheDocument()
  })
})
