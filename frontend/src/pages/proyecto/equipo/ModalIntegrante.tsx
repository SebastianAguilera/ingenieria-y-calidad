import { useEffect, useId, useState } from 'react'
import type { FormEvent } from 'react'
import type { RolProyecto } from '../../../api/types'
import Alert from '../../../ui/Alert'
import Button from '../../../ui/Button'
import TextField from '../../../ui/TextField'
import { ROLES } from './etiquetasRol'
import type { ResultadoOperacion } from './tipos'

type ModalIntegranteProps = {
  onGuardar: (email: string, rol: RolProyecto) => Promise<ResultadoOperacion>
  onCerrar: () => void
}

type Errores = { email?: string; rol?: string }

const PATRON_EMAIL = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

export default function ModalIntegrante({ onGuardar, onCerrar }: ModalIntegranteProps) {
  const idRol = useId()
  const [email, setEmail] = useState('')
  const [rol, setRol] = useState<RolProyecto>('MIEMBRO')
  const [errores, setErrores] = useState<Errores>({})
  const [errorGeneral, setErrorGeneral] = useState<string | null>(null)
  const [enviando, setEnviando] = useState(false)

  useEffect(() => {
    function alPresionarTecla(evento: KeyboardEvent) {
      if (evento.key === 'Escape' && !enviando) {
        onCerrar()
      }
    }

    document.addEventListener('keydown', alPresionarTecla)
    return () => document.removeEventListener('keydown', alPresionarTecla)
  }, [enviando, onCerrar])

  async function enviar(evento: FormEvent<HTMLFormElement>) {
    evento.preventDefault()
    if (enviando) {
      return
    }

    const emailLimpio = email.trim()
    if (emailLimpio === '') {
      setErrores({ email: 'El email es obligatorio' })
      return
    }

    if (!PATRON_EMAIL.test(emailLimpio)) {
      setErrores({ email: 'Ingresá un email válido' })
      return
    }

    setEnviando(true)
    setErrores({})
    setErrorGeneral(null)

    try {
      const resultado = await onGuardar(emailLimpio, rol)
      if (!resultado.ok) {
        if (resultado.campo === 'email' || resultado.campo === 'rol') {
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
        if (evento.target === evento.currentTarget && !enviando) {
          onCerrar()
        }
      }}
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="modal-integrante-titulo"
        className="w-full max-w-lg rounded-card border border-border bg-surface p-6 shadow-card"
      >
        <h2 id="modal-integrante-titulo" className="text-2xl font-semibold text-ink">
          Agregar integrante
        </h2>

        <form noValidate onSubmit={enviar} className="mt-5 flex flex-col gap-4">
          {errorGeneral && <Alert tono="error">{errorGeneral}</Alert>}

          <TextField
            label="Email"
            type="email"
            value={email}
            onChange={(evento) => setEmail(evento.target.value)}
            error={errores.email}
            autoFocus
            autoComplete="off"
          />

          <div className="flex flex-col gap-1.5">
            <label htmlFor={idRol} className="text-sm font-medium text-ink">
              Rol
            </label>
            <select
              id={idRol}
              value={rol}
              onChange={(evento) => setRol(evento.target.value as RolProyecto)}
              aria-invalid={errores.rol ? true : undefined}
              aria-describedby={errores.rol ? `${idRol}-error` : undefined}
              className={`rounded-tile border bg-[var(--surface-2)] px-3.5 py-2.5 text-sm text-ink focus:border-primary focus:outline-none ${
                errores.rol ? 'border-[var(--danger)]' : 'border-border'
              }`}
            >
              {ROLES.map((opcion) => (
                <option key={opcion} value={opcion}>
                  {opcion}
                </option>
              ))}
            </select>
            {errores.rol && (
              <p id={`${idRol}-error`} className="text-xs text-[var(--danger)]">
                {errores.rol}
              </p>
            )}
          </div>

          <div className="mt-2 flex flex-wrap justify-end gap-3">
            <Button variante="ghost" onClick={onCerrar} disabled={enviando}>
              Cancelar
            </Button>
            <Button type="submit" disabled={enviando}>
              {enviando ? 'Agregando…' : 'Agregar'}
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}
