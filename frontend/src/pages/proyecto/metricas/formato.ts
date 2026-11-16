/**
 * Formato de las métricas de US-20 para la UI (US-F21, RN-03 a RN-05).
 * No calcula métricas: solo presenta los valores que devuelve la API.
 */

const LOCALE = 'es-AR'

/** Signo menos tipográfico, para que la desviación negativa se lea bien. */
const SIGNO_MENOS = '\u2212'

const formatoPorcentaje = new Intl.NumberFormat(LOCALE, { maximumFractionDigits: 1 })
const formatoNumero = new Intl.NumberFormat(LOCALE, { maximumFractionDigits: 2 })

/** Recibe un valor `0..1` y lo muestra como porcentaje: `0.75` → `75%`. */
export function formatearPorcentaje(valor: number): string {
  return `${formatoPorcentaje.format(valor * 100)}%`
}

/** Desviación con signo explícito: `0.2` → `+20%`, `-0.1` → `−10%`, `0` → `0%`. */
export function formatearDesviacion(valor: number): string {
  const porcentaje = formatearPorcentaje(Math.abs(valor))
  if (porcentaje === '0%') {
    return porcentaje
  }
  return `${valor > 0 ? '+' : SIGNO_MENOS}${porcentaje}`
}

/** Número con hasta 2 decimales y coma decimal: `0.3333` → `0,33`. */
export function formatearNumero(valor: number): string {
  return formatoNumero.format(valor)
}

/** Horas con unidad: `5` → `5 h`. */
export function formatearHoras(valor: number): string {
  return `${formatearNumero(valor)} h`
}
