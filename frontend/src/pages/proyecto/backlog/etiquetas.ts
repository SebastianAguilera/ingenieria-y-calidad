import type { EstadoHistoria, PrioridadHistoria } from '../../../api/types'

export const ESTADOS_HISTORIA: readonly EstadoHistoria[] = ['Backlog', 'En Progreso', 'Hecho']

export const PRIORIDADES_HISTORIA: readonly PrioridadHistoria[] = ['Alta', 'Media', 'Baja']

const CLASES_ESTADO: Record<EstadoHistoria, string> = {
  Backlog: 'border-[var(--border-strong)] bg-[var(--primary-100)] text-primary',
  'En Progreso': 'border-[var(--accent-500)] bg-[var(--accent-100)] text-[var(--accent-600)]',
  Hecho: 'border-[var(--success)] bg-[var(--primary-100)] text-[var(--success)]',
}

const CLASES_PRIORIDAD: Record<PrioridadHistoria, string> = {
  Alta: 'border-[var(--danger)] bg-[var(--accent-100)] text-[var(--danger)]',
  Media: 'border-[var(--border-strong)] bg-[var(--primary-100)] text-primary',
  Baja: 'border-border bg-[var(--surface-2)] text-ink-soft',
}

export function claseEstado(estado: EstadoHistoria): string {
  return CLASES_ESTADO[estado]
}

export function clasePrioridad(prioridad: PrioridadHistoria): string {
  return CLASES_PRIORIDAD[prioridad]
}
