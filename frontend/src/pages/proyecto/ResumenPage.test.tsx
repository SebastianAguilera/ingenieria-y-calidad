import { render, screen, within } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { Proyecto } from '../../api/types'
import { useProyecto } from './proyectoContext'
import ResumenPage from './ResumenPage'

vi.mock('./proyectoContext', () => ({
  useProyecto: vi.fn(),
}))

const useProyectoMock = vi.mocked(useProyecto)

const alfa: Proyecto = {
  id: 1,
  nombre: 'Proyecto Alfa',
  fecha_inicio: '2026-01-15',
  fecha_fin: '2026-06-30',
  estado: 'Activo',
  cantidad_usuarios: 3,
  usuarios: [],
  creado_en: '2026-01-01T00:00:00Z',
  actualizado_en: '2026-01-01T00:00:00Z',
}

function renderizar(proyecto: Proyecto) {
  useProyectoMock.mockReturnValue({
    estado: { tipo: 'listo', proyecto },
    reintentar: vi.fn(),
  })
  render(<ResumenPage />)
}

describe('ResumenPage', () => {
  beforeEach(() => {
    useProyectoMock.mockReset()
  })

  it('muestra la tabla campo/valor con los cinco campos del proyecto', () => {
    renderizar(alfa)

    expect(screen.getByRole('table')).toBeInTheDocument()
    expect(screen.getByRole('columnheader', { name: 'Campo' })).toBeInTheDocument()
    expect(screen.getByRole('columnheader', { name: 'Valor' })).toBeInTheDocument()
    expect(screen.getByRole('rowheader', { name: 'Nombre' })).toBeInTheDocument()
    expect(screen.getByRole('rowheader', { name: 'Estado' })).toBeInTheDocument()
    expect(screen.getByRole('rowheader', { name: 'Fecha de inicio' })).toBeInTheDocument()
    expect(screen.getByRole('rowheader', { name: 'Fecha de fin' })).toBeInTheDocument()
    expect(screen.getByRole('rowheader', { name: 'Cantidad de integrantes' })).toBeInTheDocument()
    expect(screen.getByText('Proyecto Alfa')).toBeInTheDocument()
    expect(screen.getByText('Activo')).toBeInTheDocument()
    expect(screen.getByText('15/01/2026')).toBeInTheDocument()
    expect(screen.getByText('30/06/2026')).toBeInTheDocument()
    expect(screen.getByText('3')).toBeInTheDocument()
  })

  it('muestra "Sin fecha" en la fila Fecha de fin cuando el proyecto no tiene fecha de fin', () => {
    renderizar({ ...alfa, fecha_fin: null })

    const filaFechaFin = screen.getByRole('row', { name: /Fecha de fin/ })
    expect(within(filaFechaFin).getByText('Sin fecha')).toBeInTheDocument()
  })

  it('muestra cantidad_usuarios tal como la devuelve la API, sin recalcularla', () => {
    renderizar({ ...alfa, cantidad_usuarios: 7, usuarios: [] })

    expect(screen.getByText('7')).toBeInTheDocument()
  })

  it('estructura la tabla con thead y tbody accesibles', () => {
    renderizar(alfa)

    expect(screen.getAllByRole('rowgroup')).toHaveLength(2)
    expect(screen.getByRole('table')).toBeInTheDocument()
  })
})
