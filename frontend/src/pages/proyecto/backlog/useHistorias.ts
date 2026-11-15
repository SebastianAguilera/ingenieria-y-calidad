import { useCallback, useEffect, useState } from 'react'
import { esApiError } from '../../../api/errors'
import {
  actualizarHistoria as actualizarHistoriaApi,
  crearHistoria as crearHistoriaApi,
  listarHistorias,
} from '../../../api/historias'
import type { ActualizacionHistoria, AltaHistoria, Historia } from '../../../api/types'
import type { ResultadoOperacion } from './tipos'

export type { ResultadoOperacion } from './tipos'

export type EstadoCargaHistorias =
  | { tipo: 'cargando' }
  | { tipo: 'error'; mensaje: string }
  | { tipo: 'listo'; historias: Historia[] }

type Consulta = { clave: string; estado: EstadoCargaHistorias }

const MENSAJE_ERROR_LISTADO = 'No se pudieron cargar las historias. Intentá nuevamente.'
const MENSAJE_ERROR_ALTA = 'No se pudo crear la historia. Intentá nuevamente.'
const MENSAJE_ERROR_ACTUALIZACION = 'No se pudo actualizar la historia. Intentá nuevamente.'
const MENSAJE_PROYECTO_INEXISTENTE = 'No tenés acceso a este proyecto o no existe'

function aplicarCambios(historia: Historia, cambios: ActualizacionHistoria): Historia {
  return {
    ...historia,
    ...(cambios.titulo !== undefined ? { titulo: cambios.titulo } : {}),
    ...(cambios.descripcion !== undefined ? { descripcion: cambios.descripcion } : {}),
    ...(cambios.prioridad !== undefined ? { prioridad: cambios.prioridad } : {}),
    ...(cambios.estado !== undefined ? { estado: cambios.estado } : {}),
    ...(cambios.story_points !== undefined ? { story_points: cambios.story_points } : {}),
    ...(cambios.esfuerzo_real !== undefined ? { esfuerzo_real: cambios.esfuerzo_real } : {}),
  }
}

function traducirError(error: unknown, mensajeGeneral: string): ResultadoOperacion {
  if (!esApiError(error)) {
    return { ok: false, mensaje: mensajeGeneral }
  }

  if (error.status === 0) {
    return { ok: false, mensaje: mensajeGeneral }
  }

  if (error.status === 404) {
    return {
      ok: false,
      mensaje:
        error.code === 'PROYECTO_NO_ENCONTRADO'
          ? MENSAJE_PROYECTO_INEXISTENTE
          : 'La historia no existe',
    }
  }

  if (error.campo) {
    return { ok: false, mensaje: error.message, campo: error.campo }
  }

  return { ok: false, mensaje: error.message !== '' ? error.message : mensajeGeneral }
}

export function useHistorias(proyectoId: number | null) {
  const [consulta, setConsulta] = useState<Consulta | null>(null)
  const [intento, setIntento] = useState(0)
  const [guardandoId, setGuardandoId] = useState<number | null>(null)
  const clave = proyectoId === null ? 'sin-proyecto' : `${proyectoId}:${intento}`

  const estado: EstadoCargaHistorias =
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

    listarHistorias(proyectoId, controlador.signal)
      .then((historias) => {
        setConsulta({ clave, estado: { tipo: 'listo', historias } })
      })
      .catch((error: unknown) => {
        if (error instanceof Error && error.name === 'AbortError') {
          return
        }
        if (esApiError(error) && error.status === 404) {
          setConsulta({
            clave,
            estado: { tipo: 'error', mensaje: MENSAJE_PROYECTO_INEXISTENTE },
          })
          return
        }
        setConsulta({ clave, estado: { tipo: 'error', mensaje: MENSAJE_ERROR_LISTADO } })
      })

    return () => controlador.abort()
  }, [clave, proyectoId])

  const reintentar = useCallback(() => {
    setIntento((anterior) => anterior + 1)
  }, [])

  const agregar = useCallback(
    (historia: Historia) => {
      setConsulta((anterior) =>
        anterior?.clave === clave && anterior.estado.tipo === 'listo'
          ? {
              ...anterior,
              estado: { tipo: 'listo', historias: [...anterior.estado.historias, historia] },
            }
          : anterior,
      )
    },
    [clave],
  )

  const reemplazar = useCallback(
    (historia: Historia) => {
      setConsulta((anterior) =>
        anterior?.clave === clave && anterior.estado.tipo === 'listo'
          ? {
              ...anterior,
              estado: {
                tipo: 'listo',
                historias: anterior.estado.historias.map((actual) =>
                  actual.id === historia.id ? historia : actual,
                ),
              },
            }
          : anterior,
      )
    },
    [clave],
  )

  const crear = useCallback(
    async (alta: AltaHistoria): Promise<ResultadoOperacion> => {
      if (proyectoId === null) {
        return { ok: false, mensaje: MENSAJE_PROYECTO_INEXISTENTE }
      }

      try {
        const creada = await crearHistoriaApi(proyectoId, alta)
        agregar(creada)
        return { ok: true }
      } catch (error) {
        return traducirError(error, MENSAJE_ERROR_ALTA)
      }
    },
    [agregar, proyectoId],
  )

  const actualizar = useCallback(
    async (
      historia: Historia,
      cambios: ActualizacionHistoria,
    ): Promise<ResultadoOperacion> => {
      if (proyectoId === null) {
        return { ok: false, mensaje: MENSAJE_PROYECTO_INEXISTENTE }
      }

      setGuardandoId(historia.id)
      setConsulta((anterior) =>
        anterior?.clave === clave && anterior.estado.tipo === 'listo'
          ? {
              ...anterior,
              estado: {
                tipo: 'listo',
                historias: anterior.estado.historias.map((actual) =>
                  actual.id === historia.id ? aplicarCambios(actual, cambios) : actual,
                ),
              },
            }
          : anterior,
      )

      try {
        const actualizada = await actualizarHistoriaApi(proyectoId, historia.id, cambios)
        reemplazar(actualizada)
        return { ok: true }
      } catch (error) {
        setConsulta((anterior) =>
          anterior?.clave === clave && anterior.estado.tipo === 'listo'
            ? {
                ...anterior,
                estado: {
                  tipo: 'listo',
                  historias: anterior.estado.historias.map((actual) =>
                    actual.id === historia.id ? historia : actual,
                  ),
                },
              }
            : anterior,
        )
        return traducirError(error, MENSAJE_ERROR_ACTUALIZACION)
      } finally {
        setGuardandoId(null)
      }
    },
    [clave, proyectoId, reemplazar],
  )

  return { estado, reintentar, crear, actualizar, guardandoId }
}
