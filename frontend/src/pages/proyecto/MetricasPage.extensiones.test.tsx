import { render, screen, within } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiError } from '../../api/errors'
import { listarHistorias } from '../../api/historias'
import { obtenerMetricasProyecto } from '../../api/metricas'
import { listarSprints } from '../../api/sprints'
import type { Historia, MetricasProyecto } from '../../api/types'
import MetricasPage from './MetricasPage'

// US-F22: extensiones de métricas (flujo y defectos por severidad) sobre la
// pantalla de US-F21. Cubre E-01 a E-20 de
// docs/features/US-F22-extensiones-metricas.feature.

vi.mock('../../api/metricas', () => ({
  obtenerMetricasProyecto: vi.fn(),
  obtenerMetricasSprint: vi.fn(),
}))
vi.mock('../../api/historias', () => ({ listarHistorias: vi.fn() }))
vi.mock('../../api/sprints', () => ({ listarSprints: vi.fn() }))

const metricasProyectoMock = vi.mocked(obtenerMetricasProyecto)
const listarHistoriasMock = vi.mocked(listarHistorias)
const listarSprintsMock = vi.mocked(listarSprints)

const BASE: MetricasProyecto = {
  proyecto_id: 1,
  base: 'estado_actual',
  velocidad: 6,
  porcentaje_historias_completadas: 0.5,
  horas_estimadas: 40,
  horas_reales: 48,
  desviacion_esfuerzo: 0.2,
  defectos_detectados: 3,
  defectos_resueltos: 2,
  ratio_defectos: 0.33,
  sprints_restantes: 7,
  wip: 6,
  tasa_entrega: 2,
  lead_time_estimado: 3,
  lead_time_aproximado: true,
  defectos_abiertos: 4,
  defectos_por_severidad: { minor: 1, major: 1, critical: 1, blocker: 0 },
}

const CAMPOS_US21 = [
  'sprints_restantes',
  'wip',
  'tasa_entrega',
  'lead_time_estimado',
  'lead_time_aproximado',
  'defectos_abiertos',
  'defectos_por_severidad',
] as const

function historia(id: number): Historia {
  return {
    id,
    codigo: `US-${String(id).padStart(2, '0')}`,
    titulo: `Historia ${id}`,
    descripcion: '',
    prioridad: 'Media',
    estado: 'Hecho',
    story_points: 3,
    esfuerzo_real: null,
    horas_estimadas: null,
    valor_negocio: null,
    criticidad_tiempo: null,
    reduccion_riesgo: null,
    creado_en: '2026-04-01T00:00:00Z',
    actualizado_en: '2026-04-01T00:00:00Z',
  }
}

function conMetricas(cambios: Partial<MetricasProyecto>) {
  metricasProyectoMock.mockResolvedValue({ ...BASE, ...cambios })
}

function sinCamposUs21(): MetricasProyecto {
  const copia: Record<string, unknown> = { ...BASE }
  for (const campo of CAMPOS_US21) {
    delete copia[campo]
  }
  return copia as unknown as MetricasProyecto
}

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

function graficoSeveridad() {
  return screen.getByRole('region', { name: 'Defectos por severidad' })
}

async function esperarDashboard() {
  return screen.findByRole('group', { name: 'Velocidad' })
}

describe('MetricasPage - extensiones de US-F22', () => {
  beforeEach(() => {
    metricasProyectoMock.mockReset()
    listarHistoriasMock.mockReset()
    listarSprintsMock.mockReset()

    metricasProyectoMock.mockResolvedValue(BASE)
    listarHistoriasMock.mockResolvedValue([historia(1), historia(2)])
    listarSprintsMock.mockResolvedValue([])
  })

  describe('indicadores de flujo', () => {
    it('E-01 sprints restantes se muestra en plural (CA-01)', async () => {
      conMetricas({ sprints_restantes: 7 })
      renderizar()
      await esperarDashboard()

      expect(within(tarjeta('Sprints restantes')).getByText('7 sprints')).toBeInTheDocument()
    })

    it('E-02 sprints restantes se muestra en singular y en cero (CA-01)', async () => {
      conMetricas({ sprints_restantes: 1 })
      const { unmount } = renderizar()
      await esperarDashboard()
      expect(within(tarjeta('Sprints restantes')).getByText('1 sprint')).toBeInTheDocument()
      unmount()

      conMetricas({ sprints_restantes: 0 })
      renderizar()
      await esperarDashboard()
      expect(within(tarjeta('Sprints restantes')).getByText('0 sprints')).toBeInTheDocument()
    })

    it('E-03 sprints restantes null muestra "Sin datos" y no 0 (CA-02)', async () => {
      conMetricas({ sprints_restantes: null })
      renderizar()
      await esperarDashboard()

      const kpi = tarjeta('Sprints restantes')
      expect(within(kpi).getByText('Sin datos')).toBeInTheDocument()
      expect(
        within(kpi).getByText('Hace falta la velocidad de al menos un sprint cerrado.'),
      ).toBeInTheDocument()
      expect(within(kpi).queryByText(/^0/)).not.toBeInTheDocument()
    })

    it('E-04 lead time muestra la etiqueta "aprox." y la nota de la Ley de Little (CA-03)', async () => {
      conMetricas({ lead_time_estimado: 3, lead_time_aproximado: true })
      renderizar()
      await esperarDashboard()

      const kpi = tarjeta('Lead Time estimado')
      expect(within(kpi).getByText('3 sprints')).toBeInTheDocument()
      expect(within(kpi).getByText('aprox.')).toBeInTheDocument()
      expect(within(kpi).getByText(/Ley de Little/)).toBeInTheDocument()
      expect(within(kpi).getByText(/no una medición/)).toBeInTheDocument()
    })

    it('E-05 lead time null muestra "Sin datos", sin "aprox." y sin 0 (CA-04)', async () => {
      conMetricas({ lead_time_estimado: null, lead_time_aproximado: false })
      renderizar()
      await esperarDashboard()

      const kpi = tarjeta('Lead Time estimado')
      expect(within(kpi).getByText('Sin datos')).toBeInTheDocument()
      expect(
        within(kpi).getByText('Hace falta la tasa de entrega de al menos un sprint cerrado.'),
      ).toBeInTheDocument()
      expect(within(kpi).queryByText('aprox.')).not.toBeInTheDocument()
      expect(within(kpi).queryByText(/^0/)).not.toBeInTheDocument()
    })

    it('E-06 tasa de entrega null es "Sin datos" y un 0 real se muestra (CA-04)', async () => {
      conMetricas({ tasa_entrega: null })
      const { unmount } = renderizar()
      await esperarDashboard()
      expect(within(tarjeta('Tasa de entrega')).getByText('Sin datos')).toBeInTheDocument()
      unmount()

      conMetricas({ tasa_entrega: 0 })
      renderizar()
      await esperarDashboard()
      const kpi = tarjeta('Tasa de entrega')
      expect(within(kpi).getByText('0')).toBeInTheDocument()
      expect(within(kpi).queryByText('Sin datos')).not.toBeInTheDocument()
    })

    it('E-07 WiP coincide con la API (CA-05)', async () => {
      conMetricas({ wip: 6 })
      renderizar()
      await esperarDashboard()

      expect(within(tarjeta('WiP')).getByText('6')).toBeInTheDocument()
    })

    it('E-08 WiP en cero se muestra como cero (CA-05)', async () => {
      conMetricas({ wip: 0 })
      renderizar()
      await esperarDashboard()

      const kpi = tarjeta('WiP')
      expect(within(kpi).getByText('0')).toBeInTheDocument()
      expect(within(kpi).queryByText('Sin datos')).not.toBeInTheDocument()
    })
  })

  describe('defectos abiertos', () => {
    it('E-09 coincide con la API y aclara qué cuenta (CA-05)', async () => {
      conMetricas({ defectos_abiertos: 4 })
      renderizar()
      await esperarDashboard()

      const kpi = tarjeta('Defectos abiertos')
      expect(within(kpi).getByText('4')).toBeInTheDocument()
      expect(within(kpi).getByText(/estado Abierto/)).toBeInTheDocument()
    })
  })

  describe('gráfico de defectos por severidad', () => {
    it('E-10 blocker aparece con valor 0 (CA-06)', async () => {
      conMetricas({ defectos_por_severidad: { minor: 1, major: 1, critical: 1, blocker: 0 } })
      renderizar()
      await esperarDashboard()

      const fila = within(graficoSeveridad()).getByRole('row', { name: /Bloqueante/ })
      expect(within(fila).getByRole('cell', { name: '0' })).toBeInTheDocument()
    })

    it('E-11 las severidades van en orden fijo, no por valor (CA-07)', async () => {
      conMetricas({ defectos_por_severidad: { minor: 0, major: 5, critical: 2, blocker: 9 } })
      renderizar()
      await esperarDashboard()

      const filas = within(graficoSeveridad()).getAllByRole('row').slice(1)
      expect(filas.map((fila) => within(fila).getByRole('rowheader').textContent)).toEqual([
        'Menor',
        'Mayor',
        'Crítica',
        'Bloqueante',
      ])
    })

    it('E-12 tiene un resumen en texto con los mismos valores (CA-08)', async () => {
      conMetricas({ defectos_por_severidad: { minor: 1, major: 1, critical: 1, blocker: 0 } })
      renderizar()
      await esperarDashboard()

      expect(
        within(graficoSeveridad()).getByText('Menor: 1, Mayor: 1, Crítica: 1, Bloqueante: 0'),
      ).toBeInTheDocument()
    })

    it('E-13 sin defectos el resumen lo indica y las barras valen 0 (CA-08)', async () => {
      conMetricas({ defectos_por_severidad: { minor: 0, major: 0, critical: 0, blocker: 0 } })
      renderizar()
      await esperarDashboard()

      const grafico = graficoSeveridad()
      expect(within(grafico).getByText('Sin defectos registrados.')).toBeInTheDocument()
      expect(within(grafico).getAllByRole('cell', { name: '0' })).toHaveLength(4)
    })

    it('E-15 una clave faltante de severidad se toma como cero (CA-06)', async () => {
      metricasProyectoMock.mockResolvedValue({
        ...BASE,
        defectos_por_severidad: { minor: 2, major: 1, critical: 0 },
      } as unknown as MetricasProyecto)
      renderizar()
      await esperarDashboard()

      const fila = within(graficoSeveridad()).getByRole('row', { name: /Bloqueante/ })
      expect(within(fila).getByRole('cell', { name: '0' })).toBeInTheDocument()
    })

    it('E-20 cada severidad tiene su etiqueta y su valor en texto, no solo color (CA-12)', async () => {
      conMetricas({ defectos_por_severidad: { minor: 1, major: 2, critical: 3, blocker: 4 } })
      renderizar()
      await esperarDashboard()

      const leyenda = within(graficoSeveridad()).getByRole('list', {
        name: 'Leyenda de defectos por severidad',
      })
      const items = within(leyenda).getAllByRole('listitem')
      expect(items.map((item) => item.textContent)).toEqual([
        'Menor1',
        'Mayor2',
        'Crítica3',
        'Bloqueante4',
      ])
    })
  })

  describe('compatibilidad con un backend sin US-21', () => {
    it('E-14 sin los campos de US-21 no se muestran tarjetas ni gráfico nuevos (CA-09)', async () => {
      metricasProyectoMock.mockResolvedValue(sinCamposUs21())
      renderizar()
      await esperarDashboard()

      for (const nombre of [
        'Sprints restantes',
        'WiP',
        'Tasa de entrega',
        'Lead Time estimado',
        'Defectos abiertos',
      ]) {
        expect(screen.queryByRole('group', { name: nombre })).not.toBeInTheDocument()
      }
      expect(screen.queryByRole('region', { name: 'Defectos por severidad' })).not.toBeInTheDocument()
      expect(screen.queryByRole('alert')).not.toBeInTheDocument()
      // Solo las explicaciones "Sin datos" de F21, ninguna de F22.
      expect(screen.queryByText(/Hace falta la/)).not.toBeInTheDocument()
    })
  })

  describe('integración con la pantalla de US-F21', () => {
    it('E-16 las tarjetas nuevas aparecen debajo de las de F21 y estas no cambian (CA-10)', async () => {
      renderizar()
      await esperarDashboard()

      const velocidad = tarjeta('Velocidad')
      const sprintsRestantes = tarjeta('Sprints restantes')
      const severidad = graficoSeveridad()
      expect(
        velocidad.compareDocumentPosition(sprintsRestantes) & Node.DOCUMENT_POSITION_FOLLOWING,
      ).toBeTruthy()
      expect(
        sprintsRestantes.compareDocumentPosition(severidad) & Node.DOCUMENT_POSITION_FOLLOWING,
      ).toBeTruthy()
      expect(within(velocidad).getByText('6 SP')).toBeInTheDocument()
    })

    it('E-17 el aviso de estado actual aparece una sola vez (CA-10)', async () => {
      renderizar()
      await esperarDashboard()

      expect(
        screen.getAllByText('Calculado sobre el estado actual de las historias'),
      ).toHaveLength(1)
    })

    it('E-18 un 404 muestra el mensaje de acceso y ninguna tarjeta de F22 (CA-10)', async () => {
      metricasProyectoMock.mockRejectedValue(
        new ApiError({ status: 404, code: 'PROYECTO_NO_ENCONTRADO', message: 'Proyecto no encontrado' }),
      )
      renderizar()

      expect(
        await screen.findByText('No tenés acceso a este proyecto o no existe'),
      ).toBeInTheDocument()
      expect(screen.queryByRole('group', { name: 'WiP' })).not.toBeInTheDocument()
      expect(screen.queryByRole('region', { name: 'Defectos por severidad' })).not.toBeInTheDocument()
    })
  })

  describe('formato de números', () => {
    it('E-19 los números se muestran con formato local (CA-11)', async () => {
      conMetricas({ tasa_entrega: 2.5 })
      const { unmount } = renderizar()
      await esperarDashboard()
      expect(within(tarjeta('Tasa de entrega')).getByText('2,5')).toBeInTheDocument()
      unmount()

      conMetricas({ tasa_entrega: 2 })
      renderizar()
      await esperarDashboard()
      const kpi = tarjeta('Tasa de entrega')
      expect(within(kpi).getByText('2')).toBeInTheDocument()
      expect(within(kpi).queryByText('2,00')).not.toBeInTheDocument()
    })
  })
})

describe('MetricasPage - colores de severidad (US-F22)', () => {
  it('la leyenda usa la paleta de los otros gráficos', async () => {
    metricasProyectoMock.mockResolvedValue(BASE)
    listarHistoriasMock.mockResolvedValue([historia(1)])
    listarSprintsMock.mockResolvedValue([])
    renderizar()
    await esperarDashboard()

    const leyenda = within(graficoSeveridad()).getByRole('list', {
      name: 'Leyenda de defectos por severidad',
    })
    const cuadrados = [...leyenda.querySelectorAll<HTMLElement>('li > span[aria-hidden]')]

    expect(cuadrados.map((cuadrado) => cuadrado.style.background)).toEqual([
      'var(--chart-3)',
      'var(--chart-1)',
      'var(--chart-2)',
      'var(--accent-600)',
    ])
  })
})
