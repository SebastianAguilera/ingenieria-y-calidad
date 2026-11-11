import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import ModalHistoria from './ModalHistoria'

function renderizar(onGuardar = vi.fn(), onCerrar = vi.fn()) {
  render(
    <ModalHistoria
      titulo="Nueva historia"
      textoBoton="Crear historia"
      onGuardar={onGuardar}
      onCerrar={onCerrar}
    />,
  )
  return { onGuardar, onCerrar }
}

describe('ModalHistoria', () => {
  it('valida que el título sea obligatorio sin llamar al backend', async () => {
    const { onGuardar } = renderizar()
    const usuario = userEvent.setup()

    await usuario.click(screen.getByRole('button', { name: 'Crear historia' }))

    expect(screen.getByText('El título de la historia es obligatorio')).toBeInTheDocument()
    expect(onGuardar).not.toHaveBeenCalled()
  })

  it('valida que el título no supere los 200 caracteres', async () => {
    const { onGuardar } = renderizar()
    const usuario = userEvent.setup()

    await usuario.type(screen.getByLabelText('Título'), 'A'.repeat(201))
    await usuario.click(screen.getByRole('button', { name: 'Crear historia' }))

    expect(
      screen.getByText('El título de la historia no puede superar los 200 caracteres'),
    ).toBeInTheDocument()
    expect(onGuardar).not.toHaveBeenCalled()
  })

  it('acepta un título de exactamente 200 caracteres', async () => {
    const onGuardar = vi.fn().mockResolvedValue({ ok: true })
    renderizar(onGuardar)
    const usuario = userEvent.setup()
    const titulo = 'A'.repeat(200)

    fireEvent.change(screen.getByLabelText('Título'), { target: { value: titulo } })
    await usuario.click(screen.getByRole('button', { name: 'Crear historia' }))

    expect(onGuardar).toHaveBeenCalledWith(
      expect.objectContaining({ titulo }),
    )
  })

  it('envía los datos normalizados y usa los valores por defecto', async () => {
    const onGuardar = vi.fn().mockResolvedValue({ ok: true })
    renderizar(onGuardar)
    const usuario = userEvent.setup()

    await usuario.type(screen.getByLabelText('Título'), '  Historia nueva  ')
    await usuario.type(screen.getByLabelText('Descripción'), 'Detalle de la historia.')
    await usuario.click(screen.getByRole('button', { name: 'Crear historia' }))

    expect(onGuardar).toHaveBeenCalledWith({
      titulo: 'Historia nueva',
      descripcion: 'Detalle de la historia.',
      prioridad: 'Media',
      estado: 'Backlog',
    })
  })

  it('muestra el error de campo devuelto por el backend', async () => {
    const onGuardar = vi.fn().mockResolvedValue({
      ok: false,
      mensaje: 'la prioridad de la historia debe ser "Alta", "Media" o "Baja"',
      campo: 'prioridad',
    })
    renderizar(onGuardar)
    const usuario = userEvent.setup()

    await usuario.type(screen.getByLabelText('Título'), 'Historia con error')
    await usuario.click(screen.getByRole('button', { name: 'Crear historia' }))

    expect(
      await screen.findByText('la prioridad de la historia debe ser "Alta", "Media" o "Baja"'),
    ).toBeInTheDocument()
  })

  it('se cierra con Escape', () => {
    const { onCerrar } = renderizar()

    fireEvent.keyDown(document, { key: 'Escape' })

    expect(onCerrar).toHaveBeenCalled()
  })

  it('se cierra al hacer clic fuera del diálogo', () => {
    const { onCerrar } = renderizar()

    fireEvent.mouseDown(screen.getByTestId('fondo-modal'))

    expect(onCerrar).toHaveBeenCalled()
  })
})
