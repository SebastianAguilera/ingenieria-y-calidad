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
