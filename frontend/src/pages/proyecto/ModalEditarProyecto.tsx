import { useEffect, useId, useState } from 'react'
import type { FormEvent } from 'react'
import { actualizarProyecto, cambiarEstadoProyecto } from '../../api/proyectos'
import { esApiError } from '../../api/errors'
import type { EstadoProyecto, Proyecto } from '../../api/types'
import Alert from '../../ui/Alert'
import Button from '../../ui/Button'
import TextField from '../../ui/TextField'
import { validarDescripcionProyecto, validarProyecto } from '../validacionProyecto'
import type { CamposFormularioProyecto } from '../validacionProyecto'

type Formulario = CamposFormularioProyecto & {
  descripcion: string
  estado: EstadoProyecto
}

type ErroresFormulario = Partial<Record<keyof Formulario, string>>

type ModalEditarProyectoProps = {
  proyecto: Proyecto
  onCerrar: () => void
  onGuardado: (proyecto: Proyecto) => void
}

const ESTADOS: EstadoProyecto[] = ['Activo', 'Cerrado']
const CAMPOS: ReadonlyArray<keyof Formulario> = [
  'nombre',
  'descripcion',
  'fecha_inicio',
  'fecha_fin',
  'estado',
]

function fechaISO(fecha: string | null): string {
  return fecha ? fecha.slice(0, 10) : ''
}

function desdeProyecto(proyecto: Proyecto): Formulario {
  return {
    nombre: proyecto.nombre,
    descripcion: proyecto.descripcion ?? '',
    fecha_inicio: fechaISO(proyecto.fecha_inicio),
    fecha_fin: fechaISO(proyecto.fecha_fin),
    estado: proyecto.estado,
  }
}

function esCampo(valor: string | undefined): valor is keyof Formulario {
  return CAMPOS.some((campo) => campo === valor)
}

export default function ModalEditarProyecto({
  proyecto,
  onCerrar,
  onGuardado,
}: ModalEditarProyectoProps) {
  const tituloId = useId()
  const idDescripcion = useId()
  const [proyectoActual, setProyectoActual] = useState(proyecto)
  const [formulario, setFormulario] = useState<Formulario>(() => desdeProyecto(proyecto))
  const [errores, setErrores] = useState<ErroresFormulario>({})
  const [errorGuardado, setErrorGuardado] = useState<string | null>(null)
  const [exito, setExito] = useState<string | null>(null)
  const [guardando, setGuardando] = useState(false)
  const [reabriendo, setReabriendo] = useState(false)
  useEffect(() => {
    function alPresionarTecla(evento: KeyboardEvent) {
      if (evento.key === 'Escape' && !guardando && !reabriendo) {
        onCerrar()
      }
    }

    document.addEventListener('keydown', alPresionarTecla)
    return () => document.removeEventListener('keydown', alPresionarTecla)
  }, [guardando, onCerrar, reabriendo])

  function cambiar<K extends keyof Formulario>(campo: K, valor: Formulario[K]) {
    setFormulario((anterior) => ({ ...anterior, [campo]: valor }))
  }

  function mostrarErrorGuardado(error: unknown) {
    if (esApiError(error)) {
      if (error.status === 0) {
        setErrorGuardado(
          'No se pudo conectar con el servidor. Tus datos se conservaron: intentá nuevamente.',
        )
        return
      }
      if (error.status === 404) {
        setErrorGuardado('No tenés acceso a este proyecto o no existe')
        return
      }
      if (esCampo(error.campo)) {
        setErrores({ [error.campo]: error.message })
        return
      }
      setErrorGuardado(error.message)
      return
    }
    setErrorGuardado('Ocurrió un error inesperado. Intentá nuevamente.')
  }

  function mismosDatos(actual: Proyecto): boolean {
    return (
      formulario.nombre.trim() === actual.nombre &&
      formulario.descripcion.trim() === (actual.descripcion ?? '') &&
      formulario.fecha_inicio === fechaISO(actual.fecha_inicio) &&
      formulario.fecha_fin === fechaISO(actual.fecha_fin)
    )
  }

  function validarFormulario(): boolean {
    const erroresCliente: ErroresFormulario = validarProyecto(formulario)
    const errorDescripcion = validarDescripcionProyecto(formulario.descripcion)
    if (errorDescripcion) {
      erroresCliente.descripcion = errorDescripcion
    }
    setErrores(erroresCliente)
    return Object.keys(erroresCliente).length === 0
  }

  async function guardar(evento: FormEvent<HTMLFormElement>) {
    evento.preventDefault()
    if (guardando) {
      return
    }

    setErrorGuardado(null)
    setExito(null)
    if (!validarFormulario()) {
      return
    }

    const datosCambiaron = !mismosDatos(proyectoActual)
    const cambiaEstado = formulario.estado !== proyectoActual.estado

    if (!datosCambiaron && !cambiaEstado) {
      setExito('No hay cambios para guardar.')
      return
    }

    setGuardando(true)
    try {
      let actual = proyectoActual
      if (datosCambiaron) {
        actual = await actualizarProyecto(proyectoActual.id, {
          nombre: formulario.nombre.trim(),
          descripcion: formulario.descripcion.trim(),
          fecha_inicio: formulario.fecha_inicio,
          fecha_fin: formulario.fecha_fin === '' ? null : formulario.fecha_fin,
        })
        setProyectoActual(actual)
        onGuardado(actual)
      }
      if (cambiaEstado) {
        actual = await cambiarEstadoProyecto(proyectoActual.id, formulario.estado)
        setProyectoActual(actual)
        onGuardado(actual)
      }
      setFormulario(desdeProyecto(actual))
      setExito('Los cambios se guardaron correctamente.')
    } catch (error) {
      mostrarErrorGuardado(error)
    } finally {
      setGuardando(false)
    }
  }

  async function reabrir() {
    if (reabriendo) {
      return
    }
    setReabriendo(true)
    setErrorGuardado(null)
    setExito(null)
    try {
      const reabierto = await cambiarEstadoProyecto(proyectoActual.id, 'Activo')
      setProyectoActual(reabierto)
      setFormulario(desdeProyecto(reabierto))
      onGuardado(reabierto)
      setExito('Proyecto reabierto.')
    } catch (error) {
      mostrarErrorGuardado(error)
    } finally {
      setReabriendo(false)
    }
  }

  const cerrado = proyectoActual.estado === 'Cerrado'
  const ocupado = guardando || reabriendo

  return (
    <div
      data-testid="fondo-modal"
      className="modal-fondo"
      onMouseDown={(evento) => {
        if (evento.target === evento.currentTarget && !ocupado) {
          onCerrar()
        }
      }}
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby={tituloId}
        className="w-full max-w-2xl modal-panel max-h-[90vh] overflow-y-auto p-6"
      >
        <h2 id={tituloId} className="text-2xl font-semibold text-ink">
          Editar proyecto
        </h2>
        <p className="mt-1 text-sm text-ink-soft">
          Actualizá los datos de {proyecto.nombre}.
        </p>

        <form noValidate onSubmit={(evento) => void guardar(evento)} className="mt-5 flex flex-col gap-5">
          {exito && <Alert tono="success">{exito}</Alert>}
          {errorGuardado && <Alert tono="error">{errorGuardado}</Alert>}

          {cerrado && (
            <div className="flex flex-col items-start gap-3">
              <Alert tono="info">Reabrí el proyecto para editar sus datos</Alert>
              <Button variante="secondary" onClick={() => void reabrir()} disabled={ocupado}>
                {reabriendo ? 'Reabriendo…' : 'Reabrir proyecto'}
              </Button>
            </div>
          )}

          <TextField
            label="Nombre"
            value={formulario.nombre}
            onChange={(evento) => cambiar('nombre', evento.target.value)}
            error={errores.nombre}
            disabled={cerrado || guardando}
            autoComplete="off"
          />

          <div className="flex flex-col gap-1.5">
            <label htmlFor={idDescripcion} className="text-sm font-medium text-ink">
              Descripción
            </label>
            <textarea
              id={idDescripcion}
              value={formulario.descripcion}
              onChange={(evento) => cambiar('descripcion', evento.target.value)}
              disabled={cerrado || guardando}
              aria-invalid={errores.descripcion ? true : undefined}
              aria-describedby={errores.descripcion ? `${idDescripcion}-error` : undefined}
              className="min-h-24 rounded-tile border border-border bg-[var(--surface-2)] px-3.5 py-2.5 text-sm text-ink placeholder:text-ink-soft focus:border-primary focus:outline-none disabled:opacity-60"
            />
            {errores.descripcion && (
              <p id={`${idDescripcion}-error`} className="text-xs text-[var(--danger)]">
                {errores.descripcion}
              </p>
            )}
          </div>

          <div className="grid gap-4 sm:grid-cols-2">
            <TextField
              label="Fecha de inicio"
              type="date"
              value={formulario.fecha_inicio}
              onChange={(evento) => cambiar('fecha_inicio', evento.target.value)}
              error={errores.fecha_inicio}
              disabled={cerrado || guardando}
            />
            <TextField
              label="Fecha de fin"
              type="date"
              value={formulario.fecha_fin}
              onChange={(evento) => cambiar('fecha_fin', evento.target.value)}
              error={errores.fecha_fin}
              disabled={cerrado || guardando}
            />
          </div>

          <div className="flex flex-col gap-1.5">
            <label htmlFor="proyecto-estado" className="text-sm font-medium text-ink">
              Estado
            </label>
            <select
              id="proyecto-estado"
              value={formulario.estado}
              onChange={(evento) => cambiar('estado', evento.target.value as EstadoProyecto)}
              disabled={cerrado || guardando}
              className="rounded-tile border border-border bg-[var(--surface-2)] px-3.5 py-2.5 text-sm text-ink focus:border-primary focus:outline-none disabled:opacity-60"
            >
              {ESTADOS.map((valor) => (
                <option key={valor} value={valor}>
                  {valor}
                </option>
              ))}
            </select>
            {errores.estado && <p className="text-xs text-[var(--danger)]">{errores.estado}</p>}
          </div>

          <div className="flex flex-wrap items-center justify-end gap-3 border-t border-border pt-4">
            <Button variante="ghost" onClick={onCerrar} disabled={ocupado}>
              Cancelar
            </Button>
            <Button type="submit" disabled={cerrado || guardando}>
              {guardando ? 'Guardando…' : 'Guardar cambios'}
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}
