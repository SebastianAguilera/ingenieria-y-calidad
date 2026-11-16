import type { EstadoHistoria, Historia, Sprint } from '../../../api/types'

export type CantidadPorEstado = { estado: EstadoHistoria; cantidad: number }

const ORDEN_ESTADOS: readonly EstadoHistoria[] = ['Backlog', 'En Progreso', 'Hecho']

/**
 * Distribución de historias por estado para la dona. Es el único dato que el
 * dashboard cuenta en el frontend (US-F21, RN-01).
 */
export function contarPorEstado(historias: readonly Historia[]): CantidadPorEstado[] {
  return ORDEN_ESTADOS.map((estado) => ({
    estado,
    cantidad: historias.filter((historia) => historia.estado === estado).length,
  }))
}

/** Sprints del gráfico de barras: los cerrados y el activo, en el orden recibido (RN-07). */
export function sprintsDelGrafico(sprints: readonly Sprint[]): Sprint[] {
  return sprints.filter((sprint) => sprint.estado === 'Cerrado' || sprint.estado === 'Activo')
}

/** Sprint elegido al entrar: el activo; si no, el último cerrado; si no, el primero. */
export function sprintInicial(sprints: readonly Sprint[]): number | null {
  const activo = sprints.find((sprint) => sprint.estado === 'Activo')
  if (activo) {
    return activo.id
  }
  const cerrados = sprints.filter((sprint) => sprint.estado === 'Cerrado')
  if (cerrados.length > 0) {
    return cerrados[cerrados.length - 1].id
  }
  return sprints[0]?.id ?? null
}
