import { api } from './client'
import type { AltaSprint, EstadoSprint, Sprint, SprintBacklog, SprintBacklogItem } from './types'

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

type RespuestaSprintBacklog = {
  sprint: Sprint
  total: number
  items?: SprintBacklogItem[] | null
}

export async function obtenerSprintBacklog(
  proyectoId: number,
  sprintId: number,
  signal?: AbortSignal,
): Promise<SprintBacklog> {
  const respuesta = await api.get<RespuestaSprintBacklog>(
    `/proyectos/${proyectoId}/sprints/${sprintId}/historias`,
    { signal },
  )
  return { ...respuesta, items: respuesta.items ?? [] }
}

export function asignarHistoria(
  proyectoId: number,
  sprintId: number,
  historiaId: number,
): Promise<void> {
  return api.post<void>(`/proyectos/${proyectoId}/sprints/${sprintId}/historias`, {
    historia_id: historiaId,
  })
}

export function quitarHistoria(
  proyectoId: number,
  sprintId: number,
  historiaId: number,
): Promise<void> {
  return api.delete<void>(`/proyectos/${proyectoId}/sprints/${sprintId}/historias/${historiaId}`)
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
