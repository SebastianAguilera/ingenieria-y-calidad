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
      valor_negocio: null,
      criticidad_tiempo: null,
      reduccion_riesgo: null,
      horas_estimadas: null,
      creado_en: '2026-04-01T00:00:00Z',
      actualizado_en: '2026-04-01T00:00:00Z',
      ...historia,
    },
  }
}

const items = [
  crearItem(1, { codigo: 'US-01', titulo: 'Crear proyecto', story_points: 3, horas_estimadas: 8 }),
  crearItem(2, { codigo: 'US-02', titulo: 'Ver proyectos', story_points: null, estado: 'Hecho', prioridad: 'Alta' }),
]

const baseProps = () => ({
  onQuitar: vi.fn(),
  onCambiar: vi.fn(),
  onEditar: vi.fn(),
  onEstimar: vi.fn(),
})

describe('TablaSprintBacklog', () => {
  it('muestra columnas como backlog y acciones', () => {
    render(<TablaSprintBacklog items={items} estadoSprint="Activo" {...baseProps()} />)

    const encabezados = screen.getAllByRole('columnheader').map((th) => th.textContent)
    expect(encabezados).toEqual([
      'Código',
      'Título',
      'Estado',
      'Prioridad',
      'Estimación',
      'Horas estimadas',
      'Acciones',
    ])
  })

  it('muestra datos y controles inline en sprint no cerrado', () => {
    render(<TablaSprintBacklog items={items} estadoSprint="Activo" {...baseProps()} />)

    const filas = screen.getAllByRole('row').slice(1)
    expect(within(filas[0]).getByText('US-01')).toBeInTheDocument()
    expect(within(filas[0]).getByLabelText('Estado de US-01')).toHaveValue('En Progreso')
    expect(within(filas[0]).getByLabelText('Prioridad de US-01')).toHaveValue('Media')
    expect(within(filas[0]).getByText('3')).toBeInTheDocument()
    expect(within(filas[0]).getByText('8 h')).toBeInTheDocument()
    expect(within(filas[1]).getByText('Sin estimar')).toBeInTheDocument()
  })

  it('dispara cambiar, editar, estimar y quitar', async () => {
    const props = baseProps()
    const usuario = userEvent.setup()
    render(<TablaSprintBacklog items={items} estadoSprint="Activo" {...props} />)

    await usuario.selectOptions(screen.getByLabelText('Estado de US-01'), 'Hecho')
    await usuario.selectOptions(screen.getByLabelText('Prioridad de US-01'), 'Alta')
    await usuario.click(screen.getByRole('button', { name: 'Estimar US-01' }))
    await usuario.click(screen.getByRole('button', { name: 'Editar historia US-01' }))
    await usuario.click(screen.getByRole('button', { name: 'Quitar US-01 del sprint' }))

    expect(props.onCambiar).toHaveBeenCalledWith(items[0].historia, { estado: 'Hecho' })
    expect(props.onCambiar).toHaveBeenCalledWith(items[0].historia, { prioridad: 'Alta' })
    expect(props.onEstimar).toHaveBeenCalledWith(items[0].historia)
    expect(props.onEditar).toHaveBeenCalledWith(items[0].historia)
    expect(props.onQuitar).toHaveBeenCalledWith(items[0].historia)
  })

  it('en sprint cerrado muestra chips de solo lectura y no ofrece quitar', () => {
    render(<TablaSprintBacklog items={items} estadoSprint="Cerrado" {...baseProps()} />)

    expect(screen.queryByLabelText('Estado de US-01')).not.toBeInTheDocument()
    expect(screen.queryByLabelText('Prioridad de US-01')).not.toBeInTheDocument()
    expect(screen.getAllByText('En Progreso').length).toBeGreaterThan(0)
    expect(screen.queryByRole('button', { name: /Quitar .* del sprint/ })).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Estimar US-01' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Editar historia US-01' })).toBeInTheDocument()
  })
})
