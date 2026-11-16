/**
 * Estilos compartidos de los gráficos de Recharts. Solo tokens de `theme.css`
 * (skill `teal-dashboard-ui`, §1 y §7): nada de colores literales.
 */
import type { SeveridadDefecto } from '../../../api/types'

export const COLOR_PLANIFICADOS = 'var(--chart-3)'
export const COLOR_COMPLETADOS = 'var(--chart-1)'

/**
 * Severidad de defectos con la misma paleta de los demás gráficos (--chart-1,
 * --chart-2 y --chart-3), de Menor a Bloqueante, del más suave al más intenso.
 * Es la única fuente: la leyenda, las barras y los chips de Defectos la usan.
 */
export const COLOR_SEVERIDAD: Record<SeveridadDefecto, string> = {
  minor: 'var(--chart-3)',
  major: 'var(--chart-1)',
  critical: 'var(--chart-2)',
  blocker: 'var(--accent-600)',
}

export const TICK_EJE = { fill: 'var(--ink-soft)', fontSize: 11 }

export const ESTILO_TOOLTIP = {
  background: 'var(--surface)',
  border: '1px solid var(--border)',
  borderRadius: 'var(--radius-tile)',
  color: 'var(--ink)',
  fontSize: 12,
}
