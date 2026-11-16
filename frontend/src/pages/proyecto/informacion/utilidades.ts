import type { UsuarioProyecto } from '../../../api/types'

const PATRON_FECHA = /^(\d{4})-(\d{2})-(\d{2})/

function partesDeFecha(fecha: string): [number, number, number] | null {
  const coincide = PATRON_FECHA.exec(fecha)
  if (!coincide) {
    return null
  }
  return [Number(coincide[1]), Number(coincide[2]), Number(coincide[3])]
}

/** dd/mm/aaaa a partir de una fecha ISO. Se lee el texto para evitar corrimientos de zona horaria. */
export function formatearFecha(fecha: string | null): string {
  if (!fecha) {
    return 'Sin fecha'
  }
  const partes = PATRON_FECHA.exec(fecha)
  return partes ? `${partes[3]}/${partes[2]}/${partes[1]}` : 'Sin fecha'
}

export function rangoDeFechas(inicio: string, fin: string | null): string {
  if (fin === null) {
    return `Desde ${formatearFecha(inicio)}, sin fecha de fin`
  }
  return `${formatearFecha(inicio)} → ${formatearFecha(fin)}`
}

/** Días entre inicio y fin, contando ambos extremos. Null si no hay fin o el rango es inválido. */
export function duracionEnDias(inicio: string, fin: string | null): number | null {
  if (fin === null) {
    return null
  }
  const desde = partesDeFecha(inicio)
  const hasta = partesDeFecha(fin)
  if (!desde || !hasta) {
    return null
  }
  const msPorDia = 24 * 60 * 60 * 1000
  const inicioUTC = Date.UTC(desde[0], desde[1] - 1, desde[2])
  const finUTC = Date.UTC(hasta[0], hasta[1] - 1, hasta[2])
  if (finUTC < inicioUTC) {
    return null
  }
  return Math.round((finUTC - inicioUTC) / msPorDia) + 1
}

/** Líderes primero y después el resto, cada grupo por nombre. No modifica el original. */
export function ordenarEquipo(miembros: UsuarioProyecto[]): UsuarioProyecto[] {
  return [...miembros].sort((a, b) => {
    const pesoA = a.rol === 'LIDER' ? 0 : 1
    const pesoB = b.rol === 'LIDER' ? 0 : 1
    if (pesoA !== pesoB) {
      return pesoA - pesoB
    }
    return a.nombre.localeCompare(b.nombre, 'es')
  })
}

/** Hasta dos iniciales en mayúsculas a partir del nombre. */
export function iniciales(nombre: string): string {
  return nombre
    .trim()
    .split(/\s+/)
    .slice(0, 2)
    .map((palabra) => palabra.charAt(0).toUpperCase())
    .join('')
}
