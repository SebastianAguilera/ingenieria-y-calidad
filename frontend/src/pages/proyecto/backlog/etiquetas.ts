import type { EstadoHistoria, PrioridadHistoria } from '../../../api/types'

export type CampoPriorizacion = {
  clave: 'valor_negocio' | 'criticidad_tiempo' | 'reduccion_riesgo'
  etiqueta: string
  ayuda: string
}

export const ESTADOS_HISTORIA: readonly EstadoHistoria[] = ['Backlog', 'En Progreso', 'Hecho']

export const PRIORIDADES_HISTORIA: readonly PrioridadHistoria[] = ['Alta', 'Media', 'Baja']

export const VALORES_PRIORIZACION: readonly number[] = [1, 2, 3, 5, 8, 13, 21]

export const CAMPOS_PRIORIZACION: readonly CampoPriorizacion[] = [
  {
    clave: 'valor_negocio',
    etiqueta: 'Valor de negocio',
    ayuda: 'Valor que aporta la historia al negocio o al cliente.',
  },
  {
    clave: 'criticidad_tiempo',
    etiqueta: 'Criticidad en el tiempo',
    ayuda: 'Urgencia de entregar la historia en el tiempo.',
  },
  {
    clave: 'reduccion_riesgo',
    etiqueta: 'Reducción de riesgo',
    ayuda: 'Riesgos u oportunidades que reduce la historia.',
  },
]

/**
 * Estilos por estado de historia, en un solo lugar: tablero, selectores y tabla
 * usan estas clases, que leen las variables `--estado-*` de `theme.css`
 * (claro y oscuro). Backlog es índigo, En Progreso ámbar y Hecho verde.
 */
type EstilosEstado = {
  /** Chip o selector de estado. */
  chip: string
  /** Columna del tablero: franja superior y fondo teñido en degradé. */
  columna: string
  /** Punto de color junto al título de la columna. */
  punto: string
  /** Contador de la columna. */
  contador: string
  /** Tarjeta: borde izquierdo y brillo al pasar el mouse. */
  tarjeta: string
}

const ESTILOS_ESTADO: Record<EstadoHistoria, EstilosEstado> = {
  Backlog: {
    chip: 'border-[var(--estado-backlog)] bg-[color-mix(in_oklab,var(--estado-backlog)_16%,transparent)] text-[var(--estado-backlog-ink)]',
    columna:
      'border-t-4 border-t-[var(--estado-backlog)] bg-[linear-gradient(to_bottom,color-mix(in_oklab,var(--estado-backlog)_14%,transparent),transparent_70%)]',
    punto: 'bg-[var(--estado-backlog)]',
    contador:
      'bg-[color-mix(in_oklab,var(--estado-backlog)_18%,transparent)] text-[var(--estado-backlog-ink)]',
    tarjeta:
      'border-l-[3px] border-l-[var(--estado-backlog)] hover:shadow-[0_0_0_1px_color-mix(in_oklab,var(--estado-backlog)_45%,transparent),0_6px_18px_color-mix(in_oklab,var(--estado-backlog)_30%,transparent)]',
  },
  'En Progreso': {
    chip: 'border-[var(--estado-progreso)] bg-[color-mix(in_oklab,var(--estado-progreso)_16%,transparent)] text-[var(--estado-progreso-ink)]',
    columna:
      'border-t-4 border-t-[var(--estado-progreso)] bg-[linear-gradient(to_bottom,color-mix(in_oklab,var(--estado-progreso)_14%,transparent),transparent_70%)]',
    punto: 'bg-[var(--estado-progreso)]',
    contador:
      'bg-[color-mix(in_oklab,var(--estado-progreso)_18%,transparent)] text-[var(--estado-progreso-ink)]',
    tarjeta:
      'border-l-[3px] border-l-[var(--estado-progreso)] hover:shadow-[0_0_0_1px_color-mix(in_oklab,var(--estado-progreso)_45%,transparent),0_6px_18px_color-mix(in_oklab,var(--estado-progreso)_30%,transparent)]',
  },
  Hecho: {
    chip: 'border-[var(--estado-hecho)] bg-[color-mix(in_oklab,var(--estado-hecho)_16%,transparent)] text-[var(--estado-hecho-ink)]',
    columna:
      'border-t-4 border-t-[var(--estado-hecho)] bg-[linear-gradient(to_bottom,color-mix(in_oklab,var(--estado-hecho)_14%,transparent),transparent_70%)]',
    punto: 'bg-[var(--estado-hecho)]',
    contador:
      'bg-[color-mix(in_oklab,var(--estado-hecho)_18%,transparent)] text-[var(--estado-hecho-ink)]',
    tarjeta:
      'border-l-[3px] border-l-[var(--estado-hecho)] hover:shadow-[0_0_0_1px_color-mix(in_oklab,var(--estado-hecho)_45%,transparent),0_6px_18px_color-mix(in_oklab,var(--estado-hecho)_30%,transparent)]',
  },
}

const CLASES_PRIORIDAD: Record<PrioridadHistoria, string> = {
  Alta: 'border-[var(--danger)] bg-[var(--accent-100)] text-[var(--danger)]',
  Media: 'border-[var(--border-strong)] bg-[var(--primary-100)] text-primary',
  Baja: 'border-border bg-[var(--surface-2)] text-ink-soft',
}

export function claseEstado(estado: EstadoHistoria): string {
  return ESTILOS_ESTADO[estado].chip
}

export function claseColumnaEstado(estado: EstadoHistoria): string {
  return ESTILOS_ESTADO[estado].columna
}

export function clasePuntoEstado(estado: EstadoHistoria): string {
  return ESTILOS_ESTADO[estado].punto
}

export function claseContadorEstado(estado: EstadoHistoria): string {
  return ESTILOS_ESTADO[estado].contador
}

export function claseTarjetaEstado(estado: EstadoHistoria): string {
  return ESTILOS_ESTADO[estado].tarjeta
}

export function clasePrioridad(prioridad: PrioridadHistoria): string {
  return CLASES_PRIORIDAD[prioridad]
}
