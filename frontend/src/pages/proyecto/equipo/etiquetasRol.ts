import type { RolProyecto } from '../../../api/types'

const CLASES: Record<RolProyecto, string> = {
  LIDER: 'border-[var(--accent-500)] bg-[var(--accent-100)] text-[var(--accent-600)]',
  MIEMBRO: 'border-border bg-[var(--primary-100)] text-primary',
}

export function claseRol(rol: RolProyecto): string {
  return CLASES[rol]
}

export const ROLES: readonly RolProyecto[] = ['LIDER', 'MIEMBRO']
