import { useCallback, useEffect, useMemo, useState } from 'react'
import type { ReactNode } from 'react'
import { useParams } from 'react-router-dom'
import { esApiError } from '../../api/errors'
import { obtenerProyecto } from '../../api/proyectos'
import type { EstadoProyectoDetalle } from './proyectoContext'
import { ProyectoContext } from './proyectoContext'

export function ProyectoProvider({ children }: { children: ReactNode }) {
  const { id } = useParams()
  const idParseado = id === undefined ? NaN : Number(id)
  const esIdValido = Number.isInteger(idParseado) && idParseado > 0

  const [estado, setEstado] = useState<EstadoProyectoDetalle>(() =>
    esIdValido ? { tipo: 'cargando' } : { tipo: 'noEncontrado' },
  )
  const [intento, setIntento] = useState(0)

  useEffect(() => {
    if (!esIdValido) {
      return
    }
    const controlador = new AbortController()
    obtenerProyecto(idParseado, controlador.signal)
      .then((proyecto) => setEstado({ tipo: 'listo', proyecto }))
      .catch((error: unknown) => {
        if (error instanceof Error && error.name === 'AbortError') {
          return
        }
        if (esApiError(error) && error.status === 404) {
          setEstado({ tipo: 'noEncontrado' })
          return
        }
        setEstado({ tipo: 'error' })
      })
    return () => controlador.abort()
  }, [esIdValido, idParseado, intento])

  const reintentar = useCallback(() => {
    setEstado({ tipo: 'cargando' })
    setIntento((anterior) => anterior + 1)
  }, [])

  const valor = useMemo(() => ({ estado, reintentar }), [estado, reintentar])

  return <ProyectoContext.Provider value={valor}>{children}</ProyectoContext.Provider>
}
