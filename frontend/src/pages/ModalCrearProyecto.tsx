import { useEffect, useId, useRef, useState } from 'react'
import type { FormEvent } from 'react'
import { crearProyecto } from '../api/proyectos'
import { esApiError } from '../api/errors'
import type { Proyecto } from '../api/types'
import Alert from '../ui/Alert'
import Button from '../ui/Button'
import TextField from '../ui/TextField'
import { validarProyecto } from './validacionProyecto'
import type { CamposFormularioProyecto, ErroresFormularioProyecto } from './validacionProyecto'

type ModalCrearProyectoProps = {
  onCerrar: () => void
  onCreado: (proyecto: Proyecto) => void
}

const CAMPOS_FORMULARIO: ReadonlyArray<keyof CamposFormularioProyecto> = [
  'nombre',
  'fecha_inicio',
  'fecha_fin',
]

function esCampoFormulario(campo: string | undefined): campo is keyof CamposFormularioProyecto {
  return CAMPOS_FORMULARIO.some((valido) => valido === campo)
}

export default function ModalCrearProyecto({ onCerrar, onCreado }: ModalCrearProyectoProps) {
  const tituloId = useId()
  const [campos, setCampos] = useState<CamposFormularioProyecto>({
    nombre: '',
    fecha_inicio: '',
    fecha_fin: '',
  })
  const [errores, setErrores] = useState<ErroresFormularioProyecto>({})
  const [errorGeneral, setErrorGeneral] = useState<string | null>(null)
  const [enviando, setEnviando] = useState(false)
  const envioEnCurso = useRef(false)

  useEffect(() => {
    function alPresionarTecla(evento: KeyboardEvent) {
      if (evento.key === 'Escape' && !enviando) {
        onCerrar()
      }
    }

    document.addEventListener('keydown', alPresionarTecla)
    return () => document.removeEventListener('keydown', alPresionarTecla)
  }, [enviando, onCerrar])

  function cambiar(campo: keyof CamposFormularioProyecto, valor: string) {
    setCampos((anteriores) => ({ ...anteriores, [campo]: valor }))
  }

  function mostrarError(error: unknown) {
    if (esApiError(error)) {
      if (error.status === 0) {
        setErrorGeneral(
          'No se pudo conectar con el servidor. Tus datos se conservaron: intentá nuevamente.',
        )
        return
      }
      if (esCampoFormulario(error.campo)) {
        setErrores({ [error.campo]: error.message })
        return
      }
      setErrorGeneral(error.message)
      return
    }
    setErrorGeneral('Ocurrió un error inesperado. Intentá nuevamente.')
  }

  async function enviar(evento: FormEvent<HTMLFormElement>) {
    evento.preventDefault()
    if (envioEnCurso.current) {
      return
    }
    envioEnCurso.current = true
    setEnviando(true)
    setErrorGeneral(null)
    setErrores({})

    try {
      const erroresCliente = validarProyecto(campos)
      if (Object.keys(erroresCliente).length > 0) {
        setErrores(erroresCliente)
        return
      }

      try {
        const proyecto = await crearProyecto({
          nombre: campos.nombre.trim(),
          fecha_inicio: campos.fecha_inicio,
          fecha_fin: campos.fecha_fin === '' ? null : campos.fecha_fin,
        })
        onCreado(proyecto)
      } catch (error) {
        mostrarError(error)
      }
    } finally {
      envioEnCurso.current = false
      setEnviando(false)
    }
  }

  return (
    <div
      data-testid="fondo-modal"
      className="modal-fondo"
      onMouseDown={(evento) => {
        if (evento.target === evento.currentTarget && !enviando) {
          onCerrar()
        }
      }}
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby={tituloId}
        className="w-full max-w-lg modal-panel p-6"
      >
        <h2 id={tituloId} className="text-2xl font-semibold text-ink">
          Crear proyecto
        </h2>
        <p className="mt-1 text-sm text-ink-soft">
          Cargá el nombre y las fechas para empezar a planificar.
        </p>

        <form noValidate onSubmit={enviar} className="mt-5 flex flex-col gap-4">
          {errorGeneral && <Alert tono="error">{errorGeneral}</Alert>}
          <TextField
            label="Nombre"
            value={campos.nombre}
            onChange={(evento) => cambiar('nombre', evento.target.value)}
            error={errores.nombre}
            autoFocus
            autoComplete="off"
          />
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
          <div className="mt-2 flex flex-wrap justify-end gap-3">
            <Button variante="ghost" onClick={onCerrar} disabled={enviando}>
              Cancelar
            </Button>
            <Button type="submit" disabled={enviando}>
              {enviando ? 'Guardando…' : 'Guardar'}
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}
