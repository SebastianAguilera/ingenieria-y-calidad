import { api } from './client'
import type { AltaProyecto, Proyecto, ProyectoListado, RolProyecto, UsuarioProyecto } from './types'

export function crearProyecto(alta: AltaProyecto): Promise<Proyecto> {
  return api.post<Proyecto>('/proyectos', alta)
}

export function asociarUsuario(
  proyectoId: number,
  usuarioId: number,
  rol: RolProyecto,
): Promise<UsuarioProyecto> {
  return api.post<UsuarioProyecto>(`/proyectos/${proyectoId}/usuarios`, {
    usuario_id: usuarioId,
    rol,
  })
}

type RespuestaListadoProyectos = {
  total: number
  proyectos?: ProyectoListado[] | null
}

export async function listarProyectos(signal?: AbortSignal): Promise<ProyectoListado[]> {
  const respuesta = await api.get<RespuestaListadoProyectos>('/proyectos', { signal })
  return respuesta.proyectos ?? []
}
