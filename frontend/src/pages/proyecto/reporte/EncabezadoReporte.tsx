import Card from '../../../ui/Card'
import { claseEstadoSprint } from '../sprints/etiquetasSprint'
import type { Reporte } from './useReporte'

const CLASE_PILL = 'inline-flex items-center rounded-pill border px-3 py-1 text-xs font-semibold'

/** `2026-09-15` → `15/09/2026`. */
function formatearFecha(fecha: string): string {
  const [anio, mes, dia] = fecha.slice(0, 10).split('-')
  return `${dia}/${mes}/${anio}`
}

function Dato({ nombre, valor }: { nombre: string; valor: string }) {
  return (
    <div>
      <dt className="text-xs font-semibold text-ink-soft">{nombre}</dt>
      <dd className="mt-0.5 text-sm text-ink">{valor}</dd>
    </div>
  )
}

/** Encabezado del reporte: datos del proyecto o del sprint. */
export default function EncabezadoReporte({ reporte }: { reporte: Reporte }) {
  const esProyecto = reporte.alcance === 'proyecto'
  const titulo = esProyecto ? reporte.datos.nombre : reporte.datos.identificador
  const detalle = esProyecto ? reporte.datos.descripcion : reporte.datos.sprint_goal
  const fechaFin =
    reporte.datos.fecha_fin === null ? 'Sin fecha de fin' : formatearFecha(reporte.datos.fecha_fin)
  const claseEstado =
    reporte.alcance === 'sprint'
      ? claseEstadoSprint(reporte.datos.estado)
      : reporte.datos.estado === 'Activo'
        ? 'border-transparent bg-[var(--success)] text-white'
        : 'border-transparent bg-ink-soft text-[var(--on-primary)]'

  return (
    <Card>
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="text-xs font-semibold uppercase tracking-wide text-ink-soft">
            {esProyecto ? 'Proyecto' : 'Sprint'}
          </p>
          <h2 className="text-2xl font-bold text-ink [overflow-wrap:anywhere]">{titulo}</h2>
          {detalle && <p className="mt-1 text-sm text-ink-soft">{detalle}</p>}
        </div>
        <span className={`${CLASE_PILL} ${claseEstado}`}>{reporte.datos.estado}</span>
      </div>
      <dl className="mt-4 grid gap-4 sm:grid-cols-2">
        <Dato nombre="Fecha de inicio" valor={formatearFecha(reporte.datos.fecha_inicio)} />
        <Dato nombre="Fecha de fin" valor={fechaFin} />
      </dl>
    </Card>
  )
}
