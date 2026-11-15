import { api } from './client'
import type { ActualizacionHistoria, AltaHistoria, Historia } from './types'

type RespuestaListadoHistorias = {
  total: number
  historias?: Historia[] | null
}

export async function listarHistorias(
  proyectoId: number,
  signal?: AbortSignal,
): Promise<Historia[]> {
  const respuesta = await api.get<RespuestaListadoHistorias>(
    `/proyectos/${proyectoId}/historias`,
    { signal },
  )
  return respuesta.historias ?? []
}

export function crearHistoria(proyectoId: number, alta: AltaHistoria): Promise<Historia> {
  return api.post<Historia>(`/proyectos/${proyectoId}/historias`, alta)
}

export function actualizarHistoria(
  proyectoId: number,
  historiaId: number,
  actualizacion: ActualizacionHistoria,
): Promise<Historia> {
  return api.patch<Historia>(
    `/proyectos/${proyectoId}/historias/${historiaId}`,
    actualizacion,
  )
}
