import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import type { Historia } from '../../../api/types'
import TablaHistorias, { type ColumnaHistoria } from './TablaHistorias'

const historia1: Historia = {
  id: 1,
  codigo: 'US-01',
  titulo: 'Crear proyecto',
  descripcion: '',
  prioridad: 'Alta',
  estado: 'Backlog',
  story_points: 3,
  esfuerzo_real: null,
  creado_en: '2026-04-01T00:00:00Z',
  actualizado_en: '2026-04-01T00:00:00Z',
}

const historia2: Historia = {
  ...historia1,
  id: 2,
  codigo: 'US-02',
  titulo: 'Ver proyectos',
  estado: 'En Progreso',
  story_points: null,
}

const columnas: ColumnaHistoria[] = [
  { clave: 'codigo', encabezado: 'Código', render: (historia) => historia.codigo },
  { clave: 'titulo', encabezado: 'Título', render: (historia) => historia.titulo },
  { clave: 'estado', encabezado: 'Estado', render: (historia) => historia.estado },
]

describe('TablaHistorias', () => {
  it('muestra un encabezado por columna, en el orden configurado', () => {
    render(<TablaHistorias historias={[historia1]} columnas={columnas} />)

    const encabezados = screen.getAllByRole('columnheader').map((th) => th.textContent)
    expect(encabezados).toEqual(['Código', 'Título', 'Estado'])
  })

  it('muestra una fila por historia con el contenido de cada columna', () => {
    render(<TablaHistorias historias={[historia1, historia2]} columnas={columnas} />)

    const filas = screen.getAllByRole('row').slice(1)
    expect(filas).toHaveLength(2)
    expect(within(filas[0]).getByText('US-01')).toBeInTheDocument()
    expect(within(filas[0]).getByText('Crear proyecto')).toBeInTheDocument()
    expect(within(filas[1]).getByText('US-02')).toBeInTheDocument()
    expect(within(filas[1]).getByText('En Progreso')).toBeInTheDocument()
  })

  it('respeta el orden de las historias recibido', () => {
    render(<TablaHistorias historias={[historia2, historia1]} columnas={columnas} />)

    const filas = screen.getAllByRole('row').slice(1)
    expect(within(filas[0]).getByText('US-02')).toBeInTheDocument()
    expect(within(filas[1]).getByText('US-01')).toBeInTheDocument()
  })

  it('sin acciones no agrega la columna Acciones', () => {
    render(<TablaHistorias historias={[historia1]} columnas={columnas} />)

    expect(screen.queryByRole('columnheader', { name: 'Acciones' })).not.toBeInTheDocument()
  })

  it('con acciones agrega la columna Acciones y las muestra por fila', async () => {
    const alEditar = vi.fn()
    const usuario = userEvent.setup()
    render(
      <TablaHistorias
        historias={[historia1, historia2]}
        columnas={columnas}
        acciones={(historia) => (
          <button type="button" onClick={() => alEditar(historia.id)}>
            Editar {historia.codigo}
          </button>
        )}
      />,
    )

    expect(screen.getByRole('columnheader', { name: 'Acciones' })).toBeInTheDocument()
    await usuario.click(screen.getByRole('button', { name: 'Editar US-02' }))
    expect(alEditar).toHaveBeenCalledWith(2)
  })

  it('declara los encabezados con scope col para lectores de pantalla', () => {
    render(<TablaHistorias historias={[historia1]} columnas={columnas} />)

    for (const encabezado of screen.getAllByRole('columnheader')) {
      expect(encabezado).toHaveAttribute('scope', 'col')
    }
  })
})
