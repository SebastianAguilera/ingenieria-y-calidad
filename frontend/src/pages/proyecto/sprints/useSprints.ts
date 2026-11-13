import { useCallback, useEffect, useState } from 'react'
import { esApiError } from '../../../api/errors'
import {
  cambiarEstadoSprint as cambiarEstadoSprintApi,
  crearSprint as crearSprintApi,
  listarSprints,
} from '../../../api/sprints'
import type { AltaSprint, EstadoSprint, Sprint } from '../../../api/types'
import type { ResultadoOperacion } from './tipos'

export type EstadoCargaSprints =
  | { tipo: 'cargando' }
  | { tipo: 'error'; mensaje: string }
  | { tipo: 'listo'; sprints: Sprint[] }

type Consulta = { clave: string; estado: EstadoCargaSprints }

const MENSAJE_ERROR_LISTADO = 'No se pudieron cargar los sprints. Intentá nuevamente.'
const MENSAJE_ERROR_ALTA = 'No se pudo crear el sprint. Intentá nuevamente.'
const MENSAJE_ERROR_ESTADO = 'No se pudo cambiar el estado del sprint. Intentá nuevamente.'
const MENSAJE_PROYECTO_INEXISTENTE = 'No tenés acceso a este proyecto o no existe'

function traducirError(error: unknown, mensajeGeneral: string): ResultadoOperacion {
  if (!esApiError(error) || error.status === 0) {
    return { ok: false, mensaje: mensajeGeneral }
  }

  if (error.status === 404 && error.code === 'PROYECTO_NO_ENCONTRADO') {
    return { ok: false, mensaje: MENSAJE_PROYECTO_INEXISTENTE }
  }

  if (error.campo) {
    return { ok: false, mensaje: error.message, campo: error.campo }
  }

  return { ok: false, mensaje: error.message !== '' ? error.message : mensajeGeneral }
}

export function useSprints(proyectoId: number | null) {
  const [consulta, setConsulta] = useState<Consulta | null>(null)
  const [intento, setIntento] = useState(0)
  const [procesandoId, setProcesandoId] = useState<number | null>(null)
  const clave = proyectoId === null ? 'sin-proyecto' : `${proyectoId}:${intento}`

  const estado: EstadoCargaSprints =
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

    listarSprints(proyectoId, controlador.signal)
      .then((sprints) => {
        setConsulta({ clave, estado: { tipo: 'listo', sprints } })
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
    (sprint: Sprint) => {
      setConsulta((anterior) =>
        anterior?.clave === clave && anterior.estado.tipo === 'listo'
          ? {
              ...anterior,
              estado: { tipo: 'listo', sprints: [...anterior.estado.sprints, sprint] },
            }
          : anterior,
      )
    },
    [clave],
  )

  const reemplazar = useCallback(
    (sprint: Sprint) => {
      setConsulta((anterior) =>
        anterior?.clave === clave && anterior.estado.tipo === 'listo'
          ? {
              ...anterior,
              estado: {
                tipo: 'listo',
                sprints: anterior.estado.sprints.map((actual) =>
                  actual.id === sprint.id ? sprint : actual,
                ),
              },
            }
          : anterior,
      )
    },
    [clave],
  )

  const crear = useCallback(
    async (alta: AltaSprint): Promise<ResultadoOperacion> => {
      if (proyectoId === null) {
        return { ok: false, mensaje: MENSAJE_PROYECTO_INEXISTENTE }
      }

      try {
        const creado = await crearSprintApi(proyectoId, alta)
        agregar(creado)
        return { ok: true }
      } catch (error) {
        return traducirError(error, MENSAJE_ERROR_ALTA)
      }
    },
    [agregar, proyectoId],
  )

  const cambiarEstado = useCallback(
    async (sprintId: number, estadoDestino: EstadoSprint): Promise<ResultadoOperacion> => {
      if (proyectoId === null) {
        return { ok: false, mensaje: MENSAJE_PROYECTO_INEXISTENTE }
      }

      setProcesandoId(sprintId)
      try {
        const actualizado = await cambiarEstadoSprintApi(proyectoId, sprintId, estadoDestino)
        reemplazar(actualizado)
        return { ok: true }
      } catch (error) {
        return traducirError(error, MENSAJE_ERROR_ESTADO)
      } finally {
        setProcesandoId(null)
      }
    },
    [proyectoId, reemplazar],
  )

  return { estado, reintentar, crear, cambiarEstado, procesandoId }
}
