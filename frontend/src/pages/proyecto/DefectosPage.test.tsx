import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { actualizarDefecto, crearDefecto, listarDefectos } from '../../api/defectos'
import { ApiError } from '../../api/errors'
import { listarHistorias } from '../../api/historias'
import { listarSprints } from '../../api/sprints'
import type { Defecto, Historia, Sprint } from '../../api/types'
import DefectosPage from './DefectosPage'

vi.mock('../../api/defectos', () => ({
  listarDefectos: vi.fn(),
  crearDefecto: vi.fn(),
  actualizarDefecto: vi.fn(),
}))
vi.mock('../../api/historias', () => ({ listarHistorias: vi.fn() }))
vi.mock('../../api/sprints', () => ({ listarSprints: vi.fn() }))

const listarDefectosMock = vi.mocked(listarDefectos)
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

const defecto9: Defecto = {
  ...defecto7,
  id: 9,
  descripcion: 'El botón no responde',
  severidad: 'critical',
  estado: 'in_progress',
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

function celdas(fila: HTMLElement): HTMLElement[] {
  return within(fila).getAllByRole('cell')
}

function filaDe(codigo: string): HTMLElement {
  return screen.getByRole('row', { name: new RegExp(codigo) })
}

async function abrirAlta(usuario: Usuario) {
  await usuario.click(await screen.findByRole('button', { name: 'Nuevo defecto' }))
  return screen.findByRole('dialog')
}

async function completarAlta(
  usuario: Usuario,
  dialogo: HTMLElement,
  { descripcion = 'Error al guardar', severidad = 'Mayor', historiaNombre = 'US-01', sprint = 'Sprint 1' } = {},
) {
  if (descripcion !== '') {
    await usuario.type(within(dialogo).getByLabelText('Descripción'), descripcion)
  }
  if (severidad !== '') {
    await usuario.selectOptions(within(dialogo).getByLabelText('Severidad'), severidad)
  }
  if (historiaNombre !== '') {
    await usuario.selectOptions(
      within(dialogo).getByLabelText('Historia'),
      within(dialogo).getByRole('option', { name: new RegExp(historiaNombre) }),
    )
  }
  if (sprint !== '') {
    await usuario.selectOptions(within(dialogo).getByLabelText('Sprint de detección'), sprint)
  }
}

async function abrirCambioDeEstado(usuario: Usuario, codigo: string) {
  await usuario.click(await screen.findByRole('button', { name: `Cambiar estado de ${codigo}` }))
  return screen.findByRole('dialog')
}

describe('DefectosPage - tramo 1', () => {
  beforeEach(() => {
    listarDefectosMock.mockReset()
    crearMock.mockReset()
    actualizarMock.mockReset()
    listarHistoriasMock.mockReset()
    listarSprintsMock.mockReset()
    listarDefectosMock.mockResolvedValue([defecto7, defecto9])
    listarHistoriasMock.mockResolvedValue([historia])
    listarSprintsMock.mockResolvedValue([sprint1, sprint2])
  })

  describe('listado y estados (CA-1, CA-2, CA-3)', () => {
    it('muestra el indicador de carga y no la tabla mientras carga', () => {
      listarDefectosMock.mockReturnValue(new Promise(() => {}))

      renderizar()

      expect(screen.getByText('Cargando defectos…')).toBeInTheDocument()
      expect(screen.queryByRole('table')).not.toBeInTheDocument()
    })

    it('E-01 lista código, descripción, historia, severidad, estado y sprints', async () => {
      renderizar()

      await screen.findByRole('table')
      const encabezados = screen.getAllByRole('columnheader').map((th) => th.textContent)
      expect(encabezados).toEqual([
        'Código',
        'Descripción',
        'Historia',
        'Severidad',
        'Estado',
        'Sprint de detección',
        'Sprint de resolución',
        'Acciones',
      ])

      const datos = celdas(filaDe('DEF-7')).map((celda) => celda.textContent)
      expect(datos.slice(0, 7)).toEqual([
        'DEF-7',
        'Error al guardar',
        'US-01',
        'Mayor',
        'Abierto',
        'Sprint 1',
        '—',
      ])
    })

    it('E-01 ordena del defecto más reciente al más antiguo', async () => {
      renderizar()

      await screen.findByRole('table')
      const codigos = screen
        .getAllByRole('row')
        .slice(1)
        .map((fila) => celdas(fila)[0].textContent)
      expect(codigos).toEqual(['DEF-9', 'DEF-7'])
    })

    it('E-02 trunca la descripción de 81 caracteres a 80 con puntos suspensivos', async () => {
      const larga = 'a'.repeat(81)
      listarDefectosMock.mockResolvedValue([{ ...defecto7, descripcion: larga }])

      renderizar()

      expect(await screen.findByText(`${'a'.repeat(80)}…`)).toBeInTheDocument()
      expect(screen.getByTitle(larga)).toBeInTheDocument()
    })

    it('E-03 no trunca una descripción de exactamente 80 caracteres', async () => {
      const exacta = 'b'.repeat(80)
      listarDefectosMock.mockResolvedValue([{ ...defecto7, descripcion: exacta }])

      renderizar()

      expect(await screen.findByText(exacta)).toBeInTheDocument()
      expect(screen.queryByText(`${exacta}…`)).not.toBeInTheDocument()
    })

    it('E-04 muestra un guion si la historia no está en la lista y no falla', async () => {
      listarDefectosMock.mockResolvedValue([{ ...defecto7, historia_id: 99 }])

      renderizar()

      await screen.findByRole('table')
      expect(celdas(filaDe('DEF-7'))[2]).toHaveTextContent('—')
    })

    it('muestra un guion si el sprint de detección no está en la lista', async () => {
      listarDefectosMock.mockResolvedValue([{ ...defecto7, sprint_deteccion_id: 99 }])

      renderizar()

      await screen.findByRole('table')
      expect(celdas(filaDe('DEF-7'))[5]).toHaveTextContent('—')
    })

    it('E-05 muestra el estado vacío con la acción de crear', async () => {
      listarDefectosMock.mockResolvedValue([])
      const usuario = userEvent.setup()

      renderizar()

      expect(await screen.findByText('Todavía no hay defectos')).toBeInTheDocument()
      await usuario.click(screen.getByRole('button', { name: 'Nuevo defecto' }))
      expect(await screen.findByRole('dialog')).toBeInTheDocument()
    })
  })

  describe('alta de defecto (CA-4, CA-5)', () => {
    it('E-06 un alta válida agrega el defecto a la lista sin recargar', async () => {
      const creado: Defecto = {
        ...defecto7,
        id: 10,
        descripcion: 'Error al guardar el sprint',
        sprint_deteccion_id: 2,
      }
      crearMock.mockResolvedValue(creado)
      const usuario = userEvent.setup()
      renderizar()

      const dialogo = await abrirAlta(usuario)
      await completarAlta(usuario, dialogo, {
        descripcion: '  Error al guardar el sprint  ',
        sprint: 'Sprint 2',
      })
      await usuario.click(within(dialogo).getByRole('button', { name: 'Registrar defecto' }))

      await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
      expect(crearMock).toHaveBeenCalledWith(1, {
        descripcion: 'Error al guardar el sprint',
        severidad: 'major',
        historia_id: 1,
        sprint_deteccion_id: 2,
      })
      expect(listarDefectosMock).toHaveBeenCalledTimes(1)
      const fila = filaDe('DEF-10')
      expect(celdas(fila)[1]).toHaveTextContent('Error al guardar el sprint')
      expect(celdas(fila)[4]).toHaveTextContent('Abierto')
    })

    it('E-07 una descripción vacía no llama a la API', async () => {
      const usuario = userEvent.setup()
      renderizar()

      const dialogo = await abrirAlta(usuario)
      await completarAlta(usuario, dialogo, { descripcion: '' })
      await usuario.type(within(dialogo).getByLabelText('Descripción'), '   ')
      await usuario.click(within(dialogo).getByRole('button', { name: 'Registrar defecto' }))

      expect(await within(dialogo).findByText('La descripción es obligatoria')).toBeInTheDocument()
      expect(crearMock).not.toHaveBeenCalled()
    })

    it.each([
      ['severidad', { severidad: '' }, 'Elegí una severidad'],
      ['historia', { historiaNombre: '' }, 'Elegí una historia'],
      ['sprint de detección', { sprint: '' }, 'Elegí el sprint de detección'],
    ])('E-08 sin elegir %s no llama a la API', async (_campo, omision, mensaje) => {
      const usuario = userEvent.setup()
      renderizar()

      const dialogo = await abrirAlta(usuario)
      await completarAlta(usuario, dialogo, omision)
      await usuario.click(within(dialogo).getByRole('button', { name: 'Registrar defecto' }))

      expect(await within(dialogo).findByText(mensaje)).toBeInTheDocument()
      expect(crearMock).not.toHaveBeenCalled()
    })

    it('E-09 un 422 del backend se muestra junto al campo y el modal sigue abierto', async () => {
      crearMock.mockRejectedValue(
        new ApiError({
          status: 422,
          code: 'VALIDACION',
          message: 'La descripción no es válida',
          campo: 'descripcion',
        }),
      )
      const usuario = userEvent.setup()
      renderizar()

      const dialogo = await abrirAlta(usuario)
      await completarAlta(usuario, dialogo)
      await usuario.click(within(dialogo).getByRole('button', { name: 'Registrar defecto' }))

      expect(await within(dialogo).findByText('La descripción no es válida')).toBeInTheDocument()
      expect(screen.getByRole('dialog')).toBeInTheDocument()
      expect(screen.queryByText('DEF-10')).not.toBeInTheDocument()
    })

    it('E-10 sin historias avisa qué falta y no permite guardar', async () => {
      listarHistoriasMock.mockResolvedValue([])
      const usuario = userEvent.setup()
      renderizar()

      const dialogo = await abrirAlta(usuario)

      expect(
        within(dialogo).getByText(
          'No hay historias en el proyecto. Creá una historia antes de registrar un defecto.',
        ),
      ).toBeInTheDocument()
      expect(within(dialogo).getByRole('button', { name: 'Registrar defecto' })).toBeDisabled()
    })

    it('E-10 sin sprints avisa qué falta y no permite guardar', async () => {
      listarSprintsMock.mockResolvedValue([])
      const usuario = userEvent.setup()
      renderizar()

      const dialogo = await abrirAlta(usuario)

      expect(
        within(dialogo).getByText(
          'No hay sprints en el proyecto. Creá un sprint antes de registrar un defecto.',
        ),
      ).toBeInTheDocument()
      expect(within(dialogo).getByRole('button', { name: 'Registrar defecto' })).toBeDisabled()
    })
  })

  describe('cambio de estado (CA-6, CA-7)', () => {
    it('E-11 resolver sin sprint de resolución muestra el 422 junto al campo', async () => {
      actualizarMock.mockRejectedValue(
        new ApiError({
          status: 422,
          code: 'VALIDACION',
          message: 'El sprint de resolución es obligatorio',
          campo: 'sprint_resolucion_id',
        }),
      )
      const usuario = userEvent.setup()
      renderizar()

      const dialogo = await abrirCambioDeEstado(usuario, 'DEF-7')
      await usuario.selectOptions(within(dialogo).getByLabelText('Estado'), 'Resuelto')
      await usuario.click(within(dialogo).getByRole('button', { name: 'Confirmar' }))

      expect(actualizarMock).toHaveBeenCalledWith(1, 7, { estado: 'resolved' })
      expect(
        await within(dialogo).findByText('El sprint de resolución es obligatorio'),
      ).toBeInTheDocument()
      expect(within(dialogo).getByLabelText('Sprint de resolución')).toHaveAttribute(
        'aria-invalid',
        'true',
      )
      expect(celdas(filaDe('DEF-7'))[4]).toHaveTextContent('Abierto')
    })

    it('informa en el modal que resolver o cerrar exige sprint de resolución', async () => {
      const usuario = userEvent.setup()
      renderizar()

      const dialogo = await abrirCambioDeEstado(usuario, 'DEF-7')
      expect(within(dialogo).queryByLabelText('Sprint de resolución')).not.toBeInTheDocument()
      await usuario.selectOptions(within(dialogo).getByLabelText('Estado'), 'Cerrado')

      expect(within(dialogo).getByLabelText('Sprint de resolución')).toBeInTheDocument()
      expect(
        within(dialogo).getByText('Obligatorio para pasar a Resuelto o Cerrado'),
      ).toBeInTheDocument()
    })

    it('E-12 resolver con sprint de resolución actualiza la lista', async () => {
      actualizarMock.mockResolvedValue({
        ...defecto7,
        estado: 'resolved',
        sprint_resolucion_id: 2,
        resuelto_en: '2026-05-20T10:00:00Z',
      })
      const usuario = userEvent.setup()
      renderizar()

      const dialogo = await abrirCambioDeEstado(usuario, 'DEF-7')
      await usuario.selectOptions(within(dialogo).getByLabelText('Estado'), 'Resuelto')
      await usuario.selectOptions(within(dialogo).getByLabelText('Sprint de resolución'), 'Sprint 2')
      await usuario.click(within(dialogo).getByRole('button', { name: 'Confirmar' }))

      await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
      expect(actualizarMock).toHaveBeenCalledWith(1, 7, {
        estado: 'resolved',
        sprint_resolucion_id: 2,
      })
      const datos = celdas(filaDe('DEF-7'))
      expect(datos[4]).toHaveTextContent('Resuelto')
      expect(datos[6]).toHaveTextContent('Sprint 2')
    })

    it('E-13 cerrar con sprint de resolución actualiza la lista', async () => {
      actualizarMock.mockResolvedValue({
        ...defecto9,
        estado: 'closed',
        sprint_resolucion_id: 2,
      })
      const usuario = userEvent.setup()
      renderizar()

      const dialogo = await abrirCambioDeEstado(usuario, 'DEF-9')
      await usuario.selectOptions(within(dialogo).getByLabelText('Estado'), 'Cerrado')
      await usuario.selectOptions(within(dialogo).getByLabelText('Sprint de resolución'), 'Sprint 2')
      await usuario.click(within(dialogo).getByRole('button', { name: 'Confirmar' }))

      await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
      expect(celdas(filaDe('DEF-9'))[4]).toHaveTextContent('Cerrado')
    })

    it('E-14 reabrir un defecto no envía sprint de resolución', async () => {
      listarDefectosMock.mockResolvedValue([
        { ...defecto7, estado: 'resolved', sprint_resolucion_id: 2 },
      ])
      actualizarMock.mockResolvedValue({
        ...defecto7,
        estado: 'open',
        sprint_resolucion_id: null,
        resuelto_en: null,
      })
      const usuario = userEvent.setup()
      renderizar()

      const dialogo = await abrirCambioDeEstado(usuario, 'DEF-7')
      expect(within(dialogo).getByLabelText('Sprint de resolución')).toHaveValue('2')
      await usuario.selectOptions(within(dialogo).getByLabelText('Estado'), 'Abierto')
      await usuario.click(within(dialogo).getByRole('button', { name: 'Confirmar' }))

      await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
      expect(actualizarMock).toHaveBeenCalledWith(1, 7, { estado: 'open' })
      const datos = celdas(filaDe('DEF-7'))
      expect(datos[4]).toHaveTextContent('Abierto')
      expect(datos[6]).toHaveTextContent('—')
    })

    it('E-15 un estado rechazado por el backend se muestra en el modal', async () => {
      actualizarMock.mockRejectedValue(
        new ApiError({
          status: 400,
          code: 'ESTADO_INVALIDO',
          message: 'El estado no es válido',
          campo: 'estado',
        }),
      )
      const usuario = userEvent.setup()
      renderizar()

      const dialogo = await abrirCambioDeEstado(usuario, 'DEF-7')
      await usuario.selectOptions(within(dialogo).getByLabelText('Estado'), 'Rechazado')
      await usuario.click(within(dialogo).getByRole('button', { name: 'Confirmar' }))

      expect(await within(dialogo).findByText('El estado no es válido')).toBeInTheDocument()
      expect(celdas(filaDe('DEF-7'))[4]).toHaveTextContent('Abierto')
    })
  })

  describe('errores de red y acceso (CA-8, CA-9)', () => {
    it('E-16 un error de red al listar permite reintentar', async () => {
      listarDefectosMock.mockRejectedValueOnce(errorDeRed())
      const usuario = userEvent.setup()
      renderizar()

      expect(
        await screen.findByText('No se pudieron cargar los defectos. Intentá nuevamente.'),
      ).toBeInTheDocument()
      await usuario.click(screen.getByRole('button', { name: 'Reintentar' }))

      expect(await screen.findByRole('table')).toBeInTheDocument()
      expect(listarDefectosMock).toHaveBeenCalledTimes(2)
    })

    it('E-17 un error al cargar historias muestra el error con reintento y sin tabla', async () => {
      listarHistoriasMock.mockRejectedValue(errorDeRed())
      renderizar()

      expect(
        await screen.findByText('No se pudieron cargar los defectos. Intentá nuevamente.'),
      ).toBeInTheDocument()
      expect(screen.getByRole('button', { name: 'Reintentar' })).toBeInTheDocument()
      expect(screen.queryByRole('table')).not.toBeInTheDocument()
    })

    it('E-17 un error al cargar sprints muestra el error con reintento y sin tabla', async () => {
      listarSprintsMock.mockRejectedValue(errorDeRed())
      renderizar()

      expect(
        await screen.findByText('No se pudieron cargar los defectos. Intentá nuevamente.'),
      ).toBeInTheDocument()
      expect(screen.queryByRole('table')).not.toBeInTheDocument()
    })

    it('E-18 un error de red al guardar el alta mantiene el modal y permite reintentar', async () => {
      crearMock.mockRejectedValueOnce(errorDeRed())
      crearMock.mockResolvedValueOnce({ ...defecto7, id: 10 })
      const usuario = userEvent.setup()
      renderizar()

      const dialogo = await abrirAlta(usuario)
      await completarAlta(usuario, dialogo)
      await usuario.click(within(dialogo).getByRole('button', { name: 'Registrar defecto' }))

      expect(
        await within(dialogo).findByText('No se pudo crear el defecto. Intentá nuevamente.'),
      ).toBeInTheDocument()
      expect(within(dialogo).getByLabelText('Descripción')).toHaveValue('Error al guardar')

      await usuario.click(within(dialogo).getByRole('button', { name: 'Registrar defecto' }))
      await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
      expect(crearMock).toHaveBeenCalledTimes(2)
    })

    it('E-19 un error de red al cambiar el estado mantiene el modal y la tabla', async () => {
      actualizarMock.mockRejectedValue(errorDeRed())
      const usuario = userEvent.setup()
      renderizar()

      const dialogo = await abrirCambioDeEstado(usuario, 'DEF-7')
      await usuario.selectOptions(within(dialogo).getByLabelText('Estado'), 'En progreso')
      await usuario.click(within(dialogo).getByRole('button', { name: 'Confirmar' }))

      expect(
        await within(dialogo).findByText(
          'No se pudo cambiar el estado del defecto. Intentá nuevamente.',
        ),
      ).toBeInTheDocument()
      expect(celdas(filaDe('DEF-7'))[4]).toHaveTextContent('Abierto')
    })

    it('E-20 un no miembro ve el mensaje de acceso y no la tabla', async () => {
      listarDefectosMock.mockRejectedValue(
        new ApiError({
          status: 404,
          code: 'PROYECTO_NO_ENCONTRADO',
          message: 'proyecto no encontrado',
        }),
      )
      renderizar()

      expect(
        await screen.findByText('No tenés acceso a este proyecto o no existe'),
      ).toBeInTheDocument()
      expect(screen.queryByRole('table')).not.toBeInTheDocument()
    })
  })

  describe('accesibilidad de los modales (CA-10)', () => {
    it('E-21 el modal de alta tiene foco, etiquetas y cierra con Escape', async () => {
      const usuario = userEvent.setup()
      renderizar()

      const dialogo = await abrirAlta(usuario)

      expect(dialogo).toHaveAttribute('aria-modal', 'true')
      expect(dialogo).toHaveAccessibleName('Nuevo defecto')
      expect(within(dialogo).getByLabelText('Descripción')).toHaveFocus()
      expect(within(dialogo).getByLabelText('Severidad')).toBeInTheDocument()
      expect(within(dialogo).getByLabelText('Historia')).toBeInTheDocument()
      expect(within(dialogo).getByLabelText('Sprint de detección')).toBeInTheDocument()

      await usuario.keyboard('{Escape}')
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    })

    it('E-21 el botón Cancelar cierra el modal sin llamar a la API', async () => {
      const usuario = userEvent.setup()
      renderizar()

      const dialogo = await abrirAlta(usuario)
      await usuario.click(within(dialogo).getByRole('button', { name: 'Cancelar' }))

      expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
      expect(crearMock).not.toHaveBeenCalled()
    })

    it('E-21 el modal de cambio de estado es accesible y cierra con Escape', async () => {
      const usuario = userEvent.setup()
      renderizar()

      const dialogo = await abrirCambioDeEstado(usuario, 'DEF-7')

      expect(dialogo).toHaveAttribute('aria-modal', 'true')
      expect(dialogo).toHaveAccessibleName('Cambiar estado')
      expect(within(dialogo).getByLabelText('Estado')).toHaveFocus()

      await usuario.keyboard('{Escape}')
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    })

    it('E-22 un campo con error queda marcado como inválido y descrito', async () => {
      const usuario = userEvent.setup()
      renderizar()

      const dialogo = await abrirAlta(usuario)
      await usuario.click(within(dialogo).getByRole('button', { name: 'Registrar defecto' }))

      const descripcion = await within(dialogo).findByLabelText('Descripción')
      expect(descripcion).toHaveAttribute('aria-invalid', 'true')
      const idError = descripcion.getAttribute('aria-describedby')
      expect(idError).toBeTruthy()
      expect(document.getElementById(idError ?? '')).toHaveTextContent(
        'La descripción es obligatoria',
      )
    })
  })
})
