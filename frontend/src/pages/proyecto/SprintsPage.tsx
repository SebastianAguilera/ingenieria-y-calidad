import { useState } from 'react'
import { useParams } from 'react-router-dom'
import type { EstadoSprint, Sprint } from '../../api/types'
import PageHeader from '../../components/PageHeader'
import Alert from '../../ui/Alert'
import Button from '../../ui/Button'
import Card from '../../ui/Card'
import ModalConfirmacion from './sprints/ModalConfirmacion'
import ModalSprint from './sprints/ModalSprint'
import { useSprints } from './sprints/useSprints'

type FiltroEstado = 'Todos' | EstadoSprint

type Confirmacion = { sprint: Sprint; destino: EstadoSprint }

const FILTROS: readonly FiltroEstado[] = ['Todos', 'Planificado', 'Activo', 'Cerrado']

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

export default function SprintsPage() {
  const { id } = useParams()
  const proyectoId = parsearProyectoId(id)
  const { estado, reintentar, crear, cambiarEstado, procesandoId } = useSprints(proyectoId)
  const [filtro, setFiltro] = useState<FiltroEstado>('Todos')
  const [altaAbierta, setAltaAbierta] = useState(false)
  const [confirmacion, setConfirmacion] = useState<Confirmacion | null>(null)
  const [errorOperacion, setErrorOperacion] = useState<string | null>(null)

  const sprints = estado.tipo === 'listo' ? estado.sprints : []
  const hayActivo = sprints.some((sprint) => sprint.estado === 'Activo')
  const sprintsFiltrados =
    filtro === 'Todos' ? sprints : sprints.filter((sprint) => sprint.estado === filtro)

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
              <select
                aria-label="Filtrar por estado"
                value={filtro}
                onChange={(evento) => setFiltro(evento.target.value as FiltroEstado)}
                className="rounded-pill border border-border bg-[var(--surface-2)] px-3 py-1.5 text-sm text-ink focus:border-primary focus:outline-none"
              >
                {FILTROS.map((opcion) => (
                  <option key={opcion} value={opcion}>
                    {opcion}
                  </option>
                ))}
              </select>
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
                  {sprintsFiltrados.map((sprint) => (
                    <tr key={sprint.id} className="border-b border-border last:border-b-0">
                      <td className="px-6 py-5 font-medium text-ink">{sprint.identificador}</td>
                      <td className="px-6 py-5">{sprint.sprint_goal}</td>
                      <td className="px-6 py-5">{formatearFecha(sprint.fecha_inicio)}</td>
                      <td className="px-6 py-5">{formatearFecha(sprint.fecha_fin)}</td>
                      <td className="px-6 py-5">{sprint.estado}</td>
                      <td className="px-6 py-5">
                        {sprint.estado === 'Planificado' && (
                          <Button
                            variante="secondary"
                            disabled={hayActivo || procesandoId === sprint.id}
                            onClick={() => setConfirmacion({ sprint, destino: 'Activo' })}
                          >
                            Activar
                          </Button>
                        )}
                        {sprint.estado === 'Activo' && (
                          <Button
                            variante="secondary"
                            disabled={procesandoId === sprint.id}
                            onClick={() => setConfirmacion({ sprint, destino: 'Cerrado' })}
                          >
                            Cerrar
                          </Button>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
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
