import { Cell, Pie, PieChart, Tooltip } from 'recharts'
import type { EstadoHistoria, Historia } from '../../../api/types'
import { contarPorEstado } from './distribucion'
import DegradadoGrafico from './DegradadoGrafico'
import { ESTILO_TOOLTIP } from './estilosGrafico'
import { formatearPorcentaje } from './formato'
import Leyenda from './Leyenda'
import TarjetaGrafico from './TarjetaGrafico'

const TITULO = 'Historias por estado'
const LADO = 200

const COLOR_ESTADO: Record<EstadoHistoria, string> = {
  Backlog: 'var(--chart-3)',
  'En Progreso': 'var(--chart-2)',
  Hecho: 'var(--chart-1)',
}

const ESTADOS = Object.keys(COLOR_ESTADO) as EstadoHistoria[]

/** Dona de historias por estado, contada del backlog cargado (CA-5, RN-01). */
export default function GraficoEstados({
  historias,
  className,
}: {
  historias: Historia[]
  className?: string
}) {
  const distribucion = contarPorEstado(historias)
  const total = historias.length

  function porcentaje(cantidad: number): string {
    return total === 0 ? '' : ` (${formatearPorcentaje(cantidad / total)})`
  }

  return (
    <TarjetaGrafico
      titulo={TITULO}
      descripcion="Según el backlog actual"
      className={className}
      tabla={{
        columnas: ['Estado', 'Historias'],
        filas: distribucion.map(({ estado, cantidad }) => [estado, String(cantidad)]),
      }}
    >
      <div className="flex flex-col items-center gap-4 sm:flex-row sm:justify-center">
        <div aria-hidden="true" className="relative" style={{ width: LADO, height: LADO }}>
          <PieChart width={LADO} height={LADO}>
            <defs>
              {ESTADOS.map((estado, indice) => (
                <DegradadoGrafico
                  key={estado}
                  id={`degradado-estado-${indice}`}
                  color={COLOR_ESTADO[estado]}
                  arriba={0.95}
                  abajo={0.3}
                />
              ))}
            </defs>
            <Pie
              data={distribucion}
              dataKey="cantidad"
              nameKey="estado"
              innerRadius={62}
              outerRadius={92}
              paddingAngle={distribucion.filter((d) => d.cantidad > 0).length > 1 ? 2 : 0}
              stroke="var(--surface)"
              isAnimationActive={false}
            >
              {distribucion.map(({ estado }) => (
                <Cell key={estado} fill={`url(#degradado-estado-${ESTADOS.indexOf(estado)})`} />
              ))}
            </Pie>
            <Tooltip contentStyle={ESTILO_TOOLTIP} />
          </PieChart>
          <div className="pointer-events-none absolute inset-0 flex flex-col items-center justify-center">
            <span className="font-display text-2xl font-semibold text-ink">{total}</span>
            <span className="text-xs text-ink-soft">historias</span>
          </div>
        </div>
        <Leyenda
          nombre="Leyenda de historias por estado"
          className="flex-col"
          elementos={distribucion.map(({ estado, cantidad }) => ({
            etiqueta: estado,
            color: COLOR_ESTADO[estado],
            valor: `${cantidad}${porcentaje(cantidad)}`,
          }))}
        />
      </div>
    </TarjetaGrafico>
  )
}
