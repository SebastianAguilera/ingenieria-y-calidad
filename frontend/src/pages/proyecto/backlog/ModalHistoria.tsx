import { useEffect, useId, useState } from 'react'
import type { FormEvent } from 'react'
import type { EstadoHistoria, Historia, PrioridadHistoria } from '../../../api/types'
import Alert from '../../../ui/Alert'
import Button from '../../../ui/Button'
import TextField from '../../../ui/TextField'
import {
  CAMPOS_PRIORIZACION,
  ESTADOS_HISTORIA,
  PRIORIDADES_HISTORIA,
  VALORES_PRIORIZACION,
} from './etiquetas'
import type { DatosHistoria, ResultadoOperacion } from './tipos'

type ErroresFormulario = Partial<Record<keyof DatosHistoria, string>>

type ModalHistoriaProps = {
  titulo: string
  textoBoton: string
  historia?: Historia
  onGuardar: (datos: DatosHistoria) => Promise<ResultadoOperacion>
  onCerrar: () => void
}

function validar(campos: DatosHistoria): ErroresFormulario {
  const errores: ErroresFormulario = {}
  const titulo = campos.titulo.trim()

  if (titulo === '') {
    errores.titulo = 'El título de la historia es obligatorio'
  } else if (titulo.length > 200) {
    errores.titulo = 'El título de la historia no puede superar los 200 caracteres'
  }

  for (const campo of CAMPOS_PRIORIZACION) {
    const valor = campos[campo.clave]
    if (valor !== null && !VALORES_PRIORIZACION.includes(valor)) {
      errores[campo.clave] = `El valor de ${campo.etiqueta.toLowerCase()} no es válido`
    }
  }

  return errores
}

function esCampo(valor: string | undefined): valor is keyof DatosHistoria {
  return (
    valor === 'titulo' ||
    valor === 'descripcion' ||
    valor === 'prioridad' ||
    valor === 'estado' ||
    valor === 'valor_negocio' ||
    valor === 'criticidad_tiempo' ||
    valor === 'reduccion_riesgo'
  )
}

export default function ModalHistoria({
  titulo,
  textoBoton,
  historia,
  onGuardar,
  onCerrar,
}: ModalHistoriaProps) {
  const idDescripcion = useId()
  const idBase = useId()
  const [campos, setCampos] = useState<DatosHistoria>(() => ({
    titulo: historia?.titulo ?? '',
    descripcion: historia?.descripcion ?? '',
    prioridad: historia?.prioridad ?? 'Media',
    estado: historia?.estado ?? 'Backlog',
    valor_negocio: historia?.valor_negocio ?? null,
    criticidad_tiempo: historia?.criticidad_tiempo ?? null,
    reduccion_riesgo: historia?.reduccion_riesgo ?? null,
  }))
  const [errores, setErrores] = useState<ErroresFormulario>({})
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

  function cambiar<K extends keyof DatosHistoria>(campo: K, valor: DatosHistoria[K]) {
    setCampos((anteriores) => ({ ...anteriores, [campo]: valor }))
  }

  async function enviar(evento: FormEvent<HTMLFormElement>) {
    evento.preventDefault()
    if (enviando) {
      return
    }

    const erroresCliente = validar(campos)
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
        titulo: campos.titulo.trim(),
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
        aria-labelledby="modal-historia-titulo"
        className="w-full max-w-lg rounded-card border border-border bg-surface p-6 shadow-card"
      >
        <h2 id="modal-historia-titulo" className="text-2xl font-semibold text-ink">
          {titulo}
        </h2>

        <form noValidate onSubmit={enviar} className="mt-5 flex flex-col gap-4">
          {errorGeneral && <Alert tono="error">{errorGeneral}</Alert>}

          <TextField
            label="Título"
            value={campos.titulo}
            onChange={(evento) => cambiar('titulo', evento.target.value)}
            error={errores.titulo}
            maxLength={201}
            autoFocus
            autoComplete="off"
          />

          <div className="flex flex-col gap-1.5">
            <label htmlFor={idDescripcion} className="text-sm font-medium text-ink">
              Descripción
            </label>
            <textarea
              id={idDescripcion}
              value={campos.descripcion}
              onChange={(evento) => cambiar('descripcion', evento.target.value)}
              className="min-h-24 rounded-tile border border-border bg-[var(--surface-2)] px-3.5 py-2.5 text-sm text-ink placeholder:text-ink-soft focus:border-primary focus:outline-none"
            />
            {errores.descripcion && (
              <p className="text-xs text-[var(--danger)]">{errores.descripcion}</p>
            )}
          </div>

          <div className="grid gap-4 sm:grid-cols-2">
            <div className="flex flex-col gap-1.5">
              <label htmlFor="historia-prioridad" className="text-sm font-medium text-ink">
                Prioridad
              </label>
              <select
                id="historia-prioridad"
                value={campos.prioridad}
                onChange={(evento) =>
                  cambiar('prioridad', evento.target.value as PrioridadHistoria)
                }
                className="rounded-tile border border-border bg-[var(--surface-2)] px-3.5 py-2.5 text-sm text-ink focus:border-primary focus:outline-none"
              >
                {PRIORIDADES_HISTORIA.map((prioridad) => (
                  <option key={prioridad} value={prioridad}>
                    {prioridad}
                  </option>
                ))}
              </select>
              {errores.prioridad && (
                <p className="text-xs text-[var(--danger)]">{errores.prioridad}</p>
              )}
            </div>

            <div className="flex flex-col gap-1.5">
              <label htmlFor="historia-estado" className="text-sm font-medium text-ink">
                Estado
              </label>
              <select
                id="historia-estado"
                value={campos.estado}
                onChange={(evento) => cambiar('estado', evento.target.value as EstadoHistoria)}
                className="rounded-tile border border-border bg-[var(--surface-2)] px-3.5 py-2.5 text-sm text-ink focus:border-primary focus:outline-none"
              >
                {ESTADOS_HISTORIA.map((estado) => (
                  <option key={estado} value={estado}>
                    {estado}
                  </option>
                ))}
              </select>
              {errores.estado && <p className="text-xs text-[var(--danger)]">{errores.estado}</p>}
            </div>
          </div>

          <div className="grid gap-4 sm:grid-cols-3">
            {CAMPOS_PRIORIZACION.map((campo) => {
              const valor = campos[campo.clave]
              const idCampo = `historia-${campo.clave}`
              const idAyuda = `${idBase}-${campo.clave}`

              return (
                <div key={campo.clave} className="flex flex-col gap-1.5">
                  <label htmlFor={idCampo} className="text-sm font-medium text-ink">
                    {campo.etiqueta}
                  </label>
                  <select
                    id={idCampo}
                    aria-describedby={idAyuda}
                    value={valor === null ? '' : String(valor)}
                    onChange={(evento) => {
                      const seleccionado = evento.target.value
                      cambiar(
                        campo.clave,
                        seleccionado === '' ? null : Number(seleccionado),
                      )
                    }}
                    className="rounded-tile border border-border bg-[var(--surface-2)] px-3.5 py-2.5 text-sm text-ink focus:border-primary focus:outline-none"
                  >
                    <option value="" disabled={historia?.[campo.clave] != null}>
                      Sin asignar
                    </option>
                    {VALORES_PRIORIZACION.map((opcion) => (
                      <option key={opcion} value={String(opcion)}>
                        {opcion}
                      </option>
                    ))}
                  </select>
                  <p id={idAyuda} className="text-xs text-ink-soft">
                    {campo.ayuda}
                  </p>
                  {errores[campo.clave] && (
                    <p className="text-xs text-[var(--danger)]">{errores[campo.clave]}</p>
                  )}
                </div>
              )
            })}
          </div>

          <div className="mt-2 flex flex-wrap justify-end gap-3">
            <Button variante="ghost" onClick={onCerrar} disabled={enviando}>
              Cancelar
            </Button>
            <Button type="submit" disabled={enviando}>
              {enviando ? 'Guardando…' : textoBoton}
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}
