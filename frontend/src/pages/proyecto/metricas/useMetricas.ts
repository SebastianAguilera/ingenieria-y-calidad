import { useCallback, useEffect, useState } from 'react'
import { esApiError } from '../../../api/errors'
import { listarHistorias } from '../../../api/historias'
import { obtenerMetricasProyecto, obtenerMetricasSprint } from '../../../api/metricas'
import { listarSprints } from '../../../api/sprints'
import type { Historia, MetricasProyecto, MetricasSprint, Sprint } from '../../../api/types'
import { CODIGO_PROYECTO_NO_ENCONTRADO, MENSAJE_PROYECTO_INEXISTENTE } from '../mensajes'
import { sprintsDelGrafico } from './distribucion'

export type SprintConMetricas = { sprint: Sprint; metricas: MetricasSprint }

export type DatosMetricas = {
  metricas: MetricasProyecto
  historias: Historia[]
  sprints: Sprint[]
  /** Sprints cerrados y el activo, con sus métricas, para el gráfico de barras. */
  sprintsGrafico: SprintConMetricas[]
}

export type EstadoCargaMetricas =
  | { tipo: 'cargando' }
  | { tipo: 'error'; mensaje: string }
  | { tipo: 'listo'; datos: DatosMetricas }

type Consulta = { clave: string; estado: EstadoCargaMetricas }

export const MENSAJE_ERROR_METRICAS = 'No se pudieron cargar las métricas. Intentá nuevamente.'

function esAborto(error: unknown): boolean {
  return error instanceof Error && error.name === 'AbortError'
}

export function mensajeDeError(error: unknown, mensajeGeneral: string): string {
  return esApiError(error) && error.code === CODIGO_PROYECTO_NO_ENCONTRADO
    ? MENSAJE_PROYECTO_INEXISTENTE
    : mensajeGeneral
}

async function cargarDatos(proyectoId: number, signal: AbortSignal): Promise<DatosMetricas> {
  const [metricas, historias, sprints] = await Promise.all([
    obtenerMetricasProyecto(proyectoId, signal),
    listarHistorias(proyectoId, signal),
    listarSprints(proyectoId, signal),
  ])

  const sprintsGrafico = await Promise.all(
    sprintsDelGrafico(sprints).map(async (sprint) => ({
      sprint,
      metricas: await obtenerMetricasSprint(proyectoId, sprint.id, signal),
    })),
  )

  return { metricas, historias, sprints, sprintsGrafico }
}

/** Carga los datos del dashboard de métricas (US-F21, tramo 1). */
export function useMetricas(proyectoId: number | null) {
  const [consulta, setConsulta] = useState<Consulta | null>(null)
  const [intento, setIntento] = useState(0)
  const clave = proyectoId === null ? 'sin-proyecto' : `${proyectoId}:${intento}`

  const estado: EstadoCargaMetricas =
    proyectoId === null
      ? { tipo: 'error', mensaje: MENSAJE_PROYECTO_INEXISTENTE }
      : consulta?.clave === clave
        ? consulta.estado
        : { tipo: 'cargando' }

  useEffect(() => {
    if (proyectoId === null) {
      return
    }

    const controlador = new AbortController()

    cargarDatos(proyectoId, controlador.signal)
      .then((datos) => {
        setConsulta({ clave, estado: { tipo: 'listo', datos } })
      })
      .catch((error: unknown) => {
        if (esAborto(error)) {
          return
        }
        setConsulta({
          clave,
          estado: { tipo: 'error', mensaje: mensajeDeError(error, MENSAJE_ERROR_METRICAS) },
        })
      })

    return () => controlador.abort()
  }, [clave, proyectoId])

  const reintentar = useCallback(() => {
    setIntento((anterior) => anterior + 1)
  }, [])

  return { estado, reintentar }
}
