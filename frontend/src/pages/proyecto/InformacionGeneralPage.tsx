import { useEffect, useId, useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import {
  actualizarProyecto,
  cambiarEstadoProyecto,
  eliminarProyecto,
  obtenerProyecto,
} from '../../api/proyectos'
import { listarSprints } from '../../api/sprints'
import { esApiError } from '../../api/errors'
import type { EstadoProyecto, Proyecto, Sprint } from '../../api/types'
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

/** Chip de estado del proyecto: Activo usa el verde de estado y Cerrado un gris atenuado. */
const CLASE_CHIP_ESTADO: Record<EstadoProyecto, string> = {
  Activo:
    'border-[var(--estado-hecho)] bg-[color-mix(in_oklab,var(--estado-hecho)_16%,transparent)] text-[var(--estado-hecho-ink)]',
  Cerrado: 'border-border bg-[var(--surface-2)] text-ink-soft',
}

/** Misma tipografía para el valor de todas las tarjetas de datos. */
const CLASE_VALOR_TARJETA = 'font-display text-lg font-semibold leading-tight'

const CLASE_TARJETA =
  'flex items-start gap-3 rounded-tile border border-border bg-[var(--surface-2)] p-4'

type TarjetaDatoProps = {
  icono: string
  titulo: string
  children: string
  atenuado?: boolean
  /** Si se indica, toda la tarjeta es un enlace. */
  a?: string
}

function TarjetaDato({ icono, titulo, children, atenuado = false, a }: TarjetaDatoProps) {
  const contenido = (
    <>
      <span
        aria-hidden="true"
        className="flex h-10 w-10 shrink-0 items-center justify-center rounded-tile bg-primary-100 text-lg text-primary"
      >
        <i className={`bi ${icono}`} />
      </span>
      <span className="flex min-w-0 flex-col">
        <span className="text-xs font-semibold text-ink-soft">{titulo}</span>
        <span
          data-testid="valor-tarjeta"
          className={`${CLASE_VALOR_TARJETA} ${atenuado ? 'text-ink-soft' : 'text-ink'}`}
        >
          {children}
        </span>
      </span>
      {a && <i className="bi bi-chevron-right ml-auto self-center text-ink-soft" aria-hidden="true" />}
    </>
  )

  if (a) {
    return (
      <Link
        to={a}
        className={`${CLASE_TARJETA} transition hover:border-primary hover:bg-[var(--primary-100)] focus-visible:shadow-[var(--ring)]`}
      >
        {contenido}
      </Link>
    )
  }
  return <div className={CLASE_TARJETA}>{contenido}</div>
}

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
  const [sprintActivo, setSprintActivo] = useState<Sprint | null>(null)

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

  // Sprint activo (tarjeta opcional): si falla la carga, simplemente no se muestra.
  useEffect(() => {
    const controlador = new AbortController()
    listarSprints(proyectoId, controlador.signal)
      .then((sprints) => {
        setSprintActivo(sprints.find((sprint) => sprint.estado === 'Activo') ?? null)
      })
      .catch(() => setSprintActivo(null))
    return () => controlador.abort()
  }, [proyectoId])

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
  const puedeEliminar = proyecto.rol_actual === 'LIDER'

  return (
    <section className="flex flex-col gap-6">
      <PageHeader
        title="Información general"
        description="Consultá y actualizá los datos del proyecto."
      />

      <Card>
        <header className="flex flex-wrap items-start justify-between gap-3">
          <div className="min-w-0">
            <div className="flex flex-wrap items-center gap-3">
              <h2 className="text-2xl font-bold text-ink">{proyecto.nombre}</h2>
              <span
                data-testid="chip-estado-proyecto"
                className={`inline-flex items-center rounded-pill border px-3 py-1 text-xs font-semibold ${CLASE_CHIP_ESTADO[proyecto.estado]}`}
              >
                {proyecto.estado}
              </span>
            </div>
            <p className="mt-1 whitespace-pre-line text-sm text-ink-soft">
              {proyecto.descripcion === '' ? 'Sin descripción' : proyecto.descripcion}
            </p>
          </div>
          <Button
            variante="secondary"
            onClick={() =>
              document.getElementById('editar-proyecto')?.scrollIntoView({ behavior: 'smooth' })
            }
          >
            Editar
          </Button>
        </header>

        <div className="mt-6 grid grid-cols-2 gap-3 lg:grid-cols-4">
          <TarjetaDato icono="bi-calendar-event" titulo="Fecha de inicio">
            {formatearFecha(proyecto.fecha_inicio)}
          </TarjetaDato>
          <TarjetaDato
            icono="bi-calendar-check"
            titulo="Fecha de fin"
            atenuado={proyecto.fecha_fin === null}
          >
            {formatearFecha(proyecto.fecha_fin)}
          </TarjetaDato>
          <TarjetaDato
            icono="bi-people"
            titulo="Integrantes"
            a={`/proyectos/${proyecto.id}/equipo`}
          >
            {String(proyecto.cantidad_usuarios)}
          </TarjetaDato>
          {sprintActivo && (
            <TarjetaDato
              icono="bi-lightning-charge"
              titulo="Sprint activo"
              a={`/proyectos/${proyecto.id}/sprints/${sprintActivo.id}`}
            >
              {sprintActivo.identificador}
            </TarjetaDato>
          )}
        </div>
      </Card>

      <div id="editar-proyecto">
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
      </div>

      {puedeEliminar && (
        <section
          aria-labelledby="zona-peligro-titulo"
          className="rounded-card border border-[color-mix(in_oklab,var(--danger)_45%,transparent)] bg-[color-mix(in_oklab,var(--danger)_6%,transparent)] p-6"
        >
          <h2 id="zona-peligro-titulo" className="text-xl font-bold text-[var(--danger)]">
            Zona de peligro
          </h2>
          <p className="mt-1 text-sm text-ink-soft">
            Eliminar el proyecto no se puede deshacer.
          </p>
          <div className="mt-4">
            <Button variante="danger" onClick={pedirEliminar}>
              Eliminar proyecto
            </Button>
          </div>
        </section>
      )}

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
