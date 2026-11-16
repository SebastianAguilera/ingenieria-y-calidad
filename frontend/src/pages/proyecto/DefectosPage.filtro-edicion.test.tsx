import { fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import {
  actualizarDefecto,
  crearDefecto,
  listarDefectos,
  listarDefectosPorSprint,
} from '../../api/defectos'
import { ApiError } from '../../api/errors'
import { listarHistorias } from '../../api/historias'
import { listarSprints } from '../../api/sprints'
import type { Defecto, Historia, Sprint } from '../../api/types'
import DefectosPage from './DefectosPage'

vi.mock('../../api/defectos', () => ({
  listarDefectos: vi.fn(),
  listarDefectosPorSprint: vi.fn(),
  crearDefecto: vi.fn(),
  actualizarDefecto: vi.fn(),
}))
vi.mock('../../api/historias', () => ({ listarHistorias: vi.fn() }))
vi.mock('../../api/sprints', () => ({ listarSprints: vi.fn() }))

const listarDefectosMock = vi.mocked(listarDefectos)
const listarPorSprintMock = vi.mocked(listarDefectosPorSprint)
const crearMock = vi.mocked(crearDefecto)
const actualizarMock = vi.mocked(actualizarDefecto)
const listarHistoriasMock = vi.mocked(listarHistorias)
const listarSprintsMock = vi.mocked(listarSprints)

const historia: Historia = {
  id: 1,
  codigo: 'US-01',
  titulo: 'Iniciar sesión',
  descripcion: '',
  prioridad: 'Alta',
  estado: 'Backlog',
  story_points: null,
  esfuerzo_real: null,
  valor_negocio: null,
  criticidad_tiempo: null,
  reduccion_riesgo: null,
  horas_estimadas: null,
  creado_en: '2026-04-01T00:00:00Z',
  actualizado_en: '2026-04-01T00:00:00Z',
}

const sprint1: Sprint = {
  id: 1,
  identificador: 'Sprint 1',
  sprint_goal: 'Entregar el MVP',
  estado: 'Activo',
  fecha_inicio: '2026-05-01',
  fecha_fin: '2026-05-15',
  creado_en: '2026-04-01T00:00:00Z',
  actualizado_en: '2026-04-01T00:00:00Z',
}

const sprint2: Sprint = {
  ...sprint1,
  id: 2,
  identificador: 'Sprint 2',
  estado: 'Planificado',
  fecha_inicio: '2026-05-16',
  fecha_fin: '2026-05-30',
}

// Detectado en el sprint 1 y todavía sin resolver.
const defecto7: Defecto = {
  id: 7,
  historia_id: 1,
  sprint_deteccion_id: 1,
  sprint_resolucion_id: null,
  reportado_por: null,
  descripcion: 'Error al guardar',
  severidad: 'major',
  estado: 'open',
  creado_en: '2026-05-02T10:00:00Z',
  resuelto_en: null,
}

// Detectado en el sprint 1 y resuelto en el sprint 2.
const defecto9: Defecto = {
  ...defecto7,
  id: 9,
  descripcion: 'El botón no responde',
  severidad: 'critical',
  estado: 'resolved',
  sprint_resolucion_id: 2,
  resuelto_en: '2026-05-20T10:00:00Z',
}

// Detectado en el sprint 2.
const defecto11: Defecto = {
  ...defecto7,
  id: 11,
  descripcion: 'Falla al exportar',
  sprint_deteccion_id: 2,
}

type Usuario = ReturnType<typeof userEvent.setup>

function renderizar() {
  return render(
    <MemoryRouter initialEntries={['/proyectos/1/defectos']}>
      <Routes>
        <Route path="/proyectos/:id/defectos" element={<DefectosPage />} />
      </Routes>
    </MemoryRouter>,
  )
}

function errorDeRed() {
  return new ApiError({ status: 0, code: 'NETWORK_ERROR', message: 'Failed to fetch' })
}

function codigosEnTabla(): string[] {
  return screen
    .getAllByRole('row')
    .slice(1)
    .map((fila) => within(fila).getAllByRole('cell')[0].textContent ?? '')
}

async function abrirEdicion(usuario: Usuario, codigo: string) {
  await usuario.click(await screen.findByRole('button', { name: `Editar ${codigo}` }))
  return screen.findByRole('dialog')
}

describe('DefectosPage - tramo 2: filtro por sprint', () => {
  beforeEach(() => {
    listarDefectosMock.mockReset()
    listarPorSprintMock.mockReset()
    crearMock.mockReset()
    actualizarMock.mockReset()
    listarHistoriasMock.mockReset()
    listarSprintsMock.mockReset()
    listarDefectosMock.mockResolvedValue([defecto7, defecto9, defecto11])
    listarHistoriasMock.mockResolvedValue([historia])
    listarSprintsMock.mockResolvedValue([sprint1, sprint2])
  })

  it('no muestra el filtro mientras el proyecto no tiene defectos', async () => {
    listarDefectosMock.mockResolvedValue([])

    renderizar()

    expect(await screen.findByText('Todavía no hay defectos')).toBeInTheDocument()
    expect(screen.queryByLabelText('Filtrar por sprint')).not.toBeInTheDocument()
  })

  it('el filtro arranca en "Todos los sprints" con el tipo deshabilitado', async () => {
    renderizar()

    expect(await screen.findByLabelText('Filtrar por sprint')).toHaveValue('')
    expect(screen.getByLabelText('Tipo de filtro')).toBeDisabled()
  })

  it('E-23 filtrar por sprint de detección consulta el listado del sprint', async () => {
    listarPorSprintMock.mockResolvedValue([defecto7, defecto9])
    const usuario = userEvent.setup()
    renderizar()

    await usuario.selectOptions(await screen.findByLabelText('Filtrar por sprint'), 'Sprint 1')

    await waitFor(() => expect(codigosEnTabla()).toEqual(['DEF-9', 'DEF-7']))
    expect(listarPorSprintMock).toHaveBeenCalledWith(1, 1, 'deteccion', expect.any(AbortSignal))
    expect(screen.queryByText('DEF-11')).not.toBeInTheDocument()
  })

  it('E-24 filtrar por sprint de resolución consulta con tipo resolucion', async () => {
    listarPorSprintMock.mockResolvedValue([defecto9])
    const usuario = userEvent.setup()
    renderizar()

    await usuario.selectOptions(await screen.findByLabelText('Filtrar por sprint'), 'Sprint 2')
    await usuario.selectOptions(screen.getByLabelText('Tipo de filtro'), 'Resolución')

    await waitFor(() => expect(codigosEnTabla()).toEqual(['DEF-9']))
    expect(listarPorSprintMock).toHaveBeenLastCalledWith(
      1,
      2,
      'resolucion',
      expect.any(AbortSignal),
    )
  })

  it('el filtro y sus controles siguen visibles mientras se carga el nuevo listado', async () => {
    listarPorSprintMock.mockReturnValue(new Promise(() => {}))
    const usuario = userEvent.setup()
    renderizar()

    await usuario.selectOptions(await screen.findByLabelText('Filtrar por sprint'), 'Sprint 1')

    expect(await screen.findByText('Cargando defectos…')).toBeInTheDocument()
    expect(screen.getByLabelText('Filtrar por sprint')).toHaveValue('1')
    expect(screen.queryByRole('table')).not.toBeInTheDocument()
  })

  it('E-25 un filtro sin resultados muestra un aviso y no el estado vacío', async () => {
    listarPorSprintMock.mockResolvedValue([])
    const usuario = userEvent.setup()
    renderizar()

    await usuario.selectOptions(await screen.findByLabelText('Filtrar por sprint'), 'Sprint 2')

    expect(await screen.findByText('No hay defectos con ese filtro')).toBeInTheDocument()
    expect(screen.queryByText('Todavía no hay defectos')).not.toBeInTheDocument()
    expect(screen.getByLabelText('Filtrar por sprint')).toBeInTheDocument()
  })

  it('E-26 quitar el filtro vuelve al listado completo', async () => {
    listarPorSprintMock.mockResolvedValue([defecto9])
    const usuario = userEvent.setup()
    renderizar()

    const filtro = await screen.findByLabelText('Filtrar por sprint')
    await usuario.selectOptions(filtro, 'Sprint 2')
    await waitFor(() => expect(codigosEnTabla()).toEqual(['DEF-9']))
    await usuario.selectOptions(screen.getByLabelText('Filtrar por sprint'), 'Todos los sprints')

    await waitFor(() => expect(codigosEnTabla()).toEqual(['DEF-11', 'DEF-9', 'DEF-7']))
    expect(listarDefectosMock).toHaveBeenCalledTimes(2)
    expect(screen.getByLabelText('Tipo de filtro')).toBeDisabled()
  })

  it('un error al aplicar el filtro muestra el error con reintento', async () => {
    listarPorSprintMock.mockRejectedValueOnce(errorDeRed())
    const usuario = userEvent.setup()
    renderizar()

    await usuario.selectOptions(await screen.findByLabelText('Filtrar por sprint'), 'Sprint 1')

    expect(
      await screen.findByText('No se pudieron cargar los defectos. Intentá nuevamente.'),
    ).toBeInTheDocument()
    listarPorSprintMock.mockResolvedValue([defecto7])
    await usuario.click(screen.getByRole('button', { name: 'Reintentar' }))

    await waitFor(() => expect(codigosEnTabla()).toEqual(['DEF-7']))
  })

  it('E-36 con el filtro por detección, un defecto nuevo de otro sprint no aparece', async () => {
    listarPorSprintMock.mockResolvedValue([defecto7, defecto9])
    crearMock.mockResolvedValue({ ...defecto11, id: 12 })
    const usuario = userEvent.setup()
    renderizar()

    await usuario.selectOptions(await screen.findByLabelText('Filtrar por sprint'), 'Sprint 1')
    await waitFor(() => expect(codigosEnTabla()).toEqual(['DEF-9', 'DEF-7']))
    await usuario.click(screen.getByRole('button', { name: 'Nuevo defecto' }))
    const dialogo = await screen.findByRole('dialog')
    await usuario.type(within(dialogo).getByLabelText('Descripción'), 'Falla al exportar')
    await usuario.selectOptions(within(dialogo).getByLabelText('Severidad'), 'Mayor')
    await usuario.selectOptions(within(dialogo).getByLabelText('Historia'), '1')
    await usuario.selectOptions(within(dialogo).getByLabelText('Sprint de detección'), 'Sprint 2')
    await usuario.click(within(dialogo).getByRole('button', { name: 'Registrar defecto' }))

    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
    expect(crearMock).toHaveBeenCalledTimes(1)
    expect(codigosEnTabla()).toEqual(['DEF-9', 'DEF-7'])
  })

  it('E-37 con el filtro por resolución, reabrir un defecto lo quita de la lista', async () => {
    listarPorSprintMock.mockResolvedValue([defecto9])
    actualizarMock.mockResolvedValue({
      ...defecto9,
      estado: 'open',
      sprint_resolucion_id: null,
      resuelto_en: null,
    })
    const usuario = userEvent.setup()
    renderizar()

    await usuario.selectOptions(await screen.findByLabelText('Filtrar por sprint'), 'Sprint 2')
    await usuario.selectOptions(screen.getByLabelText('Tipo de filtro'), 'Resolución')
    await waitFor(() => expect(codigosEnTabla()).toEqual(['DEF-9']))
    await usuario.click(screen.getByRole('button', { name: 'Cambiar estado de DEF-9' }))
    const dialogo = await screen.findByRole('dialog')
    await usuario.selectOptions(within(dialogo).getByLabelText('Estado'), 'Abierto')
    await usuario.click(within(dialogo).getByRole('button', { name: 'Confirmar' }))

    expect(await screen.findByText('No hay defectos con ese filtro')).toBeInTheDocument()
  })
})

describe('DefectosPage - tramo 2: edición de descripción y severidad', () => {
  beforeEach(() => {
    listarDefectosMock.mockReset()
    listarPorSprintMock.mockReset()
    crearMock.mockReset()
    actualizarMock.mockReset()
    listarHistoriasMock.mockReset()
    listarSprintsMock.mockReset()
    listarDefectosMock.mockResolvedValue([defecto7, defecto9])
    listarHistoriasMock.mockResolvedValue([historia])
    listarSprintsMock.mockResolvedValue([sprint1, sprint2])
  })

  it('el modal de edición parte de los valores actuales y es accesible', async () => {
    const usuario = userEvent.setup()
    renderizar()

    const dialogo = await abrirEdicion(usuario, 'DEF-7')

    expect(dialogo).toHaveAttribute('aria-modal', 'true')
    expect(dialogo).toHaveAccessibleName('Editar defecto')
    expect(within(dialogo).getByLabelText('Descripción')).toHaveValue('Error al guardar')
    expect(within(dialogo).getByLabelText('Descripción')).toHaveFocus()
    expect(within(dialogo).getByLabelText('Severidad')).toHaveValue('major')

    await usuario.keyboard('{Escape}')
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('E-27 editar la descripción actualiza la lista y envía solo ese campo', async () => {
    actualizarMock.mockResolvedValue({ ...defecto7, descripcion: 'Error al guardar el sprint' })
    const usuario = userEvent.setup()
    renderizar()

    const dialogo = await abrirEdicion(usuario, 'DEF-7')
    const descripcion = within(dialogo).getByLabelText('Descripción')
    await usuario.clear(descripcion)
    await usuario.type(descripcion, '  Error al guardar el sprint  ')
    await usuario.click(within(dialogo).getByRole('button', { name: 'Guardar cambios' }))

    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
    expect(actualizarMock).toHaveBeenCalledWith(1, 7, { descripcion: 'Error al guardar el sprint' })
    expect(screen.getByText('Error al guardar el sprint')).toBeInTheDocument()
  })

  it('E-28 editar la severidad actualiza la lista y envía solo ese campo', async () => {
    actualizarMock.mockResolvedValue({ ...defecto7, severidad: 'critical' })
    const usuario = userEvent.setup()
    renderizar()

    const dialogo = await abrirEdicion(usuario, 'DEF-7')
    await usuario.selectOptions(within(dialogo).getByLabelText('Severidad'), 'Crítica')
    await usuario.click(within(dialogo).getByRole('button', { name: 'Guardar cambios' }))

    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
    expect(actualizarMock).toHaveBeenCalledWith(1, 7, { severidad: 'critical' })
    const fila = screen.getByRole('row', { name: /DEF-7/ })
    expect(within(fila).getAllByRole('cell')[3]).toHaveTextContent('Crítica')
  })

  it('E-29 la edición nunca envía sprint de detección, estado ni sprint de resolución', async () => {
    actualizarMock.mockResolvedValue({
      ...defecto7,
      descripcion: 'Otra descripción',
      severidad: 'blocker',
    })
    const usuario = userEvent.setup()
    renderizar()

    const dialogo = await abrirEdicion(usuario, 'DEF-7')
    const descripcion = within(dialogo).getByLabelText('Descripción')
    await usuario.clear(descripcion)
    await usuario.type(descripcion, 'Otra descripción')
    await usuario.selectOptions(within(dialogo).getByLabelText('Severidad'), 'Bloqueante')
    await usuario.click(within(dialogo).getByRole('button', { name: 'Guardar cambios' }))

    await waitFor(() => expect(actualizarMock).toHaveBeenCalledTimes(1))
    const cuerpo = actualizarMock.mock.calls[0][2]
    expect(Object.keys(cuerpo).sort()).toEqual(['descripcion', 'severidad'])
  })

  it('E-35 guardar sin cambios cierra el modal sin llamar a la API', async () => {
    const usuario = userEvent.setup()
    renderizar()

    const dialogo = await abrirEdicion(usuario, 'DEF-7')
    await usuario.click(within(dialogo).getByRole('button', { name: 'Guardar cambios' }))

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    expect(actualizarMock).not.toHaveBeenCalled()
  })

  it('E-30 más de 1000 caracteres muestra el error y no llama a la API', async () => {
    const usuario = userEvent.setup()
    renderizar()

    const dialogo = await abrirEdicion(usuario, 'DEF-7')
    fireEvent.change(within(dialogo).getByLabelText('Descripción'), {
      target: { value: 'a'.repeat(1001) },
    })
    await usuario.click(within(dialogo).getByRole('button', { name: 'Guardar cambios' }))

    expect(
      await within(dialogo).findByText('La descripción no puede superar los 1000 caracteres'),
    ).toBeInTheDocument()
    expect(within(dialogo).getByLabelText('Descripción')).toHaveAttribute('aria-invalid', 'true')
    expect(actualizarMock).not.toHaveBeenCalled()
  })

  it('E-31 exactamente 1000 caracteres es válido', async () => {
    const texto = 'a'.repeat(1000)
    actualizarMock.mockResolvedValue({ ...defecto7, descripcion: texto })
    const usuario = userEvent.setup()
    renderizar()

    const dialogo = await abrirEdicion(usuario, 'DEF-7')
    fireEvent.change(within(dialogo).getByLabelText('Descripción'), { target: { value: texto } })
    await usuario.click(within(dialogo).getByRole('button', { name: 'Guardar cambios' }))

    await waitFor(() => expect(actualizarMock).toHaveBeenCalledWith(1, 7, { descripcion: texto }))
  })

  it('E-32 el límite cuenta caracteres y no bytes: 1000 "ñ" es válido', async () => {
    const texto = 'ñ'.repeat(1000)
    actualizarMock.mockResolvedValue({ ...defecto7, descripcion: texto })
    const usuario = userEvent.setup()
    renderizar()

    const dialogo = await abrirEdicion(usuario, 'DEF-7')
    fireEvent.change(within(dialogo).getByLabelText('Descripción'), { target: { value: texto } })
    await usuario.click(within(dialogo).getByRole('button', { name: 'Guardar cambios' }))

    await waitFor(() => expect(actualizarMock).toHaveBeenCalledWith(1, 7, { descripcion: texto }))
  })

  it('E-33 una descripción vacía o con solo espacios no llama a la API', async () => {
    const usuario = userEvent.setup()
    renderizar()

    const dialogo = await abrirEdicion(usuario, 'DEF-7')
    fireEvent.change(within(dialogo).getByLabelText('Descripción'), { target: { value: '   ' } })
    await usuario.click(within(dialogo).getByRole('button', { name: 'Guardar cambios' }))

    expect(await within(dialogo).findByText('La descripción es obligatoria')).toBeInTheDocument()
    expect(actualizarMock).not.toHaveBeenCalled()
  })

  it('E-34 un 422 del backend se muestra junto al campo y la tabla no cambia', async () => {
    actualizarMock.mockRejectedValue(
      new ApiError({
        status: 422,
        code: 'VALIDACION',
        message: 'La descripción es demasiado larga',
        campo: 'descripcion',
      }),
    )
    const usuario = userEvent.setup()
    renderizar()

    const dialogo = await abrirEdicion(usuario, 'DEF-7')
    const descripcion = within(dialogo).getByLabelText('Descripción')
    await usuario.clear(descripcion)
    await usuario.type(descripcion, 'Texto nuevo')
    await usuario.click(within(dialogo).getByRole('button', { name: 'Guardar cambios' }))

    expect(await within(dialogo).findByText('La descripción es demasiado larga')).toBeInTheDocument()
    expect(screen.getByRole('dialog')).toBeInTheDocument()
    const tabla = screen.getByRole('table')
    expect(within(tabla).queryByText('Texto nuevo')).not.toBeInTheDocument()
    expect(within(tabla).getByText('Error al guardar')).toBeInTheDocument()
  })

  it('un error de red al editar mantiene el modal y permite reintentar', async () => {
    actualizarMock.mockRejectedValueOnce(errorDeRed())
    actualizarMock.mockResolvedValueOnce({ ...defecto7, severidad: 'minor' })
    const usuario = userEvent.setup()
    renderizar()

    const dialogo = await abrirEdicion(usuario, 'DEF-7')
    await usuario.selectOptions(within(dialogo).getByLabelText('Severidad'), 'Menor')
    await usuario.click(within(dialogo).getByRole('button', { name: 'Guardar cambios' }))

    expect(
      await within(dialogo).findByText('No se pudo editar el defecto. Intentá nuevamente.'),
    ).toBeInTheDocument()
    await usuario.click(within(dialogo).getByRole('button', { name: 'Guardar cambios' }))
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
    expect(actualizarMock).toHaveBeenCalledTimes(2)
  })

  it('editar no modifica el estado ni el sprint de resolución que muestra la tabla', async () => {
    actualizarMock.mockResolvedValue({ ...defecto9, severidad: 'minor' })
    const usuario = userEvent.setup()
    renderizar()

    const dialogo = await abrirEdicion(usuario, 'DEF-9')
    await usuario.selectOptions(within(dialogo).getByLabelText('Severidad'), 'Menor')
    await usuario.click(within(dialogo).getByRole('button', { name: 'Guardar cambios' }))

    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
    const celdas = within(screen.getByRole('row', { name: /DEF-9/ })).getAllByRole('cell')
    expect(celdas[3]).toHaveTextContent('Menor')
    expect(celdas[4]).toHaveTextContent('Resuelto')
    expect(celdas[6]).toHaveTextContent('Sprint 2')
  })
})
