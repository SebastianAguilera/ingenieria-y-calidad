import { render, screen, within } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import type { DispersionPoker } from '../../../api/types'
import PanelDispersion from './PanelDispersion'

const dispersion: DispersionPoker = {
  promedio: 4,
  mediana: 3,
  minimo: 1,
  maximo: 8,
  rango: 7,
  desviacion_estandar: 2.9439202887759492,
  hay_dispersion: true,
  total_votos: 3,
  votos_por_valor: { '8': 1, '1': 1, '3': 1 },
}

describe('PanelDispersion', () => {
  it('muestra las estadísticas formateadas en es-AR', () => {
    render(<PanelDispersion dispersion={dispersion} />)

    const panel = screen.getByRole('region', { name: 'Dispersión de votos' })

    expect(within(panel).getByText('Promedio: 4')).toBeInTheDocument()
    expect(within(panel).getByText('Mediana: 3')).toBeInTheDocument()
    expect(within(panel).getByText('Mínimo: 1')).toBeInTheDocument()
    expect(within(panel).getByText('Máximo: 8')).toBeInTheDocument()
    expect(within(panel).getByText('Rango: 7')).toBeInTheDocument()
    expect(within(panel).getByText('Desviación estándar: 2,94')).toBeInTheDocument()
    expect(within(panel).getByText('Total de votos: 3')).toBeInTheDocument()
  })

  it('muestra la distribución ordenada por valor', () => {
    render(<PanelDispersion dispersion={dispersion} />)

    const items = within(
      screen.getByRole('list', { name: 'Distribución de votos' }),
    ).getAllByRole('listitem')

    expect(items.map((item) => item.textContent)).toEqual([
      '1 punto: 1 voto',
      '3 puntos: 1 voto',
      '8 puntos: 1 voto',
    ])
  })

  it('muestra el badge de diferencias cuando hay dispersión', () => {
    render(<PanelDispersion dispersion={dispersion} />)

    expect(screen.getByText('Hay diferencias')).toBeInTheDocument()
  })

  it('muestra el badge de consenso y usa singular/plural en la distribución', () => {
    render(
      <PanelDispersion
        dispersion={{
          ...dispersion,
          promedio: 3,
          mediana: 3,
          minimo: 3,
          maximo: 3,
          rango: 0,
          desviacion_estandar: 0,
          hay_dispersion: false,
          total_votos: 2,
          votos_por_valor: { '3': 2 },
        }}
      />,
    )

    expect(screen.getByText('Consenso')).toBeInTheDocument()
    expect(screen.getByText('3 puntos: 2 votos')).toBeInTheDocument()
  })

  it('formatea valores decimales de la escala', () => {
    render(
      <PanelDispersion
        dispersion={{
          ...dispersion,
          promedio: 0.5,
          mediana: 0.5,
          minimo: 0.5,
          maximo: 0.5,
          rango: 0,
          desviacion_estandar: 0,
          hay_dispersion: false,
          total_votos: 1,
          votos_por_valor: { '0.5': 1 },
        }}
      />,
    )

    expect(screen.getByText('Promedio: 0,5')).toBeInTheDocument()
    expect(screen.getByText('0,5 puntos: 1 voto')).toBeInTheDocument()
  })
})
