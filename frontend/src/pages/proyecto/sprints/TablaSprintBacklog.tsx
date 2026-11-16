import type { ActualizacionHistoria, EstadoSprint, Historia, SprintBacklogItem } from '../../../api/types'
import {
  ESTADOS_HISTORIA,
  PRIORIDADES_HISTORIA,
  claseEstado,
  clasePrioridad,
} from '../backlog/etiquetas'
import SelectPill from '../backlog/SelectPill'
import TablaHistorias from '../backlog/TablaHistorias'
import type { ColumnaHistoria } from '../backlog/TablaHistorias'
import { formatearHoras } from '../metricas/formato'

type TablaSprintBacklogProps = {
  items: SprintBacklogItem[]
  estadoSprint: EstadoSprint
  onQuitar: (historia: Historia) => void
  onCambiar: (historia: Historia, cambios: ActualizacionHistoria) => void
  onEditar: (historia: Historia) => void
  onEstimar: (historia: Historia) => void
  deshabilitado?: boolean
}

const SELECT_PILL = 'min-w-[125px]'
const BOTON_ACCION =
  'inline-flex items-center gap-1.5 rounded-tile border border-border bg-surface px-3 py-1.5 text-xs font-semibold text-ink-soft transition hover:border-primary hover:bg-[var(--primary-100)] hover:text-primary focus-visible:shadow-[var(--ring)] disabled:cursor-not-allowed disabled:opacity-60'

function Chip({ texto, clase }: { texto: string; clase: string }) {
  return (
    <span className={`inline-flex items-center rounded-pill border px-3 py-1.5 text-xs font-semibold ${clase}`}>
      {texto}
    </span>
  )
}

export default function TablaSprintBacklog({
  items,
  estadoSprint,
  onQuitar,
  onCambiar,
  onEditar,
  onEstimar,
  deshabilitado = false,
}: TablaSprintBacklogProps) {
  const cerrado = estadoSprint === 'Cerrado'

  const columnas: ColumnaHistoria[] = [
    {
      clave: 'codigo',
      encabezado: 'Código',
      ancho: 'w-[8%]',
      alineacion: 'centro',
      render: (historia) => (
        <span className="font-mono text-xs font-semibold text-ink-soft">{historia.codigo}</span>
      ),
    },
    {
      clave: 'titulo',
      encabezado: 'Título',
      ancho: 'w-[18%]',
      alineacion: 'centro',
      render: (historia) => (
        <span className="block w-full min-w-0 truncate font-medium text-ink" title={historia.titulo}>
          {historia.titulo}
        </span>
      ),
    },
    {
      clave: 'estado',
      encabezado: 'Estado',
      ancho: 'w-[15%]',
      alineacion: 'centro',
      render: (historia) =>
        cerrado ? (
          <Chip texto={historia.estado} clase={claseEstado(historia.estado)} />
        ) : (
          <SelectPill
            label={`Estado de ${historia.codigo}`}
            value={historia.estado}
            opciones={ESTADOS_HISTORIA}
            onChange={(estado) => onCambiar(historia, { estado })}
            disabled={deshabilitado}
            className={`${claseEstado(historia.estado)} ${SELECT_PILL}`}
          />
        ),
    },
    {
      clave: 'prioridad',
      encabezado: 'Prioridad',
      ancho: 'w-[15%]',
      alineacion: 'centro',
      render: (historia) =>
        cerrado ? (
          <Chip texto={historia.prioridad} clase={clasePrioridad(historia.prioridad)} />
        ) : (
          <SelectPill
            label={`Prioridad de ${historia.codigo}`}
            value={historia.prioridad}
            opciones={PRIORIDADES_HISTORIA}
            onChange={(prioridad) => onCambiar(historia, { prioridad })}
            disabled={deshabilitado}
            className={`${clasePrioridad(historia.prioridad)} ${SELECT_PILL}`}
          />
        ),
    },
    {
      clave: 'estimacion',
      encabezado: 'Estimación',
      ancho: 'w-[11%]',
      alineacion: 'centro',
      render: (historia) =>
        historia.story_points === null ? (
          <span className="text-xs font-medium text-ink-soft">Sin estimar</span>
        ) : (
          <span
            className="inline-flex min-w-8 items-center justify-center rounded-pill border border-[var(--primary-300)] bg-[var(--primary-100)] px-2.5 py-0.5 text-xs font-semibold text-primary"
            title="Puntos de historia estimados"
          >
            {historia.story_points}
          </span>
        ),
    },
    {
      clave: 'horas_estimadas',
      encabezado: 'Horas estimadas',
      ancho: 'w-[10%]',
      alineacion: 'centro',
      render: (historia) =>
        historia.horas_estimadas === null ? (
          <span className="text-xs font-medium text-ink-soft" title="Sin horas estimadas">
            —
          </span>
        ) : (
          <span className="text-xs font-semibold text-ink">{formatearHoras(historia.horas_estimadas)}</span>
        ),
    },
  ]

  return (
    <div className="[&_table]:min-w-[60rem]">
      <TablaHistorias
        historias={items.map((item) => item.historia)}
        columnas={columnas}
        layoutFijo
        anchoAcciones="w-[23%]"
        acciones={(historia) => (
          <div className="tabla-acciones">
            <button
              type="button"
              onClick={() => onEstimar(historia)}
              disabled={deshabilitado}
              aria-label={`Estimar ${historia.codigo}`}
              className={BOTON_ACCION}
            >
              <i className="bi bi-suit-spade-fill" aria-hidden="true" />
              Estimar
            </button>
            <button
              type="button"
              onClick={() => onEditar(historia)}
              disabled={deshabilitado}
              aria-label={`Editar historia ${historia.codigo}`}
              className={BOTON_ACCION}
            >
              <i className="bi bi-pencil-square" aria-hidden="true" />
              Editar
            </button>
            {!cerrado && (
              <button
                type="button"
                onClick={() => onQuitar(historia)}
                disabled={deshabilitado}
                aria-label={`Quitar ${historia.codigo} del sprint`}
                className={BOTON_ACCION}
              >
                Quitar del sprint
              </button>
            )}
          </div>
        )}
      />
    </div>
  )
}
