import { useState } from 'react'
import { useParams } from 'react-router-dom'
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

function ordenarPorFechaInicio(sprints: Sprint[]): Sprint[] {
  return [...sprints].sort(
    (a, b) => a.fecha_inicio.localeCompare(b.fecha_inicio) || a.id - b.id,
  )
}

export default function SprintsPage() {
  const { id } = useParams()
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
            <label className="flex items-center gap-2 text-sm text-ink-soft">
              Filtrar por estado
              <span className="relative inline-flex items-center">
                <select
                  aria-label="Filtrar por estado"
                  value={filtro}
                  onChange={(evento) => {
                    setFiltro(evento.target.value as FiltroEstado)
                    setPagina(1)
                  }}
                  className="appearance-none rounded-pill border border-border bg-[var(--surface-2)] py-1.5 pl-3 pr-8 text-sm text-ink focus:border-primary focus:outline-none"
                >
                  {FILTROS.map((opcion) => (
                    <option key={opcion} value={opcion}>
                      {opcion}
                    </option>
                  ))}
                </select>
                <i
                  className="bi bi-chevron-down pointer-events-none absolute right-2.5 text-xs text-ink-soft"
                  aria-hidden="true"
                />
              </span>
            </label>
            <Button onClick={() => setAltaAbierta(true)}>Nuevo sprint</Button>
          </div>

          {sprintsFiltrados.length === 0 ? (
            <Alert tono="info">No hay sprints con ese estado</Alert>
          ) : (
            <div className="overflow-x-auto rounded-tile border border-border bg-surface">
              <table className="w-full min-w-[48rem] text-left text-sm">
                <thead className="border-b border-border text-ink-soft">
                  <tr>
                    <th scope="col" className="px-6 py-4 font-semibold">
                      Identificador
                    </th>
                    <th scope="col" className="px-6 py-4 font-semibold">
                      Sprint Goal
                    </th>
                    <th scope="col" className="px-6 py-4 font-semibold">
                      Fecha de inicio
                    </th>
                    <th scope="col" className="px-6 py-4 font-semibold">
                      Fecha de fin
                    </th>
                    <th scope="col" className="px-6 py-4 font-semibold">
                      Estado
                    </th>
                    <th scope="col" className="px-6 py-4 font-semibold">
                      Acciones
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {sprintsDePagina.map((sprint) => (
                    <tr key={sprint.id} className="border-b border-border last:border-b-0">
                      <td className="px-6 py-5 font-medium text-ink">{sprint.identificador}</td>
                      <td className="px-6 py-5">
                        <GoalSprint texto={sprint.sprint_goal} />
                      </td>
                      <td className="px-6 py-5">{formatearFecha(sprint.fecha_inicio)}</td>
                      <td className="px-6 py-5">{formatearFecha(sprint.fecha_fin)}</td>
                      <td className="px-6 py-5">
                        <span
                          className={`inline-flex items-center rounded-pill border px-3 py-1.5 text-xs font-semibold ${claseEstadoSprint(sprint.estado)}`}
                        >
                          {sprint.estado}
                        </span>
                      </td>
                      <td className="px-6 py-5">
                        <div className="flex flex-wrap gap-2">
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
                        </div>
                      </td>
                    </tr>
                  ))}
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
