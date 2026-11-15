import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiError } from '../../api/errors'
import { listarHistorias } from '../../api/historias'
import { asignarHistoria, obtenerSprintBacklog, quitarHistoria } from '../../api/sprints'
import type { EstadoSprint, Historia, Sprint, SprintBacklog } from '../../api/types'
import SprintDetallePage from './SprintDetallePage'

vi.mock('../../api/sprints', () => ({
  obtenerSprintBacklog: vi.fn(),
  asignarHistoria: vi.fn(),
  quitarHistoria: vi.fn(),
}))

vi.mock('../../api/historias', () => ({
  listarHistorias: vi.fn(),
}))

const obtenerMock = vi.mocked(obtenerSprintBacklog)
const asignarMock = vi.mocked(asignarHistoria)
const quitarMock = vi.mocked(quitarHistoria)
const listarHistoriasMock = vi.mocked(listarHistorias)

function crearSprint(estado: EstadoSprint): Sprint {
  return {
    id: 1,
    identificador: 'Sprint 1',
    sprint_goal: 'Entregar el MVP',
    estado,
    fecha_inicio: '2026-10-06',
    fecha_fin: '2026-10-20',
    creado_en: '2026-10-01T09:15:00Z',
    actualizado_en: '2026-10-06T09:00:00Z',
  }
}

function crearHistoria(id: number, codigo: string, titulo: string, extra: Partial<Historia> = {}): Historia {
  return {
    id,
    codigo,
    titulo,
    descripcion: '',
    prioridad: 'Media',
    estado: 'Backlog',
    story_points: null,
    esfuerzo_real: null,
    creado_en: '2026-10-01T09:00:00Z',
    actualizado_en: '2026-10-01T09:00:00Z',
    ...extra,
  }
}

const us01 = crearHistoria(1, 'US-01', 'Crear proyecto', { estado: 'En Progreso', story_points: 3 })
const us02 = crearHistoria(2, 'US-02', 'Ver proyectos', { estado: 'En Progreso' })
const us05 = crearHistoria(5, 'US-05', 'Exportar informe')

function crearBacklog(estado: EstadoSprint, historias: Historia[] = []): SprintBacklog {
  return {
    sprint: crearSprint(estado),
    total: historias.length,
    items: historias.map((historia, indice) => ({
      id: 10 + indice,
      orden: indice + 1,
      fecha_asignacion: '2026-10-06T10:00:00Z',
      historia,
    })),
  }
}

function renderizar(sprintId = '1') {
  return render(
    <MemoryRouter initialEntries={[`/proyectos/1/sprints/${sprintId}`]}>
      <Routes>
        <Route path="/proyectos/:id/sprints/:sprintId" element={<SprintDetallePage />} />
        <Route path="/proyectos/:id/sprints" element={<p>Lista de sprints</p>} />
      </Routes>
    </MemoryRouter>,
  )
}

const errorDeRed = () =>
  new ApiError({ status: 0, code: 'ERROR_DE_RED', message: 'Failed to fetch' })

describe('SprintDetallePage', () => {
  beforeEach(() => {
    obtenerMock.mockReset()
    asignarMock.mockReset()
    quitarMock.mockReset()
    listarHistoriasMock.mockReset()
    listarHistoriasMock.mockResolvedValue([])
  })

  it('muestra el indicador de carga mientras consulta el sprint', () => {
    obtenerMock.mockReturnValue(new Promise(() => {}))

    renderizar()

    expect(screen.getByText('Cargando sprint…')).toBeInTheDocument()
    expect(screen.queryByRole('table')).not.toBeInTheDocument()
  })

  it('E-01 consulta el sprint del proyecto y muestra la cabecera', async () => {
    obtenerMock.mockResolvedValue(crearBacklog('Activo', [us01]))

    renderizar()

    expect(await screen.findByRole('heading', { name: /Sprint 1/ })).toBeInTheDocument()
    expect(obtenerMock).toHaveBeenCalledWith(1, 1, expect.anything())
    expect(screen.getByText('Entregar el MVP')).toBeInTheDocument()
    expect(screen.getByText('06/10/2026')).toBeInTheDocument()
    expect(screen.getByText('20/10/2026')).toBeInTheDocument()
    expect(screen.getByText('Activo')).toBeInTheDocument()
  })

  it('E-02 muestra las historias asignadas en el orden de la API', async () => {
    obtenerMock.mockResolvedValue(crearBacklog('Activo', [us02, us01]))

    renderizar()

    await screen.findByRole('table')
    for (const encabezado of ['Código', 'Título', 'Estado', 'Story Points']) {
      expect(screen.getByRole('columnheader', { name: encabezado })).toBeInTheDocument()
    }
    const filas = screen.getAllByRole('row').slice(1)
    expect(within(filas[0]).getByText('US-02')).toBeInTheDocument()
    expect(within(filas[1]).getByText('US-01')).toBeInTheDocument()
    expect(within(filas[1]).getByText('3')).toBeInTheDocument()
  })

  it('E-03 un sprint activo sin historias muestra el estado vacío y la acción de asignar', async () => {
    obtenerMock.mockResolvedValue(crearBacklog('Activo'))

    renderizar()

    expect(await screen.findByText('Este sprint todavía no tiene historias')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Asignar historia' })).toBeInTheDocument()
  })

  it('E-04 un sprint cerrado sin historias muestra el vacío sin acción de asignar', async () => {
    obtenerMock.mockResolvedValue(crearBacklog('Cerrado'))

    renderizar()

    expect(await screen.findByText('Este sprint todavía no tiene historias')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Asignar historia' })).not.toBeInTheDocument()
  })

  it('E-09 un sprint planificado avisa que hay que activarlo y no permite asignar', async () => {
    obtenerMock.mockResolvedValue(crearBacklog('Planificado'))

    renderizar()

    expect(
      await screen.findByText('Activá el sprint para poder asignar historias'),
    ).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Asignar historia' })).not.toBeInTheDocument()
  })

  it('E-10 un sprint cerrado es de solo lectura: ni asignar ni quitar', async () => {
    obtenerMock.mockResolvedValue(crearBacklog('Cerrado', [us01]))

    renderizar()

    expect(await screen.findByText('US-01')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Asignar historia' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /Quitar/ })).not.toBeInTheDocument()
  })

  describe('asignar historia', () => {
    it('E-05 / E-06 asigna una historia en Backlog y refresca la tabla sin recargar', async () => {
      obtenerMock
        .mockResolvedValueOnce(crearBacklog('Activo', [us01]))
        .mockResolvedValueOnce(crearBacklog('Activo', [us01, us05]))
      asignarMock.mockResolvedValue(undefined)
      listarHistoriasMock.mockResolvedValue([us01, us05, us02])
      const usuario = userEvent.setup()
      renderizar()

      await usuario.click(await screen.findByRole('button', { name: 'Asignar historia' }))
      const dialogo = await screen.findByRole('dialog', { name: 'Asignar historia' })

      // US-01 ya está en el sprint y US-02 no está en Backlog: no se ofrecen.
      expect(await within(dialogo).findByRole('radio', { name: /US-05/ })).toBeInTheDocument()
      expect(within(dialogo).queryByRole('radio', { name: /US-01/ })).not.toBeInTheDocument()
      expect(within(dialogo).queryByRole('radio', { name: /US-02/ })).not.toBeInTheDocument()

      await usuario.click(within(dialogo).getByRole('radio', { name: /US-05/ }))
      await usuario.click(within(dialogo).getByRole('button', { name: 'Asignar' }))

      await waitFor(() => expect(asignarMock).toHaveBeenCalledWith(1, 1, 5))
      expect(await screen.findByText('US-05')).toBeInTheDocument()
      expect(obtenerMock).toHaveBeenCalledTimes(2)
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    })

    it('E-20 sin historias disponibles muestra el aviso en el modal', async () => {
      obtenerMock.mockResolvedValue(crearBacklog('Activo', [us01]))
      listarHistoriasMock.mockResolvedValue([us01])
      const usuario = userEvent.setup()
      renderizar()

      await usuario.click(await screen.findByRole('button', { name: 'Asignar historia' }))

      expect(
        await screen.findByText('No hay historias disponibles para asignar'),
      ).toBeInTheDocument()
    })

    it('E-07 si la historia ya está asignada muestra el error del backend y no cambia la tabla', async () => {
      obtenerMock.mockResolvedValue(crearBacklog('Activo', [us01]))
      listarHistoriasMock.mockResolvedValue([us05])
      asignarMock.mockRejectedValue(
        new ApiError({
          status: 409,
          code: 'HISTORIA_YA_ASIGNADA',
          message: 'la historia ya está asignada a un Sprint',
        }),
      )
      const usuario = userEvent.setup()
      renderizar()

      await usuario.click(await screen.findByRole('button', { name: 'Asignar historia' }))
      await usuario.click(await screen.findByRole('radio', { name: /US-05/ }))
      await usuario.click(screen.getByRole('button', { name: 'Asignar' }))

      expect(await screen.findByRole('alert')).toHaveTextContent(
        'la historia ya está asignada a un Sprint',
      )
      expect(screen.getByRole('dialog')).toBeInTheDocument()
      expect(obtenerMock).toHaveBeenCalledTimes(1)
      const tabla = screen.getByRole('table')
      expect(within(tabla).queryByText('US-05')).not.toBeInTheDocument()
    })

    it('E-08 si la historia no está pendiente muestra el error del backend', async () => {
      obtenerMock.mockResolvedValue(crearBacklog('Activo'))
      listarHistoriasMock.mockResolvedValue([us05])
      asignarMock.mockRejectedValue(
        new ApiError({
          status: 400,
          code: 'HISTORIA_NO_PENDIENTE',
          message: 'solo se pueden asignar historias en estado pendiente',
        }),
      )
      const usuario = userEvent.setup()
      renderizar()

      await usuario.click(await screen.findByRole('button', { name: 'Asignar historia' }))
      await usuario.click(await screen.findByRole('radio', { name: /US-05/ }))
      await usuario.click(screen.getByRole('button', { name: 'Asignar' }))

      expect(await screen.findByRole('alert')).toHaveTextContent(
        'solo se pueden asignar historias en estado pendiente',
      )
    })

    it('E-13 una asignación fuera de la ventana temporal muestra el error y no agrega la historia', async () => {
      obtenerMock.mockResolvedValue(crearBacklog('Activo', [us01]))
      listarHistoriasMock.mockResolvedValue([us05])
      asignarMock.mockRejectedValue(
        new ApiError({
          status: 400,
          code: 'FUERA_DE_VENTANA_TEMPORAL',
          message: 'la asignación está fuera de la ventana temporal del Sprint',
        }),
      )
      const usuario = userEvent.setup()
      renderizar()

      await usuario.click(await screen.findByRole('button', { name: 'Asignar historia' }))
      await usuario.click(await screen.findByRole('radio', { name: /US-05/ }))
      await usuario.click(screen.getByRole('button', { name: 'Asignar' }))

      expect(await screen.findByRole('alert')).toHaveTextContent(
        'la asignación está fuera de la ventana temporal del Sprint',
      )
      expect(within(screen.getByRole('table')).queryByText('US-05')).not.toBeInTheDocument()
    })

    it('E-16 un error de red al asignar muestra un mensaje y permite reintentar', async () => {
      obtenerMock.mockResolvedValue(crearBacklog('Activo'))
      listarHistoriasMock.mockResolvedValue([us05])
      asignarMock.mockRejectedValueOnce(errorDeRed()).mockResolvedValueOnce(undefined)
      const usuario = userEvent.setup()
      renderizar()

      await usuario.click(await screen.findByRole('button', { name: 'Asignar historia' }))
      await usuario.click(await screen.findByRole('radio', { name: /US-05/ }))
      await usuario.click(screen.getByRole('button', { name: 'Asignar' }))

      expect(await screen.findByRole('alert')).toBeInTheDocument()
      expect(screen.getByRole('button', { name: 'Asignar' })).toBeEnabled()

      await usuario.click(screen.getByRole('button', { name: 'Asignar' }))

      await waitFor(() => expect(asignarMock).toHaveBeenCalledTimes(2))
    })

    it('no envía dos veces la asignación mientras la primera está en curso', async () => {
      obtenerMock.mockResolvedValue(crearBacklog('Activo'))
      listarHistoriasMock.mockResolvedValue([us05])
      asignarMock.mockReturnValue(new Promise(() => {}))
      const usuario = userEvent.setup()
      renderizar()

      await usuario.click(await screen.findByRole('button', { name: 'Asignar historia' }))
      await usuario.click(await screen.findByRole('radio', { name: /US-05/ }))
      await usuario.click(screen.getByRole('button', { name: 'Asignar' }))
      await usuario.click(screen.getByRole('button', { name: 'Asignar' }))

      expect(asignarMock).toHaveBeenCalledTimes(1)
    })
  })

  describe('quitar del sprint', () => {
    it('E-11 pide confirmación y, al confirmar, quita la historia de la tabla', async () => {
      obtenerMock
        .mockResolvedValueOnce(crearBacklog('Activo', [us01, us02]))
        .mockResolvedValueOnce(crearBacklog('Activo', [us02]))
      quitarMock.mockResolvedValue(undefined)
      const usuario = userEvent.setup()
      renderizar()

      await usuario.click(await screen.findByRole('button', { name: 'Quitar US-01 del sprint' }))
      const dialogo = await screen.findByRole('dialog')
      expect(quitarMock).not.toHaveBeenCalled()
      await usuario.click(within(dialogo).getByRole('button', { name: 'Quitar' }))

      await waitFor(() => expect(quitarMock).toHaveBeenCalledWith(1, 1, 1))
      await waitFor(() =>
        expect(within(screen.getByRole('table')).queryByText('US-01')).not.toBeInTheDocument(),
      )
      expect(screen.getByText('US-02')).toBeInTheDocument()
    })

    it('E-12 al cancelar la confirmación no se envía nada y la historia sigue', async () => {
      obtenerMock.mockResolvedValue(crearBacklog('Activo', [us01]))
      const usuario = userEvent.setup()
      renderizar()

      await usuario.click(await screen.findByRole('button', { name: 'Quitar US-01 del sprint' }))
      await usuario.click(within(await screen.findByRole('dialog')).getByRole('button', { name: 'Cancelar' }))

      expect(quitarMock).not.toHaveBeenCalled()
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
      expect(screen.getByText('US-01')).toBeInTheDocument()
    })

    it('un sprint planificado permite quitar historias', async () => {
      obtenerMock.mockResolvedValue(crearBacklog('Planificado', [us01]))

      renderizar()

      expect(
        await screen.findByRole('button', { name: 'Quitar US-01 del sprint' }),
      ).toBeInTheDocument()
    })

    it('E-16 un error de red al quitar muestra un mensaje y la historia sigue en la tabla', async () => {
      obtenerMock.mockResolvedValue(crearBacklog('Activo', [us01]))
      quitarMock.mockRejectedValue(errorDeRed())
      const usuario = userEvent.setup()
      renderizar()

      await usuario.click(await screen.findByRole('button', { name: 'Quitar US-01 del sprint' }))
      await usuario.click(within(await screen.findByRole('dialog')).getByRole('button', { name: 'Quitar' }))

      expect(await screen.findByRole('alert')).toBeInTheDocument()
      expect(screen.getByText('US-01')).toBeInTheDocument()
      expect(obtenerMock).toHaveBeenCalledTimes(1)
    })

    it('muestra el error del backend si el sprint ya está cerrado', async () => {
      obtenerMock.mockResolvedValue(crearBacklog('Activo', [us01]))
      quitarMock.mockRejectedValue(
        new ApiError({
          status: 409,
          code: 'SPRINT_CERRADO',
          message: 'no se pueden modificar Sprints cerrados',
        }),
      )
      const usuario = userEvent.setup()
      renderizar()

      await usuario.click(await screen.findByRole('button', { name: 'Quitar US-01 del sprint' }))
      await usuario.click(within(await screen.findByRole('dialog')).getByRole('button', { name: 'Quitar' }))

      expect(await screen.findByRole('alert')).toHaveTextContent(
        'no se pueden modificar Sprints cerrados',
      )
    })
  })

  describe('errores al cargar', () => {
    it('E-14 un sprint inexistente o de otro proyecto muestra sprint no encontrado', async () => {
      obtenerMock.mockRejectedValue(
        new ApiError({
          status: 404,
          code: 'SPRINT_NO_ENCONTRADA',
          message: 'el sprint no existe',
        }),
      )

      renderizar('99')

      expect(await screen.findByText('Sprint no encontrado')).toBeInTheDocument()
      const volver = screen.getByRole('link', { name: 'Volver a sprints' })
      expect(volver).toHaveAttribute('href', '/proyectos/1/sprints')
    })

    it('un proyecto sin acceso muestra el mensaje de acceso', async () => {
      obtenerMock.mockRejectedValue(
        new ApiError({
          status: 404,
          code: 'PROYECTO_NO_ENCONTRADO',
          message: 'el proyecto no existe',
        }),
      )

      renderizar()

      expect(
        await screen.findByText('No tenés acceso a este proyecto o no existe'),
      ).toBeInTheDocument()
    })

    it('un sprintId no numérico se trata como no encontrado sin llamar a la API', async () => {
      renderizar('abc')

      expect(await screen.findByText('Sprint no encontrado')).toBeInTheDocument()
      expect(obtenerMock).not.toHaveBeenCalled()
    })

    it('E-15 un error de red al listar muestra el mensaje y permite reintentar', async () => {
      obtenerMock
        .mockRejectedValueOnce(errorDeRed())
        .mockResolvedValueOnce(crearBacklog('Activo', [us01]))
      const usuario = userEvent.setup()
      renderizar()

      expect(
        await screen.findByText('No se pudo cargar el sprint. Intentá nuevamente.'),
      ).toBeInTheDocument()
      await usuario.click(screen.getByRole('button', { name: 'Reintentar' }))

      expect(await screen.findByText('US-01')).toBeInTheDocument()
      expect(obtenerMock).toHaveBeenCalledTimes(2)
    })
  })

  it('E-19 no ofrece editar goal o fechas, ni crear o cerrar el sprint', async () => {
    obtenerMock.mockResolvedValue(crearBacklog('Activo', [us01]))

    renderizar()

    await screen.findByText('US-01')
    expect(screen.queryByRole('button', { name: /editar/i })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /cerrar sprint/i })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /nuevo sprint/i })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /activar/i })).not.toBeInTheDocument()
  })

  it('ofrece volver al listado de sprints', async () => {
    obtenerMock.mockResolvedValue(crearBacklog('Activo', [us01]))

    renderizar()

    const volver = await screen.findByRole('link', { name: 'Volver a sprints' })
    expect(volver).toHaveAttribute('href', '/proyectos/1/sprints')
  })
})
