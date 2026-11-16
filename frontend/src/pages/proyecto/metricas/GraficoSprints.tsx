import { Bar, BarChart, CartesianGrid, LabelList, Tooltip, XAxis, YAxis } from 'recharts'
import { formatearNumero } from './formato'
import Leyenda from './Leyenda'
import {
  COLOR_COMPLETADOS,
  COLOR_PLANIFICADOS,
  ESTILO_TOOLTIP,
  TICK_EJE,
} from './estilosGrafico'
import TarjetaGrafico from './TarjetaGrafico'
import { useAnchoContenedor } from './useAnchoContenedor'
import type { SprintConMetricas } from './useMetricas'

const TITULO = 'Story points por sprint'
const ALTO = 260

/** Barras de SP planificados contra completados por sprint cerrado y activo (CA-4). */
export default function GraficoSprints({
  sprints,
  className,
}: {
  sprints: SprintConMetricas[]
  className?: string
}) {
  const { ref, ancho } = useAnchoContenedor<HTMLDivElement>()
  const datos = sprints.map(({ sprint, metricas }) => ({
    sprint: sprint.identificador,
    planificados: metricas.story_points_planificados,
    completados: metricas.story_points_completados,
  }))
  const vacio = datos.length === 0

  return (
    <TarjetaGrafico
      titulo={TITULO}
      descripcion="Sprints cerrados y el sprint activo"
      className={className}
      tabla={
        vacio
          ? undefined
          : {
              columnas: ['Sprint', 'SP planificados', 'SP completados'],
              filas: datos.map((fila) => [
                fila.sprint,
                formatearNumero(fila.planificados),
                formatearNumero(fila.completados),
              ]),
            }
      }
    >
      {vacio ? (
        <p className="py-10 text-center text-sm text-ink-soft">
          Todavía no hay sprints cerrados ni activos
        </p>
      ) : (
        <>
          <Leyenda
            nombre="Leyenda de story points por sprint"
            className="mb-3"
            elementos={[
              { etiqueta: 'SP planificados', color: COLOR_PLANIFICADOS },
              { etiqueta: 'SP completados', color: COLOR_COMPLETADOS },
            ]}
          />
          <div ref={ref} aria-hidden="true" className="w-full" style={{ height: ALTO }}>
            {ancho > 0 && (
              <BarChart
                width={ancho}
                height={ALTO}
                data={datos}
                margin={{ top: 20, right: 8, bottom: 0, left: 0 }}
                barGap={4}
              >
                <CartesianGrid vertical={false} strokeDasharray="3 3" stroke="var(--chart-grid)" />
                <XAxis dataKey="sprint" axisLine={false} tickLine={false} tick={TICK_EJE} />
                <YAxis
                  allowDecimals={false}
                  axisLine={false}
                  tickLine={false}
                  tick={TICK_EJE}
                  width={32}
                />
                <Tooltip
                  cursor={{ fill: 'var(--primary-100)' }}
                  contentStyle={ESTILO_TOOLTIP}
                  formatter={(valor) => formatearNumero(Number(valor))}
                />
                <Bar
                  dataKey="planificados"
                  name="SP planificados"
                  fill={COLOR_PLANIFICADOS}
                  radius={[6, 6, 0, 0]}
                  maxBarSize={28}
                  isAnimationActive={false}
                >
                  <LabelList dataKey="planificados" position="top" style={TICK_EJE} />
                </Bar>
                <Bar
                  dataKey="completados"
                  name="SP completados"
                  fill={COLOR_COMPLETADOS}
                  radius={[6, 6, 0, 0]}
                  maxBarSize={28}
                  isAnimationActive={false}
                >
                  <LabelList dataKey="completados" position="top" style={TICK_EJE} />
                </Bar>
              </BarChart>
            )}
          </div>
        </>
      )}
    </TarjetaGrafico>
  )
}
