import { render, screen, within } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import type { Historia } from '../../../api/types'
import TablaBacklog from './TablaBacklog'

const estimada: Historia = {
  id: 1,
  codigo: 'US-01',
  titulo: 'Crear proyecto',
  descripcion: '',
  prioridad: 'Alta',
  estado: 'Backlog',
  story_points: 5,
  esfuerzo_real: null,
  creado_en: '2026-04-01T00:00:00Z',
  actualizado_en: '2026-04-01T00:00:00Z',
}

const sinEstimar: Historia = {
  ...estimada,
  id: 2,
  codigo: 'US-02',
  titulo: 'Ver proyectos',
  story_points: null,
}

function renderizar(historias: Historia[]) {
  return render(
    <TablaBacklog
      historias={historias}
      onCambiar={vi.fn()}
      onEditar={vi.fn()}
      guardandoId={null}
    />,
  )
}

describe('TablaBacklog', () => {
  it('agrega la columna Estimación junto a Prioridad', () => {
    renderizar([estimada])

    const encabezados = screen.getAllByRole('columnheader').map((th) => th.textContent)
    expect(encabezados).toEqual([
      'Código',
      'Título',
      'Estado',
      'Prioridad',
      'Estimación',
      'Acciones',
    ])
  })

  it('muestra los story points cuando la historia está estimada', () => {
    renderizar([estimada])

    const fila = screen.getAllByRole('row')[1]
    expect(within(fila).getByText('5')).toBeInTheDocument()
  })

  it('indica "Sin estimar" cuando la historia no tiene story points', () => {
    renderizar([sinEstimar])

    const fila = screen.getAllByRole('row')[1]
    expect(within(fila).getByText('Sin estimar')).toBeInTheDocument()
  })
})
