import { fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiError } from '../../api/errors'
import { cambiarEstadoSprint, crearSprint, listarSprints } from '../../api/sprints'
import type { Sprint } from '../../api/types'
import SprintsPage from './SprintsPage'

vi.mock('../../api/sprints', () => ({
  listarSprints: vi.fn(),
  crearSprint: vi.fn(),
  cambiarEstadoSprint: vi.fn(),
}))

const listarMock = vi.mocked(listarSprints)
const crearMock = vi.mocked(crearSprint)
const cambiarEstadoMock = vi.mocked(cambiarEstadoSprint)

const planificado: Sprint = {
  id: 1,
  identificador: 'Sprint 1',
  sprint_goal: 'Entregar el MVP',
  estado: 'Planificado',
  fecha_inicio: '2026-05-01',
  fecha_fin: '2026-05-15',
  creado_en: '2026-04-01T00:00:00Z',
  actualizado_en: '2026-04-01T00:00:00Z',
}

const activo: Sprint = {
  ...planificado,
  id: 2,
  identificador: 'Sprint 2',
  sprint_goal: 'Estabilizar la entrega',
  estado: 'Activo',
  fecha_inicio: '2026-05-16',
  fecha_fin: '2026-05-30',
}

const cerrado: Sprint = {
  ...planificado,
  id: 3,
  identificador: 'Sprint 0',
  sprint_goal: 'Preparar el equipo',
  estado: 'Cerrado',
  fecha_inicio: '2026-04-01',
  fecha_fin: '2026-04-15',
}

function renderizar() {
  return render(
    <MemoryRouter initialEntries={['/proyectos/1/sprints']}>
      <Routes>
        <Route path="/proyectos/:id/sprints" element={<SprintsPage />} />
      </Routes>
    </MemoryRouter>,
  )
}

async function abrirAlta(usuario: ReturnType<typeof userEvent.setup>) {
  await usuario.click(await screen.findByRole('button', { name: 'Nuevo sprint' }))
  return screen.findByRole('dialog')
}

async function completar(
  usuario: ReturnType<typeof userEvent.setup>,
  { identificador = 'Sprint 3', goal = 'Entregar valor', inicio = '2026-06-01', fin = '2026-06-15' } = {},
) {
  if (identificador !== '') {
    await usuario.clear(screen.getByLabelText('Identificador'))
    await usuario.type(screen.getByLabelText('Identificador'), identificador)
  }
  await usuario.clear(screen.getByLabelText('Sprint Goal'))
  await usuario.type(screen.getByLabelText('Sprint Goal'), goal)
  fireEvent.change(screen.getByLabelText('Fecha de inicio'), { target: { value: inicio } })
  fireEvent.change(screen.getByLabelText('Fecha de fin'), { target: { value: fin } })
}

describe('SprintsPage', () => {
  beforeEach(() => {
    listarMock.mockReset()
    crearMock.mockReset()
    cambiarEstadoMock.mockReset()
  })

  it('muestra el indicador de carga y no la tabla mientras carga', () => {
    listarMock.mockReturnValue(new Promise(() => {}))

    renderizar()

    expect(screen.getByText('Cargando sprints…')).toBeInTheDocument()
    expect(screen.queryByRole('table')).not.toBeInTheDocument()
  })

  it('muestra el estado vacío con la acción de crear', async () => {
    listarMock.mockResolvedValue([])
    const usuario = userEvent.setup()

    renderizar()

    expect(await screen.findByText('Todavía no hay sprints')).toBeInTheDocument()
    await usuario.click(screen.getByRole('button', { name: 'Nuevo sprint' }))
    expect(await screen.findByRole('dialog')).toBeInTheDocument()
  })

  it('lista identificador, Sprint Goal, fechas y estado', async () => {
    listarMock.mockResolvedValue([planificado, activo])

    renderizar()

    expect(await screen.findByRole('table')).toBeInTheDocument()
    for (const encabezado of [
      'Identificador',
      'Sprint Goal',
      'Fecha de inicio',
      'Fecha de fin',
      'Estado',
      'Acciones',
    ]) {
      expect(screen.getByRole('columnheader', { name: encabezado })).toBeInTheDocument()
    }
    expect(screen.getByText('Sprint 1')).toBeInTheDocument()
    expect(screen.getByText('Entregar el MVP')).toBeInTheDocument()
    expect(screen.getByText('01/05/2026')).toBeInTheDocument()
    expect(screen.getByText('15/05/2026')).toBeInTheDocument()
    const tabla = screen.getByRole('table')
    expect(within(tabla).getByText('Planificado')).toBeInTheDocument()
    expect(within(tabla).getByText('Activo')).toBeInTheDocument()
  })

  it('filtra por estado y muestra solo los sprints cerrados', async () => {
    listarMock.mockResolvedValue([planificado, activo, cerrado])
    const usuario = userEvent.setup()

    renderizar()

    await screen.findByRole('table')
    await usuario.selectOptions(screen.getByLabelText('Filtrar por estado'), 'Cerrado')

    expect(screen.getByText('Sprint 0')).toBeInTheDocument()
    expect(screen.queryByText('Sprint 1')).not.toBeInTheDocument()
    expect(screen.queryByText('Sprint 2')).not.toBeInTheDocument()
  })

  it('muestra un aviso cuando el filtro no tiene coincidencias', async () => {
    listarMock.mockResolvedValue([planificado, cerrado])
    const usuario = userEvent.setup()

    renderizar()

    await screen.findByRole('table')
    await usuario.selectOptions(screen.getByLabelText('Filtrar por estado'), 'Activo')

    expect(screen.getByText('No hay sprints con ese estado')).toBeInTheDocument()
    expect(screen.queryByRole('table')).not.toBeInTheDocument()
  })

  it('crea un sprint y lo muestra como Planificado', async () => {
    listarMock.mockResolvedValue([])
    crearMock.mockResolvedValue({
      ...planificado,
      id: 9,
      identificador: 'Sprint 3',
      sprint_goal: 'Entregar valor',
      fecha_inicio: '2026-06-01',
      fecha_fin: '2026-06-15',
    })
    const usuario = userEvent.setup()

    renderizar()

    const modal = await abrirAlta(usuario)
    await completar(usuario)
    await usuario.click(within(modal).getByRole('button', { name: 'Crear sprint' }))

    await waitFor(() =>
      expect(crearMock).toHaveBeenCalledWith(1, {
        identificador: 'Sprint 3',
        sprint_goal: 'Entregar valor',
        fecha_inicio: '2026-06-01',
        fecha_fin: '2026-06-15',
      }),
    )
    expect(await screen.findByText('Sprint 3')).toBeInTheDocument()
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('no llama a la API con el identificador vacío', async () => {
    listarMock.mockResolvedValue([])
    const usuario = userEvent.setup()

    renderizar()

    const modal = await abrirAlta(usuario)
    await usuario.click(within(modal).getByRole('button', { name: 'Crear sprint' }))

    expect(await within(modal).findByText('El identificador es obligatorio')).toBeInTheDocument()
    expect(crearMock).not.toHaveBeenCalled()
  })

  it('no llama a la API con la fecha de fin anterior al inicio', async () => {
    listarMock.mockResolvedValue([])
    const usuario = userEvent.setup()

    renderizar()

    const modal = await abrirAlta(usuario)
    await completar(usuario, { inicio: '2026-06-15', fin: '2026-06-01' })
    await usuario.click(within(modal).getByRole('button', { name: 'Crear sprint' }))

    expect(
      await within(modal).findByText('La fecha de fin debe ser igual o posterior a la de inicio'),
    ).toBeInTheDocument()
    expect(crearMock).not.toHaveBeenCalled()
  })

  it('activa un sprint Planificado tras confirmar', async () => {
    listarMock.mockResolvedValue([planificado])
    cambiarEstadoMock.mockResolvedValue({ ...planificado, estado: 'Activo' })
    const usuario = userEvent.setup()

    renderizar()

    const fila = (await screen.findByText('Sprint 1')).closest('tr') as HTMLElement
    await usuario.click(within(fila).getByRole('button', { name: 'Activar' }))

    const dialogo = await screen.findByRole('dialog')
    await usuario.click(within(dialogo).getByRole('button', { name: 'Confirmar' }))

    expect(cambiarEstadoMock).toHaveBeenCalledWith(1, 1, 'Activo')
    await waitFor(() => {
      const filaActualizada = screen.getByText('Sprint 1').closest('tr') as HTMLElement
      expect(within(filaActualizada).getByText('Activo')).toBeInTheDocument()
    })
  })

  it('deshabilita Activar cuando ya hay un sprint Activo', async () => {
    listarMock.mockResolvedValue([planificado, activo])

    renderizar()

    const fila = (await screen.findByText('Sprint 1')).closest('tr') as HTMLElement
    expect(within(fila).getByRole('button', { name: 'Activar' })).toBeDisabled()
  })

  it('cierra el sprint Activo tras confirmar', async () => {
    listarMock.mockResolvedValue([activo])
    cambiarEstadoMock.mockResolvedValue({ ...activo, estado: 'Cerrado' })
    const usuario = userEvent.setup()

    renderizar()

    const fila = (await screen.findByText('Sprint 2')).closest('tr') as HTMLElement
    await usuario.click(within(fila).getByRole('button', { name: 'Cerrar' }))
    await usuario.click(
      within(await screen.findByRole('dialog')).getByRole('button', { name: 'Confirmar' }),
    )

    expect(cambiarEstadoMock).toHaveBeenCalledWith(1, 2, 'Cerrado')
    await waitFor(() => {
      const filaActualizada = screen.getByText('Sprint 2').closest('tr') as HTMLElement
      expect(within(filaActualizada).getByText('Cerrado')).toBeInTheDocument()
    })
  })

  it('un sprint Cerrado no muestra acciones', async () => {
    listarMock.mockResolvedValue([cerrado])

    renderizar()

    const fila = (await screen.findByText('Sprint 0')).closest('tr') as HTMLElement
    expect(within(fila).queryByRole('button', { name: 'Activar' })).not.toBeInTheDocument()
    expect(within(fila).queryByRole('button', { name: 'Cerrar' })).not.toBeInTheDocument()
  })

  it('muestra el error de la API al activar ante una carrera', async () => {
    listarMock.mockResolvedValue([planificado])
    cambiarEstadoMock.mockRejectedValue(
      new ApiError({
        status: 409,
        code: 'SPRINT_ACTIVO_EXISTENTE',
        message: 'ya existe un sprint activo en el proyecto',
      }),
    )
    const usuario = userEvent.setup()

    renderizar()

    const fila = (await screen.findByText('Sprint 1')).closest('tr') as HTMLElement
    await usuario.click(within(fila).getByRole('button', { name: 'Activar' }))
    await usuario.click(
      within(await screen.findByRole('dialog')).getByRole('button', { name: 'Confirmar' }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'ya existe un sprint activo en el proyecto',
    )
    const filaOriginal = screen.getByText('Sprint 1').closest('tr') as HTMLElement
    expect(within(filaOriginal).getByText('Planificado')).toBeInTheDocument()
  })

  it('muestra el mensaje de acceso cuando el proyecto no existe o no hay membresía', async () => {
    listarMock.mockRejectedValue(
      new ApiError({ status: 404, code: 'PROYECTO_NO_ENCONTRADO', message: 'el proyecto no existe' }),
    )

    renderizar()

    expect(
      await screen.findByText('No tenés acceso a este proyecto o no existe'),
    ).toBeInTheDocument()
  })

  it('permite reintentar cuando falla la carga', async () => {
    listarMock.mockRejectedValueOnce(new Error('falla'))
    listarMock.mockResolvedValueOnce([planificado])
    const usuario = userEvent.setup()

    renderizar()

    expect(await screen.findByRole('alert')).toHaveTextContent('No se pudieron cargar los sprints')
    await usuario.click(screen.getByRole('button', { name: 'Reintentar' }))

    expect(await screen.findByText('Sprint 1')).toBeInTheDocument()
    expect(listarMock).toHaveBeenCalledTimes(2)
  })

  it('enfoca el identificador y cierra el modal con Escape', async () => {
    listarMock.mockResolvedValue([])
    const usuario = userEvent.setup()

    renderizar()

    await abrirAlta(usuario)
    expect(screen.getByLabelText('Identificador')).toHaveFocus()

    await usuario.keyboard('{Escape}')
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
  })

  it('conserva el formulario y muestra el error ante un fallo de red al crear', async () => {
    listarMock.mockResolvedValue([])
    crearMock.mockRejectedValue(
      new ApiError({ status: 0, code: 'NETWORK_ERROR', message: 'Failed to fetch' }),
    )
    const usuario = userEvent.setup()

    renderizar()

    const modal = await abrirAlta(usuario)
    await completar(usuario)
    await usuario.click(within(modal).getByRole('button', { name: 'Crear sprint' }))

    expect(await within(modal).findByRole('alert')).toHaveTextContent(
      'No se pudo crear el sprint',
    )
    expect(screen.getByLabelText('Identificador')).toHaveValue('Sprint 3')
    expect(screen.getByLabelText('Sprint Goal')).toHaveValue('Entregar valor')
  })
})
