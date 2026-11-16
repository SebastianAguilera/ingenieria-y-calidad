import type { ReporteDefecto, ReporteHistoria, Sprint } from '../../../api/types'
import Card from '../../../ui/Card'
import { claseEstado } from '../backlog/etiquetas'
import {
  claseEstadoDefecto,
  claseSeveridadDefecto,
  etiquetaEstado,
  etiquetaSeveridad,
  truncarDescripcion,
} from '../defectos/etiquetasDefecto'
import { formatearHoras, formatearNumero } from '../metricas/formato'

/** Valor ausente en una tabla (RN-02). */
export const SIN_VALOR = '—'

const CLASE_TH = 'px-4 py-3 font-semibold'
const CLASE_TD = 'px-4 py-3'
const CLASE_PILL = 'inline-flex items-center rounded-pill border px-3 py-1 text-xs font-semibold'

function siHay(valor: number | null, formatear: (valor: number) => string): string {
  return valor === null ? SIN_VALOR : formatear(valor)
}

function Vacio({ children }: { children: string }) {
  return <p className="text-sm text-ink-soft">{children}</p>
}

function cantidad(total: number, singular: string, plural: string): string {
  return `${total} ${total === 1 ? singular : plural}`
}

/** Historias del reporte (US-F23). */
export function TablaHistoriasReporte({ historias }: { historias: ReporteHistoria[] }) {
  return (
    <Card title="Historias" description={cantidad(historias.length, 'historia', 'historias')}>
      {historias.length === 0 ? (
        <Vacio>Este reporte no tiene historias</Vacio>
      ) : (
        <div className="overflow-x-auto">
          <table aria-label="Historias del reporte" className="w-full min-w-[40rem] text-left text-sm">
            <thead className="border-b border-border text-ink-soft">
              <tr>
                <th scope="col" className={CLASE_TH}>Código</th>
                <th scope="col" className={CLASE_TH}>Título</th>
                <th scope="col" className={CLASE_TH}>Estado</th>
                <th scope="col" className={`${CLASE_TH} text-right`}>SP</th>
                <th scope="col" className={`${CLASE_TH} text-right`}>Horas estimadas</th>
                <th scope="col" className={`${CLASE_TH} text-right`}>Horas reales</th>
              </tr>
            </thead>
            <tbody>
              {historias.map((historia) => (
                <tr key={historia.id} className="border-b border-border last:border-b-0">
                  <td className={`${CLASE_TD} whitespace-nowrap font-medium text-ink`}>{historia.codigo}</td>
                  <td className={`${CLASE_TD} text-ink [overflow-wrap:anywhere]`}>{historia.titulo}</td>
                  <td className={CLASE_TD}>
                    <span className={`${CLASE_PILL} whitespace-nowrap ${claseEstado(historia.estado)}`}>
                      {historia.estado}
                    </span>
                  </td>
                  <td className={`${CLASE_TD} text-right`}>{siHay(historia.story_points, formatearNumero)}</td>
                  <td className={`${CLASE_TD} whitespace-nowrap text-right`}>
                    {siHay(historia.horas_estimadas, formatearHoras)}
                  </td>
                  <td className={`${CLASE_TD} whitespace-nowrap text-right`}>
                    {siHay(historia.horas_reales, formatearHoras)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </Card>
  )
}

/** Defectos del reporte, con los sprints por nombre (US-F23, CA-10). */
export function TablaDefectosReporte({
  defectos,
  sprints,
}: {
  defectos: ReporteDefecto[]
  sprints: Sprint[]
}) {
  const nombres = new Map(sprints.map((sprint) => [sprint.id, sprint.identificador]))
  const nombreSprint = (id: number | null) => (id === null ? SIN_VALOR : (nombres.get(id) ?? SIN_VALOR))

  return (
    <Card title="Defectos" description={cantidad(defectos.length, 'defecto', 'defectos')}>
      {defectos.length === 0 ? (
        <Vacio>Este reporte no tiene defectos</Vacio>
      ) : (
        <div className="overflow-x-auto">
          <table aria-label="Defectos del reporte" className="w-full min-w-[48rem] text-left text-sm">
            <thead className="border-b border-border text-ink-soft">
              <tr>
                <th scope="col" className={CLASE_TH}>Código</th>
                <th scope="col" className={CLASE_TH}>Descripción</th>
                <th scope="col" className={CLASE_TH}>Severidad</th>
                <th scope="col" className={CLASE_TH}>Estado</th>
                <th scope="col" className={CLASE_TH}>Sprint de detección</th>
                <th scope="col" className={CLASE_TH}>Sprint de resolución</th>
              </tr>
            </thead>
            <tbody>
              {defectos.map((defecto) => (
                <tr key={defecto.id} className="border-b border-border last:border-b-0">
                  <td className={`${CLASE_TD} whitespace-nowrap font-medium text-ink`}>{defecto.codigo}</td>
                  <td className={CLASE_TD}>
                    <span title={defecto.descripcion} className="block max-w-sm [overflow-wrap:anywhere]">
                      {truncarDescripcion(defecto.descripcion)}
                    </span>
                  </td>
                  <td className={CLASE_TD}>
                    <span className={`${CLASE_PILL} ${claseSeveridadDefecto(defecto.severidad)}`}>
                      {etiquetaSeveridad(defecto.severidad)}
                    </span>
                  </td>
                  <td className={CLASE_TD}>
                    <span className={`${CLASE_PILL} whitespace-nowrap ${claseEstadoDefecto(defecto.estado)}`}>
                      {etiquetaEstado(defecto.estado)}
                    </span>
                  </td>
                  <td className={CLASE_TD}>{nombreSprint(defecto.sprint_deteccion_id)}</td>
                  <td className={CLASE_TD}>{nombreSprint(defecto.sprint_resolucion_id)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </Card>
  )
}
