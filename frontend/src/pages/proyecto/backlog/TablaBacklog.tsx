import type { ActualizacionHistoria, Historia } from '../../../api/types'
import Button from '../../../ui/Button'
import { ESTADOS_HISTORIA, PRIORIDADES_HISTORIA, claseEstado, clasePrioridad } from './etiquetas'
import SelectPill from './SelectPill'
import TablaHistorias from './TablaHistorias'
import type { ColumnaHistoria } from './TablaHistorias'

type TablaBacklogProps = {
  historias: Historia[]
  onCambiar: (historia: Historia, cambios: ActualizacionHistoria) => void
  onEditar: (historia: Historia) => void
  onPoker?: (historia: Historia) => void
  pokerActivo?: ReadonlySet<number>
  guardandoId: number | null
}

const SIN_POKER_ACTIVO: ReadonlySet<number> = new Set()

export default function TablaBacklog({
  historias,
  onCambiar,
  onEditar,
  onPoker,
  pokerActivo = SIN_POKER_ACTIVO,
  guardandoId,
}: TablaBacklogProps) {
  const columnas: ColumnaHistoria[] = [
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
        <SelectPill
          label={`Estado de ${historia.codigo}`}
          value={historia.estado}
          opciones={ESTADOS_HISTORIA}
          onChange={(estado) => onCambiar(historia, { estado })}
          disabled={guardandoId === historia.id}
          className={claseEstado(historia.estado)}
        />
      ),
    },
    {
      clave: 'prioridad',
      encabezado: 'Prioridad',
      render: (historia) => (
        <SelectPill
          label={`Prioridad de ${historia.codigo}`}
          value={historia.prioridad}
          opciones={PRIORIDADES_HISTORIA}
          onChange={(prioridad) => onCambiar(historia, { prioridad })}
          disabled={guardandoId === historia.id}
          className={clasePrioridad(historia.prioridad)}
        />
      ),
    },
  ]

  return (
    <TablaHistorias
      historias={historias}
      columnas={columnas}
      acciones={(historia) => (
        <div className="flex flex-wrap items-center gap-1">
          {pokerActivo.has(historia.id) && (
            <span className="inline-flex items-center rounded-pill border border-[var(--primary-300)] bg-[var(--primary-100)] px-2.5 py-1 text-xs font-semibold text-primary">
              Planning Poker en curso
            </span>
          )}
          {onPoker && (
            <Button
              variante="ghost"
              onClick={() => onPoker(historia)}
              aria-label={`Planning Poker de ${historia.codigo}`}
              disabled={guardandoId === historia.id}
            >
              Planning Poker
            </Button>
          )}
          <Button
            variante="ghost"
            onClick={() => onEditar(historia)}
            aria-label={`Editar ${historia.codigo}`}
            disabled={guardandoId === historia.id}
          >
            Editar
          </Button>
        </div>
      )}
    />
  )
}
