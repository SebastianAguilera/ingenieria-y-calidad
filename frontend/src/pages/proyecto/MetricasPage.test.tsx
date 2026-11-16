import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiError } from '../../api/errors'
import { listarHistorias } from '../../api/historias'
import { obtenerMetricasProyecto, obtenerMetricasSprint } from '../../api/metricas'
import { listarSprints } from '../../api/sprints'
import type {
  EstadoHistoria,
  EstadoSprint,
  Historia,
  MetricasProyecto,
  MetricasSprint,
  Sprint,
} from '../../api/types'
import MetricasPage from './MetricasPage'

vi.mock('../../api/metricas', () => ({
  obtenerMetricasProyecto: vi.fn(),
  obtenerMetricasSprint: vi.fn(),
}))
vi.mock('../../api/historias', () => ({ listarHistorias: vi.fn() }))
vi.mock('../../api/sprints', () => ({ listarSprints: vi.fn() }))

const metricasProyectoMock = vi.mocked(obtenerMetricasProyecto)
const metricasSprintMock = vi.mocked(obtenerMetricasSprint)
const listarHistoriasMock = vi.mocked(listarHistorias)
const listarSprintsMock = vi.mocked(listarSprints)

const metricasProyecto: MetricasProyecto = {
  proyecto_id: 1,
  base: 'estado_actual',
  velocidad: 8,
  porcentaje_historias_completadas: 0.5,
  horas_estimadas: 40,
  horas_reales: 48,
  desviacion_esfuerzo: 0.2,
  defectos_detectados: 4,
  defectos_resueltos: 3,
  ratio_defectos: 0.25,
  sprints_restantes: null,
  wip: 1,
  tasa_entrega: null,
  lead_time_estimado: null,
  lead_time_aproximado: false,
  defectos_abiertos: 1,
  defectos_por_severidad: { minor: 1, major: 2, critical: 1, blocker: 0 },
}

const metricasNulas: MetricasProyecto = {
  ...metricasProyecto,
  velocidad: null,
  porcentaje_historias_completadas: null,
  desviacion_esfuerzo: null,
  ratio_defectos: null,
  horas_estimadas: 0,
  horas_reales: 0,
  defectos_detectados: 0,
  defectos_resueltos: 0,
}

function metricasDeSprint(
  sprintId: number,
  planificados: number,
  completados: number,
  porcentaje: number | null,
): MetricasSprint {
  return {
    sprint_id: sprintId,
    base: 'estado_actual',
    story_points_planificados: planificados,
    story_points_completados: completados,
    porcentaje_completado: porcentaje,
    historias_completadas: completados > 0 ? 2 : 0,
    horas_estimadas: 20,
    horas_reales: 24,
    desviacion_esfuerzo: 0.2,
    defectos_detectados: 2,
    defectos_resueltos: 1,
  }
}

const METRICAS_POR_SPRINT: Record<number, MetricasSprint> = {
  1: metricasDeSprint(1, 13, 8, 0.6153846153846154),
  2: metricasDeSprint(2, 10, 2, 0.2),
  3: metricasDeSprint(3, 0, 0, null),
}

function sprint(id: number, estado: EstadoSprint): Sprint {
  return {
    id,
    identificador: `Sprint ${id}`,
    sprint_goal: '',
    estado,
    fecha_inicio: '2026-05-01',
    fecha_fin: '2026-05-15',
    creado_en: '2026-04-01T00:00:00Z',
    actualizado_en: '2026-04-01T00:00:00Z',
  }
}

const SPRINTS: Sprint[] = [sprint(1, 'Cerrado'), sprint(2, 'Activo'), sprint(3, 'Planificado')]

function historia(id: number, estado: EstadoHistoria): Historia {
  return {
    id,
    codigo: `US-${String(id).padStart(2, '0')}`,
    titulo: `Historia ${id}`,
    descripcion: '',
    prioridad: 'Media',
    estado,
    story_points: 3,
    esfuerzo_real: null,
    horas_estimadas: null,
    creado_en: '2026-04-01T00:00:00Z',
    actualizado_en: '2026-04-01T00:00:00Z',
  }
}

const HISTORIAS: Historia[] = [
  historia(1, 'Backlog'),
  historia(2, 'Backlog'),
  historia(3, 'En Progreso'),
  historia(4, 'Hecho'),
  historia(5, 'Hecho'),
  historia(6, 'Hecho'),
]

function renderizar() {
  return render(
    <MemoryRouter initialEntries={['/proyectos/1/metricas']}>
      <Routes>
        <Route path="/proyectos/:id/metricas" element={<MetricasPage />} />
      </Routes>
    </MemoryRouter>,
  )
}

function tarjeta(nombre: string) {
  return screen.getByRole('group', { name: nombre })
}

async function esperarDashboard() {
  return screen.findByRole('group', { name: 'Velocidad' })
}

describe('MetricasPage', () => {
  beforeEach(() => {
    metricasProyectoMock.mockReset()
    metricasSprintMock.mockReset()
    listarHistoriasMock.mockReset()
    listarSprintsMock.mockReset()

    metricasProyectoMock.mockResolvedValue(metricasProyecto)
    metricasSprintMock.mockImplementation(async (_proyectoId, sprintId) => {
      const metricas = METRICAS_POR_SPRINT[sprintId]
      if (!metricas) {
        throw new Error(`sprint inesperado ${sprintId}`)
      }
      return metricas
    })
    listarHistoriasMock.mockResolvedValue(HISTORIAS)
    listarSprintsMock.mockResolvedValue(SPRINTS)
  })

  describe('tramo 1 - KPI', () => {
    it('E-01 muestra las tarjetas de KPI con los valores de GET /metricas', async () => {
      renderizar()
      await esperarDashboard()

      expect(metricasProyectoMock).toHaveBeenCalledWith(1, expect.any(AbortSignal))
      expect(within(tarjeta('Velocidad')).getByText('8 SP')).toBeInTheDocument()
      expect(within(tarjeta('Historias completadas')).getByText('50%')).toBeInTheDocument()
      expect(within(tarjeta('Desviación de esfuerzo')).getByText('+20%')).toBeInTheDocument()
      expect(
        within(tarjeta('Desviación de esfuerzo')).getByText('48 h reales / 40 h estimadas'),
      ).toBeInTheDocument()
      expect(within(tarjeta('Ratio de defectos')).getByText('0,25')).toBeInTheDocument()
      expect(within(tarjeta('Defectos detectados')).getByText('4')).toBeInTheDocument()
      expect(within(tarjeta('Defectos resueltos')).getByText('3')).toBeInTheDocument()
    })

    it('E-02 muestra "Sin datos" con explicación para las métricas nulas', async () => {
      metricasProyectoMock.mockResolvedValue(metricasNulas)
      renderizar()
      await esperarDashboard()

      const casos: Array<[string, string]> = [
        ['Velocidad', 'Todavía no hay sprints cerrados'],
        ['Historias completadas', 'El proyecto no tiene historias'],
        ['Desviación de esfuerzo', 'Ninguna historia tiene horas estimadas y reales'],
        ['Ratio de defectos', 'Todavía no hay historias completadas'],
      ]
      for (const [nombre, explicacion] of casos) {
        const tarjetaKpi = tarjeta(nombre)
        expect(within(tarjetaKpi).getByText('Sin datos')).toBeInTheDocument()
        expect(within(tarjetaKpi).getByText(explicacion)).toBeInTheDocument()
        for (const cero of ['0', '0%', '0 SP', '0,00']) {
          expect(within(tarjetaKpi).queryByText(cero)).not.toBeInTheDocument()
        }
      }
    })

    it('E-03 un cero real se muestra como 0 y no como "Sin datos"', async () => {
      metricasProyectoMock.mockResolvedValue({ ...metricasProyecto, velocidad: 0 })
      renderizar()
      await esperarDashboard()

      expect(within(tarjeta('Velocidad')).getByText('0 SP')).toBeInTheDocument()
      expect(within(tarjeta('Velocidad')).queryByText('Sin datos')).not.toBeInTheDocument()
    })

    it('E-04 multiplica por 100 el porcentaje de la API', async () => {
      metricasProyectoMock.mockResolvedValue({
        ...metricasProyecto,
        porcentaje_historias_completadas: 0.75,
      })
      renderizar()
      await esperarDashboard()

      expect(within(tarjeta('Historias completadas')).getByText('75%')).toBeInTheDocument()
    })

    it('E-05 muestra decimales con coma y la desviación negativa con signo', async () => {
      metricasProyectoMock.mockResolvedValue({
        ...metricasProyecto,
        porcentaje_historias_completadas: 0.6153846153846154,
        desviacion_esfuerzo: -0.1,
      })
      renderizar()
      await esperarDashboard()

      expect(within(tarjeta('Historias completadas')).getByText('61,5%')).toBeInTheDocument()
      expect(within(tarjeta('Desviación de esfuerzo')).getByText('−10%')).toBeInTheDocument()
    })
  })

  describe('tramo 1 - gráficos', () => {
    it('E-06 el gráfico de barras tiene planificados y completados por sprint cerrado y activo', async () => {
      renderizar()
      const tabla = await screen.findByRole('table', { name: 'Story points por sprint' })

      const encabezados = within(tabla)
        .getAllByRole('columnheader')
        .map((celda) => celda.textContent)
      expect(encabezados).toEqual(['Sprint', 'SP planificados', 'SP completados'])

      const filas = within(tabla).getAllByRole('row').slice(1)
      expect(filas.map((fila) => fila.textContent)).toEqual(['Sprint 1138', 'Sprint 2102'])
      expect(within(tabla).queryByText('Sprint 3')).not.toBeInTheDocument()
      expect(metricasSprintMock).toHaveBeenCalledWith(1, 1, expect.any(AbortSignal))
      expect(metricasSprintMock).toHaveBeenCalledWith(1, 2, expect.any(AbortSignal))
    })

    it('E-07 sin sprints cerrados ni activos el gráfico de barras muestra un estado vacío', async () => {
      listarSprintsMock.mockResolvedValue([sprint(3, 'Planificado')])
      renderizar()
      await esperarDashboard()

      expect(
        await screen.findByText('Todavía no hay sprints cerrados ni activos'),
      ).toBeInTheDocument()
      expect(
        screen.queryByRole('table', { name: 'Story points por sprint' }),
      ).not.toBeInTheDocument()
    })

    it('E-08 la dona muestra la cantidad y la etiqueta de cada estado', async () => {
      renderizar()
      const tabla = await screen.findByRole('table', { name: 'Historias por estado' })

      const filas = within(tabla).getAllByRole('row').slice(1)
      expect(filas.map((fila) => fila.textContent)).toEqual([
        'Backlog2',
        'En Progreso1',
        'Hecho3',
      ])

      const leyenda = screen.getByRole('list', { name: 'Leyenda de historias por estado' })
      expect(within(leyenda).getByText('Backlog')).toBeInTheDocument()
      expect(within(leyenda).getByText('En Progreso')).toBeInTheDocument()
      expect(within(leyenda).getByText('Hecho')).toBeInTheDocument()
    })

    it('E-13 cada gráfico tiene una tabla que se puede mostrar con "Ver como tabla"', async () => {
      const usuario = userEvent.setup()
      renderizar()
      await screen.findByRole('table', { name: 'Story points por sprint' })

      // Dos gráficos de US-F21 más el de defectos por severidad de US-F22.
      const botones = screen.getAllByRole('button', { name: 'Ver como tabla' })
      expect(botones).toHaveLength(3)
      expect(botones[0]).toHaveAttribute('aria-expanded', 'false')

      await usuario.click(botones[0])

      expect(screen.getByRole('button', { name: 'Ocultar tabla' })).toHaveAttribute(
        'aria-expanded',
        'true',
      )
      expect(screen.getByRole('table', { name: 'Story points por sprint' })).not.toHaveClass(
        'sr-only',
      )
      expect(screen.getByRole('table', { name: 'Historias por estado' })).toHaveClass('sr-only')
    })
  })

  describe('tramo 1 - panel de sprint', () => {
    it('E-09 elige el sprint activo y muestra sus métricas', async () => {
      renderizar()
      await esperarDashboard()

      const selector = await screen.findByLabelText('Sprint')
      expect(selector).toHaveValue('2')

      const panel = screen.getByRole('region', { name: 'Detalle por sprint' })
      await waitFor(() =>
        expect(within(panel).getByRole('group', { name: 'Completado' })).toHaveTextContent('20%'),
      )
      expect(within(panel).getByRole('group', { name: 'SP planificados' })).toHaveTextContent(
        '10',
      )
    })

    it('E-10 al cambiar de sprint pide y muestra sus métricas', async () => {
      const usuario = userEvent.setup()
      renderizar()
      await esperarDashboard()

      await usuario.selectOptions(await screen.findByLabelText('Sprint'), '1')

      const panel = screen.getByRole('region', { name: 'Detalle por sprint' })
      await waitFor(() =>
        expect(within(panel).getByRole('group', { name: 'Completado' })).toHaveTextContent(
          '61,5%',
        ),
      )
      expect(within(panel).getByRole('group', { name: 'SP planificados' })).toHaveTextContent(
        '13',
      )
      expect(within(panel).getByRole('group', { name: 'SP completados' })).toHaveTextContent('8')
      expect(within(panel).getByRole('group', { name: 'Horas estimadas' })).toHaveTextContent(
        '20 h',
      )
      expect(within(panel).getByRole('group', { name: 'Horas reales' })).toHaveTextContent(
        '24 h',
      )
      expect(within(panel).getByRole('group', { name: 'Desviación' })).toHaveTextContent('+20%')
    })

    it('E-11 un sprint sin story points muestra "Sin datos" en el porcentaje', async () => {
      const usuario = userEvent.setup()
      renderizar()
      await esperarDashboard()

      await usuario.selectOptions(await screen.findByLabelText('Sprint'), '3')

      const panel = screen.getByRole('region', { name: 'Detalle por sprint' })
      await waitFor(() =>
        expect(metricasSprintMock).toHaveBeenCalledWith(1, 3, expect.any(AbortSignal)),
      )
      const completado = await within(panel).findByRole('group', { name: 'Completado' })
      await waitFor(() => expect(completado).toHaveTextContent('Sin datos'))
      expect(completado).toHaveTextContent('El sprint no tiene story points planificados')
    })

    it('sin sprints el panel avisa que todavía no hay sprints', async () => {
      listarSprintsMock.mockResolvedValue([])
      renderizar()
      await esperarDashboard()

      const panel = screen.getByRole('region', { name: 'Detalle por sprint' })
      expect(within(panel).getByText('Todavía no hay sprints')).toBeInTheDocument()
      expect(within(panel).queryByLabelText('Sprint')).not.toBeInTheDocument()
    })

    it('si fallan las métricas del sprint muestra el error en el panel y permite reintentar', async () => {
      const usuario = userEvent.setup()
      // La primera consulta del sprint 2 es la del gráfico; la segunda, la del panel.
      let fallar = true
      let consultasSprint2 = 0
      metricasSprintMock.mockImplementation(async (_proyectoId, sprintId) => {
        if (sprintId === 2) {
          consultasSprint2 += 1
        }
        if (sprintId === 2 && consultasSprint2 > 1 && fallar) {
          throw new ApiError({ status: 500, code: 'ERROR_INTERNO', message: 'fallo' })
        }
        return METRICAS_POR_SPRINT[sprintId]
      })
      renderizar()
      await esperarDashboard()

      const panel = screen.getByRole('region', { name: 'Detalle por sprint' })
      expect(
        await within(panel).findByText('No se pudieron cargar las métricas del sprint.'),
      ).toBeInTheDocument()

      fallar = false
      await usuario.click(within(panel).getByRole('button', { name: 'Reintentar' }))

      await waitFor(() =>
        expect(within(panel).getByRole('group', { name: 'Completado' })).toHaveTextContent('20%'),
      )
    })
  })

  describe('tramo 1 - estados', () => {
    it('E-12 un proyecto sin historias muestra el estado vacío', async () => {
      listarHistoriasMock.mockResolvedValue([])
      renderizar()

      expect(await screen.findByText('Todavía no hay historias')).toBeInTheDocument()
      expect(screen.getByRole('link', { name: 'Ir al backlog' })).toHaveAttribute(
        'href',
        '/proyectos/1/backlog',
      )
      expect(screen.queryByRole('group', { name: 'Velocidad' })).not.toBeInTheDocument()
    })

    it('E-14 muestra el aviso de cálculo sobre el estado actual', async () => {
      renderizar()
      await esperarDashboard()

      expect(
        screen.getByText('Calculado sobre el estado actual de las historias'),
      ).toBeInTheDocument()
    })

    it('muestra el estado de carga mientras espera la respuesta', () => {
      metricasProyectoMock.mockReturnValue(new Promise(() => {}))
      renderizar()

      expect(screen.getByRole('status')).toHaveTextContent('Cargando métricas…')
    })

    it('E-15 un no miembro ve el mensaje de acceso', async () => {
      metricasProyectoMock.mockRejectedValue(
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
    })

    it('E-16 ante un error de red permite reintentar', async () => {
      const usuario = userEvent.setup()
      metricasProyectoMock.mockRejectedValueOnce(
        new ApiError({ status: 0, code: 'NETWORK_ERROR', message: 'Failed to fetch' }),
      )
      renderizar()

      expect(
        await screen.findByText('No se pudieron cargar las métricas. Intentá nuevamente.'),
      ).toBeInTheDocument()

      await usuario.click(screen.getByRole('button', { name: 'Reintentar' }))

      expect(await esperarDashboard()).toBeInTheDocument()
    })
  })
})
