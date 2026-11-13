import { useEffect, useId, useState } from 'react'
import type { FormEvent } from 'react'
import Alert from '../../../ui/Alert'
import Button from '../../../ui/Button'
import TextField from '../../../ui/TextField'
import type { ResultadoOperacion } from './tipos'
import { validarSprint } from './validacionSprint'
import type { CamposSprint, ErroresSprint } from './validacionSprint'

type ModalSprintProps = {
  onGuardar: (campos: CamposSprint) => Promise<ResultadoOperacion>
  onCerrar: () => void
}

const CAMPOS: ReadonlyArray<keyof CamposSprint> = [
  'identificador',
  'sprint_goal',
  'fecha_inicio',
  'fecha_fin',
]

function esCampo(valor: string | undefined): valor is keyof CamposSprint {
  return CAMPOS.some((campo) => campo === valor)
}

export default function ModalSprint({ onGuardar, onCerrar }: ModalSprintProps) {
  const idGoal = useId()
  const [campos, setCampos] = useState<CamposSprint>({
    identificador: '',
    sprint_goal: '',
    fecha_inicio: '',
    fecha_fin: '',
  })
  const [errores, setErrores] = useState<ErroresSprint>({})
  const [errorGeneral, setErrorGeneral] = useState<string | null>(null)
  const [enviando, setEnviando] = useState(false)

  useEffect(() => {
    function alPresionarTecla(evento: KeyboardEvent) {
      if (evento.key === 'Escape') {
        onCerrar()
      }
    }

    document.addEventListener('keydown', alPresionarTecla)
    return () => document.removeEventListener('keydown', alPresionarTecla)
  }, [onCerrar])

  function cambiar(campo: keyof CamposSprint, valor: string) {
    setCampos((anteriores) => ({ ...anteriores, [campo]: valor }))
  }

  async function enviar(evento: FormEvent<HTMLFormElement>) {
    evento.preventDefault()
    if (enviando) {
      return
    }

    const erroresCliente = validarSprint(campos)
    if (Object.keys(erroresCliente).length > 0) {
      setErrores(erroresCliente)
      return
    }

    setEnviando(true)
    setErrorGeneral(null)
    setErrores({})

    try {
      const resultado = await onGuardar({
        ...campos,
        identificador: campos.identificador.trim(),
        sprint_goal: campos.sprint_goal.trim(),
      })

      if (!resultado.ok) {
        if (esCampo(resultado.campo)) {
          setErrores({ [resultado.campo]: resultado.mensaje })
        } else {
          setErrorGeneral(resultado.mensaje)
        }
      }
    } finally {
      setEnviando(false)
    }
  }

  return (
    <div
      data-testid="fondo-modal"
      className="fixed inset-0 z-50 flex items-center justify-center bg-[color-mix(in_oklab,var(--ink)_45%,transparent)] p-4"
      onMouseDown={(evento) => {
        if (evento.target === evento.currentTarget) {
          onCerrar()
        }
      }}
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="modal-sprint-titulo"
        className="w-full max-w-lg rounded-card border border-border bg-surface p-6 shadow-card"
      >
        <h2 id="modal-sprint-titulo" className="text-2xl font-semibold text-ink">
          Nuevo sprint
        </h2>

        <form noValidate onSubmit={enviar} className="mt-5 flex flex-col gap-4">
          {errorGeneral && <Alert tono="error">{errorGeneral}</Alert>}

          <TextField
            label="Identificador"
            value={campos.identificador}
            onChange={(evento) => cambiar('identificador', evento.target.value)}
            error={errores.identificador}
            maxLength={51}
            autoFocus
            autoComplete="off"
          />

          <div className="flex flex-col gap-1.5">
            <label htmlFor={idGoal} className="text-sm font-medium text-ink">
              Sprint Goal
            </label>
            <textarea
              id={idGoal}
              value={campos.sprint_goal}
              onChange={(evento) => cambiar('sprint_goal', evento.target.value)}
              aria-invalid={errores.sprint_goal ? true : undefined}
              aria-describedby={errores.sprint_goal ? `${idGoal}-error` : undefined}
              className="min-h-24 rounded-tile border border-border bg-[var(--surface-2)] px-3.5 py-2.5 text-sm text-ink placeholder:text-ink-soft focus:border-primary focus:outline-none"
            />
            {errores.sprint_goal && (
              <p id={`${idGoal}-error`} className="text-xs text-[var(--danger)]">
                {errores.sprint_goal}
              </p>
            )}
          </div>

          <div className="grid gap-4 sm:grid-cols-2">
            <TextField
              label="Fecha de inicio"
              type="date"
              value={campos.fecha_inicio}
              onChange={(evento) => cambiar('fecha_inicio', evento.target.value)}
              error={errores.fecha_inicio}
            />
            <TextField
              label="Fecha de fin"
              type="date"
              value={campos.fecha_fin}
              onChange={(evento) => cambiar('fecha_fin', evento.target.value)}
              error={errores.fecha_fin}
            />
          </div>

          <div className="mt-2 flex flex-wrap justify-end gap-3">
            <Button variante="ghost" onClick={onCerrar} disabled={enviando}>
              Cancelar
            </Button>
            <Button type="submit" disabled={enviando}>
              {enviando ? 'Guardando…' : 'Crear sprint'}
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}
