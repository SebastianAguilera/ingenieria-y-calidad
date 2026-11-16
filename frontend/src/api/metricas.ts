import { api } from './client'
import type { MetricasProyecto, MetricasSprint } from './types'

/** Métricas del proyecto completo (US-20 / US-21). */
export function obtenerMetricasProyecto(
  proyectoId: number,
  signal?: AbortSignal,
): Promise<MetricasProyecto> {
  return api.get<MetricasProyecto>(`/proyectos/${proyectoId}/metricas`, { signal })
}

/** Métricas de un sprint del proyecto (US-20). */
export function obtenerMetricasSprint(
  proyectoId: number,
  sprintId: number,
  signal?: AbortSignal,
): Promise<MetricasSprint> {
  return api.get<MetricasSprint>(`/proyectos/${proyectoId}/sprints/${sprintId}/metricas`, {
    signal,
  })
}
