import type { SeveridadDefecto } from '../../../api/types'

export type CamposAltaDefecto = {
  descripcion: string
  severidad: SeveridadDefecto | ''
  historia_id: string
  sprint_deteccion_id: string
}

export type ErroresAltaDefecto = Partial<Record<keyof CamposAltaDefecto, string>>

/**
 * Valida el alta en el cliente. El alta no limita el largo de la descripción
 * porque el backend solo lo hace al editar (US-18).
 */
export function validarAltaDefecto(campos: CamposAltaDefecto): ErroresAltaDefecto {
  const errores: ErroresAltaDefecto = {}

  if (campos.descripcion.trim() === '') {
    errores.descripcion = 'La descripción es obligatoria'
  }

  if (campos.severidad === '') {
    errores.severidad = 'Elegí una severidad'
  }

  if (campos.historia_id === '') {
    errores.historia_id = 'Elegí una historia'
  }

  if (campos.sprint_deteccion_id === '') {
    errores.sprint_deteccion_id = 'Elegí el sprint de detección'
  }

  return errores
}

/** Máximo de caracteres de la descripción al editar (fijado por US-18 en el backend). */
export const LIMITE_DESCRIPCION_EDICION = 1000

export type CamposEdicionDefecto = {
  descripcion: string
  severidad: SeveridadDefecto
}

export type ErroresEdicionDefecto = Partial<Record<keyof CamposEdicionDefecto, string>>

/**
 * Valida la edición en el cliente. El largo se cuenta en caracteres (runas),
 * no en bytes ni en unidades UTF-16, igual que el backend.
 */
export function validarEdicionDefecto(campos: CamposEdicionDefecto): ErroresEdicionDefecto {
  const errores: ErroresEdicionDefecto = {}
  const descripcion = campos.descripcion.trim()

  if (descripcion === '') {
    errores.descripcion = 'La descripción es obligatoria'
  } else if ([...descripcion].length > LIMITE_DESCRIPCION_EDICION) {
    errores.descripcion = `La descripción no puede superar los ${LIMITE_DESCRIPCION_EDICION} caracteres`
  }

  return errores
}
