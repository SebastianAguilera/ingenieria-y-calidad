import { useId } from 'react'
import type { MetricasSprint } from '../../../api/types'
import { formatearDesviacion, formatearHoras, formatearNumero, formatearPorcentaje } from './formato'
import { TEXTO_SIN_DATOS } from './TarjetaKpi'

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

/** Métricas de un sprint (US-F21); también las usa el reporte (US-F23). */
export default function DatosSprint({ metricas }: { metricas: MetricasSprint }) {
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
