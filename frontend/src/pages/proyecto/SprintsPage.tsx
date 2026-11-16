import { type KeyboardEvent, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import type { EstadoSprint, Sprint } from '../../api/types'
import PageHeader from '../../components/PageHeader'
import Alert from '../../ui/Alert'
import Button from '../../ui/Button'
import Card from '../../ui/Card'
import ModalConfirmacion from './sprints/ModalConfirmacion'
import ModalSprint from './sprints/ModalSprint'
import GoalSprint from './sprints/GoalSprint'
import { claseEstadoSprint } from './sprints/etiquetasSprint'
import { useSprints } from './sprints/useSprints'

type FiltroEstado = 'Todos' | EstadoSprint

type Confirmacion = { sprint: Sprint; destino: EstadoSprint }

const FILTROS: readonly FiltroEstado[] = ['Todos', 'Planificado', 'Activo', 'Cerrado']

export const TAMANO_PAGINA_SPRINTS = 10

function parsearProyectoId(id: string | undefined): number | null {
  if (id === undefined) {
    return null
  }
  const numero = Number(id)
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

function ordenarPorFechaInicio(sprints: Sprint[]): Sprint[] {
  return [...sprints].sort(
    (a, b) => a.fecha_inicio.localeCompare(b.fecha_inicio) || a.id - b.id,
  )
}

function esInteractivo(objetivo: EventTarget | null): boolean {
  return objetivo instanceof HTMLElement && objetivo.closest('a, button, select, input, textarea') !== null
}

export default function SprintsPage() {
  const { id } = useParams()
  const navegar = useNavigate()
  const proyectoId = parsearProyectoId(id)
  const { estado, reintentar, crear, cambiarEstado, procesandoId } = useSprints(proyectoId)
  const [filtro, setFiltro] = useState<FiltroEstado>('Todos')
  const [pagina, setPagina] = useState(1)
  const [altaAbierta, setAltaAbierta] = useState(false)
  const [confirmacion, setConfirmacion] = useState<Confirmacion | null>(null)
  const [errorOperacion, setErrorOperacion] = useState<string | null>(null)

  const sprints = estado.tipo === 'listo' ? estado.sprints : []
  const hayActivo = sprints.some((sprint) => sprint.estado === 'Activo')
  const sprintsFiltrados = ordenarPorFechaInicio(
    filtro === 'Todos' ? sprints : sprints.filter((sprint) => sprint.estado === filtro),
  )
  const totalPaginas = Math.max(1, Math.ceil(sprintsFiltrados.length / TAMANO_PAGINA_SPRINTS))
  const paginaActual = Math.min(pagina, totalPaginas)
  const sprintsDePagina = sprintsFiltrados.slice(
    (paginaActual - 1) * TAMANO_PAGINA_SPRINTS,
    paginaActual * TAMANO_PAGINA_SPRINTS,
  )

  function alPresionarFila(evento: KeyboardEvent<HTMLTableRowElement>, sprintId: number) {
    if (!esInteractivo(evento.target) && (evento.key === 'Enter' || evento.key === ' ')) {
      evento.preventDefault()
      irADetalle(sprintId)
    }
  }

  const contadores: Record<FiltroEstado, number> = {
    Todos: sprints.length,
    Planificado: sprints.filter((s) => s.estado === 'Planificado').length,
    Activo: sprints.filter((s) => s.estado === 'Activo').length,
    Cerrado: sprints.filter((s) => s.estado === 'Cerrado').length,
  }

  function irADetalle(sprintId: number) {
    navegar(`/proyectos/${proyectoId}/sprints/${sprintId}`)
  }

  async function confirmarAccion() {
    if (confirmacion === null) {
      return
    }
    setErrorOperacion(null)
    const resultado = await cambiarEstado(confirmacion.sprint.id, confirmacion.destino)
    setConfirmacion(null)
    if (!resultado.ok) {
      setErrorOperacion(resultado.mensaje)
    }
  }

  return (
    <section>
      <PageHeader
        title="Sprints"
        description="Planificá las iteraciones del proyecto y gestioná su ciclo de vida."
      />

      {errorOperacion && (
        <div className="mb-5">
          <Alert tono="error">{errorOperacion}</Alert>
        </div>
      )}

      {estado.tipo === 'cargando' && (
        <p role="status" className="text-ink-soft">
          Cargando sprints…
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

      {estado.tipo === 'listo' && sprints.length === 0 && (
        <Card
          title="Todavía no hay sprints"
          description="Creá el primer sprint para empezar a planificar el proyecto."
        >
          <Button onClick={() => setAltaAbierta(true)}>Nuevo sprint</Button>
        </Card>
      )}

      {estado.tipo === 'listo' && sprints.length > 0 && (
        <>
          <div className="mb-5 flex flex-wrap items-center justify-between gap-3">
            <div className="flex flex-wrap items-center gap-2" role="group" aria-label="Filtrar por estado">
              {FILTROS.map((opcion) => (
                <button
                  key={opcion}
                  type="button"
                  aria-pressed={filtro === opcion}
                  onClick={() => {
                    setFiltro(opcion)
                    setPagina(1)
                  }}
                  className={`inline-flex items-center gap-2 rounded-pill border px-3 py-1.5 text-sm font-semibold transition ${
                    filtro === opcion
                      ? 'border-primary bg-primary text-[var(--on-primary)]'
                      : 'border-border bg-surface text-ink-soft hover:border-primary hover:text-ink'
                  }`}
                >
                  {opcion}
                  <span
                    className={`rounded-pill px-2 py-0.5 text-xs ${
                      filtro === opcion ? 'bg-white/20' : 'bg-[var(--surface-2)]'
                    }`}
                  >
                    {contadores[opcion]}
                  </span>
                </button>
              ))}
            </div>
            <Button onClick={() => setAltaAbierta(true)}>Nuevo sprint</Button>
          </div>

          {sprintsFiltrados.length === 0 ? (
            <Alert tono="info">No hay sprints con ese estado</Alert>
          ) : (
            <div className="tabla-contenedor">
              <table className="tabla min-w-[48rem]">
                <thead>
                  <tr>
                    <th scope="col" className="text-center">
                      Identificador
                    </th>
                    <th scope="col" className="text-center">
                      Sprint Goal
                    </th>
                    <th scope="col" className="text-center">
                      Fecha de inicio
                    </th>
                    <th scope="col" className="text-center">
                      Fecha de fin
                    </th>
                    <th scope="col" className="text-center">
                      Estado
                    </th>
                    <th scope="col" className="col-acciones">
                      Acciones
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {sprintsDePagina.map((sprint) => {
                    const detalle = `/proyectos/${proyectoId}/sprints/${sprint.id}`
                    const dias = duracionEnDias(sprint.fecha_inicio, sprint.fecha_fin)
                    return (
                      <tr
                        key={sprint.id}
                        tabIndex={0}
                        className="cursor-pointer focus-visible:shadow-[var(--ring)]"
                        onClick={(evento) => {
                          if (!esInteractivo(evento.target)) {
                            irADetalle(sprint.id)
                          }
                        }}
                        onKeyDown={(evento) => alPresionarFila(evento, sprint.id)}
                      >
                        <td className="text-center font-medium text-ink">
                          <div className="flex flex-col items-center gap-0.5">
                            <Link
                              to={detalle}
                              className="hover:text-primary hover:underline"
                            >
                              {sprint.identificador}
                            </Link>
                            <span className="text-xs font-normal text-ink-soft">
                              {dias} {dias === 1 ? 'día' : 'días'}
                            </span>
                          </div>
                        </td>
                        <td className="text-center">
                          <GoalSprint texto={sprint.sprint_goal} />
                        </td>
                        <td className="text-center">{formatearFecha(sprint.fecha_inicio)}</td>
                        <td className="text-center">{formatearFecha(sprint.fecha_fin)}</td>
                        <td className="text-center">
                          <span
                            className={`inline-flex items-center rounded-pill border px-3 py-1.5 text-xs font-semibold ${claseEstadoSprint(sprint.estado)}`}
                          >
                            {sprint.estado}
                          </span>
                        </td>
                        <td className="col-acciones">
                          <div className="tabla-acciones">
                            {sprint.estado === 'Planificado' && (
                              <Button
                                variante="success"
                                tamano="sm"
                                disabled={hayActivo || procesandoId === sprint.id}
                                onClick={() => setConfirmacion({ sprint, destino: 'Activo' })}
                              >
                                Activar
                              </Button>
                            )}
                            {sprint.estado === 'Activo' && (
                              <Button
                                variante="danger"
                                tamano="sm"
                                disabled={procesandoId === sprint.id}
                                onClick={() => setConfirmacion({ sprint, destino: 'Cerrado' })}
                              >
                                Cerrar
                              </Button>
                            )}
                            <Link
                              to={detalle}
                              className="inline-flex items-center gap-1.5 rounded-tile border border-border bg-surface px-3 py-1.5 text-xs font-semibold text-ink-soft transition hover:border-primary hover:bg-[var(--primary-100)] hover:text-primary focus-visible:shadow-[var(--ring)]"
                            >
                              Ver sprint
                            </Link>
                          </div>
                        </td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            </div>
          )}

          {totalPaginas > 1 && (
            <nav
              aria-label="Paginación de sprints"
              className="mt-4 flex flex-wrap items-center justify-end gap-3"
            >
              <Button
                variante="ghost"
                disabled={paginaActual === 1}
                onClick={() => setPagina(paginaActual - 1)}
              >
                Anterior
              </Button>
              <span className="text-sm text-ink-soft">
                Página {paginaActual} de {totalPaginas}
              </span>
              <Button
                variante="ghost"
                disabled={paginaActual === totalPaginas}
                onClick={() => setPagina(paginaActual + 1)}
              >
                Siguiente
              </Button>
            </nav>
          )}
        </>
      )}

      {altaAbierta && (
        <ModalSprint
          onCerrar={() => setAltaAbierta(false)}
          onGuardar={async (campos) => {
            const resultado = await crear(campos)
            if (resultado.ok) {
              setAltaAbierta(false)
            }
            return resultado
          }}
        />
      )}

      {confirmacion && (
        <ModalConfirmacion
          titulo={confirmacion.destino === 'Activo' ? 'Activar sprint' : 'Cerrar sprint'}
          mensaje={
            confirmacion.destino === 'Activo'
              ? `¿Querés activar "${confirmacion.sprint.identificador}"?`
              : `Al cerrar "${confirmacion.sprint.identificador}" no se puede volver a abrir. ¿Querés cerrarlo?`
          }
          textoConfirmar="Confirmar"
          deshabilitado={procesandoId !== null}
          onConfirmar={() => void confirmarAccion()}
          onCerrar={() => setConfirmacion(null)}
        />
      )}
    </section>
  )
}
