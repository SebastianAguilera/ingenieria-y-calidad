import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiError } from '../api/errors'
import { obtenerProyecto } from '../api/proyectos'
import type { Proyecto } from '../api/types'
import BarraLateralProyecto from './BarraLateralProyecto'
import ProyectoLayout from './ProyectoLayout'

vi.mock('../api/proyectos', () => ({
  obtenerProyecto: vi.fn(),
}))

const obtenerMock = vi.mocked(obtenerProyecto)

const alfa: Proyecto = {
  id: 1,
  nombre: 'Proyecto Alfa',
  fecha_inicio: '2026-01-15',
  fecha_fin: '2026-06-30',
  estado: 'Activo',
  cantidad_usuarios: 3,
  usuarios: [
    { id: 10, nombre: 'Ana Pérez', email: 'ana@ejemplo.com', creado_en: '2026-01-01T00:00:00Z' },
  ],
  creado_en: '2026-01-01T00:00:00Z',
  actualizado_en: '2026-01-01T00:00:00Z',
}

function renderizar(entrada = '/proyectos/1', conMenu = false) {
  return render(
    <MemoryRouter initialEntries={[entrada]}>
      {conMenu && <BarraLateralProyecto />}
      <Routes>
        <Route path="/proyectos" element={<p>Listado de proyectos</p>} />
        <Route path="/proyectos/:id" element={<ProyectoLayout />}>
          <Route index element={<p>Contenido de Resumen</p>} />
          <Route path="backlog" element={<p>Contenido de Backlog</p>} />
          <Route path="equipo" element={<p>Contenido de Equipo</p>} />
        </Route>
      </Routes>
    </MemoryRouter>,
  )
}

describe('ProyectoLayout', () => {
  beforeEach(() => {
    obtenerMock.mockReset()
  })

  it('muestra "Cargando proyecto…" con role status mientras resuelve, sin contenido de sección', () => {
    obtenerMock.mockReturnValue(new Promise(() => {}))

    renderizar()

    expect(screen.getByRole('status')).toHaveTextContent('Cargando proyecto…')
    expect(screen.queryByText('Contenido de Resumen')).not.toBeInTheDocument()
  })

  it('muestra el nombre del proyecto en el encabezado y el contenido de la sección', async () => {
    obtenerMock.mockResolvedValue(alfa)

    renderizar()

    expect(await screen.findByRole('heading', { name: 'Proyecto Alfa' })).toBeInTheDocument()
    expect(screen.getByText('Contenido de Resumen')).toBeInTheDocument()
    expect(obtenerMock).toHaveBeenCalledWith(1, expect.any(AbortSignal))
  })

  it('no muestra el chip "Cerrado" cuando el proyecto está activo', async () => {
    obtenerMock.mockResolvedValue(alfa)

    renderizar()

    await screen.findByRole('heading', { name: 'Proyecto Alfa' })
    expect(screen.queryByText('Cerrado')).not.toBeInTheDocument()
  })

  it('muestra el chip "Cerrado" en el encabezado cuando el proyecto está cerrado', async () => {
    obtenerMock.mockResolvedValue({ ...alfa, estado: 'Cerrado' })

    renderizar()

    expect(await screen.findByText('Cerrado')).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Proyecto Alfa' })).toBeInTheDocument()
  })

  it('muestra el error con "Reintentar" y recupera el detalle al reintentar', async () => {
    obtenerMock.mockRejectedValueOnce(
      new ApiError({ status: 500, code: 'ERROR_INTERNO', message: 'boom' }),
    )
    obtenerMock.mockResolvedValueOnce(alfa)
    const usuario = userEvent.setup()

    renderizar()

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'No se pudo cargar el proyecto. Intentá nuevamente.',
    )
    expect(screen.queryByText('Contenido de Resumen')).not.toBeInTheDocument()
    await usuario.click(screen.getByRole('button', { name: 'Reintentar' }))

    expect(await screen.findByRole('heading', { name: 'Proyecto Alfa' })).toBeInTheDocument()
    expect(obtenerMock).toHaveBeenCalledTimes(2)
  })

  it('muestra "Proyecto no encontrado" con enlace a /proyectos ante un 404 de la API', async () => {
    obtenerMock.mockRejectedValueOnce(
      new ApiError({ status: 404, code: 'PROYECTO_NO_ENCONTRADO', message: 'el proyecto no existe' }),
    )
    const usuario = userEvent.setup()

    renderizar()

    expect(await screen.findByText('Proyecto no encontrado')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Reintentar' })).not.toBeInTheDocument()
    expect(screen.queryByText('Contenido de Resumen')).not.toBeInTheDocument()
    await usuario.click(screen.getByRole('link', { name: 'Volver a proyectos' }))

    expect(screen.getByText('Listado de proyectos')).toBeInTheDocument()
  })

  it('trata un id no numérico como no encontrado sin llamar a la API', async () => {
    renderizar('/proyectos/abc')

    expect(await screen.findByText('Proyecto no encontrado')).toBeInTheDocument()
    expect(obtenerMock).not.toHaveBeenCalled()
  })

  it('trata un id no positivo como no encontrado sin llamar a la API', async () => {
    renderizar('/proyectos/0')

    expect(await screen.findByText('Proyecto no encontrado')).toBeInTheDocument()
    expect(obtenerMock).not.toHaveBeenCalled()
  })

  it('cambia de sección sin recargar la página ni volver a pedir el proyecto', async () => {
    obtenerMock.mockResolvedValue(alfa)
    const usuario = userEvent.setup()

    renderizar('/proyectos/1', true)

    await screen.findByRole('heading', { name: 'Proyecto Alfa' })
    await usuario.click(screen.getByRole('link', { name: 'Equipo' }))

    expect(screen.getByText('Contenido de Equipo')).toBeInTheDocument()
    expect(screen.queryByText('Contenido de Resumen')).not.toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Proyecto Alfa' })).toBeInTheDocument()
    expect(obtenerMock).toHaveBeenCalledTimes(1)
  })

  it('carga el proyecto en un deep link directo a una sección', async () => {
    obtenerMock.mockResolvedValue(alfa)

    renderizar('/proyectos/1/equipo')

    expect(await screen.findByRole('heading', { name: 'Proyecto Alfa' })).toBeInTheDocument()
    expect(screen.getByText('Contenido de Equipo')).toBeInTheDocument()
    expect(obtenerMock).toHaveBeenCalledTimes(1)
  })

  it('no muestra el contenido de la sección en un deep link a un proyecto inexistente', async () => {
    obtenerMock.mockRejectedValueOnce(
      new ApiError({ status: 404, code: 'PROYECTO_NO_ENCONTRADO', message: 'el proyecto no existe' }),
    )

    renderizar('/proyectos/999/equipo')

    expect(await screen.findByText('Proyecto no encontrado')).toBeInTheDocument()
    expect(screen.queryByText('Contenido de Equipo')).not.toBeInTheDocument()
  })
})
