import { render, screen, within } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { Proyecto } from '../../api/types'
import { useProyecto } from './proyectoContext'
import EquipoPage from './EquipoPage'

vi.mock('./proyectoContext', () => ({
  useProyecto: vi.fn(),
}))

const useProyectoMock = vi.mocked(useProyecto)

const proyectoConEquipo: Proyecto = {
  id: 1,
  nombre: 'Proyecto Alfa',
  fecha_inicio: '2026-01-15',
  fecha_fin: '2026-06-30',
  estado: 'Activo',
  cantidad_usuarios: 2,
  usuarios: [
    { id: 10, nombre: 'Ana Pérez', email: 'ana@ejemplo.com', creado_en: '2026-01-01T00:00:00Z' },
    { id: 11, nombre: 'Luis Gómez', email: 'luis@ejemplo.com', creado_en: '2025-12-20T10:00:00Z' },
  ],
  creado_en: '2026-01-01T00:00:00Z',
  actualizado_en: '2026-01-01T00:00:00Z',
}

function renderizar(proyecto: Proyecto) {
  useProyectoMock.mockReturnValue({
    estado: { tipo: 'listo', proyecto },
    reintentar: vi.fn(),
  })
  render(<EquipoPage />)
}

describe('EquipoPage', () => {
  beforeEach(() => {
    useProyectoMock.mockReset()
  })

  it('muestra una fila por integrante con nombre, email y fecha de registro', () => {
    renderizar(proyectoConEquipo)

    expect(screen.getByRole('table')).toBeInTheDocument()
    expect(screen.getByRole('columnheader', { name: 'Nombre' })).toBeInTheDocument()
    expect(screen.getByRole('columnheader', { name: 'Email' })).toBeInTheDocument()
    expect(screen.getByRole('columnheader', { name: 'Registrado el' })).toBeInTheDocument()

    const cuerpo = screen.getAllByRole('rowgroup')[1]
    const filas = within(cuerpo).getAllByRole('row')
    expect(filas).toHaveLength(2)

    expect(screen.getByText('Ana Pérez')).toBeInTheDocument()
    expect(screen.getByText('ana@ejemplo.com')).toBeInTheDocument()
    expect(screen.getByText('01/01/2026')).toBeInTheDocument()
    expect(screen.getByText('Luis Gómez')).toBeInTheDocument()
    expect(screen.getByText('luis@ejemplo.com')).toBeInTheDocument()
    expect(screen.getByText('20/12/2025')).toBeInTheDocument()
  })

  it('muestra el estado vacío y no la tabla cuando no hay integrantes', () => {
    renderizar({ ...proyectoConEquipo, usuarios: [], cantidad_usuarios: 0 })

    expect(screen.getByText('Este proyecto todavía no tiene integrantes')).toBeInTheDocument()
    expect(screen.queryByRole('table')).not.toBeInTheDocument()
  })

  it('muestra la fecha de registro del usuario en formato dd/mm/yyyy', () => {
    renderizar(proyectoConEquipo)

    expect(screen.getByText('01/01/2026')).toBeInTheDocument()
    expect(screen.getByText('20/12/2025')).toBeInTheDocument()
  })
})
