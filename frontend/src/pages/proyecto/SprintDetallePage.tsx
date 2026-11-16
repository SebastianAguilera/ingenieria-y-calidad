import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { esApiError } from '../../api/errors'
import { listarHistorias } from '../../api/historias'
import { obtenerMetricasSprint } from '../../api/metricas'
import type {
  ActualizacionHistoria,
  Historia,
  MetricasSprint,
  Sprint,
  SprintBacklog,
} from '../../api/types'
import Alert from '../../ui/Alert'
import Button from '../../ui/Button'
import Card from '../../ui/Card'
import ModalHistoria from './backlog/ModalHistoria'
import { claseEstadoSprint } from './sprints/etiquetasSprint'
import ModalAsignarHistoria from './sprints/ModalAsignarHistoria'
import ModalConfirmacion from './sprints/ModalConfirmacion'
import TablaSprintBacklog from './sprints/TablaSprintBacklog'
import { useSprintBacklog } from './sprints/useSprintBacklog'
import { formatearHoras, formatearNumero, formatearPorcentaje } from './metricas/formato'

const MENSAJE_ERROR_CANDIDATAS = 'No se pudieron cargar las historias disponibles.'

type Candidatas =
  | { tipo: 'cargando' }
  | { tipo: 'error'; mensaje: string }
  | { tipo: 'listo'; historias: Historia[] }

function parsearId(valor: string | undefined): number | null {
  if (valor === undefined) return null
  const numero = Number(valor)
  return Number.isInteger(numero) && numero > 0 ? numero : null
}

function formatearFecha(fecha: string): string {
  const [anio, mes, dia] = fecha.slice(0, 10).split('-')
  return `${dia}/${mes}/${anio}`
}

function duracionEnDias(inicio: string, fin: string): number {
  const a = Date.parse(inicio.slice(0, 10))
  const b = Date.parse(fin.slice(0, 10))
  return Math.floor((b - a) / 86400000) + 1
}

function Kpi({ nombre, valor }: { nombre: string; valor: string }) {
  return (
    <div className="rounded-tile border border-border bg-[var(--surface-2)] p-3 text-center">
      <p className="text-xs font-semibold text-ink-soft">{nombre}</p>
      <p className="mt-1 font-display text-lg font-semibold text-ink">{valor}</p>
    </div>
  )
}

function PanelKpis({
  proyectoId,
  sprintId,
  refresco,
}: {
  proyectoId: number
  sprintId: number
  refresco: number
}) {
  const clave = `${proyectoId}:${sprintId}:${refresco}`
  const [consulta, setConsulta] = useState<{
    clave: string
    estado: { tipo: 'error' } | { tipo: 'listo'; metricas: MetricasSprint }
  } | null>(null)

  useEffect(() => {
    let activo = true
    obtenerMetricasSprint(proyectoId, sprintId)
      .then((metricas) => {
        if (activo) setConsulta({ clave, estado: { tipo: 'listo', metricas } })
      })
      .catch(() => {
        if (activo) setConsulta({ clave, estado: { tipo: 'error' } })
      })
    return () => {
      activo = false
    }
  }, [clave, proyectoId, sprintId])

  const estado = consulta?.clave === clave ? consulta.estado : { tipo: 'cargando' as const }
  const sinDatos = estado.tipo !== 'listo'
  const m = estado.tipo === 'listo' ? estado.metricas : null
  const pct = m?.porcentaje_completado
  return (
    <section aria-label="Indicadores del sprint" className="mb-6 grid gap-3 sm:grid-cols-2 lg:grid-cols-6">
      <Kpi nombre="SP planificados" valor={sinDatos ? '—' : formatearNumero(m!.story_points_planificados)} />
      <Kpi nombre="SP completados" valor={sinDatos ? '—' : formatearNumero(m!.story_points_completados)} />
      <Kpi
        nombre="% completado"
        valor={sinDatos || pct == null ? '—' : formatearPorcentaje(pct)}
      />
      <Kpi nombre="Historias completadas" valor={sinDatos ? '—' : String(m!.historias_completadas)} />
      <Kpi nombre="Horas estimadas" valor={sinDatos ? '—' : formatearHoras(m!.horas_estimadas)} />
      <Kpi nombre="Horas reales" valor={sinDatos ? '—' : formatearHoras(m!.horas_reales)} />
    </section>
  )
}

function CabeceraSprint({
  sprint,
  proyectoId,
  onAsignar,
  mostrarAsignar,
}: {
  sprint: Sprint
  proyectoId: number
  onAsignar: () => void
  mostrarAsignar: boolean
}) {
  const dias = duracionEnDias(sprint.fecha_inicio, sprint.fecha_fin)
  return (
    <header className="mb-6">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 className="font-display text-3xl font-bold text-ink">{sprint.identificador}</h1>
          <p className="mt-1 text-sm text-ink-soft">{sprint.sprint_goal}</p>
          <dl className="mt-3 flex flex-wrap items-center gap-x-6 gap-y-2 text-sm">
            <div className="flex items-center gap-2">
              <dt className="text-ink-soft">Estado</dt>
              <dd>
                <span
                  className={`inline-flex items-center rounded-pill border px-3 py-1.5 text-xs font-semibold ${claseEstadoSprint(sprint.estado)}`}
                >
                  {sprint.estado}
                </span>
              </dd>
            </div>
            <div className="flex items-center gap-2">
              <dt className="text-ink-soft">Fechas</dt>
              <dd className="font-medium text-ink">
                {formatearFecha(sprint.fecha_inicio)} — {formatearFecha(sprint.fecha_fin)}
              </dd>
            </div>
            <div className="flex items-center gap-2">
              <dt className="text-ink-soft">Duración</dt>
              <dd className="font-medium text-ink">
                {dias} {dias === 1 ? 'día' : 'días'}
              </dd>
            </div>
          </dl>
        </div>
        <div className="flex flex-wrap gap-2">
          {mostrarAsignar && <Button onClick={onAsignar}>Asignar historia</Button>}
          <Link
            to={`/proyectos/${proyectoId}/reporte`}
            className="inline-flex items-center gap-2 rounded-tile border border-border bg-surface px-4 py-2 text-sm font-semibold text-ink-soft hover:border-primary hover:text-primary"
          >
            Ver reporte
          </Link>
          <Link
            to={`/proyectos/${proyectoId}/metricas`}
            className="inline-flex items-center gap-2 rounded-tile border border-border bg-surface px-4 py-2 text-sm font-semibold text-ink-soft hover:border-primary hover:text-primary"
          >
            Ver métricas
          </Link>
        </div>
      </div>
    </header>
  )
}

export default function SprintDetallePage() {
  const { id, sprintId } = useParams()
  const navigate = useNavigate()
  const proyectoId = parsearId(id)
  const { estado, reintentar, asignar, quitar, actualizarHistoria, procesando } = useSprintBacklog(
    proyectoId,
    parsearId(sprintId),
  )
  const [asignacionAbierta, setAsignacionAbierta] = useState(false)
  const [candidatas, setCandidatas] = useState<Candidatas>({ tipo: 'cargando' })
  const [errorAsignacion, setErrorAsignacion] = useState<string | null>(null)
  const [aQuitar, setAQuitar] = useState<Historia | null>(null)
  const [aEditar, setAEditar] = useState<Historia | null>(null)
  const [errorOperacion, setErrorOperacion] = useState<string | null>(null)
  const [refrescoKpis, setRefrescoKpis] = useState(0)

  const urlLista = `/proyectos/${id ?? ''}/sprints`

  async function abrirAsignacion() {
    if (proyectoId === null) return
    setErrorOperacion(null)
    setErrorAsignacion(null)
    setCandidatas({ tipo: 'cargando' })
    setAsignacionAbierta(true)
    try {
      const historias = await listarHistorias(proyectoId)
      setCandidatas({ tipo: 'listo', historias })
    } catch (error) {
      const mensaje =
        esApiError(error) && error.status !== 0 && error.message !== ''
          ? error.message
          : MENSAJE_ERROR_CANDIDATAS
      setCandidatas({ tipo: 'error', mensaje })
    }
  }

  function cerrarAsignacion() {
    setAsignacionAbierta(false)
    setErrorAsignacion(null)
  }

  async function confirmarAsignacion(historiaId: number) {
    setErrorAsignacion(null)
    const resultado = await asignar(historiaId)
    if (resultado.ok) {
      cerrarAsignacion()
      setRefrescoKpis((n) => n + 1)
      return
    }
    setErrorAsignacion(resultado.mensaje)
  }

  async function confirmarQuitar() {
    if (aQuitar === null) return
    const historia = aQuitar
    setErrorOperacion(null)
    const resultado = await quitar(historia.id)
    setAQuitar(null)
    if (!resultado.ok) {
      setErrorOperacion(resultado.mensaje)
    } else {
      setRefrescoKpis((n) => n + 1)
    }
  }

  function historiasDisponibles(backlog: SprintBacklog): Historia[] {
    if (candidatas.tipo !== 'listo') return []
    const asignadas = new Set(backlog.items.map((item) => item.historia.id))
    return candidatas.historias.filter((h) => h.estado === 'Backlog' && !asignadas.has(h.id))
  }

  return (
    <section>
      <Link
        to={urlLista}
        className="mb-4 inline-flex items-center gap-2 text-sm font-medium text-primary hover:underline"
      >
        Volver a sprints
      </Link>

      {estado.tipo === 'cargando' && (
        <p role="status" className="text-ink-soft">
          Cargando sprint…
        </p>
      )}

      {estado.tipo === 'error' && (
        <div className="flex flex-col items-start gap-4">
          <Alert tono="error">{estado.mensaje}</Alert>
          <Button variante="secondary" onClick={reintentar}>
            Reintentar
          </Button>
        </div>
      )}

      {estado.tipo === 'noEncontrado' && <Alert tono="error">{estado.mensaje}</Alert>}

      {estado.tipo === 'listo' && proyectoId !== null && (
        <>
          <CabeceraSprint
            sprint={estado.backlog.sprint}
            proyectoId={proyectoId}
            onAsignar={() => void abrirAsignacion()}
            mostrarAsignar={estado.backlog.sprint.estado === 'Activo'}
          />

          <PanelKpis
            proyectoId={proyectoId}
            sprintId={estado.backlog.sprint.id}
            refresco={refrescoKpis}
          />

          {errorOperacion && (
            <div className="mb-5">
              <Alert tono="error">{errorOperacion}</Alert>
            </div>
          )}

          {estado.backlog.sprint.estado === 'Planificado' && (
            <div className="mb-5">
              <Alert tono="info">Activá el sprint para poder asignar historias</Alert>
            </div>
          )}

          {estado.backlog.items.length === 0 ? (
            <Card title="Este sprint todavía no tiene historias">
              <p className="text-sm text-ink-soft">
                {estado.backlog.sprint.estado === 'Activo'
                  ? 'Asigná historias del Product Backlog para planificar el trabajo.'
                  : 'No hay historias asignadas a este sprint.'}
              </p>
            </Card>
          ) : (
            <TablaSprintBacklog
              items={estado.backlog.items}
              estadoSprint={estado.backlog.sprint.estado}
              onQuitar={(historia) => setAQuitar(historia)}
              onCambiar={(historia, cambios) => {
                setErrorOperacion(null)
                void actualizarHistoria(historia.id, cambios).then((r) => {
                  if (!r.ok) setErrorOperacion(r.mensaje)
                  else setRefrescoKpis((n) => n + 1)
                })
              }}
              onEditar={(historia) => setAEditar(historia)}
              onEstimar={(historia) => navigate(`/proyectos/${proyectoId}/backlog/${historia.id}/poker`)}
              deshabilitado={procesando}
            />
          )}

          {asignacionAbierta && (
            <ModalAsignarHistoria
              historias={historiasDisponibles(estado.backlog)}
              cargando={candidatas.tipo === 'cargando'}
              error={candidatas.tipo === 'error' ? candidatas.mensaje : errorAsignacion}
              enviando={procesando}
              onAsignar={(historiaId) => void confirmarAsignacion(historiaId)}
              onCerrar={cerrarAsignacion}
            />
          )}

          {aQuitar && (
            <ModalConfirmacion
              titulo="Quitar historia"
              mensaje={`¿Querés quitar "${aQuitar.codigo} ${aQuitar.titulo}" de ${estado.backlog.sprint.identificador}?`}
              textoConfirmar="Quitar"
              deshabilitado={procesando}
              onConfirmar={() => void confirmarQuitar()}
              onCerrar={() => setAQuitar(null)}
            />
          )}

          {aEditar && (
            <ModalHistoria
              titulo="Editar historia"
              textoBoton="Guardar cambios"
              historia={aEditar}
              onCerrar={() => setAEditar(null)}
              onGuardar={async (datos) => {
                const resultado = await actualizarHistoria(
                  aEditar.id,
                  datos as ActualizacionHistoria,
                )
                if (resultado.ok) {
                  setAEditar(null)
                  setRefrescoKpis((n) => n + 1)
                }
                return resultado
              }}
            />
          )}
        </>
      )}
    </section>
  )
}
