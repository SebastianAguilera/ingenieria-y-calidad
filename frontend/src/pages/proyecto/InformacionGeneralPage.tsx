import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { eliminarProyecto, listarIntegrantes, obtenerProyecto } from '../../api/proyectos'
import { esApiError } from '../../api/errors'
import type { Proyecto, UsuarioProyecto } from '../../api/types'
import Alert from '../../ui/Alert'
import Button from '../../ui/Button'
import Card from '../../ui/Card'
import ModalConfirmacion from '../../ui/ModalConfirmacion'
import {
  duracionEnDias,
  formatearFecha,
  iniciales,
  ordenarEquipo,
  rangoDeFechas,
} from './informacion/utilidades'
import ModalEditarProyecto from './ModalEditarProyecto'

type EstadoCarga =
  | { tipo: 'cargando' }
  | { tipo: 'error'; es404: boolean }
  | { tipo: 'listo'; proyecto: Proyecto }

type EstadoEquipo =
  | { tipo: 'cargando' }
  | { tipo: 'error' }
  | { tipo: 'listo'; miembros: UsuarioProyecto[] }

function EtiquetaEstado({ estado }: { estado: Proyecto['estado'] }) {
  const cerrado = estado === 'Cerrado'
  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-semibold ${
        cerrado
          ? 'bg-[var(--accent-100)] text-[var(--danger)]'
          : 'bg-[var(--primary-100)] text-[var(--primary-700)]'
      }`}
    >
      <i className={`bi ${cerrado ? 'bi-lock-fill' : 'bi-circle-fill'} text-[0.6rem]`} aria-hidden="true" />
      {estado}
    </span>
  )
}

function EtiquetaRol({ rol }: { rol: UsuarioProyecto['rol'] }) {
  const esLider = rol === 'LIDER'
  return (
    <span
      className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${
        esLider
          ? 'bg-[var(--primary-500)] text-[var(--on-primary)]'
          : 'border border-border text-ink-soft'
      }`}
    >
      {esLider ? 'Líder' : 'Miembro'}
    </span>
  )
}

function FilaDato({ etiqueta, valor }: { etiqueta: string; valor: string }) {
  return (
    <div className="flex items-baseline justify-between gap-4 border-b border-border py-2.5 last:border-b-0">
      <dt className="text-sm text-ink-soft">{etiqueta}</dt>
      <dd className="text-right text-sm font-medium text-ink">{valor}</dd>
    </div>
  )
}

function textoDuracion(proyecto: Proyecto): string {
  if (proyecto.fecha_fin === null) {
    return 'Sin fecha de fin'
  }
  const dias = duracionEnDias(proyecto.fecha_inicio, proyecto.fecha_fin)
  return dias === null ? 'Sin dato' : `${dias} ${dias === 1 ? 'día' : 'días'}`
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

function ContenidoEquipo({ equipo, proyectoId }: { equipo: EstadoEquipo; proyectoId: number }) {
  if (equipo.tipo === 'cargando') {
    return (
      <p role="status" className="text-sm text-ink-soft">
        Cargando equipo…
      </p>
    )
  }

  if (equipo.tipo === 'error') {
    return <Alert tono="error">No se pudo cargar el equipo. Recargá la página.</Alert>
  }

  if (equipo.miembros.length === 0) {
    return (
      <p className="text-sm text-ink-soft">
        Todavía no hay integrantes en este proyecto.{' '}
        <Link to={`/proyectos/${proyectoId}/equipo`} className="font-medium text-ink underline">
          Agregar integrantes
        </Link>
      </p>
    )
  }

  return (
    <ul className="grid gap-3 sm:grid-cols-2 xl:grid-cols-3" aria-label="Integrantes del proyecto">
      {ordenarEquipo(equipo.miembros).map((miembro) => (
        <li
          key={miembro.id}
          className="flex items-center gap-3 rounded-tile border border-border bg-[var(--surface-2)] p-3"
        >
          <span
            aria-hidden="true"
            className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-[var(--primary-100)] text-sm font-semibold text-[var(--primary-700)]"
          >
            {iniciales(miembro.nombre)}
          </span>
          <div className="min-w-0 flex-1">
            <p className="truncate text-sm font-semibold text-ink">{miembro.nombre}</p>
            <p className="truncate text-xs text-ink-soft">{miembro.email}</p>
          </div>
          {miembro.rol && <EtiquetaRol rol={miembro.rol} />}
        </li>
      ))}
    </ul>
  )
}

export default function InformacionGeneralPage() {
  const { id } = useParams()
  const proyectoId = Number(id)
  const navegar = useNavigate()

  const [estado, setEstado] = useState<EstadoCarga>({ tipo: 'cargando' })
  const [equipo, setEquipo] = useState<EstadoEquipo>({ tipo: 'cargando' })
  const [intento, setIntento] = useState(0)
  const [editarAbierto, setEditarAbierto] = useState(false)
  const [confirmarEliminar, setConfirmarEliminar] = useState(false)
  const [eliminando, setEliminando] = useState(false)
  const [errorEliminar, setErrorEliminar] = useState<string | null>(null)

  useEffect(() => {
    const controlador = new AbortController()
    obtenerProyecto(proyectoId, controlador.signal)
      .then((proyecto) => setEstado({ tipo: 'listo', proyecto }))
      .catch((error: unknown) => {
        if (error instanceof Error && error.name === 'AbortError') {
          return
        }
        setEstado({ tipo: 'error', es404: esApiError(error) && error.status === 404 })
      })
    return () => controlador.abort()
  }, [proyectoId, intento])

  useEffect(() => {
    const controlador = new AbortController()
    setEquipo({ tipo: 'cargando' })
    listarIntegrantes(proyectoId, controlador.signal)
      .then((miembros) => setEquipo({ tipo: 'listo', miembros }))
      .catch((error: unknown) => {
        if (error instanceof Error && error.name === 'AbortError') {
          return
        }
        setEquipo({ tipo: 'error' })
      })
    return () => controlador.abort()
  }, [proyectoId, intento])

  if (estado.tipo === 'cargando') {
    return (
      <p role="status" className="text-ink-soft">
        Cargando proyecto…
      </p>
    )
  }

  if (estado.tipo === 'error') {
    if (estado.es404) {
      return <Alert tono="error">No tenés acceso a este proyecto o no existe</Alert>
    }
    return (
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
    )
  }

  const proyecto = estado.proyecto
  const totalIntegrantes =
    equipo.tipo === 'listo' ? equipo.miembros.length : proyecto.cantidad_usuarios

  function pedirEliminar() {
    setErrorEliminar(null)
    setConfirmarEliminar(true)
  }

  function cancelarEliminar() {
    if (eliminando) {
      return
    }
    setConfirmarEliminar(false)
    setErrorEliminar(null)
  }

  async function confirmarEliminacion() {
    if (eliminando) {
      return
    }
    setEliminando(true)
    setErrorEliminar(null)
    try {
      await eliminarProyecto(proyecto.id)
      navegar('/proyectos')
    } catch (error) {
      setErrorEliminar(mensajeDeEliminacion(error))
      setEliminando(false)
    }
  }

  return (
    <section className="flex flex-col gap-6">
      <header className="flex flex-wrap items-start justify-between gap-4">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-3">
            <h1 className="text-3xl font-bold tracking-tight text-ink">{proyecto.nombre}</h1>
            <EtiquetaEstado estado={proyecto.estado} />
          </div>
          <p className="mt-2 text-ink-soft">
            {rangoDeFechas(proyecto.fecha_inicio, proyecto.fecha_fin)}
          </p>
        </div>
        <div className="flex flex-wrap gap-3">
          <Button variante="secondary" onClick={() => setEditarAbierto(true)}>
            <i className="bi bi-pencil" aria-hidden="true" />
            Editar proyecto
          </Button>
          <Button variante="danger" onClick={pedirEliminar}>
            <i className="bi bi-trash" aria-hidden="true" />
            Eliminar proyecto
          </Button>
        </div>
      </header>

      <div className="grid gap-6 lg:grid-cols-3">
        <Card title="Descripción" className="lg:col-span-2">
          {proyecto.descripcion.trim() === '' ? (
            <p className="text-sm text-ink-soft">
              Sin descripción todavía.{' '}
              <button
                type="button"
                onClick={() => setEditarAbierto(true)}
                className="font-medium text-ink underline"
              >
                Agregarla desde Editar proyecto
              </button>
            </p>
          ) : (
            <p className="whitespace-pre-line text-sm leading-relaxed text-ink">
              {proyecto.descripcion}
            </p>
          )}
        </Card>

        <Card title="Resumen">
          <dl>
            <FilaDato etiqueta="Estado" valor={proyecto.estado} />
            <FilaDato etiqueta="Inicio" valor={formatearFecha(proyecto.fecha_inicio)} />
            <FilaDato etiqueta="Fin" valor={formatearFecha(proyecto.fecha_fin)} />
            <FilaDato etiqueta="Duración" valor={textoDuracion(proyecto)} />
            <FilaDato etiqueta="Integrantes" valor={String(totalIntegrantes)} />
            <FilaDato
              etiqueta="Creado por"
              valor={proyecto.creado_por?.nombre ?? 'No disponible'}
            />
            <FilaDato etiqueta="Creado el" valor={formatearFecha(proyecto.creado_en)} />
          </dl>
        </Card>
      </div>

      <Card title="Equipo" description="Personas asignadas al proyecto y su rol.">
        <ContenidoEquipo equipo={equipo} proyectoId={proyectoId} />
      </Card>

      {editarAbierto && (
        <ModalEditarProyecto
          proyecto={proyecto}
          onCerrar={() => setEditarAbierto(false)}
          onGuardado={(actualizado) => setEstado({ tipo: 'listo', proyecto: actualizado })}
        />
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
