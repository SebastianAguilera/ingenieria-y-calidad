export const LIMITE_NOMBRE_PROYECTO = 150
export const LIMITE_DESCRIPCION_PROYECTO = 2000

export type CamposFormularioProyecto = {
  nombre: string
  fecha_inicio: string
  fecha_fin: string
}

export type ErroresFormularioProyecto = Partial<Record<keyof CamposFormularioProyecto, string>>

export function validarProyecto(campos: CamposFormularioProyecto): ErroresFormularioProyecto {
  const errores: ErroresFormularioProyecto = {}
  const nombre = campos.nombre.trim()

  if (nombre === '') {
    errores.nombre = 'El nombre es obligatorio'
  } else if ([...nombre].length > LIMITE_NOMBRE_PROYECTO) {
    errores.nombre = `El nombre no puede superar los ${LIMITE_NOMBRE_PROYECTO} caracteres`
  }

  if (campos.fecha_inicio === '') {
    errores.fecha_inicio = 'La fecha de inicio es obligatoria'
  }

  if (campos.fecha_fin !== '' && campos.fecha_inicio !== '' && campos.fecha_fin < campos.fecha_inicio) {
    errores.fecha_fin = 'La fecha de fin debe ser igual o posterior a la de inicio'
  }

  return errores
}

export function validarDescripcionProyecto(descripcion: string): string | undefined {
  if ([...descripcion.trim()].length > LIMITE_DESCRIPCION_PROYECTO) {
    return `La descripción no puede superar los ${LIMITE_DESCRIPCION_PROYECTO} caracteres`
  }
  return undefined
}
