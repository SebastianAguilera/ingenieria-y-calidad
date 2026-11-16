import { useCallback, useEffect, useState } from 'react'
import { obtenerMetricasSprint } from '../../../api/metricas'
import type { MetricasSprint } from '../../../api/types'
import { mensajeDeError } from './useMetricas'

export type EstadoMetricasSprint =
  | { tipo: 'cargando' }
  | { tipo: 'error'; mensaje: string }
  | { tipo: 'listo'; metricas: MetricasSprint }

type Consulta = { clave: string; estado: EstadoMetricasSprint }

export const MENSAJE_ERROR_METRICAS_SPRINT = 'No se pudieron cargar las métricas del sprint.'

/** Métricas del sprint elegido en el panel (`GET .../sprints/{sprintId}/metricas`). */
export function useMetricasSprint(proyectoId: number, sprintId: number | null) {
  const [consulta, setConsulta] = useState<Consulta | null>(null)
  const [intento, setIntento] = useState(0)
  const clave = `${proyectoId}:${sprintId ?? 'ninguno'}:${intento}`

  const estado: EstadoMetricasSprint =
    consulta?.clave === clave ? consulta.estado : { tipo: 'cargando' }

  useEffect(() => {
    if (sprintId === null) {
      return
    }

    const controlador = new AbortController()

    obtenerMetricasSprint(proyectoId, sprintId, controlador.signal)
      .then((metricas) => {
        setConsulta({ clave, estado: { tipo: 'listo', metricas } })
      })
      .catch((error: unknown) => {
        if (error instanceof Error && error.name === 'AbortError') {
          return
        }
        setConsulta({
          clave,
          estado: { tipo: 'error', mensaje: mensajeDeError(error, MENSAJE_ERROR_METRICAS_SPRINT) },
        })
      })

    return () => controlador.abort()
  }, [clave, proyectoId, sprintId])

  const reintentar = useCallback(() => {
    setIntento((anterior) => anterior + 1)
  }, [])

  return { estado, reintentar }
}
