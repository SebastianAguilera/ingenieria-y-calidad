import type { EstadoDefecto, SeveridadDefecto } from '../../../api/types'

export const LIMITE_DESCRIPCION_TABLA = 80

export const ESTADOS_DEFECTO: readonly EstadoDefecto[] = [
  'open',
  'in_progress',
  'resolved',
  'closed',
  'rejected',
]

export const SEVERIDADES_DEFECTO: readonly SeveridadDefecto[] = [
  'minor',
  'major',
  'critical',
  'blocker',
]

const ETIQUETAS_ESTADO: Record<EstadoDefecto, string> = {
  open: 'Abierto',
  in_progress: 'En progreso',
  resolved: 'Resuelto',
  closed: 'Cerrado',
  rejected: 'Rechazado',
}

const ETIQUETAS_SEVERIDAD: Record<SeveridadDefecto, string> = {
  minor: 'Menor',
  major: 'Mayor',
  critical: 'Crítica',
  blocker: 'Bloqueante',
}

const CLASES_ESTADO: Record<EstadoDefecto, string> = {
  open: 'border-primary bg-transparent text-ink',
  in_progress: 'border-transparent bg-primary text-[var(--on-primary)]',
  resolved: 'border-transparent bg-[var(--success)] text-white',
  closed: 'border-transparent bg-ink-soft text-[var(--on-primary)]',
  rejected: 'border-border bg-[var(--surface-2)] text-ink-soft',
}

const CLASES_SEVERIDAD: Record<SeveridadDefecto, string> = {
  minor: 'border-border bg-[var(--surface-2)] text-ink',
  major: 'border-primary bg-transparent text-ink',
  critical: 'border-[var(--danger)] bg-[var(--accent-100)] text-[var(--danger)]',
  blocker: 'border-transparent bg-[var(--danger)] text-white',
}

export function etiquetaEstado(estado: EstadoDefecto): string {
  return ETIQUETAS_ESTADO[estado]
}

export function etiquetaSeveridad(severidad: SeveridadDefecto): string {
  return ETIQUETAS_SEVERIDAD[severidad]
}

export function claseEstadoDefecto(estado: EstadoDefecto): string {
  return CLASES_ESTADO[estado]
}

export function claseSeveridadDefecto(severidad: SeveridadDefecto): string {
  return CLASES_SEVERIDAD[severidad]
}

/** El código `DEF-{id}` es solo de presentación: el backend no lo guarda ni lo devuelve. */
export function codigoDefecto(id: number): string {
  return `DEF-${id}`
}

/** Recorta por caracteres (no por bytes) y agrega `…` solo si realmente se recortó. */
export function truncarDescripcion(texto: string): string {
  const caracteres = [...texto]
  return caracteres.length > LIMITE_DESCRIPCION_TABLA
    ? `${caracteres.slice(0, LIMITE_DESCRIPCION_TABLA).join('')}…`
    : texto
}
