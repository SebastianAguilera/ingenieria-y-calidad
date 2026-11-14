import type { DragEvent } from 'react'
import type { EstadoHistoria, Historia } from '../../../api/types'
import Button from '../../../ui/Button'
import EtiquetaPrioridad from './EtiquetaPrioridad'
import { ESTADOS_HISTORIA, claseEstado } from './etiquetas'
import SelectPill from './SelectPill'

type TableroBacklogProps = {
  historias: Historia[]
  onMover: (historia: Historia, estado: EstadoHistoria) => void
  onEditar: (historia: Historia) => void
  onPoker?: (historia: Historia) => void
  pokerActivo?: ReadonlySet<number>
  guardandoId: number | null
}

const SIN_POKER_ACTIVO: ReadonlySet<number> = new Set()

function alArrastrar(evento: DragEvent<HTMLElement>, historia: Historia) {
  evento.dataTransfer.setData('text/plain', String(historia.id))
  evento.dataTransfer.effectAllowed = 'move'
}

export default function TableroBacklog({
  historias,
  onMover,
  onEditar,
  onPoker,
  pokerActivo = SIN_POKER_ACTIVO,
  guardandoId,
}: TableroBacklogProps) {
  function alSoltar(evento: DragEvent<HTMLElement>, estado: EstadoHistoria) {
    evento.preventDefault()
    const id = Number(evento.dataTransfer.getData('text/plain'))
    const historia = historias.find((actual) => actual.id === id)
    if (historia && historia.estado !== estado) {
      onMover(historia, estado)
    }
  }

  return (
    <div className="grid gap-4 md:grid-cols-3">
      {ESTADOS_HISTORIA.map((estado) => {
        const tarjetas = historias.filter((historia) => historia.estado === estado)
        return (
          <section
            key={estado}
            aria-label={`Columna ${estado}`}
            onDragOver={(evento) => evento.preventDefault()}
            onDrop={(evento) => alSoltar(evento, estado)}
            className="flex min-h-[16rem] flex-col rounded-card border border-border bg-surface p-4"
          >
            <header className="mb-4 flex items-center justify-between gap-2">
              <h2 className="text-base font-semibold text-ink">{estado}</h2>
              <span className="rounded-pill bg-[var(--primary-100)] px-2.5 py-0.5 text-xs font-semibold text-primary">
                {tarjetas.length}
              </span>
            </header>

            <div className="flex flex-1 flex-col gap-3">
              {tarjetas.length === 0 && (
                <p className="rounded-tile border border-dashed border-border px-3 py-4 text-center text-xs text-ink-soft">
                  Sin historias
                </p>
              )}

              {tarjetas.map((historia) => {
                const guardando = guardandoId === historia.id
                return (
                  <article
                    key={historia.id}
                    aria-label={`Historia ${historia.codigo}`}
                    draggable
                    onDragStart={(evento) => alArrastrar(evento, historia)}
                    className="cursor-grab rounded-tile border border-border bg-[var(--surface-2)] p-3 shadow-card active:cursor-grabbing"
                  >
                    <div className="flex items-start justify-between gap-2">
                      <span className="font-mono text-xs font-semibold text-ink-soft">
                        {historia.codigo}
                      </span>
                      <EtiquetaPrioridad prioridad={historia.prioridad} />
                    </div>
                    <p className="mt-1 font-medium text-ink">{historia.titulo}</p>
                    {pokerActivo.has(historia.id) && (
                      <span className="mt-2 inline-flex items-center rounded-pill border border-[var(--primary-300)] bg-[var(--primary-100)] px-2.5 py-1 text-xs font-semibold text-primary">
                        Planning Poker en curso
                      </span>
                    )}
                    <div className="mt-3 flex flex-wrap items-center justify-between gap-2">
                      <SelectPill
                        label={`Mover ${historia.codigo} a`}
                        value={historia.estado}
                        opciones={ESTADOS_HISTORIA}
                        onChange={(nuevoEstado) => {
                          if (nuevoEstado !== historia.estado) {
                            onMover(historia, nuevoEstado)
                          }
                        }}
                        disabled={guardando}
                        className={claseEstado(historia.estado)}
                      />
                      <div className="flex flex-wrap items-center gap-1">
                        {onPoker && (
                          <Button
                            variante="ghost"
                            onClick={() => onPoker(historia)}
                            aria-label={`Planning Poker de ${historia.codigo}`}
                            disabled={guardando}
                          >
                            Planning Poker
                          </Button>
                        )}
                        <Button
                          variante="ghost"
                          onClick={() => onEditar(historia)}
                          aria-label={`Editar ${historia.codigo}`}
                          disabled={guardando}
                        >
                          Editar
                        </Button>
                      </div>
                    </div>
                  </article>
                )
              })}
            </div>
          </section>
        )
      })}
    </div>
  )
}
