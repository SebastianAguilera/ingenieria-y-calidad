import { api } from './client'
import type {
  ActualizacionProyecto,
  AltaProyecto,
  EstadoProyecto,
  Proyecto,
  ProyectoListado,
  RolProyecto,
  UsuarioProyecto,
} from './types'

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

export function listarIntegrantes(
  proyectoId: number,
  signal?: AbortSignal,
): Promise<UsuarioProyecto[]> {
  return api.get<UsuarioProyecto[]>(`/proyectos/${proyectoId}/usuarios`, { signal })
}

export function agregarIntegrantePorEmail(
  proyectoId: number,
  email: string,
  rol: RolProyecto = 'MIEMBRO',
): Promise<UsuarioProyecto> {
  return api.post<UsuarioProyecto>(`/proyectos/${proyectoId}/usuarios`, { email, rol })
}

export function cambiarRolIntegrante(
  proyectoId: number,
  usuarioId: number,
  rol: RolProyecto,
): Promise<{ usuario_id: number; rol: RolProyecto }> {
  return api.patch<{ usuario_id: number; rol: RolProyecto }>(
    `/proyectos/${proyectoId}/usuarios/${usuarioId}`,
    { rol },
  )
}

export function quitarIntegrante(proyectoId: number, usuarioId: number): Promise<void> {
  return api.delete<void>(`/proyectos/${proyectoId}/usuarios/${usuarioId}`)
}

type RespuestaListadoProyectos = {
  total: number
  proyectos?: ProyectoListado[] | null
}

export async function listarProyectos(signal?: AbortSignal): Promise<ProyectoListado[]> {
  const respuesta = await api.get<RespuestaListadoProyectos>('/proyectos', { signal })
  return respuesta.proyectos ?? []
}

export function obtenerProyecto(id: number, signal?: AbortSignal): Promise<Proyecto> {
  return api.get<Proyecto>(`/proyectos/${id}`, { signal })
}

export function actualizarProyecto(
  id: number,
  actualizacion: ActualizacionProyecto,
): Promise<Proyecto> {
  return api.put<Proyecto>(`/proyectos/${id}`, actualizacion)
}

export function cambiarEstadoProyecto(id: number, estado: EstadoProyecto): Promise<Proyecto> {
  return api.patch<Proyecto>(`/proyectos/${id}/estado`, { estado })
}

export function eliminarProyecto(id: number): Promise<void> {
  return api.delete<void>(`/proyectos/${id}`)
}
