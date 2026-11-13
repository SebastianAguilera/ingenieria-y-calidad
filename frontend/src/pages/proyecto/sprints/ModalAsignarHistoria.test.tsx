import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import type { Historia } from '../../../api/types'
import ModalAsignarHistoria from './ModalAsignarHistoria'

function crearHistoria(id: number, codigo: string, titulo: string): Historia {
  return {
    id,
    codigo,
    titulo,
    descripcion: '',
    prioridad: 'Media',
    estado: 'Backlog',
    story_points: null,
    esfuerzo_real: null,
    creado_en: '2026-04-01T00:00:00Z',
    actualizado_en: '2026-04-01T00:00:00Z',
  }
}

const historias = [
  crearHistoria(5, 'US-05', 'Exportar informe'),
  crearHistoria(6, 'US-06', 'Ver métricas'),
]

function renderizar(propiedades: Partial<Parameters<typeof ModalAsignarHistoria>[0]> = {}) {
  const alAsignar = vi.fn()
  const alCerrar = vi.fn()
  render(
    <ModalAsignarHistoria
      historias={historias}
      onAsignar={alAsignar}
      onCerrar={alCerrar}
      {...propiedades}
    />,
  )
  return { alAsignar, alCerrar }
}

describe('ModalAsignarHistoria', () => {
  it('es un diálogo modal accesible titulado Asignar historia', () => {
    renderizar()

    const dialogo = screen.getByRole('dialog', { name: 'Asignar historia' })
    expect(dialogo).toHaveAttribute('aria-modal', 'true')
  })

  it('lista las historias disponibles para elegir una', () => {
    renderizar()

    const dialogo = screen.getByRole('dialog')
    expect(within(dialogo).getByRole('radio', { name: /US-05.*Exportar informe/ })).toBeInTheDocument()
    expect(within(dialogo).getByRole('radio', { name: /US-06.*Ver métricas/ })).toBeInTheDocument()
  })

  it('no permite asignar hasta elegir una historia', () => {
    renderizar()

    expect(screen.getByRole('button', { name: 'Asignar' })).toBeDisabled()
  })

  it('al elegir una historia y confirmar entrega su id', async () => {
    const usuario = userEvent.setup()
    const { alAsignar } = renderizar()

    await usuario.click(screen.getByRole('radio', { name: /US-06/ }))
    await usuario.click(screen.getByRole('button', { name: 'Asignar' }))

    expect(alAsignar).toHaveBeenCalledTimes(1)
    expect(alAsignar).toHaveBeenCalledWith(6)
  })

  it('sin historias disponibles muestra el aviso y no permite asignar', () => {
    renderizar({ historias: [] })

    expect(screen.getByText('No hay historias disponibles para asignar')).toBeInTheDocument()
    expect(screen.queryByRole('radio')).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Asignar' })).toBeDisabled()
  })

  it('muestra el error recibido como alerta sin cerrar el modal', () => {
    renderizar({ error: 'La historia ya está asignada a un Sprint' })

    expect(screen.getByRole('alert')).toHaveTextContent('La historia ya está asignada a un Sprint')
    expect(screen.getByRole('dialog')).toBeInTheDocument()
  })

  it('mientras envía deshabilita las acciones', async () => {
    const usuario = userEvent.setup()
    renderizar({ enviando: true })

    expect(screen.getByRole('button', { name: 'Asignar' })).toBeDisabled()
    expect(screen.getByRole('button', { name: 'Cancelar' })).toBeDisabled()
    await usuario.click(screen.getByRole('radio', { name: /US-05/ }))
    expect(screen.getByRole('button', { name: 'Asignar' })).toBeDisabled()
  })

  it('se cierra con Cancelar', async () => {
    const usuario = userEvent.setup()
    const { alCerrar } = renderizar()

    await usuario.click(screen.getByRole('button', { name: 'Cancelar' }))

    expect(alCerrar).toHaveBeenCalledTimes(1)
  })

  it('se cierra con Escape', async () => {
    const usuario = userEvent.setup()
    const { alCerrar } = renderizar()

    await usuario.keyboard('{Escape}')

    expect(alCerrar).toHaveBeenCalledTimes(1)
  })

  it('recibe el foco al abrirse', () => {
    renderizar()

    expect(screen.getByRole('dialog')).toContainElement(document.activeElement as HTMLElement)
  })

  it('se puede operar con el teclado: elegir con espacio y confirmar con Enter', async () => {
    const usuario = userEvent.setup()
    const { alAsignar } = renderizar()

    screen.getByRole('radio', { name: /US-05/ }).focus()
    await usuario.keyboard(' ')
    screen.getByRole('button', { name: 'Asignar' }).focus()
    await usuario.keyboard('{Enter}')

    expect(alAsignar).toHaveBeenCalledWith(5)
  })
})
