import { useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { esApiError } from '../../api/errors'
import { listarHistorias } from '../../api/historias'
import type { Historia, Sprint, SprintBacklog } from '../../api/types'
import PageHeader from '../../components/PageHeader'
import Alert from '../../ui/Alert'
import Button from '../../ui/Button'
import Card from '../../ui/Card'
import { claseEstadoSprint } from './sprints/etiquetasSprint'
import ModalAsignarHistoria from './sprints/ModalAsignarHistoria'
import ModalConfirmacion from './sprints/ModalConfirmacion'
import TablaSprintBacklog from './sprints/TablaSprintBacklog'
import { useSprintBacklog } from './sprints/useSprintBacklog'

const MENSAJE_ERROR_CANDIDATAS = 'No se pudieron cargar las historias disponibles.'

type Candidatas =
  | { tipo: 'cargando' }
  | { tipo: 'error'; mensaje: string }
  | { tipo: 'listo'; historias: Historia[] }

function parsearId(valor: string | undefined): number | null {
  if (valor === undefined) {
    return null
  }
  const numero = Number(valor)
  return Number.isInteger(numero) && numero > 0 ? numero : null
}

function formatearFecha(fecha: string): string {
  const [anio, mes, dia] = fecha.slice(0, 10).split('-')
  return `${dia}/${mes}/${anio}`
}

function CabeceraSprint({ sprint }: { sprint: Sprint }) {
  return (
    <>
      <PageHeader title={sprint.identificador} description={sprint.sprint_goal} />
      <dl className="mb-6 flex flex-wrap items-center gap-x-8 gap-y-3 text-sm">
        <div className="flex items-center gap-2">
          <dt className="text-ink-soft">Fecha de inicio</dt>
          <dd className="font-medium text-ink">{formatearFecha(sprint.fecha_inicio)}</dd>
        </div>
        <div className="flex items-center gap-2">
          <dt className="text-ink-soft">Fecha de fin</dt>
          <dd className="font-medium text-ink">{formatearFecha(sprint.fecha_fin)}</dd>
        </div>
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
      </dl>
    </>
  )
}

export default function SprintDetallePage() {
  const { id, sprintId } = useParams()
  const proyectoId = parsearId(id)
  const { estado, reintentar, asignar, quitar, procesando } = useSprintBacklog(
    proyectoId,
    parsearId(sprintId),
  )
  const [asignacionAbierta, setAsignacionAbierta] = useState(false)
  const [candidatas, setCandidatas] = useState<Candidatas>({ tipo: 'cargando' })
  const [errorAsignacion, setErrorAsignacion] = useState<string | null>(null)
  const [aQuitar, setAQuitar] = useState<Historia | null>(null)
  const [errorOperacion, setErrorOperacion] = useState<string | null>(null)

  const urlLista = `/proyectos/${id ?? ''}/sprints`

  async function abrirAsignacion() {
    if (proyectoId === null) {
      return
    }
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
      return
    }
    setErrorAsignacion(resultado.mensaje)
  }

  async function confirmarQuitar() {
    if (aQuitar === null) {
      return
    }
    const historia = aQuitar
    setErrorOperacion(null)
    const resultado = await quitar(historia.id)
    setAQuitar(null)
    if (!resultado.ok) {
      setErrorOperacion(resultado.mensaje)
    }
  }

  function historiasDisponibles(backlog: SprintBacklog): Historia[] {
    if (candidatas.tipo !== 'listo') {
      return []
    }
    const asignadas = new Set(backlog.items.map((item) => item.historia.id))
    return candidatas.historias.filter(
      (historia) => historia.estado === 'Backlog' && !asignadas.has(historia.id),
    )
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

      {estado.tipo === 'listo' && (
        <>
          <CabeceraSprint sprint={estado.backlog.sprint} />

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

          {estado.backlog.sprint.estado === 'Activo' && (
            <div className="mb-5 flex justify-end">
              <Button onClick={() => void abrirAsignacion()}>Asignar historia</Button>
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
        </>
      )}
    </section>
  )
}
