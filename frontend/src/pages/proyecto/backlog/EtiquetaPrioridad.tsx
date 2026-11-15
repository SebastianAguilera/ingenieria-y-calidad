import type { PrioridadHistoria } from '../../../api/types'
import { clasePrioridad } from './etiquetas'

export default function EtiquetaPrioridad({ prioridad }: { prioridad: PrioridadHistoria }) {
  return (
    <span
      className={`inline-flex rounded-pill border px-2.5 py-0.5 text-xs font-semibold ${clasePrioridad(prioridad)}`.trim()}
    >
      {prioridad}
    </span>
  )
}
