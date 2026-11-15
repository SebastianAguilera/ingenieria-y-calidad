import { useCallback, useEffect, useState } from 'react'
import {
  actualizarDefecto as actualizarDefectoApi,
  crearDefecto as crearDefectoApi,
  listarDefectos,
} from '../../../api/defectos'
import { esApiError } from '../../../api/errors'
import { listarHistorias } from '../../../api/historias'
import { listarSprints } from '../../../api/sprints'
import type {
  ActualizacionDefecto,
  AltaDefecto,
  Defecto,
  Historia,
  Sprint,
} from '../../../api/types'
import { CODIGO_PROYECTO_NO_ENCONTRADO, MENSAJE_PROYECTO_INEXISTENTE } from '../mensajes'
import type { ResultadoOperacion } from '../sprints/tipos'

export type EstadoCargaDefectos =
  | { tipo: 'cargando' }
  | { tipo: 'error'; mensaje: string }
  | {
      tipo: 'listo'
      defectos: Defecto[]
      historias: Historia[]
      sprints: Sprint[]
      cargandoDefectos: boolean
    }

type Referencias = { historias: Historia[]; sprints: Sprint[] }

type Resultado<T> = { ok: true; datos: T } | { ok: false; mensaje: string }

type ConsultaReferencias = { clave: string; resultado: Resultado<Referencias> }
type ConsultaDefectos = { clave: string; resultado: Resultado<Defecto[]> }

const MENSAJE_ERROR_LISTADO = 'No se pudieron cargar los defectos. Intentá nuevamente.'
const MENSAJE_ERROR_ALTA = 'No se pudo crear el defecto. Intentá nuevamente.'
const MENSAJE_ERROR_ESTADO = 'No se pudo cambiar el estado del defecto. Intentá nuevamente.'

function esAborto(error: unknown): boolean {
  return error instanceof Error && error.name === 'AbortError'
}

function mensajeDeCarga(error: unknown): string {
  return esApiError(error) && error.code === CODIGO_PROYECTO_NO_ENCONTRADO
    ? MENSAJE_PROYECTO_INEXISTENTE
    : MENSAJE_ERROR_LISTADO
}

function traducirError(error: unknown, mensajeGeneral: string): ResultadoOperacion {
  if (!esApiError(error) || error.status === 0) {
    return { ok: false, mensaje: mensajeGeneral }
  }

  if (error.status === 404 && error.code === CODIGO_PROYECTO_NO_ENCONTRADO) {
    return { ok: false, mensaje: MENSAJE_PROYECTO_INEXISTENTE }
  }

  if (error.campo) {
    return { ok: false, mensaje: error.message, campo: error.campo }
  }

  return { ok: false, mensaje: error.message !== '' ? error.message : mensajeGeneral }
}

export function useDefectos(proyectoId: number | null) {
  const [consultaReferencias, setConsultaReferencias] = useState<ConsultaReferencias | null>(null)
  const [consultaDefectos, setConsultaDefectos] = useState<ConsultaDefectos | null>(null)
  const [intento, setIntento] = useState(0)
  const [procesandoId, setProcesandoId] = useState<number | null>(null)

  const claveReferencias = proyectoId === null ? 'sin-proyecto' : `${proyectoId}:${intento}`
  const claveDefectos = claveReferencias

  const referencias =
    consultaReferencias?.clave === claveReferencias ? consultaReferencias.resultado : null
  const defectos = consultaDefectos?.clave === claveDefectos ? consultaDefectos.resultado : null

  let estado: EstadoCargaDefectos
  if (proyectoId === null) {
    estado = { tipo: 'error', mensaje: MENSAJE_PROYECTO_INEXISTENTE }
  } else if (referencias !== null && !referencias.ok) {
    estado = { tipo: 'error', mensaje: referencias.mensaje }
  } else if (defectos !== null && !defectos.ok) {
    estado = { tipo: 'error', mensaje: defectos.mensaje }
  } else if (referencias !== null && referencias.ok) {
    estado = {
      tipo: 'listo',
      defectos: defectos !== null && defectos.ok ? defectos.datos : [],
      historias: referencias.datos.historias,
      sprints: referencias.datos.sprints,
      cargandoDefectos: defectos === null,
    }
  } else {
    estado = { tipo: 'cargando' }
  }

  useEffect(() => {
    if (proyectoId === null) {
      return
    }

    const controlador = new AbortController()

    Promise.all([
      listarHistorias(proyectoId, controlador.signal),
      listarSprints(proyectoId, controlador.signal),
    ])
      .then(([historias, sprints]) => {
        setConsultaReferencias({
          clave: claveReferencias,
          resultado: { ok: true, datos: { historias, sprints } },
        })
      })
      .catch((error: unknown) => {
        if (esAborto(error)) {
          return
        }
        setConsultaReferencias({
          clave: claveReferencias,
          resultado: { ok: false, mensaje: mensajeDeCarga(error) },
        })
      })

    return () => controlador.abort()
  }, [claveReferencias, proyectoId])

  useEffect(() => {
    if (proyectoId === null) {
      return
    }

    const controlador = new AbortController()

    listarDefectos(proyectoId, controlador.signal)
      .then((lista) => {
        setConsultaDefectos({ clave: claveDefectos, resultado: { ok: true, datos: lista } })
      })
      .catch((error: unknown) => {
        if (esAborto(error)) {
          return
        }
        setConsultaDefectos({
          clave: claveDefectos,
          resultado: { ok: false, mensaje: mensajeDeCarga(error) },
        })
      })

    return () => controlador.abort()
  }, [claveDefectos, proyectoId])

  const reintentar = useCallback(() => {
    setIntento((anterior) => anterior + 1)
  }, [])

  const modificarLista = useCallback(
    (transformar: (lista: Defecto[]) => Defecto[]) => {
      setConsultaDefectos((anterior) =>
        anterior?.clave === claveDefectos && anterior.resultado.ok
          ? {
              ...anterior,
              resultado: { ok: true, datos: transformar(anterior.resultado.datos) },
            }
          : anterior,
      )
    },
    [claveDefectos],
  )

  const crear = useCallback(
    async (alta: AltaDefecto): Promise<ResultadoOperacion> => {
      if (proyectoId === null) {
        return { ok: false, mensaje: MENSAJE_PROYECTO_INEXISTENTE }
      }

      try {
        const creado = await crearDefectoApi(proyectoId, alta)
        modificarLista((lista) => [...lista, creado])
        return { ok: true }
      } catch (error) {
        return traducirError(error, MENSAJE_ERROR_ALTA)
      }
    },
    [modificarLista, proyectoId],
  )

  const cambiarEstado = useCallback(
    async (
      defectoId: number,
      actualizacion: ActualizacionDefecto,
    ): Promise<ResultadoOperacion> => {
      if (proyectoId === null) {
        return { ok: false, mensaje: MENSAJE_PROYECTO_INEXISTENTE }
      }

      setProcesandoId(defectoId)
      try {
        const actualizado = await actualizarDefectoApi(proyectoId, defectoId, actualizacion)
        modificarLista((lista) =>
          lista.map((actual) => (actual.id === actualizado.id ? actualizado : actual)),
        )
        return { ok: true }
      } catch (error) {
        return traducirError(error, MENSAJE_ERROR_ESTADO)
      } finally {
        setProcesandoId(null)
      }
    },
    [modificarLista, proyectoId],
  )

  return { estado, reintentar, crear, cambiarEstado, procesandoId }
}
