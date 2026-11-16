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
  horas_estimadas: null,
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
      horas_estimadas: null,
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

const historiaConHoras: Historia = {
  id: 4,
  codigo: 'US-04',
  titulo: 'Historia estimada',
  descripcion: '',
  prioridad: 'Alta',
  estado: 'En Progreso',
  story_points: 5,
  esfuerzo_real: null,
  horas_estimadas: 8,
  valor_negocio: null,
  criticidad_tiempo: null,
  reduccion_riesgo: null,
  creado_en: '2026-04-01T00:00:00Z',
  actualizado_en: '2026-04-01T00:00:00Z',
}

function renderizarEdicion(historia: Historia, onGuardar = vi.fn()) {
  render(
    <ModalHistoria
      titulo="Editar historia"
      textoBoton="Guardar cambios"
      historia={historia}
      onGuardar={onGuardar}
      onCerrar={vi.fn()}
    />,
  )
  return { onGuardar }
}

describe('ModalHistoria - horas estimadas (US-F21 tramo 2)', () => {
  it('E-17 envía las horas estimadas cuando son mayores a 0', async () => {
    const onGuardar = vi.fn().mockResolvedValue({ ok: true })
    renderizar(onGuardar)
    const usuario = userEvent.setup()

    await usuario.type(screen.getByLabelText('Título'), 'Historia con horas')
    await usuario.type(screen.getByLabelText('Horas estimadas'), '5')
    await usuario.click(screen.getByRole('button', { name: 'Crear historia' }))

    expect(onGuardar).toHaveBeenCalledWith(
      expect.objectContaining({ titulo: 'Historia con horas', horas_estimadas: 5 }),
    )
  })

  it('E-19 en el alta envía horas_estimadas null cuando el campo está vacío', async () => {
    const onGuardar = vi.fn().mockResolvedValue({ ok: true })
    renderizar(onGuardar)
    const usuario = userEvent.setup()

    await usuario.type(screen.getByLabelText('Título'), 'Sin horas')
    await usuario.click(screen.getByRole('button', { name: 'Crear historia' }))

    expect(onGuardar).toHaveBeenCalledWith(expect.objectContaining({ horas_estimadas: null }))
  })

  it.each(['0', '-3'])(
    'E-21 rechaza %s en el cliente, marca el campo y no llama a la API',
    async (valor) => {
      const { onGuardar } = renderizar()
      const usuario = userEvent.setup()

      await usuario.type(screen.getByLabelText('Título'), 'Historia')
      await usuario.type(screen.getByLabelText('Horas estimadas'), valor)
      await usuario.click(screen.getByRole('button', { name: 'Crear historia' }))

      const campo = screen.getByLabelText('Horas estimadas')
      expect(screen.getByText('Las horas estimadas deben ser mayores a 0')).toBeInTheDocument()
      expect(campo).toHaveAttribute('aria-invalid', 'true')
      expect(campo).toHaveAccessibleDescription(
        expect.stringContaining('Las horas estimadas deben ser mayores a 0'),
      )
      expect(onGuardar).not.toHaveBeenCalled()
    },
  )

  it('E-22 acepta un cuarto de hora', async () => {
    const onGuardar = vi.fn().mockResolvedValue({ ok: true })
    renderizar(onGuardar)
    const usuario = userEvent.setup()

    await usuario.type(screen.getByLabelText('Título'), 'Historia corta')
    await usuario.type(screen.getByLabelText('Horas estimadas'), '0.25')
    await usuario.click(screen.getByRole('button', { name: 'Crear historia' }))

    expect(onGuardar).toHaveBeenCalledWith(expect.objectContaining({ horas_estimadas: 0.25 }))
  })

  it('E-23 muestra en el campo el 422 del backend sobre horas_estimadas', async () => {
    const onGuardar = vi.fn().mockResolvedValue({
      ok: false,
      mensaje: 'las horas estimadas deben ser mayores a 0',
      campo: 'horas_estimadas',
    })
    renderizar(onGuardar)
    const usuario = userEvent.setup()

    await usuario.type(screen.getByLabelText('Título'), 'Historia')
    await usuario.type(screen.getByLabelText('Horas estimadas'), '3')
    await usuario.click(screen.getByRole('button', { name: 'Crear historia' }))

    const mensaje = await screen.findByText('las horas estimadas deben ser mayores a 0')
    expect(screen.getByLabelText('Horas estimadas')).toHaveAttribute('aria-invalid', 'true')
    expect(mensaje.id).toBe(
      screen.getByLabelText('Horas estimadas').getAttribute('aria-describedby')?.split(' ')[0],
    )
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('E-24 precarga las horas estimadas al editar', () => {
    renderizarEdicion(historiaConHoras)

    expect(screen.getByLabelText('Horas estimadas')).toHaveValue(8)
  })

  it('al editar con el campo vacío no envía horas_estimadas', async () => {
    const onGuardar = vi.fn().mockResolvedValue({ ok: true })
    renderizarEdicion({ ...historiaConHoras, horas_estimadas: null }, onGuardar)
    const usuario = userEvent.setup()

    await usuario.click(screen.getByRole('button', { name: 'Guardar cambios' }))

    expect(onGuardar).toHaveBeenCalledTimes(1)
    expect(onGuardar.mock.calls[0][0]).not.toHaveProperty('horas_estimadas')
  })

  it('al editar envía el nuevo valor de horas estimadas', async () => {
    const onGuardar = vi.fn().mockResolvedValue({ ok: true })
    renderizarEdicion(historiaConHoras, onGuardar)
    const usuario = userEvent.setup()

    const campo = screen.getByLabelText('Horas estimadas')
    await usuario.clear(campo)
    await usuario.type(campo, '12')
    await usuario.click(screen.getByRole('button', { name: 'Guardar cambios' }))

    expect(onGuardar).toHaveBeenCalledWith(expect.objectContaining({ horas_estimadas: 12 }))
  })
})
