import { api } from './client'
import type {
  DetalleSesionPoker,
  ResultadoVotoPoker,
  SesionPokerActiva,
  SesionPokerCreada,
  SesionPokerHistorial,
} from './types'

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

export function obtenerDetalleSesionPoker(
  proyectoId: number,
  sesionId: number,
  signal?: AbortSignal,
): Promise<DetalleSesionPoker> {
  return api.get<DetalleSesionPoker>(`/proyectos/${proyectoId}/poker-sessions/${sesionId}`, {
    signal,
  })
}

export function revelarSesionPoker(
  proyectoId: number,
  sesionId: number,
): Promise<DetalleSesionPoker> {
  return api.post<DetalleSesionPoker>(`/proyectos/${proyectoId}/poker-sessions/${sesionId}/revelar`)
}

export function abrirNuevaRondaSesionPoker(
  proyectoId: number,
  sesionId: number,
): Promise<DetalleSesionPoker> {
  return api.post<DetalleSesionPoker>(`/proyectos/${proyectoId}/poker-sessions/${sesionId}/rondas`)
}

export function votarSesionPoker(
  proyectoId: number,
  sesionId: number,
  puntos: number,
): Promise<ResultadoVotoPoker> {
  return api.post<ResultadoVotoPoker>(`/proyectos/${proyectoId}/poker-sessions/${sesionId}/votos`, {
    puntos,
  })
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
