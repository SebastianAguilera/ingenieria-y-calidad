import { api } from './client'
import type { SesionPokerActiva, SesionPokerCreada, SesionPokerHistorial } from './types'

export async function listarSesionesPoker(
  proyectoId: number,
  historiaId: number,
  signal?: AbortSignal,
): Promise<SesionPokerHistorial[]> {
  const sesiones = await api.get<SesionPokerHistorial[] | null>(
    `/proyectos/${proyectoId}/historias/${historiaId}/poker-sessions`,
    { signal },
  )
  return sesiones ?? []
}

export function crearSesionPoker(
  proyectoId: number,
  historiaId: number,
): Promise<SesionPokerCreada> {
  return api.post<SesionPokerCreada>(
    `/proyectos/${proyectoId}/historias/${historiaId}/poker-sessions`,
  )
}

export async function listarSesionesActivas(
  proyectoId: number,
  signal?: AbortSignal,
): Promise<SesionPokerActiva[]> {
  const sesiones = await api.get<SesionPokerActiva[] | null>(
    `/proyectos/${proyectoId}/poker-sessions/activas`,
    { signal },
  )
  return sesiones ?? []
}
