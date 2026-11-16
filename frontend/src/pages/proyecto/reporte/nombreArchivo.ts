/**
 * Nombres de los archivos descargados del reporte (US-F23, D-02 y D-03).
 * Siguen el esquema de `pdf_service.go` (US-24). El frontend arma el nombre
 * porque, en producción, el CORS no expone `Content-Disposition`.
 */

const LARGO_MAXIMO_SLUG = 50
const SLUG_VACIO = 'reporte'

/** Slug ASCII: minúsculas, sin tildes, palabras separadas por un guion. */
export function slug(texto: string): string {
  const resultado = texto
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .slice(0, LARGO_MAXIMO_SLUG)
    .replace(/-+$/, '')
  return resultado === '' ? SLUG_VACIO : resultado
}

/** Día local en formato `AAAA-MM-DD`. */
export function fechaLocal(fecha: Date): string {
  const mes = String(fecha.getMonth() + 1).padStart(2, '0')
  const dia = String(fecha.getDate()).padStart(2, '0')
  return `${fecha.getFullYear()}-${mes}-${dia}`
}

export type ExtensionReporte = 'json' | 'pdf'

/** `reporte-{nombre}-proyecto-{fecha}.{ext}` */
export function nombreArchivoProyecto(
  nombreProyecto: string,
  fecha: string,
  extension: ExtensionReporte,
): string {
  return `reporte-${slug(nombreProyecto)}-proyecto-${fecha}.${extension}`
}

/** `reporte-proyecto-{id}-sprint-{sprint}-{fecha}.{ext}`: el reporte de sprint no trae el nombre del proyecto. */
export function nombreArchivoSprint(
  proyectoId: number,
  identificadorSprint: string,
  fecha: string,
  extension: ExtensionReporte,
): string {
  return `reporte-proyecto-${proyectoId}-sprint-${slug(identificadorSprint)}-${fecha}.${extension}`
}
