import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { listarProyectos } from '../api/proyectos'
import type { ProyectoListado } from '../api/types'
import ProyectosPage from './ProyectosPage'

vi.mock('../api/proyectos', () => ({
  listarProyectos: vi.fn(),
}))

const listarMock = vi.mocked(listarProyectos)

const alfa: ProyectoListado = {
  id: 1,
  nombre: 'Proyecto Alfa',
  fecha_inicio: '2026-01-15',
  fecha_fin: '2026-06-30',
  estado: 'Activo',
  cantidad_usuarios: 3,
  creado_en: '2026-01-01T00:00:00Z',
  actualizado_en: '2026-01-01T00:00:00Z',
}

const beta: ProyectoListado = {
  ...alfa,
  id: 2,
  nombre: 'Proyecto Beta',
  fecha_fin: null,
  estado: 'Cerrado',
}

function renderizar() {
  return render(
    <MemoryRouter initialEntries={['/proyectos']}>
      <Routes>
        <Route path="/proyectos" element={<ProyectosPage />} />
        <Route path="/proyectos/nuevo" element={<p>Pantalla nuevo proyecto</p>} />
        <Route path="/proyectos/:id" element={<p>Detalle del proyecto</p>} />
      </Routes>
    </MemoryRouter>,
  )
}

describe('ProyectosPage', () => {
  beforeEach(() => {
    listarMock.mockReset()
  })

  it('muestra un indicador de carga y no una tabla mientras se cargan los proyectos', () => {
    listarMock.mockReturnValue(new Promise(() => {}))

    renderizar()

    expect(screen.getByText('Cargando proyectos…')).toBeInTheDocument()
    expect(screen.queryByRole('table')).not.toBeInTheDocument()
  })

  it('muestra la tabla con nombre, estado, fecha de inicio y fecha de fin', async () => {
    listarMock.mockResolvedValue([alfa])

    renderizar()

    expect(await screen.findByRole('table')).toBeInTheDocument()
    for (const encabezado of ['Nombre', 'Estado', 'Fecha de inicio', 'Fecha de fin']) {
      expect(screen.getByRole('columnheader', { name: encabezado })).toBeInTheDocument()
    }
    expect(screen.getByText('Proyecto Alfa')).toBeInTheDocument()
    expect(screen.getByText('Activo')).toBeInTheDocument()
    expect(screen.getByText('15/01/2026')).toBeInTheDocument()
    expect(screen.getByText('30/06/2026')).toBeInTheDocument()
  })

  it('muestra "Sin fecha" cuando el proyecto no tiene fecha de fin', async () => {
    listarMock.mockResolvedValue([beta])

    renderizar()

    expect(await screen.findByText('Sin fecha')).toBeInTheDocument()
  })

  it('muestra el estado vacío y navega a /proyectos/nuevo con "Crear proyecto"', async () => {
    listarMock.mockResolvedValue([])
    const usuario = userEvent.setup()

    renderizar()

    expect(await screen.findByText('Todavía no participás en ningún proyecto')).toBeInTheDocument()
    await usuario.click(screen.getByRole('button', { name: 'Crear proyecto' }))

    expect(screen.getByText('Pantalla nuevo proyecto')).toBeInTheDocument()
  })

  it('muestra el error con "Reintentar" y vuelve a cargar al presionarlo', async () => {
    listarMock.mockRejectedValueOnce(new Error('falla'))
    listarMock.mockResolvedValueOnce([alfa])
    const usuario = userEvent.setup()

    renderizar()

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'No se pudieron cargar los proyectos',
    )
    await usuario.click(screen.getByRole('button', { name: 'Reintentar' }))

    expect(await screen.findByText('Proyecto Alfa')).toBeInTheDocument()
    expect(listarMock).toHaveBeenCalledTimes(2)
  })

  it('navega a /proyectos/:id al hacer clic en una fila', async () => {
    listarMock.mockResolvedValue([alfa, beta])
    const usuario = userEvent.setup()

    renderizar()

    await usuario.click(await screen.findByText('Proyecto Beta'))

    expect(screen.getByText('Detalle del proyecto')).toBeInTheDocument()
  })

  it('navega a /proyectos/:id al presionar Enter sobre una fila enfocada', async () => {
    listarMock.mockResolvedValue([alfa])
    const usuario = userEvent.setup()

    renderizar()

    const fila = (await screen.findByText('Proyecto Alfa')).closest('tr') as HTMLElement
    fila.focus()
    await usuario.keyboard('{Enter}')

    await waitFor(() => expect(screen.getByText('Detalle del proyecto')).toBeInTheDocument())
  })
})
