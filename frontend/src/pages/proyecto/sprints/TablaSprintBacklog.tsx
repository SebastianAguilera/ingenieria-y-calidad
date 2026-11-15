import type { EstadoSprint, Historia, SprintBacklogItem } from '../../../api/types'
import Button from '../../../ui/Button'
import { claseEstado } from '../backlog/etiquetas'
import TablaHistorias from '../backlog/TablaHistorias'
import type { ColumnaHistoria } from '../backlog/TablaHistorias'

type TablaSprintBacklogProps = {
  items: SprintBacklogItem[]
  estadoSprint: EstadoSprint
  onQuitar: (historia: Historia) => void
  deshabilitado?: boolean
}

const COLUMNAS: ColumnaHistoria[] = [
  {
    clave: 'codigo',
    encabezado: 'Código',
    render: (historia) => (
      <span className="font-mono text-xs font-semibold text-ink-soft">{historia.codigo}</span>
    ),
  },
  {
    clave: 'titulo',
    encabezado: 'Título',
    render: (historia) => <span className="font-medium text-ink">{historia.titulo}</span>,
  },
  {
    clave: 'estado',
    encabezado: 'Estado',
    render: (historia) => (
      <span
        className={`inline-flex items-center rounded-pill border px-3 py-1.5 text-xs font-semibold ${claseEstado(historia.estado)}`}
      >
        {historia.estado}
      </span>
    ),
  },
  {
    clave: 'story_points',
    encabezado: 'Story Points',
    render: (historia) => historia.story_points ?? '—',
  },
]

/**
 * Sprint Backlog: las historias asignadas a un sprint. Un sprint cerrado es de
 * solo lectura, por lo que no ofrece la acción de quitar.
 */
export default function TablaSprintBacklog({
  items,
  estadoSprint,
  onQuitar,
  deshabilitado = false,
}: TablaSprintBacklogProps) {
  const puedeQuitar = estadoSprint !== 'Cerrado'

  return (
    <TablaHistorias
      historias={items.map((item) => item.historia)}
      columnas={COLUMNAS}
      acciones={
        puedeQuitar
          ? (historia) => (
              <Button
                variante="ghost"
                onClick={() => onQuitar(historia)}
                aria-label={`Quitar ${historia.codigo} del sprint`}
                disabled={deshabilitado}
              >
                Quitar del sprint
              </Button>
            )
          : undefined
      }
    />
  )
}
