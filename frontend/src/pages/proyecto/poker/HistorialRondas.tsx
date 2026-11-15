import { useId } from 'react'
import type { HistorialRondaPoker } from '../../../api/types'
import PanelDispersion from './PanelDispersion'

type HistorialRondasProps = {
  rondas: HistorialRondaPoker[]
  rondaActual: number
}

const formatoNumero = new Intl.NumberFormat('es-AR', {
  maximumFractionDigits: 2,
})

function formatearFecha(fecha: string | null): string {
  if (!fecha) {
    return ''
  }
  const [anio, mes, dia] = fecha.slice(0, 10).split('-')
  return `${dia}/${mes}/${anio}`
}

function etiquetaPuntos(puntos: number): string {
  const valor = formatoNumero.format(puntos)
  return `${valor} ${puntos === 1 ? 'punto' : 'puntos'}`
}

function Ronda({ ronda }: { ronda: HistorialRondaPoker }) {
  const tituloId = useId()
  const fecha = formatearFecha(ronda.revelada_en)

  return (
    <article aria-labelledby={tituloId} className="rounded-tile border border-border bg-surface p-4">
      <div className="mb-3 flex flex-wrap items-baseline justify-between gap-2">
        <h3 id={tituloId} className="text-base font-semibold text-ink">
          Ronda {ronda.numero}
        </h3>
        {fecha && <p className="text-xs text-ink-soft">Revelada el {fecha}</p>}
      </div>

      <ul aria-label={`Votos de la ronda ${ronda.numero}`} className="mb-4 space-y-1 text-sm text-ink">
        {ronda.votos.map((voto) => (
          <li key={voto.usuario_id}>
            {voto.usuario_nombre}: {etiquetaPuntos(voto.puntos)}
          </li>
        ))}
      </ul>

      {ronda.dispersion && <PanelDispersion dispersion={ronda.dispersion} />}
    </article>
  )
}

export default function HistorialRondas({ rondas, rondaActual }: HistorialRondasProps) {
  const anteriores = rondas
    .filter((ronda) => ronda.numero < rondaActual)
    .sort((a, b) => a.numero - b.numero)

  if (anteriores.length === 0) {
    return null
  }

  return (
    <section aria-label="Historial de rondas" className="mt-5">
      <h2 className="mb-3 text-lg font-semibold text-ink">Historial de rondas</h2>
      <div className="space-y-3">
        {anteriores.map((ronda) => (
          <Ronda key={ronda.numero} ronda={ronda} />
        ))}
      </div>
    </section>
  )
}
