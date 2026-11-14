import type { DispersionPoker } from '../../../api/types'

type PanelDispersionProps = {
  dispersion: DispersionPoker
}

const formatoNumero = new Intl.NumberFormat('es-AR', {
  maximumFractionDigits: 2,
})

function formatear(valor: number): string {
  return formatoNumero.format(valor)
}

function etiquetaPuntos(valor: number): string {
  return `${formatear(valor)} ${valor === 1 ? 'punto' : 'puntos'}`
}

function etiquetaVotos(cantidad: number): string {
  return `${cantidad} ${cantidad === 1 ? 'voto' : 'votos'}`
}

function Distribucion({ votosPorValor }: { votosPorValor: DispersionPoker['votos_por_valor'] }) {
  const filas = Object.entries(votosPorValor)
    .map(([valor, cantidad]) => ({ valor: Number(valor), cantidad }))
    .sort((a, b) => a.valor - b.valor)

  return (
    <div>
      <h4 className="mb-2 text-sm font-semibold text-ink">Distribución de votos</h4>
      <ul aria-label="Distribución de votos" className="space-y-1 text-sm text-ink-soft">
        {filas.map(({ valor, cantidad }) => (
          <li key={valor}>
            {etiquetaPuntos(valor)}: {etiquetaVotos(cantidad)}
          </li>
        ))}
      </ul>
    </div>
  )
}

export default function PanelDispersion({ dispersion }: PanelDispersionProps) {
  return (
    <section
      aria-label="Dispersión de votos"
      className="rounded-tile border border-border bg-[var(--surface-2)] p-4"
    >
      <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
        <h3 className="text-base font-semibold text-ink">Dispersión de votos</h3>
        <span
          className={`inline-flex items-center rounded-pill px-3 py-1 text-xs font-semibold ${
            dispersion.hay_dispersion
              ? 'bg-[var(--accent-100)] text-[var(--warning)]'
              : 'bg-[var(--primary-100)] text-primary'
          }`}
        >
          {dispersion.hay_dispersion ? 'Hay diferencias' : 'Consenso'}
        </span>
      </div>

      <div className="grid gap-x-6 gap-y-1 text-sm text-ink-soft sm:grid-cols-2">
        <p>Promedio: {formatear(dispersion.promedio)}</p>
        <p>Mediana: {formatear(dispersion.mediana)}</p>
        <p>Mínimo: {formatear(dispersion.minimo)}</p>
        <p>Máximo: {formatear(dispersion.maximo)}</p>
        <p>Rango: {formatear(dispersion.rango)}</p>
        <p>Desviación estándar: {formatear(dispersion.desviacion_estandar)}</p>
        <p>Total de votos: {dispersion.total_votos}</p>
      </div>

      <div className="mt-4 border-t border-border pt-3">
        <Distribucion votosPorValor={dispersion.votos_por_valor} />
      </div>
    </section>
  )
}
