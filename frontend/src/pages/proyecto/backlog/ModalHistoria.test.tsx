import { fireEvent, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import type { Historia } from '../../../api/types'
import ModalHistoria from './ModalHistoria'

const historiaConValores: Historia = {
  id: 10,
  codigo: 'US-01',
  titulo: 'Historia priorizada',
  descripcion: 'Con valores cargados',
  prioridad: 'Alta',
  estado: 'Backlog',
  story_points: 5,
  esfuerzo_real: null,
  valor_negocio: 8,
  criticidad_tiempo: 3,
  reduccion_riesgo: 2,
  creado_en: '2026-11-15T10:00:00Z',
  actualizado_en: '2026-11-15T10:00:00Z',
}

const AYUDA = {
  valorNegocio: 'Valor que aporta la historia al negocio o al cliente.',
  criticidadTiempo: 'Urgencia de entregar la historia en el tiempo.',
  reduccionRiesgo: 'Riesgos u oportunidades que reduce la historia.',
}

function renderizar(onGuardar = vi.fn(), onCerrar = vi.fn(), historia?: Historia) {
  render(
    <ModalHistoria
      titulo="Nueva historia"
      textoBoton="Crear historia"
      historia={historia}
      onGuardar={onGuardar}
      onCerrar={onCerrar}
    />,
  )
  return { onGuardar, onCerrar }
}

const ESCALA = ['Sin asignar', '1', '2', '3', '5', '8', '13', '21']

function opcionesDe(etiqueta: string): string[] {
  return within(screen.getByLabelText(etiqueta))
    .getAllByRole('option')
    .map((opcion) => opcion.textContent ?? '')
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
      valor_negocio: null,
      criticidad_tiempo: null,
      reduccion_riesgo: null,
    })
  })

  it('E-01: muestra los tres campos de priorización con su ayuda visible', () => {
    renderizar()

    const valor = screen.getByLabelText('Valor de negocio')
    const criticidad = screen.getByLabelText('Criticidad en el tiempo')
    const riesgo = screen.getByLabelText('Reducción de riesgo')

    expect(valor).toHaveAccessibleDescription(AYUDA.valorNegocio)
    expect(criticidad).toHaveAccessibleDescription(AYUDA.criticidadTiempo)
    expect(riesgo).toHaveAccessibleDescription(AYUDA.reduccionRiesgo)
  })

  it('E-02: los selectores solo ofrecen Sin asignar y la escala Fibonacci', () => {
    renderizar()

    expect(opcionesDe('Valor de negocio')).toEqual(ESCALA)
    expect(opcionesDe('Criticidad en el tiempo')).toEqual(ESCALA)
    expect(opcionesDe('Reducción de riesgo')).toEqual(ESCALA)
  })

  it('E-03: envía los tres valores seleccionados', async () => {
    const onGuardar = vi.fn().mockResolvedValue({ ok: true })
    renderizar(onGuardar)
    const usuario = userEvent.setup()

    await usuario.type(screen.getByLabelText('Título'), 'Historia priorizada')
    await usuario.selectOptions(screen.getByLabelText('Valor de negocio'), '8')
    await usuario.selectOptions(screen.getByLabelText('Criticidad en el tiempo'), '3')
    await usuario.selectOptions(screen.getByLabelText('Reducción de riesgo'), '2')
    await usuario.click(screen.getByRole('button', { name: 'Crear historia' }))

    expect(onGuardar).toHaveBeenCalledWith(
      expect.objectContaining({
        valor_negocio: 8,
        criticidad_tiempo: 3,
        reduccion_riesgo: 2,
      }),
    )
  })

  it('E-04: carga los valores existentes y deshabilita Sin asignar', () => {
    renderizar(vi.fn(), vi.fn(), historiaConValores)

    for (const [etiqueta, valor] of [
      ['Valor de negocio', '8'],
      ['Criticidad en el tiempo', '3'],
      ['Reducción de riesgo', '2'],
    ] as const) {
      const selector = screen.getByLabelText(etiqueta)
      expect(selector).toHaveValue(valor)
      expect(within(selector).getByRole('option', { name: 'Sin asignar' })).toBeDisabled()
    }
  })

  it('E-08: muestra el 422 en el campo de priorización correspondiente', async () => {
    const campo = 'criticidad_tiempo'
    const mensaje = 'la criticidad de tiempo debe ser un número de la escala'
    const onGuardar = vi.fn().mockResolvedValue({ ok: false, mensaje, campo })
    renderizar(onGuardar)
    const usuario = userEvent.setup()

    await usuario.type(screen.getByLabelText('Título'), 'Historia con error')
    await usuario.selectOptions(screen.getByLabelText('Criticidad en el tiempo'), '5')
    await usuario.click(screen.getByRole('button', { name: 'Crear historia' }))

    expect(await screen.findByText(mensaje)).toBeInTheDocument()
    expect(onGuardar).toHaveBeenCalled()
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
