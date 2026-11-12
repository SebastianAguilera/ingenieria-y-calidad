import { useEffect, useId, useState } from 'react'
import type { FormEvent } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import {
  actualizarProyecto,
  cambiarEstadoProyecto,
  eliminarProyecto,
  obtenerProyecto,
} from '../../api/proyectos'
import { esApiError } from '../../api/errors'
import type { EstadoProyecto, Proyecto } from '../../api/types'
import PageHeader from '../../components/PageHeader'
import Alert from '../../ui/Alert'
import Button from '../../ui/Button'
import Card from '../../ui/Card'
import ModalConfirmacion from '../../ui/ModalConfirmacion'
import TextField from '../../ui/TextField'
import {
  validarDescripcionProyecto,
  validarProyecto,
} from '../validacionProyecto'
import type { CamposFormularioProyecto } from '../validacionProyecto'

type Formulario = CamposFormularioProyecto & {
  descripcion: string
  estado: EstadoProyecto
}

type ErroresFormulario = Partial<Record<keyof Formulario, string>>

type EstadoCarga =
  | { tipo: 'cargando' }
  | { tipo: 'error'; es404: boolean }
  | { tipo: 'listo'; proyecto: Proyecto }

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

function formatearFecha(fecha: string | null): string {
  if (!fecha) {
    return 'Sin fecha'
  }
  const [anio, mes, dia] = fecha.slice(0, 10).split('-')
  return `${dia}/${mes}/${anio}`
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

function mensajeDeEliminacion(error: unknown): string {
  if (esApiError(error)) {
    if (error.status === 0) {
      return 'No se pudo conectar con el servidor. Intentá nuevamente.'
    }
    if (error.status === 404) {
      return 'No tenés acceso a este proyecto o no existe'
    }
    if (error.code === 'PROYECTO_CON_HISTORIAL') {
      return error.message
    }
  }
  return 'No se pudo eliminar el proyecto. Intentá nuevamente.'
}

export default function InformacionGeneralPage() {
  const { id } = useParams()
  const proyectoId = Number(id)
  const navegar = useNavigate()
  const idDescripcion = useId()

  const [estado, setEstado] = useState<EstadoCarga>({ tipo: 'cargando' })
  const [formulario, setFormulario] = useState<Formulario>({
    nombre: '',
    descripcion: '',
    fecha_inicio: '',
    fecha_fin: '',
    estado: 'Activo',
  })
  const [errores, setErrores] = useState<ErroresFormulario>({})
  const [errorGuardado, setErrorGuardado] = useState<string | null>(null)
  const [exito, setExito] = useState<string | null>(null)
  const [guardando, setGuardando] = useState(false)
  const [reabriendo, setReabriendo] = useState(false)
  const [confirmarEliminar, setConfirmarEliminar] = useState(false)
  const [eliminando, setEliminando] = useState(false)
  const [errorEliminar, setErrorEliminar] = useState<string | null>(null)
  const [intento, setIntento] = useState(0)

  useEffect(() => {
    const controlador = new AbortController()
    obtenerProyecto(proyectoId, controlador.signal)
      .then((proyecto) => {
        setEstado({ tipo: 'listo', proyecto })
        setFormulario(desdeProyecto(proyecto))
      })
      .catch((error: unknown) => {
        if (error instanceof Error && error.name === 'AbortError') {
          return
        }
        setEstado({ tipo: 'error', es404: esApiError(error) && error.status === 404 })
      })
    return () => controlador.abort()
  }, [proyectoId, intento])

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

  function mismosDatos(proyecto: Proyecto): boolean {
    return (
      formulario.nombre.trim() === proyecto.nombre &&
      formulario.descripcion.trim() === (proyecto.descripcion ?? '') &&
      formulario.fecha_inicio === fechaISO(proyecto.fecha_inicio) &&
      formulario.fecha_fin === fechaISO(proyecto.fecha_fin)
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
    if (estado.tipo !== 'listo' || guardando) {
      return
    }

    setErrorGuardado(null)
    setExito(null)
    if (!validarFormulario()) {
      return
    }

    const proyectoActual = estado.proyecto
    const datosCambiaron = !mismosDatos(proyectoActual)
    const cambiaEstado = formulario.estado !== proyectoActual.estado

    if (!datosCambiaron && !cambiaEstado) {
      setExito('No hay cambios para guardar.')
      return
    }

    setGuardando(true)
    try {
      if (datosCambiaron) {
        const actualizado = await actualizarProyecto(proyectoActual.id, {
          nombre: formulario.nombre.trim(),
          descripcion: formulario.descripcion.trim(),
          fecha_inicio: formulario.fecha_inicio,
          fecha_fin: formulario.fecha_fin === '' ? null : formulario.fecha_fin,
        })
        setEstado({ tipo: 'listo', proyecto: actualizado })
        setFormulario((anterior) => ({
          ...anterior,
          nombre: actualizado.nombre,
          descripcion: actualizado.descripcion ?? '',
          fecha_inicio: fechaISO(actualizado.fecha_inicio),
          fecha_fin: fechaISO(actualizado.fecha_fin),
        }))
      }

      if (cambiaEstado) {
        const conEstado = await cambiarEstadoProyecto(proyectoActual.id, formulario.estado)
        setEstado({ tipo: 'listo', proyecto: conEstado })
        setFormulario(desdeProyecto(conEstado))
      }

      setExito('Los cambios se guardaron correctamente.')
    } catch (error) {
      mostrarErrorGuardado(error)
    } finally {
      setGuardando(false)
    }
  }

  async function reabrir() {
    if (estado.tipo !== 'listo' || reabriendo) {
      return
    }
    setReabriendo(true)
    setErrorGuardado(null)
    setExito(null)
    try {
      const reabierto = await cambiarEstadoProyecto(estado.proyecto.id, 'Activo')
      setEstado({ tipo: 'listo', proyecto: reabierto })
      setFormulario(desdeProyecto(reabierto))
      setExito('Proyecto reabierto.')
    } catch (error) {
      mostrarErrorGuardado(error)
    } finally {
      setReabriendo(false)
    }
  }

  function pedirEliminar() {
    setErrorEliminar(null)
    setConfirmarEliminar(true)
  }

  function cancelarEliminar() {
    if (!eliminando) {
      setConfirmarEliminar(false)
      setErrorEliminar(null)
    }
  }

  async function confirmarEliminacion() {
    if (estado.tipo !== 'listo' || eliminando) {
      return
    }
    setEliminando(true)
    setErrorEliminar(null)
    try {
      await eliminarProyecto(estado.proyecto.id)
      navegar('/proyectos')
    } catch (error) {
      setErrorEliminar(mensajeDeEliminacion(error))
      setEliminando(false)
    }
  }

  if (estado.tipo === 'cargando') {
    return (
      <section>
        <PageHeader title="Información general" description="Consultá y actualizá los datos del proyecto." />
        <p role="status" className="text-ink-soft">
          Cargando proyecto…
        </p>
      </section>
    )
  }

  if (estado.tipo === 'error') {
    return (
      <section>
        <PageHeader title="Información general" description="Consultá y actualizá los datos del proyecto." />
        {estado.es404 ? (
          <Alert tono="error">No tenés acceso a este proyecto o no existe</Alert>
        ) : (
          <div className="flex flex-col items-start gap-4">
            <Alert tono="error">No se pudo cargar el proyecto. Intentá nuevamente.</Alert>
            <Button
              variante="secondary"
              onClick={() => {
                setEstado({ tipo: 'cargando' })
                setIntento((anterior) => anterior + 1)
              }}
            >
              Reintentar
            </Button>
          </div>
        )}
      </section>
    )
  }

  const proyecto = estado.proyecto
  const cerrado = proyecto.estado === 'Cerrado'

  return (
    <section className="flex flex-col gap-6">
      <PageHeader
        title="Información general"
        description="Consultá y actualizá los datos del proyecto."
      />

      <Card title="Datos del proyecto">
        <dl className="grid gap-4 sm:grid-cols-2">
          <div>
            <dt className="text-xs uppercase tracking-wide text-ink-soft">Nombre</dt>
            <dd className="text-sm text-ink">{proyecto.nombre}</dd>
          </div>
          <div>
            <dt className="text-xs uppercase tracking-wide text-ink-soft">Estado</dt>
            <dd className="text-sm text-ink">{proyecto.estado}</dd>
          </div>
          <div>
            <dt className="text-xs uppercase tracking-wide text-ink-soft">Fecha de inicio</dt>
            <dd className="text-sm text-ink">{formatearFecha(proyecto.fecha_inicio)}</dd>
          </div>
          <div>
            <dt className="text-xs uppercase tracking-wide text-ink-soft">Fecha de fin</dt>
            <dd className="text-sm text-ink">{formatearFecha(proyecto.fecha_fin)}</dd>
          </div>
          <div>
            <dt className="text-xs uppercase tracking-wide text-ink-soft">Cantidad de usuarios</dt>
            <dd className="text-sm text-ink">{proyecto.cantidad_usuarios}</dd>
          </div>
          <div className="sm:col-span-2">
            <dt className="text-xs uppercase tracking-wide text-ink-soft">Descripción</dt>
            <dd className="whitespace-pre-line text-sm text-ink">
              {proyecto.descripcion === '' ? 'Sin descripción' : proyecto.descripcion}
            </dd>
          </div>
        </dl>
        <div className="mt-6">
          <Button variante="secondary" onClick={pedirEliminar}>
            Eliminar proyecto
          </Button>
        </div>
      </Card>

      <Card title="Editar proyecto">
        {exito && <Alert tono="success">{exito}</Alert>}
        {errorGuardado && <Alert tono="error">{errorGuardado}</Alert>}

        {cerrado && (
          <div className="mb-4 flex flex-col items-start gap-3">
            <Alert tono="info">Reabrí el proyecto para editar sus datos</Alert>
            <Button variante="secondary" onClick={() => void reabrir()} disabled={reabriendo}>
              {reabriendo ? 'Reabriendo…' : 'Reabrir proyecto'}
            </Button>
          </div>
        )}

        <form noValidate onSubmit={(evento) => void guardar(evento)} className="mt-4 flex flex-col gap-5">
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

          <div>
            <Button type="submit" disabled={cerrado || guardando}>
              {guardando ? 'Guardando…' : 'Guardar cambios'}
            </Button>
          </div>
        </form>
      </Card>

      {confirmarEliminar && (
        <ModalConfirmacion
          titulo="Eliminar proyecto"
          textoConfirmar="Eliminar"
          procesando={eliminando}
          onConfirmar={() => void confirmarEliminacion()}
          onCancelar={cancelarEliminar}
        >
          {errorEliminar && <Alert tono="error">{errorEliminar}</Alert>}
          <p>
            ¿Querés eliminar <strong className="text-ink">{proyecto.nombre}</strong>?
          </p>
          <p>Esta acción no se puede deshacer desde la interfaz.</p>
        </ModalConfirmacion>
      )}
    </section>
  )
}
