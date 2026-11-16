import { beforeEach, describe, expect, it, vi } from 'vitest'
import { api } from './client'
import { obtenerMetricasProyecto, obtenerMetricasSprint } from './metricas'
import type { MetricasProyecto, MetricasSprint } from './types'

vi.mock('./client', () => ({
  api: {
    get: vi.fn(),
  },
}))

const getMock = vi.mocked(api.get)

const metricasProyecto: MetricasProyecto = {
  proyecto_id: 1,
  base: 'estado_actual',
  velocidad: 6,
  porcentaje_historias_completadas: 0.5,
  horas_estimadas: 40,
  horas_reales: 48,
  desviacion_esfuerzo: 0.2,
  defectos_detectados: 3,
  defectos_resueltos: 2,
  ratio_defectos: 0.3333333333333333,
  sprints_restantes: null,
  wip: 0,
  tasa_entrega: null,
  lead_time_estimado: null,
  lead_time_aproximado: false,
  defectos_abiertos: 1,
  defectos_por_severidad: { minor: 1, major: 1, critical: 1, blocker: 0 },
}

const metricasSprint: MetricasSprint = {
  sprint_id: 10,
  base: 'estado_actual',
  story_points_planificados: 13,
  story_points_completados: 8,
  porcentaje_completado: 0.6153846153846154,
  historias_completadas: 2,
  horas_estimadas: 20,
  horas_reales: 24,
  desviacion_esfuerzo: 0.2,
  defectos_detectados: 2,
  defectos_resueltos: 1,
}

describe('obtenerMetricasProyecto', () => {
  beforeEach(() => {
    getMock.mockReset()
  })

  it('consulta GET /proyectos/:id/metricas y devuelve el cuerpo', async () => {
    getMock.mockResolvedValue(metricasProyecto)

    const metricas = await obtenerMetricasProyecto(1)

    expect(getMock).toHaveBeenCalledWith('/proyectos/1/metricas', { signal: undefined })
    expect(metricas).toEqual(metricasProyecto)
  })

  it('pasa la señal de aborto al cliente', async () => {
    getMock.mockResolvedValue(metricasProyecto)
    const controlador = new AbortController()

    await obtenerMetricasProyecto(1, controlador.signal)

    expect(getMock).toHaveBeenCalledWith('/proyectos/1/metricas', {
      signal: controlador.signal,
    })
  })

  it('propaga el error del cliente', async () => {
    const error = new Error('fallo')
    getMock.mockRejectedValue(error)

    await expect(obtenerMetricasProyecto(1)).rejects.toBe(error)
  })
})

describe('obtenerMetricasSprint', () => {
  beforeEach(() => {
    getMock.mockReset()
  })

  it('consulta GET /proyectos/:id/sprints/:sprintId/metricas', async () => {
    getMock.mockResolvedValue(metricasSprint)
    const controlador = new AbortController()

    const metricas = await obtenerMetricasSprint(1, 10, controlador.signal)

    expect(getMock).toHaveBeenCalledWith('/proyectos/1/sprints/10/metricas', {
      signal: controlador.signal,
    })
    expect(metricas).toEqual(metricasSprint)
  })
})
