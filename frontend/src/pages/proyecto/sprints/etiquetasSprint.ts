import type { EstadoSprint } from '../../../api/types'

const CLASES_ESTADO: Record<EstadoSprint, string> = {
  Planificado: 'border-transparent bg-primary text-[var(--on-primary)]',
  Activo: 'border-transparent bg-[var(--success)] text-white',
  Cerrado: 'border-transparent bg-ink-soft text-[var(--on-primary)]',
}

export function claseEstadoSprint(estado: EstadoSprint): string {
  return CLASES_ESTADO[estado]
}
