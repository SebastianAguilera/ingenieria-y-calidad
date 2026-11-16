import { useState } from 'react'
import { useParams } from 'react-router-dom'
import type { Defecto, Historia, Sprint, TipoListadoDefecto } from '../../api/types'
import PageHeader from '../../components/PageHeader'
import Alert from '../../ui/Alert'
import Button from '../../ui/Button'
import Card from '../../ui/Card'
import {
  claseEstadoDefecto,
  claseSeveridadDefecto,
  codigoDefecto,
  etiquetaEstado,
  etiquetaSeveridad,
  truncarDescripcion,
} from './defectos/etiquetasDefecto'
import ModalDefecto from './defectos/ModalDefecto'
import ModalEdicionDefecto from './defectos/ModalEdicionDefecto'
import ModalEstadoDefecto from './defectos/ModalEstadoDefecto'
import { useDefectos } from './defectos/useDefectos'
import type { FiltroDefectos } from './defectos/useDefectos'

const SIN_DATO = '—'
const TODOS_LOS_SPRINTS = ''

const CLASES_SELECT =
  'appearance-none rounded-pill border border-border bg-[var(--surface-2)] py-1.5 pl-3 pr-8 text-sm text-ink focus:border-primary focus:outline-none disabled:cursor-not-allowed disabled:opacity-60'

function parsearProyectoId(id: string | undefined): number | null {
  if (id === undefined) {
    return null
  }
  const numero = Number(id)
  return Number.isInteger(numero) && numero > 0 ? numero : null
}

function ordenarMasRecientePrimero(defectos: Defecto[]): Defecto[] {
  return [...defectos].sort((a, b) => b.id - a.id)
}

function indexarPorId<T extends { id: number }>(elementos: T[]): Map<number, T> {
  return new Map(elementos.map((elemento) => [elemento.id, elemento]))
}

function esTipo(valor: string): valor is TipoListadoDefecto {
  return valor === 'deteccion' || valor === 'resolucion'
}

export default function DefectosPage() {
  const { id } = useParams()
  const proyectoId = parsearProyectoId(id)
  const [sprintFiltro, setSprintFiltro] = useState(TODOS_LOS_SPRINTS)
  const [tipoFiltro, setTipoFiltro] = useState<TipoListadoDefecto>('deteccion')
  const filtro: FiltroDefectos | null =
    sprintFiltro === TODOS_LOS_SPRINTS ? null : { sprintId: Number(sprintFiltro), tipo: tipoFiltro }
  const { estado, reintentar, crear, cambiarEstado, editar, procesandoId } = useDefectos(
    proyectoId,
    filtro,
  )
  const [altaAbierta, setAltaAbierta] = useState(false)
  const [defectoEnEstado, setDefectoEnEstado] = useState<Defecto | null>(null)
  const [defectoEnEdicion, setDefectoEnEdicion] = useState<Defecto | null>(null)

  const defectos = estado.tipo === 'listo' ? estado.defectos : []
  const historias = estado.tipo === 'listo' ? estado.historias : []
  const sprints = estado.tipo === 'listo' ? estado.sprints : []
  const historiaPorId: Map<number, Historia> = indexarPorId(historias)
  const sprintPorId: Map<number, Sprint> = indexarPorId(sprints)
  const cargandoDefectos = estado.tipo === 'listo' && estado.cargandoDefectos
  const hayFiltro = filtro !== null
  const mostrarBarra =
    estado.tipo === 'listo' && (hayFiltro || (!cargandoDefectos && defectos.length > 0))

  function nombreSprint(sprintId: number | null): string {
    if (sprintId === null) {
      return SIN_DATO
    }
    return sprintPorId.get(sprintId)?.identificador ?? SIN_DATO
  }

  return (
    <section>
      <PageHeader
        title="Defectos"
        description="Registrá los defectos del proyecto y seguí su estado hasta la resolución."
      />

      {(estado.tipo === 'cargando' || cargandoDefectos) && (
        <p role="status" className="mb-5 text-ink-soft">
          Cargando defectos…
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

      {estado.tipo === 'listo' && !cargandoDefectos && defectos.length === 0 && !hayFiltro && (
        <Card
          title="Todavía no hay defectos"
          description="Registrá el primer defecto para empezar a seguir la calidad del proyecto."
        >
          <Button onClick={() => setAltaAbierta(true)}>Nuevo defecto</Button>
        </Card>
      )}

      {mostrarBarra && (
        <div className="mb-5 flex flex-wrap items-center justify-between gap-3">
          <div className="flex flex-wrap items-center gap-4">
            <label className="flex items-center gap-2 text-sm text-ink-soft">
              Filtrar por sprint
              <span className="relative inline-flex items-center">
                <select
                  value={sprintFiltro}
                  onChange={(evento) => setSprintFiltro(evento.target.value)}
                  className={CLASES_SELECT}
                >
                  <option value={TODOS_LOS_SPRINTS}>Todos los sprints</option>
                  {sprints.map((sprint) => (
                    <option key={sprint.id} value={String(sprint.id)}>
                      {sprint.identificador}
                    </option>
                  ))}
                </select>
                <i
                  className="bi bi-chevron-down pointer-events-none absolute right-2.5 text-xs text-ink-soft"
                  aria-hidden="true"
                />
              </span>
            </label>
            <label className="flex items-center gap-2 text-sm text-ink-soft">
              Tipo de filtro
              <span className="relative inline-flex items-center">
                <select
                  value={tipoFiltro}
                  disabled={!hayFiltro}
                  onChange={(evento) => {
                    if (esTipo(evento.target.value)) {
                      setTipoFiltro(evento.target.value)
                    }
                  }}
                  className={CLASES_SELECT}
                >
                  <option value="deteccion">Detección</option>
                  <option value="resolucion">Resolución</option>
                </select>
                <i
                  className="bi bi-chevron-down pointer-events-none absolute right-2.5 text-xs text-ink-soft"
                  aria-hidden="true"
                />
              </span>
            </label>
          </div>
          <Button onClick={() => setAltaAbierta(true)}>Nuevo defecto</Button>
        </div>
      )}

      {estado.tipo === 'listo' && !cargandoDefectos && defectos.length === 0 && hayFiltro && (
        <Alert tono="info">No hay defectos con ese filtro</Alert>
      )}

      {estado.tipo === 'listo' && !cargandoDefectos && defectos.length > 0 && (
        <div className="tabla-contenedor barra-fina">
          <table className="tabla tabla-compacta">
            <thead>
              <tr>
                <th scope="col" className="text-center">
                  Código
                </th>
                <th scope="col" className="text-center">
                  Descripción
                </th>
                <th scope="col" className="text-center">
                  Historia
                </th>
                <th scope="col" className="text-center">
                  Severidad
                </th>
                <th scope="col" className="text-center">
                  Estado
                </th>
                <th scope="col" className="text-center">
                  Sprint de detección
                </th>
                <th scope="col" className="text-center">
                  Sprint de resolución
                </th>
                <th scope="col" className="col-acciones">
                  Acciones
                </th>
              </tr>
            </thead>
            <tbody>
              {ordenarMasRecientePrimero(defectos).map((defecto) => (
                <tr key={defecto.id}>
                  <td className="text-center whitespace-nowrap font-medium text-ink">
                    {codigoDefecto(defecto.id)}
                  </td>
                  <td className="text-center">
                    <span
                      title={defecto.descripcion}
                      className="mx-auto block max-w-[16rem] [overflow-wrap:anywhere]"
                    >
                      {truncarDescripcion(defecto.descripcion)}
                    </span>
                  </td>
                  <td className="text-center">
                    {historiaPorId.get(defecto.historia_id)?.codigo ?? SIN_DATO}
                  </td>
                  <td className="text-center">
                    <span
                      className={`inline-flex items-center rounded-pill border px-3 py-1.5 text-xs font-semibold ${claseSeveridadDefecto(defecto.severidad)}`}
                    >
                      {etiquetaSeveridad(defecto.severidad)}
                    </span>
                  </td>
                  <td className="text-center">
                    <span
                      className={`inline-flex items-center whitespace-nowrap rounded-pill border px-3 py-1.5 text-xs font-semibold ${claseEstadoDefecto(defecto.estado)}`}
                    >
                      {etiquetaEstado(defecto.estado)}
                    </span>
                  </td>
                  <td className="text-center">{nombreSprint(defecto.sprint_deteccion_id)}</td>
                  <td className="text-center">{nombreSprint(defecto.sprint_resolucion_id)}</td>
                  <td className="col-acciones">
                    <div className="tabla-acciones">
                      <Button
                        variante="outline"
                        tamano="sm"
                        className="whitespace-nowrap"
                        disabled={procesandoId === defecto.id}
                        aria-label={`Cambiar estado de ${codigoDefecto(defecto.id)}`}
                        onClick={() => setDefectoEnEstado(defecto)}
                      >
                        Cambiar estado
                      </Button>
                      <Button
                        variante="outline"
                        tamano="sm"
                        className="whitespace-nowrap"
                        disabled={procesandoId === defecto.id}
                        aria-label={`Editar ${codigoDefecto(defecto.id)}`}
                        onClick={() => setDefectoEnEdicion(defecto)}
                      >
                        Editar
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {altaAbierta && (
        <ModalDefecto
          historias={historias}
          sprints={sprints}
          onCerrar={() => setAltaAbierta(false)}
          onGuardar={async (alta) => {
            const resultado = await crear(alta)
            if (resultado.ok) {
              setAltaAbierta(false)
            }
            return resultado
          }}
        />
      )}

      {defectoEnEstado && (
        <ModalEstadoDefecto
          defecto={defectoEnEstado}
          sprints={sprints}
          onCerrar={() => setDefectoEnEstado(null)}
          onGuardar={async (actualizacion) => {
            const resultado = await cambiarEstado(defectoEnEstado.id, actualizacion)
            if (resultado.ok) {
              setDefectoEnEstado(null)
            }
            return resultado
          }}
        />
      )}

      {defectoEnEdicion && (
        <ModalEdicionDefecto
          defecto={defectoEnEdicion}
          onCerrar={() => setDefectoEnEdicion(null)}
          onGuardar={async (actualizacion) => {
            const resultado = await editar(defectoEnEdicion.id, actualizacion)
            if (resultado.ok) {
              setDefectoEnEdicion(null)
            }
            return resultado
          }}
        />
      )}
    </section>
  )
}
