import type { MetricasProyecto } from '../../../api/types'
import Card from '../../../ui/Card'
import { formatearDesviacion, formatearHoras, formatearNumero, formatearPorcentaje } from './formato'
import TarjetaKpi from './TarjetaKpi'

/** Aplica `formatear` solo si hay valor: `null` sigue siendo `null` (RN-02). */
function siHay(valor: number | null, formatear: (valor: number) => string): string | null {
  return valor === null ? null : formatear(valor)
}

/** Tarjetas de KPI del proyecto (US-F21); también las usa el reporte (US-F23). */
export default function ResumenKpi({ metricas }: { metricas: MetricasProyecto }) {
  const ratio = metricas.ratio_defectos

  return (
    <Card title="Resumen">
      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
        <TarjetaKpi
          titulo="Velocidad"
          icono="bi-speedometer2"
          valor={siHay(metricas.velocidad, (valor) => `${formatearNumero(valor)} SP`)}
          detalle="Promedio por sprint cerrado"
          explicacionSinDatos="Todavía no hay sprints cerrados"
        />
        <TarjetaKpi
          titulo="Historias completadas"
          icono="bi-check2-circle"
          valor={siHay(metricas.porcentaje_historias_completadas, formatearPorcentaje)}
          detalle="Historias en Hecho sobre el total"
          explicacionSinDatos="El proyecto no tiene historias"
        />
        <TarjetaKpi
          titulo="Desviación de esfuerzo"
          icono="bi-hourglass-split"
          valor={siHay(metricas.desviacion_esfuerzo, formatearDesviacion)}
          detalle={`${formatearHoras(metricas.horas_reales)} reales / ${formatearHoras(metricas.horas_estimadas)} estimadas`}
          explicacionSinDatos="Ninguna historia tiene horas estimadas y reales"
        />
        <TarjetaKpi
          titulo="Ratio de defectos"
          icono="bi-bug"
          valor={siHay(ratio, formatearNumero)}
          detalle="Defectos por historia completada"
          explicacionSinDatos="Todavía no hay historias completadas"
          alerta={ratio !== null && ratio > 0}
        />
        <TarjetaKpi titulo="Defectos detectados" icono="bi-exclamation-octagon" valor={String(metricas.defectos_detectados)} />
        <TarjetaKpi titulo="Defectos resueltos" icono="bi-check2-all" valor={String(metricas.defectos_resueltos)} />
      </div>
    </Card>
  )
}
