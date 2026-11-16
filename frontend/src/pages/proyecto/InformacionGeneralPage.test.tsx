import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import {
  actualizarProyecto,
  cambiarEstadoProyecto,
  eliminarProyecto,
  obtenerProyecto,
} from '../../api/proyectos'
import { ApiError } from '../../api/errors'
import { listarSprints } from '../../api/sprints'
import type { Proyecto, Sprint } from '../../api/types'
import InformacionGeneralPage from './InformacionGeneralPage'

vi.mock('../../api/proyectos', () => ({
  obtenerProyecto: vi.fn(),
  actualizarProyecto: vi.fn(),
  cambiarEstadoProyecto: vi.fn(),
  eliminarProyecto: vi.fn(),
}))

vi.mock('../../api/sprints', () => ({ listarSprints: vi.fn() }))

const listarSprintsMock = vi.mocked(listarSprints)
const obtenerMock = vi.mocked(obtenerProyecto)
const actualizarMock = vi.mocked(actualizarProyecto)
const estadoMock = vi.mocked(cambiarEstadoProyecto)
const eliminarMock = vi.mocked(eliminarProyecto)

const proyecto: Proyecto = {
  id: 1,
  nombre: 'Proyecto Alfa',
  descripcion: 'Línea uno\nLínea dos',
  fecha_inicio: '2026-01-15',
  fecha_fin: '2026-06-30',
  estado: 'Activo',
  cantidad_usuarios: 3,
  usuarios: [],
  rol_actual: 'LIDER',
  creado_en: '2026-01-15T00:00:00Z',
  actualizado_en: '2026-01-15T00:00:00Z',
}

function renderizar() {
  return render(
    <MemoryRouter initialEntries={['/proyectos/1']}>
      <Routes>
        <Route path="/proyectos" element={<p>Listado de proyectos</p>} />
        <Route path="/proyectos/:id" element={<InformacionGeneralPage />} />
      </Routes>
    </MemoryRouter>,
  )
}

describe('InformacionGeneralPage', () => {
  beforeEach(() => {
    obtenerMock.mockReset()
    actualizarMock.mockReset()
    estadoMock.mockReset()
    eliminarMock.mockReset()
    listarSprintsMock.mockReset()
    listarSprintsMock.mockResolvedValue([])
  })

  it('muestra los datos del proyecto, incluida la descripción multilínea', async () => {
    obtenerMock.mockResolvedValue(proyecto)

    renderizar()

    expect(await screen.findByRole('heading', { name: 'Información general' })).toBeInTheDocument()
    expect(screen.getByText('Proyecto Alfa')).toBeInTheDocument()
    expect(screen.getByText('15/01/2026')).toBeInTheDocument()
    expect(screen.getByText('30/06/2026')).toBeInTheDocument()
    expect(screen.getByText(/Línea uno\s+Línea dos/, { selector: 'p' })).toBeInTheDocument()
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

  it('edita solo los datos y envía un PUT', async () => {
    obtenerMock.mockResolvedValue(proyecto)
    actualizarMock.mockResolvedValue({ ...proyecto, nombre: 'Proyecto Renombrado' })
    const usuario = userEvent.setup()

    renderizar()

    const nombre = await screen.findByLabelText('Nombre')
    await usuario.clear(nombre)
    await usuario.type(nombre, 'Proyecto Renombrado')
    await usuario.click(screen.getByRole('button', { name: 'Guardar cambios' }))

    await waitFor(() =>
      expect(screen.getByText('Los cambios se guardaron correctamente.')).toBeInTheDocument(),
    )
    expect(actualizarMock).toHaveBeenCalledWith(1, {
      nombre: 'Proyecto Renombrado',
      descripcion: 'Línea uno\nLínea dos',
      fecha_inicio: '2026-01-15',
      fecha_fin: '2026-06-30',
    })
    expect(estadoMock).not.toHaveBeenCalled()
  })

  it('cambia solo el estado y envía un PATCH sin PUT', async () => {
    obtenerMock.mockResolvedValue(proyecto)
    estadoMock.mockResolvedValue({ ...proyecto, estado: 'Cerrado' })
    const usuario = userEvent.setup()

    renderizar()

    await usuario.selectOptions(await screen.findByLabelText('Estado'), 'Cerrado')
    await usuario.click(screen.getByRole('button', { name: 'Guardar cambios' }))

    await waitFor(() => expect(estadoMock).toHaveBeenCalledWith(1, 'Cerrado'))
    expect(actualizarMock).not.toHaveBeenCalled()
  })

  it('envía PUT y luego PATCH cuando cambian datos y estado', async () => {
    obtenerMock.mockResolvedValue(proyecto)
    actualizarMock.mockResolvedValue({ ...proyecto, nombre: 'Proyecto Renombrado' })
    estadoMock.mockResolvedValue({
      ...proyecto,
      nombre: 'Proyecto Renombrado',
      estado: 'Cerrado',
    })
    const usuario = userEvent.setup()

    renderizar()

    const nombre = await screen.findByLabelText('Nombre')
    await usuario.clear(nombre)
    await usuario.type(nombre, 'Proyecto Renombrado')
    await usuario.selectOptions(screen.getByLabelText('Estado'), 'Cerrado')
    await usuario.click(screen.getByRole('button', { name: 'Guardar cambios' }))

    await waitFor(() => expect(estadoMock).toHaveBeenCalledWith(1, 'Cerrado'))
    expect(actualizarMock).toHaveBeenCalledWith(1, {
      nombre: 'Proyecto Renombrado',
      descripcion: 'Línea uno\nLínea dos',
      fecha_inicio: '2026-01-15',
      fecha_fin: '2026-06-30',
    })
    expect(actualizarMock.mock.invocationCallOrder[0]).toBeLessThan(
      estadoMock.mock.invocationCallOrder[0],
    )
  })

  it('muestra el error de validación de descripción sin llamar a la API', async () => {
    obtenerMock.mockResolvedValue(proyecto)
    const usuario = userEvent.setup()

    renderizar()

    const descripcion = await screen.findByLabelText('Descripción')
    await usuario.click(descripcion)
    await usuario.paste('a'.repeat(2001))
    await usuario.click(screen.getByRole('button', { name: 'Guardar cambios' }))

    expect(
      screen.getByText('La descripción no puede superar los 2000 caracteres'),
    ).toBeInTheDocument()
    expect(actualizarMock).not.toHaveBeenCalled()
    expect(estadoMock).not.toHaveBeenCalled()
  })

  it('en un proyecto cerrado bloquea los datos y permite reabrirlo', async () => {
    obtenerMock.mockResolvedValue({ ...proyecto, estado: 'Cerrado' })
    estadoMock.mockResolvedValue(proyecto)
    const usuario = userEvent.setup()

    renderizar()

    expect(
      await screen.findByText('Reabrí el proyecto para editar sus datos'),
    ).toBeInTheDocument()
    expect(screen.getByLabelText('Nombre')).toBeDisabled()
    expect(screen.getByRole('button', { name: 'Guardar cambios' })).toBeDisabled()

    await usuario.click(screen.getByRole('button', { name: 'Reabrir proyecto' }))

    await waitFor(() => expect(estadoMock).toHaveBeenCalledWith(1, 'Activo'))
    await waitFor(() => expect(screen.getByLabelText('Nombre')).toBeEnabled())
  })

  it('conserva el PUT y permite reintentar el estado si el PATCH falla', async () => {
    obtenerMock.mockResolvedValue(proyecto)
    actualizarMock.mockResolvedValue({ ...proyecto, nombre: 'Proyecto Renombrado' })
    estadoMock.mockRejectedValueOnce(
      new ApiError({ status: 500, code: 'ERROR_INTERNO', message: 'no se pudo cambiar el estado' }),
    )
    const usuario = userEvent.setup()

    renderizar()

    const nombre = await screen.findByLabelText('Nombre')
    await usuario.clear(nombre)
    await usuario.type(nombre, 'Proyecto Renombrado')
    await usuario.selectOptions(screen.getByLabelText('Estado'), 'Cerrado')
    await usuario.click(screen.getByRole('button', { name: 'Guardar cambios' }))

    expect(await screen.findByText('no se pudo cambiar el estado')).toBeInTheDocument()
    expect(screen.getByLabelText('Nombre')).toHaveValue('Proyecto Renombrado')

    estadoMock.mockResolvedValue({ ...proyecto, nombre: 'Proyecto Renombrado', estado: 'Cerrado' })
    await usuario.click(screen.getByRole('button', { name: 'Guardar cambios' }))

    await waitFor(() => expect(estadoMock).toHaveBeenCalledTimes(2))
    expect(actualizarMock).toHaveBeenCalledTimes(1)
    expect(estadoMock).toHaveBeenLastCalledWith(1, 'Cerrado')
  })

  it('elimina el proyecto y navega al listado', async () => {
    obtenerMock.mockResolvedValue(proyecto)
    eliminarMock.mockResolvedValue(undefined)
    const usuario = userEvent.setup()

    renderizar()

    await usuario.click(await screen.findByRole('button', { name: 'Eliminar proyecto' }))
    await usuario.click(screen.getByRole('button', { name: 'Eliminar' }))

    await waitFor(() => expect(screen.getByText('Listado de proyectos')).toBeInTheDocument())
    expect(eliminarMock).toHaveBeenCalledWith(1)
  })

  it('ante 409 conserva el proyecto y muestra el mensaje', async () => {
    obtenerMock.mockResolvedValue(proyecto)
    eliminarMock.mockRejectedValue(
      new ApiError({
        status: 409,
        code: 'PROYECTO_CON_HISTORIAL',
        message: 'el proyecto tiene historial asociado y no puede eliminarse',
      }),
    )
    const usuario = userEvent.setup()

    renderizar()

    await usuario.click(await screen.findByRole('button', { name: 'Eliminar proyecto' }))
    await usuario.click(screen.getByRole('button', { name: 'Eliminar' }))

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'el proyecto tiene historial asociado y no puede eliminarse',
    )
    expect(screen.getByRole('dialog', { name: 'Eliminar proyecto' })).toBeInTheDocument()
  })

  describe('encabezado y tarjetas de datos', () => {
    const sprintActivo: Sprint = {
      id: 7,
      identificador: 'Sprint 3',
      sprint_goal: 'Cerrar el módulo',
      estado: 'Activo',
      fecha_inicio: '2026-06-01',
      fecha_fin: '2026-06-14',
      creado_en: '2026-05-20T00:00:00Z',
      actualizado_en: '2026-05-20T00:00:00Z',
    }

    it('muestra el nombre como título, el chip de estado a su derecha y la descripción debajo', async () => {
      obtenerMock.mockResolvedValue(proyecto)

      renderizar()

      const titulo = await screen.findByRole('heading', { name: 'Proyecto Alfa' })
      const encabezado = titulo.closest('header') as HTMLElement
      expect(within(encabezado).getByText('Activo')).toBeInTheDocument()
      expect(within(encabezado).getByText(/Línea uno\s+Línea dos/)).toBeInTheDocument()
    })

    it('el chip de Activo es verde y el de Cerrado es gris', async () => {
      obtenerMock.mockResolvedValue(proyecto)
      const { unmount } = renderizar()
      expect((await screen.findByTestId('chip-estado-proyecto')).className).toContain('--estado-hecho')
      unmount()

      obtenerMock.mockResolvedValue({ ...proyecto, estado: 'Cerrado' })
      renderizar()
      const cerrado = (await screen.findByTestId('chip-estado-proyecto')).className
      expect(cerrado).toContain('text-ink-soft')
      expect(cerrado).not.toContain('--estado-hecho')
    })

    it('ofrece un botón secundario Editar junto al título', async () => {
      obtenerMock.mockResolvedValue(proyecto)

      renderizar()

      const titulo = await screen.findByRole('heading', { name: 'Proyecto Alfa' })
      const encabezado = titulo.closest('header') as HTMLElement
      expect(within(encabezado).getByRole('button', { name: 'Editar' })).toBeInTheDocument()
    })

    it('muestra las fechas, "Sin fecha" atenuado y los integrantes con enlace a Equipo', async () => {
      obtenerMock.mockResolvedValue({ ...proyecto, fecha_fin: null })

      renderizar()

      expect(await screen.findByText('15/01/2026')).toBeInTheDocument()
      expect(screen.getByText('Sin fecha')).toHaveClass('text-ink-soft')
      const integrantes = screen.getByRole('link', { name: /Integrantes/ })
      expect(integrantes).toHaveAttribute('href', '/proyectos/1/equipo')
      expect(integrantes).toHaveTextContent('3')
    })

    it('todos los valores de las tarjetas usan la misma tipografía', async () => {
      obtenerMock.mockResolvedValue(proyecto)

      renderizar()

      const valores = await screen.findAllByTestId('valor-tarjeta')
      expect(valores.length).toBeGreaterThanOrEqual(3)
      const clases = new Set(valores.map((valor) => valor.className))
      expect(clases.size).toBe(1)
      for (const valor of valores) {
        expect(valor.className).not.toMatch(/font-mono/)
      }
    })

    it('muestra el sprint activo con enlace a su detalle', async () => {
      obtenerMock.mockResolvedValue(proyecto)
      listarSprintsMock.mockResolvedValue([sprintActivo])

      renderizar()

      const enlace = await screen.findByRole('link', { name: /Sprint activo/ })
      expect(enlace).toHaveAttribute('href', '/proyectos/1/sprints/7')
      expect(enlace).toHaveTextContent('Sprint 3')
    })

    it('no muestra la tarjeta de sprint activo si no hay ninguno', async () => {
      obtenerMock.mockResolvedValue(proyecto)
      listarSprintsMock.mockResolvedValue([{ ...sprintActivo, estado: 'Cerrado' }])

      renderizar()

      await screen.findByRole('heading', { name: 'Proyecto Alfa' })
      await waitFor(() => expect(listarSprintsMock).toHaveBeenCalled())
      expect(screen.queryByText('Sprint activo')).not.toBeInTheDocument()
    })

    it('si no se pueden cargar los sprints, la pantalla sigue funcionando sin esa tarjeta', async () => {
      obtenerMock.mockResolvedValue(proyecto)
      listarSprintsMock.mockRejectedValue(new Error('sin conexión'))

      renderizar()

      expect(await screen.findByRole('heading', { name: 'Proyecto Alfa' })).toBeInTheDocument()
      expect(screen.queryByText('Sprint activo')).not.toBeInTheDocument()
    })
  })

  describe('zona de peligro', () => {
    it('la muestra al final, con la advertencia y el botón de eliminar, si el rol es LIDER', async () => {
      obtenerMock.mockResolvedValue(proyecto)

      renderizar()

      const zona = await screen.findByRole('region', { name: 'Zona de peligro' })
      expect(within(zona).getByText(/no se puede deshacer/i)).toBeInTheDocument()
      expect(within(zona).getByRole('button', { name: 'Eliminar proyecto' })).toBeInTheDocument()
      const formulario = screen.getByRole('button', { name: 'Guardar cambios' })
      expect(
        formulario.compareDocumentPosition(zona) & Node.DOCUMENT_POSITION_FOLLOWING,
      ).toBeTruthy()
    })

    it.each([
      ['MIEMBRO', 'MIEMBRO' as const],
      ['sin rol', undefined],
    ])('no la muestra si el rol es %s', async (_nombre, rol) => {
      obtenerMock.mockResolvedValue({ ...proyecto, rol_actual: rol })

      renderizar()

      await screen.findByRole('heading', { name: 'Proyecto Alfa' })
      expect(screen.queryByText('Zona de peligro')).not.toBeInTheDocument()
      expect(screen.queryByRole('button', { name: 'Eliminar proyecto' })).not.toBeInTheDocument()
    })
  })
})
