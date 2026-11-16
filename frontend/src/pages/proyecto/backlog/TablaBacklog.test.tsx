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
  valor_negocio: null,
  criticidad_tiempo: null,
  reduccion_riesgo: null,
  horas_estimadas: 5,
  creado_en: '2026-04-01T00:00:00Z',
  actualizado_en: '2026-04-01T00:00:00Z',
}

const sinEstimar: Historia = {
  ...estimada,
  id: 2,
  codigo: 'US-02',
  titulo: 'Ver proyectos',
  story_points: null,
  horas_estimadas: null,
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
  it('agrega las columnas Estimación y Horas estimadas junto a Prioridad', () => {
    renderizar([estimada])

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

  it('E-18 muestra las horas estimadas de la historia', () => {
    renderizar([estimada])

    const fila = screen.getAllByRole('row')[1]
    expect(within(fila).getByText('5 h')).toBeInTheDocument()
  })

  it('E-20 muestra "—" cuando la historia no tiene horas estimadas', () => {
    renderizar([sinEstimar])

    const fila = screen.getAllByRole('row')[1]
    expect(within(fila).getByText('—')).toBeInTheDocument()
  })
})

describe('TablaBacklog - alineación de columnas', () => {
  function celdaDe(encabezado: string): { th: HTMLElement; td: HTMLElement } {
    const th = screen.getByRole('columnheader', { name: encabezado })
    const fila = screen.getAllByRole('row')[1] as HTMLTableRowElement
    return { th, td: fila.cells[(th as HTMLTableCellElement).cellIndex] }
  }

  it.each(['Código', 'Título', 'Estado', 'Prioridad', 'Estimación', 'Horas estimadas'])(
    '%s: el encabezado y el contenido están centrados',
    (encabezado) => {
      renderizar([estimada])

      const { th, td } = celdaDe(encabezado)
      expect(th).toHaveClass('text-center')
      expect(td).toHaveClass('text-center')
    },
  )

  it('Acciones: el encabezado y los botones comparten la clase de columna centrada', () => {
    renderizar([estimada])

    const { th, td } = celdaDe('Acciones')
    expect(th).toHaveClass('col-acciones')
    expect(td).toHaveClass('col-acciones')
  })

  it('Horas estimadas: el guion de una historia sin horas queda centrado bajo el título', () => {
    renderizar([sinEstimar])

    const { th, td } = celdaDe('Horas estimadas')
    expect(td).toHaveTextContent('—')
    expect(td).toHaveClass('text-center')
    expect(th).toHaveClass('text-center')
  })

  it('los encabezados de la tabla de ancho fijo pueden envolverse para no desalinearse del contenido', () => {
    renderizar([estimada])

    for (const encabezado of ['Estimación', 'Horas estimadas']) {
      expect(screen.getByRole('columnheader', { name: encabezado }).className).toContain(
        '!whitespace-normal',
      )
    }
  })
})
