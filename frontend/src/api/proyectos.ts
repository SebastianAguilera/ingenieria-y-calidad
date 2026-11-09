import { api } from './client'
import type { ProyectoListado } from './types'

type RespuestaListadoProyectos = {
  total: number
  proyectos?: ProyectoListado[] | null
}

export async function listarProyectos(signal?: AbortSignal): Promise<ProyectoListado[]> {
  const respuesta = await api.get<RespuestaListadoProyectos>('/proyectos', { signal })
  return respuesta.proyectos ?? []
}
