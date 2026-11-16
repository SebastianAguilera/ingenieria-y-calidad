import { useId } from 'react'

type TarjetaKpiProps = {
  titulo: string
  icono: string
  /** Valor ya formateado; `null` significa "sin datos" (RN-02). */
  valor: string | null
  /** Por qué no hay datos; se muestra solo cuando `valor` es `null`. */
  explicacionSinDatos?: string
  detalle?: string
  alerta?: boolean
}

export const TEXTO_SIN_DATOS = 'Sin datos'

/** Tarjeta de KPI del sistema visual `teal-dashboard-ui`. */
export default function TarjetaKpi({
  titulo,
  icono,
  valor,
  explicacionSinDatos,
  detalle,
  alerta = false,
}: TarjetaKpiProps) {
  const idTitulo = useId()
  const sinDatos = valor === null

  return (
    <div
      role="group"
      aria-labelledby={idTitulo}
      className={`flex items-start gap-3 rounded-tile border p-4 ${
        alerta
          ? 'border-[color-mix(in_oklab,var(--accent-500)_50%,transparent)] bg-accent-100'
          : 'border-border bg-surface'
      }`}
    >
      <span
        aria-hidden="true"
        className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-tile text-lg ${
          alerta ? 'bg-[var(--accent-300)] text-accent-600' : 'bg-primary-100 text-primary'
        }`}
      >
        <i className={`bi ${icono}`} />
      </span>
      <div className="min-w-0">
        <p id={idTitulo} className="text-xs font-semibold text-ink-soft">
          {titulo}
        </p>
        {sinDatos ? (
          <>
            <p className="font-display text-xl font-semibold text-ink-soft">{TEXTO_SIN_DATOS}</p>
            {explicacionSinDatos && (
              <p className="mt-0.5 text-xs text-ink-soft">{explicacionSinDatos}</p>
            )}
          </>
        ) : (
          <>
            <p className="font-display text-2xl font-semibold text-ink">{valor}</p>
            {detalle && <p className="mt-0.5 text-xs text-ink-soft">{detalle}</p>}
          </>
        )}
      </div>
    </div>
  )
}
