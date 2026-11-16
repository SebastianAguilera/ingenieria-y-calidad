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
