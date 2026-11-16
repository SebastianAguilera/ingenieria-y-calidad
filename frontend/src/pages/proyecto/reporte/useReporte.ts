import { useCallback, useEffect, useState } from 'react'
import { obtenerReporteProyecto, obtenerReporteSprint } from '../../../api/reportes'
import { listarSprints } from '../../../api/sprints'
import type { ReporteProyecto, ReporteSprint, Sprint } from '../../../api/types'
import { MENSAJE_PROYECTO_INEXISTENTE } from '../mensajes'
import { mensajeDeError } from '../metricas/useMetricas'

export const MENSAJE_ERROR_REPORTE = 'No se pudo cargar el reporte. Intentá nuevamente.'

/** Reporte cargado, con su alcance: proyecto completo o un sprint. */
export type Reporte =
  | { alcance: 'proyecto'; datos: ReporteProyecto }
  | { alcance: 'sprint'; datos: ReporteSprint }

export type EstadoCarga<T> =
  | { tipo: 'cargando' }
  | { tipo: 'error'; mensaje: string }
  | { tipo: 'listo'; valor: T }

type Consulta<T> = { clave: string; estado: EstadoCarga<T> }

function esAborto(error: unknown): boolean {
  return error instanceof Error && error.name === 'AbortError'
}

/**
 * Ejecuta `cargar` cada vez que cambia `clave` (o se reintenta) y expone el
 * estado de esa consulta. Cancela la consulta anterior al cambiar de clave.
 */
function useConsulta<T>(
  clave: string | null,
  cargar: (signal: AbortSignal) => Promise<T>,
): { estado: EstadoCarga<T>; reintentar: () => void } {
  const [consulta, setConsulta] = useState<Consulta<T> | null>(null)
  const [intento, setIntento] = useState(0)
  const claveIntento = clave === null ? null : `${clave}:${intento}`

  useEffect(() => {
    if (claveIntento === null) {
      return
    }
    const controlador = new AbortController()
    cargar(controlador.signal)
      .then((valor) => setConsulta({ clave: claveIntento, estado: { tipo: 'listo', valor } }))
      .catch((error: unknown) => {
        if (esAborto(error)) {
          return
        }
        setConsulta({
          clave: claveIntento,
          estado: { tipo: 'error', mensaje: mensajeDeError(error, MENSAJE_ERROR_REPORTE) },
        })
      })
    return () => controlador.abort()
    // `cargar` depende solo de lo que ya está en la clave.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [claveIntento])

  const reintentar = useCallback(() => setIntento((anterior) => anterior + 1), [])

  const estado: EstadoCarga<T> =
    claveIntento === null
      ? { tipo: 'error', mensaje: MENSAJE_PROYECTO_INEXISTENTE }
      : consulta?.clave === claveIntento
        ? consulta.estado
        : { tipo: 'cargando' }

  return { estado, reintentar }
}

/** Sprints del proyecto, para el selector de alcance (US-F23). */
export function useSprintsDelReporte(proyectoId: number | null) {
  return useConsulta<Sprint[]>(proyectoId === null ? null : `sprints:${proyectoId}`, (signal) =>
    listarSprints(proyectoId as number, signal),
  )
}

/** Reporte del proyecto (`sprintId` `null`) o de un sprint (US-F23, RN-03). */
export function useReporte(proyectoId: number | null, sprintId: number | null) {
  return useConsulta<Reporte>(
    proyectoId === null ? null : `reporte:${proyectoId}:${sprintId ?? 'proyecto'}`,
    async (signal) => {
      const id = proyectoId as number
      if (sprintId === null) {
        return { alcance: 'proyecto', datos: await obtenerReporteProyecto(id, signal) }
      }
      return { alcance: 'sprint', datos: await obtenerReporteSprint(id, sprintId, signal) }
    },
  )
}
