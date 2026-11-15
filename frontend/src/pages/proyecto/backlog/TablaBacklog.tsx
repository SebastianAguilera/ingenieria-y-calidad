import type { ActualizacionHistoria, Historia } from '../../../api/types'
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

/** Cápsulas de Estado y Prioridad: ancho mínimo uniforme (el padding lo define SelectPill). */
const SELECT_PILL = 'min-w-[125px]'

const BOTON_ACCION =
  'inline-flex items-center gap-1.5 rounded-tile border border-border bg-surface px-3 py-1.5 text-xs font-semibold text-ink-soft transition hover:border-primary hover:bg-[var(--primary-100)] hover:text-primary focus-visible:shadow-[var(--ring)] disabled:cursor-not-allowed disabled:opacity-60'

/** Encabezados compactos y tabla con ancho mínimo propio del Backlog. */
const CLASE_TABLA =
  '[&_table]:min-w-[60rem] [&_thead_th]:py-3 [&_thead_th]:text-xs [&_thead_th]:uppercase [&_thead_th]:tracking-wider'

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
      ancho: 'w-[8%]',
      render: (historia) => (
        <span className="font-mono text-xs font-semibold text-ink-soft">{historia.codigo}</span>
      ),
    },
    {
      clave: 'titulo',
      encabezado: 'Título',
      ancho: 'w-[23%]',
      render: (historia) => (
        <span className="block w-full min-w-0 truncate font-medium text-ink" title={historia.titulo}>
          {historia.titulo}
        </span>
      ),
    },
    {
      clave: 'estado',
      encabezado: 'Estado',
      ancho: 'w-[17%]',
      render: (historia) => (
        <SelectPill
          label={`Estado de ${historia.codigo}`}
          value={historia.estado}
          opciones={ESTADOS_HISTORIA}
          onChange={(estado) => onCambiar(historia, { estado })}
          disabled={guardandoId === historia.id}
          className={`${claseEstado(historia.estado)} ${SELECT_PILL}`}
        />
      ),
    },
    {
      clave: 'prioridad',
      encabezado: 'Prioridad',
      ancho: 'w-[17%]',
      render: (historia) => (
        <SelectPill
          label={`Prioridad de ${historia.codigo}`}
          value={historia.prioridad}
          opciones={PRIORIDADES_HISTORIA}
          onChange={(prioridad) => onCambiar(historia, { prioridad })}
          disabled={guardandoId === historia.id}
          className={`${clasePrioridad(historia.prioridad)} ${SELECT_PILL}`}
        />
      ),
    },
    {
      clave: 'estimacion',
      encabezado: 'Estimación',
      ancho: 'w-[12%]',
      alineacion: 'centro',
      render: (historia) => {
        if (pokerActivo.has(historia.id)) {
          return (
            <span
              className="inline-flex items-center gap-1.5 rounded-pill border border-[var(--primary-300)] bg-[var(--primary-100)] px-2 py-0.5 text-[11px] font-semibold text-primary"
              title="Sesión de Planning Poker en curso"
            >
              <i className="bi bi-broadcast text-[10px]" aria-hidden="true" />
              En votación
            </span>
          )
        }
        if (historia.story_points === null) {
          return <span className="text-xs font-medium text-ink-soft">Sin estimar</span>
        }
        return (
          <span
            className="inline-flex min-w-8 items-center justify-center rounded-pill border border-[var(--primary-300)] bg-[var(--primary-100)] px-2.5 py-0.5 text-xs font-semibold text-primary"
            title="Puntos de historia estimados"
          >
            {historia.story_points}
          </span>
        )
      },
    },
  ]

  return (
    <div className={CLASE_TABLA}>
      <TablaHistorias
        historias={historias}
        columnas={columnas}
        layoutFijo
        anchoAcciones="w-[23%]"
        acciones={(historia) => (
          <div className="flex items-center justify-end gap-1.5">
            {onPoker && (
              <button
                type="button"
                onClick={() => onPoker(historia)}
                disabled={guardandoId === historia.id}
                title="Iniciar / ver Planning Poker"
                aria-label={`Iniciar / ver Planning Poker de ${historia.codigo}`}
                className={BOTON_ACCION}
              >
                <i className="bi bi-suit-spade-fill" aria-hidden="true" />
                {pokerActivo.has(historia.id) ? 'Ver sala' : 'Estimar'}
              </button>
            )}
            <button
              type="button"
              onClick={() => onEditar(historia)}
              disabled={guardandoId === historia.id}
              title="Editar historia"
              aria-label={`Editar historia ${historia.codigo}`}
              className={BOTON_ACCION}
            >
              <i className="bi bi-pencil-square" aria-hidden="true" />
              Editar
            </button>
          </div>
        )}
      />
    </div>
  )
}
