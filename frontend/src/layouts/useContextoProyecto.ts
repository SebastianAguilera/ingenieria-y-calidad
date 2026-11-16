import { useEffect, useState } from 'react'
import { listarSprints } from '../api/sprints'
import { obtenerProyecto } from '../api/proyectos'
import type { Proyecto, Sprint } from '../api/types'

export type ContextoProyecto = {
  proyecto: Proyecto | null
  sprintActivo: Sprint | null
}

type Cargado = ContextoProyecto & { proyectoId: number }

/**
 * Datos del proyecto abierto que necesita la barra superior (nombre, rol del
 * usuario y sprint activo). Reutiliza los endpoints existentes y se recarga
 * cuando cambia la ruta, para reflejar un sprint recién activado o cerrado.
 */
export function useContextoProyecto(proyectoId: number | null, ruta: string): ContextoProyecto {
  const [cargado, setCargado] = useState<Cargado | null>(null)

  useEffect(() => {
    if (proyectoId === null) {
      return
    }
    const control = new AbortController()
    void Promise.allSettled([
      obtenerProyecto(proyectoId, control.signal),
      listarSprints(proyectoId, control.signal),
    ]).then(([proyecto, sprints]) => {
      if (control.signal.aborted) {
        return
      }
      setCargado({
        proyectoId,
        proyecto: proyecto.status === 'fulfilled' ? proyecto.value : null,
        sprintActivo:
          sprints.status === 'fulfilled'
            ? (sprints.value.find((sprint) => sprint.estado === 'Activo') ?? null)
            : null,
      })
    })
    return () => control.abort()
  }, [proyectoId, ruta])

  if (cargado === null || cargado.proyectoId !== proyectoId) {
    return { proyecto: null, sprintActivo: null }
  }
  return { proyecto: cargado.proyecto, sprintActivo: cargado.sprintActivo }
}
