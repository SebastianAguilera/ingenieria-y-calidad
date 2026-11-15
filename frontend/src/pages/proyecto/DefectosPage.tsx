import { useState } from 'react'
import { useParams } from 'react-router-dom'
import type { Defecto, Historia, Sprint } from '../../api/types'
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
import ModalEstadoDefecto from './defectos/ModalEstadoDefecto'
import { useDefectos } from './defectos/useDefectos'

const SIN_DATO = '—'

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

export default function DefectosPage() {
  const { id } = useParams()
  const proyectoId = parsearProyectoId(id)
  const { estado, reintentar, crear, cambiarEstado, procesandoId } = useDefectos(proyectoId)
  const [altaAbierta, setAltaAbierta] = useState(false)
  const [defectoEnEstado, setDefectoEnEstado] = useState<Defecto | null>(null)

  const defectos = estado.tipo === 'listo' ? estado.defectos : []
  const historias = estado.tipo === 'listo' ? estado.historias : []
  const sprints = estado.tipo === 'listo' ? estado.sprints : []
  const historiaPorId: Map<number, Historia> = indexarPorId(historias)
  const sprintPorId: Map<number, Sprint> = indexarPorId(sprints)
  const cargandoDefectos = estado.tipo === 'listo' && estado.cargandoDefectos

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
        <p role="status" className="text-ink-soft">
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

      {estado.tipo === 'listo' && !cargandoDefectos && defectos.length === 0 && (
        <Card
          title="Todavía no hay defectos"
          description="Registrá el primer defecto para empezar a seguir la calidad del proyecto."
        >
          <Button onClick={() => setAltaAbierta(true)}>Nuevo defecto</Button>
        </Card>
      )}

      {estado.tipo === 'listo' && !cargandoDefectos && defectos.length > 0 && (
        <>
          <div className="mb-5 flex flex-wrap items-center justify-end gap-3">
            <Button onClick={() => setAltaAbierta(true)}>Nuevo defecto</Button>
          </div>

          <div className="overflow-x-auto rounded-tile border border-border bg-surface">
            <table className="w-full min-w-[64rem] text-left text-sm">
              <thead className="border-b border-border text-ink-soft">
                <tr>
                  <th scope="col" className="px-6 py-4 font-semibold">
                    Código
                  </th>
                  <th scope="col" className="px-6 py-4 font-semibold">
                    Descripción
                  </th>
                  <th scope="col" className="px-6 py-4 font-semibold">
                    Historia
                  </th>
                  <th scope="col" className="px-6 py-4 font-semibold">
                    Severidad
                  </th>
                  <th scope="col" className="px-6 py-4 font-semibold">
                    Estado
                  </th>
                  <th scope="col" className="px-6 py-4 font-semibold">
                    Sprint de detección
                  </th>
                  <th scope="col" className="px-6 py-4 font-semibold">
                    Sprint de resolución
                  </th>
                  <th scope="col" className="px-6 py-4 font-semibold">
                    Acciones
                  </th>
                </tr>
              </thead>
              <tbody>
                {ordenarMasRecientePrimero(defectos).map((defecto) => (
                  <tr key={defecto.id} className="border-b border-border last:border-b-0">
                    <td className="whitespace-nowrap px-6 py-5 font-medium text-ink">
                      {codigoDefecto(defecto.id)}
                    </td>
                    <td className="px-6 py-5">
                      <span
                        title={defecto.descripcion}
                        className="block max-w-sm [overflow-wrap:anywhere]"
                      >
                        {truncarDescripcion(defecto.descripcion)}
                      </span>
                    </td>
                    <td className="px-6 py-5">
                      {historiaPorId.get(defecto.historia_id)?.codigo ?? SIN_DATO}
                    </td>
                    <td className="px-6 py-5">
                      <span
                        className={`inline-flex items-center rounded-pill border px-3 py-1.5 text-xs font-semibold ${claseSeveridadDefecto(defecto.severidad)}`}
                      >
                        {etiquetaSeveridad(defecto.severidad)}
                      </span>
                    </td>
                    <td className="px-6 py-5">
                      <span
                        className={`inline-flex items-center rounded-pill border px-3 py-1.5 text-xs font-semibold ${claseEstadoDefecto(defecto.estado)}`}
                      >
                        {etiquetaEstado(defecto.estado)}
                      </span>
                    </td>
                    <td className="px-6 py-5">{nombreSprint(defecto.sprint_deteccion_id)}</td>
                    <td className="px-6 py-5">{nombreSprint(defecto.sprint_resolucion_id)}</td>
                    <td className="px-6 py-5">
                      <div className="flex flex-wrap gap-2">
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
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </>
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
    </section>
  )
}
