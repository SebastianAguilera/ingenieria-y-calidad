import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import type { Historia } from '../../../api/types'
import TableroBacklog from './TableroBacklog'

const historia: Historia = {
  id: 10,
  codigo: 'US-01',
  titulo: 'Registrar horas reales',
  descripcion: 'Permitir registrar horas por usuario.',
  prioridad: 'Alta',
  estado: 'Backlog',
  story_points: 5,
  esfuerzo_real: null,
  creado_en: '2026-10-01T09:15:00Z',
  actualizado_en: '2026-10-01T09:15:00Z',
}

function crearDataTransfer() {
  const data: Record<string, string> = {}
  return {
    setData: (tipo: string, valor: string) => {
      data[tipo] = valor
    },
    getData: (tipo: string) => data[tipo] ?? '',
    effectAllowed: '',
    dropEffect: '',
  }
}

describe('TableroBacklog', () => {
  it('mueve una tarjeta al soltarla en otra columna', () => {
    const onMover = vi.fn()
    render(
      <TableroBacklog
        historias={[historia]}
        onMover={onMover}
        onEditar={vi.fn()}
        guardandoId={null}
      />,
    )

    const tarjeta = screen.getByRole('article', { name: 'Historia US-01' })
    const columnaHecho = screen.getByRole('region', { name: 'Columna Hecho' })
    const dataTransfer = crearDataTransfer()

    fireEvent.dragStart(tarjeta, { dataTransfer })
    fireEvent.dragOver(columnaHecho, { dataTransfer })
    fireEvent.drop(columnaHecho, { dataTransfer })

    expect(onMover).toHaveBeenCalledWith(historia, 'Hecho')
  })

  it('no mueve la tarjeta si se suelta en su misma columna', () => {
    const onMover = vi.fn()
    render(
      <TableroBacklog
        historias={[historia]}
        onMover={onMover}
        onEditar={vi.fn()}
        guardandoId={null}
      />,
    )

    const tarjeta = screen.getByRole('article', { name: 'Historia US-01' })
    const columnaBacklog = screen.getByRole('region', { name: 'Columna Backlog' })
    const dataTransfer = crearDataTransfer()

    fireEvent.dragStart(tarjeta, { dataTransfer })
    fireEvent.drop(columnaBacklog, { dataTransfer })

    expect(onMover).not.toHaveBeenCalled()
  })
})
