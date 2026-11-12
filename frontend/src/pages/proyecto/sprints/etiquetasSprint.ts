import type { EstadoSprint } from '../../../api/types'

const CLASES_ESTADO: Record<EstadoSprint, string> = {
  Planificado: 'border-[var(--border-strong)] bg-[var(--primary-100)] text-primary',
  Activo: 'border-[var(--success)] bg-[var(--primary-100)] text-[var(--success)]',
  Cerrado: 'border-border bg-[var(--surface-2)] text-ink-soft',
}

export function claseEstadoSprint(estado: EstadoSprint): string {
  return CLASES_ESTADO[estado]
}
