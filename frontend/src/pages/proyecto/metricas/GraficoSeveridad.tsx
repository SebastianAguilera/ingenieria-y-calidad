import { useId } from 'react'
import { Bar, BarChart, CartesianGrid, Cell, LabelList, Tooltip, XAxis, YAxis } from 'recharts'
import type { DefectosPorSeveridad } from '../../../api/types'
import { colorSeveridad } from '../defectos/etiquetasDefecto'
import { formatearNumero } from './formato'
import Leyenda from './Leyenda'
import { desgloseSeveridad, resumenSeveridad } from './severidad'
import { ESTILO_TOOLTIP, TICK_EJE } from './estilosGrafico'
import TarjetaGrafico from './TarjetaGrafico'
import { useAnchoContenedor } from './useAnchoContenedor'

const TITULO = 'Defectos por severidad'
const ALTO = 220

/**
 * Barras horizontales de defectos por severidad (US-F22). Las cuatro
 * severidades se muestran siempre, también `blocker` con `0` (CA-06).
 */
export default function GraficoSeveridad({
  defectos,
  className,
}: {
  defectos: DefectosPorSeveridad
  className?: string
}) {
  const { ref, ancho } = useAnchoContenedor<HTMLDivElement>()
  const desglose = desgloseSeveridad(defectos)
  const prefijoDegrade = useId().replace(/:/g, '')

  return (
    <TarjetaGrafico
      titulo={TITULO}
      descripcion="Todos los defectos del proyecto, de menor a mayor severidad"
      className={className}
      tabla={{
        columnas: ['Severidad', 'Defectos'],
        filas: desglose.map(({ etiqueta, valor }) => [etiqueta, formatearNumero(valor)]),
      }}
    >
      <p className="sr-only">{resumenSeveridad(desglose)}</p>
      <Leyenda
        nombre="Leyenda de defectos por severidad"
        className="mb-3"
        elementos={desglose.map(({ clave, etiqueta, valor }) => ({
          etiqueta,
          color: colorSeveridad(clave),
          valor: formatearNumero(valor),
        }))}
      />
      <div ref={ref} aria-hidden="true" className="w-full" style={{ height: ALTO }}>
        {ancho > 0 && (
          <BarChart
            layout="vertical"
            width={ancho}
            height={ALTO}
            data={desglose}
            margin={{ top: 4, right: 32, bottom: 0, left: 0 }}
          >
            <defs>
              {desglose.map(({ clave }) => (
                <linearGradient
                  key={clave}
                  id={`${prefijoDegrade}-${clave}`}
                  x1="0"
                  y1="0"
                  x2="1"
                  y2="0"
                >
                  <stop offset="0%" stopColor={colorSeveridad(clave)} stopOpacity={0.2} />
                  <stop offset="100%" stopColor={colorSeveridad(clave)} stopOpacity={0.9} />
                </linearGradient>
              ))}
            </defs>
            <CartesianGrid
              horizontal={false}
              strokeDasharray="3 3"
              stroke="var(--chart-grid)"
              strokeOpacity={0.4}
            />
            <XAxis
              type="number"
              allowDecimals={false}
              axisLine={false}
              tickLine={false}
              tick={TICK_EJE}
            />
            <YAxis
              type="category"
              dataKey="etiqueta"
              axisLine={false}
              tickLine={false}
              tick={TICK_EJE}
              width={80}
            />
            <Tooltip
              cursor={{ fill: 'var(--primary-100)' }}
              contentStyle={ESTILO_TOOLTIP}
              formatter={(valor) => formatearNumero(Number(valor))}
            />
            <Bar
              dataKey="valor"
              name="Defectos"
              radius={[0, 6, 6, 0]}
              maxBarSize={28}
              isAnimationActive={false}
            >
              {desglose.map(({ clave }) => (
                <Cell
                  key={clave}
                  fill={`url(#${prefijoDegrade}-${clave})`}
                  stroke={colorSeveridad(clave)}
                  strokeWidth={1.5}
                />
              ))}
              <LabelList dataKey="valor" position="right" style={TICK_EJE} />
            </Bar>
          </BarChart>
        )}
      </div>
    </TarjetaGrafico>
  )
}
