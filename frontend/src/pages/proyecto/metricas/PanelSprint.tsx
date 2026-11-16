import { useId, useState } from 'react'
import type { MetricasSprint, Sprint } from '../../../api/types'
import Alert from '../../../ui/Alert'
import Button from '../../../ui/Button'
import { sprintInicial } from './distribucion'
import { formatearDesviacion, formatearHoras, formatearNumero, formatearPorcentaje } from './formato'
import { TEXTO_SIN_DATOS } from './TarjetaKpi'
import { useMetricasSprint } from './useMetricasSprint'

const CLASES_SELECT =
  'appearance-none rounded-pill border border-border bg-[var(--surface-2)] py-1.5 pl-3 pr-8 text-sm text-ink focus:border-primary focus:outline-none'

type DatoSprintProps = {
  nombre: string
  valor: string | null
  explicacionSinDatos?: string
}

function DatoSprint({ nombre, valor, explicacionSinDatos }: DatoSprintProps) {
  const idNombre = useId()
  return (
    <div role="group" aria-labelledby={idNombre} className="rounded-tile bg-[var(--surface-2)] p-3">
      <p id={idNombre} className="text-xs font-semibold text-ink-soft">
        {nombre}
      </p>
      {valor === null ? (
        <>
          <p className="font-display text-lg font-semibold text-ink-soft">{TEXTO_SIN_DATOS}</p>
          {explicacionSinDatos && <p className="text-xs text-ink-soft">{explicacionSinDatos}</p>}
        </>
      ) : (
        <p className="font-display text-lg font-semibold text-ink">{valor}</p>
      )}
    </div>
  )
}

function DatosSprint({ metricas }: { metricas: MetricasSprint }) {
  const porcentaje = metricas.porcentaje_completado
  const desviacion = metricas.desviacion_esfuerzo

  return (
    <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
      <DatoSprint nombre="SP planificados" valor={formatearNumero(metricas.story_points_planificados)} />
      <DatoSprint nombre="SP completados" valor={formatearNumero(metricas.story_points_completados)} />
      <DatoSprint
        nombre="Completado"
        valor={porcentaje === null ? null : formatearPorcentaje(porcentaje)}
        explicacionSinDatos="El sprint no tiene story points planificados"
      />
      <DatoSprint nombre="Historias hechas" valor={String(metricas.historias_completadas)} />
      <DatoSprint nombre="Horas estimadas" valor={formatearHoras(metricas.horas_estimadas)} />
      <DatoSprint nombre="Horas reales" valor={formatearHoras(metricas.horas_reales)} />
      <DatoSprint
        nombre="Desviación"
        valor={desviacion === null ? null : formatearDesviacion(desviacion)}
        explicacionSinDatos="Ninguna historia del sprint tiene horas estimadas y reales"
      />
      <DatoSprint
        nombre="Defectos del sprint"
        valor={`${metricas.defectos_detectados} detectados · ${metricas.defectos_resueltos} resueltos`}
      />
    </div>
  )
}

function ContenidoSprint({ proyectoId, sprintId }: { proyectoId: number; sprintId: number }) {
  const { estado, reintentar } = useMetricasSprint(proyectoId, sprintId)

  if (estado.tipo === 'cargando') {
    return (
      <p role="status" className="text-sm text-ink-soft">
        Cargando métricas del sprint…
      </p>
    )
  }

  if (estado.tipo === 'error') {
    return (
      <div className="flex flex-col items-start gap-3">
        <Alert tono="error">{estado.mensaje}</Alert>
        <Button variante="secondary" tamano="sm" onClick={reintentar}>
          Reintentar
        </Button>
      </div>
    )
  }

  return <DatosSprint metricas={estado.metricas} />
}

/** Selector de sprint y sus métricas (CA-6). */
export default function PanelSprint({ proyectoId, sprints }: { proyectoId: number; sprints: Sprint[] }) {
  const idTitulo = useId()
  const idSelector = useId()
  const [sprintId, setSprintId] = useState<number | null>(() => sprintInicial(sprints))

  return (
    <section aria-labelledby={idTitulo} className="rounded-card border border-border bg-surface p-5 shadow-card">
      <header className="mb-4 flex flex-wrap items-center justify-between gap-3">
        <h2 id={idTitulo} className="text-lg font-semibold text-ink">
          Detalle por sprint
        </h2>
        {sprintId !== null && (
          <div className="flex items-center gap-2 text-sm text-ink-soft">
            <label htmlFor={idSelector}>Sprint</label>
            <span className="relative inline-flex items-center">
              <select
                id={idSelector}
                value={String(sprintId)}
                onChange={(evento) => setSprintId(Number(evento.target.value))}
                className={CLASES_SELECT}
              >
                {sprints.map((sprint) => (
                  <option key={sprint.id} value={String(sprint.id)}>
                    {`${sprint.identificador} (${sprint.estado})`}
                  </option>
                ))}
              </select>
              <i
                className="bi bi-chevron-down pointer-events-none absolute right-2.5 text-xs text-ink-soft"
                aria-hidden="true"
              />
            </span>
          </div>
        )}
      </header>

      {sprintId === null ? (
        <p className="text-sm text-ink-soft">Todavía no hay sprints</p>
      ) : (
        <ContenidoSprint key={sprintId} proyectoId={proyectoId} sprintId={sprintId} />
      )}
    </section>
  )
}
