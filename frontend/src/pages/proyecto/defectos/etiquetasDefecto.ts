import type { EstadoDefecto, SeveridadDefecto } from '../../../api/types'
import { COLOR_SEVERIDAD } from '../metricas/estilosGrafico'

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

/**
 * Chips de estado: pastilla sin relleno, con borde y punto del color del
 * estado. Así se distinguen de los de severidad, que son pastillas
 * con fondo teñido. El color del punto sigue las variables `--estado-*` de
 * `theme.css`: Abierto índigo, En progreso ámbar, Resuelto verde, Cerrado gris
 * y Rechazado rojo.
 */
const BASE_ESTADO =
  'gap-2 bg-transparent text-ink before:size-2 before:shrink-0 before:rounded-full before:content-[""]'

const CLASES_ESTADO: Record<EstadoDefecto, string> = {
  open: `${BASE_ESTADO} border-[var(--estado-backlog)] before:bg-[var(--estado-backlog)]`,
  in_progress: `${BASE_ESTADO} border-[var(--estado-progreso)] before:bg-[var(--estado-progreso)]`,
  resolved: `${BASE_ESTADO} border-[var(--estado-hecho)] before:bg-[var(--estado-hecho)]`,
  closed: `${BASE_ESTADO} border-ink-soft before:bg-ink-soft`,
  rejected: `${BASE_ESTADO} border-[var(--danger)] before:bg-[var(--danger)]`,
}

/**
 * Chips de severidad: usan la paleta de los gráficos de métricas
 * (`COLOR_SEVERIDAD` en `metricas/estilosGrafico.ts`), de Menor a Bloqueante.
 * Las clases son literales para que Tailwind las detecte; una prueba comprueba
 * que coincidan con `COLOR_SEVERIDAD`.
 */
const CLASES_SEVERIDAD: Record<SeveridadDefecto, string> = {
  minor: 'border-[var(--chart-3)] bg-[color-mix(in_oklab,var(--chart-3)_30%,transparent)] text-ink',
  major: 'border-[var(--chart-1)] bg-[color-mix(in_oklab,var(--chart-1)_20%,transparent)] text-ink',
  critical: 'border-[var(--chart-2)] bg-[color-mix(in_oklab,var(--chart-2)_22%,transparent)] text-ink',
  blocker: 'border-[var(--accent-600)] bg-[color-mix(in_oklab,var(--accent-600)_28%,transparent)] text-ink',
}

/** Color sólido de la severidad, para la leyenda y las barras del gráfico. */
export function colorSeveridad(severidad: SeveridadDefecto): string {
  return COLOR_SEVERIDAD[severidad]
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
