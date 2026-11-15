import { render, screen, within } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import type { HistorialRondaPoker } from '../../../api/types'
import HistorialRondas from './HistorialRondas'

const rondaAnterior: HistorialRondaPoker = {
  numero: 1,
  estado: 'Revelada',
  revelada_en: '2026-11-13T10:05:00Z',
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
}

const rondaActual: HistorialRondaPoker = {
  ...rondaAnterior,
  numero: 2,
  estado: 'Votando',
  revelada_en: null,
  votos: [],
}

describe('HistorialRondas', () => {
  it('muestra las rondas anteriores con fecha, votos y estadísticas', () => {
    render(<HistorialRondas rondas={[rondaAnterior]} rondaActual={2} />)

    const historial = screen.getByRole('region', { name: 'Historial de rondas' })

    expect(within(historial).getByRole('heading', { name: 'Ronda 1' })).toBeInTheDocument()
    expect(within(historial).getByText('Revelada el 13/11/2026')).toBeInTheDocument()
    expect(within(historial).getByText('Ana: 1 punto')).toBeInTheDocument()
    expect(within(historial).getByText('Ben: 3 puntos')).toBeInTheDocument()
    expect(within(historial).getByText('Cris: 8 puntos')).toBeInTheDocument()
    expect(within(historial).getByText('Promedio: 4')).toBeInTheDocument()
  })

  it('no muestra la ronda actual entre las rondas anteriores', () => {
    render(<HistorialRondas rondas={[rondaAnterior, rondaActual]} rondaActual={2} />)

    const historial = screen.getByRole('region', { name: 'Historial de rondas' })

    expect(within(historial).getByRole('heading', { name: 'Ronda 1' })).toBeInTheDocument()
    expect(within(historial).queryByRole('heading', { name: 'Ronda 2' })).not.toBeInTheDocument()
  })

  it('no renderiza la sección cuando no hay rondas anteriores', () => {
    render(<HistorialRondas rondas={[rondaActual]} rondaActual={2} />)

    expect(
      screen.queryByRole('region', { name: 'Historial de rondas' }),
    ).not.toBeInTheDocument()
  })

  it('omite la fecha cuando la ronda no tiene revelada_en', () => {
    render(
      <HistorialRondas
        rondas={[{ ...rondaAnterior, revelada_en: null }]}
        rondaActual={2}
      />,
    )

    expect(screen.queryByText(/Revelada el/)).not.toBeInTheDocument()
  })
})
