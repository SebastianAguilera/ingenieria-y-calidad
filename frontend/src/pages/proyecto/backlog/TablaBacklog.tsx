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
  guardandoId: number | null
}

export default function TablaBacklog({
  historias,
  onCambiar,
  onEditar,
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
        <Button
          variante="ghost"
          onClick={() => onEditar(historia)}
          aria-label={`Editar ${historia.codigo}`}
          disabled={guardandoId === historia.id}
        >
          Editar
        </Button>
      )}
    />
  )
}
