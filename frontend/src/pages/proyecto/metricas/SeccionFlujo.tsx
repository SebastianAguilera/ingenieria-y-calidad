import type { MetricasProyecto } from '../../../api/types'
import Card from '../../../ui/Card'
import { formatearNumero, formatearSprints } from './formato'
import GraficoSeveridad from './GraficoSeveridad'
import TarjetaKpi from './TarjetaKpi'

const NOTA_LEAD_TIME =
  'Estimación según la Ley de Little (WiP / tasa de entrega), no una medición.'
const SIN_VELOCIDAD = 'Hace falta la velocidad de al menos un sprint cerrado.'
const SIN_TASA_ENTREGA = 'Hace falta la tasa de entrega de al menos un sprint cerrado.'

/** Aplica `formatear` solo si hay valor: `null` sigue siendo `null` (RN-02). */
function siHay(valor: number | null, formatear: (valor: number) => string): string | null {
  return valor === null ? null : formatear(valor)
}

/**
 * Indicadores de flujo y desglose de defectos por severidad (US-F22).
 * Cada tarjeta se oculta si la API no devuelve su campo (RB-09): `undefined`
 * (backend sin US-21) es distinto de `null` ("Sin datos").
 */
export default function SeccionFlujo({ metricas }: { metricas: MetricasProyecto }) {
  const {
    sprints_restantes: sprintsRestantes,
    wip,
    tasa_entrega: tasaEntrega,
    lead_time_estimado: leadTime,
    lead_time_aproximado: aproximado,
    defectos_abiertos: defectosAbiertos,
    defectos_por_severidad: porSeveridad,
  } = metricas

  const hayTarjetas = [sprintsRestantes, wip, tasaEntrega, leadTime, defectosAbiertos].some(
    (valor) => valor !== undefined,
  )
  if (!hayTarjetas && porSeveridad === undefined) {
    return null
  }

  const textoLeadTime = leadTime === undefined ? null : siHay(leadTime, formatearSprints)

  return (
    <div className="flex flex-col gap-5">
      {hayTarjetas && (
        <Card title="Flujo y defectos abiertos">
          <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
            {sprintsRestantes !== undefined && (
              <TarjetaKpi
                titulo="Sprints restantes"
                icono="bi-calendar-check"
                valor={siHay(sprintsRestantes, formatearSprints)}
                detalle="Para completar los story points pendientes"
                explicacionSinDatos={SIN_VELOCIDAD}
              />
            )}
            {wip !== undefined && (
              <TarjetaKpi
                titulo="WiP"
                icono="bi-hammer"
                valor={String(wip)}
                detalle="Historias en progreso"
              />
            )}
            {tasaEntrega !== undefined && (
              <TarjetaKpi
                titulo="Tasa de entrega"
                icono="bi-box-arrow-right"
                valor={siHay(tasaEntrega, formatearNumero)}
                detalle="Historias completadas por sprint cerrado"
                explicacionSinDatos={SIN_TASA_ENTREGA}
              />
            )}
            {leadTime !== undefined && (
              <TarjetaKpi
                titulo="Lead Time estimado"
                icono="bi-clock-history"
                valor={textoLeadTime}
                etiqueta={aproximado === true ? 'aprox.' : undefined}
                detalle={NOTA_LEAD_TIME}
                explicacionSinDatos={SIN_TASA_ENTREGA}
              />
            )}
            {defectosAbiertos !== undefined && (
              <TarjetaKpi
                titulo="Defectos abiertos"
                icono="bi-exclamation-triangle"
                valor={String(defectosAbiertos)}
                detalle="Defectos en estado Abierto"
              />
            )}
          </div>
        </Card>
      )}

      {porSeveridad !== undefined && <GraficoSeveridad defectos={porSeveridad} />}
    </div>
  )
}
