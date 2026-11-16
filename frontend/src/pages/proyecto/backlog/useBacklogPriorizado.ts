import { useCallback, useEffect, useState } from 'react'
import { esApiError } from '../../../api/errors'
import { listarBacklogPriorizado } from '../../../api/historias'
import type { HistoriaPriorizada } from '../../../api/types'

export type EstadoBacklogPriorizado =
  | { tipo: 'cargando' }
  | { tipo: 'error'; mensaje: string }
  | { tipo: 'listo'; historias: HistoriaPriorizada[] }

type Consulta = { clave: string; estado: EstadoBacklogPriorizado }

export const MENSAJE_ERROR_PRIORIZADO =
  'No se pudo cargar el backlog priorizado. Intentá nuevamente.'
export const MENSAJE_PROYECTO_INEXISTENTE = 'No tenés acceso a este proyecto o no existe'

export function useBacklogPriorizado(proyectoId: number) {
  const [consulta, setConsulta] = useState<Consulta | null>(null)
  const [intento, setIntento] = useState(0)
  const clave = `${proyectoId}:${intento}`

  const estado: EstadoBacklogPriorizado =
    consulta?.clave === clave ? consulta.estado : { tipo: 'cargando' }

  useEffect(() => {
    const controlador = new AbortController()

    listarBacklogPriorizado(proyectoId, controlador.signal)
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
        setConsulta({ clave, estado: { tipo: 'error', mensaje: MENSAJE_ERROR_PRIORIZADO } })
      })

    return () => controlador.abort()
  }, [clave, proyectoId])

  const reintentar = useCallback(() => {
    setIntento((anterior) => anterior + 1)
  }, [])

  return { estado, reintentar }
}
