import { api } from './client'
import type { AltaSprint, EstadoSprint, Sprint } from './types'

type RespuestaListadoSprints = {
  total: number
  sprints?: Sprint[] | null
}

export async function listarSprints(
  proyectoId: number,
  signal?: AbortSignal,
): Promise<Sprint[]> {
  const respuesta = await api.get<RespuestaListadoSprints>(
    `/proyectos/${proyectoId}/sprints`,
    { signal },
  )
  return respuesta.sprints ?? []
}

export function crearSprint(proyectoId: number, alta: AltaSprint): Promise<Sprint> {
  return api.post<Sprint>(`/proyectos/${proyectoId}/sprints`, alta)
}

export function cambiarEstadoSprint(
  proyectoId: number,
  sprintId: number,
  estado: EstadoSprint,
): Promise<Sprint> {
  return api.patch<Sprint>(`/proyectos/${proyectoId}/sprints/${sprintId}/estado`, {
    estado,
  })
}
