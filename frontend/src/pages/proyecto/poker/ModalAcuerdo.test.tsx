import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import ModalAcuerdo from './ModalAcuerdo'

describe('ModalAcuerdo', () => {
  it('muestra el diálogo, el título y los nueve valores de la escala', () => {
    render(
      <ModalAcuerdo valorInicial={null} onConfirmar={vi.fn()} onCerrar={vi.fn()} />,
    )

    expect(screen.getByRole('dialog')).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Registrar acuerdo' })).toBeInTheDocument()
    expect(screen.getAllByRole('radio')).toHaveLength(9)
    expect(screen.getByRole('radio', { name: '0.5 puntos' })).toBeInTheDocument()
    expect(screen.getByRole('radio', { name: '21 puntos' })).toBeInTheDocument()
  })

  it('marca y enfoca la opción prefijada', () => {
    render(
      <ModalAcuerdo valorInicial={3} onConfirmar={vi.fn()} onCerrar={vi.fn()} />,
    )

    const opcion = screen.getByRole('radio', { name: '3 puntos' })
    expect(opcion).toBeChecked()
    expect(opcion).toHaveFocus()
  })

  it('enfoca la primera opción cuando no hay prefill', () => {
    render(
      <ModalAcuerdo valorInicial={null} onConfirmar={vi.fn()} onCerrar={vi.fn()} />,
    )

    expect(screen.getByRole('radio', { name: '0 puntos' })).toHaveFocus()
  })

  it('confirma con el valor elegido', async () => {
    const onConfirmar = vi.fn()
    const usuario = userEvent.setup()
    render(
      <ModalAcuerdo valorInicial={3} onConfirmar={onConfirmar} onCerrar={vi.fn()} />,
    )

    await usuario.click(screen.getByRole('radio', { name: '8 puntos' }))
    await usuario.click(screen.getByRole('button', { name: 'Confirmar acuerdo' }))

    expect(onConfirmar).toHaveBeenCalledWith(8)
  })

  it('sin selección muestra un error y no confirma', async () => {
    const onConfirmar = vi.fn()
    const usuario = userEvent.setup()
    render(
      <ModalAcuerdo valorInicial={null} onConfirmar={onConfirmar} onCerrar={vi.fn()} />,
    )

    await usuario.click(screen.getByRole('button', { name: 'Confirmar acuerdo' }))

    expect(onConfirmar).not.toHaveBeenCalled()
    expect(screen.getByText('Elegí una estimación de la escala.')).toBeInTheDocument()
  })

  it('muestra el error del backend y lo limpia al elegir otra opción', async () => {
    const usuario = userEvent.setup()
    render(
      <ModalAcuerdo
        valorInicial={3}
        error="El valor elegido no es válido. Elegí otro de la escala."
        onConfirmar={vi.fn()}
        onCerrar={vi.fn()}
      />,
    )

    expect(
      screen.getByText('El valor elegido no es válido. Elegí otro de la escala.'),
    ).toBeInTheDocument()

    await usuario.click(screen.getByRole('radio', { name: '5 puntos' }))

    expect(screen.queryByText('Elegí una estimación de la escala.')).not.toBeInTheDocument()
  })

  it('se cierra con Escape sin confirmar', async () => {
    const onConfirmar = vi.fn()
    const onCerrar = vi.fn()
    const usuario = userEvent.setup()
    render(
      <ModalAcuerdo valorInicial={3} onConfirmar={onConfirmar} onCerrar={onCerrar} />,
    )

    await usuario.keyboard('{Escape}')

    expect(onCerrar).toHaveBeenCalledTimes(1)
    expect(onConfirmar).not.toHaveBeenCalled()
  })

  it('se cierra al hacer clic en el fondo', async () => {
    const onCerrar = vi.fn()
    const usuario = userEvent.setup()
    render(
      <ModalAcuerdo valorInicial={3} onConfirmar={vi.fn()} onCerrar={onCerrar} />,
    )

    await usuario.click(screen.getByTestId('fondo-modal'))

    expect(onCerrar).toHaveBeenCalledTimes(1)
  })

  it('deshabilita las opciones y el envío mientras procesa', () => {
    render(
      <ModalAcuerdo valorInicial={3} procesando onConfirmar={vi.fn()} onCerrar={vi.fn()} />,
    )

    screen.getAllByRole('radio').forEach((radio) => {
      expect(radio).toBeDisabled()
    })
    expect(screen.getByRole('button', { name: 'Registrando…' })).toBeDisabled()
  })
})
