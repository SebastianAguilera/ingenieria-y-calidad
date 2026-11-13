export const LIMITE_IDENTIFICADOR_SPRINT = 50
export const LIMITE_SPRINT_GOAL = 300

export type CamposSprint = {
  identificador: string
  sprint_goal: string
  fecha_inicio: string
  fecha_fin: string
}

export type ErroresSprint = Partial<Record<keyof CamposSprint, string>>

export function validarSprint(campos: CamposSprint): ErroresSprint {
  const errores: ErroresSprint = {}
  const identificador = campos.identificador.trim()
  const sprintGoal = campos.sprint_goal.trim()

  if (identificador === '') {
    errores.identificador = 'El identificador es obligatorio'
  } else if ([...identificador].length > LIMITE_IDENTIFICADOR_SPRINT) {
    errores.identificador = `El identificador no puede superar los ${LIMITE_IDENTIFICADOR_SPRINT} caracteres`
  }

  if (sprintGoal === '') {
    errores.sprint_goal = 'El Sprint Goal es obligatorio'
  } else if ([...sprintGoal].length > LIMITE_SPRINT_GOAL) {
    errores.sprint_goal = `El Sprint Goal no puede superar los ${LIMITE_SPRINT_GOAL} caracteres`
  }

  if (campos.fecha_inicio === '') {
    errores.fecha_inicio = 'La fecha de inicio es obligatoria'
  }

  if (campos.fecha_fin === '') {
    errores.fecha_fin = 'La fecha de fin es obligatoria'
  } else if (campos.fecha_inicio !== '' && campos.fecha_fin < campos.fecha_inicio) {
    errores.fecha_fin = 'La fecha de fin debe ser igual o posterior a la de inicio'
  }

  return errores
}
