import { api } from './client'
import type { ActualizacionDefecto, AltaDefecto, Defecto, TipoListadoDefecto } from './types'

type RespuestaListadoDefectos = {
  total: number
  defectos?: Defecto[] | null
}

export async function listarDefectos(
  proyectoId: number,
  signal?: AbortSignal,
): Promise<Defecto[]> {
  const respuesta = await api.get<RespuestaListadoDefectos>(
    `/proyectos/${proyectoId}/defectos`,
    { signal },
  )
  return respuesta.defectos ?? []
}

/** Defectos detectados (`deteccion`) o resueltos (`resolucion`) en un sprint (US-18). */
export async function listarDefectosPorSprint(
  proyectoId: number,
  sprintId: number,
  tipo: TipoListadoDefecto,
  signal?: AbortSignal,
): Promise<Defecto[]> {
  const respuesta = await api.get<RespuestaListadoDefectos>(
    `/proyectos/${proyectoId}/sprints/${sprintId}/defectos`,
    { query: { tipo }, signal },
  )
  return respuesta.defectos ?? []
}

export function crearDefecto(proyectoId: number, alta: AltaDefecto): Promise<Defecto> {
  return api.post<Defecto>(`/proyectos/${proyectoId}/defectos`, alta)
}

export function actualizarDefecto(
  proyectoId: number,
  defectoId: number,
  actualizacion: ActualizacionDefecto,
): Promise<Defecto> {
  return api.patch<Defecto>(
    `/proyectos/${proyectoId}/defectos/${defectoId}`,
    actualizacion,
  )
}
