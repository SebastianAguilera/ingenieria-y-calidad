import { useCallback, useEffect, useState } from 'react'
import { esApiError } from '../../../api/errors'
import {
  asignarHistoria as asignarHistoriaApi,
  obtenerSprintBacklog,
  quitarHistoria as quitarHistoriaApi,
} from '../../../api/sprints'
import type { SprintBacklog } from '../../../api/types'
import type { ResultadoOperacion } from './tipos'

export type EstadoCargaSprintBacklog =
  | { tipo: 'cargando' }
  | { tipo: 'error'; mensaje: string }
  | { tipo: 'noEncontrado'; mensaje: string }
  | { tipo: 'listo'; backlog: SprintBacklog }

type Consulta = { clave: string; estado: EstadoCargaSprintBacklog }

const MENSAJE_ERROR_CARGA = 'No se pudo cargar el sprint. Intentá nuevamente.'
const MENSAJE_ERROR_ASIGNAR = 'No se pudo asignar la historia. Intentá nuevamente.'
const MENSAJE_ERROR_QUITAR = 'No se pudo quitar la historia. Intentá nuevamente.'
const MENSAJE_ERROR_REFRESCO =
  'La operación se realizó, pero no se pudo actualizar la lista. Recargá la página.'
const MENSAJE_SPRINT_INEXISTENTE = 'Sprint no encontrado'
const MENSAJE_PROYECTO_INEXISTENTE = 'No tenés acceso a este proyecto o no existe'

function traducirError(error: unknown, mensajeGeneral: string): ResultadoOperacion {
  if (esApiError(error) && error.status !== 0 && error.message !== '') {
    return { ok: false, mensaje: error.message }
  }
  return { ok: false, mensaje: mensajeGeneral }
}

/**
 * Datos del detalle de un sprint: el sprint con las historias asignadas y las
 * operaciones de asignar y quitar. Tras cada operación vuelve a consultar el
 * backend (el POST y el DELETE responden 204 sin cuerpo).
 */
export function useSprintBacklog(proyectoId: number | null, sprintId: number | null) {
  const [consulta, setConsulta] = useState<Consulta | null>(null)
  const [intento, setIntento] = useState(0)
  const [procesando, setProcesando] = useState(false)
  const ids = proyectoId !== null && sprintId !== null ? { proyectoId, sprintId } : null
  const clave = ids === null ? 'sin-sprint' : `${ids.proyectoId}:${ids.sprintId}:${intento}`

  const estado: EstadoCargaSprintBacklog =
    ids === null
      ? { tipo: 'noEncontrado', mensaje: MENSAJE_SPRINT_INEXISTENTE }
      : consulta?.clave === clave
        ? consulta.estado
        : { tipo: 'cargando' }

  const proyectoActual = ids?.proyectoId
  const sprintActual = ids?.sprintId

  useEffect(() => {
    if (proyectoActual === undefined || sprintActual === undefined) {
      return
    }

    const controlador = new AbortController()

    obtenerSprintBacklog(proyectoActual, sprintActual, controlador.signal)
      .then((backlog) => {
        setConsulta({ clave, estado: { tipo: 'listo', backlog } })
      })
      .catch((error: unknown) => {
        if (error instanceof Error && error.name === 'AbortError') {
          return
        }
        if (esApiError(error) && error.status === 404) {
          const mensaje =
            error.code === 'PROYECTO_NO_ENCONTRADO'
              ? MENSAJE_PROYECTO_INEXISTENTE
              : MENSAJE_SPRINT_INEXISTENTE
          setConsulta({ clave, estado: { tipo: 'noEncontrado', mensaje } })
          return
        }
        setConsulta({ clave, estado: { tipo: 'error', mensaje: MENSAJE_ERROR_CARGA } })
      })

    return () => controlador.abort()
  }, [clave, proyectoActual, sprintActual])

  const reintentar = useCallback(() => {
    setIntento((anterior) => anterior + 1)
  }, [])

  // Vuelve a consultar sin pasar por "cargando", para que la tabla no parpadee.
  const refrescar = useCallback(async (): Promise<boolean> => {
    if (proyectoActual === undefined || sprintActual === undefined) {
      return false
    }
    try {
      const backlog = await obtenerSprintBacklog(proyectoActual, sprintActual)
      setConsulta({ clave, estado: { tipo: 'listo', backlog } })
      return true
    } catch {
      return false
    }
  }, [clave, proyectoActual, sprintActual])

  const ejecutar = useCallback(
    async (
      operacion: () => Promise<void>,
      mensajeGeneral: string,
    ): Promise<ResultadoOperacion> => {
      setProcesando(true)
      try {
        await operacion()
      } catch (error) {
        return traducirError(error, mensajeGeneral)
      } finally {
        setProcesando(false)
      }

      const actualizado = await refrescar()
      return actualizado ? { ok: true } : { ok: false, mensaje: MENSAJE_ERROR_REFRESCO }
    },
    [refrescar],
  )

  const asignar = useCallback(
    (historiaId: number): Promise<ResultadoOperacion> => {
      if (proyectoActual === undefined || sprintActual === undefined) {
        return Promise.resolve({ ok: false, mensaje: MENSAJE_SPRINT_INEXISTENTE })
      }
      return ejecutar(
        () => asignarHistoriaApi(proyectoActual, sprintActual, historiaId),
        MENSAJE_ERROR_ASIGNAR,
      )
    },
    [ejecutar, proyectoActual, sprintActual],
  )

  const quitar = useCallback(
    (historiaId: number): Promise<ResultadoOperacion> => {
      if (proyectoActual === undefined || sprintActual === undefined) {
        return Promise.resolve({ ok: false, mensaje: MENSAJE_SPRINT_INEXISTENTE })
      }
      return ejecutar(
        () => quitarHistoriaApi(proyectoActual, sprintActual, historiaId),
        MENSAJE_ERROR_QUITAR,
      )
    },
    [ejecutar, proyectoActual, sprintActual],
  )

  return { estado, reintentar, asignar, quitar, procesando }
}
