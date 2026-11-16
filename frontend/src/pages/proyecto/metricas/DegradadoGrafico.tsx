/**
 * Degradado de transparencia para series de Recharts: el color de la serie
 * (un token de `theme.css`) pasa de casi opaco arriba a muy translúcido abajo.
 * Se declara dentro de un `<defs>` del gráfico y se usa con `fill="url(#id)"`.
 */
export default function DegradadoGrafico({
  id,
  color,
  arriba = 0.9,
  abajo = 0.12,
}: {
  id: string
  color: string
  arriba?: number
  abajo?: number
}) {
  return (
    <linearGradient id={id} x1="0" y1="0" x2="0" y2="1">
      <stop offset="0%" style={{ stopColor: color, stopOpacity: arriba }} />
      <stop offset="100%" style={{ stopColor: color, stopOpacity: abajo }} />
    </linearGradient>
  )
}
