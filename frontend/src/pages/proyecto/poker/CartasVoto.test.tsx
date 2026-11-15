import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import CartasVoto from './CartasVoto'

const ETIQUETAS = [
  '0 puntos',
  '0.5 puntos',
  '1 puntos',
  '2 puntos',
  '3 puntos',
  '5 puntos',
  '8 puntos',
  '13 puntos',
  '21 puntos',
]

describe('CartasVoto', () => {
  it('muestra las cartas de la escala en orden', () => {
    render(<CartasVoto valor={null} onSeleccionar={vi.fn()} />)

    const cartas = screen.getAllByRole('radio')

    expect(cartas).toHaveLength(9)
    expect(cartas.map((carta) => carta.getAttribute('value'))).toEqual([
      '0',
      '0.5',
      '1',
      '2',
      '3',
      '5',
      '8',
      '13',
      '21',
    ])
    ETIQUETAS.forEach((etiqueta) => {
      expect(screen.getByRole('radio', { name: etiqueta })).toBeInTheDocument()
    })
  })

  it('marca como seleccionado el voto propio', () => {
    render(<CartasVoto valor={5} onSeleccionar={vi.fn()} />)

    expect(screen.getByRole('radio', { name: '5 puntos' })).toBeChecked()
    expect(screen.getByRole('radio', { name: '8 puntos' })).not.toBeChecked()
  })

  it('informa el valor elegido al hacer clic', async () => {
    const onSeleccionar = vi.fn()
    const usuario = userEvent.setup()
    render(<CartasVoto valor={5} onSeleccionar={onSeleccionar} />)

    await usuario.click(screen.getByRole('radio', { name: '8 puntos' }))

    expect(onSeleccionar).toHaveBeenCalledWith(8)
  })

  it('cambia la carta seleccionada al recibir un nuevo voto', () => {
    const { rerender } = render(<CartasVoto valor={5} onSeleccionar={vi.fn()} />)

    expect(screen.getByRole('radio', { name: '5 puntos' })).toBeChecked()

    rerender(<CartasVoto valor={0.5} onSeleccionar={vi.fn()} />)

    expect(screen.getByRole('radio', { name: '0.5 puntos' })).toBeChecked()
    expect(screen.getByRole('radio', { name: '5 puntos' })).not.toBeChecked()
  })

  it('deshabilita todas las cartas fuera de Votando', () => {
    render(<CartasVoto valor={5} onSeleccionar={vi.fn()} deshabilitado />)

    screen.getAllByRole('radio').forEach((carta) => {
      expect(carta).toBeDisabled()
    })
  })

  it('permite enfocar y seleccionar una carta con teclado', async () => {
    const onSeleccionar = vi.fn()
    const usuario = userEvent.setup()
    render(<CartasVoto valor={null} onSeleccionar={onSeleccionar} />)

    await usuario.tab()

    expect(screen.getByRole('radio', { name: '0 puntos' })).toHaveFocus()

    await usuario.keyboard(' ')

    expect(onSeleccionar).toHaveBeenCalledWith(0)
  })
})
