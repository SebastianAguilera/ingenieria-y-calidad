import { api, descargarArchivo } from './client'
import type { ReporteProyecto, ReporteSprint } from './types'

/** Reporte consolidado del proyecto completo (US-22). */
export function obtenerReporteProyecto(
  proyectoId: number,
  signal?: AbortSignal,
): Promise<ReporteProyecto> {
  return api.get<ReporteProyecto>(`/proyectos/${proyectoId}/reporte`, { signal })
}

/** Reporte consolidado de un sprint del proyecto (US-22). */
export function obtenerReporteSprint(
  proyectoId: number,
  sprintId: number,
  signal?: AbortSignal,
): Promise<ReporteSprint> {
  return api.get<ReporteSprint>(`/proyectos/${proyectoId}/sprints/${sprintId}/reporte`, {
    signal,
  })
}

/** PDF del reporte (US-24): del proyecto si `sprintId` es `null`, si no del sprint. */
export function descargarPdfReporte(
  proyectoId: number,
  sprintId: number | null,
  signal?: AbortSignal,
): Promise<Blob> {
  const ruta =
    sprintId === null
      ? `/proyectos/${proyectoId}/reporte.pdf`
      : `/proyectos/${proyectoId}/sprints/${sprintId}/reporte.pdf`
  return descargarArchivo(ruta, 'application/pdf', signal)
}
