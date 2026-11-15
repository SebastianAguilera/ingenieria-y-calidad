import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import type { Historia, SprintBacklogItem } from '../../../api/types'
import TablaSprintBacklog from './TablaSprintBacklog'

function crearItem(id: number, historia: Partial<Historia>): SprintBacklogItem {
  return {
    id,
    orden: id,
    fecha_asignacion: '2026-05-01T10:00:00Z',
    historia: {
      id: 100 + id,
      codigo: `US-0${id}`,
      titulo: `Historia ${id}`,
      descripcion: '',
      prioridad: 'Media',
      estado: 'En Progreso',
      story_points: 3,
      esfuerzo_real: null,
      creado_en: '2026-04-01T00:00:00Z',
      actualizado_en: '2026-04-01T00:00:00Z',
      ...historia,
    },
  }
}

const items = [
  crearItem(1, { codigo: 'US-01', titulo: 'Crear proyecto', story_points: 3 }),
  crearItem(2, { codigo: 'US-02', titulo: 'Ver proyectos', story_points: null, estado: 'Hecho' }),
]

describe('TablaSprintBacklog', () => {
  it('muestra Código, Título, Estado, Story Points y Acciones', () => {
    render(<TablaSprintBacklog items={items} estadoSprint="Activo" onQuitar={vi.fn()} />)

    const encabezados = screen.getAllByRole('columnheader').map((th) => th.textContent)
    expect(encabezados).toEqual(['Código', 'Título', 'Estado', 'Story Points', 'Acciones'])
  })

  it('muestra los datos de cada historia asignada en el orden recibido', () => {
    render(<TablaSprintBacklog items={items} estadoSprint="Activo" onQuitar={vi.fn()} />)

    const filas = screen.getAllByRole('row').slice(1)
    expect(filas).toHaveLength(2)
    expect(within(filas[0]).getByText('US-01')).toBeInTheDocument()
    expect(within(filas[0]).getByText('Crear proyecto')).toBeInTheDocument()
    expect(within(filas[0]).getByText('En Progreso')).toBeInTheDocument()
    expect(within(filas[0]).getByText('3')).toBeInTheDocument()
    expect(within(filas[1]).getByText('US-02')).toBeInTheDocument()
    expect(within(filas[1]).getByText('Hecho')).toBeInTheDocument()
  })

  it('muestra un guion cuando la historia no tiene story points', () => {
    render(<TablaSprintBacklog items={items} estadoSprint="Activo" onQuitar={vi.fn()} />)

    const filas = screen.getAllByRole('row').slice(1)
    expect(within(filas[1]).getByText('—')).toBeInTheDocument()
  })

  it('ofrece Quitar del sprint por historia y avisa cuál se quiere quitar', async () => {
    const alQuitar = vi.fn()
    const usuario = userEvent.setup()
    render(<TablaSprintBacklog items={items} estadoSprint="Activo" onQuitar={alQuitar} />)

    await usuario.click(screen.getByRole('button', { name: 'Quitar US-02 del sprint' }))

    expect(alQuitar).toHaveBeenCalledTimes(1)
    expect(alQuitar).toHaveBeenCalledWith(items[1].historia)
  })

  it('en un sprint Planificado todavía permite quitar', () => {
    render(<TablaSprintBacklog items={items} estadoSprint="Planificado" onQuitar={vi.fn()} />)

    expect(screen.getAllByRole('button', { name: /Quitar .* del sprint/ })).toHaveLength(2)
  })

  it('en un sprint Cerrado es de solo lectura: sin acciones ni columna Acciones', () => {
    render(<TablaSprintBacklog items={items} estadoSprint="Cerrado" onQuitar={vi.fn()} />)

    expect(screen.queryByRole('button')).not.toBeInTheDocument()
    expect(screen.queryByRole('columnheader', { name: 'Acciones' })).not.toBeInTheDocument()
    expect(screen.getByText('US-01')).toBeInTheDocument()
  })

  it('deshabilita las acciones mientras hay una operación en curso', () => {
    render(
      <TablaSprintBacklog
        items={items}
        estadoSprint="Activo"
        onQuitar={vi.fn()}
        deshabilitado
      />,
    )

    for (const boton of screen.getAllByRole('button')) {
      expect(boton).toBeDisabled()
    }
  })
})
