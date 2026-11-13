import { useCallback, useEffect, useState } from 'react'
import { esApiError } from '../../../api/errors'
import {
  agregarIntegrantePorEmail,
  cambiarRolIntegrante,
  listarIntegrantes,
  obtenerProyecto,
  quitarIntegrante,
} from '../../../api/proyectos'
import type { RolProyecto, UsuarioProyecto } from '../../../api/types'
import type { ResultadoOperacion } from './tipos'

export type EstadoEquipo =
  | { tipo: 'cargando' }
  | { tipo: 'error'; mensaje: string; reintentable: boolean }
  | {
      tipo: 'listo'
      integrantes: UsuarioProyecto[]
      rolActual: RolProyecto | null
      cerrado: boolean
    }

type Consulta = { clave: string; estado: EstadoEquipo }

const MENSAJE_ERROR_LISTADO = 'No se pudieron cargar los integrantes. Intentá nuevamente.'
const MENSAJE_ERROR_ALTA = 'No se pudo agregar al integrante. Intentá nuevamente.'
const MENSAJE_ERROR_ROL = 'No se pudo cambiar el rol. Intentá nuevamente.'
const MENSAJE_ERROR_BAJA = 'No se pudo quitar al integrante. Intentá nuevamente.'
const MENSAJE_SIN_CONEXION = 'No se pudo conectar con el servidor. Intentá nuevamente.'
export const MENSAJE_PROYECTO_INEXISTENTE = 'No tenés acceso a este proyecto o no existe'

function traducirError(error: unknown, mensajeGeneral: string): ResultadoOperacion {
  if (!esApiError(error) || error.status === 0) {
    return { ok: false, mensaje: MENSAJE_SIN_CONEXION }
  }

  if (error.status === 404 && error.code === 'PROYECTO_NO_ENCONTRADO') {
    return { ok: false, mensaje: MENSAJE_PROYECTO_INEXISTENTE }
  }

  if (error.status === 404 && error.code === 'USUARIO_NO_ENCONTRADO') {
    return { ok: false, mensaje: error.message, campo: 'email' }
  }

  if (error.campo) {
    return { ok: false, mensaje: error.message, campo: error.campo }
  }

  return { ok: false, mensaje: error.message !== '' ? error.message : mensajeGeneral }
}

function ordenarPorId(integrantes: UsuarioProyecto[]): UsuarioProyecto[] {
  return [...integrantes].sort((a, b) => a.id - b.id)
}

export function useEquipo(proyectoId: number | null) {
  const [consulta, setConsulta] = useState<Consulta | null>(null)
  const [intento, setIntento] = useState(0)
  const [procesandoId, setProcesandoId] = useState<number | null>(null)
  const clave = proyectoId === null ? 'sin-proyecto' : `${proyectoId}:${intento}`

  const estado: EstadoEquipo =
    proyectoId === null
      ? { tipo: 'error', mensaje: MENSAJE_PROYECTO_INEXISTENTE, reintentable: false }
      : consulta?.clave === clave
        ? consulta.estado
        : { tipo: 'cargando' }

  useEffect(() => {
    if (proyectoId === null) {
      return
    }

    const controlador = new AbortController()

    Promise.all([
      obtenerProyecto(proyectoId, controlador.signal),
      listarIntegrantes(proyectoId, controlador.signal),
    ])
      .then(([proyecto, integrantes]) => {
        setConsulta({
          clave,
          estado: {
            tipo: 'listo',
            integrantes: ordenarPorId(integrantes),
            rolActual: proyecto.rol_actual ?? null,
            cerrado: proyecto.estado === 'Cerrado',
          },
        })
      })
      .catch((error: unknown) => {
        if (error instanceof Error && error.name === 'AbortError') {
          return
        }
        if (esApiError(error) && error.status === 404) {
          setConsulta({
            clave,
            estado: {
              tipo: 'error',
              mensaje: MENSAJE_PROYECTO_INEXISTENTE,
              reintentable: false,
            },
          })
          return
        }
        setConsulta({
          clave,
          estado: { tipo: 'error', mensaje: MENSAJE_ERROR_LISTADO, reintentable: true },
        })
      })

    return () => controlador.abort()
  }, [clave, proyectoId])

  const reintentar = useCallback(() => {
    setIntento((anterior) => anterior + 1)
  }, [])

  const actualizarIntegrantes = useCallback(
    (transformar: (integrantes: UsuarioProyecto[]) => UsuarioProyecto[]) => {
      setConsulta((anterior) =>
        anterior?.clave === clave && anterior.estado.tipo === 'listo'
          ? {
              ...anterior,
              estado: {
                ...anterior.estado,
                integrantes: transformar(anterior.estado.integrantes),
              },
            }
          : anterior,
      )
    },
    [clave],
  )

  const agregar = useCallback(
    async (email: string, rol: RolProyecto): Promise<ResultadoOperacion> => {
      if (proyectoId === null) {
        return { ok: false, mensaje: MENSAJE_PROYECTO_INEXISTENTE }
      }

      try {
        const creado = await agregarIntegrantePorEmail(proyectoId, email, rol)
        actualizarIntegrantes((integrantes) =>
          ordenarPorId([...integrantes.filter((actual) => actual.id !== creado.id), creado]),
        )
        return { ok: true }
      } catch (error) {
        return traducirError(error, MENSAJE_ERROR_ALTA)
      }
    },
    [actualizarIntegrantes, proyectoId],
  )

  const cambiarRol = useCallback(
    async (usuarioId: number, rol: RolProyecto): Promise<ResultadoOperacion> => {
      if (proyectoId === null) {
        return { ok: false, mensaje: MENSAJE_PROYECTO_INEXISTENTE }
      }

      setProcesandoId(usuarioId)
      try {
        const actualizado = await cambiarRolIntegrante(proyectoId, usuarioId, rol)
        actualizarIntegrantes((integrantes) =>
          integrantes.map((actual) =>
            actual.id === usuarioId ? { ...actual, rol: actualizado.rol } : actual,
          ),
        )
        return { ok: true }
      } catch (error) {
        return traducirError(error, MENSAJE_ERROR_ROL)
      } finally {
        setProcesandoId(null)
      }
    },
    [actualizarIntegrantes, proyectoId],
  )

  const quitar = useCallback(
    async (usuarioId: number): Promise<ResultadoOperacion> => {
      if (proyectoId === null) {
        return { ok: false, mensaje: MENSAJE_PROYECTO_INEXISTENTE }
      }

      setProcesandoId(usuarioId)
      try {
        await quitarIntegrante(proyectoId, usuarioId)
        actualizarIntegrantes((integrantes) =>
          integrantes.filter((actual) => actual.id !== usuarioId),
        )
        return { ok: true }
      } catch (error) {
        return traducirError(error, MENSAJE_ERROR_BAJA)
      } finally {
        setProcesandoId(null)
      }
    },
    [actualizarIntegrantes, proyectoId],
  )

  return { estado, reintentar, agregar, cambiarRol, quitar, procesandoId }
}
