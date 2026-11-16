import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { listarIntegrantes, obtenerProyecto } from '../../api/proyectos'
import { ApiError } from '../../api/errors'
import type { Proyecto, UsuarioProyecto } from '../../api/types'
import InformacionGeneralPage from './InformacionGeneralPage'

vi.mock('../../api/proyectos', () => ({
  obtenerProyecto: vi.fn(),
  listarIntegrantes: vi.fn(),
  actualizarProyecto: vi.fn(),
  cambiarEstadoProyecto: vi.fn(),
  eliminarProyecto: vi.fn(),
}))

const obtenerMock = vi.mocked(obtenerProyecto)
const integrantesMock = vi.mocked(listarIntegrantes)

const proyecto: Proyecto = {
  id: 1,
  nombre: 'Proyecto Alfa',
  descripcion: 'Línea uno\nLínea dos',
  fecha_inicio: '2026-01-15',
  fecha_fin: '2026-06-30',
  estado: 'Activo',
  cantidad_usuarios: 2,
  usuarios: [],
  creado_en: '2026-01-02T00:00:00Z',
  actualizado_en: '2026-01-02T00:00:00Z',
}

const equipo: UsuarioProyecto[] = [
  { id: 2, nombre: 'Bruno Díaz', email: 'bruno@x.com', rol: 'MIEMBRO' },
  { id: 1, nombre: 'Ana Ruiz', email: 'ana@x.com', rol: 'LIDER' },
]

function renderizar() {
  return render(
    <MemoryRouter initialEntries={['/proyectos/1']}>
      <Routes>
        <Route path="/proyectos/:id" element={<InformacionGeneralPage />} />
        <Route path="/proyectos/:id/editar" element={<p>Pantalla de edición</p>} />
        <Route path="/proyectos/:id/equipo" element={<p>Pantalla de equipo</p>} />
      </Routes>
    </MemoryRouter>,
  )
}

describe('InformacionGeneralPage', () => {
  beforeEach(() => {
    obtenerMock.mockReset()
    integrantesMock.mockReset()
    integrantesMock.mockResolvedValue(equipo)
  })

  it('muestra el nombre, el estado, el rango de fechas y la descripción multilínea', async () => {
    obtenerMock.mockResolvedValue(proyecto)

    renderizar()

    expect(await screen.findByRole('heading', { name: 'Proyecto Alfa' })).toBeInTheDocument()
    expect(screen.getByText('15/01/2026 → 30/06/2026')).toBeInTheDocument()
    expect(screen.getByText(/Línea uno\s+Línea dos/, { selector: 'p' })).toBeInTheDocument()
  })

  it('resume duración, fechas e integrantes en la tarjeta de resumen', async () => {
    obtenerMock.mockResolvedValue(proyecto)

    renderizar()

    const resumen = (await screen.findByRole('heading', { name: 'Resumen' })).closest('section')!
    expect(within(resumen).getByText('167 días')).toBeInTheDocument()
    expect(within(resumen).getByText('15/01/2026')).toBeInTheDocument()
    expect(within(resumen).getByText('30/06/2026')).toBeInTheDocument()
    expect(within(resumen).getByText('2')).toBeInTheDocument()
  })

  it('muestra "No disponible" como creador mientras el backend no lo informe', async () => {
    obtenerMock.mockResolvedValue(proyecto)

    renderizar()

    expect(await screen.findByText('Creado por')).toBeInTheDocument()
    expect(screen.getByText('No disponible')).toBeInTheDocument()
  })

  it('muestra el nombre del creador cuando el backend lo informa', async () => {
    obtenerMock.mockResolvedValue({ ...proyecto, creado_por: { id: 1, nombre: 'Ana Ruiz' } })

    renderizar()

    expect(await screen.findByText('Ana Ruiz', { selector: 'dd' })).toBeInTheDocument()
  })

  it('lista el equipo con su rol y pone primero a los líderes', async () => {
    obtenerMock.mockResolvedValue(proyecto)

    renderizar()

    const lista = await screen.findByRole('list', { name: 'Integrantes del proyecto' })
    const items = within(lista).getAllByRole('listitem')
    expect(items[0]).toHaveTextContent('Ana Ruiz')
    expect(items[0]).toHaveTextContent('Líder')
    expect(items[1]).toHaveTextContent('Bruno Díaz')
    expect(items[1]).toHaveTextContent('Miembro')
  })

  it('indica cuando el proyecto todavía no tiene integrantes', async () => {
    obtenerMock.mockResolvedValue(proyecto)
    integrantesMock.mockResolvedValue([])

    renderizar()

    expect(await screen.findByText('Todavía no hay integrantes en este proyecto.')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Agregar integrantes' })).toHaveAttribute(
      'href',
      '/proyectos/1/equipo',
    )
  })

  it('muestra el proyecto aunque falle la carga del equipo', async () => {
    obtenerMock.mockResolvedValue(proyecto)
    integrantesMock.mockRejectedValue(new Error('caída'))

    renderizar()

    expect(await screen.findByRole('heading', { name: 'Proyecto Alfa' })).toBeInTheDocument()
    expect(await screen.findByText('No se pudo cargar el equipo. Recargá la página.')).toBeInTheDocument()
  })

  it('muestra un mensaje cuando el proyecto no existe o no hay acceso', async () => {
    obtenerMock.mockRejectedValue(
      new ApiError({ status: 404, code: 'NO_ENCONTRADO', message: 'proyecto no encontrado' }),
    )

    renderizar()

    expect(
      await screen.findByText('No tenés acceso a este proyecto o no existe'),
    ).toBeInTheDocument()
  })

  it('el botón Editar proyecto lleva a la pantalla de edición', async () => {
    obtenerMock.mockResolvedValue(proyecto)
    const usuario = userEvent.setup()

    renderizar()

    await usuario.click(await screen.findByRole('button', { name: 'Editar proyecto' }))

    expect(await screen.findByText('Pantalla de edición')).toBeInTheDocument()
  })
})
